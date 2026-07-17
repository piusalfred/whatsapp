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

package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
)

// MaxPayloadBytes is the maximum webhook payload size.
// WhatsApp documents a 3 MB limit; we allow 4 MB (1 MB grace margin)
// to avoid rejecting borderline payloads.
const MaxPayloadBytes = 4 << 20

// SignatureHeaderKey is the key for the X-Hub-Signature-256 header.
const SignatureHeaderKey = "X-Hub-Signature-256"

var (
	ErrInvalidSignature          = errors.New("signature is invalid")
	ErrSignatureNotFound         = errors.New("signature not found")
	ErrSignatureVerification     = errors.New("signature verification failed")
	ErrReadNotification          = errors.New("error reading request body")
	ErrMessageDecode             = errors.New("error decoding message")
	ErrPayloadTooLarge           = errors.New("webhook payload exceeds 4 MB limit")
	ErrHandlerNotConfigured      = errors.New("webhooks: handler not configured")
	ErrConfigReaderNotConfigured = errors.New("webhooks: config reader not configured")
)

type (
	// Middleware wraps a NotificationHandler to add cross-cutting behavior.
	// Applied inside-out by NewListener so middlewares[0] runs outermost.
	Middleware func(NotificationHandler) NotificationHandler

	// Listener is the HTTP entry point for WhatsApp webhook callbacks. It
	// validates signatures, decodes the notification payload, and delegates
	// to the wrapped NotificationHandler. Construct via NewListener.
	Listener struct {
		middlewares     []Middleware
		originalHandler NotificationHandler
		handler         NotificationHandler
		configReader    ConfigReader
		sigVerifier     SignatureVerifier
		onError         func(ctx context.Context, r *http.Request, err error)
	}

	// Config holds the webhook configuration for a single business account.
	//
	// Token is the verify token used during subscription verification.
	// WhatsApp sends this value as hub.verify_token; the listener compares
	// it to confirm the request is authentic.
	//
	// AppSecret is used to validate X-Hub-Signature-256 headers on incoming
	// notifications. Leave empty to skip validation.
	//
	// Validate enables HMAC signature verification. When true, every POST
	// notification must carry a valid X-Hub-Signature-256 header.
	Config struct {
		Token           string
		ValidatePayload bool
		AppSecret       string
	}

	// ConfigReaderFunc implements the ConfigReader interface.
	ConfigReaderFunc func(request *http.Request) (*Config, error)

	// ConfigReader resolves per-request webhook configuration. The *http.Request
	// parameter enables multi-tenant setups where different phone numbers or
	// WABAs require different tokens and app secrets (e.g., keyed by request
	// path, host, or header).
	//
	// Design trade-off: accepting *http.Request couples the config layer to
	// the HTTP transport. For simpler single-tenant deployments, return a
	// static Config from ConfigReaderFunc ignoring the request:
	//
	// reader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
	//     return &webhooks.Config{
	//         Token:     os.Getenv("WEBHOOK_TOKEN"),
	//         AppSecret: os.Getenv("APP_SECRET"),
	//         Validate:  true,
	//     }, nil
	// })
	//
	// For zero-allocation config caching, wrap a ConfigReader in a sync.Once
	// or use middleware to inject config via the request context.
	ConfigReader interface {
		ReadConfig(request *http.Request) (*Config, error)
	}

	// Response carries the HTTP status code to return to WhatsApp after
	// processing a notification. Return 200 to acknowledge receipt.
	Response struct {
		StatusCode int
	}

	// NotificationHandlerFunc adapts a bare function to the NotificationHandler
	// interface so callers can pass inline functions where an interface is expected.
	NotificationHandlerFunc func(ctx context.Context, notification *Notification) *Response

	// NotificationHandler processes a decoded webhook notification.
	// Implementations route the notification to domain-specific handlers
	// based on the entry changes.
	NotificationHandler interface {
		HandleNotification(ctx context.Context, notification *Notification) *Response
	}

	// PanicError captures a panic that occurred during webhook handler dispatch.
	// It is returned through the normal error path so callers can inspect it
	// with [IsPanicError] instead of having the panic crash the server.
	//
	// The Stack field contains the goroutine stack trace at the point of the panic,
	// captured via [runtime/debug.Stack].
	PanicError struct {
		// Value is the argument passed to panic().
		Value any
		// Stack is the formatted goroutine stack trace at the panic point.
		Stack []byte
	}

	// ParseNotificationOptions controls whether and how incoming payloads are authenticated.
	ParseNotificationOptions struct {
		VerifyPayloadSignature bool
		AppSecret              string
		SignatureVerifier      SignatureVerifier
	}

	// VerifySignatureOptions holds the parameters required for signature validation.
	// It combines the payload (which is the raw request body), the signature string extracted from the header,
	// and the app's secret used to generate the HMAC signature.
	VerifySignatureOptions struct {
		Signature string // Extracted signature (without the sha256= prefix)
		AppSecret string // App secret used for signature generation
	}
)

