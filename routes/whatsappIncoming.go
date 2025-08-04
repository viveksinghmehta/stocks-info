package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"stocks-info-channel/helper"
	"stocks-info-channel/model"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/twilio/twilio-go"
	openApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func twillioClient(phone, message string) (*openApi.ApiV2010Message, error) {
	sid := os.Getenv(helper.EnvironmentConstant().TWILIO_ACCOUNT_SID)
	password := os.Getenv(helper.EnvironmentConstant().TWILIO_AUTH_TOKEN)

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: sid,
		Password: password,
	})

	params := &openApi.CreateMessageParams{}
	params.SetFrom(helper.AppConstant().WhatsApp + os.Getenv(helper.EnvironmentConstant().PHONE_NUMBER)) // Twilio whatsapp phone number

	params.SetTo(helper.AppConstant().WhatsApp + phone) // Recipient's phone number
	params.SetBody(message)

	return client.Api.CreateMessage(params)
}

func GetUserByPhone(db *sql.DB, phoneNumber string) (*model.User, error) {
	var user model.User
	query := `
        SELECT
            id,
            phone_number,
            name,
            last_message_time,
            last_two_messages_to_user,
            last_two_messages_from_user,
            is_subscribed,
            subscribed_stocks
        FROM users WHERE phone_number = $1
    `
	err := db.QueryRow(query, phoneNumber).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.LastMessageTime,
		(*pq.StringArray)(&user.LastTwoMessagesToUser),
		(*pq.StringArray)(&user.LastTwoMessagesFromUser),
		&user.IsSubscribed,
		(*pq.StringArray)(&user.SubscribedStocks),
	)
	if err == sql.ErrNoRows {
		// User not found
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func InsertUser(db *sql.DB, phoneNumber string) error {
	insertQuery := `
        INSERT INTO users (phone_number, last_message_time) VALUES ($1, NOW())
    `
	_, err := db.Exec(insertQuery, phoneNumber)
	return err
}

func UpdateLastMessageTime(db *sql.DB, phoneNumber string) error {
	updateQuery := `
        UPDATE users
        SET last_message_time = NOW()
        WHERE phone_number = $1
    `
	_, err := db.Exec(updateQuery, phoneNumber)
	return err
}

// SearchStocks finds stocks by symbol or company name based on user input logic.
func SearchStocks(db *sql.DB, stockName string) ([]model.Stock, error) {
	var results []model.Stock

	if strings.Contains(stockName, " ") {
		// Space in query: direct company name search
		likePattern := "%" + stockName + "%"
		rows, err := db.Query(
			`SELECT symbol, company_name FROM stocks WHERE company_name ILIKE $1`,
			likePattern,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var s model.Stock
			if err := rows.Scan(&s.Symbol, &s.CompanyName); err == nil {
				results = append(results, s)
			}
		}
		return results, nil
	}

	// No space: try exact symbol, then fallback to company name
	var direct model.Stock
	err := db.QueryRow(
		`SELECT symbol, company_name FROM stocks WHERE symbol ILIKE $1 LIMIT 1`,
		stockName,
	).Scan(&direct.Symbol, &direct.CompanyName)
	if err == nil {
		results = append(results, direct)
		return results, nil
	}

	likePattern := "%" + stockName + "%"
	rows, err := db.Query(
		`SELECT symbol, company_name FROM stocks WHERE company_name ILIKE $1`,
		likePattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s model.Stock
		if err := rows.Scan(&s.Symbol, &s.CompanyName); err == nil {
			results = append(results, s)
		}
	}
	return results, nil
}

// Generates a WhatsApp message listing companies with emojis
func GenerateCompanyMessage(companies []model.Stock) string {
	var sb strings.Builder
	sb.WriteString("🤔 Are you trying to search for one of these companies?\n")
	sb.WriteString("📋 Please use either the full company name or its symbol when making your request. Here are the options:\n\n")
	for _, c := range companies {
		sb.WriteString(fmt.Sprintf("🔹 %s (%s)\n", c.CompanyName, c.Symbol))
	}
	sb.WriteString("\n💡 For example: try searching \"HINDUNILVR\" or \"Hindustan Unilever Limited\".\n")
	sb.WriteString("❓ Let me know which company you’re interested in!")
	return sb.String()
}

func WhatsAppIncoming(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var message model.TwillioWhatsappMessageRequest

		// 1. Receive the message and decode it
		if err := c.Bind(&message); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Could not decode the values",
			})
			return
		}

		phoneNumber := strings.TrimPrefix(message.From, helper.AppConstant().WhatsApp)

		// Check if the user exists
		user, error := GetUserByPhone(db, phoneNumber)
		if error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}
		if user == nil {
			// User not found, insert new user with timestamp
			err := InsertUser(db, phoneNumber)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Could not save new user"})
				return
			}
		} else {
			// User exists, update last_message_time
			err := UpdateLastMessageTime(db, phoneNumber)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Could not update timestamp"})
				return
			}
		}

		userMessage := strings.ToLower(message.Body)

		// Check if the user has send a normal message or from the options
		switch {
		case strings.HasPrefix(userMessage, "stock "):
			// Reply with stock price logic
			companyName := strings.TrimPrefix(userMessage, "stock ")

			matches, error := SearchStocks(db, companyName)
			if error != nil {
				// handle DB error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": error.Error(),
				})
			}
			if len(matches) == 0 {
				// reply: "No company found"
				messageBody := `❌ No results found. 🤔
Did you mean something else? Try entering the full company name or stock symbol. 🔍`
				response, error := twillioClient(phoneNumber, messageBody)
				if error != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"status_code": http.StatusInternalServerError,
						"message":     error.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"OK":       "OK",
					"response": response,
				})
				return
			} else if len(matches) == 1 {
				// reply: match found, show symbol and name

				return
			} else {
				// reply: show all matches, let user pick
				messageBody := GenerateCompanyMessage(matches)
				response, error := twillioClient(phoneNumber, messageBody)
				if error != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"status_code": http.StatusInternalServerError,
						"message":     error.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"OK":       "OK",
					"response": response,
				})
				return
			}
		case userMessage == "top stocks":
			// Reply with top stocks logic
			c.JSON(http.StatusOK, gin.H{
				"OK":      "OK",
				"Message": message.Body,
			})
		case strings.HasPrefix(userMessage, "alert "):
			// Reply with alert logic
			c.JSON(http.StatusOK, gin.H{
				"OK":      "OK",
				"Message": message.Body,
			})
		default:
			messageBody := `👋🏻 Welcome to the Stocks Info Channel!
Send one of these messages:
• 🔍 Stock HUL — for HUL stock price
• ⭐ top stocks — to get today's top stocks
• 📢 alert NIFTY 50 — for index price alerts

You can get alerts and info for *any* Indian stocks!

💬 Try them now, or ask me anything about stocks.

---
Made with ❤️ in 🇮🇳 by Vivek Singh Mehta`

			response, error := twillioClient(phoneNumber, messageBody)
			if error != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status_code": http.StatusInternalServerError,
					"message":     error.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"OK":       "OK",
				"response": response,
			})
		}
	}
}
