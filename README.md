# whatsapp

[![GoDoc](https://godoc.org/github.com/piusalfred/whatsapp?status.svg)](https://godoc.org/github.com/piusalfred/whatsapp)
[![Go Report Card](https://goreportcard.com/badge/github.com/piusalfred/whatsapp)](https://goreportcard.com/report/github.com/piusalfred/whatsapp)
![Status](https://img.shields.io/badge/status-alpha-red)

A Go client for the [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api) covering outbound messaging, inbound webhooks, and the Business Management API.

> [!IMPORTANT]
> This is a third-party library. Not affiliated with or maintained by Meta.

## Installation

```bash
go get github.com/piusalfred/whatsapp
```

## Packages

| Package | Purpose |
|----------|---------|
| `message` | Send text, media, interactive, location, contacts, reactions, stickers |
| `message/interactive` | Build interactive messages (buttons, lists, flows, carousels) |
| `message/template` | Build template messages (text, media, carousel, auth) |
| `message/media` | Media type helpers |
| `webhooks` | Receive and dispatch inbound notifications (messages, statuses, calls, flows, groups, account alerts) |
| `webhooks/callbacks` | Manage alternate webhook callback URLs |
| `webhooks/router` | HTTP router for webhook endpoints |
| `groups` | Create, delete, manage groups and participants |
| `qrcode` | Create, read, update, delete QR codes |
| `media` | Upload, retrieve, delete, download media |
| `phonenumber` | List, get, configure phone numbers |
| `auth` | System users, tokens, 2FA, app installation |
| `business` | Get and update business profile |
| `business/analytics` | Messaging, conversation, and pricing analytics |
| `conversation/automation` | Conversational components, welcome messages |
| `user` | Block, unblock, list blocked users |
| `uploads` | Chunked upload sessions |
| `settings` | Business settings |
| `calls` | Calling API |
| `flow` | WhatsApp Flows management |
| `config` | Shared configuration types |
| `pkg/http` | HTTP client with middleware, interceptors, and request building |
| `pkg/types` | Shared types (metadata, phone numbers) |

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

### Handle webhooks

See **[webhooks/README.md](./webhooks/README.md)** for the full webhook guide — architecture, dispatch pipeline, domain handlers, mode, fallback cascade, middleware, and error handling.

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

## Client types

Every domain package exposes two client constructors:

- `NewClient(conf)` — holds a fixed `*config.Config`, ideal for single-tenant services.
- `NewBaseClient()` — accepts per-call config via `WithSenderConfig`, ideal for multi-tenant workloads or dynamic credential rotation.

## Testing

Generated mocks are available in [`mocks/`](./mocks/). Each interface has a corresponding mock.

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

## Documentation

- **[docs/README.md](./docs/README.md)** — full project guide
- **[webhooks/README.md](./webhooks/README.md)** — webhook architecture, dispatch, and configuration
- Official [WhatsApp Cloud API Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started)

## Reference

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

**Management APIs**

- [Phone Numbers](https://developers.facebook.com/docs/whatsapp/cloud-api/phone-numbers)
- [QR Codes](https://developers.facebook.com/docs/whatsapp/business-management-api/qr-codes/)
- [System Users](https://developers.facebook.com/docs/marketing-api/system-users/overview)
- [Install Apps, Generate, Refresh, and Revoke Tokens](https://developers.facebook.com/docs/marketing-api/system-users/install-apps-and-generate-tokens/#revoke-token)
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
