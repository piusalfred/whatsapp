# Webhooks Package

Type-safe, handler-based webhook receiver for the [WhatsApp Cloud API](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks).

## Design Decisions

### Handler Registration Pattern (`On*`)

Every handler uses an `On<Event>` registration method. This keeps registration declarative
and lets the compiler verify type safety through generics:

```go
handler := webhooks.NewHandler()
handler.OnTextMessage(myTextHandler)       // MessageHandler[Text]
handler.OnAlerts(myAlertHandler)           // BusinessEventHandler[AlertNotification]
handler.OnFlowStatusChange(myFlowHandler)  // FlowEventHandler[StatusChangeDetails]
```

### Type Aliases

The package exports 42 type aliases (e.g. `TextMessageHandler = MessageHandler[Text]`).
These exist solely to make `On*` method signatures self-documenting. Without them,
registrations would read `handler.OnTextMessage(h MessageHandler[Text])` — technically
equivalent but noisier. The trade-off is a larger godoc surface.

### Three-Layer Fallback Architecture

The fallback cascade has three layers that catch unhandled webhook events at
increasing levels of generality. Every layer that returns `nil` silently
acknowledges the event (HTTP 200) so WhatsApp does not retry.

```
┌─────────────────────────────────────────────────┐
│ Layer 1 — Unknown field                          │
│ change.Field is NOT in the implemented set       │
│ → handler.Fallback (Handler.OnFallback)          │
└────────────────┬────────────────────────────────┘
                 │ field IS implemented
┌────────────────▼────────────────────────────────┐
│ Layer 2 — Nil sub-handler                        │
│ handler.<domain> is nil (no On* called yet)      │
│ → handler.Fallback (same Handler.OnFallback)     │
└────────────────┬────────────────────────────────┘
                 │ sub-handler is non-nil
┌────────────────▼────────────────────────────────┐
│ Layer 3 — Unhandled sub-field                    │
│ Sub-handler exists but specific field is nil     │
│ → <domain>.Fallback (OnFallback on sub-handler)  │
└────────────────┬────────────────────────────────┘
                 │ no sub-fallback set
               HTTP 200 (silent)
```

Each sub-handler (`FlowNotificationHandler`, `BusinessNotificationHandler`,
`GroupManagementHandler`, `MessagesHandler`, `HistoryHandler`) has its own
`Fallback` field that catches unhandled fields **within** its domain without
affecting other domains.

Calling `Handler.OnFallback` **propagates** the fallback to every non-nil
sub-handler that does not already have its own `Fallback` set. This means you
can register a single catch-all and it will work for all domains, or you can
set per-domain fallbacks for finer control.

**Concrete example — Group Management:**

```go
h := webhooks.NewHandler()

// Layer 3: Set a fallback on the Groups sub-handler for unhandled group fields.
h.Groups().OnFallback(webhooks.FallbackHandlerFunc(
    func(ctx context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
        log.Printf("unhandled group field: %s", c.Field)
        return nil // acknowledge, don't error
    },
))

// Register a dedicated handler for group lifecycle events.
h.OnGroupLifecycleUpdate(webhooks.ChangeValueHandlerFunc[webhooks.Group](
    func(ctx context.Context, req *webhooks.ChangeValueRequest[webhooks.Group]) error {
        for _, g := range req.Payload {
            log.Printf("group %s: %s", g.GroupID, g.Type)
        }
        return nil
    },
))

// Now:
//   group_lifecycle_update → dedicated handler fires (layer 3 match)
//   group_status_update   → Groups.Fallback fires (layer 3 fallthrough)
//   account_alerts        → handler.Fallback fires (layer 1, unknown field)
//
// If no Groups.OnFallback had been set, group_status_update would
// silently return 200 (the layer 3 default).
```

**Layer-by-layer behavior:**

| Scenario | Layer | Outcome |
|---|---|---|
| Field not in implemented set, no fallback | 1 | HTTP 200 (silent) |
| Field not in implemented set, handler.OnFallback set | 1 | Fallback invoked |
| Sub-handler nil (no On* called for that domain) | 2 | Same as Layer 1 |
| Sub-handler exists, dedicated handler set | 3 | Dedicated handler invoked |
| Sub-handler exists, dedicated handler nil, sub-fallback set | 3 | Sub-fallback invoked |
| Sub-handler exists, dedicated handler nil, no sub-fallback | 3 | HTTP 200 (silent) |

The public accessors `Handler.Messages()`, `Handler.Flows()`,
`Handler.Business()`, `Handler.Groups()`, and `Handler.History()` lazily
initialise their sub-handler on first call and return it for direct
configuration. Use them when you need to set a sub-handler `Fallback`,
`ErrorHandler`, or nested fields (`Media`, `Interactive`) before any
dedicated `On*` registration.

> **Messages dispatch is different.** The `"messages"` webhook field carries
> three kinds of payload in a single change: notification errors, message
> status updates, *and* incoming messages. `Handler.OnNotificationErrors`
> and `Handler.OnMessageStatusChange` are registered directly on `Handler`
> (not on `MessagesHandler`), so they fire even when `handler.messages` is
> nil. The nil-guard for layer 2 only applies to the messages loop — status
> and error handlers always run when the `"messages"` field arrives.

### Error Routing with `handleError`

Every sub-handler has a private `handleError(ctx, err)` method that routes user
handler errors through its `ErrorHandler`. When `ErrorHandler` is nil, errors
pass through unchanged. When set, it decides whether the error is fatal or
non-fatal:

```go
handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
    log.Printf("non-fatal: %v", err)
    return nil // continue processing
}))
```

The central `Handler` also has a `handleError` that routes through its own
`errorHandler`. Sub-handler `handleError` methods were added to eliminate ~20
repeated boilerplate blocks.

### Concurrency

All handlers are safe for concurrent reads but not concurrent writes.
Register all handlers during initialisation, then treat them as immutable.

### History Webhook

The history webhook arrives in two forms, both with `"field": "history"`:

1. **Chat history entries** — `"history"` array with threads of messages and
   delivery statuses. Dispatched to `HistoryHandler.Messages`.
2. **Media content** — `"messages"` array with full message payloads (including
   media asset IDs) for messages that appeared as `"media_placeholder"` in the
   history threads. Dispatched to `HistoryHandler.MediaMessages`.

This separation prevents history media from mixing with live incoming messages.

## File Map

| File | Purpose |
|---|---|
| `webhooks.go` | `Listener`, `Config`, signature verification, `ErrorHandler`, `FallbackHandler` |
| `handler.go` | `Handler`, `MessageRequest`, `MessageInfo`, dispatch logic |
| `notification.go` | JSON wire format: `Notification`, `Entry`, `Change`, `Value` |
| `message.go` | `MessagesHandler`, `MediaHandler`, `InteractiveHandler`, message types |
| `business.go` | `BusinessNotificationHandler`, business notification types |
| `groups.go` | `GroupManagementHandler`, group webhook types |
| `flows.go` | `FlowNotificationHandler`, flows webhook types |
| `history.go` | `HistoryHandler`, history sync types |
| `status.go` | Status types, user preferences, helper functions |
| `calls.go` | Call types |
| `change_field.go` | Change field classification and dispatch lookup |
| `callbacks/` | Alternate callback URL API client |
| `router/` | HTTP router integration with middleware support |
