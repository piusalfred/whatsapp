# Webhooks

Receive and dispatch WhatsApp Cloud API webhook notifications with typed handlers.

## 30-Second Setup

```go
handler := webhooks.NewHandler()

handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
        fmt.Printf("From %s: %s\n", req.Info.From, req.Payload.Body)
        return nil
    },
))

listener := webhooks.NewListener(handler, webhooks.ConfigReaderFunc(
    func(r *http.Request) (*webhooks.Config, error) {
        return &webhooks.Config{
            Token:     "your-verify-token",
            AppSecret: "your-app-secret",
        }, nil
    },
))

http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        listener.HandleSubscriptionVerification(w, r)
    case http.MethodPost:
        listener.HandleNotification(w, r)
    }
})
http.ListenAndServe(":8080", nil)
```

## Registering Handlers

Every webhook event type has an `On<Event>` registration method:

```go
// Messages
handler.OnTextMessage(h)
handler.OnImageMessage(h)
handler.OnReactionMessage(h)
handler.OnOrderMessage(h)
handler.OnLocationMessage(h)
handler.OnContactsMessage(h)
handler.OnButtonMessage(h)

// Message delivery statuses
handler.OnMessageStatusChange(h)

// Account & business events
handler.OnBusinessAlertNotification(h)
handler.OnBusinessTemplateStatusUpdate(h)
handler.OnBusinessSecurityNotification(h)

// Groups
handler.OnGroupLifecycleUpdate(h)
handler.OnGroupParticipantsUpdate(h)
handler.OnGroupSettingsUpdate(h)
handler.OnGroupStatusUpdate(h)

// Flows
handler.OnFlowStatusChange(h)

// Calls
handler.OnCallConnect(h)
handler.OnCallCreated(h)
handler.OnCallTerminate(h)
handler.OnCallStatus(h)

// History sync
handler.OnHistorySync(h)
handler.OnHistoryMediaMessages(h)

// SMB
handler.OnSMBMessageEcho(h)
handler.OnSMBAppStateSync(h)

// User preferences
handler.OnUserPreferencesUpdate(h)
```

Each `On*` method accepts a typed interface. Use the `Func` adapter for inline functions:

```go
handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
        // req.Payload is *webhooks.Text
        return nil
    },
))
```

## Handling Unhandled Events — The Fallback

A fallback catches every event that doesn't have a dedicated handler. Always register one to avoid silent drops:

```go
handler.OnFallback(webhooks.FallbackHandlerFunc(
    func(ctx context.Context, ev webhooks.NotificationEvent) error {
        log.Printf("unhandled event: field=%s", ev.Field)
        return nil // acknowledge so WhatsApp doesn't retry
    },
))
```

You can also set per-domain fallbacks for finer control:

```go
// Catches unhandled group fields only
handler.Groups().OnFallback(webhooks.FallbackHandlerFunc(
    func(ctx context.Context, ev webhooks.NotificationEvent) error {
        log.Printf("unhandled group event: field=%s", ev.Field)
        return nil
    },
))
```

## Handling Errors

Errors from your handlers can be fatal (stop processing, WhatsApp retries) or non-fatal (log and continue):

```go
handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
    log.Printf("handler error: %v", err)
    return nil // non-fatal, continue processing
}))
```

Per-domain error handlers override the global one:

```go
handler.Groups().OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
    return err // fatal for group events only
}))
```

## Adding Middleware

Middleware wraps the notification handler and runs in order (`middlewares[0]` outermost):

```go
// Auth middleware — short-circuits before handler runs
authMiddleware := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
    return webhooks.NotificationHandlerFunc(
        func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
            if !isAuthorized(ctx) {
                return &webhooks.Response{StatusCode: http.StatusForbidden}
            }
            return next.HandleNotification(ctx, n)
        },
    )
}

// Logging middleware — wraps around the handler
logMiddleware := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
    return webhooks.NotificationHandlerFunc(
        func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
            start := time.Now()
            resp := next.HandleNotification(ctx, n)
            log.Printf("status=%d duration=%s", resp.StatusCode, time.Since(start))
            return resp
        },
    )
}

listener := webhooks.NewListener(handler, configReader, authMiddleware, logMiddleware)
```

