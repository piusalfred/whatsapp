/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"sort"

	"github.com/piusalfred/whatsapp/pkg/crypto"
	"github.com/piusalfred/whatsapp/pkg/types"
)

type (
	// Request carries all the information needed to construct an HTTP request
	// from a typed domain payload. It is built via [MakeRequest] or the
	// [RequestBuilder] + [Build] pattern.
	//
	// Required fields (no defaults): Method, BaseURL.
	// Optional: all other fields; Headers and QueryParams default to empty
	// maps when constructed via [MakeRequest].
	//
	// Body encoding uses a single-source priority: if Message is non-nil it
	// is JSON-encoded; otherwise, if Form is set it is encoded as multipart;
	// otherwise BodyReader is used raw. Supplying more than one body source
	// causes [RequestWithContext] to return [ErrMultipleBodySources].
	Request[T any] struct {
		Type           RequestType
		Method         string
		Bearer         string
		Headers        map[string]string
		QueryParams    map[string]string
		BaseURL        string
		Endpoints      []string
		Metadata       types.Metadata
		Message        *T
		Form           *RequestForm
		AppSecret      string
		SecureRequests bool
		DownloadURL    string
		BodyReader     io.Reader
		debugLogLevel  DebugLogLevel
	}

	// RequestForm represents a multipart form body. Fields are serialized as
	// form fields; FormFile, if non-nil, attaches a file to the request.
	RequestForm struct {
		Fields   map[string]string
		FormFile *FormFile
	}

	// FormFile describes a file attachment for a [RequestForm]. Type defaults
	// to "application/octet-stream" when empty.
	FormFile struct {
		Name string
		Path string
		Type string
	}

	// RequestOption is a functional option for configuring a [Request].
	// Apply with [MakeRequest] or the chained [RequestBuilder] methods.
	RequestOption[T any] func(request *Request[T])
)

func (request *Request[T]) SetDebugLogLevel(level DebugLogLevel) {
	request.debugLogLevel = level
}

// MakeRequest creates a [Request[T]] with empty Headers/QueryParams maps
// and [DebugLogLevelNone] as defaults. Options are applied in the order
// provided; later options overwrite earlier ones for the same field.
func MakeRequest[T any](method, baseURL string, options ...RequestOption[T]) *Request[T] {
	req := &Request[T]{
		Method:        method,
		BaseURL:       baseURL,
		Headers:       make(map[string]string),
		QueryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
	}

	for _, option := range options {
		if option != nil {
			option(req)
		}
	}

	return req
}

func MakeDownloadRequest[T any](downloadURL string, options ...RequestOption[T]) *Request[T] {
	return MakeRequest(stdhttp.MethodGet, "", append(options, func(r *Request[T]) {
		r.DownloadURL = downloadURL
	})...)
}

// NewRequestWithContext builds and validates a [*net/http.Request] in one
// call. It is a convenience alias for [MakeRequest] followed by
// [RequestWithContext].
func NewRequestWithContext[T any](ctx context.Context, method, baseURL string,
	options ...RequestOption[T],
) (*stdhttp.Request, error) {
	req := MakeRequest(method, baseURL, options...)

	return RequestWithContext(ctx, req)
}

func WithRequestType[T any](requestType RequestType) RequestOption[T] {
	return func(request *Request[T]) {
		request.Type = requestType
	}
}

func WithRequestBearer[T any](bearer string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Bearer = bearer
	}
}

func WithRequestEndpoints[T any](endpoints ...string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Endpoints = endpoints
	}
}

func WithRequestMetadata[T any](metadata types.Metadata) RequestOption[T] {
	return func(request *Request[T]) {
		request.Metadata = metadata
	}
}

func WithRequestHeaders[T any](headers map[string]string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Headers = headers
	}
}

func WithRequestQueryParams[T any](queryParams map[string]string) RequestOption[T] {
	return func(request *Request[T]) {
		request.QueryParams = queryParams
	}
}

func WithRequestMessage[T any](message *T) RequestOption[T] {
	return func(request *Request[T]) {
		request.Message = message
	}
}

