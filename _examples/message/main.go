/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/interactive"
	"github.com/piusalfred/whatsapp/message/template"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func main() {
	ctx := context.Background()
	telemetry, err := InitTelemetry(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = telemetry.Close(context.Background()) }()

	conf := &config.Config{
		BaseURL:           whatsapp.BaseURL,
		APIVersion:        whatsapp.LowestSupportedAPIVersion,
		AccessToken:       os.Getenv("WHATSAPP_CLOUD_API_ACCESS_TOKEN"),
		PhoneNumberID:     os.Getenv("WHATSAPP_CLOUD_API_PHONE_NUMBER_ID"),
		BusinessAccountID: "",
		AppSecret:         "",
		AppID:             "",
		SecureRequests:    false,
		DebugLogLevel:     "",
	}

	// Client with OTEL-instrumented transport and sender middleware.
	client := message.NewClient(conf,
		whttp.WithSenderHTTPClient(&http.Client{
			Transport: OTelHTTPTransport(&TransportParams{
				Propagators:    telemetry.Propagator,
				MeterProvider:  telemetry.MeterProvider,
				TracerProvider: telemetry.TraceProvider,
				ServerName:     "whatsapp-cloud-api-server",
			}),
			Timeout: 30 * time.Second,
		}),
	)

	client.SetMiddlewares(telemetry.Middleware())

	recipient := message.SendTo(os.Getenv("WHATSAPP_CLOUD_API_TEST_NUMBER"))

	tmpl := template.NewInteractiveTemplate(
		"hello_world",
		&template.Language{Code: "en_US"},
		nil, nil, nil,
	)
	resp, err := client.SendTemplateMessage(ctx, recipient, tmpl)
	if err != nil {
		telemetry.Logger.Error("template message", "error", err)
		return
	}
	telemetry.Logger.Info("template message sent", "id", resp.Messages[0].ID)

	resp, err = client.SendTextMessage(ctx, recipient, &message.Text{
		PreviewURL: true,
		Body:       "Visit the repo at https://github.com/piusalfred/whatsapp",
	})
	if err != nil {
		telemetry.Logger.Error("text message", "error", err)
		return
	}
	telemetry.Logger.InfoContext(ctx, "text message sent",
		slog.String("id", resp.Messages[0].ID),
		slog.Any("debug", resp.Debug),
		slog.Any("debug_headers", resp.DebugHeaders),
	)

	resp, err = client.SendInteractiveMessage(ctx, recipient,
		interactive.CTAURLButton(&interactive.CTAURLRequest{
			DisplayText: "Github Link",
			URL:         "https://github.com/piusalfred/whatsapp",
			Body:        "A highly configurable client for WhatsApp chatbots",
			Header:      interactive.HeaderText("WhatsApp Cloud API Client"),
			Footer:      "Frequently updated",
		}),
	)
	if err != nil {
		telemetry.Logger.Error("interactive CTA", "error", err)
		return
	}

	telemetry.Logger.Info("interactive CTA sent", "id", resp.Messages[0].ID)

	resp, err = client.SendInteractiveMessage(ctx, recipient,
		interactive.LocationRequest("Where are you?"),
	)
	if err != nil {
		telemetry.Logger.Error("location request", "error", err)
		return
	}

	telemetry.Logger.Info("location request sent", "id", resp.Messages[0].ID)

	resp, err = client.SendLocationMessage(ctx, recipient, &message.Location{
		Longitude: -3.688344,
		Latitude:  40.453053,
		Name:      "Estadio Santiago Bernabéu",
		Address:   "Av. de Concha Espina, 1, Chamartín, 28036 Madrid, Spain",
	})
	if err != nil {
		telemetry.Logger.Error("location", "error", err)
		return
	}
	telemetry.Logger.Info("location sent", "id", resp.Messages[0].ID)

	contacts := &message.Contacts{
		message.NewContact(
			message.WithContactName(&message.Name{
				FormattedName: "Dr. John A. Doe Sr.",
				FirstName:     "John",
				LastName:      "Doe",
			}),
			message.WithContactURLs(&message.URL{
				Type: "Github",
				URL:  "https://github.com/piusalfred/whatsapp",
			}),
		),
	}
	resp, err = client.SendContactsMessage(ctx, recipient, contacts)
	if err != nil {
		telemetry.Logger.Error("contacts", "error", err)
		return
	}
	telemetry.Logger.Info("contacts sent", "id", resp.Messages[0].ID)

	resp, err = client.SendReactionMessage(ctx, recipient, &message.Reaction{
		MessageID: resp.Messages[0].ID,
		Emoji:     "🤝",
	})
	if err != nil {
		telemetry.Logger.Error("reaction", "error", err)
		return
	}
	telemetry.Logger.Info("reaction sent", "id", resp.Messages[0].ID)
}
