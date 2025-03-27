package thumbnail

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	digitRegex = regexp.MustCompile(`\d`) // Компилируем регулярное выражение один раз

	host        = os.Getenv("DB_ADDRESS")
	port        = os.Getenv("HOST_DB_PORT")
	dbUser      = os.Getenv("DB_USER")
	dbPassword  = os.Getenv("DB_PASS")
	guildDBName = os.Getenv("DB_NAME")
)

func DownloadThumbnail(foldierPath string) {

	var imgCount int
	var newImgCount int

	type Player struct {
		name          string
		thumbnail_url string
	}

	// Формируем строку подключения
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPassword, guildDBName)

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println("Ошибка подключения к БД", err)
		return
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Println("Ошибка проверке подключения к БД", err)
		return
	}

	// Выполняем запрос
	rows, err := db.Query("SELECT name, thumbnail_url FROM members")
	if err != nil {
		log.Println("Ошибка выполнения запроса к БД", err)
		return
	}
	defer rows.Close()

	players := make(map[string]string)

	// Обрабатываем результаты запроса
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.name, &player.thumbnail_url)
		if err != nil {
			log.Println("Ошибка обработчика запроса в БД", err)
			return
		}

		// Исключаем ники с цифрами, их невозможно использовать в запросе к сторонним апи.
		match := digitRegex.MatchString(players[player.name])
		if !match {
			players[player.name] = player.thumbnail_url
		}
	}

	// Проверяем наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		log.Println("Ошибка при обработке результата из БД", err)
		return
	}

	for player, thumb := range players {
		if thumb == "" {

		} else {

			fileName := fmt.Sprintf("%s", player+".jpg")

			filePath := filepath.Join(foldierPath, fileName)

			// Проверяем, существует ли файл
			if _, err := os.Stat(filePath); err == nil {
				// fmt.Printf("Файл %s уже существует.\n", fileName)
				imgCount++
				continue
			} else if os.IsNotExist(err) {

				url := thumb

				resp, err := http.Get(url)
				if err != nil {
					log.Println("Ошибка получения картинки", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Println("Ошибка статуса", resp.Status)
				}

				// Создаем файл для записи
				file, err := os.Create(filePath)
				if err != nil {
					fmt.Println("Ошибка при создании файла:", err)
					return
				}
				defer file.Close()

				// Записываем содержимое ответа в файл
				_, err = io.Copy(file, resp.Body)
				if err != nil {
					fmt.Println("Ошибка при записи файла:", err)
					return
				}
				newImgCount++
			} else {
				// Другая ошибка
				fmt.Println("Ошибка при проверке существования файла:", err)
				return
			}

			time.Sleep(5 * time.Second)
		}
	}

	log.Printf("Аватарок в папке: %d, новых: %d", imgCount, newImgCount)

}
