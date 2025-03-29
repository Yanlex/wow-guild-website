package structure

// #### С Создаем структуру БД, после запускаем заполнение стартовой информацией о гильдии с АПИ raid.io через defer firstfilldb.FirstFillDB()

import (
	"context"
	"fmt"
	filldb "kvd/deployments/db/filldb"
	"kvd/internal/db"
	"log"
	"os"
)

var (
	dbUser, dbPassword, dbName, dbhost, dbPort string
)

func Init() {
	// Конфигурация подключения
	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASS")
	dbhost = os.Getenv("DB_ADDRESS")
	dbPort = os.Getenv("HOST_DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbBuild()
}

func dbBuild() {
	ctx := context.Background()

	// Подключаемся к БД
	db := db.NewPostgreSQL(dbPort, dbUser, dbPassword, dbhost, "")
	err := db.Connect(ctx)
	if err != nil {
		log.Println(err)
	}
	defer db.Disconnect()

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

	// Проверяем, существует ли база данных
	checkDBExistsQuery := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)

	existsDB, err := db.QueryRow(ctx, checkDBExistsQuery)
	if err != nil {
		log.Fatalf("Ошибка проверки существует ли БД: %s\n", err)
	}

	if !existsDB {
		// SQL запрос для создания БД
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)

		// Выполняем запрос
		err = db.Exec(ctx, createDBQuery)
		if err != nil {
			log.Printf("Ошибка при создании БД: %v\n", err)
			logger.Printf("Ошибка при создании БД: %s %v\n", dbName, err)
		} else {
			logger.Printf("БД: %s успешно создана", dbName)
			log.Printf("БД: %s успешно создана\n", dbName)
			createTable()
		}
	} else {
		logger.Printf("БД: %s уже существует", dbName)
		log.Printf("БД: %s уже существует\n", dbName)
	}
}

func createTable() {

	ctx := context.Background()

	// Подключаемся к БД
	db := db.NewPostgreSQL(dbPort, dbUser, dbPassword, dbhost, dbName)
	err := db.Connect(ctx)
	if err != nil {
		log.Println(err)
	}
	defer db.Disconnect()

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
	);`

	err = db.Exec(ctx, createTableAndRow)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %v\n", err)
	} else {
		log.Printf("Таблица успешно создана\n")
	}
	defer filldb.FirstFillDB()
}
