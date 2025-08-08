package routes

import (
	"database/sql"
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

		user, err := services.GetOrCreateUser(db, phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		switch {
		case strings.HasPrefix(body, "stock "):
			handleStockQuery(db, phone, user, strings.TrimPrefix(body, "stock "), c)
		case strings.HasPrefix(body, "alert "):
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