func WithRequestForm[T any](form *RequestForm) RequestOption[T] {
	return func(request *Request[T]) {
		request.Form = form
	}
}

// WithRequestAppSecret stores the app secret used to generate an
// appsecret_proof query parameter. The proof is only appended to the
// URL when secure mode is enabled via [WithRequestSecured]; setting the
// secret alone has no visible effect.
func WithRequestAppSecret[T any](appSecret string) RequestOption[T] {
	return func(request *Request[T]) {
		request.AppSecret = appSecret
	}
}

func WithRequestSecured[T any](secured bool) RequestOption[T] {
	return func(request *Request[T]) {
		request.SecureRequests = secured
	}
}

func WithRequestBodyReader[T any](bodyReader io.Reader) RequestOption[T] {
	return func(request *Request[T]) {
		request.BodyReader = bodyReader
	}
}

// URL constructs the final request URL. When [Request.DownloadURL] is
// set it is returned unchanged. Otherwise, query parameters are sorted
// for deterministic output, and the "appsecret_proof" parameter is
// appended when [Request.SecureRequests] is true.
func (request *Request[T]) URL() (string, error) {
	if request.DownloadURL != "" {
		return request.DownloadURL, nil
	}

	parsedURL, err := url.Parse(request.BaseURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	parsedURL = parsedURL.JoinPath(request.Endpoints...)

	q := parsedURL.Query()

	keys := make([]string, 0, len(request.QueryParams))
	for key := range request.QueryParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		q.Set(key, request.QueryParams[key])
	}

	shouldEnableDebugLogging := request.debugLogLevel != DebugLogLevelNone && request.debugLogLevel != ""
	if shouldEnableDebugLogging {
		q.Set("debug", string(request.debugLogLevel))
	}

	if request.SecureRequests {
		proof, proofErr := crypto.GenerateAppSecretProof(request.Bearer, request.AppSecret)
		if proofErr != nil {
			return "", fmt.Errorf("failed to generate app secret proof: %w", proofErr)
		}
		q.Set("appsecret_proof", proof)
	}

	parsedURL.RawQuery = q.Encode()

	return parsedURL.String(), nil
}

// RequestWithContext validates the [Request] (rejecting nil and multiple
// body sources), injects [types.Metadata] into the context, encodes the
// body, and builds a [*net/http.Request]. It is the single entry point
// that all senders use to materialize an HTTP request.
func RequestWithContext[T any](ctx context.Context, req *Request[T]) (*stdhttp.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("request: %w", ErrNilRequest)
	}
	ctx = InjectMessageMetadata(ctx, req.Metadata)

	parsedURL, err := req.URL()
	if err != nil {
		return nil, fmt.Errorf("format url: %w", err)
	}

	var body io.Reader
	var contentType string
	var bodySources int

	if req.Message != nil {
		bodySources++
		encodeResp, encodeErr := EncodePayloadWithContext(ctx, req.Message)
		if encodeErr != nil {
			return nil, fmt.Errorf("failed to encode request payload: %w", encodeErr)
		}
		body = encodeResp.Body
		contentType = encodeResp.ContentType
	}

	if req.Form != nil {
		bodySources++
		encodeResp, encodeErr := EncodePayloadWithContext(ctx, req.Form)
		if encodeErr != nil {
			return nil, fmt.Errorf("failed to encode request payload: %w", encodeErr)
		}
		body = encodeResp.Body
		contentType = encodeResp.ContentType
	}

	if req.BodyReader != nil {
		bodySources++
		body = req.BodyReader
		contentType = "application/octet-stream"
	}

	if bodySources > 1 {
		return nil, ErrMultipleBodySources
	}

	r, err := stdhttp.NewRequestWithContext(ctx, req.Method, parsedURL, body)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	if body != nil {
		r.Header.Set("Content-Type", contentType)
	}

	if req.Bearer != "" {
		r.Header.Set("Authorization", "Bearer "+req.Bearer)
	}

	for key, value := range req.Headers {
		r.Header.Set(key, value)
	}

	return r, nil
}
