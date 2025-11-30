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
- [User Management](./user)
  - [Block Users](./user)
  - [Unblock Users](./user)
  - [Get Blocked Users](./user)
- [Conversation Automation](./conversation/automation)
- [File Uploads](./uploads)
- [Call Settings](./settings)


## Initial Steps
Start by reading the official [WhatsApp Cloud API Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started) then go to [Get Started Guide](./docs/README.md) 
for initial steps in setting up your developing environment.

> [!NOTE]
> You will find `BaseClient` and `Client`.
> `Client` provides a stateful approach, reusing the same configuration across multiple requests until manually refreshed, making it ideal for long-running services where consistency and thread safety are required.
> `BaseClient` is stateless, reloading the configuration on each request, making it more flexible for dynamic environments like multi-tenancy, where different configurations may be needed for each request.


### Usage
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
	"net/http"
	"os"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const recipient = "XXXXXXXXXXXXX" // Placeholder for recipient number

func main() {
	ctx := context.Background()

	coreClient := whttp.NewSender[message.Message]()
	coreClient.SetHTTPClient(http.DefaultClient)

	reader := config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
		// TODO: Replace with your config reader implementation
		conf := &config.Config{
			BaseURL:           "",
			APIVersion:        "",
			AccessToken:       "",
			PhoneNumberID:     "",
			BusinessAccountID: "",
			AppSecret:         "",
			AppID:             "",
			SecureRequests:    false,
		}

		return conf, nil
	})

	baseClient, err := message.NewBaseClient(coreClient, reader)
	if err != nil {
		fmt.Printf("error creating base client: %v\n", err)
		os.Exit(1)
	}

	initTmpl := message.WithTemplateMessage(&message.Template{
		Name: "hello_world",
		Language: &message.TemplateLanguage{
			Code: "en_US",
		},
	})

	initTmplMessage, err := message.New(recipient, initTmpl)
	if err != nil {
		fmt.Printf("error creating initial template message: %v\n", err)
		os.Exit(1)
	}

	response, err := baseClient.SendMessage(ctx, initTmplMessage)
	if err != nil {
		fmt.Printf("error sending initial template message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("response: %+v\n", response)
}
```

See more in [examples](./_examples/) and [docs](./docs/README.md)


## Testing
There is provision of [**mocks**](./mocks) that may come handy in testing.

## Extras
The extras package contains some useful utilities for working with this library. It is experimental and may change in future releases.
- [OpenTelemetry Adapter](./extras/otel) provides OpenTelemetry instrumentation for tracing and monitoring sending and receiving whatsapp messages.
- [Model Context Protocol](./extras/mcp) a simple implementation of the Model Context Protocol (MCP) server for sending whatapp messages.


## Development
After making some changes run `make all` to format and test the code. You can also run `make help` to see other available commands

## Documentation Links
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
- [Calling API Settings](https://developers.facebook.com/docs/whatsapp/cloud-api/calling/call-settings)
- [Address Message—Currently Supported only in India](https://developers.facebook.com/docs/whatsapp/cloud-api/messages/address-messages)
- [Get Started with Groups API](https://developers.facebook.com/docs/whatsapp/cloud-api/groups/getting-started)


## Video Links 
- [Get Started with Whatsapp Business Calling API](https://www.youtube.com/watch?v=SRDjj3KAMIE) 
- [Building end-to-end Experiences with the WhatsApp Business Platform](https://www.youtube.com/watch?v=KP6_BUw3i0U)