# WhatsApp Cloud API Go Client — Guide

## Quick Start

### Installation

```bash
go get github.com/piusalfred/whatsapp
```

### Credentials

From your [Facebook Developer App](https://developers.facebook.com/apps) dashboard, under **WhatsApp > API Setup**:

| You need | Config field | Env var example |
|----------|-------------|-----------------|
| Access token | `AccessToken` | `WHATSAPP_TOKEN` |
| Phone number ID | `PhoneNumberID` | `WHATSAPP_PHONE_NUMBER_ID` |
| Business account ID | `BusinessAccountID` | `WHATSAPP_BUSINESS_ID` |
| App secret (optional, for secure requests) | `AppSecret` | `WHATSAPP_APP_SECRET` |

### Send a Message

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

All message types follow the same pattern — `client.SendXxxMessage(ctx, message.SendTo(to), payload)`:

```go
// Interactive list
resp, _ := client.SendInteractiveMessage(ctx, message.SendTo("+123"),
    interactive.List(&interactive.ListRequest{...}))

// Template
resp, _ := client.SendTemplateMessage(ctx, message.SendTo("+123"),
    template.NewInteractiveTemplate("hello_world", &template.Language{Code: "en_US"}, nil, nil, nil))

// Image, audio, video, document, sticker, location, reaction, contacts...
// All follow: client.SendXxxMessage(ctx, message.SendTo(to), payload)
```

### Mark as Read

```go
client.UpdateMessageStatus(ctx, &message.StatusUpdateRequest{
    MessageID: "wamid.xxx",
    Status:    message.StatusRead,
})
```

### Receive Webhooks

```go
listener := webhooks.NewListener(
    webhooks.NewHandler(),
    webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
        return &webhooks.Config{Token: "my-token", AppSecret: "my-secret", Validate: true}, nil
    }),
)

http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        listener.HandleSubscriptionVerification(w, r)
    case http.MethodPost:
        listener.HandleNotification(w, r)
    }
})
```

---

## How the Library Is Organized

The library follows a few principles consistently:

**Every domain has its own package** — `message`, `groups`, `media`, `webhooks`, `auth`, `qrcode`, etc. Each is self-contained and can be used independently.

**Two client variants per package** — `Client` holds a fixed configuration (single-tenant services), `BaseClient` accepts configuration per call (multi-tenant or credential-rotation workloads).

**HTTP stays in `pkg/http`** — domain packages never touch `*http.Request` or `*http.Response`. The transport layer is generic, typed, and testable via the `Sender[T]` interface. Middleware, request/response interceptors, and body encoding/decoding are all pluggable.

**The `api` package unifies everything** — if you need multiple APIs from one place, `api.NewClient(conf)` lazily initializes all 15 sub-clients on first use.

```
User code
   ↓
api.Client / message.Client / groups.Client  ...   ← domain packages
   ↓
whttp.BaseClient[T]  →  CoreClient[T]              ← transport layer
   ↓
net/http
```

A message send flows through:

```
SendTextMessage(ctx, si, text)
  → build BaseRequest (domain struct → JSON wire format)
  → RequestBuilder.Auth(conf.AuthConfig())
  → BuildRequest(builder, body)
  → CoreClient.Send(ctx, req, decoder)
      → request interceptor (snapshot body, call hook, restore)
      → http.Client.Do
      → response interceptor (snapshot body, call hook, restore)
      → decoder.Decode
```

---

## Single-Tenant vs Multi-Tenant

**Single-tenant** — one phone number, known at startup:

```go
client := message.NewClient(conf)
resp, _ := client.SendTextMessage(ctx, message.SendTo("+123"), &message.Text{Body: "hi"})
```

**Multi-tenant** — many phone numbers, dynamic credentials:

```go
base := message.NewBaseClient()
resp, _ := base.SendTextMessage(ctx, tenantConf, message.SendTo("+123"), &message.Text{Body: "hi"})
```

**Unified** — need multiple APIs, single tenant:

```go
client := api.NewClient(conf)
client.SendMessage(ctx, message.New(message.SendTo("+123"), message.WithTextMessage(...)))
client.CreateGroup(ctx, &groups.CreateGroupRequest{Name: "Team Chat"})
```

---

## Testing

Every domain client exposes `SetBaseClient(sender Sender[T])` for mock injection. Generated mocks live in `mocks/`:

```go
import mockhttp "github.com/piusalfred/whatsapp/mocks/http"

func TestMyService(t *testing.T) {
    ctrl := gomock.NewController(t)
    mockSender := mockhttp.NewMockSender[message.BaseRequest](ctrl)
    mockSender.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

    client := message.NewClient(conf)
    client.SetBaseClient(mockSender)
    // calls use the mock
}
```

---

## Middleware

Attach cross-cutting concerns (logging, metrics, retries) as middleware — never inside business logic:

```go
loggingMW := func(next whttp.SenderFunc[message.BaseRequest]) whttp.SenderFunc[message.BaseRequest] {
    return func(ctx context.Context, req *whttp.Request[message.BaseRequest], dec whttp.ResponseDecoder) error {
        start := time.Now()
        err := next(ctx, req, dec)
        log.Printf("request took %v, err=%v", time.Since(start), err)
        return err
    }
}
client.SetMiddlewares(loggingMW, metricsMW)
```

Middlewares are applied inside-out: `middlewares[0]` runs outermost (first on the way in, last on the way out).

---

## Secure Requests

Enable appsecret_proof by setting `SecureRequests` and providing your app secret:

```go
conf := &config.Config{
    SecureRequests: true,
    AppSecret:      os.Getenv("WHATSAPP_APP_SECRET"),
}
```

The proof is automatically appended to every request URL as `?appsecret_proof=...`.

---

## Validate Config Early

Fail fast at startup — don't discover missing credentials mid-request:

```go
conf, err := config.ReadValidate(ctx, config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
    return &config.Config{...}, nil
}))
if err != nil {
    log.Fatal("invalid config:", err)
}
```

---

## Things to Watch Out For

### Don't return errors from webhook handlers for permanent failures

WhatsApp retries non-200 responses for up to 7 days. If your handler returns an error for something that will never succeed (unsupported message type, invalid payload), WhatsApp retries forever. Only return errors for transient failures (DB down, network timeout). For permanent failures, log and return nil.

```go
// ❌ WhatsApp retries this for 7 days
handler.OnTextMessage(func(...) error {
    return fmt.Errorf("unsupported message")
})

// ✅ Log and move on
handler.OnTextMessage(func(...) error {
    if !isSupported(msg) {
        log.Printf("skipping unsupported message %s", info.MessageID)
        return nil
    }
    return processMessage(msg)
})
```

### Don't use the webhook context for background work

The `ctx` passed to webhook handlers is cancelled after the HTTP response is written. If you spawn a goroutine that uses this context, it will fail:

```go
// ❌ ctx is cancelled after the handler returns
go func() { sendReply(ctx, ...) }()

// ✅ Use a fresh context
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    sendReply(ctx, ...)
}()
```

### Always construct clients with NewClient or NewBaseClient

A zero-value `Client` has a nil sender — calling any method panics:

```go
// ❌ PANIC
var client message.Client
client.SendTextMessage(...)

// ✅
client := message.NewClient(conf)
```

### Set middlewares before any Send calls

`SetMiddlewares` and `SetSender` write to a shared field without synchronization. Call them at startup, before any goroutines start sending:

```go
client := message.NewClient(conf)
client.SetMiddlewares(loggingMW)  // ✅ setup phase
go startServer(client)             // ✅ concurrent use
```

### SendTo, not a raw phone number

The typed helpers take `*SendInfo`, not a raw string:

```go
// ❌
client.SendTextMessage(ctx, "+16505551234", &message.Text{Body: "hi"})

// ✅
client.SendTextMessage(ctx, message.SendTo("+16505551234"), &message.Text{Body: "hi"})
```

### `pkg/http` shadows `net/http` — use an alias

```go
import whttp "github.com/piusalfred/whatsapp/pkg/http"
```

### Unrecognized webhook fields are intentionally ignored

If WhatsApp adds a new notification type that the library doesn't handle yet, the handler returns 200 (not an error). This prevents retry storms. If you need an unsupported field, open an issue.

### `config.Config` has no defaults — every required field must be set

`BaseURL`, `APIVersion`, `AccessToken`, and `PhoneNumberID` are all required. Use `config.Validate()` to catch missing fields.

### Large webhook payloads may be rejected

The listener enforces a 4 MB payload limit. If you receive many events in one notification, configure your WhatsApp webhook settings to use smaller batches.

### sync.Once fragility in api.BaseClient

If a sub-client initialization panics inside `sync.Once`, that sub-client is permanently nil. Future calls panic. Validate all options before creating the client.
