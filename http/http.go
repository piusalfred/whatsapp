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
	// Request is a struct that holds the details that can be used to make a http request.
	// It is used by the Send function to make a request.
	// It contains Payload which is an interface that can be used to pass any data type
	// to the Send function. Payload is expected to be a struct that can be marshalled
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
		ApiVersion string
		SenderID   string
		Endpoints  []string
	}

	// ResponseHook is a function that takes a Context and *http.Response and returns nothing.
	// It is used to execute a function after a request has been made and response received.
	// Send calls multiple hooks and pass the returned *http.Response Do not close the response body
	// in your hooks implementations as it is closed by the Send function.
	ResponseHook func(ctx context.Context, response *http.Response)
)

// executeResponseHooks take a Context, *http.Response and a slice of ResponseHook and executes
// each hook in the slice.
func executeResponseHooks(ctx context.Context, response *http.Response, hooks []ResponseHook) {
	// range over the hooks and execute each one
	for i := 0; i < len(hooks); i++ {
		hooks[i](ctx, response)
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

func NewRequestWithContext(ctx context.Context, request *Request) (*http.Request, error) {
	if request == nil {
		return nil, errors.New("request cannot be nil")
	}
	var (
		body io.Reader
		req  *http.Request
	)
	if request.Form != nil {
		form := url.Values{}
		for key, value := range request.Form {
			form.Add(key, value)
		}
		body = strings.NewReader(form.Encode())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if request.Payload != nil {
		rdr, err := extractPayloadFromRequest(request.Payload)
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

func extractPayloadFromRequest(payload interface{}) (io.Reader, error) {
	if payload != nil {
		// Payload is expected to be a struct that can be marshalled to json, or a slice
		// of bytes or an io.Reader.
		// We need to check the type of the payload and convert it to an io.Reader
		// if it is not already.
		switch payload.(type) {
		case []byte:
			return bytes.NewReader(payload.([]byte)), nil
		case io.Reader:
			return payload.(io.Reader), nil
		case string:
			return strings.NewReader(payload.(string)), nil
		default:
			payload, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}
			return bytes.NewReader(payload), nil
		}
	}

	// FIXME
	return nil, nil
}

func Send(ctx context.Context, client *http.Client, request *Request, v any, hooks ...ResponseHook) error {
	req, err := NewRequestWithContext(ctx, request)
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	response, err := client.Do(req)
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	// Execute the response hooks
	executeResponseHooks(ctx, response, hooks)

	if response.Body == nil && v == nil {
		return nil
	}

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("http send: %w", err)
	}
	bodyBytes := buff.Bytes()

	// Sometimes when there is an error, the response body is not empty
	// as the error description is returned in the body. So we need to
	// check the status code and the body to determine if there is an error
	isResponseOk := response.StatusCode >= 200 && response.StatusCode <= 299
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
