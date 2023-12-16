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
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	hooks "github.com/piusalfred/whatsapp/webhooks"
)

type verifier struct {
	secret string
	logger io.Writer
	// other places where you pull the secret e.g database
	// other fields for tracing and logging etc
}

// This is first implementation
func (v *verifier) Verify(ctx context.Context, vr *hooks.VerificationRequest) error {
	if vr.Token != v.secret {
		log.Println("invalid token")
		return errors.New("invalid token")
	}

	log.Println("valid token")
	return nil
}

// This is second implementation
func VerifyFn(secret string) hooks.SubscriptionVerifier {
	return func(ctx context.Context, vr *hooks.VerificationRequest) error {
		if vr.Token != secret {
			log.Println("invalid token")
			return errors.New("invalid token")
		}
		log.Println("valid token")
		return nil
	}
}

// func HandleNotificationError(ctx context.Context, writer http.ResponseWriter, request *http.Request, err error) error {
// 	if err != nil {
// 		log.Printf("HandleError: %+v\n", err)
// 		return err
// 	}

// 	log.Printf("HandleError: NIL")
// 	return nil
// }

func HandleGeneralNotification(ctx context.Context, writer http.ResponseWriter, notification *hooks.Notification) error {
	os.Stdout.WriteString("HandleEvent")
	jsonb, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	// print the string representation of the json
	// os.Stdout.WriteString(string(jsonb))
	log.Printf("\n%s\n", string(jsonb))
	writer.WriteHeader(http.StatusOK)
	return nil
}

func main() {
	router := httprouter.New()

	h := VerifyFn("testtoken").HandlerFunc
	router.HandlerFunc(http.MethodGet, "/webhooks", h)
	/*
		// verifyHandler2
		verifier := &verifier{
			secret: "mytesttoken",
			logger: log.Writer(),
		}

		verifyHandler2 := hooks.VerifySubscriptionHandler(verifier.Verify)
		router.Handler(http.MethodGet, "/webhooks", verifyHandler2)

	*/

	listener := hooks.NewEventListener(hooks.WithGlobalNotificationHandler(HandleGeneralNotification))
	router.Handler(http.MethodPost, "/webhooks", listener.GlobalHandler())

	//router.Handler(http.MethodPost, "/webhooks", http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//
	//})))

	log.Fatal(http.ListenAndServe(":8080", router))
}
