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
	"net/http"

	"github.com/piusalfred/libwhatsapp/webhooks/flow"

	"github.com/piusalfred/libwhatsapp/webhooks"
	"github.com/piusalfred/libwhatsapp/webhooks/business"
	"github.com/piusalfred/libwhatsapp/webhooks/message"
)

func HandleBusinessNotification(ctx context.Context, notification *business.Notification) *webhooks.Response {
	fmt.Printf("Business notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

func HandleMessageNotification(ctx context.Context, notification *message.Notification) *webhooks.Response {
	fmt.Printf("Message notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

func HandleFlowNotification(ctx context.Context, notification *flow.Notification) *webhooks.Response {
	fmt.Printf("Message notification received: %+v\n", notification)
	return &webhooks.Response{StatusCode: http.StatusOK}
}

func main() {
	messageWebhooksHandler := webhooks.OnEventNotification(
		message.NotificationHandlerFunc(
			HandleMessageNotification,
		),
	)

	businessWebhooksHandler := webhooks.OnEventNotification(
		business.NotificationHandlerFunc(
			HandleBusinessNotification,
		),
	)

	flowEventsHandler := webhooks.OnEventNotification(
		flow.NotificationHandlerFunc(HandleFlowNotification),
	)

	messageVerifyToken := webhooks.VerifyTokenReader(func(ctx context.Context) (string, error) {
		return "message-subscription-token", nil
	})

	businessVerifyToken := webhooks.VerifyTokenReader(func(ctx context.Context) (string, error) {
		return "business-subscription-token", nil
	})

	validatorWrapper := func(next http.HandlerFunc) http.HandlerFunc {
		fn := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if err := webhooks.ValidateRequestPayloadSignature(request, "validate-secret"); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)

				return
			}
			next(writer, request)
		})

		return fn
	}

	businessVerifyHandler1 := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		webhooks.SubscriptionVerificationHandlerFunc("business-subscription-token")(writer, request)
	})

	http.HandleFunc("POST /webhooks/messages", validatorWrapper(messageWebhooksHandler))
	http.HandleFunc("POST /webhooks/business", validatorWrapper(businessWebhooksHandler))
	http.HandleFunc("POST /webhooks/flows", flowEventsHandler)
	http.HandleFunc("POST /webhooks/messages/verify", messageVerifyToken.VerifySubscription)
	http.HandleFunc("POST /webhooks/business/verify", businessVerifyToken.VerifySubscription)
	http.HandleFunc("POST /webhooks/business/verify/2", businessVerifyHandler1)
	http.ListenAndServe(":8080", nil)
}
