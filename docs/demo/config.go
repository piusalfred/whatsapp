// Package demo provides shared configuration and helpers for interactive
// examples demonstrating the WhatsApp Cloud API library.
package demo

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/piusalfred/whatsapp/config"
)

// ErrMissingConfig is returned when required environment variables are not set.
var ErrMissingConfig = errors.New("missing required configuration")

// LoadConfig reads configuration from environment variables. If a .env file
// exists in the working directory, it is loaded first. Callers can override
// any value by setting the corresponding environment variable.
func LoadConfig() (*config.Config, error) {
	_ = godotenv.Load()

	conf := &config.Config{
		BaseURL:           getEnv("WHATSAPP_BASE_URL", "https://graph.facebook.com"),
		APIVersion:        getEnv("WHATSAPP_API_VERSION", "v22.0"),
		AccessToken:       os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		BusinessAccountID: os.Getenv("WHATSAPP_BUSINESS_ACCOUNT_ID"),
		AppSecret:         os.Getenv("WHATSAPP_APP_SECRET"),
		AppID:             os.Getenv("WHATSAPP_APP_ID"),
		SecureRequests:    os.Getenv("WHATSAPP_SECURE_REQUESTS") == "true",
		DebugLogLevel:     os.Getenv("WHATSAPP_DEBUG_LOG_LEVEL"),
	}

	if conf.AccessToken == "" || conf.PhoneNumberID == "" {
		return nil, ErrMissingConfig
	}

	return conf, nil
}

// TestNumber returns the phone number used for test/demo recipients.
func TestNumber() string {
	return os.Getenv("WHATSAPP_TEST_NUMBER")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
