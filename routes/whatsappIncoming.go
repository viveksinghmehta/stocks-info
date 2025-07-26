package routes

import (
	"net/http"
	"os"
	"stocks-info-channel/helper"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	openApi "github.com/twilio/twilio-go/rest/api/v2010"
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

func twillioClient(phone, message string) (*openApi.ApiV2010Message, error) {
	sid := os.Getenv(helper.EnvironmentConstant().TWILIO_ACCOUNT_SID)
	password := os.Getenv(helper.EnvironmentConstant().TWILIO_AUTH_TOKEN)

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: sid,
		Password: password,
	})

	params := &openApi.CreateMessageParams{}
	params.SetFrom(helper.AppConstant().WhatsApp + helper.AppConstant().Phone_Number) // Twilio phone number
	println("### ", helper.AppConstant().WhatsApp+helper.AppConstant().Phone_Number)
	params.SetTo(helper.AppConstant().WhatsApp + phone) // Recipient's phone number
	params.SetBody(message)

	return client.Api.CreateMessage(params)
}

func WhatsAppIncoming(c *gin.Context) {
	var message TwillioWhatsappMessageRequest

	if err := c.Bind(&message); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Could not decode the values",
		})
	}

	messageBody := "Hello 👋🏻, This is Stocks Info Account"

	println("FROM :- ", message.To)
	println("TO :- ", message.From)

	phone := strings.TrimPrefix(message.From, helper.AppConstant().WhatsApp)

	response, error := twillioClient(phone, messageBody)
	if error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"message":     error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"OK":       "OK",
		"response": response,
	})
}
