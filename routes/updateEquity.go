package routes

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func UpdateEquity(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open the CSV file
		file, err := os.Open("EQUITY_L.csv")
		if err != nil {
			log.Fatal("Failed to open CSV file:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}
		defer file.Close()

		// Read the CSV file
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			log.Fatal("Failed to read CSV file:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}

		// If your first row is the header, skip it
		start := 1 // skip header row
		if records[0][0] != "symbol" || records[0][1] != "company_name" {
			start = 0
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatal("Failed to start transaction:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}

		stmt, err := tx.Prepare(`
        INSERT INTO stocks (symbol, company_name)
        VALUES ($1, $2)
        ON CONFLICT(symbol) DO UPDATE SET company_name = EXCLUDED.company_name;
    `)
		if err != nil {
			log.Fatal("Failed to prepare statement:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}
		defer stmt.Close()

		for i := start; i < len(records); i++ {
			symbol := records[i][0]
			name := records[i][1]
			_, err := stmt.Exec(symbol, name)
			if err != nil {
				log.Printf("Failed to upsert [%s, %s]: %v\n", symbol, name, err)
				c.JSON(500, gin.H{
					"error": err.Error(),
				})
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatal("Failed to commit transaction:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		}

		fmt.Println("Bulk upsert completed successfully!")
		c.JSON(200, gin.H{
			"success": "Bulk upsert completed successfully!",
		})
	}
}
