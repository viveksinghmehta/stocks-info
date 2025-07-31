package routes

import (
	"database/sql"
	"net/http"
	"os"
	"stocks-info-channel/helper"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/twilio/twilio-go"
	openApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwillioWhatsappMessageRequest struct {
	MessageSid          string `json:"MessageSid" form:"MessageSid"`
	SmsSid              string `json:"SmsSid" form:"SmsSid"`
	SmsMessageSid       string `json:"SmsMessageSid" form:"SmsMessageSid"`
	AccountSid          string `json:"AccountSid" form:"AccountSid"`
	MessagingServiceSid string `json:"MessagingServiceSid" form:"MessagingServiceSid"`
	From                string `json:"From" form:"From"`
	To                  string `json:"To" form:"To"`
	Body                string `json:"Body" form:"Body"`
}

type User struct {
	ID                      string
	PhoneNumber             string
	Name                    sql.NullString
	LastMessageTime         sql.NullTime
	LastTwoMessagesToUser   pq.StringArray
	LastTwoMessagesFromUser pq.StringArray
	IsSubscribed            bool
	SubscribedStocks        pq.StringArray
}

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

func GetUserByPhone(db *sql.DB, phoneNumber string) (*User, error) {
	var user User
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

func WhatsAppIncoming(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var message TwillioWhatsappMessageRequest

		// 1. Receive the message and decode it
		if err := c.Bind(&message); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Could not decode the values",
			})
			return
		}

		// Check if the user exists
		user, error := GetUserByPhone(db, message.From)
		if error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}
		if user == nil {
			// User not found, insert new user with timestamp
			err := InsertUser(db, message.From)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Could not save new user"})
				return
			}
		} else {
			// User exists, update last_message_time
			err := UpdateLastMessageTime(db, message.From)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Could not update timestamp"})
				return
			}
		}
		// 2 - 1. if new user save it the DB
		// 2 - 2. ask for the name of the user and save it against its phone number

		messageBody := `👋🏻 Hello and welcome to the Stocks Info Channel!

📈 You can try:

• 🔍 Search stock price: send "stock <ticker>" (e.g., stock AAPL)
• ⭐ Get top stocks: send "top stocks"
• 📢 Index price alert: send "alert <index>" (e.g., alert NASDAQ)

💬 Feel free to ask me anything about stocks, and I'll help you out!


Made with ❤️ in 🇮🇳`

		phone := strings.TrimPrefix(message.From, helper.AppConstant().WhatsApp)

		response, error := twillioClient(phone, messageBody)
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
