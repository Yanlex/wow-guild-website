package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	Connect() *pgxpool.Pool
	Disconnect() error
	Query(query string) ([]string, error)
	Exec(query string) error
}

type PostgreSQL struct {
	User     string
	Password string
	Port     string
	Host     string
	DBName   string
	pool     *pgxpool.Pool
}

func (db *PostgreSQL) Connect(ctx context.Context) error {
	var err error
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if db.pool != nil {
		fmt.Println("Не удалось создать пул подключений к Postgres или пул уже создан.")
		return err
	}

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.DBName)
	db.pool, err = pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Printf("Ошибка при попытке создать пул БД: %s\n", err)
		return err
	}

	err = db.pool.Ping(ctx)
	if err != nil {
		log.Printf("Ошибка при проверке соединения с БД: %s\n", err)
		return err
	} else {
		log.Printf("Подключение к PostgreSQL успешно установленно.")
	}
	return nil
}

func (db *PostgreSQL) Disconnect() error {
	if db.pool == nil {
		return fmt.Errorf("пул соединений не инициализирован")
	}

	db.pool.Close()
	log.Println("Пул PostgreSQL успешно отключен")
	db.pool = nil

	return nil
}

func (db *PostgreSQL) Query(ctx context.Context, query string) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if db.pool == nil {
		return nil, fmt.Errorf("пул соединений не инициализирован")
	}

	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить соединение из пула %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	return rows, nil
}

func (db *PostgreSQL) Exec(ctx context.Context, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Резервируем подключение к БД из пула
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить соединение из пула %w", err)
	}
	// Возвращаем подключение в пул
	defer conn.Release()

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса в БД: %w", err)
	}
	return nil
}

func (db *PostgreSQL) QueryRow(ctx context.Context, query string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var existsDB bool

	err := db.pool.QueryRow(ctx, query).Scan(&existsDB)
	if err != nil {
		if err == sql.ErrNoRows {
			return existsDB, nil
		}
		return existsDB, fmt.Errorf("ошибка при проверке существования базы данных: %w", err)
	}
	return existsDB, nil
}

func NewPostgreSQL(port, user, password, host, dbname string) *PostgreSQL {
	return &PostgreSQL{
		Port:     port,
		User:     user,
		Password: password,
		Host:     host,
		DBName:   dbname,
	}
}
