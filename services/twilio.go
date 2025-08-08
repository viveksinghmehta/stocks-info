package services

import (
	"log"
	"os"
	"stocks-info-channel/helper"

	"github.com/twilio/twilio-go"
	openApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func SendWhatsApp(to, body string) (*openApi.ApiV2010Message, error) {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv(helper.EnvironmentConstant().TWILIO_ACCOUNT_SID),
		Password: os.Getenv(helper.EnvironmentConstant().TWILIO_AUTH_TOKEN),
	})
	params := &openApi.CreateMessageParams{}
	log.Println(" Phone Number :- ", os.Getenv(helper.EnvironmentConstant().PHONE_NUMBER))
	params.SetFrom(helper.AppConstant().WhatsApp + os.Getenv(helper.EnvironmentConstant().PHONE_NUMBER))
	params.SetTo(helper.AppConstant().WhatsApp + to)
	params.SetBody(body)
	return client.Api.CreateMessage(params)
}
