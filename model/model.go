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
