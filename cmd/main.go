package main

import (
	"fmt"
	"stocks-info-channel/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	router := gin.Default()
	router.POST("whatsapp", routes.WhatsAppIncoming)

	router.Run()
}
