/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/piusalfred/whatsapp/webhooks"
)

type (
	Middleware func(http.Handler) http.Handler

	Server struct {
		Logger            *slog.Logger
		SubscriptionToken string
		Listener          *webhooks.NotificationListener
		Endpoint          string
	}
)

func main() {
	if err := ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

// ListenAndServe starts the server and listens for incoming webhook notifications.
func ListenAndServe() error {
	server := NewServer("1234", "/webhook")
	mux := http.NewServeMux()
	notificationEndpoint := fmt.Sprintf("POST %s", server.Endpoint)
	verificationEndpoint := fmt.Sprintf("GET %s", server.Endpoint)
	notificationHandler := wrapMiddleware(server.LoggingMiddleware, server.Listener.HandleNotificationX)
	verificationHandler := wrapMiddleware(server.LoggingMiddleware, server.Listener.HandleSubscriptionVerification)

	mux.HandleFunc(notificationEndpoint, notificationHandler)
	mux.HandleFunc(verificationEndpoint, verificationHandler)

	return http.ListenAndServe(":8080", mux)
}

// NewServer returns a new instance of the Server.
func NewServer(subscriptionToken, endpoint string) *Server {
	s := &Server{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})).WithGroup("webhooks"),
		SubscriptionToken: subscriptionToken,
		Endpoint:          endpoint,
	}

	g := webhooks.GeneralNotificationHandler(s.HandleNotification)
	s.Listener = webhooks.NewGeneralListener(&webhooks.Config{
		AppSecret:      "",
		ShouldValidate: false,
		VerifyToken:    "",
	}, g)

	return s
}

// wrapMiddleware wraps the given http.HandlerFunc with the provided middleware.
// and returns http.HandlerFunc.
func wrapMiddleware(middleware Middleware, handler http.HandlerFunc) http.HandlerFunc {
	return middleware(handler).ServeHTTP
}

// LoggingMiddleware is a middleware that logs incoming requests.
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		s.Logger.LogAttrs(ctx, slog.LevelInfo, "request received",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("host", r.Host),
		)

		next.ServeHTTP(w, r)

		s.Logger.LogAttrs(ctx, slog.LevelInfo, "request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("host", r.Host),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// HandleNotification is a handler function that processes incoming webhook notifications.
func (s *Server) HandleNotification(ctx context.Context, notification *webhooks.Notification) *webhooks.Response {
	id := notification.Entry[0].ID
	object := notification.Object
	product := notification.Entry[0].Changes[0].Value.MessagingProduct

	s.Logger.LogAttrs(ctx, slog.LevelInfo, "notification received",
		slog.String("id", id),
		slog.String("object", object),
		slog.String("product", product),
	)

	return &webhooks.Response{
		StatusCode: http.StatusOK,
	}
}
