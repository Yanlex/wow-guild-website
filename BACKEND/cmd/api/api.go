package api

import (
	"encoding/json"
	"fmt"
	rapi "kvd/internal/api/RaiderIoRequest"
	guild "kvd/internal/api/getAll"
	"kvd/internal/api/limits"
	d "kvd/internal/api/thumbnail"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

var (
	folderPath  string
	playerClass string
	homeDir     string
	err         error
)

// Функция создания папки в домашнем каталоге, так же нужна для определения пути до папки куда сохраняем аватарки игроков.
func makeNewFolder() {
	// Получаем домашний каталог
	homeDir, err = os.UserHomeDir()
	if err != nil {
		fmt.Println("Ошибка при получении домашнего каталога:", err)
	}

	// Название папки
	folderName := "assets/thumbnail"

	// Пусть к папке
	folderPath = filepath.Join(homeDir, folderName)
	playerClass = filepath.Join(homeDir, "assets/class")

	// Проверяем существует ли папка
	// os.Stat(foldierPath) проверяет есть ли папка и в err возвращает true если есть
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Папка не существует, создаем
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			fmt.Println("Ошибка при создании папки", err)
		}
		fmt.Println("Папка успешно создана", folderPath)
	} else {
		fmt.Println("Папка сущестует:", folderPath)
	}
}

// Функция получения имен всех игроков в гильдии
// Вызывает функцию скачиваниях аватарок
func getMembers(w http.ResponseWriter, r *http.Request) {
	guildMembers := rapi.GetAllMembers()

	guildMembersJson, err := json.Marshal(guildMembers)
	if err != nil {
		log.Println("Ошибка cериализации", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(guildMembersJson)
}

func getData(w http.ResponseWriter, r *http.Request) {
	guild := guild.GetAll()

	// Установка заголовка Content-Type
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(guild)
}

// Функция ограничивает доступ к ручке если в Query нету нужного параметра с паролем
// func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// Ожидаемый пароль в Query
// 		p := "kvd"
// 		// Ожидаемая переменная в Query - ?p=kvd
// 		reqPassword := r.URL.Query().Get("p")

// 		if reqPassword != p {
// 			// Здесь просто отдаем StatusNotFound чтобы не подсказывать про пароль.
// 			http.Error(w, "404", http.StatusNotFound)
// 			return
// 		}
// 		next(w, r)
// 	}
// }

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://wow-guild-front-nginx")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			return
		}

		// Если запрос разрешен, передаем его дальше.
		next.ServeHTTP(w, r)
	})
}

func Api() {
	//Пробуем создать папку
	makeNewFolder()

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			d.DownloadThumbnail(folderPath)
			time.Sleep(360 * time.Minute)
		}
	}()

	// Создаем лимитер с ограничением в 300 запросов в минуту.
	limiter := limits.NewLimiter(300)

	// Создаем новый ServeMux для маршрутизации HTTP запросов.
	mux := http.NewServeMux()

	// Папка с аватарками jpg
	playerAvatars := http.FileServer(http.Dir(folderPath))
	// Папка с классами jpg
	playerClass := http.FileServer(http.Dir(playerClass))
	/*
		РОУТЫ
	*/
	// Список игроков
	mux.HandleFunc("GET /api/get-members", getMembers)
	// API ручка, отдаем Rank, Name, Mythic Rating, Guild, Class
	mux.HandleFunc("GET /api/guild-data", getData)
	// Шарим папку с аватарками в WEB
	mux.Handle("/api/avatar/", http.StripPrefix("/api/avatar/", playerAvatars))
	// Шарим папку с классами в WEB
	mux.Handle("/api/class/", http.StripPrefix("/api/class/", playerClass))

	log.Println("Api сервер запущен на порте: 3000")

	// Настраиваем сервер с адресом и обработчиком, обернутым нашим middleware.
	server := &http.Server{
		Addr:    ":3000",
		Handler: enableCORS(limits.RateLimitMiddleware(mux, limiter, 300)),
	}

	log.Fatal(server.ListenAndServe())
}
