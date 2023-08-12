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
	"net/http"
	"net/url"
	"strings"

	werrors "github.com/piusalfred/whatsapp/errors"
)

const BaseURL = "https://graph.facebook.com"

type (
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
		Context *RequestContext
		Method  string
		Headers map[string]string
		Query   map[string]string
		Bearer  string
		Form    map[string]string
		Payload any
	}

	RequestOption func(*Request)

	RequestContext struct {
		Name       string
		BaseURL    string
		ApiVersion string //nolint: revive,stylecheck
		SenderID   string
		Endpoints  []string
	}

	// Hook is a function that takes a Context, a *http.Request and *http.Response and returns nothing.
	// It exposes the request and response to the user.
	// Do calls multiple hooks and pass the returned *http.Response Do not close the response body
	// in your hooks implementations as it is closed by the Do function.
	Hook func(ctx context.Context, request *http.Request, response *http.Response)
)

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

// executeHooks take a Context,*http.Request, *http.Response and a slice of Hook and executes
// each hook in the slice.
func executeHooks(ctx context.Context, request *http.Request, response *http.Response, hooks []Hook) {
	// range over the hooks and execute each one
	for i := 0; i < len(hooks); i++ {
		hooks[i](ctx, request, response)
	}
}

// requestURLFromContext joins the elements of url parts into a single url string.
func requestURLFromContext(parts *RequestContext) (string, error) {
	return CreateRequestURL(parts.BaseURL, parts.ApiVersion, parts.SenderID, parts.Endpoints...)
}

// CreateRequestURL creates a new request url by joining the base url, api version
// sender id and endpoints.
// It is called by the NewRequestWithContext function where these details are
// passed from the Request.
func CreateRequestURL(baseURL, apiVersion, senderID string, endpoints ...string) (string, error) {
	elems := append([]string{apiVersion, senderID}, endpoints...)

	path, err := url.JoinPath(baseURL, elems...)
	if err != nil {
		return "", fmt.Errorf("failed to join url path: %w", err)
	}

	return path, nil
}

// NewRequest takes a context.Context and a slice of RequestOption and returns a new *http.Request.
// Internally, it calls the NewRequestWithContext function.
func NewRequest(ctx context.Context, options ...RequestOption) (*http.Request, error) {
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

	return NewRequestWithContext(ctx, request)
}

func WithContext(ctx *RequestContext) RequestOption {
	return func(request *Request) {
		request.Context = ctx
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
//
//nolint:cyclop
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

	requestURL, err := requestURLFromContext(request.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to create request url: %w", err)
	}

	// Create the http request
	req, err = http.NewRequestWithContext(ctx, request.Method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
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

// Do calls http.Client.Do after creating a *http.Request. It takes hooks that are executed after the request
// is sent. Even when an error occurs, the hooks are executed. Except for the errors caused by NewRequestWithContext,
// here Do terminates the execution and returns the error. Hooks passed here should always check for nil values of
// the requests and responses before using them. As in some cases hooks may be called with nil values. Example when
// http.Client.Do returns an error.
//
//nolint:cyclop
func Do(ctx context.Context, client *http.Client, r *Request, v any, hooks ...Hook) error {
	ctx = withRequestName(ctx, r.Context.Name)
	request, err := NewRequestWithContext(ctx, r)
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}
	reqBodyBytes, err := r.BodyBytes()
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		defer executeHooks(ctx, request, response, hooks)
		request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))

		return fmt.Errorf("http send: %w", err)
	}
	defer func() {
		// restore the request body
		request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		executeHooks(ctx, request, response, hooks)
		_ = response.Body.Close()
	}()

	if v == nil {
		return nil
	}

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("http send: %w", err)
	}
	bodyBytes := buff.Bytes()

	// restore the response body
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Sometimes when there is an error, the response body is not empty
	// as the error description is returned in the body. So we need to
	// check the status code and the body to determine if there is an error
	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed
	bodyIsEmpty := len(bodyBytes) == 0
	if !isResponseOk && !bodyIsEmpty {
		var errResponse ResponseError
		if err = json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&errResponse); err != nil {
			return fmt.Errorf("http send: status (%d): body (%s): %w", response.StatusCode, string(bodyBytes), err)
		}
		errResponse.Code = response.StatusCode

		return &errResponse
	}

	// Response is OK and the body is available
	if isResponseOk && !bodyIsEmpty {
		if v != nil {
			if err = json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(v); err != nil {
				return fmt.Errorf("http send: status (%d): body (%s): %w", response.StatusCode, string(bodyBytes), err)
			}

			return nil
		}
	}

	return nil
}

type ResponseError struct {
	Code int            `json:"code,omitempty"`
	Err  *werrors.Error `json:"error,omitempty"`
}

// Error returns the error message for ResponseError.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("whatsapp error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

// Unwrap returns the underlying error for ResponseError.
func (e *ResponseError) Unwrap() error {
	return e.Err
}
