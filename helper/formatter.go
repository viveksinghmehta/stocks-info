package helper

import (
	"fmt"
	"stocks-info-channel/model"
	"strings"
)

func WelcomeMessage() string {
	return `👋🏻 Welcome to *Stocks Info Channel*!

Send one of these:
• 🔍 *Stock HUL* — HUL stock price
• ⭐ *Top Stocks* — Today's trending stocks
• 📢 *Alert NIFTY TCS INFY* — Get prices individually

Made with ❤️ in 🇮🇳`
}

func NoStockFoundMessage() string {
	return `❌ No results found.
🔍 Try full company name or stock symbol.`
}

func GenerateCompanyMessage(stocks []model.Stock) string {
	var sb strings.Builder
	sb.WriteString("🔍 Found multiple results:\n")
	for _, s := range stocks {
		sb.WriteString(fmt.Sprintf("🔹 *%s* (%s)\n", s.CompanyName, s.Symbol))
	}
	return sb.String()
}

func SingleStockMessage(stock model.Stock, price float64) string {
	return fmt.Sprintf("📈 *%s (%s)*\nCurrent Price: ₹%.2f", stock.CompanyName, stock.Symbol, price)
}

func AlertStockMessage(symbol string, price float64) string {
	return fmt.Sprintf("🔔 Alert: *%s*\nCurrent Price: ₹%.2f", symbol, price)
}
