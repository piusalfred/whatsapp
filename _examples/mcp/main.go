package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	mcpmessage "github.com/piusalfred/whatsapp/extras/mcp/message"
	"github.com/piusalfred/whatsapp/message"
)

var _ mcpmessage.Sender = (*fakeClient)(nil)

type fakeClient struct{}

func (f fakeClient) SendMessage(ctx context.Context, message *message.Message) (*message.Response, error) {
	// TODO implement me
	panic("implement me")
}

func (f fakeClient) SendText(ctx context.Context, request *message.Request[message.Text]) (
	*message.Response, error,
) {
	resp := &message.Response{
		Product: "whatsapp",
		Contacts: []*message.ResponseContact{
			{
				Input:      request.Recipient,
				WhatsappID: "abjbv12uyt627r56r2",
			},
		},
		Messages: []*message.ID{
			{
				ID:            "message.response.id",
				MessageStatus: "DELIVERED",
			},
		},
	}

	return resp, nil
}

func main() {
	// Structured logging (adjust handler as you like)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)

	cfg := &mcpmessage.Config{
		HTTPAddress:                      ":9099",
		HTTPDisableGeneralOptionsHandler: false,
		HTTPReadTimeout:                  15 * time.Second,
		HTTPReadHeaderTimeout:            10 * time.Second,
		HTTPWriteTimeout:                 30 * time.Second,
		HTTPIdleTimeout:                  60 * time.Second,
		HTTPMaxHeaderBytes:               http.DefaultMaxHeaderBytes,
		HTTPServerShutdownTimeout:        10 * time.Second,

		LogLevel:   slog.LevelDebug,
		LogHandler: handler,

		MCPServerName:    "whatsapp-mcp",
		MCPServerTitle:   "WhatsApp MCP",
		MCPServerVersion: "0.1.0",
		MCPOptions:       &mcp.ServerOptions{}, // tune if needed
	}

	srv, err := mcpmessage.NewServer(cfg, &fakeClient{},
		// Optional: attach middlewares
		mcpmessage.WithServerHTTPMiddlewares(
			requestIDMiddleware,
			corsMiddleware,
		),
		// Optional init hooks
		mcpmessage.WithServerHTTPServerInitHook(func(s *http.Server) error {
			// Example: set BaseContext for per-conn context if needed
			return nil
		}),
		mcpmessage.WithServerMCPServerInitHook(func(m *mcp.Server) error {
			// Example: register additional MCP tools here
			return nil
		}),
	)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Root context cancels on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("server exited cleanly")
}

// ---- Example middlewares (optional) ----

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inject a simple request id if absent
		if r.Header.Get("X-Request-Id") == "" {
			w.Header().Set("X-Request-Id", time.Now().UTC().Format("20060102T150405.000000000Z07:00"))
		}
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust allowed origins/methods to your needs
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
