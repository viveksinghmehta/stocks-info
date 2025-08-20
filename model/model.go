package model

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
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

// type Stock struct {
// 	Symbol      string `json:"symbol"`
// 	CompanyName string `json:"company_name"`
// }

type Stock struct {
	Symbol      string
	CompanyName string
}

type GrowthEntry struct {
	FromPrice float64
	ToPrice   float64
	Growth    float64
}

type StockGrowthStats struct {
	Symbol    string
	Timestamp time.Time
	Entries   map[string]GrowthEntry
}

type StockAPIResponse struct {
	CurrentPrice float64 `json:"current_price"`
	OpenPrice    float64 `json:"open_price"`
	Price1m      float64 `json:"price_1m"`
	Price1y      float64 `json:"price_1y"`
	Price3y      float64 `json:"price_3y"`
	Price5y      float64 `json:"price_5y"`
	Symbol       string  `json:"symbol"`
}

type HistoricalEntry struct {
	FromPrice float64
	ToPrice   float64
	Growth    float64 // in percentage
}

type StockPerformance struct {
	CompanyName string
	Symbol      string
	Current     float64
	Open        float64
	Timestamp   time.Time
	Entries     map[string]HistoricalEntry
}

// NSEResponse - exported struct (capitalized name)
type NSEResponse struct {
	Data []struct {
		Symbol       string  `json:"symbol"`
		OpenPrice    float64 `json:"openPrice"`
		DayHighPrice float64 `json:"dayHighPrice"`
		DayLowPrice  float64 `json:"dayLowPrice"`
		Ltp          float64 `json:"ltp"`
		NetPrice     float64 `json:"netPrice"` // % change
		TradedVolume int64   `json:"tradedVolume"`
	} `json:"data"`
}
