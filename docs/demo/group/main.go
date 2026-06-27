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
	"github.com/piusalfred/whatsapp/groups"
	"github.com/piusalfred/whatsapp/message"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
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

	logger.Debug("Config loaded", "debug_log_level", config.DebugLogLevel)

	ctx := context.Background()

	// ---- Groups client ----
	groupClient := groups.NewClient(conf, whttp.WithSenderTimeout(30*time.Second))
	groupClient.SetMiddlewares(responseCapturer[groups.BaseRequest](logger))

	// ---- Message client ----
	msgClient := message.NewClient(conf, whttp.WithSenderTimeout(30*time.Second))
	msgClient.SetMiddlewares(responseCapturer[message.BaseRequest](logger))

	// ---- 1. Create a group ----
	createResp, err := groupClient.CreateGroup(ctx, &groups.CreateGroupRequest{
		Subject:          "SDK Test Group",
		Description:      "Created by WhatsApp SDK demo",
		JoinApprovalMode: groups.JoinApprovalModeAuto,
	})
	if err != nil {
		logger.Error("create group", "error", err)
		return
	}
	logger.Info("group created", "id", createResp.ID, "subject", createResp.Subject)

	groupRecipient := message.SendTo(createResp.ID).AsGroupMessage()

	// ---- 2. Send a text message to the group ----
	textResp, err := msgClient.SendTextMessage(ctx, groupRecipient, &message.Text{
		Body: "Hello from the SDK! This message will be pinned.",
	})
	if err != nil {
		logger.Error("send text to group", "error", err)
		return
	}
	logger.Info("text sent to group", "id", textResp.Messages[0].ID)

	// ---- 3. Pin that message ----
	pinResp, err := msgClient.SendPinMessage(ctx, groupRecipient, &message.Pin{
		Type:           message.PinOperationPinMessage,
		MessageID:      textResp.Messages[0].ID,
		ExpirationDays: 7,
	})
	if err != nil {
		logger.Error("pin message", "error", err)
		return
	}
	logger.Info("message pinned", "id", pinResp.Messages[0].ID)
}

// responseCapturer wraps every request with a ResponseCapturer and logs the raw response.
func responseCapturer[T any](logger *slog.Logger) whttp.Middleware[T] {
	return func(next whttp.SenderFunc[T]) whttp.SenderFunc[T] {
		return whttp.SenderFunc[T](
			func(ctx context.Context, request *whttp.Request[T], decoder whttp.ResponseDecoder) error {
				capturer := whttp.NewResponseCapturer(decoder)
				url, _ := request.URL()

				logger.Info("sending request", "url", url)
				if err := next(ctx, request, capturer); err != nil {
					logger.Error("request failed",
						"error", err,
						"status", capturer.StatusCode,
						"header", capturer.Header,
						"body", string(capturer.Body),
					)
					return err
				}

				logger.Info("request succeeded", "status", capturer.StatusCode, "header", capturer.Header)
				logger.Debug("response body", "body", string(capturer.Body))

				return nil
			},
		)
	}
}