// NewListener creates a Listener that wraps handler with the given middlewares.
// Middlewares are applied inside-out so middlewares[0] is the outermost wrapper.
// The ConfigReader is called on every HTTP request to resolve per-tenant config.
func NewListener(
	handler NotificationHandler,
	reader ConfigReader,
	middlewares ...Middleware,
) *Listener {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		m := middlewares[i]
		wrapped = m(wrapped)
	}

	return &Listener{
		middlewares:     middlewares,
		originalHandler: handler,
		handler:         wrapped,
		configReader:    reader,
		sigVerifier:     SignatureVerifierFunc(VerifyPayloadSignature),
	}
}

// OnError registers a callback that is invoked when the Listener encounters an
// error that prevents it from completing a subscription verification or notification
// request. This includes nil/missing configuration, config-read failures, signature
// validation failures, payload decode errors, and panics recovered from user handlers.
//
// The callback is purely observational — it cannot change the HTTP response sent
// to WhatsApp. Use [Middleware] if you need to intercept or rewrite responses.
//
// Only one callback is supported; calling OnError multiple times overwrites
// the previous callback. Pass nil to clear.
func (ls *Listener) OnError(handler func(ctx context.Context, r *http.Request, err error)) {
	ls.onError = handler
}

// SetSignatureVerifier allows you to override the default signature validation logic. By default,
// the Listener uses the [VerifyPayloadSignature] function to validate the X-Hub-Signature-256 header
// against the request body and app secret. You can provide a custom implementation of the SignatureVerifier
// interface to change this behavior.
func (ls *Listener) SetSignatureVerifier(validator SignatureVerifier) {
	ls.sigVerifier = validator
}

// HandleSubscriptionVerification responds to WhatsApp's GET handshake. It reads the hub.mode,
// hub.challenge, and hub.verify_token query parameters, validates the token, and writes the challenge
// back to complete verification.
func (ls *Listener) HandleSubscriptionVerification(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if ls.configReader == nil {
		ls.respondError(ctx, writer, request, ErrConfigReaderNotConfigured, http.StatusInternalServerError)
		return
	}

	config, err := ls.configReader.ReadConfig(request)
	if err != nil {
		ls.respondError(ctx, writer, request, err, http.StatusInternalServerError)
		return
	}
	challenge, err := verifySubscriptionRequest(request, config.Token)
	if err != nil {
		ls.respondError(ctx, writer, request, err, http.StatusForbidden)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(html.EscapeString(challenge)))
}

// HandleNotification processes an incoming POST webhook event. It reads the request body,
// optionally validates the X-Hub-Signature-256 header, decodes the notification JSON, and
// dispatches to the wrapped handler.
//
// A defer/recover guard catches panics from user-supplied middlewares or handlers and returns
// HTTP 500 to WhatsApp instead of crashing the server.
//
// WhatsApp retries non-200 responses for up to 7 days with decreasing frequency. Ensure your
// handler is idempotent — the same notification may be delivered more than once.
//
// Unrecognized webhook fields are silently dropped (no error) so retries are not
// triggered for unimplemented types.
func (ls *Listener) HandleNotification(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if ls.handler == nil {
		ls.respondError(ctx, writer, request, ErrHandlerNotConfigured, http.StatusInternalServerError)
		return
	}
	if ls.configReader == nil {
		ls.respondError(ctx, writer, request, ErrConfigReaderNotConfigured, http.StatusInternalServerError)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			pe := &PanicError{Value: r, Stack: debug.Stack()}
			ls.respondError(ctx, writer, request, pe, http.StatusInternalServerError)
		}
	}()

	config, err := ls.configReader.ReadConfig(request)
	if err != nil {
		ls.respondError(ctx, writer, request, err, http.StatusInternalServerError)

		return
	}

	notification, err := ParseNotification(request, &ParseNotificationOptions{
		VerifyPayloadSignature: config.ValidatePayload,
		AppSecret:              config.AppSecret,
		SignatureVerifier:      ls.sigVerifier,
	})
	if err != nil {
		ls.respondError(ctx, writer, request, err, http.StatusBadRequest)

		return
	}

	response := ls.handler.HandleNotification(ctx, notification)

	writer.WriteHeader(response.StatusCode)
}

// ParseNotification reads the request body, optionally validates the
// signature header against the app secret, and decodes the JSON into a
// Notification. The request body is restored afterward so it can be re-read.
//
// This function assumes the Webhooks "Include Values" setting is enabled in
// the App Dashboard. If values are disabled, changes arrive as
// "changed_fields" arrays instead of "changes" objects with values, and
// decoding will produce empty Change.Value fields.
func ParseNotification(request *http.Request, options *ParseNotificationOptions) (*Notification, error) {
	var buff bytes.Buffer
	_, err := io.Copy(&buff, io.LimitReader(request.Body, MaxPayloadBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadNotification, err)
	}
	if buff.Len() == MaxPayloadBytes {
		return nil, ErrPayloadTooLarge
	}

	request.Body = io.NopCloser(&buff)

	if options.VerifyPayloadSignature {
		if validateErr := options.SignatureVerifier.VerifySignature(
			request.Header, buff.Bytes(), options.AppSecret); validateErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrSignatureVerification, validateErr)
		}
	}

	var notification Notification
	if err = json.NewDecoder(&buff).Decode(&notification); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: %w", ErrMessageDecode, err)
	}

	return &notification, nil
}

