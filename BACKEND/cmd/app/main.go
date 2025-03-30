package main

import (
	"fmt"
	a "kvd/cmd/api"
	deploy "kvd/deployments/db"
	"kvd/internal/db/update"

	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
)

func updatePlayersHandler() {
	fmt.Println("Задача CRON запущена")
	update.UpdateAllPlayers()
}

func init() {
	// Крон планировщик
	// Загрузка локации
	est, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		// Логирование ошибки вместо паники
		log.Printf("Ошибка загрузки региона: %v", err)
		return
	}

	// Создание планировщика с указанной локацией
	s := gocron.NewScheduler(est)

	// Планирование задачи
	_, _ = s.Every(1).Day().At("02:30").Do(updatePlayersHandler)

	// Запуск планировщика асинхронно
	s.StartAsync()
}

// var err error

func main() {
	go a.Api()
	go func() {
		time.Sleep(10 * time.Minute)
		update.UpdateAllPlayers()
		time.Sleep(12 * time.Hour)
	}()
	// Создаем канал для сигналов
	signals := make(chan os.Signal, 1)
	// Регистрируем канал для получения сигналов
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	deploy.Deploy()
	time.Sleep(2 * time.Second)
	log.Println("Backend запущен")
	time.Sleep(2 * time.Second)
	// Блокируемся до получения сигнала
	sig := <-signals
	fmt.Println("Получен сигнал, закрываем программу:", sig)
}
