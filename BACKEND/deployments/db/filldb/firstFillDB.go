package filldb

// #### Этот код запускается из db_createStructure.go
// #### Запускется один раз после создания базы данных для заполнения базы данных стартовой информацией о гильдии с АПИ Raider.IO

import (
	"context"
	"fmt"
	fetch "kvd/internal/api/raiderio"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
)

var pool *pgxpool.Pool
var err error
var ctx = context.Background()

func init() {
	// config.InitConfigDB()
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbhost := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("HOST_DB_PORT")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbhost, dbPort, dbName)
	// dbUrl := viper.GetString("db.urlKvd")
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Println("Ошибка в конфигурации подключения к БД: %v", err)
	}
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Println("Ошибка подключения к БД %v", err)
	}
}

type Player struct {
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

func FirstFillDB() {
	// Получение пути к домашнему каталогу
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
	}

	logFilePath := fmt.Sprintf("%s/kvd/logs/deploy.log", homeDir)

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
	defer file.Close()

	logger := log.New(file, "[FirstFillDB] ", log.LstdFlags|log.Lshortfile)

	resp := fetch.FetchRaiderIo()
	if resp == "" {
		log.Println("Ошибка получения данных из API")
		logger.Println("Ошибка получения данных из API")
	}
	// if err != nil {
	// 	log.Fatalf("Unable to create pool: %v", err)
	// 	logger.Fatalf("Unable to create pool: %v", err)
	// }

	// ctx := context.Background()
	rows, err := pool.Query(ctx, "SELECT name FROM guild")
	if err != nil {
		log.Println("Ошибка в запросе к БД: %v\n", err)
	}
	defer rows.Close()

	// Считаем количество строк
	var count int
	for rows.Next() {
		count++
		if count > 1 {
			break
		}
	}
	defer pool.Close()
	if count == 0 {
		// Имя гильдии
		name := gjson.Get(resp, "name").String()
		if name == "" {
			logger.Println("Ошибка при попытке извлечь имя игрока", err)
			log.Println("Ошибка при попытке извлечь имя игрока", err)
		}

		// Фракция
		faction := gjson.Get(resp, "faction").String()

		// Регион
		region := gjson.Get(resp, "region").String()

		// Реалм
		realm := gjson.Get(resp, "realm").String()

		// Профиль
		profile_url := gjson.Get(resp, "profile_url").String()

		// Вставка данных в таблицу guild
		_, err = pool.Exec(ctx, `
        INSERT INTO guild (name, faction, region, realm, profile_url, created_at)
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
    `, name, faction, region, realm, profile_url)
		if err != nil {
			logger.Printf("Ошибка, не удалось вставить данные: %v", err)
			log.Println("Ошибка, не удалось вставить данные: %v", err)
		}
		defer log.Println("Успех, данные гильдии добавлены")
	} else {
		log.Println("Похоже данные в БД уже сущесуют, идем дальше")
	}
	defer fillPlayers(resp, file)
}

func fillPlayers(resp string, file *os.File) {
	logger := log.New(file, "[fillPlayers] ", log.LstdFlags|log.Lshortfile)

	totalMembers := gjson.Get(resp, "members.#")

	// ctx := context.Background()
	rows, err := pool.Query(ctx, "SELECT name FROM members")
	if err != nil {
		log.Println("Ошибка, не удалось выполнить запрос: %v\n", err)
	}
	defer rows.Close()

	// Считаем количество строк
	var count int
	for rows.Next() {
		count++
		if count > 1 {
			break
		}
	}

	if count == 0 {
		semaphoreBD := make(chan struct{}, 10)
		wg := sync.WaitGroup{}

		for i := 0; i < int(totalMembers.Int()); i++ {
			wg.Add(1)
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

			player := Player{
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
			// Логируем полученного игрока
			defer logger.Println(player)

			go func(p Player) {
				defer wg.Done()
				semaphoreBD <- struct{}{}
				insertObject(p)
				defer func() { <-semaphoreBD }()
			}(player)
		}
		wg.Wait()
		defer log.Println("Данные об игроках гильдии успешно вставлены в БД")
	} else {
		defer log.Println("Похоже в БД уже есть данные об игроках, идем дальше.")
	}
	defer file.Close()

}

func insertObject(p Player) {
	// ctx := context.Background()
	// Вставка данных в таблицу members
	_, err = pool.Exec(ctx, `
        INSERT INTO members (rank, name, guild, realm, race, class, gender, faction, achievement_points, profile_url, profile_banner, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
    `, p.rank, p.name, p.guild, p.realm, p.race, p.class, p.gender, p.faction, p.achievementPoints, p.profileURL, p.profileBanner)
	if err != nil {
		log.Println("Ошибка, не удалось добавить игрока: %v\n", err)
		log.Println(p)
	}

}
