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
Package webhooks implements a WhatsApp Cloud API webhook receiver with two
usage paths: a fine-grained event-driven API built on [Handler], and a
general-purpose HTTP listener built on [Listener].

# Architecture

The package operates at three layers:

  - [ParseNotification] — reads the HTTP request body, optionally verifies
    the X-Hub-Signature-256 header, and decodes the JSON into a
    [Notification] envelope.

  - [Handler] — routes typed webhook events to your callbacks. It owns the
    dispatch table and can operate standalone (fine-grained events path) or
    be wrapped by a [Listener] (general HTTP path).

  - [Listener] — HTTP entry point that wires [ParseNotification], signature
    verification, middleware, and a [Handler] together. It handles both
    subscription verification (GET) and event notifications (POST).

# Choosing a Path

Use the fine-grained events path when you need:

  - Selective, per-event-type async processing
  - Direct control over concurrency and error propagation
  - To feed individual events into a queue or pipeline
  - To bypass middleware and HTTP wiring

Use the general Listener path when you need:

  - Standard HTTP handler integration (net/http, chi, gorilla/mux)
  - Middleware for logging, auth, rate-limiting
  - Signature verification via X-Hub-Signature-256
  - A turn-key solution for subscription verification and event delivery

# Fine-Grained Events Path

Register typed callbacks on a [Handler], then call [Handler.HandleNotificationEvents]
to flatten and concurrently dispatch all events in a notification:

	handler := webhooks.NewHandler()
	handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
	    log.Printf("webhook error: %v", err)
	    return nil // continue processing remaining events
	}))

	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        log.Printf("text from %s: %s", req.MessageInfo.From, req.Payload.Body)
	        return nil
	    },
	))

	handler.OnReactionMessage(webhooks.MessageHandlerFunc[media.Reaction](
	    func(ctx context.Context, req *webhooks.MessageRequest[media.Reaction]) error {
	        log.Printf("reaction: %s", req.Payload.Emoji)
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

	resp := handler.HandleNotificationEvents(r.Context(), notif)
	w.WriteHeader(resp.StatusCode)

You can also process individual events selectively by calling
[Notification.Events] and [Handler.HandleNotificationEvent] directly:

	for _, event := range notif.Events() {
	    if event.Field == "messages" {
	        go handler.HandleNotificationEvent(context.Background(), event)
	    }
	}

# General Listener Path

Wrap a [Handler] in a [Listener] to get HTTP routing, middleware, and
signature verification:

	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        log.Printf("text from %s: %s", req.MessageInfo.From, req.Payload.Body)
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

	// Attach an error observer (purely observational — cannot change responses).
	listener.OnError(func(ctx context.Context, r *http.Request, err error) {
	    log.Printf("listener error: %v", err)
	})

	// Wire into net/http.
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
	    switch r.Method {
	    case http.MethodGet:
	        listener.HandleSubscriptionVerification(w, r)
	    case http.MethodPost:
	        listener.HandleNotification(w, r)
	    }
	})

For more complex routing, use [router.WebhookRouter] which supports
path prefixes, global and route-specific middleware:

	wr, _ := router.NewWebhookRouter(listener,
	    router.WithWebhookRouterEndpoints(router.Endpoints{
	        Webhook:                  "/webhooks",
	        SubscriptionVerification: "/webhooks",
	    }),
	    router.WithWebhookRouterGlobalMiddlewares(loggingMiddleware),
	    router.WithWebhookRouterWebhookMiddlewares(rateLimitMiddleware),
	)
	mux := http.NewServeMux()
	mux.Handle("/whatsapp/", http.StripPrefix("/whatsapp", wr))

# Middleware

[Listener] middleware wraps [NotificationHandler]. Middleware is applied
inside-out: middlewares[0] is the outermost wrapper.

	authMiddleware := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
	    return webhooks.NotificationHandlerFunc(
	        func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
	            // Short-circuit: return early to skip the handler.
	            if !isAuthorized(ctx) {
	                return &webhooks.Response{StatusCode: http.StatusForbidden}
	            }
	            resp := next.HandleNotification(ctx, n)
	            // Modify the response after the handler runs.
	            if resp.StatusCode == http.StatusOK {
	                log.Printf("processed notification with %d entries", len(n.Entry))
	            }
	            return resp
	        },
	    )
	}

	listener := webhooks.NewListener(handler, configReader, authMiddleware)

# Signature Verification

The Listener verifies X-Hub-Signature-256 headers using HMAC-SHA256.
Verification is enabled when [Config.ValidatePayload] is true and
Config.AppSecret is set.

To customise the verification logic (e.g., constant-time comparison,
multi-tenant secret lookup), implement [SignatureVerifier] and set it:

	listener.SetSignatureVerifier(webhooks.SignatureVerifierFunc(
	    func(header http.Header, payload []byte, secret string) error {
	        return myCustomVerify(header, payload, secret)
	    },
	))

# Error Handling

Handler callbacks and middleware return errors. By default, errors
propagate through the [ErrorHandler] chain:

  - Return nil to acknowledge success. The error is logged/discarded and
    processing continues to the next event.

  - Return a non-nil error to signal a fatal failure. The Handler stops
    processing and returns an error to the Listener, which returns a
    non-200 status to WhatsApp. WhatsApp retries non-200 responses for up
    to 7 days with decreasing frequency.

Use [Handler.OnError] to install a custom error handler that can suppress
non-fatal errors:

	handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
	    if errors.Is(err, webhooks.ErrInvalidSignature) {
	        return nil // don't retry — this payload will never be valid
	    }
	    return err // everything else triggers a retry
	}))

For the Listener level, [Listener.OnError] is purely observational — it
receives every error but cannot change the HTTP response. Use it for
logging, metrics, and alerting.

# Panic Recovery

Panics in handler callbacks and middleware are caught at both the
[Handler] layer (via [Handler.dispatchChange]) and the [Listener]
layer (via [Listener.HandleNotification]). A panic is wrapped in
[PanicError] (which includes the goroutine stack trace) and surfaced
through the error path. Use [IsPanicError] to distinguish panics from
expected errors:

	if pe, ok := webhooks.IsPanicError(err); ok {
	    log.Printf("handler panic: %v\n%s", pe.Value, pe.Stack)
	}

# Context Lifetime

The context passed to webhook callbacks is the HTTP request context. It
is cancelled after the HTTP response is written. If you need to perform
background work after acknowledging receipt, use [context.Background]:

	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
	    func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
	        go processAsync(context.Background(), req)
	        return nil
	    },
	))

# Subscription Verification

WhatsApp sends a GET request with hub.mode, hub.challenge, and
hub.verify_token. [Listener.HandleSubscriptionVerification] validates the
token and writes the challenge back. If the token does not match or the
mode is not "subscribe", it returns HTTP 403.

# Payload Limits

WhatsApp documents a 3 MB limit. The package enforces [MaxPayloadBytes]
(4 MB, including a 1 MB grace margin). Payloads exceeding this limit are
rejected with a 400 status.

Reference: https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks
*/
package webhooks
