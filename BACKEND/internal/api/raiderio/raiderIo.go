package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func init() {
	// config.InitConfigDB()
}

func tryFetchRio(url string) (*http.Response, error) {
	for {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		log.Println("Failed to fetch player data from API, trying again in 5 minutes", url, err)
		time.Sleep(5 * time.Minute)
	}
}

func FetchRaiderIo() string {

	// URL по кторому получаем данные
	// КВД https://raider.io/api/v1/guilds/profile?region=eu&realm=howling-fjord&name=%D0%9A%D0%BB%D1%8E%D1%87%D0%B8%D0%BA%20%D0%B2%20%D0%B4%D1%83%D1%80%D0%BA%D1%83&fields=members
	guildRegion := os.Getenv("GUILD_REGION")
	guildRealm := os.Getenv("GUILD_REALM")
	guildName := os.Getenv("GUILD_NAME")
	encodeGuildName := url.QueryEscape(guildName)
	encodeGuildRealm := url.QueryEscape(guildRealm)

	url := fmt.Sprintf(`https://raider.io/api/v1/guilds/profile?region=%s&realm=%s&name=%s&fields=members`, guildRegion, encodeGuildRealm, encodeGuildName)
	log.Printf("Make API request to %s\n", url)
	// url := viper.GetString("guild.raiderio_api_url")

	// Гет запрос
	resp, err := tryFetchRio(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	// Читаем данные
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// Преобразование в строку
	bodyStr := string(body)
	// fmt.Println(bodyStr)

	// // Создать файл
	// file, err := os.Create("fetch.json")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer file.Close()

	// // Копируем все данные в файл
	// _, err = io.Copy(file, resp.Body)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Читаем данные из файла
	// body, err := os.ReadFile("fetch.json")

	// result := gjson.Get(bodyStr, "members.#(character.name==\"Коррозийный\").character")
	// sd := gjson.Get(bodyStr, "members.#(character.name==\"Коррозийный\").character.profile_url")
	// fmt.Println(sd.String())

	return bodyStr
	// Блокировка выполнения программы, чтобы она не завершалась
}
