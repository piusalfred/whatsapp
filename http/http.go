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

	werrors "github.com/piusalfred/whatsapp/errors"
)

const BaseURL = "https://graph.facebook.com"

type (
	Client struct {
		http          *http.Client
		RequestHooks  []RequestHook
		ResponseHooks []ResponseHook
	}

	ClientOption func(*Client)
)

// NewClient creates a new client with the given options, The client is used
// to create a new http request and send it to the server.
// Example:
//
//	client := NewClient(
//		WithHTTPClient(http.DefaultClient),
//		WithRequestHooks(
//			// Add your request hooks here
//		),
//		WithResponseHooks(
//			// Add your response hooks here
//		),
//	)
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		http: http.DefaultClient,
	}
	for _, option := range options {
		option(client)
	}

	return client
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.http = httpClient
	}
}

func WithRequestHooks(hooks ...RequestHook) ClientOption {
	return func(client *Client) {
		client.RequestHooks = hooks
	}
}

func WithResponseHooks(hooks ...ResponseHook) ClientOption {
	return func(client *Client) {
		client.ResponseHooks = hooks
	}
}

// SetRequestHooks sets the request hooks for the client, This removes any previously set request hooks.
// and replaces it with the new ones.
func (client *Client) SetRequestHooks(hooks ...RequestHook) {
	client.RequestHooks = hooks
}

// AppendRequestHooks appends the request hooks to the client, This adds the new request hooks to the
// existing ones.
func (client *Client) AppendRequestHooks(hooks ...RequestHook) {
	client.RequestHooks = append(client.RequestHooks, hooks...)
}

// PrependRequestHooks prepends the request hooks to the client, This adds the new request hooks to the
// existing ones.
func (client *Client) PrependRequestHooks(hooks ...RequestHook) {
	if hooks == nil {
		return
	}
	client.RequestHooks = append(hooks, client.RequestHooks...)
}

// SetResponseHooks sets the response hooks for the client, This removes any previously set response hooks.
// and replaces it with the new ones.
func (client *Client) SetResponseHooks(hooks ...ResponseHook) {
	client.ResponseHooks = hooks
}

// AppendResponseHooks appends the response hooks to the client, This adds the new response hooks to the
// existing ones.
func (client *Client) AppendResponseHooks(hooks ...ResponseHook) {
	client.ResponseHooks = append(client.ResponseHooks, hooks...)
}

// PrependResponseHooks prepends the response hooks to the client, This adds the new response hooks to the
// existing ones.
func (client *Client) PrependResponseHooks(hooks ...ResponseHook) {
	if hooks == nil {
		return
	}
	client.ResponseHooks = append(hooks, client.ResponseHooks...)
}

// Do send a http request to the server and returns the response, It accepts a context,
// a request and a pointer to a variable to decode the response into.
func (client *Client) Do(ctx context.Context, r *Request, v any) error {
	request, err := prepareRequest(ctx, r, client.RequestHooks...)
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading response body: %w", err)
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err = runResponseHooks(ctx, response, client.ResponseHooks...); err != nil {
		return fmt.Errorf("response hooks: %w", err)
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return decodeResponseJson(response, v)
}

var ErrRequestFailed = errors.New("request failed")

func decodeResponseJson(response *http.Response, v interface{}) error {
	if v == nil || response == nil {
		return nil
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed

	if !isResponseOk {
		if len(responseBody) == 0 {
			return fmt.Errorf("%w: status code: %d", ErrRequestFailed, response.StatusCode)
		}

		var errorResponse ResponseError
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return fmt.Errorf("error decoding response error body: %w", err)
		}

		return &errorResponse
	}

	if len(responseBody) != 0 {
		if err := json.Unmarshal(responseBody, v); err != nil {
			return fmt.Errorf("error decoding response body: %w", err)
		}
	}

	return nil
}

func runResponseHooks(ctx context.Context, response *http.Response, hooks ...ResponseHook) error {
	for _, hook := range hooks {
		if hook != nil {
			if err := hook(ctx, response); err != nil {
				return fmt.Errorf("response hooks: %w", err)
			}
		}
	}

	return nil
}

func prepareRequest(ctx context.Context, r *Request, hooks ...RequestHook) (*http.Request, error) {
	// create a new request, run hooks and return the request after restoring the body
	ctx = withRequestName(ctx, r.Context.Name)
	request, err := NewRequestWithContext(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}
	// run request hooks
	for _, hook := range hooks {
		if hook != nil {
			if err = hook(ctx, request); err != nil {
				return nil, fmt.Errorf("prepare request: %w", err)
			}
		}
	}
	// restore the request body
	body, err := r.BodyBytes()
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}
	request.Body = io.NopCloser(bytes.NewBuffer(body))

	return request, nil
}

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

// ResponseDecoder decodes the response body into the given interface.
type ResponseDecoder interface {
	DecodeResponse(response *http.Response, v interface{}) error
}

// ResponseDecoderFunc is an adapter to allow the use of ordinary functions as
// response decoders. If f is a function with the appropriate signature,
// ResponseDecoderFunc(f) is a ResponseDecoder that calls f.
type ResponseDecoderFunc func(response *http.Response, v interface{}) error

// DecodeResponse calls f(ctx, response, v).
func (f ResponseDecoderFunc) DecodeResponse(response *http.Response, v interface{}) error {
	return f(response, v)
}

type (
	RequestHook  func(ctx context.Context, request *http.Request) error
	ResponseHook func(ctx context.Context, response *http.Response) error
)

func LogRequestHook(logger *slog.Logger) RequestHook {
	return func(ctx context.Context, request *http.Request) error {
		name := RequestNameFromContext(ctx)
		reader, err := request.GetBody()
		if err != nil {
			return fmt.Errorf("log request hook: %w", err)
		}
		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(reader); err != nil {
			return fmt.Errorf("log request hook: %w", err)
		}

		hb := &strings.Builder{}
		hb.WriteString("[")
		for k, v := range request.Header {
			hb.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		}
		hb.WriteString("]")

		logger.LogAttrs(ctx, slog.LevelDebug, "request", slog.String("name", name),
			slog.String("body", buf.String()), slog.String("headers", hb.String()))

		return nil
	}
}

func LogResponseHook(logger *slog.Logger) ResponseHook {
	return func(ctx context.Context, response *http.Response) error {
		name := RequestNameFromContext(ctx)
		reader := response.Body
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(reader); err != nil {
			return fmt.Errorf("log response hook: %w", err)
		}

		hb := &strings.Builder{}
		hb.WriteString("[")
		for k, v := range response.Header {
			hb.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		}
		hb.WriteString("]")

		logger.LogAttrs(ctx, slog.LevelDebug, "response", slog.String("name", name),
			slog.Int("status", response.StatusCode),
			slog.String("status_text", response.Status),
			slog.String("headers", hb.String()),
			slog.String("body", buf.String()),
		)

		return nil
	}
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
