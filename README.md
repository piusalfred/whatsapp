# whatsapp

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

A highly configurable golang client for [Whatsapp Cloud API](https://www.postman.com/meta/whatsapp-business-platform/collection/wlk6lh4/whatsapp-cloud-api)

> [!IMPORTANT]  
> This is the third-party library and not the official one. Not affiliated nor maintained by Meta.


## Supported API

- **Messages** — text, image, video, audio, document, sticker, location, reaction, contacts, pin
  - **Interactive** — CTA URL, reply buttons, list picker, flow, media carousel, address, location request, call permission
  - **Templates** — text, media, carousel, coupon, limited-time offer, authentication
- **QR Code Management** — create, read, update, delete, list
- **Phone Number Management** — list, get, settings
- **Media Management** — upload, retrieve, delete, download
- **Webhooks** — messages, statuses, calls, flows, groups, security, templates, account alerts
- **User Management** — block, unblock, list blocked
- **Conversation Automation** — components, welcome messages, bot details
- **Groups** — create, delete, manage participants, invite links, join requests
- **Business Profile** — get, update
- **Analytics** — messaging, conversation, pricing
- **Auth (System Users)** — create, list, update, tokens, 2FA
- **Uploads** — chunked upload sessions
- **Callbacks** — alternate webhook URLs
- **Settings** — business settings


## Quick Start

```bash
go get github.com/piusalfred/whatsapp
```

### Send a text message

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
)

func main() {
	conf := &config.Config{
		BaseURL:       "https://graph.facebook.com",
		APIVersion:    "v22.0",
		AccessToken:   os.Getenv("WHATSAPP_TOKEN"),
		PhoneNumberID: os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
	}

	client := message.NewClient(conf)

	resp, err := client.SendTextMessage(
		context.Background(),
		message.SendTo("+16505551234"),
		&message.Text{Body: "Hello from Go!"},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Message ID:", resp.Messages[0].ID)
}
```

### Send an interactive list

```go
import (
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/interactive"
)

resp, err := client.SendInteractiveMessage(ctx, message.SendTo("+16505551234"),
	interactive.List(&interactive.ListRequest{
		Body:   "Which shipping option do you prefer?",
		Button: "Shipping Options",
		Sections: []*interactive.Section{{
			Title: "I want it ASAP!",
			Rows: []*interactive.SectionRow{
				{ID: "priority_express", Title: "Priority Mail Express", Description: "Next Day to 2 Days"},
			},
		}},
	}),
)
```

### Send a template

```go
import "github.com/piusalfred/whatsapp/message/template"

tmpl := template.NewInteractiveTemplate("hello_world",
	&template.Language{Code: "en_US"},
	nil, nil, nil,
)
resp, err := client.SendTemplateMessage(ctx, message.SendTo("+16505551234"), tmpl)
```

### Mark a message as read

```go
resp, err := client.UpdateMessageStatus(ctx, &message.StatusUpdateRequest{
	MessageID: "wamid.xxx",
	Status:    message.StatusRead,
})
```

### Unified client (all APIs)

```go
import "github.com/piusalfred/whatsapp/api"

client := api.NewClient(conf)

// Messages
client.SendMessage(ctx, message.New(
	message.SendTo("+16505551234"),
	message.WithTextMessage(&message.Text{Body: "Hello"}),
))

// Groups
client.CreateGroup(ctx, &groups.CreateGroupRequest{Name: "Team Chat"})

// QR Codes
client.CreateQR(ctx, &qrcode.CreateRequest{PrefilledMessage: "Hi"})

// Media
client.UploadMedia(ctx, &media.UploadRequest{...})

// System users
client.CreateSystemUser(ctx, &auth.CreateSystemUserRequest{Name: "bot"})
```

See more in [examples](./_examples/) and [docs](./docs/)

> [!NOTE]
> Every domain package exposes both `Client` and `BaseClient`.
> `Client` holds a fixed `*config.Config` — ideal for single-tenant services.
> `BaseClient` accepts a per-call config — ideal for multi-tenant workloads
> or dynamic credential rotation.

> [!NOTE]
> The [webhooks](./webhooks) package is an HTTP server that receives inbound
> notifications from WhatsApp (messages, statuses, calls, flows, groups,
> templates, account alerts). It validates signatures and dispatches events
> to your handlers.
> The [message](./message) package is the outbound client for sending messages.
> They serve opposite directions and are configured independently.


## Documentation

Read the full guide at **[docs/README.md](./docs/README.md)** — it covers quick start, architecture, testing, middleware, secure requests, and things to watch out for.

Start by reading the official [WhatsApp Cloud API Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started).


## Testing
There is provision of [**mocks**](./mocks) that may come handy in testing.

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
