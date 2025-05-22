package db

import (
	"log"
	"os"

	"github.com/joho/godotenv" // Для загрузки .env
	"gorm.io/driver/postgres"  // Драйвер PostgreSQL для GORM
	"gorm.io/gorm"             // Основная библиотека GORM

	"effective-mobile/config" // Импортируем наш логгер (используем effective-mobile, как вы указали)
	"effective-mobile/models" // Импортируем нашу модель Person (используем effective-mobile, как вы указали)
)

// DB - это глобальный экземпляр подключения к базе данных GORM.
var DB *gorm.DB

// InitDB инициализирует подключение к базе данных и выполняет миграции.
func InitDB() {
	// 1. Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		// Если .env не найден или ошибка при загрузке, это критично.
		log.Fatalf("Error loading .env file: %v", err)
	}
	config.Log.Info(".env file loaded successfully") // Информационный лог

	// 2. Получение строки подключения к базе данных из переменной окружения
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Если переменная DATABASE_URL не установлена, это критично.
		log.Fatal("DATABASE_URL environment variable not set")
	}
	config.Log.Debugf("Connecting to database with URL: %s", databaseURL) // Отладочный лог

	// 3. Установка подключения к базе данных с помощью GORM
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		// Если не удалось подключиться, это критично.
		log.Fatalf("Failed to connect to database: %v", err)
	}
	config.Log.Info("Database connection established") // Информационный лог

	// 4. Выполнение автоматических миграций GORM
	err = DB.AutoMigrate(&models.Person{})
	if err != nil {
		// Если миграции не удалось выполнить, это критично.
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	config.Log.Info("Database migrations completed") // Информационный лог
}

// CloseDB закрывает соединение с базой данных.
func CloseDB() {
	sqlDB, err := DB.DB()
	if err != nil {
		config.Log.Errorf("Error getting underlying SQL DB: %v", err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		config.Log.Errorf("Error closing database connection: %v", err)
	}
	config.Log.Info("Database connection closed")
}