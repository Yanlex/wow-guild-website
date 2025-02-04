package update

import (
	"context"
	"fmt"
	fetch "kvd/internal/api/raiderio"
	"kvd/internal/db"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var digitRegex = regexp.MustCompile(`\d`) // Компилируем регулярное выражение один раз
var (
	dbUser, dbPassword, dbName, dbhost, dbPort, guildRegion string
)

func init() {
	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASS")
	guildRegion = os.Getenv("GUILD_REGION")
	dbName = os.Getenv("DB_NAME")
	dbhost = os.Getenv("DB_ADDRESS")
	dbPort = os.Getenv("HOST_DB_PORT")
}

// Структура для игроков с API
type apiPlayer struct {
	rank              int
	name              string
	guild             string
	realm             string
	race              string
	class             string
	gender            string
	faction           string
	achievementPoints int
	profileURL        string
	profileBanner     string
}
type PlayerDB struct {
	rank                         int
	name                         string
	mythic_plus_scores_by_season int
	guild                        string
	realm                        string
	race                         string
	class                        string
	gender                       string
	faction                      string
	achievementPoints            int
	profileURL                   string
	thumbnail_url                string
	profileBanner                string
}

func UpdateAllPlayers() {

	ctx := context.Background()

	// Подключаемся к БД
	db := db.NewPostgreSQL(dbPort, dbUser, dbPassword, dbhost, dbName)
	err := db.Connect(ctx)
	if err != nil {
		log.Println(err)
	}
	defer db.Disconnect()

	// Логирование в файл
	logger, file := logsUpdateAllPlayers()

	queryPlayersData := "SELECT rank, name, mythic_plus_scores_by_season, guild, realm, race, class, gender, faction, achievement_points, profile_url,thumbnail_url, profile_banner FROM members"
	// Получаем из Базы данных таблицу members
	rows, err := db.Query(ctx, queryPlayersData)
	if err != nil {
		log.Printf("Ошибка в запросе к БД: %v\n", err)
	}
	defer rows.Close()

	// var dbPlayersAllInfoSlice []PlayerDB
	dbPlayersAllInfoMap := make(map[string]PlayerDB)

	for rows.Next() {
		var player PlayerDB
		if err := rows.Scan(
			&player.rank,
			&player.name,
			&player.mythic_plus_scores_by_season,
			&player.guild,
			&player.realm,
			&player.race,
			&player.class,
			&player.gender,
			&player.faction,
			&player.achievementPoints,
			&player.profileURL,
			&player.thumbnail_url,
			&player.profileBanner); err != nil {
			log.Println(err)
		}

		// dbPlayersAllInfoSlice = append(dbPlayersAllInfoSlice, player)
		dbPlayersAllInfoMap[player.name] = player

		if err := rows.Err(); err != nil {
			log.Println(err)
		}
	}

	// Получаем список ников из таблицы members
	selectAllPlayers := `SELECT name FROM members;`
	playerRows, err := db.Query(ctx, selectAllPlayers)
	if err != nil {
		log.Printf("Ошибка, Не могу получить список игроков: %v\n", err)
	} else {
		fmt.Println("Успешно получен список игроков из БД")
	}
	defer playerRows.Close()

	playerCh := make(chan apiPlayer)
	go processGuildMembersJSON(playerCh)

	for player := range playerCh {

		// Проверяем есть ли игрок из API в нашей БД
		_, found := dbPlayersAllInfoMap[player.name]
		// Исключаем ники в которых есть цифры, они нам не подходят.
		match := digitRegex.MatchString(player.name)

		if !found {
			insertObject(ctx, player, db, logger)
		}

		if found && !match {
			p := dbPlayersAllInfoMap[player.name]

			// time sleep нужен из за ограничения запросов на сторонний API
			time.Sleep(900 * time.Millisecond)
			// mythic plus requests
			// Гет запрос

			// Кодирование имени персонажа в URL-кодированный формат, если не кодировать имя персонажа, то API вернет ошибку, почему не знаю.
			encodedName := url.QueryEscape(player.name)
			encodedPlayerRealm := url.QueryEscape(player.realm)

			// Делаем запрос на API
			playerResp, err := fetch.MemberRio(guildRegion, encodedPlayerRealm, encodedName)
			if err != nil {
				log.Println(err)
			}

			// Достаем текущий рейтинг из gjson.Response
			playerRio := gjson.Get(playerResp, "mythic_plus_scores_by_season.#.scores.all")
			var currRioRating int
			// Конвертируем gjson.Response в string
			for _, s := range playerRio.Array() {
				currRioRating = int(s.Int())
			}

			// Достаем thumbnail_url
			playerThumbnailUrl := gjson.Get(playerResp, "thumbnail_url")
			playerThumbnailUrlString := playerThumbnailUrl.String()

			if p.thumbnail_url == "" || player.rank != p.rank || p.mythic_plus_scores_by_season != currRioRating || player.guild != p.guild || player.realm != p.realm || player.race != p.race || player.gender != p.gender || player.achievementPoints != p.achievementPoints || player.profileURL != p.profileURL || player.profileBanner != p.profileBanner {
				updateQuery := "UPDATE members SET "

				var updates []string

				if player.rank != p.rank {
					updates = append(updates, fmt.Sprintf(`rank = '%d'`, player.rank)) // Использование двойных кавычек для строки и %s для интерполяции
				}
				// if player.guild != guild {
				// 	updates = append(updates, fmt.Sprintf("guild = '%s'", player.guild))
				// }
				if player.realm != p.realm {
					updates = append(updates, fmt.Sprintf("realm = '%s'", player.realm))
				}
				if player.race != p.race && player.race != "Mag'har Orc" {
					raceFix := strings.ReplaceAll(player.race, "'", " ")
					updates = append(updates, fmt.Sprintf("race = '%s'", raceFix))
				}
				if player.gender != p.gender {
					updates = append(updates, fmt.Sprintf("gender = '%s'", player.gender))
				}
				if player.achievementPoints != p.achievementPoints {
					updates = append(updates, fmt.Sprintf("achievement_points = '%d'", player.achievementPoints))
				}
				if player.profileURL != p.profileURL {
					updates = append(updates, fmt.Sprintf("profile_url = '%s'", player.profileURL))
				}

				if p.mythic_plus_scores_by_season != currRioRating {
					updates = append(updates, fmt.Sprintf("mythic_plus_scores_by_season = '%d'", currRioRating))
				}
				if p.thumbnail_url == "" {
					updates = append(updates, fmt.Sprintf("thumbnail_url = '%s'", playerThumbnailUrlString))
				}

				if len(updates) > 0 {
					updateQuery += strings.Join(updates, ", ")
					updateQuery += fmt.Sprintf(` WHERE name = '%s'`, player.name)
					fmt.Println(updateQuery)
					err := db.Exec(ctx, updateQuery)
					if err != nil {
						log.Println(err)
					} else {
						logger.Println("Обновили данные игрока: ", player.name, updateQuery)
					}
				}
			}

		}
	}

	defer file.Close()
}

// Добавляем игрока в базу данных
func insertObject(ctx context.Context, p apiPlayer, db *db.PostgreSQL, logger *log.Logger) {
	// Вставка данных в таблицу members
	err := db.Exec(ctx, `
        INSERT INTO members (rank, name, guild, realm, race, class, gender, faction, achievement_points, profile_url, profile_banner, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
    `, p.rank, p.name, p.guild, p.realm, p.race, p.class, p.gender, p.faction, p.achievementPoints, p.profileURL, p.profileBanner)
	if err != nil {
		logger.Println("Ошибка добавления игрока: ", p.name, `в БД`, err)
	} else {
		logger.Println(p.name, `новый игрок`)
	}
}

func logsUpdateAllPlayers() (*log.Logger, *os.File) {
	// Получение пути к домашнему каталогу
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	logFilePath := fmt.Sprintf("%s/kvd/logs/updatePlayers.log", homeDir)

	// Создание всех необходимых каталогов, если они еще не существуют
	err = os.MkdirAll(fmt.Sprintf("%s/kvd/logs", homeDir), 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// Создаем логирование в файл logs/update/updatePlayers.log
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log := log.New(file, "[UPDATEPlAYERS] ", log.LstdFlags|log.Lshortfile)
	return log, file
}

func processGuildMembersJSON(playerCh chan apiPlayer) {
	var resp string

	for {
		// Получаем данные из API
		resp = fetch.GuildRio()
		if resp == "" {
			log.Println("Ошибка подключения к API")
		} else {
			break
		}
		time.Sleep(5 * time.Minute)
	}

	// Итерация по очень большому json объекту.
	totalMembers := gjson.Get(resp, "members.#")
	for i := 0; i < int(totalMembers.Int()); i++ {

		// i к int64
		// Создание пути с использованием fmt.Sprintf иначе gjson.Get выдаст ошибку too many arguments in call to gjson.Get
		rankPath := fmt.Sprintf("members.%d.rank", i) // Создание пути с использованием fmt.Sprintf
		rank := gjson.Get(resp, rankPath)

		namePath := fmt.Sprintf("members.%d.character.name", i)
		name := gjson.Get(resp, namePath)

		guild := "ключик в дурку"

		realmPath := fmt.Sprintf("members.%d.character.realm", i)
		realm := gjson.Get(resp, realmPath)

		racePath := fmt.Sprintf("members.%d.character.race", i)
		race := gjson.Get(resp, racePath)

		classPath := fmt.Sprintf("members.%d.character.class", i)
		class := gjson.Get(resp, classPath)

		genderPath := fmt.Sprintf("members.%d.character.gender", i)
		gender := gjson.Get(resp, genderPath)

		factionPath := fmt.Sprintf("members.%d.character.faction", i)
		faction := gjson.Get(resp, factionPath)

		achievementPointsPath := fmt.Sprintf("members.%d.character.achievement_points", i)
		achievement_points := gjson.Get(resp, achievementPointsPath)

		profileURLPath := fmt.Sprintf("members.%d.character.profile_url", i)
		profile_url := gjson.Get(resp, profileURLPath)

		profileBannerPath := fmt.Sprintf("members.%d.character.profile_banner", i)
		profile_banner := gjson.Get(resp, profileBannerPath)

		player := apiPlayer{
			rank:              int(rank.Int()),
			name:              name.String(),
			guild:             guild,
			realm:             realm.String(),
			race:              race.String(),
			class:             class.String(),
			gender:            gender.String(),
			faction:           faction.String(),
			achievementPoints: int(achievement_points.Int()),
			profileURL:        profile_url.String(),
			profileBanner:     profile_banner.String(),
		}
		playerCh <- player
	}
	defer close(playerCh)
}
