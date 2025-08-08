package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"stocks-info-channel/helper"
	"stocks-info-channel/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
}

func connectTODB() *sql.DB {
	connStr := os.Getenv(helper.EnvironmentConstant().DB_URL)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Check connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to PostgreSQL successfully!")
	return db
}

func main() {
	db := connectTODB()
	router := gin.Default()

	router.POST("whatsapp", routes.WhatsAppIncomingHandler(db))
	router.GET("alert", routes.StockAlertHandler(db))

	port := os.Getenv(helper.EnvironmentConstant().PORT)
	if port == "" {
		port = helper.AppConstant().DefaultPort
	}
	router.Run(":" + port)
}
