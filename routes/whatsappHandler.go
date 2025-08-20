package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"stocks-info-channel/helper"
	"stocks-info-channel/model"
	"stocks-info-channel/services"

	"github.com/gin-gonic/gin"
)

func WhatsAppIncomingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var message model.TwillioWhatsappMessageRequest
		if err := c.Bind(&message); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		phone := strings.TrimPrefix(message.From, helper.AppConstant().WhatsApp)
		body := strings.ToLower(message.Body)

		log.Println("Message from :- ", message.From)
		log.Println("Message To :- ", message.To)
		log.Println("Message Body :- ", message.Body)
		log.Println("Message AccountSid :- ", message.AccountSid)
		log.Println("Message MessageSid :- ", message.MessageSid)
		log.Println("Message SmsSid :- ", message.SmsSid)
		log.Println("Message SmsMessageSid :- ", message.SmsMessageSid)

		user, err := services.GetOrCreateUser(db, phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		switch {
		case strings.HasPrefix(body, "stock "):
			log.Println("Handling Stock search query...")
			handleStockQuery(db, phone, user, strings.TrimPrefix(body, "stock "), c)
		case strings.HasPrefix(body, "alert "):
			log.Println("Handling Stock alert query...")
			handleStockAlerts(db, phone, user, strings.TrimPrefix(body, "alert "), c)
		case body == "top stocks":
			// TODO: implement top stocks logic
			c.JSON(http.StatusOK, gin.H{"msg": "Coming soon!"})
		default:
			resp := helper.WelcomeMessage()
			_, _ = services.SendWhatsApp(phone, resp)
			c.JSON(http.StatusOK, gin.H{"message": "Default welcome sent"})
		}
	}
}

func handleStockQuery(db *sql.DB, phone string, user *model.User, query string, c *gin.Context) {
	matches, err := services.SearchStocks(db, query)
	log.Println("Matches :- ", matches)
	if err != nil {
		log.Println("Error :- ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch len(matches) {
	case 0: // No stock found
		log.Println("No stock found...")

		userHasCheckFor2Times, error := services.CheckForTwoStockSeachTries(db, user)
		if error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Could not save the message in DB",
				"error":  error.Error(),
			})
		}
		if userHasCheckFor2Times {
			msg := helper.StockNotInDatabaseMessage()
			err := services.ClearLastTwoMessages(db, user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Could not clear out the last 2 messages sent to user from DB",
					"error":   err.Error(),
				})
			}
			services.SendWhatsApp(phone, msg)
			c.JSON(http.StatusOK, gin.H{
				"message": "Could not clear out the last 2 messages sent to user from DB",
				"error":   err.Error(),
			})
			return
		} else {
			msg := helper.NoStockFoundMessage()
			err := services.UpdateSentMessagesToUser(db, user, msg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Could not save the message in DB",
					"error":   err.Error(),
				})
				return
			}
			services.SendWhatsApp(phone, msg)
			return
		}
	case 1: // exact match found for the stock
		log.Println(" Stock Symbol :- ", matches[0].Symbol)
		log.Println(" Company Name :- ", matches[0].CompanyName)
		stockPerformance, err := services.GetStockPerformance(matches[0].Symbol+".NS", matches[0].CompanyName)
		log.Println("Stock Performance :- ", stockPerformance)
		if err != nil {
			log.Println("Failed to fetch stock price...")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price"})
			return
		}
		msg := helper.SingleStockPerformanceMessage(stockPerformance)
		services.SendWhatsApp(phone, msg)
	default: // multiple company found with stock name
		msg := helper.GenerateCompanyMessage(matches)
		err := services.UpdateSentMessagesToUser(db, user, msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Could not save the message in DB",
				"error":   err.Error(),
			})
		}
		services.SendWhatsApp(phone, msg)
	}
	c.JSON(http.StatusOK, gin.H{"status": "Stock response sent"})
}

func handleStockAlerts(db *sql.DB, phone string, user *model.User, query string, c *gin.Context) {
	matches, err := services.SearchStocks(db, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	switch len(matches) {
	case 0: // No stock font
		msg := helper.NoStockFoundMessage()
		services.SendWhatsApp(phone, msg)
	case 1: // exact match found for the stock
		stockPerformance, err := services.GetStockPerformance(matches[0].Symbol+".NS", matches[0].CompanyName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price"})
			return
		}
		msg := helper.SingleStockPerformanceMessage(stockPerformance)
		services.SendWhatsApp(phone, msg)
	default: // multiple company found with stock name
		msg := helper.GenerateCompanyMessage(matches)
		err := services.UpdateSentMessagesToUser(db, user, msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Could not save the message in DB",
				"error":  err.Error(),
			})
		}
		services.SendWhatsApp(phone, msg)
	}

	c.JSON(http.StatusOK, gin.H{"status": "Alert messages dispatched"})
}