## Wiring a Listener

```go
// Config reader — provides per-request verification and app secret
reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
    return &webhooks.Config{
        Token:           os.Getenv("VERIFY_TOKEN"),
        AppSecret:       os.Getenv("APP_SECRET"),
        ValidatePayload: true,
    }, nil
})

listener := webhooks.NewListener(handler, reader)

// Observe errors (logging/metrics only, can't change response)
listener.OnError(func(ctx context.Context, r *http.Request, err error) {
    log.Printf("listener error: %v", err)
})

// Custom signature verification
listener.SetSignatureVerifier(webhooks.SignatureVerifierFunc(
    func(header http.Header, payload []byte, secret string) error {
        // your custom HMAC comparison
        return nil
    },
))

// Wire into HTTP
http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        listener.HandleSubscriptionVerification(w, r)
    case http.MethodPost:
        listener.HandleNotification(w, r)
    }
})
```

## Dispatch Mode

Control how events within a notification are processed:

```go
// Sequential (default) — one at a time, first error stops processing
handler := webhooks.NewHandler()

// Parallel — all events concurrently, best for latency
handler := webhooks.NewHandler(webhooks.WithMode(webhooks.Parallel))
```

| Mode | Behavior |
|---|---|
| `Sequential` | Events one-by-one. Error → 500. |
| `Parallel` | Events concurrently. Context cancelled → 504. |

## Selective Event Processing

Flatten a notification into individual events for selective async processing:

```go
notif, _ := webhooks.ParseNotification(r, &webhooks.ParseNotificationOptions{...})

for _, ev := range notif.Events() {
    if ev.Field == "messages" {
        go handler.HandleNotificationEvent(context.Background(), ev)
    }
}
```

## Custom Signature Verification

```go
listener.SetSignatureVerifier(webhooks.SignatureVerifierFunc(
    func(header http.Header, payload []byte, secret string) error {
        mac := hmac.New(sha256.New, []byte(secret))
        mac.Write(payload)
        expected := fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
        actual := header.Get("X-Hub-Signature-256")
        if !hmac.Equal([]byte(expected), []byte(actual)) {
            return webhooks.ErrInvalidSignature
        }
        return nil
    },
))
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/piusalfred/whatsapp/webhooks"
)

func main() {
    handler := webhooks.NewHandler(
        webhooks.WithMode(webhooks.Parallel),
    )

    // Text messages
    handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
        func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
            log.Printf("Text from %s: %s", req.Info.From, req.Payload.Body)
            return nil
        },
    ))

    // Catch-all for events without a dedicated handler
    handler.OnFallback(webhooks.FallbackHandlerFunc(
        func(ctx context.Context, ev webhooks.NotificationEvent) error {
            log.Printf("Unhandled: field=%s", ev.Field)
            return nil
        },
    ))

    // Global error handler
    handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
        log.Printf("Error: %v", err)
        return nil
    }))

    reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
        return &webhooks.Config{
            Token:           os.Getenv("VERIFY_TOKEN"),
            AppSecret:       os.Getenv("APP_SECRET"),
            ValidatePayload: true,
        }, nil
    })

    listener := webhooks.NewListener(handler, reader)

    listener.OnError(func(ctx context.Context, r *http.Request, err error) {
        log.Printf("Listener error: %v", err)
    })

    http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            listener.HandleSubscriptionVerification(w, r)
        case http.MethodPost:
            listener.HandleNotification(w, r)
        }
    })

    log.Println("Listening on :8080")
    http.ListenAndServe(":8080", nil)
}
```
