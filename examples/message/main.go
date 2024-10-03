/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func LoadConfigFromFile(filepath string) (config.ReaderFunc, string) {
	fn := func(ctx context.Context) (*config.Config, error) {
		err := godotenv.Load(filepath) // Load the .env file from the given path
		if err != nil {
			return nil, err
		}

		return &config.Config{
			BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
			APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
			AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
			PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
			BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		}, nil
	}

	recipient := os.Getenv("WHATSAPP_CLOUD_API_TEST_RECIPIENT")

	return fn, recipient
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientRequestInterceptor[message.Message](
			func(ctx context.Context, req *http.Request) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "request intercepted",
					slog.String("http.request.method", req.Method),
					slog.String("http.request.url", req.URL.String()),
				)
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[message.Message](
			func(ctx context.Context, resp *http.Response) error {
				logger.LogAttrs(ctx, slog.LevelInfo, "request intercepted",
					slog.String("http.response.status", resp.Status),
					slog.Int("http.response.code", resp.StatusCode),
				)
				return nil
			},
		),
	}

	ctx := context.Background()

	coreClient := whttp.NewSender[message.Message](clientOptions...)
	reader, recipient := LoadConfigFromFile("api.env")
	baseClient, err := message.NewBaseClient(coreClient, reader)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error creating base client", slog.String("error", err.Error()))
		return
	}

	initTmpl := message.WithTemplateMessage(&message.Template{
		Name: "hello_world",
		Language: &message.TemplateLanguage{
			Code: "en_US",
		},
	})

	initTmplMessage, err := message.New(recipient, initTmpl)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error creating initial template message", slog.String("error", err.Error()))
		return
	}

	response, err := baseClient.SendMessage(ctx, initTmplMessage)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending initial template message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from initial template message", slog.Any("response", response))

	interactiveMessage := message.NewInteractiveCTAURLButton(&message.InteractiveCTARequest{
		DisplayText: "",
		URL:         "",
		Body:        "",
		Header:      nil,
		Footer:      "",
	})

	ir := message.NewRequest(recipient, interactiveMessage, "")

	// use dedicated method
	response, err = baseClient.SendInteractiveMessage(ctx, ir)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending interactive CTA message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from interactive CTA message", slog.Any("response", response))

	msg := "where are you?"

	locationMessage, err := message.New(recipient, message.WithRequestLocationMessage(&msg))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error creating location request message", slog.String("error", err.Error()))
		return
	}

	response, err = baseClient.SendMessage(ctx, locationMessage)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending location request message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from location request message", slog.Any("response", response))
}
