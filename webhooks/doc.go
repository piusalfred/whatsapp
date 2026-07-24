//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

/*
Package webhooks implements a WhatsApp Cloud API webhook receiver.

# Architecture

Three layers:

  - [ParseNotification] — reads the HTTP request body, verifies the
    X-Hub-Signature-256 header, and decodes JSON into a [Notification].

  - [Handler] — the central dispatch unit. Owns nine domain handlers
    (Flows, Business, Messages, Groups, History, Calls,
    SMBMessageEchoes, SMBAppStateSyncs, UserPreferences) that each
    implement [EventHandler]. Routes events by change field and
    delegates to [HandleEvent] for the error/fallback pipeline.

  - [Listener] — HTTP entry point that wires [ParseNotification],
    signature verification, middleware, and a [Handler] together.
    Handles subscription verification (GET) and event notifications
    (POST).

# Dispatch Pipeline

Each event flows through:

	dispatchEvent → EventHandler → HandleEvent
	                      │
	                      ├─ IsEventHandlerImplemented?
	                      │    NO  → Fallback()
	                      │    YES → Handle()
	                      │           ├─ ErrEventNotHandled → Fallback()
	                      │           ├─ other error → HandleError()
	                      │           └─ nil → done

# Choosing a Path

Use [Handler] directly when you need:

  - Selective, per-event-type processing
  - Direct control over concurrency and error propagation
  - To feed individual events into a queue or pipeline
  - To bypass middleware and HTTP wiring

Use [Listener] when you need:

  - Standard HTTP handler integration (net/http, chi, gorilla/mux)
  - Middleware for logging, auth, rate-limiting
  - Signature verification via X-Hub-Signature-256
  - A turn-key solution for subscription verification and event delivery

# Handler Path

Register typed callbacks, then call [Handler.HandleNotification]:

	handler := webhooks.NewHandler(
	    webhooks.WithMode(webhooks.Parallel), // optional: concurrent dispatch
	)

	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        log.Printf("text from %s: %s", req.Info.From, req.Payload.Body)
	        return nil
	    },
	))

	handler.OnFallback(webhooks.FallbackHandlerFunc(
	    func(ctx context.Context, ev webhooks.NotificationEvent) error {
	        log.Printf("unhandled event: %s", ev.Field)
	        return nil
	    },
	))

	// Parse and dispatch.
	notif, err := webhooks.ParseNotification(r, &webhooks.ParseNotificationOptions{
	    ValidatePayload: true,
	    AppSecret:       os.Getenv("APP_SECRET"),
	})
	if err != nil {
	    http.Error(w, "Bad Request", http.StatusBadRequest)
	    return
	}

	resp := handler.HandleNotification(r.Context(), notif)
	w.WriteHeader(resp.StatusCode)

For selective processing, flatten events with [Notification.Events] and
dispatch individually with [Handler.HandleNotificationEvent]:

	for _, event := range notif.Events() {
	    if event.Field == "messages" {
	        go handler.HandleNotificationEvent(context.Background(), event)
	    }
	}

# Dispatch Mode

[Handler.HandleNotification] supports two dispatch modes controlled via
[WithMode] (default: [Sequential]):

  - [Sequential] — events are processed one after another. The first
    handler error stops processing and returns HTTP 500.

  - [Parallel] — events are processed concurrently via an errgroup.
    All events run independently; a context cancellation returns
    HTTP 504. Use when event ordering is not important.

# Domain Handlers

Each domain handler implements [EventHandler] and owns typed dispatch:

  - [FlowNotificationHandler] — flow status, client/endpoint errors
  - [BusinessNotificationHandler] — account alerts, template updates,
    phone number changes, security, calls
  - [MessagesHandler] — text, media, interactive, reactions, orders,
    statuses, notification errors
  - [GroupManagementHandler] — lifecycle, participants, settings, status
  - [HistoryHandler] — chat history entries, media content
  - [CallsHandler] — connect, created, terminate, status
  - [SMBMessageEchoesHandler] — SMB message echoes
  - [SMBAppStateSyncsHandler] — SMB app state sync
  - [UserPreferencesHandler] — marketing message opt-in/out

Each has its own [EventHandler.OnError] and [EventHandler.OnFallback]
for fine-grained control. Setting them on [Handler] propagates to all
domain handlers.

# Listener Path

	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        log.Printf("text from %s: %s", req.Info.From, req.Payload.Body)
	        return nil
	    },
	))

	listener := webhooks.NewListener(handler, webhooks.ConfigReaderFunc(
	    func(r *http.Request) (*webhooks.Config, error) {
	        return &webhooks.Config{
	            Token:           os.Getenv("WEBHOOK_VERIFY_TOKEN"),
	            AppSecret:       os.Getenv("APP_SECRET"),
	            ValidatePayload: true,
	        }, nil
	    },
	))

	listener.OnError(func(ctx context.Context, r *http.Request, err error) {
	    log.Printf("listener error: %v", err)
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
	    switch r.Method {
	    case http.MethodGet:
	        listener.HandleSubscriptionVerification(w, r)
	    case http.MethodPost:
	        listener.HandleNotification(w, r)
	    }
	})

For more complex routing, use [router.WebhookRouter]:

	wr, _ := router.NewWebhookRouter(listener,
	    router.WithWebhookRouterEndpoints(router.Endpoints{
	        Webhook:                  "/webhooks",
	        SubscriptionVerification: "/webhooks",
	    }),
	    router.WithWebhookRouterGlobalMiddlewares(loggingMiddleware),
	)
	mux := http.NewServeMux()
	mux.Handle("/whatsapp/", http.StripPrefix("/whatsapp", wr))

# Middleware

[Listener] middleware wraps [NotificationHandler]. Middleware is applied
inside-out: middlewares[0] is the outermost wrapper.

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

	listener := webhooks.NewListener(handler, configReader, authMiddleware)

# Error Handling

[EventHandler.HandleError] is called when a typed handler returns an
error. Return nil to suppress the error and continue processing; return
a non-nil error to stop processing and trigger an HTTP 500 response
(which causes WhatsApp to retry).

Set error handlers per-domain or globally:

	handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
	    if errors.Is(err, webhooks.ErrInvalidSignature) {
	        return nil // don't retry
	    }
	    return err
	}))

[Listener.OnError] is purely observational — it receives every error but
cannot change the HTTP response. Use it for logging, metrics, and
alerting.

# Panic Recovery

Panics in handler callbacks are caught by [HandleEvent] and wrapped in
[PanicError] (which includes the goroutine stack trace). Use
[IsPanicError] to distinguish panics from expected errors:

	if pe, ok := webhooks.IsPanicError(err); ok {
	    log.Printf("handler panic: %v\n%s", pe.Value, pe.Stack)
	}

# Signature Verification

The Listener verifies X-Hub-Signature-256 headers using HMAC-SHA256.
Verification is enabled when [Config.ValidatePayload] is true and
Config.AppSecret is set.

Customise verification by implementing [SignatureVerifier]:

	listener.SetSignatureVerifier(webhooks.SignatureVerifierFunc(
	    func(header http.Header, payload []byte, secret string) error {
	        return myCustomVerify(header, payload, secret)
	    },
	))

# Context Lifetime

The context passed to webhook callbacks is the HTTP request context. It
is cancelled after the HTTP response is written. If you need background
work after acknowledging receipt, use [context.Background]:

	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        go processAsync(context.Background(), req)
	        return nil
	    },
	))

# Subscription Verification

WhatsApp sends a GET request with hub.mode, hub.challenge, and
hub.verify_token. [Listener.HandleSubscriptionVerification] validates
the token and writes the challenge back. If the token does not match or
the mode is not "subscribe", it returns HTTP 403.

# Payload Limits

WhatsApp documents a 3 MB limit. The package enforces [MaxPayloadBytes]
(4 MB, including a 1 MB grace margin). Payloads exceeding this limit are
rejected with a 400 status.

Reference: https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks
*/
package webhooks
