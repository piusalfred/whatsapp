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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/piusalfred/whatsapp"
	hooks "github.com/piusalfred/whatsapp/webhooks"
)

var ErrInvalidToken = errors.New("invalid token")

type Config struct {
	BaseURL           string
	Version           string
	AccessToken       string
	PhoneNumberID     string
	BusinessAccountID string
	WebhookSecret     string
	Port              int
}

type Service struct {
	Config   *Config
	whatsapp *whatsapp.Client
	Logger   *slog.Logger
}

func (svc *Service) WebhookSubVerifier() http.HandlerFunc {
	verif := hooks.SubscriptionVerifier(func(ctx context.Context, request *hooks.VerificationRequest) error {
		if svc.Config.WebhookSecret != request.Token {
			return fmt.Errorf("%w: %s", ErrInvalidToken, request.Token)
		}

		return nil
	})

	return verif.HandlerFunc
}

// ListenAndServe starts the server and listens for incoming requests.
func (svc *Service) ListenAndServe(ctx context.Context) error {
	listener := hooks.NewEventListener()
	listener.OnTextMessage(svc.OnTextMessageHook)
	router := httprouter.New()
	router.HandlerFunc(
		http.MethodGet,
		"/webhooks",
		svc.WebhookSubVerifier(),
	)
	router.Handler(
		http.MethodPost,
		"/webhooks",
		listener.NotificationHandler(),
	)
	server := &http.Server{
		Addr:                         fmt.Sprintf(":%d", svc.Config.Port),
		Handler:                      router,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     slog.NewLogLogger(svc.Logger.Handler(), slog.LevelDebug),
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	go func() {
		// Waiting for context cancellation
		<-ctx.Done()

		// Create a context for shutdown with a timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Try to gracefully shut down the server
		if err := server.Shutdown(shutdownCtx); err != nil {
			svc.Logger.ErrorContext(shutdownCtx, "error during server shutdown:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func NewService(config *Config) (*Service, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	client, err := whatsapp.NewClientWithConfig(&whatsapp.Config{
		BaseURL:           config.BaseURL,
		Version:           config.Version,
		AccessToken:       config.AccessToken,
		PhoneNumberID:     config.PhoneNumberID,
		BusinessAccountID: config.BusinessAccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("new client with config: %w", err)
	}
	service := &Service{
		Config:   config,
		whatsapp: client,
		Logger:   logger,
	}

	return service, nil
}

func Run(config *Config) error {
	service, err := NewService(config)
	if err != nil {
		return fmt.Errorf("new service: %w", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())

	// Capture system signals to trigger a shutdown
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // Capture Ctrl+C and terminated signal
		<-sigs
		cancel(fmt.Errorf("received signal"))
	}()

	defer cancel(nil)

	return service.ListenAndServe(ctx)
}

func main() {
	config := &Config{
		BaseURL:           whatsapp.BaseURL,
		Version:           "v16.0",
		AccessToken:       "EAALLrT0ok6UBO6cpaRw0iHybYbwLq1xMtyMWXFVXJUJOHRgapUoA4PQwzksAqUQ4zqb3QJPsh9GXo8AACEe7hrFRjtpexY6qG05O9YQr2e1d0orrYIkD0B11mOK0llEls5KhDTsZCDPd1QqZAsbZAxZCX9jC65Aiz3gnzvaLx3nBGyXkA6aNZCPHYpxYryfpTtx7nkLsOvobzyvAZD",
		PhoneNumberID:     "114425371552711",
		BusinessAccountID: "113592508304116",
		WebhookSecret:     "testtoken",
		Port:              8080,
	}

	if err := Run(config); err != nil {
		fmt.Printf("error running service: %v\n", err)
		os.Exit(1)
	}
}
