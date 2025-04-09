/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/webhooks"
)

// LoggingMiddleware is an example middleware that logs
// before and after handling a single incoming webhook request.
func LoggingMiddleware(next webhooks.NotificationHandlerFunc) webhooks.NotificationHandlerFunc {
	return func(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
		fmt.Println("[LoggingMiddleware] --> Before handling notification")
		response := next(ctx, notification)
		fmt.Println("[LoggingMiddleware] <-- After handling notification")
		return response
	}
}

// ReactionHandler is an example of a more advanced handler type
// that implements logic for dealing with Reaction messages.
type ReactionHandler struct {
	Logger *slog.Logger
	Store  map[string]any
}

// Ensure that ReactionHandler satisfies the interface for reaction messages.
// This is optional, but helpful if you want static type checking.
var _ webhooks.ReactionHandler = (*ReactionHandler)(nil)

// Handle processes reaction messages (type: reaction).
// It logs the incoming reaction and stores the emoji in a map.
func (r *ReactionHandler) Handle(
	ctx context.Context,
	nctx *webhooks.MessageNotificationContext,
	mctx *webhooks.MessageInfo,
	reaction *message.Reaction,
) error {
	r.Logger.Info("Received reaction message",
		"context", nctx,
		"message_info", mctx,
		"reaction", reaction,
	)

	r.Logger.Info("Reaction emoji", "emoji", reaction.Emoji)

	// For example, you can store the emoji in an in-memory map keyed by the message context ID.
	if r.Store == nil {
		r.Store = make(map[string]any)
	}
	if mctx.Context != nil {
		r.Store[mctx.Context.ID] = reaction.Emoji
	}

	return nil
}

func main() {
	// Create a Handler that can process various types of webhook notifications.
	// This initializes no-op handlers for all message types. The no-op handlers
	// does nothing and returns nil.
	handler := webhooks.NewHandler()

	// There are two ways to register handlers for specific message types:
	// 1. A simple callback function for a specific message type.
	// 2. Implementing a specific handler interface (e.g., webhooks.ReactionHandler) for more complex logic.

	// Register a simple callback for text messages using OnTextMessage. This will replace the
	// no-op handler that was initialized by called webhooks.NewHandler().
	// This callback will be invoked whenever we receive a "type": "text" message.
	handler.OnTextMessage(func(ctx context.Context,
		nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo,
		text *webhooks.Text,
	) error {
		fmt.Printf("[OnTextMessage] New text message received:\n")
		fmt.Printf("  Notification context: %+v\n", nctx)
		fmt.Printf("  Message info:         %+v\n", mctx)
		fmt.Printf("  Text content:         %+v\n", text)
		return nil
	})

	// In case of complex handlers you can implement a certain notification type handler interface
	// for example for reaction messages you can implement the webhooks.ReactionHandler interface
	// which is an alias for MessageHandler[message.Reaction]
	reactionHandler := &ReactionHandler{
		Logger: slog.Default(),
		Store:  make(map[string]any),
	}

	// this is how you can register it with the main handler.
	handler.SetReactionMessageHandler(reactionHandler)

	// Create a Listener that wraps our handler with optional middlewares (like LoggingMiddleware).
	// We also provide a VerifyTokenReader and ValidateOptions for optional signature verification, etc.
	messageListener := webhooks.NewListener(
		handler.HandleNotification, // The core Handler’s method
		func(ctx context.Context) (string, error) { // VerifyTokenReader: returns your verify token
			return "TOKEN", nil
		},
		&webhooks.ValidateOptions{
			Validate:  false, // Skip signature validation for simplicity
			AppSecret: "SUPERSECRET",
		},
		LoggingMiddleware, // Example middleware
	)

	// Set up two HTTP endpoints: one for subscription verification ("GET")
	// and one for receiving actual webhook notifications ("POST").
	//
	// In actual code, you might do http.HandleFunc("/webhooks/messages", ...) (without "POST " prefix).
	// For clarity, here we show "POST /webhooks/messages" and "POST /webhooks/verify" in a single snippet.
	http.HandleFunc("POST /webhooks/messages", messageListener.HandleNotification)
	http.HandleFunc("POST /webhooks/verify", messageListener.HandleSubscriptionVerification)

	// Start an HTTP server on port :8080 to listen for incoming requests.
	fmt.Println("[main] Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
