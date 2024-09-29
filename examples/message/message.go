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
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"

	"github.com/piusalfred/whatsapp/message"
)

type ProfanityChecker struct {
	profanities map[string]struct{}
}

func NewProfanityChecker(profanities []string) *ProfanityChecker {
	bank := make(map[string]struct{})
	for _, word := range profanities {
		bank[word] = struct{}{}
	}
	return &ProfanityChecker{profanities: bank}
}

// Censor checks for profane words in the message and replaces all occurrences by looping through the profanity list.
func (p *ProfanityChecker) Censor(message string) string {
	for profaneWord := range p.profanities {
		// Perform a case-insensitive replace to cover all cases
		message = strings.ReplaceAll(message, profaneWord, "****")
	}
	return message
}

func (p *ProfanityChecker) ProfanityMiddleware(next whttp.SenderFunc[message.Message]) whttp.SenderFunc[message.Message] {
	return func(ctx context.Context, req *whttp.Request[message.Message], v whttp.ResponseDecoder) error {
		req.Message.Text.Body = p.Censor(req.Message.Text.Body)
		return next(ctx, req, v)
	}
}

type Translator struct {
	bank map[string]map[string]string
}

type TranslationOptions struct {
	ShouldTranslate bool
	Language        string
}

func NewTranslator(bank map[string]map[string]string) *Translator {
	return &Translator{bank: bank}
}

func (t *Translator) Translate(message string, lang string) string {
	words := strings.Fields(message)
	for i, word := range words {
		if translation, exists := t.bank[word][lang]; exists {
			words[i] = translation
		}
	}
	return strings.Join(words, " ")
}

func (t *Translator) TranslationMiddleware(next whttp.SenderFunc[message.Message]) whttp.SenderFunc[message.Message] {
	return func(ctx context.Context, req *whttp.Request[message.Message], v whttp.ResponseDecoder) error {
		metadata := message.RetrieveMessageMetadata(ctx)
		if options, ok := metadata["TranslationOptions"].(TranslationOptions); ok && options.ShouldTranslate {
			req.Message.Text.Body = t.Translate(req.Message.Text.Body, options.Language)
		}
		return next(ctx, req, v)
	}
}

type Bot struct {
	http        *http.Client
	logger      *slog.Logger
	client      *message.BaseClient
	translator  *Translator
	profanities *ProfanityChecker
}

func NewBot(hc *http.Client, logger *slog.Logger) *Bot {
	b := &Bot{
		http:   hc,
		logger: logger,
	}

	b.profanities = NewProfanityChecker([]string{"badword1", "badword2"})
	b.translator = NewTranslator(map[string]map[string]string{
		"hello": {"germany": "hallo"},
	})

	middlewares := []whttp.Middleware[message.Message]{
		b.translator.TranslationMiddleware,
		b.profanities.ProfanityMiddleware,
	}

	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientHTTPClient[message.Message](b.http),
		whttp.WithCoreClientRequestInterceptor[message.Message](
			func(ctx context.Context, req *http.Request) error {
				fmt.Println("Request Intercepted")
				return nil
			},
		),
		whttp.WithCoreClientResponseInterceptor[message.Message](
			func(ctx context.Context, resp *http.Response) error {
				fmt.Println("Response Intercepted")
				return nil
			},
		),
		whttp.WithCoreClientMiddlewares(middlewares),
	}

	coreClient := whttp.NewSender[message.Message](clientOptions...)
	coreClient.PrependMiddlewares(b.SendMessageLogger)
	baseClient, _ := message.NewBaseClient(coreClient, config.ReaderFunc(b.ReadConf))
	b.client = baseClient

	return b
}

func (b *Bot) SendMessageLogger(next whttp.SenderFunc[message.Message]) whttp.SenderFunc[message.Message] {
	return func(ctx context.Context, req *whttp.Request[message.Message], decoder whttp.ResponseDecoder) error {
		b.logger.InfoContext(ctx, "incoming request", slog.Any("request", req))
		err := next(ctx, req, decoder)
		metadata := message.RetrieveMessageMetadata(ctx)
		b.logger.InfoContext(ctx, "metadata retrieved", slog.Any("metadata", metadata))
		return err
	}
}

func (b *Bot) ReadConf(ctx context.Context) (*config.Config, error) {
	return &config.Config{
		BaseURL:           "",
		APIVersion:        "",
		AccessToken:       "",
		PhoneNumberID:     "",
		BusinessAccountID: "",
	}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	bot := NewBot(&http.Client{}, logger)

	recipient := ""

	ctx := context.Background()
	interactiveMessage := message.NewInteractiveCTAURLButton(&message.InteractiveCTARequest{
		DisplayText: "",
		URL:         "",
		Body:        "",
		Header:      nil,
		Footer:      "",
	})

	ir := message.NewRequest(recipient, interactiveMessage, "")

	// use dedicated method
	response, err := bot.client.SendInteractiveMessage(ctx, ir)
	if err != nil {
		return
	}

	fmt.Printf("%v", response)

	msg := "where are you?"

	locationMessage, _ := message.New(recipient, message.WithRequestLocationMessage(&msg))

	// use generic method
	response, err = bot.client.SendMessage(ctx, locationMessage)
	if err != nil {
		return
	}
}
