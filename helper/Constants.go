package helper

type EnvironmentConstants struct {
	DB_URL             string
	GIN_MODE           string
	TWILIO_ACCOUNT_SID string
	TWILIO_AUTH_TOKEN  string
}

type AppConstants struct {
	Phone_Number string
	WhatsApp     string
}

func EnvironmentConstant() EnvironmentConstants {
	return EnvironmentConstants{
		DB_URL:             "DB_URL",
		GIN_MODE:           "GIN_MODE",
		TWILIO_AUTH_TOKEN:  "TWILIO_AUTH_TOKEN",
		TWILIO_ACCOUNT_SID: "TWILIO_ACCOUNT_SID",
	}
}

func AppConstant() AppConstants {
	return AppConstants{
		Phone_Number: "+14155238886",
		WhatsApp:     "whatsapp:",
	}
}
