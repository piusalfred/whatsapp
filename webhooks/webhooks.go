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
	"io"
	"net/http"
	"strings"
)

type Listener[T any] struct {
	middlewares       []HandleMiddleware[T]
	originalHandler   NotificationHandlerFunc[T]
	Handler           NotificationHandlerFunc[T]
	VerifyTokenReader VerifyTokenReader
	ValidateOptions   *ValidateOptions
}

func NewListener[T any](handler NotificationHandlerFunc[T],
	reader VerifyTokenReader, validateOpts *ValidateOptions, middlewares ...HandleMiddleware[T],
) *Listener[T] {
	wrappedHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		wrappedHandler = middleware(wrappedHandler)
	}

	return &Listener[T]{
		middlewares:       middlewares,
		originalHandler:   handler,
		Handler:           wrappedHandler,
		VerifyTokenReader: reader,
		ValidateOptions:   validateOpts,
	}
}

func (listener *Listener[T]) HandleSubscriptionVerification(writer http.ResponseWriter, request *http.Request) {
	listener.VerifyTokenReader.VerifySubscription(writer, request)
}

func (listener *Listener[T]) HandleNotification(writer http.ResponseWriter, request *http.Request) {
	var (
		notification *T
		ctx          = request.Context()
		err          error
	)

	notification, err = ExtractAndValidatePayload[T](request, listener.ValidateOptions)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	response := listener.Handler.HandleNotification(ctx, notification)

	writer.WriteHeader(response.StatusCode)
}

type (
	HandleMiddleware[T any] func(handlerFunc NotificationHandlerFunc[T]) NotificationHandlerFunc[T]

	Response struct {
		StatusCode int
	}

	NotificationHandlerFunc[T any] func(ctx context.Context, notification *T) *Response

	NotificationHandler[T any] interface {
		HandleNotification(ctx context.Context, notification *T) *Response
	}
)

func (fn NotificationHandlerFunc[T]) HandleNotification(ctx context.Context, notification *T) *Response {
	return fn(ctx, notification)
}

// OnEventNotification creates an HTTP handler function for processing webhook event notifications.
func OnEventNotification[T any](handler NotificationHandler[T]) http.HandlerFunc {
	fn := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			msgErr := fmt.Errorf("%w: %w", ErrReadNotification, err)
			http.Error(writer, msgErr.Error(), http.StatusInternalServerError)

			return
		}

		request.Body = io.NopCloser(bytes.NewBuffer(body))

		var payload T

		if err = json.Unmarshal(body, &payload); err != nil {
			msgErr := fmt.Errorf("%w: %w", ErrMessageDecode, err)
			http.Error(writer, msgErr.Error(), http.StatusInternalServerError)

			return
		}

		response := handler.HandleNotification(request.Context(), &payload)

		writer.WriteHeader(response.StatusCode)
	})

	return fn
}

type ValidateOptions struct {
	Validate  bool
	AppSecret string
}

