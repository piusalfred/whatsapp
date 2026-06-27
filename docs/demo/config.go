//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package demo provides shared configuration and helpers for interactive
// examples demonstrating the WhatsApp Cloud API library.
package demo

import (
	"errors"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/config"
)

type Config struct {
	config.Config
	TestSender   string
	TestReceiver string
}

// ErrMissingConfig is returned when required environment variables are not set.
var ErrMissingConfig = errors.New("missing required configuration")

// LoadConfig reads configuration from environment variables. If envFile is
// non-empty, the file is loaded first via godotenv. Callers can override any
// value by setting the corresponding environment variable.
func LoadConfig(envFile string) (*Config, error) {
	if envFile == "" {
		envFile = ".env"
	}

	values, err := godotenv.Read(envFile)
	if err != nil {
		return nil, fmt.Errorf("read env file: %w", err)
	}

	conf := &Config{
		Config: config.Config{
			BaseURL:           getEnv(values, "WHATSAPP_CLOUD_API_BASE_URL", "https://graph.facebook.com"),
			APIVersion:        getEnv(values, "WHATSAPP_CLOUD_API_API_VERSION", "v22.0"),
			AccessToken:       values["WHATSAPP_CLOUD_API_ACCESS_TOKEN"],
			PhoneNumberID:     values["WHATSAPP_CLOUD_API_PHONENUMBER_ID"],
			BusinessAccountID: values["WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"],
			AppSecret:         values["WHATSAPP_CLOUD_API_APP_SECRET"],
			AppID:             values["WHATSAPP_CLOUD_API_APP_ID"],
			SecureRequests:    values["WHATSAPP_CLOUD_API_SECURE_REQUESTS"] == "true",
			DebugLogLevel:     values["WHATSAPP_CLOUD_API_DEBUG_LOG_LEVEL"],
		},
		TestSender:   values["WHATSAPP_CLOUD_API_TEST_SENDER"],
		TestReceiver: values["WHATSAPP_CLOUD_API_TEST_RECEIVER"],
	}

	if conf.AccessToken == "" || conf.PhoneNumberID == "" {
		return nil, ErrMissingConfig
	}

	return conf, nil
}

func getEnv(values map[string]string, key, fallback string) string {
	if v := values[key]; v != "" {
		return v
	}
	return fallback
}
