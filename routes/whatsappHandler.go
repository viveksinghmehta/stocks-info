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
		println("### USER :- ", user)
		// if user == nil {
		// 	_ = services.InsertUser(db, phone)
		// } else {
		// 	_ = services.UpdateLastMessageTime(db, phone)
		// }

		switch {
		case strings.HasPrefix(body, "stock "):
			println("### Stock search :- ", body)
			handleStockQuery(db, phone, strings.TrimPrefix(body, "stock "), c)
		case strings.HasPrefix(body, "alert "):
			println("### alert :- ", body)
			handleStockAlerts(db, phone, strings.TrimPrefix(body, "alert "), c)
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

func handleStockQuery(db *sql.DB, phone, query string, c *gin.Context) {
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
		price, err := services.GetYahooFinancePrice(matches[0].Symbol + ".NS")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price"})
			return
		}
		msg := helper.SingleStockMessage(matches[0], price)
		services.SendWhatsApp(phone, msg)
	default: // multiple company found with stock name
		msg := helper.GenerateCompanyMessage(matches)
		services.SendWhatsApp(phone, msg)
	}
	c.JSON(http.StatusOK, gin.H{"status": "Stock response sent"})
}

func handleStockAlerts(db *sql.DB, phone, alert string, c *gin.Context) {
	stocks := strings.Fields(alert)

	for _, symbol := range stocks {
		go func(s string) {
			price, err := services.GetYahooFinancePrice(s + ".NS")
			if err != nil {
				services.SendWhatsApp(phone, "⚠️ Error fetching "+s)
				return
			}
			msg := helper.AlertStockMessage(s, price)
			services.SendWhatsApp(phone, msg)
		}(symbol)
	}

	c.JSON(http.StatusOK, gin.H{"status": "Alert messages dispatched"})
}
