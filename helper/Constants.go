package helper

type EnvironmentConstants struct {
	DB_URL             string
	GIN_MODE           string
	TWILIO_ACCOUNT_SID string
	TWILIO_AUTH_TOKEN  string
	PHONE_NUMBER       string
	STOCK_PRICE_URL    string
	PORT               string
}

func EnvironmentConstant() EnvironmentConstants {
	return EnvironmentConstants{
		DB_URL:             "DB_URL",
		GIN_MODE:           "GIN_MODE",
		TWILIO_AUTH_TOKEN:  "TWILIO_AUTH_TOKEN",
		TWILIO_ACCOUNT_SID: "TWILIO_ACCOUNT_SID",
		PHONE_NUMBER:       "PHONE_NUMBER",
		STOCK_PRICE_URL:    "STOCK_PRICE_URL",
		PORT:               "PORT",
	}
}

type AppConstants struct {
	WhatsApp    string
	DefaultPort string
}

func AppConstant() AppConstants {
	return AppConstants{
		WhatsApp:    "whatsapp:",
		DefaultPort: "8080",
	}
}
