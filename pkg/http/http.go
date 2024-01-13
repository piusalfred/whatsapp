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
	"sync"

	"github.com/piusalfred/whatsapp/pkg/config"
)

var (
	ErrInvalidRequestValue = errors.New("invalid request value")
	ErrRequestFailed       = errors.New("request failed")
)

type (
	BaseClient struct {
		mu            *sync.Mutex
		http          *http.Client
		requestHooks  []RequestHook
		responseHooks []ResponseHook
		mw            []SendMiddleware
		errorChannel  chan error
		config        *config.Values
	}

	BaseClientOption func(*BaseClient)

	RequestHook  func(ctx context.Context, request *http.Request) error
	ResponseHook func(ctx context.Context, response *http.Response) error
	SenderFunc   func(ctx context.Context, r *Request) (*ResponseMessage, error)

	Sender interface {
		Send(ctx context.Context, r *Request) (*ResponseMessage, error)
	}

	SendMiddleware func(Sender) Sender
)

func wrapSender(sender Sender, middleware ...SendMiddleware) Sender {
	for i := len(middleware) - 1; i >= 0; i-- {
		sender = middleware[i](sender)
	}

	return sender
}

func (f SenderFunc) Send(ctx context.Context, r *Request) (*ResponseMessage, error) {
	return f(ctx, r)
}

// InitBaseClient initializes the whatsapp cloud api http client.
func InitBaseClient(ctx context.Context, c config.Reader, opts ...BaseClientOption) (*BaseClient, error) {
	values, err := c.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("init client: load values: %w", err)
	}

	client := &BaseClient{
		mu:            &sync.Mutex{},
		http:          http.DefaultClient,
		requestHooks:  []RequestHook{},
		responseHooks: []ResponseHook{},
		errorChannel:  make(chan error),
		config:        values,
	}

	client.ApplyOptions(opts...)

	return client, nil
}

// ApplyOptions applies the options to the client.
func (client *BaseClient) ApplyOptions(opts ...BaseClientOption) {
	for _, opt := range opts {
		opt(client)
	}
}

// ReloadConfig reloads the config for the client.
func (client *BaseClient) ReloadConfig(ctx context.Context, c config.Reader) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	values, err := c.Read(ctx)
	if err != nil {
		return fmt.Errorf("reload config: load values: %w", err)
	}

	client.config = values

	return nil
}

// ListenErrors takes a func(error) and returns nothing.
// Every error sent to the client's errorChannel will be passed to the function.
func (client *BaseClient) ListenErrors(errorHandler func(error)) {
	for err := range client.errorChannel {
		errorHandler(err)
	}
}

// Close closes the client.
func (client *BaseClient) Close() error {
	close(client.errorChannel)

	return nil
}

func WithHTTPClient(httpClient *http.Client) BaseClientOption {
	return func(client *BaseClient) {
		client.http = httpClient
	}
}

func WithRequestHooks(hooks ...RequestHook) BaseClientOption {
	return func(client *BaseClient) {
		client.requestHooks = hooks
	}
}

func WithResponseHooks(hooks ...ResponseHook) BaseClientOption {
	return func(client *BaseClient) {
		client.responseHooks = hooks
	}
}

// WithSendMiddleware adds a middleware to the base client.
func WithSendMiddleware(mw ...SendMiddleware) BaseClientOption {
	return func(client *BaseClient) {
		client.mw = append(client.mw, mw...)
	}
}

// SetRequestHooks sets the request hooks for the client, This removes any previously set request hooks.
// and replaces it with the new ones.
func (client *BaseClient) SetRequestHooks(hooks ...RequestHook) {
	client.requestHooks = hooks
}

// AppendRequestHooks appends the request hooks to the client, This adds the new request hooks to the
// existing ones.
func (client *BaseClient) AppendRequestHooks(hooks ...RequestHook) {
	client.requestHooks = append(client.requestHooks, hooks...)
}

