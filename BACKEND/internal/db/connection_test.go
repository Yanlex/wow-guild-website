package db

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestNewPostgreSQL(t *testing.T) {
	testCases := []struct {
		name     string
		port     string
		user     string
		password string
		host     string
		dbname   string
	}{
		{
			name:     "success case 1",
			port:     "5432",
			user:     "postgres",
			password: "secret",
			host:     "localhost",
			dbname:   "mydb",
		}, {
			name:     "success case 2",
			port:     "5433",
			user:     "admin",
			password: "admin123",
			host:     "127.0.0.1",
			dbname:   "testdb",
		}, // Тест для минимального порта
		{
			name:     "min port",
			port:     "1",
			user:     "postgres",
			password: "secret",
			host:     "localhost",
			dbname:   "mydb",
		},
		// Тест для максимального порта
		{
			name:     "max port",
			port:     "65535", // Максимальный порт TCP/IP
			user:     "postgres",
			password: "secret",
			host:     "localhost",
			dbname:   "mydb",
		},
		// Тест для пустого пароля
		{
			name:     "empty password",
			port:     "5432",
			user:     "postgres",
			password: "",
			host:     "localhost",
			dbname:   "mydb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем объект PostgreSQL
			pg := NewPostgreSQL(tc.port, tc.user, tc.password, tc.host, tc.dbname)

			// Проверяем, что значения совпали
			if pg.Port != tc.port {
				t.Errorf("Port %v != expected %v", pg.Port, tc.port)
			}
			if pg.User != tc.user {
				t.Errorf("User %v != expected %v", pg.User, tc.user)
			}
			if pg.Password != tc.password {
				t.Errorf("Password %v != expected %v", pg.Password, tc.password)
			}
			if pg.Host != tc.host {
				t.Errorf("Host %v != expected %v", pg.Host, tc.host)
			}
			if pg.DBName != tc.dbname {
				t.Errorf("DBName %v != expected %v", pg.DBName, tc.dbname)
			}
		})
	}
}

// Дополнительный тест с использованием quick.Check для автоматической проверки случайных значений
func TestNewPostgreSQLQuick(t *testing.T) {
	quick.Check(func(port string, user, password, host, dbname string) bool {
		pg := NewPostgreSQL(port, user, password, host, dbname)
		return reflect.DeepEqual(pg, &PostgreSQL{
			Port:     port,
			User:     user,
			Password: password,
			Host:     host,
			DBName:   dbname,
		})
	}, nil)
}
