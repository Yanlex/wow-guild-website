package raideriorequest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type getAll struct {
	Name          string    `json:"name"`
	Faction       string    `json:"faction"`
	Region        string    `json:"region"`
	Realm         string    `json:"realm"`
	LastCrawledAt time.Time `json:"last_crawled_at"`
	ProfileURL    string    `json:"profile_url"`
	Members       []struct {
		Rank      int `json:"rank"`
		Character struct {
			Name              string    `json:"name"`
			Race              string    `json:"race"`
			Class             string    `json:"class"`
			ActiveSpecName    string    `json:"active_spec_name"`
			ActiveSpecRole    string    `json:"active_spec_role"`
			Gender            string    `json:"gender"`
			Faction           string    `json:"faction"`
			AchievementPoints int       `json:"achievement_points"`
			Region            string    `json:"region"`
			Realm             string    `json:"realm"`
			LastCrawledAt     time.Time `json:"last_crawled_at"`
			ProfileURL        string    `json:"profile_url"`
			ProfileBanner     string    `json:"profile_banner"`
		} `json:"character"`
	} `json:"members"`
}

var (
	guildRegion      = os.Getenv("GUILD_REGION")
	guildRealm       = os.Getenv("GUILD_REALM")
	guildName        = os.Getenv("GUILD_NAME")
	encodeGuildName  = url.QueryEscape(guildName)
	encodeGuildRealm = url.QueryEscape(guildRealm)
)

/*
Делаем запрос к АПИ Raider.io.
Читаем JSON в переменную body
Десериализуем JSON в дату
Парсим дату, забираем членов гильдии в срез members
*/
func GetAllMembers() []string {
	var members []string

	url := fmt.Sprintf("https://raider.io/api/v1/guilds/profile?region=%s&realm=%s&name=%s&fields=members", guildRegion, encodeGuildRealm, encodeGuildName)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Ошибка при запросе к API Raider.io", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения тела запроса", err)
	}

	var data getAll
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Ошибка десериализации JSON", err)
	}

	for _, member := range data.Members {
		members = append(members, member.Character.Name)
	}
	return members
}
