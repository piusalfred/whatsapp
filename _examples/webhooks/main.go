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

	"github.com/go-chi/chi/v5"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/webhooks"
	"github.com/piusalfred/whatsapp/webhooks/router"
)

func LoggingMiddleware(next webhooks.NotificationHandler) webhooks.NotificationHandler {
	handler := webhooks.NotificationHandlerFunc(func(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
		log.Println("[LoggingMiddleware] --> Before handling notification")
		response := next.HandleNotification(ctx, notification)
		log.Printf("response is %+v", response)
		log.Println("[LoggingMiddleware] <-- After handling notification")
		return response
	})
	return handler
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

	if r.Store == nil {
		r.Store = make(map[string]any)
	}
	if mctx.Context != nil {
		r.Store[mctx.Context.ID] = reaction.Emoji
	}

	return nil
}

func main() {
	handler := webhooks.NewHandler()
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

	conf := webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
		value := &webhooks.Config{
			Token:     "TOKEN",
			Validate:  true,
			AppSecret: "SUPERSECRET",
		}

		return value, nil
	})

	messageListener := webhooks.NewListener(
		handler,
		conf,
		LoggingMiddleware,
	)

	muxOptions := []router.SimpleRouterOption{
		router.WithSimpleRouterEndpoints(router.Endpoints{
			Webhook:                  "/webhooks/messages",
			SubscriptionVerification: "/webhooks/messages",
		}),
		router.WithSimpleRouterMux(chi.NewMux()),
	}

	mux, err := router.NewSimpleRouter(messageListener, muxOptions...)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