// VerifyPayloadSignature extracts and validates the signature from the HTTP request header.
//
// It performs the following steps:
//  1. Extracts the signature from the "X-Hub-Signature-256" header using ExtractSignatureFromHeader.
//  2. Validates the extracted signature against the payload using VerifySignature.
//
// This function is designed to work with signed webhook events, ensuring that the request
// is authentic and has not been tampered with.
//
// Parameters:
//   - header: HTTP headers from the incoming request.
//   - payload: The raw body (payload) of the request.
//   - secret: The app's secret used to generate the expected signature.
//
// Returns an error if the signature is invalid or missing.
func VerifyPayloadSignature(header http.Header, payload []byte, secret string) error {
	signature, err := ExtractSignatureFromHeader(header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	params := VerifySignatureOptions{
		Signature: signature,
		AppSecret: secret,
	}

	if err = VerifySignature(payload, params); err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	return nil
}

// ExtractSignatureFromHeader extracts the signature from the HTTP header.
//
// The X-Hub-Signature-256 header contains the signature as a SHA256 hash of the payload,
// prefixed with "sha256=". This function strips that prefix and returns the actual signature.
//
// Returns the signature string without the prefix or an error if the header is missing
// or improperly formatted.
func ExtractSignatureFromHeader(header http.Header) (string, error) {
	signature := header.Get(SignatureHeaderKey)
	if !strings.HasPrefix(signature, "sha256=") {
		return "", fmt.Errorf("signature is missing or does not have prefix \"sha256\": %w", ErrSignatureNotFound)
	}

	return signature[7:], nil
}

// VerifySignature validates the signature of a payload using the provided VerifySignatureOptions.
//
// The validation process involves generating an HMAC-SHA256 signature using the payload and the app's secret.
// The signature is then compared to the one provided in the request header.
//
// To validate the payload:
//  1. Generate a SHA256 signature using the payload and your app's AppSecret.
//  2. Compare your signature to the signature in the X-Hub-Signature-256 header (after stripping the "sha256=" prefix).
//
// If the signatures match, the payload is considered genuine. It's important to note that the signature is
// generated using an escaped Unicode version of the payload (e.g., special characters are encoded as \u00e4).
// This function assumes the payload is provided in its final byte form.
//
// Errors are returned if the signature is invalid or the decoding process fails.
func VerifySignature(payload []byte, params VerifySignatureOptions) error {
	// Decode the provided signature from hexadecimal to raw bytes.
	decodedSig, err := hex.DecodeString(params.Signature)
	if err != nil {
		return fmt.Errorf("error decoding signature: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(params.AppSecret))
	if _, err = mac.Write(payload); err != nil {
		return fmt.Errorf("error hashing payload: %w", err)
	}
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(decodedSig, expectedSignature) {
		return ErrInvalidSignature
	}

	return nil
}

// IsPanicError checks whether err is a [PanicError] and returns it.
// Use this to distinguish programmatic bugs (panics in user callbacks)
// from expected operational errors.
func IsPanicError(err error) (*PanicError, bool) {
	if pe, ok := errors.AsType[*PanicError](err); ok {
		return pe, true
	}
	return nil, false
}

// verifySubscriptionRequest validates the GET query parameters during the initial Webhook connection.
func verifySubscriptionRequest(request *http.Request, token string) (string, error) {
	q := request.URL.Query()
	mode := q.Get("hub.mode")
	challenge := q.Get("hub.challenge")
	providedToken := q.Get("hub.verify_token")

	if providedToken != token || mode != "subscribe" {
		return "", ErrInvalidSignature
	}

	return challenge, nil
}

// respondError notifies the registered onError callback (if set) and writes
// a safe HTTP error response. The statusText is sent as the response body
// so internal error details are never leaked to the client.
func (ls *Listener) respondError(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	statusCode int,
) {
	if ls.onError != nil {
		ls.onError(ctx, r, err)
	}
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func (fn ConfigReaderFunc) ReadConfig(request *http.Request) (*Config, error) {
	return fn(request)
}

func (fn NotificationHandlerFunc) HandleNotification(ctx context.Context, notification *Notification) *Response {
	return fn(ctx, notification)
}

// Error returns a description of the panic including the original value.
func (e *PanicError) Error() string {
	return fmt.Sprintf("handler panic: %v", e.Value)
}

type (
	SignatureVerifier interface {
		VerifySignature(header http.Header, payload []byte, secret string) error
	}

	SignatureVerifierFunc func(header http.Header, payload []byte, secret string) error
)

func (fn SignatureVerifierFunc) VerifySignature(header http.Header, payload []byte, secret string) error {
	return fn(header, payload, secret)
}
