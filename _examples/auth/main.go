//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/piusalfred/whatsapp/auth"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	_ = godotenv.Load("api.env")

	conf := &config.Config{
		BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
		APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
		AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
		BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		AppSecret:         os.Getenv("WHATSAPP_CLOUD_API_APP_SECRET"),
		AppID:             os.Getenv("WHATSAPP_CLOUD_API_APP_ID"),
		SecureRequests:    os.Getenv("WHATSAPP_CLOUD_API_SECURE_REQUESTS") == "true",
		DebugLogLevel:     os.Getenv("WHATSAPP_CLOUD_API_DEBUG_LOG_LEVEL"),
	}

	client := auth.NewClient(conf,
		whttp.WithSenderRequestInterceptor(
			func(ctx context.Context, req *http.Request) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "request intercepted",
					slog.String("method", req.Method),
					slog.String("url", req.URL.String()),
				)
				return nil
			},
		),
	)

	ctx := context.Background()

	response, err := client.GenerateAccessToken(ctx, &auth.GenerateAccessTokenParams{
		AccessToken:  conf.AccessToken,
		AppID:        conf.AppID,
		SystemUserID: conf.SystemUserID,
		AppSecret:    conf.AppSecret,
		Scopes: []string{
			auth.TokenScopeWhatsappBusinessManagement,
			auth.TokenScopeWhatsappBusinessMessaging,
		},
		SetTokenExpiresIn60: true,
	})
	if err != nil {
		logger.Error("generate access token", "error", err)
		return
	}

	logger.Info("token generated", "response", response)
}
