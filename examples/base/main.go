/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"fmt"
	"github.com/piusalfred/whatsapp/pkg/models/factories"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/pkg/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

var _ config.Reader = (*dotEnvReader)(nil)

type dotEnvReader struct {
	filePath string
}

func (d *dotEnvReader) Read(ctx context.Context) (*config.Values, error) {
	vm, err := godotenv.Read(d.filePath)
	if err != nil {
		return nil, err
	}

	return &config.Values{
		BaseURL:           vm["BASE_URL"],
		Version:           vm["VERSION"],
		AccessToken:       vm["ACCESS_TOKEN"],
		PhoneNumberID:     vm["PHONE_NUMBER_ID"],
		BusinessAccountID: vm["BUSINESS_ACCOUNT_ID"],
	}, nil
}

func initBaseClient(ctx context.Context) (*whatsapp.Client, error) {
	reader := &dotEnvReader{filePath: ".env"}
	b, err := whatsapp.NewClient(ctx, reader,
		whatsapp.WithBaseClientOptions(
			[]whttp.BaseClientOption{
				whttp.WithHTTPClient(http.DefaultClient),
				whttp.WithRequestHooks(),
				whttp.WithResponseHooks(),
				whttp.WithSendMiddleware(),
			},
		),
		whatsapp.WithSendMiddlewares(),
	)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type textRequest struct {
	Recipient string
	Message   string
}

// sendTextMessage sends a text message to a whatsapp user.
func sendTextMessage(ctx context.Context, request *textRequest) error {
	b, err := initBaseClient(ctx)
	if err != nil {
		return err
	}

	rc := whttp.MakeRequestContext(
		whttp.WithRequestContextAction(whttp.RequestActionSend),
		whttp.WithRequestContextCategory(whttp.RequestCategoryMessage),
	)

	tmp := &models.Message{
		Product:       "whatsapp",
		Type:          "template",
		RecipientType: "individual",
		To:            request.Recipient,
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code: "en_US",
			},
			Name: "hello_world",
		},
	}

	// send a template message first and make sure the user has accepted the template message (replied).
	resp, err := b.Send(ctx, rc, tmp)
	if err != nil {
		return err
	}

	fmt.Printf("\n%+v\n", resp)

	time.Sleep(2 * time.Second)

	message := &models.Message{
		Product:       "whatsapp",
		Type:          "text",
		RecipientType: "individual",
		To:            request.Recipient,
		Text: &models.Text{
			Body:       request.Message,
			PreviewURL: true,
		},
	}

	resp, err = b.Send(ctx, rc, message)
	if err != nil {
		return err
	}

	fmt.Printf("\n%+v\n", resp)

	params := &whatsapp.RequestParams{
		ID:        "1234567890",
		Metadata:  map[string]string{"foo": "bar"},
		Recipient: "+255767001828",
		ReplyID:   "",
	}

	resp, err = b.RequestLocation(ctx, params, "Where are you mate?")
	if err != nil {
		return err
	}

	fmt.Printf("\n%+v\n", resp)

	buttonURL, err := b.InteractiveCTAButtonURL(ctx, params, &factories.CTAButtonURLParameters{
		DisplayText: "link to github repo",
		URL:         "https://github.com/piusalfred/whatsapp",
		Body:        "The Golang client for the WhatsApp Business API offers a rich set of features for building interactive WhatsApp experiences.",
		Footer:      "You can fork,stargaze and contribute to this repo",
		Header:      "Hey look",
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n%+v\n", buttonURL)

	return nil
}

func main() {
	ctx := context.Background()
	err := sendTextMessage(ctx, &textRequest{
		Recipient: "+255767001828",
		Message:   "Hello World From github.com/piusalfred/whatsapp",
	})
	if err != nil {
		panic(err)
	}
}
