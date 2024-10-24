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
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/pkg/crypto"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func LoadConfigFromFile(filepath string) (config.ReaderFunc, string) {
	err := godotenv.Load(filepath) // Load the .env file from the given path
	if err != nil {
		panic(err)
	}
	recipient := os.Getenv("WHATSAPP_CLOUD_API_TEST_RECIPIENT")

	secureRequestsStr := os.Getenv("WHATSAPP_CLOUD_API_SECURE_REQUESTS")

	secureRequests := false

	if secureRequestsStr == "true" {
		secureRequests = true
	}

	conf := &config.Config{
		BaseURL:           os.Getenv("WHATSAPP_CLOUD_API_BASE_URL"),
		APIVersion:        os.Getenv("WHATSAPP_CLOUD_API_API_VERSION"),
		AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
		BusinessAccountID: os.Getenv("WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID"),
		AppSecret:         os.Getenv("WHATSAPP_CLOUD_API_APP_SECRET"),
		SecureRequests:    secureRequests,
	}
	fn := func(ctx context.Context) (*config.Config, error) {
		return conf, nil
	}

	prrof, _ := crypto.GenerateAppSecretProof(conf.AccessToken, conf.AppSecret)
	fmt.Println("PROOOOOF" + prrof)

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
				logger.LogAttrs(ctx, slog.LevelInfo, "response intercepted",
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

	textMessage := &message.Text{
		PreviewURL: true,
		Body:       "Visit the repo at https://github.com/piusalfred/whatsapp",
	}

	response, err = baseClient.SendText(ctx, message.NewRequest(recipient, textMessage, ""))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending text message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from sending text message", slog.Any("response", response))

	interactiveMessage := message.NewInteractiveCTAURLButton(&message.InteractiveCTARequest{
		DisplayText: "Github Link",
		URL:         "https://github.com/piusalfred/whatsapp",
		Body:        "A highly configurable client to build a wide range of whatsapp chatbots",
		Header: &message.InteractiveHeader{
			Text: "Whatsapp Cloud API Client",
			Type: "text",
		},
		Footer: "Frequently updated",
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

	location := &message.Location{
		Longitude: -3.688344,
		Latitude:  40.453053,
		Name:      "Estadio Santiago Bernabéu",
		Address:   "Av. de Concha Espina, 1, Chamartín, 28036 Madrid, Spain",
	}

	response, err = baseClient.SendLocation(ctx, message.NewRequest(recipient, location, ""))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending location message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from location message", slog.Any("response", response))

	contact := message.NewContact(
		message.WithContactURLs(
			[]*message.URL{
				{
					Type: "Github",
					URL:  "https://github.com/piusalfred/whatsapp",
				},
			}...,
		),
		message.WithContactName(&message.Name{
			FormattedName: "Dr. John A. Doe Sr.",
			FirstName:     "John",
			LastName:      "Doe",
			MiddleName:    "Anon",
			Suffix:        "Sr.",
			Prefix:        "Dr.",
		}),
	)

	contacts := &message.Contacts{contact}

	mc, err := message.New(recipient, message.WithContacts(contacts))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error creating contacts message", slog.String("error", err.Error()))
		return

	}

	response, err = baseClient.SendMessage(ctx, mc)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error sending contacts message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from contacts message", slog.Any("response", response))

	messageID := response.Messages[0].ID

	// react to the above message
	reaction := &message.Reaction{
		MessageID: messageID,
		Emoji:     "🤝",
	}

	response, err = baseClient.SendReaction(ctx, message.NewRequest(recipient, reaction, ""))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error reacting to message", slog.String("error", err.Error()))
		return
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "response from reaction message", slog.Any("response", response))
}
