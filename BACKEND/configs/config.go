package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func InitConfigDB() {
	// Получаем домашний каталог пользователя
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Println("Ошибка при получении домашнего каталога:", err)
	// 	return
	// }

	// Относительный путь от домашнего каталога до интересующей нас директории
	// relativePath := ".myFolder/subFolder"

	// Объединяем домашний каталог и относительный путь
	// fullPath := filepath.Join(homeDir, relativePath)

	viper.SetConfigName("db")            // Имя файла конфигурации без расширения
	viper.SetConfigType("yaml")          // Тип файла конфигурации
	viper.AddConfigPath("/app/configs/") // Путь к директории с конфигурацией
	// viper.SetConfigFile("./db.yaml") // Путь к конфигурационному файла
	// Чтение конфигурации
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}
}
