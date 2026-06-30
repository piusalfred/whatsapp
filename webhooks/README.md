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

### Two-Level Fallback Architecture

Each sub-handler has its own `Fallback` field. The dispatch chain is:

```
dispatcher → dedicated handler (if set)
           → sub-handler fallback (if set)
           → general fallback (Handler.OnFallback)
           → silent skip (HTTP 200)
```

This lets users catch unknown event types within a domain (e.g., new flow events)
without affecting other domains. See each sub-handler's `OnFallback` method.

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
