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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

var _ slog.LogValuer = (*Request)(nil)

type (
	BasicAuth struct {
		Username string
		Password string
	}

	// requestNameKey is a type that holds the name of a request. This is usually passed
	// extracted from Request.Context.Name and passed down to the Do function.
	// then passed down with to the request hooks. In request hooks, the name can be
	// used to identify the request and other multiple use cases like instrumentation,
	// logging etc.
	requestNameKey string

	// Request is a struct that holds the details that can be used to make a http request.
	// It is used by the Do function to make a request.
	// It contains Payload which is an interface that can be used to pass any data type
	// to the Do function. Payload is expected to be a struct that can be marshalled
	// to json, or a slice of bytes or an io.Reader.
	Request struct {
		Context   *RequestContext
		Method    string
		BasicAuth *BasicAuth
		Headers   map[string]string
		Query     map[string]string
		Bearer    string
		Form      map[string]string
		Payload   any
	}

	RequestOption func(*Request)

	RequestContext struct {
		Name              string
		BaseURL           string
		ApiVersion        string //nolint: revive,stylecheck
		PhoneNumberID     string
		Bearer            string
		BusinessAccountID string
		Endpoints         []string
	}
)

func (request *Request) LogValue() slog.Value {
	if request == nil {
		return slog.StringValue("nil")
	}
	var reqURL string
	if request.Context != nil {
		reqURL, _ = url.JoinPath(request.Context.BaseURL, request.Context.Endpoints...)
	}
	value := slog.GroupValue(
		slog.String("name", request.Context.Name),
		slog.String("method", request.Method),
		slog.String("url", reqURL),
	)

	return value
}

// MakeRequest takes a context.Context and a slice of RequestOption and returns a new *http.Request.
// Internally, it calls the NewRequestWithContext function.
func MakeRequest(options ...RequestOption) *Request {
	request := &Request{
		Context: &RequestContext{
			BaseURL:    BaseURL,
			ApiVersion: "v16.0",
		},
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
	}
	for _, option := range options {
		option(request)
	}

	return request
}

func WithBasicAuth(username, password string) RequestOption {
	return func(request *Request) {
		request.BasicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

func WithRequestName(name string) RequestOption {
	return func(request *Request) {
		request.Context.Name = name
	}
}

func WithForm(form map[string]string) RequestOption {
	return func(request *Request) {
		request.Form = form
	}
}

func WithPayload(payload any) RequestOption {
	return func(request *Request) {
		request.Payload = payload
	}
}

func WithMethod(method string) RequestOption {
	return func(request *Request) {
		request.Method = method
	}
}

func WithHeaders(headers map[string]string) RequestOption {
	return func(request *Request) {
		request.Headers = headers
	}
}

func WithQuery(query map[string]string) RequestOption {
	return func(request *Request) {
		request.Query = query
	}
}

func WithEndpoints(endpoints ...string) RequestOption {
	return func(request *Request) {
		request.Context.Endpoints = endpoints
	}
}

func WithBearer(bearer string) RequestOption {
	return func(request *Request) {
		request.Bearer = bearer
	}
}

// ReaderFunc is a function that takes a *Request and returns a func that takes nothing
// but returns an io.Reader and an error.
func (request *Request) ReaderFunc() func() (io.Reader, error) {
	return func() (io.Reader, error) {
		return extractRequestBody(request.Payload)
	}
}

// BodyBytes takes a *Request and returns a slice of bytes or an error.
func (request *Request) BodyBytes() ([]byte, error) {
	if request.Payload == nil {
		return nil, nil
	}
	body, err := request.ReaderFunc()()
	if err != nil {
		return nil, fmt.Errorf("reader func: %w", err)
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return nil, fmt.Errorf("read from: %w", err)
	}

	return buf.Bytes(), nil
}

var ErrInvalidRequestValue = errors.New("invalid request value")

// NewRequestWithContext takes a context and a *Request and returns a new *http.Request.
func NewRequestWithContext(ctx context.Context, request *Request) (*http.Request, error) {
	if request == nil || request.Context == nil {
		return nil, fmt.Errorf("%w: request or request context should not be nil", ErrInvalidRequestValue)
	}
	var (
		body    io.Reader
		req     *http.Request
		headers = map[string]string{}
	)
	if request.Form != nil {
		form := url.Values{}
		for key, value := range request.Form {
			form.Add(key, value)
		}
		body = strings.NewReader(form.Encode())
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	} else if request.Payload != nil {
		rdr, err := extractRequestBody(request.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to extract payload from request: %w", err)
		}

		body = rdr
	}

	requestURL, err := RequestURLFromContext(request.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to create request url: %w", err)
	}

	// Create the http request
	req, err = http.NewRequestWithContext(ctx, request.Method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	if request.BasicAuth != nil {
		req.SetBasicAuth(request.BasicAuth.Username, request.BasicAuth.Password)
	}

	// Set the request headers
	if request.Headers != nil {
		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}
	}

	// extra headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Set the bearer token header
	if request.Bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", request.Bearer))
	}

	// Add the query parameters to the request URL
	if request.Query != nil {
		query := req.URL.Query()
		for key, value := range request.Query {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

// extractRequestBody takes an interface{} and returns an io.Reader.
// It is called by the NewRequestWithContext function to convert the payload in the
// Request to an io.Reader. The io.Reader is then used to set the body of the http.Request.
// Only the following types are supported:
// 1. []byte
// 2. io.Reader
// 3. string
// 4. any value that can be marshalled to json
// 5. nil.
func extractRequestBody(payload interface{}) (io.Reader, error) {
	if payload == nil {
		return nil, nil
	}
	switch p := payload.(type) {
	case []byte:
		return bytes.NewReader(p), nil
	case io.Reader:
		return p, nil
	case string:
		return strings.NewReader(p), nil
	default:
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(p)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload: %w", err)
		}

		return buf, nil
	}
}

// withRequestName takes a string and a context and returns a new context with the string
// as the request name.
func withRequestName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, requestNameKey("request-name"), name)
}

// RequestNameFromContext returns the request name from the context.
func RequestNameFromContext(ctx context.Context) string {
	name, ok := ctx.Value(requestNameKey("request-name")).(string)
	if !ok {
		return "unknown request name"
	}

	return name
}

// RequestURLFromContext returns the request url from the context.
func RequestURLFromContext(ctx *RequestContext) (string, error) {
	elems := append([]string{ctx.ApiVersion, ctx.PhoneNumberID}, ctx.Endpoints...)
	path, err := url.JoinPath(ctx.BaseURL, elems...)
	if err != nil {
		return "", fmt.Errorf("failed to join url path: %w", err)
	}

	return path, nil
}
