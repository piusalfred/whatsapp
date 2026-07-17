# whatsapp

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

A highly configurable Go client for the [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api). It covers outbound messaging (text, media, interactive, templates), inbound webhooks (signature validation, typed event dispatch), and the full Business Management API (groups, QR codes, phone numbers, media, analytics, system users, and more).

Every domain is a self-contained package with its own `Client` (single-tenant) and `BaseClient` (multi-tenant) — use one or compose them all behind the unified `api.Client`. The HTTP transport is generic, typed, and fully mockable, with middleware chains, request/response interceptors, and functional options for every tunable.

> [!IMPORTANT]
> This is a third-party library. Not affiliated with or maintained by Meta.

## Installation

```bash
go get github.com/piusalfred/whatsapp
```

## Supported APIs

| Category | Package | Capabilities |
|----------|---------|-------------|
| **Messages** | `message` | Text, image, video, audio, document, sticker, location, reaction, contacts, pin |
| **Interactive** | `message/interactive` | CTA URL, reply buttons, list picker, flow, media carousel, address, location request, call permission |
| **Templates** | `message/template` | Text, media, carousel, coupon, limited-time offer, authentication |
| **Webhooks** | `webhooks` | Messages, statuses, calls, flows, groups, security, templates, account alerts |
| **Groups** | `groups` | Create, delete, participants, invite links, join requests |
| **QR Codes** | `qrcode` | Create, read, update, delete, list |
| **Media** | `media` | Upload, retrieve, delete, download |
| **Phone Numbers** | `phonenumber` | List, get, settings |
| **Auth** | `auth` | System users (create, list, update), tokens, 2FA, install apps |
| **Business Profile** | `business` | Get, update |
| **Analytics** | `business/analytics` | Messaging, conversation, pricing |
| **Conversation** | `conversation/automation` | Components, welcome messages, bot details |
| **Users** | `user` | Block, unblock, list blocked |
| **Uploads** | `uploads` | Chunked upload sessions |
| **Callbacks** | `webhooks/callbacks` | Alternate webhook URLs |
| **Settings** | `settings` | Business settings |
| **Calls** | `calls` | Calling API |
| **Flows** | `flow` | WhatsApp Flows management |

## Quick Start

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

See more in [examples](./_examples/) and the [full guide](./docs/).

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

Generated mocks for every interface are available in [`mocks/`](./mocks/).

```go
import mockhttp "github.com/piusalfred/whatsapp/mocks/http"

ctrl := gomock.NewController(t)
mockSender := mockhttp.NewMockSender[message.BaseRequest](ctrl)
mockSender.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

client := message.NewClient(conf)
client.SetBaseClient(mockSender)
```

## Development

```bash
make all    # format, lint, generate mocks, run tests with race detector
make help   # list all available targets
```

## Reference Links

**Getting started**

- [WhatsApp Cloud API Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started)
- [Application Dashboard](https://developers.facebook.com/apps/)
- [Postman Collection](https://www.postman.com/meta/whatsapp-business-platform/collection/wlk6lh4/whatsapp-cloud-api)
- [Error Codes](https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes/)
- [Securing Requests](https://developers.facebook.com/docs/graph-api/guides/secure-requests)
- [Graph API Reference](https://developers.facebook.com/docs/graph-api)

**Messaging**

- [Messages Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages)
- [Address Message](https://developers.facebook.com/docs/whatsapp/cloud-api/messages/address-messages) (India only)

**Webhooks**

- [Webhooks Getting Started](https://developers.facebook.com/docs/graph-api/webhooks/getting-started)
- [Webhooks for WhatsApp Business Account](https://developers.facebook.com/docs/graph-api/webhooks/getting-started/webhooks-for-whatsapp)
- [Notification Payload Reference](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components)
- [Webhooks Override](https://developers.facebook.com/docs/whatsapp/embedded-signup/webhooks/override)

**Flows**

- [WhatsApp Flows](https://developers.facebook.com/docs/whatsapp/flows/)
- [Flows Guide](https://developers.facebook.com/docs/whatsapp/flows/guides)
- [Flows Reference](https://developers.facebook.com/docs/whatsapp/flows/reference/)
- [FlowJSON](https://developers.facebook.com/docs/whatsapp/flows/reference/flowjson)
- [Flows Best Practices](https://developers.facebook.com/docs/whatsapp/flows/guides/bestpractices)
- [Flows Webhooks](https://developers.facebook.com/docs/whatsapp/flows/reference/flowswebhooks)
- [Flow Encryption](https://developers.facebook.com/docs/whatsapp/cloud-api/reference/whatsapp-business-encryption)

**Management APIs**

- [Phone Numbers](https://developers.facebook.com/docs/whatsapp/cloud-api/phone-numbers)
- [QR Codes](https://developers.facebook.com/docs/whatsapp/business-management-api/qr-codes/)
- [System Users](https://developers.facebook.com/docs/marketing-api/system-users/overview)
- [Install Apps, Generate, Refresh, and Revoke Tokens](https://developers.facebook.com/docs/marketing-api/system-users/install-apps-and-generate-tokens/#revoke-token)
- [Create, Retrieve and Update a System User](https://developers.facebook.com/docs/marketing-api/system-users/create-retrieve-update)
- [Access Token Debugger](https://developers.facebook.com/tools/accesstoken/)
- [Analytics](https://developers.facebook.com/docs/whatsapp/business-management-api/analytics#analytics-parameters)
- [Conversational Components](https://developers.facebook.com/docs/whatsapp/cloud-api/phone-numbers/conversational-components)
- [Groups API](https://developers.facebook.com/docs/whatsapp/cloud-api/groups/getting-started)
- [Calling API Settings](https://developers.facebook.com/docs/whatsapp/cloud-api/calling/call-settings)

**Reference**

- [WhatsApp Business Platform Documentation](https://developers.facebook.com/docs/whatsapp)
- [WhatsApp Business Account Graph API Reference](https://developers.facebook.com/docs/graph-api/reference/whats-app-business-account/)

## Videos

- [Get Started with WhatsApp Business Calling API](https://www.youtube.com/watch?v=SRDjj3KAMIE)
- [Building end-to-end Experiences with the WhatsApp Business Platform](https://www.youtube.com/watch?v=KP6_BUw3i0U)