// PrependRequestHooks prepends the request hooks to the client, This adds the new request hooks to the
// existing ones.
func (client *BaseClient) PrependRequestHooks(hooks ...RequestHook) {
	if hooks == nil {
		return
	}
	client.requestHooks = append(hooks, client.requestHooks...)
}

// SetResponseHooks sets the response hooks for the client, This removes any previously set response hooks.
// and replaces it with the new ones.
func (client *BaseClient) SetResponseHooks(hooks ...ResponseHook) {
	client.responseHooks = hooks
}

// AppendResponseHooks appends the response hooks to the client, This adds the new response hooks to the
// existing ones.
func (client *BaseClient) AppendResponseHooks(hooks ...ResponseHook) {
	client.responseHooks = append(client.responseHooks, hooks...)
}

// PrependResponseHooks prepends the response hooks to the client, This adds the new response hooks to the
// existing ones.
func (client *BaseClient) PrependResponseHooks(hooks ...ResponseHook) {
	if hooks == nil {
		return
	}
	client.responseHooks = append(hooks, client.responseHooks...)
}

// Send sends a message to the server. This is used to Send requests to /messages endpoint.
// They all have the same response structure as represented by ResponseMessage.
func (client *BaseClient) Send(ctx context.Context, r *Request) (*ResponseMessage, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	sender := wrapSender(SenderFunc(client.send), client.mw...)
	resp, err := sender.Send(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("base client: %w", err)
	}

	return resp, nil
}

func (client *BaseClient) send(ctx context.Context, r *Request) (*ResponseMessage, error) {
	request, err := client.prepareRequest(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	response, err := client.http.Do(request) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("http send: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// send error to error channel
			client.errorChannel <- fmt.Errorf("closing response body: %w", err)
		}
	}(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err = client.runResponseHooks(ctx, response); err != nil {
		return nil, fmt.Errorf("response hooks: %w", err)
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var resp ResponseMessage

	if err := decodeResponseJSON(response, &resp); err != nil {
		return nil, fmt.Errorf("response decoder: %w", err)
	}

	return &resp, nil
}

func (client *BaseClient) runResponseHooks(ctx context.Context, response *http.Response) error {
	for _, hook := range client.responseHooks {
		if hook != nil {
			if err := hook(ctx, response); err != nil {
				return fmt.Errorf("response hooks: %w", err)
			}
		}
	}

	return nil
}

// authorize attaches the bearer token to the request.
func (client *BaseClient) authorize(ctx *RequestContext, request *http.Request) error {
	if request == nil || ctx == nil {
		return fmt.Errorf("%w: request or request context should not be nil", ErrInvalidRequestValue)
	}

	if ctx.Category == RequestCategoryMessage {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.config.AccessToken))

		return nil
	}

	return nil
}

// prepareRequest prepares the request for sending.
func (client *BaseClient) prepareRequest(ctx1 context.Context, r *Request) (*http.Request, error) {
	ctx := attachRequestContext(ctx1, r.Context)
	request, err := client.newHTTPRequest(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	// run request hooks
	for _, hook := range client.requestHooks {
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

// NewRequestWithContext takes a context and a *Request and returns a new *http.Request.
func (client *BaseClient) newHTTPRequest(ctx context.Context, request *Request) (*http.Request, error) {
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
		headers["Content-MessageType"] = "application/x-www-form-urlencoded"
	} else if request.Payload != nil {
		rdr, err := extractRequestBody(request.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to extract payload from request: %w", err)
		}

		body = rdr
	}

	requestURL, err := RequestURLFmt(client.config, request.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to create request url: %w", err)
	}

	// CreateQR the http request
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

	// Add the query parameters to the request URL
	if request.Query != nil {
		query := req.URL.Query()
		for key, value := range request.Query {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	// authorize the request
	if err = client.authorize(request.Context, req); err != nil {
		return nil, fmt.Errorf("failed to authorize request: %w", err)
	}

	return req, nil
}

func decodeResponseJSON(response *http.Response, v interface{}) error {
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
