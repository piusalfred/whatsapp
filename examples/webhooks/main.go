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
	"log"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
)

// LoggingMiddleware logs the start and end of request processing.
func LoggingMiddleware(next webhooks.NotificationHandlerFunc) webhooks.NotificationHandlerFunc {
	return func(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
		fmt.Println("Logging: Before handling notification")
		response := next(ctx, notification)
		fmt.Println("Logging: After handling notification")
		return response
	}
}

// AddMetadataMiddleware adds some metadata to the context.
func AddMetadataMiddleware(next webhooks.NotificationHandlerFunc) webhooks.NotificationHandlerFunc {
	return func(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
		fmt.Println("Adding metadata to the context")
		ctx = context.WithValue(ctx, "metadata", "some value")
		return next(ctx, notification)
	}
}

func main() {
	handler := webhooks.NewHandler()
	handler.OnTextMessage(func(ctx context.Context, nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo, text *webhooks.Text,
	) error {
		fmt.Printf("new text message received: context: %+v,message info: %+v, text: %+v\n",
			nctx, mctx, text)

		return nil
	})
	messageListener := webhooks.NewListener(
		handler.HandleNotification,
		func(ctx context.Context) (string, error) {
			return "TOKEN", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false,
			AppSecret: "SUPERSECRET",
		},
		LoggingMiddleware,
		AddMetadataMiddleware,
	)

	http.HandleFunc("POST /webhooks/messages", messageListener.HandleNotification)
	http.HandleFunc("GET /webhooks/verify", messageListener.HandleSubscriptionVerification)

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
