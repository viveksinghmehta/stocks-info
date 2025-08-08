package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"stocks-info-channel/helper"
	"stocks-info-channel/model"
)

// SearchStocks looks up company symbols or names
func SearchStocks(db *sql.DB, query string) ([]model.Stock, error) {
	hasSpace := strings.Contains(query, " ")

	var rows *sql.Rows
	var err error

	if hasSpace {
		// User is likely searching for a company name
		rows, err = db.Query(`
			SELECT symbol, company_name FROM stocks
			WHERE LOWER(company_name) LIKE '%' || LOWER($1) || '%'
			LIMIT 10
		`, query)
	} else {
		// User is likely searching for a symbol
		// First try exact match
		rows, err = db.Query(`
			SELECT symbol, company_name FROM stocks
			WHERE LOWER(symbol) = LOWER($1)
			LIMIT 1
		`, query)

		if err == nil && rows != nil {
			defer rows.Close()
			var exactMatch []model.Stock
			for rows.Next() {
				var stock model.Stock
				if err := rows.Scan(&stock.Symbol, &stock.CompanyName); err == nil {
					exactMatch = append(exactMatch, stock)
				}
			}
			if len(exactMatch) > 0 {
				return exactMatch, nil
			}
		}

		// Fallback to fuzzy match if no exact match found
		rows, err = db.Query(`
			SELECT symbol, company_name FROM stocks
			WHERE LOWER(symbol) LIKE '%' || LOWER($1) || '%'
			OR LOWER(company_name) LIKE '%' || LOWER($1) || '%'
			LIMIT 10
		`, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []model.Stock
	for rows.Next() {
		var stock model.Stock
		if err := rows.Scan(&stock.Symbol, &stock.CompanyName); err == nil {
			stocks = append(stocks, stock)
		}
	}
	return stocks, nil
}

// GetYahooFinancePrice fetches the current price from Yahoo Finance
func GetStockPerformance(symbol string, companyName string) (model.StockPerformance, error) {
	url := fmt.Sprintf(os.Getenv(helper.EnvironmentConstant().STOCK_PRICE_URL)+"%s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return model.StockPerformance{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.StockPerformance{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp model.StockAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return model.StockPerformance{}, err
	}

	// Prepare Entries map with growth calculation
	entries := make(map[string]model.HistoricalEntry)

	// 1M growth
	if apiResp.Price1m != 0 {
		growth := ((apiResp.CurrentPrice - apiResp.Price1m) / apiResp.Price1m) * 100
		entries["1M"] = model.HistoricalEntry{
			FromPrice: apiResp.Price1m,
			ToPrice:   apiResp.CurrentPrice,
			Growth:    growth,
		}
	}

	// 1Y growth
	if apiResp.Price1y != 0 {
		growth := ((apiResp.CurrentPrice - apiResp.Price1y) / apiResp.Price1y) * 100
		entries["1Y"] = model.HistoricalEntry{
			FromPrice: apiResp.Price1y,
			ToPrice:   apiResp.CurrentPrice,
			Growth:    growth,
		}
	}

	// 5Y growth
	if apiResp.Price5y != 0 {
		growth := ((apiResp.CurrentPrice - apiResp.Price5y) / apiResp.Price5y) * 100
		entries["5Y"] = model.HistoricalEntry{
			FromPrice: apiResp.Price5y,
			ToPrice:   apiResp.CurrentPrice,
			Growth:    growth,
		}
	}

	// Create StockPerformance object
	stockPerf := model.StockPerformance{
		CompanyName: companyName,
		Symbol:      apiResp.Symbol,
		Current:     apiResp.CurrentPrice,
		Open:        apiResp.OpenPrice,
		Timestamp:   time.Now(),
		Entries:     entries,
	}

	return stockPerf, nil
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
