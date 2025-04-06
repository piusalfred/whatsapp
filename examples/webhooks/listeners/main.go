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
	"github.com/piusalfred/whatsapp/webhooks/business"
	"github.com/piusalfred/whatsapp/webhooks/message"
)

func HandleBusinessNotification(ctx context.Context, notification *business.Notification) *webhooks.Response {
	fmt.Printf("Business notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

func HandleMessageNotification(ctx context.Context, notification *message.Notification) *webhooks.Response {
	fmt.Printf("Message notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

// LoggingMiddleware logs the start and end of request processing.
func LoggingMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		fmt.Println("Logging: Before handling notification")
		response := next(ctx, notification)
		fmt.Println("Logging: After handling notification")
		return response
	}
}

// AddMetadataMiddleware adds some metadata to the context.
func AddMetadataMiddleware[T any](next webhooks.NotificationHandlerFunc[T]) webhooks.NotificationHandlerFunc[T] {
	return func(ctx context.Context, notification *T) *webhooks.Response {
		fmt.Println("Adding metadata to the context")
		ctx = context.WithValue(ctx, "metadata", "some value")
		return next(ctx, notification)
	}
}

func main() {
	messageListener := webhooks.NewListener(
		HandleMessageNotification,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false,
			AppSecret: "",
		},
		LoggingMiddleware[message.Notification],
		AddMetadataMiddleware[message.Notification],
	)

	businessListener := webhooks.NewListener(
		HandleBusinessNotification,
		func(ctx context.Context) (string, error) {
			return "", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false,
			AppSecret: "",
		},
		LoggingMiddleware[business.Notification],
		AddMetadataMiddleware[business.Notification],
	)

	http.HandleFunc("POST /webhooks/messages", messageListener.HandleNotification)
	http.HandleFunc("POST /webhooks/business", businessListener.HandleNotification)
	http.HandleFunc("POST /webhooks/messages/verify", messageListener.HandleSubscriptionVerification)
	http.HandleFunc("POST /webhooks/business/verify", businessListener.HandleSubscriptionVerification)

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
