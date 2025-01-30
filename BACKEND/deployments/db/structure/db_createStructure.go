package structure

// #### С Создаем структуру БД, после запускаем заполнение стартовой информацией о гильдии с АПИ raid.io через defer firstfilldb.FirstFillDB()

import (
	"context"
	"fmt"
	filldb "kvd/deployments/db/filldb"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	// "github.com/spf13/viper"
)

func Init() {
	// Конфигурация подключения
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")
	dbhost := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("HOST_DB_PORT")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s", dbUser, dbPassword, dbhost, dbPort)
	connConfig, err := pgx.ParseConfig(dbUrl)
	if err != nil {
		log.Printf("Ошибка в конфигурации: %v\n", err)
	}
	dbBuild(connConfig)
}

var conn *pgx.Conn
var retryDelay = 60 * time.Second

func dbBuild(connConfig *pgx.ConnConfig) {
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
	logger := log.New(file, "[DEPLOY] ", log.LstdFlags|log.Lshortfile)

	// Попытка подключения, если ошибка, ждем и пытаемся еще раз
	for {
		conn, err = pgx.ConnectConfig(context.Background(), connConfig)
		if err != nil {
			log.Printf("Ошибка подключения к PostgreSQL: %v\n", err)
			logger.Printf("Ошибка подключения к PostgreSQL: %v\n", err)
			log.Printf("Пытаемся переподключиться %s...\n", retryDelay)
			time.Sleep(retryDelay)
		} else {
			logger.Printf("Успешно подключились к PostgreSQL\n")
			log.Printf("Успешно подключились к PostgreSQL\n")
			break
		}
	}

	// Закрыть соединение после выполнения функции
	defer conn.Close(context.Background())

	dbName := os.Getenv("DB_NAME")

	// Проверяем, существует ли база данных
	checkDBExistsQuery := fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s'", dbName)
	var existsDB string // Изменено с bool на string
	err = conn.QueryRow(context.Background(), checkDBExistsQuery).Scan(&existsDB)
	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Ошибка проверки существования базы данных: %v\n", err)
		logger.Printf("Ошибка проверки существования базы данных: %v\n", err)
	}

	if existsDB == "" { // Проверяем, что exists пустая строка, что означает отсутствие базы данных
		// SQL запрос для создания БД
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)

		// Выполняем запрос
		_, err = conn.Exec(context.Background(), createDBQuery)
		if err != nil {
			log.Printf("Ошибка при создании БД: %v\n", err)
			logger.Println("Ошибка при создании БД: %s %v\n", dbName, err)
		}
		logger.Printf("БД: %s успешно создана", dbName)
		log.Printf("БД: %s успешно создана\n", dbName)
	} else {
		logger.Printf("БД: %s уже существует", dbName)
		log.Printf("БД: %s уже существует\n", dbName)
		// Удаление комментариев о необходимости удаления базы данных перед созданием, так как это не требуется
	}

	// Подключение к базе данных kvd_guild
	connConfig.Database = os.Getenv("DB_NAME")
	conn, err = pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		logger.Printf("Ошибка подключения к новой БД: %v\n", err)
		log.Printf("Ошибка подключения к новой БД: %v\n", err)
	}
	// for {
	// 	conn, err = pgx.ConnectConfig(context.Background(), connConfig)

	// 	if err == nil {
	// 		logger.Printf("Connected to PostgreSQL\n")
	// 		log.Printf("Connected to PostgreSQL\n")
	// 		break
	// 	} else {
	// 		fmt.Printf("Error connecting to PostgreSQL: %v\n", err)
	// 		logger.Fatalf("Error connecting to PostgreSQL: %v\n", err)
	// 		fmt.Printf("Retrying connection in %s...\n", retryDelay)
	// 		time.Sleep(retryDelay)
	// 	}
	// }
	defer conn.Close(context.Background())

	// SQL запрос для создания таблиц и столбцов в базе kvd_guild
	createTableAndRow := `
	CREATE TABLE IF NOT EXISTS guild (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		faction VARCHAR(255) ,
		region VARCHAR(255) ,
		realm VARCHAR(255),
		profile_url VARCHAR(255) ,
		created_at TIMESTAMP DEFAULT now()
	);
	CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    rank INTEGER,
    name VARCHAR(255) NOT NULL,
    mythic_plus_scores_by_season INTEGER DEFAULT 0,
    guild VARCHAR(255),
    realm VARCHAR(255) DEFAULT '',
    race VARCHAR(255),
    class VARCHAR(255),
    gender VARCHAR(255),
    faction VARCHAR(255),
    achievement_points INTEGER,
    profile_url VARCHAR(255),
    thumbnail_url VARCHAR(255) DEFAULT '',
    profile_banner VARCHAR(255),
    created_at TIMESTAMP DEFAULT now()
);

	`

	_, err = conn.Exec(context.Background(), createTableAndRow)
	if err != nil {
		log.Println("Ошибка при создании таблицы: %v\n", err)
		logger.Println("Ошибка при создании таблицы: %v\n", err)
	} else {
		logger.Printf("Таблица успешно создана\n")
		log.Printf("Таблица успешно создана\n")
	}
	defer filldb.FirstFillDB()
}
