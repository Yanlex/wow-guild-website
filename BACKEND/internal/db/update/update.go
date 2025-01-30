package update

import (
	"context"
	"fmt"
	"io"
	fetch "kvd/internal/api/raiderio"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
)

var pool *pgxpool.Pool
var ctx context.Context
var logger *log.Logger
var digitRegex = regexp.MustCompile(`\d`) // Компилируем регулярное выражение один раз

// var file *os.File

func init() {

}

type PlayerBase struct {
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
	// Получаем конфигурацию соединения с БД
	// config.InitConfigDB()

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")
	guildRegion := os.Getenv("GUILD_REGION")
	guildDBName := os.Getenv("DB_NAME")
	dbhost := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("HOST_DB_PORT")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbhost, dbPort, guildDBName)

	// dbUrl := viper.GetString("db.urlKvd")
	ctx = context.Background()
	// fmt.Println(dbUrl)
	connConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Println("Ошибка в конфигурации: %v\n", err)
	}
	// Создаем пул соединений
	pool, err = pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		log.Println("Ошибка подключения к БД: %v\n", err)
	} else {
		fmt.Printf("Успешно подключились к БД\n")
	}

	// Получение пути к домашнему каталогу
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
	}

	logFilePath := fmt.Sprintf("%s/kvd/logs/updatePlayers.log", homeDir)

	// Создание всех необходимых каталогов, если они еще не существуют
	err = os.MkdirAll(fmt.Sprintf("%s/kvd/logs", homeDir), 0755)
	if err != nil {
		log.Println(err)
	}

	// Создаем логирование в файл logs/update/updatePlayers.log
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	logger = log.New(file, "[UPDATEPlAYERS] ", log.LstdFlags|log.Lshortfile)
	fmt.Println("Обновление данных игроков начато")

	// Получаем данные из API
	resp := fetch.FetchRaiderIo()
	if resp == "" {
		log.Println("Ошибка подключения к API")
	}

	// Получаем из Базы данных таблицу members
	rows, err := pool.Query(context.Background(), "SELECT rank, name, mythic_plus_scores_by_season, guild, realm, race, class, gender, faction, achievement_points, profile_url,thumbnail_url, profile_banner FROM members")
	if err != nil {
		log.Println("Ошибка в запросе к БД: %v\n", err)
	}
	defer rows.Close()

	var players []PlayerDB
	for rows.Next() {

		var player PlayerDB
		if err := rows.Scan(&player.rank, &player.name, &player.mythic_plus_scores_by_season, &player.guild, &player.realm, &player.race, &player.class, &player.gender, &player.faction, &player.achievementPoints, &player.profileURL, &player.thumbnail_url, &player.profileBanner); err != nil {
			log.Println(err)
		}

		players = append(players, player)

		if err := rows.Err(); err != nil {
			log.Println(err)
		}
	}

	// Получаем список ников из таблицы members
	playersFromDB := `SELECT name FROM members;`
	playerRows, err := pool.Query(context.Background(), playersFromDB)
	if err != nil {
		log.Println("Ошибка, Не могу получить список игроков: %v\n", err)
	} else {
		fmt.Println("Успешно получен список игроков из БД")
	}
	defer playerRows.Close()

	var playerNames []string
	// Помещаем наймена игроков в playerNames
	for playerRows.Next() {
		var name string
		err := playerRows.Scan(&name)
		if err != nil {
			log.Println("Scan error: %v\n", err)
		}
		playerNames = append(playerNames, name)
	}

	totalMembers := gjson.Get(resp, "members.#")
	for i := 0; i < int(totalMembers.Int()); i++ {
		// Приведение i к int64
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

		player := PlayerBase{
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
		// M+ scores
		// mythic_plus_scores_by_season
		// https://raider.io/api/v1/characters/profile?region=eu&realm=howling-fjord&name=%D0%A7%D0%BE%D1%81%D0%BA%D0%B8&fields=mythic_plus_scores_by_season%3Acurrent
		found2 := slices.Contains(playerNames, player.name)
		match := digitRegex.MatchString(player.name)

		if found2 && !match {

			// fmt.Println("Found", player.name)
			// Это итеррация по всей полученной таблице members
			for _, p := range players {
				// fmt.Println(p.rank, p.name, p.guild, p.realm, p.race, p.class, p.gender, p.faction, p.achievementPoints, p.profileURL, p.profileBanner)
				if player.name == p.name {
					// time sleep нужыен из за ограничения запросов на стороний API
					time.Sleep(900 * time.Millisecond)
					// mythic plus requests
					// Гет запрос

					// Кодирование имени персонажа в URL-кодированный формат, если не кодировать имя персонажа, то API вернет ошибку, почему не знаю.
					fmt.Println("Имя игрока", player.name)
					encodedName := url.QueryEscape(player.name)
					playerRealm := url.QueryEscape(player.realm)
					// playeerGuild := url.QueryEscape(player.guild)

					// Делаем запрос на API
					url := fmt.Sprintf("https://raider.io/api/v1/characters/profile?region=%s&realm=%s&name=%s&fields=mythic_plus_scores_by_season:current", guildRegion, playerRealm, encodedName)
					respRio, err := tryFetchRio(url)
					if err != nil {
						log.Println(err)
					}

					// if err != nil {
					// 	// Здесь убрал фатал чтобы не крашить приложение, скорее всего превысили ограничение на количество запросов поэтому просто сделаем таймаут.
					// 	logger.Println("respRio: ", err)
					// 	time.Sleep(120 * time.Second)
					// 	// Пробуем еще раз через 2 минуты
					// 	// respRio, _ = http.Get(url)
					// }
					defer respRio.Body.Close()

					// Читаем данные из запроса
					body, err := io.ReadAll(respRio.Body)
					if err != nil {
						log.Println(err)
					}

					// Преобразование в строку
					playerResp := string(body)
					if playerResp == "" {
						log.Println("Failed to fetch player data from API")
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

					// fmt.Println("О, привет:" + player.name + " " + p.name)
					if p.thumbnail_url == "" || player.rank != p.rank || p.mythic_plus_scores_by_season != currRioRating || player.guild != p.guild || player.realm != p.realm || player.race != p.race || player.gender != p.gender || player.achievementPoints != p.achievementPoints || player.profileURL != p.profileURL || player.profileBanner != p.profileBanner {
						updateQuery := "UPDATE members SET "
						fmt.Println("Провалилсь в условие", player.name)

						var updates []string

						if player.rank != p.rank {
							updates = append(updates, fmt.Sprintf(`rank = '%d'`, player.rank)) // Использование двойных кавычек для строки и %s для интерполяции
						}
						if player.guild != guild {
							updates = append(updates, fmt.Sprintf("guild = '%s'", player.guild))
						}
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
							_, err := pool.Exec(ctx, updateQuery)
							if err != nil {
								log.Println(err)
							} else {
								logger.Println("Обновили данные игрока: ", player.name, updateQuery)
							}
						}
					}
				}
			}
		} else {
			// fmt.Println(playerJson)
			logger.Println("Игрок ", name.String(), `не найден в БД, вносим нового игрока.`)
			insertObject(player, pool)
		}
	}
	defer fmt.Println("Программа обновления данных игроков завершилась")
	defer file.Close()
	defer pool.Close()
}

// Ошибку возвращаем просто для практики.
func tryFetchRio(url string) (*http.Response, error) {
	for {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		logger.Println("Ошибка при запросе к API, повторная попытка через 5 минут", url, err)
		log.Println("Ошибка при запросе к API, повторная попытка через 5 минут", url, err)
		time.Sleep(5 * time.Minute)
	}
}

// Добавляем игрока в базу данных
func insertObject(p PlayerBase, pool *pgxpool.Pool) {
	ctx := context.Background()
	// Вставка данных в таблицу members
	_, err := pool.Exec(ctx, `
        INSERT INTO members (rank, name, guild, realm, race, class, gender, faction, achievement_points, profile_url, profile_banner, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
    `, p.rank, p.name, p.guild, p.realm, p.race, p.class, p.gender, p.faction, p.achievementPoints, p.profileURL, p.profileBanner)
	if err != nil {
		logger.Println("Ошибка добавления игрока: ", p.name, `в БД`, err)
	} else {
		logger.Println("Игрок ", p.name, `добавлен в БД`)
	}
}
