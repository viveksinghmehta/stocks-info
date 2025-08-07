package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/psanford/finance-go/chart"
	"github.com/psanford/finance-go/quote"

	"stocks-info-channel/model"
)

// SearchStocks looks up company symbols or names
func SearchStocks(db *sql.DB, query string) ([]model.Stock, error) {
	rows, err := db.Query(`
		SELECT symbol, company_name FROM stocks
		WHERE LOWER(company_name) LIKE '%' || LOWER($1) || '%'
		OR LOWER(symbol) LIKE '%' || LOWER($1) || '%'
		LIMIT 10
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []model.Stock
	for rows.Next() {
		var stock model.Stock
		if err := rows.Scan(&stock.Symbol, &stock.CompanyName); err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}
	return stocks, nil
}

// GetYahooFinancePrice fetches the current price from Yahoo Finance
func GetYahooFinancePrice(symbol string) (float64, error) {
	q, err := quote.Get(symbol)
	if err != nil {
		return 0, err
	}
	return q.RegularMarketPrice, nil
}

// GetGrowthStats fetches historical data and calculates performance
func GetGrowthStats(symbol string) (model.StockGrowthStats, error) {
	now := time.Now()
	durations := []struct {
		Label string
		Start time.Time
	}{
		{"1M", now.AddDate(0, -1, 0)},
		{"1Y", now.AddDate(-1, 0, 0)},
		{"5Y", now.AddDate(-5, 0, 0)},
	}

	result := model.StockGrowthStats{
		Symbol:    symbol,
		Timestamp: now,
		Entries:   make(map[string]model.GrowthEntry),
	}

	current, err := GetYahooFinancePrice(symbol)
	if err != nil {
		return result, err
	}

	for _, d := range durations {
		params := &chart.Params{
			Symbol:   symbol,
			Start:    d.Start.Unix(),
			End:      now.Unix(),
			Interval: chart.Interval1d,
		}
		iter := chart.Get(params)
		var firstClose float64
		for iter.Next() {
			bar := iter.Bar()
			if bar.Close != 0 {
				firstClose = bar.Close
				break
			}
		}
		if firstClose == 0 {
			continue
		}

		percent := ((current - firstClose) / firstClose) * 100
		result.Entries[d.Label] = model.GrowthEntry{
			FromPrice: firstClose,
			ToPrice:   current,
			Growth:    percent,
		}
	}

	return result, nil
}

func FormatGrowthMessage(stats model.StockGrowthStats) string {
	var sb strings.Builder

	current := stats.Entries["1M"].ToPrice // use any entry to fetch current price safely

	sb.WriteString(fmt.Sprintf(
		"💼 *%s* Stock Update\n", strings.ToUpper(stats.Symbol),
	))
	sb.WriteString(fmt.Sprintf(
		"💰 *Current Price*: ₹%.2f\n", current,
	))
	sb.WriteString(fmt.Sprintf(
		"🕒 *As of*: %s\n\n", stats.Timestamp.Format("02 Jan 2006 15:04"),
	))

	sb.WriteString("📈 *Performance Overview:*\n")

	for _, label := range []string{"1M", "1Y", "5Y"} {
		entry, ok := stats.Entries[label]
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
