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

package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/piusalfred/whatsapp/docs/demo"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/interactive"
	"github.com/piusalfred/whatsapp/message/template"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	config, err := demo.LoadConfig("../.env")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}

	conf := &config.Config

	logger.Debug("Config loaded successfully", "debug_log_level", config.DebugLogLevel, "secure", config.SecureRequests)

	lm := whttp.Middleware[message.BaseRequest](
		func(next whttp.SenderFunc[message.BaseRequest]) whttp.SenderFunc[message.BaseRequest] {
			fn := whttp.SenderFunc[message.BaseRequest](
				func(ctx context.Context, request *whttp.Request[message.BaseRequest], decoder whttp.ResponseDecoder) error {
					capturer := whttp.NewResponseCapturer(decoder)
					url, _ := request.URL()

					if request.Metadata != nil {
						logger.Info("request has metadata", "metadata", request.Metadata)
					}

					logger.Info("sending message", "url", url)
					if err := next(ctx, request, capturer); err != nil {
						logger.Error("send message", "error",
							err, "status",
							capturer.StatusCode,
							"header", capturer.Header,
							"body", string(capturer.Body))
						return err
					}

					logger.Info("message sent successfully", "status", capturer.StatusCode, "header", capturer.Header)
					logger.Debug("response body", "body", string(capturer.Body))

					if request.Metadata != nil {
						logger.Info("metadata still on request after send", "metadata", request.Metadata)
					}

					return nil
				},
			)

			return fn
		},
	)

	client := message.NewClient(conf, whttp.WithSenderTimeout(30*time.Second))
	client.SetMiddlewares(lm)

	si := message.SendTo(config.TestReceiver)
	si.BizOpaqueCallbackData("")
	si.AsGroupMessage()
	si.Metadata(map[string]any{
		"initial_key": "initial_value",
	})
	ctx := context.Background()

	tmpl := template.NewInteractiveTemplate(
		"hello_world",
		&template.Language{Code: "en_US"},
		nil, nil, nil,
	)
	resp, err := client.SendTemplateMessage(ctx, si, tmpl)
	if err != nil {
		logger.Error("template message", "error", err)
		return
	}
	logger.Info("template message sent", "id", resp.Messages[0].ID)

	resp, err = client.SendTextMessage(ctx, si.ReplyTo(resp.Messages[0].ID), &message.Text{
		PreviewURL: true,
		Body:       "Visit the repo at https://github.com/piusalfred/whatsapp",
	})
	if err != nil {
		logger.Error("text message", "error", err)
		return
	}
	logger.Info("text message sent", "id", resp.Messages[0].ID)

	resp, err = client.SendInteractiveMessage(ctx, si,
		interactive.CTAURLButton(&interactive.CTAURLRequest{
			DisplayText: "Github Link",
			URL:         "https://github.com/piusalfred/whatsapp",
			Body:        "A highly configurable client for WhatsApp chatbots",
			Header:      interactive.HeaderText("WhatsApp Cloud API Client"),
			Footer:      "Frequently updated",
		}),
	)
	if err != nil {
		logger.Error("interactive CTA", "error", err)
		return
	}
	logger.Info("interactive CTA sent", "id", resp.Messages[0].ID)

	resp, err = client.SendInteractiveMessage(ctx, si,
		interactive.LocationRequest("Where are you?"),
	)
	if err != nil {
		logger.Error("location request", "error", err)
		return
	}
	logger.Info("location request sent", "id", resp.Messages[0].ID)

	resp, err = client.SendLocationMessage(ctx, si, &message.Location{
		Longitude: -3.688344,
		Latitude:  40.453053,
		Name:      "Estadio Santiago Bernabéu",
		Address:   "Av. de Concha Espina, 1, Chamartín, 28036 Madrid, Spain",
	})
	if err != nil {
		logger.Error("location", "error", err)
		return
	}
	logger.Info("location sent", "id", resp.Messages[0].ID)

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
	resp, err = client.SendContactsMessage(ctx, si, contacts)
	if err != nil {
		logger.Error("contacts", "error", err)
		return
	}
	logger.Info("contacts sent", "id", resp.Messages[0].ID)

	resp, err = client.SendReactionMessage(ctx, si, &message.Reaction{
		MessageID: resp.Messages[0].ID,
		Emoji:     "🤝",
	})
	if err != nil {
		logger.Error("reaction", "error", err)
		return
	}
	logger.Info("reaction sent", "id", resp.Messages[0].ID)

	// ---- Metadata demonstration using SendInfo.WithMetadata ----
	info := message.SendTo(config.TestReceiver).Metadata(types.Metadata{
		"correlation_id": "demo-run-001",
		"timestamp":      time.Now().Unix(),
	})
	metaMsg := message.New(info,
		message.WithTextMessage(&message.Text{
			Body: "This message carries metadata.",
		}),
	)
	metaResp, metaErr := client.SendMessage(ctx, info, metaMsg)
	if metaErr != nil {
		logger.Error("metadata message", "error", metaErr)
		return
	}
	logger.Info("metadata message sent",
		"id", metaResp.Messages[0].ID,
		"message_metadata", metaResp.MessageMetadata,
	)
}
