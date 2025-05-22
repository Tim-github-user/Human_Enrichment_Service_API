// main.go
package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"effective-mobile/config"
	"effective-mobile/db"
	"effective-mobile/docs" // Сгенерированный Swagger
	"effective-mobile/handlers"
)

// @title Human Enrichment Service API
// @version 1.0
// @description This is a human enrichment service API using Go and Gin.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Инициализация логгера
	config.InitLogger()
	config.Log.Info("Starting Human Enrichment Service")

	// Загрузка переменных окружения
	err := godotenv.Load()
	if err != nil {
		config.Log.Fatalf("Error loading .env file: %v", err)
	}

	// Инициализация базы данных
	db.InitDB()

	router := gin.Default()

	// Настройка Swagger
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	config.Log.Info("Swagger UI available at /swagger/index.html")

	// Группировка роутов
	v1 := router.Group("/api/v1")
	{
		people := v1.Group("/people")
		{
			people.GET("", handlers.GetPeople)
			people.GET("/:id", handlers.GetPersonByID)
			people.POST("", handlers.CreatePerson)
			people.PUT("/:id", handlers.UpdatePerson)
			people.DELETE("/:id", handlers.DeletePerson)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Порт по умолчанию
	}

	config.Log.Infof("Server starting on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		config.Log.Fatalf("Server failed to start: %v", err)
	}
}