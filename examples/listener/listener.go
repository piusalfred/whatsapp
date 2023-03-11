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
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/piusalfred/whatsapp/webhooks"
)

type Middleware func(next http.Handler) http.Handler

// Wrap wraps a http.Handler with a middlewares
func Wrap(handler http.Handler, middlewares ...Middleware) http.Handler {
	// wraps backwards
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// StartTimeMiddleware is a middleware that adds the start time to the request context
func StartTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("StartTimeMiddleware")
		ctx := r.Context()
		ctx = context.WithValue(ctx, "startTime", time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EndTimeMiddleware is a middleware that adds the end time to the request context
func EndTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("EndTimeMiddleware")
		ctx := r.Context()
		ctx = context.WithValue(ctx, "endTime", time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GenericNotificationHandler() webhooks.GenericNotificationHandler {
	return func(ctx context.Context, writer http.ResponseWriter, notification *webhooks.Notification,
	) error {
		// print the notification and the time it took to process it
		startTime := ctx.Value("startTime").(time.Time)
		time.Sleep(2 * time.Second)
		endTime := ctx.Value("endTime").(time.Time)
		err := json.NewEncoder(writer).Encode(notification)
		if err != nil {
			return err
		}
		log.Printf("Notification: %+v, Time: %s\n", notification, endTime.Sub(startTime))
		return nil
	}
}

/*
	curl -X POST --location "http://localhost:8080/webhooks" \
		   -H "Content-Type: application/json" \
		   -d "{
		         "object": "whatsapp_business_account",
		         "entry": [{
		           "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
		           "changes": [{
		             "value": {
		               "messaging_product": "whatsapp",
		               "metadata": {
		                 "display_phone_number": "PHONE_NUMBER",
		                 "phone_number_id": "PHONE_NUMBER_ID"
		               },
		               "contacts": [{
		                 "profile": {
		                  	"name": "NAME"
		                 },
		                 "wa_id": "PHONE_NUMBER_ID"
		               }],
		               "messages": [{
		                 "from": "PHONE_NUMBER",
		                 "id": "wamid.ID",
		                 "timestamp": "TIMESTAMP",
		                 "text": {
		                   "body": "MESSAGE_BODY"
		                 },
		                 "type": "text"
		               }]
		             },
		             "field": "messages"
		           }]
		         }]
		       }"
*/
func main() {
	ls := webhooks.NewEventListener(
		webhooks.WithGenericNotificationHandler(GenericNotificationHandler()),
		webhooks.WithNotificationErrorHandler(
			func(ctx context.Context, request *http.Request, err error) *webhooks.Response {
				return &webhooks.Response{
					StatusCode: http.StatusBadRequest,
					Headers:    nil,
					Body:       []byte(err.Error()),
					Skip:       false,
				}
			}),
	)
	middlewares := []Middleware{
		StartTimeMiddleware,
		EndTimeMiddleware,
	}
	finalHandler := Wrap(ls.GlobalHandler(), middlewares...)
	mux := http.NewServeMux()
	mux.Handle("/webhooks", finalHandler)

	// Create a new server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
