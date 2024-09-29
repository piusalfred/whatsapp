# WhatsApp Cloud API Go Client

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

**Note:** This library is currently in alpha and not yet stable. Breaking changes may occur.

## Supported API

- [**Message**](./message)
    - Text
    - Media
    - Templates
    - Interactive Messages
    - Replies and Reactions

- [**QR Code Management**](./qrcode)
    - Generate QR Code
    - Retrieve QR Code

- [**Phone Number Management**](./phonenumber)
    - Get Phone Number Information
    - Update Phone Number

- [**Media Management**](./media)
    - Upload Media
    - Download Media
    - Delete Media

- [**Webhooks**](./webhooks)
    - Message Webhooks
    - Business Management Webhooks
    - Flow Management Webhooks


## examples

You will find `BaseClient` and `Client`.

`Client` provides a stateful approach, reusing the same configuration across multiple requests until manually refreshed, making it ideal for long-running services where consistency and thread safety are required.

`BaseClient` is stateless, reloading the configuration on each request, making it more flexible for dynamic environments like multi-tenancy, where different configurations may be needed for each request.

See more in [examples](./examples/)

Install the library by running

```bash
go get github.com/piusalfred/whatsapp
```

### messages

```go
package main

import (
	"context"
	"fmt"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"log"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
)

func main() {
	// Create the HTTP client
	httpClient := &http.Client{}

	// Define a configuration reader function
	configReader := config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
		return &config.Config{
			BaseURL:           "https://your-whatsapp-api-url",   // Replace with your API URL
			AccessToken:       "your-access-token",              // Replace with your access token
			APIVersion:        "v14.0",                          // WhatsApp API version
			BusinessAccountID: "your-business-account-id",       // Replace with your business account ID
		}, nil
	})

	clientOptions := []whttp.CoreClientOption[message.Message]{
		whttp.WithCoreClientHTTPClient[message.Message](httpClient),
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
	}

	coreClient := whttp.NewSender[message.Message](clientOptions...)
	client, err := message.NewBaseClient(coreClient, configReader)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Define the recipient's WhatsApp phone number (including country code)
	recipient := "1234567890"

	// Create a new text message request
	textMessage := message.NewRequest(recipient, &message.Text{
		Body: "Hello, this is a test message!",
	}, "")

	// Send the text message
	ctx := context.Background()
	response, err := client.SendText(ctx, textMessage)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	// Print the response from the API
	fmt.Printf("Message sent successfully! Response: %+v\n", response)
}
```

### webhooks

```go

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks/flow"

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
```