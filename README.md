# whatsapp

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

A highly configurable golang client for [Whatsapp Cloud API](https://www.postman.com/meta/whatsapp-business-platform/collection/wlk6lh4/whatsapp-cloud-api)

> [!IMPORTANT]  
> This is the third-party library and not the official one. Not affiliated nor maintained by Meta.


## Supported API

- [Message](./message)
  - [Text](./message)
  - [Media](./message)
  - [Templates](./message)
  - [Interactive Messages](./message)
  - [Replies and Reactions](./message)
- [QR Code Management](./qrcode)
- [Phone Number Management](./phonenumber)
  - [Get Phone Number Information](./phonenumber)
  - [Update Phone Number](./phonenumber)
- [Media Management](./media)
- [Webhooks](./webhooks)
  - [Message Webhooks](./webhooks/message)
  - [Business Management Webhooks](./webhooks/business)
  - [Flow Management Webhooks](./webhooks/flow)
- [User Management](./user)
  - [Block Users](./user)
  - [Unblock Users](./user)
  - [Get Blocked Users](./user)
- [Conversation Automation](./conversation/automation)


## setup

- Go to [apps](https://developers.facebook.com/apps) you have registered click on the one you want to develop API for and go to `API setup` if not create one.
- Get these details `access token`,`phone number id`,`business account id` and make sure you have authorized one or more whatsapp phone numbers to receive these messages
- While on the root directory of the project run `task build-examples` to build all the examples that will be in  [examples/bin](./examples/bin) directory
- Then run `cp examples/api.env examples/bin/api.env`
- And populate the values you have got to this `api.env` file, The given values here are not valid
```dotenv
WHATSAPP_CLOUD_API_BASE_URL=https://graph.facebook.com
WHATSAPP_CLOUD_API_API_VERSION=v20.0
WHATSAPP_CLOUD_API_ACCESS_TOKEN=EAALLrT0ok6UBJZB0ZB3gvzk9hJaEjGM8ISZAxPR5e3ZAFn4RmBIThoeK0XOdbKv8y2zB3YQ7uaijShZBjVIcZD
WHATSAPP_CLOUD_API_PHONE_NUMBER_ID=111112271552333
WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID=222222508300000
WHATSAPP_CLOUD_API_TEST_RECIPIENT=+XXXX7374665453
```
- now you can navigate to the  [examples/bin](./examples/bin) directory and run the examples. But before that make sure you have sent the hello world
template message to the recipient and **replied** and this should be within *24 hrs window*

> [!NOTE]
> You will find `BaseClient` and `Client`.
> `Client` provides a stateful approach, reusing the same configuration across multiple requests until manually refreshed, making it ideal for long-running services where consistency and thread safety are required.
> `BaseClient` is stateless, reloading the configuration on each request, making it more flexible for dynamic environments like multi-tenancy, where different configurations may be needed for each request.


### usage
Install the library by running

```bash
go get github.com/piusalfred/whatsapp
```

> [!NOTE]
> The webhooks and messaging clients are separated to allow for different configurations and use cases. The webhooks client is designed to handle incoming notifications,
> while the messaging client is focused on sending messages.

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
    "log"
    "log/slog"
    "net/http"
    
    "github.com/piusalfred/whatsapp/message"
    "github.com/piusalfred/whatsapp/webhooks"
)

// LoggingMiddleware is an example middleware that logs
// before and after handling a single incoming webhook request.
func LoggingMiddleware(next webhooks.NotificationHandlerFunc) webhooks.NotificationHandlerFunc {
  return func(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
	  fmt.Println("[LoggingMiddleware] --> Before handling notification")
	  response := next(ctx, notification)
	  fmt.Println("[LoggingMiddleware] <-- After handling notification")
	  return response
  }
}

// ReactionHandler is an example of a more advanced handler type
// that implements logic for dealing with Reaction messages.
type ReactionHandler struct {
    Logger *slog.Logger
    Store  map[string]any
}

// Ensure that ReactionHandler satisfies the interface for reaction messages.
// This is optional, but helpful if you want static type checking.
var _ webhooks.ReactionHandler = (*ReactionHandler)(nil)

// Handle processes reaction messages (type: reaction).
// It logs the incoming reaction and stores the emoji in a map.
func (r *ReactionHandler) Handle(
        ctx context.Context,
        nctx *webhooks.MessageNotificationContext,
        mctx *webhooks.MessageInfo,
        reaction *message.Reaction,
) error {
    r.Logger.Info("Received reaction message",
      "context", nctx,
      "message_info", mctx,
      "reaction", reaction,
    )
    
    r.Logger.Info("Reaction emoji", "emoji", reaction.Emoji)
    
    // For example, you can store the emoji in an in-memory map keyed by the message context ID.
    if r.Store == nil {
      r.Store = make(map[string]any)
    }
    if mctx.Context != nil {
      r.Store[mctx.Context.ID] = reaction.Emoji
    }
    
    return nil
}

func main() {
    // Create a Handler that can process various types of webhook notifications.
    // This initializes no-op handlers for all message types. The no-op handlers
    // does nothing and returns nil.
    handler := webhooks.NewHandler()
    
    // There are two ways to register handlers for specific message types:
    // 1. A simple callback function for a specific message type.
    // 2. Implementing a specific handler interface (e.g., webhooks.ReactionHandler) for more complex logic.
    
    // Register a simple callback for text messages using OnTextMessage. This will replace the
    // no-op handler that was initialized by called webhooks.NewHandler().
    // This callback will be invoked whenever we receive a "type": "text" message.
    handler.OnTextMessage(func(ctx context.Context,
            nctx *webhooks.MessageNotificationContext,
            mctx *webhooks.MessageInfo,
            text *webhooks.Text,
    ) error {
      fmt.Printf("[OnTextMessage] New text message received:\n")
      fmt.Printf("  Notification context: %+v\n", nctx)
      fmt.Printf("  Message info:         %+v\n", mctx)
      fmt.Printf("  Text content:         %+v\n", text)
      return nil
    })
    
    // In case of complex handlers you can implement a certain notification type handler interface
    // for example for reaction messages you can implement the webhooks.ReactionHandler interface
    // which is an alias for MessageHandler[message.Reaction]
    reactionHandler := &ReactionHandler{
      Logger: slog.Default(),
      Store:  make(map[string]any),
    }
    handler.SetReactionMessageHandler(reactionHandler)
    messageListener := webhooks.NewListener(
      handler.HandleNotification, // The core Handler’s method
      func(ctx context.Context) (string, error) { // VerifyTokenReader: returns your verify token
        return "TOKEN", nil
      },
      &webhooks.ValidateOptions{
        Validate:  false, // Skip signature validation for simplicity
        AppSecret: "SUPERSECRET",
      },
      LoggingMiddleware, // Example middleware
    )
    
    http.HandleFunc("POST /webhooks/messages", messageListener.HandleNotification)
    http.HandleFunc("GET /webhooks/messages", messageListener.HandleSubscriptionVerification)
    
    // Start an HTTP server on port :8080 to listen for incoming requests.
    fmt.Println("[main] Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

#### block users from sending messages to your business account

```go
        reader, recipient := LoadConfigFromFile("api.env")
	coreClient1 := whttp.NewSender[user.BlockBaseRequest]()
	blocker := user.NewBlockClient(reader, coreClient1)
	resp, err := blocker.Block(ctx, []string{"1234567890"})
	if err != nil {
		return 
	}
	fmt.Println(resp)
```

See more in [examples](./examples/)


## Testing
There is provision of [**mocks**](./mocks) that may come handy in testing.

## extras
The extras package contains some useful utilities for working with this library. It is experimental and may change in future releases.

## Links
- [Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started)
- [Postman Collection](https://www.postman.com/meta/whatsapp-business-platform/collection/wlk6lh4/whatsapp-cloud-api)
- [Webhooks Get Started](https://developers.facebook.com/docs/graph-api/webhooks/getting-started)
- [Messages Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages)
- [Phone Numbers Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api/phone-numbers)
- [Whatsapp Flows](https://developers.facebook.com/docs/whatsapp/flows/)
- [Whatsapp Flows Guide](https://developers.facebook.com/docs/whatsapp/flows/guides)
- [Flows Reference](https://developers.facebook.com/docs/whatsapp/flows/reference/)
- [QR Codes Documentation](https://developers.facebook.com/docs/whatsapp/business-management-api/qr-codes/)
- [Flows Best Practices](https://developers.facebook.com/docs/whatsapp/flows/guides/bestpractices)
- [FlowJSON](https://developers.facebook.com/docs/whatsapp/flows/reference/flowjson)
- [Error Codes](https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes/)
- [Application Dashboard](https://developers.facebook.com/apps/)
- [Webhooks for Whatsapp Business Account](https://developers.facebook.com/docs/graph-api/webhooks/getting-started/webhooks-for-whatsapp)
- [Webhooks Notification Payload Reference](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components)
- [System Users (Important in Tokens Management)](https://developers.facebook.com/docs/marketing-api/system-users/overview)
- [Install Apps, Generate, Refresh, and Revoke Tokens](https://developers.facebook.com/docs/marketing-api/system-users/install-apps-and-generate-tokens/#revoke-token)
- [Create, Retrieve and Update a System User](https://developers.facebook.com/docs/marketing-api/system-users/create-retrieve-update)
- [Access Token Debugger](https://developers.facebook.com/tools/accesstoken/)
- [Securing Requests](https://developers.facebook.com/docs/graph-api/guides/secure-requests)
- [Analytics](https://developers.facebook.com/docs/whatsapp/business-management-api/analytics#analytics-parameters)
- [Conversational Components](https://developers.facebook.com/docs/whatsapp/cloud-api/phone-numbers/conversational-components)
- [Graph API Reference](https://developers.facebook.com/docs/graph-api)
- [Flows Webhooks](https://developers.facebook.com/docs/whatsapp/flows/reference/flowswebhooks)
- [Whatsapp Business Platform Documentation](https://developers.facebook.com/docs/whatsapp)
- [Flow Encryption](https://developers.facebook.com/docs/whatsapp/cloud-api/reference/whatsapp-business-encryption)
- [Whatsapp Business Account Graph API Reference](https://developers.facebook.com/docs/graph-api/reference/whats-app-business-account/)
- [Webhooks Override](https://developers.facebook.com/docs/whatsapp/embedded-signup/webhooks/override)