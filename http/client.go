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
)

type (
	RequestHook func(ctx context.Context, request *http.Request) error

	ResponseHook func(ctx context.Context, response *http.Response) error

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

	return DecodeResponseJSON(response, v)
}

var ErrRequestFailed = errors.New("request failed")

func DecodeResponseJSON(response *http.Response, v interface{}) error {
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
