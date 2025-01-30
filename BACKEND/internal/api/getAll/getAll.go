package getAll

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

var (
	digitRegex = regexp.MustCompile(`\d`) // Компилируем регулярное выражение один раз

	host        = os.Getenv("DB_ADDRESS")
	port        = os.Getenv("HOST_DB_PORT")
	dbUser      = os.Getenv("DB_USER")
	dbPassword  = os.Getenv("DB_PASS")
	guildDBName = os.Getenv("DB_NAME")
)

func GetAll() []byte {

	type Player struct {
		Rank                         int    `json:"rank"`
		Name                         string `json:"name"`
		Mythic_plus_scores_by_season int    `json:"mythic_plus_scores_by_season"`
		Guild                        string `json:"guild"`
		Class                        string `json:"class"`
	}

	// Формируем строку подключения
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable", host, port, dbUser, dbPassword, guildDBName)

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println("Ошибка подключения к БД", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Println("Ошибка пинга к БД", err)
	} else {
		log.Println("Успешно подключились к базе данных")
	}

	// Выполняем запрос
	rows, err := db.Query("SELECT rank, name, mythic_plus_scores_by_season, guild, class FROM members")
	if err != nil {
		log.Println("Ошибка выполнения запроса к БД", err)
	}
	defer rows.Close()

	var players []Player

	// Обрабатываем результаты запроса
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.Rank, &player.Name, &player.Mythic_plus_scores_by_season, &player.Guild, &player.Class)
		if err != nil {
			log.Println("Ошибка обработчика запроса в БД", err)
		}

		// Исключаем ники с цифрами, их невозможно использовать в запросе к сторонним апи.
		match := digitRegex.MatchString(player.Name)
		if !match {
			players = append(players, player)
		}
	}

	// Проверяем наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		log.Println("Ошибка при обработке результата из БД", err)
	}

	// Преобразование структуры в JSON
	jsonPlayers, err := json.Marshal(players)
	if err != nil {
		log.Println("Ошибка сериализации", err)
	}

	return jsonPlayers
}
