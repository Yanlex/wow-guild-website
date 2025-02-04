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
		log.Println("Ошибка при запросе к API, повторная попытка через 5 минут", url, err)
		time.Sleep(5 * time.Minute)
	}
}

// Ошибку возвращаем просто для практики.
func MemberRio(guildRegion, encodedPlayerRealm, encodedName string) (string, error) {
	url := fmt.Sprintf("https://raider.io/api/v1/characters/profile?region=%s&realm=%s&name=%s&fields=mythic_plus_scores_by_season:current", guildRegion, encodedPlayerRealm, encodedName)
	for {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			// Читаем данные из запроса
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
			}

			// Преобразование в строку
			playerResp := string(body)
			if playerResp == "" {
				log.Println("Не удалось преобразовать в строку", playerResp)
			}

			return playerResp, nil
		} else {
			defer resp.Body.Close()
		}
		log.Println("Ошибка при запросе к API, повторная попытка через 5 минут", url, err)
		time.Sleep(5 * time.Minute)
	}
}

func GuildRio() string {
	// КВД https://raider.io/api/v1/guilds/profile?region=eu&realm=howling-fjord&name=%D0%9A%D0%BB%D1%8E%D1%87%D0%B8%D0%BA%20%D0%B2%20%D0%B4%D1%83%D1%80%D0%BA%D1%83&fields=members
	guildRegion := os.Getenv("GUILD_REGION")
	guildRealm := os.Getenv("GUILD_REALM")
	guildName := os.Getenv("GUILD_NAME")
	encodeGuildName := url.QueryEscape(guildName)
	encodeGuildRealm := url.QueryEscape(guildRealm)

	url := fmt.Sprintf(`https://raider.io/api/v1/guilds/profile?region=%s&realm=%s&name=%s&fields=members`, guildRegion, encodeGuildRealm, encodeGuildName)

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
	return bodyStr
	// Блокировка выполнения программы, чтобы она не завершалась
}
