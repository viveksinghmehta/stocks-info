package helper

import (
	"fmt"
	"stocks-info-channel/model"
	"strings"
)

func WelcomeMessage() string {
	return `👋🏻 Welcome to *Stocks Info Channel*!

You can send:
• 🔍 *Stock HUL* — Get the latest HUL stock price
• ⭐ *Top Stocks* — Today's trending stocks *(coming soon 🚧)*
• 📢 *Alert NIFTY* — Set a stock price alert

Made with ❤️ in 🇮🇳`
}

func NoStockFoundMessage() string {
	return `❌ No results found.
🔍 Try full company name or stock symbol.`
}

func GenerateCompanyMessage(stocks []model.Stock) string {
	var sb strings.Builder

	sb.WriteString("📈 *Multiple companies matched your query:*\n\n")

	for i, s := range stocks {
		sb.WriteString(fmt.Sprintf("%d. *%s*\n    *(%s)*\n\n", i+1, s.CompanyName, s.Symbol))
	}

	sb.WriteString("🔁 Please reply with the *number* (e.g., 1 or 2) to choose.")

	return sb.String()
}

func SingleStockPerformanceMessage(stock model.StockPerformance) string {
	var sb strings.Builder

	// Daily change
	change := stock.Current - stock.Open
	var percentChange float64
	if stock.Open != 0 {
		percentChange = (change / stock.Open) * 100
	}

	var dailyTrendEmoji string
	switch {
	case change > 0:
		dailyTrendEmoji = "📈🟢"
	case change < 0:
		dailyTrendEmoji = "📉🔴"
	default:
		dailyTrendEmoji = "⏸️"
	}

	// Header
	sb.WriteString(fmt.Sprintf(
		"%s *%s (%s)*\n\n", dailyTrendEmoji, stock.CompanyName, stock.Symbol,
	))

	// Current & open price
	sb.WriteString(fmt.Sprintf(
		"💰 *Current Price*: ₹%.2f\n🔓 *Opened At*: ₹%.2f\n📊 *Today's Change*: ₹%.2f (%.2f%%)\n",
		stock.Current, stock.Open, change, percentChange,
	))

	// Timestamp
	sb.WriteString(fmt.Sprintf(
		"🕒 *As of*: %s\n\n", stock.Timestamp.Format("02 Jan 2006 03:04 PM"),
	))

	// Performance overview
	sb.WriteString("📈 *Performance Overview:*\n")

	for _, label := range []string{"1M", "1Y", "5Y"} {
		entry, ok := stock.Entries[label]
		if !ok {
			continue
		}

		emoji := "📈"
		trend := "gain"
		if entry.Growth < 0 {
			emoji = "📉"
			trend = "loss"
		} else if entry.Growth == 0 {
			emoji = "⚖️"
			trend = "no change"
		}

		comment := ""
		switch {
		case entry.Growth > 100:
			comment = "🚀 Massive rally!"
		case entry.Growth > 50:
			comment = "🔥 Strong performer!"
		case entry.Growth > 10:
			comment = "👍 Decent growth"
		case entry.Growth > 0:
			comment = "📊 Mild uptick"
		case entry.Growth > -10:
			comment = "🔻 Slight dip"
		case entry.Growth > -50:
			comment = "⚠️ Weak trend"
		default:
			comment = "💥 Major crash!"
		}

		sb.WriteString(fmt.Sprintf(
			"%s *%s*: %.2f%% %s (₹%.2f → ₹%.2f) %s\n",
			emoji, label, entry.Growth, trend, entry.FromPrice, entry.ToPrice, comment,
		))
	}

	sb.WriteString("\n📬 _This is an automated stock alert. Stay informed!_")

	return sb.String()
}

func AlertStockMessage(symbol string, price float64) string {
	return fmt.Sprintf("🔔 Alert: *%s*\nCurrent Price: ₹%.2f", symbol, price)
}
