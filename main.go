package main

import (
	"log"
	"restful-api-gin/config"
	_ "restful-api-gin/docs"
	"restful-api-gin/user"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title restful-api-gin
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	config.InitializeMongoDB()

	gin := gin.Default()
	user.RegisterRoutes(gin)
	gin.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := gin.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