func ExtractAndValidatePayload[T any](request *http.Request, options *ValidateOptions) (*T, error) {
	var buff bytes.Buffer
	_, err := io.Copy(&buff, request.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	request.Body = io.NopCloser(&buff)

	if options.Validate {
		if err = ValidatePayloadSignature(request.Header, buff.Bytes(), options.AppSecret); err != nil {
			return nil, err
		}
	}

	var notification T
	if err = json.NewDecoder(&buff).Decode(&notification); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	return &notification, nil
}

// SignatureHeaderKey is the key for the X-Hub-Signature-256 header.
const SignatureHeaderKey = "X-Hub-Signature-256"

// ValidateSignatureOptions holds the parameters required for signature validation.
// It combines the payload (which is the raw request body), the signature string extracted from the header,
// and the app's secret used to generate the HMAC signature.
type ValidateSignatureOptions struct {
	Signature string // Extracted signature (without the sha256= prefix)
	AppSecret string // App secret used for signature generation
}

// ValidateSignature validates the signature of a payload using the provided ValidateSignatureOptions.
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
func ValidateSignature(payload []byte, params ValidateSignatureOptions) error {
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

// ExtractSignatureFromHeader extracts the signature from the HTTP header.
//
// The X-Hub-Signature-256 header contains the signature as a SHA256 hash of the payload,
// prefixed with "sha256=". This function strips that prefix and returns the actual signature.
//
// Returns the signature string without the prefix, or an error if the header is missing
// or improperly formatted.
func ExtractSignatureFromHeader(header http.Header) (string, error) {
	signature := header.Get(SignatureHeaderKey)
	if !strings.HasPrefix(signature, "sha256=") {
		return "", fmt.Errorf("signature is missing or does not have prefix \"sha256\": %w", ErrSignatureNotFound)
	}

	return signature[7:], nil
}

// ValidatePayloadSignature extracts and validates the signature from the HTTP request header.
//
// It performs the following steps:
//  1. Extracts the signature from the "X-Hub-Signature-256" header using ExtractSignatureFromHeader.
//  2. Validates the extracted signature against the payload using ValidateSignature.
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
func ValidatePayloadSignature(header http.Header, payload []byte, secret string) error {
	signature, err := ExtractSignatureFromHeader(header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	params := ValidateSignatureOptions{
		Signature: signature,
		AppSecret: secret,
	}

	if err = ValidateSignature(payload, params); err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	return nil
}

func ValidateRequestPayloadSignature(request *http.Request, secret string) error {
	signature, err := ExtractSignatureFromHeader(request.Header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	params := ValidateSignatureOptions{
		Signature: signature,
		AppSecret: secret,
	}

	var buff bytes.Buffer
	_, err = io.Copy(&buff, request.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	request.Body = io.NopCloser(&buff)

	if err = ValidateSignature(buff.Bytes(), params); err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	return nil
}

// SubscriptionVerificationHandlerFunc returns an http.HandlerFunc that handles the verification of webhook
// subscriptions using a fixed verification token.
//
// This function verifies a subscription request sent by the Webhooks product. It uses a fixed verification
// token provided as an argument (`verifyToken`). If the provided token matches and the `hub.mode` is "subscribe",
// it responds with the `hub.challenge` value, completing the verification process.
//
// Use this function if you do not require dynamic token lookup. For that use VerifyTokenReader.VerifySubscription.
func SubscriptionVerificationHandlerFunc(verifyToken string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		q := request.URL.Query()
		mode := q.Get("hub.mode")
		challenge := q.Get("hub.challenge")
		token := q.Get("hub.verify_token")

		if token != verifyToken || mode != "subscribe" {
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(challenge))
	}
}

// VerifyTokenReader is a function signature that retrieves the verification token.
// It takes a context and returns the token string and any error encountered.
type VerifyTokenReader func(ctx context.Context) (string, error)

// VerifySubscription handles the verification of webhook subscriptions using a dynamic token retrieved
// by the VerifyTokenReader.
//
// This function verifies a subscription request sent by the Webhooks product. It dynamically obtains the
// verification token by calling the VerifyTokenReader. If the provided token matches and the `hub.mode`
// is "subscribe", it responds with the `hub.challenge` value, completing the verification process.
//
// if you don't want dynamic token lookup use SubscriptionVerificationHandlerFunc.
func (reader VerifyTokenReader) VerifySubscription(writer http.ResponseWriter, request *http.Request) {
	token, err := reader(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	q := request.URL.Query()
	mode := q.Get("hub.mode")
	challenge := q.Get("hub.challenge")
	providedToken := q.Get("hub.verify_token")

	if providedToken != token || mode != "subscribe" {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(challenge))
}

// webhookError is a custom error type for webhook errors.
type webhookError string

func (e webhookError) Error() string {
	return string(e)
}

const (
	ErrInvalidSignature      = webhookError("signature is invalid")
	ErrSignatureNotFound     = webhookError("signature not found")
	ErrSignatureVerification = webhookError("signature verification failed")
	ErrReadNotification      = webhookError("error reading request body")
	ErrMessageDecode         = webhookError("error decoding message")
	ErrBadRequest            = webhookError("could not retrieve the notification content")
)

// ChangeField represent the name of the field in which the webhook notification payload
// is embedded.
type ChangeField string

const (
	ChangeFieldFlows                    ChangeField = "flows"
	ChangeFieldAccountAlerts            ChangeField = "account_alerts"
	ChangeFieldTemplateStatusUpdate     ChangeField = "message_template_status_update"
	ChangeFieldTemplateCategoryUpdate   ChangeField = "template_category_update"
	ChangeFieldTemplateQualityUpdate    ChangeField = "message_template_quality_update"
	ChangeFieldPhoneNumberNameUpdate    ChangeField = "phone_number_name_update"
	ChangeFieldBusinessCapabilityUpdate ChangeField = "business_capability_update"
	ChangeFieldAccountUpdate            ChangeField = "account_update"
	ChangeFieldAccountReviewUpdate      ChangeField = "account_review_update"
	ChangeFieldPhoneNumberQualityUpdate ChangeField = "phone_number_quality_update"
)

func (c ChangeField) String() string {
	return string(c)
}

// MaxPayloadSize is the maximum payload size for webhooks which is 3MB.
const MaxPayloadSize = 3 * 1024 * 1024 // 3MB
