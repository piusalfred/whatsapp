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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/config"
)

const (
	EndpointMessages  = "messages"
	EndpointTemplates = "message_templates"
)

type (

	// Request is a struct that holds the details that can be used to make a http request.
	// It is used by the Do function to make a request.
	// It contains Payload which is an interface that can be used to pass any data type
	// to the Do function. Payload is expected to be a struct that can be marshalled
	// to json, or a slice of bytes or an io.Reader.
	Request struct {
		Context   *RequestContext
		Method    string
		Headers   map[string]string
		Query     map[string]string
		Form      map[string]string
		Payload   any
		Endpoints []string
	}

	RequestOption func(*Request)
)

// MakeRequest creates a new request with the given options. Default values are used
// for the request if no options are passed.
// The default values are:
// 1. RequestContext.BaseURL: https://graph.facebook.com
// 2. RequestContext.ApiVersion: v16.0
// 3. Request.Method: http.MethodPost
// 4. Request.Headers: map[string]string{"Content-MessageType": "application/json"}
// 5. RequestContext.Endpoints: []string{"/messages"}
//
// Most importantly remember to set the request payload and the request type and the
// bearer token and phone number id if needed.
func MakeRequest(options ...RequestOption) *Request {
	req := &Request{
		Context: &RequestContext{
			Action:   RequestActionSend,
			Category: RequestCategoryMessage,
			Name:     RequestNameTextMessage,
		},
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-MessageType": "application/json"},
	}
	for _, option := range options {
		option(req)
	}

	if req.Context.CacheOptions != nil {
		co := req.Context.CacheOptions
		if co.CacheControl != "" {
			req.Headers["Cache-Control"] = co.CacheControl
		} else if co.Expires > 0 {
			req.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", co.Expires)
		}
		if co.LastModified != "" {
			req.Headers["Last-Modified"] = co.LastModified
		}
		if co.ETag != "" {
			req.Headers["ETag"] = co.ETag
		}
	}

	return req
}

func WithRequestContext(ctx *RequestContext) RequestOption {
	return func(request *Request) {
		request.Context = ctx
	}
}

func WithRequestForm(form map[string]string) RequestOption {
	return func(request *Request) {
		request.Form = form
	}
}

func WithRequestPayload(payload any) RequestOption {
	return func(request *Request) {
		request.Payload = payload
	}
}

func WithRequestMethod(method string) RequestOption {
	return func(request *Request) {
		request.Method = method
	}
}

func WithRequestHeaders(headers map[string]string) RequestOption {
	return func(request *Request) {
		request.Headers = headers
	}
}

func WithRequestQuery(query map[string]string) RequestOption {
	return func(request *Request) {
		request.Query = query
	}
}

func WithRequestEndpoints(endpoints ...string) RequestOption {
	return func(request *Request) {
		request.Context.Endpoints = endpoints
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

// RequestURLFmt returns the request url from the context.
func RequestURLFmt(values *config.Values, request *RequestContext) (string, error) {
	if request == nil {
		return "", fmt.Errorf("%w: request should not be nil", ErrInvalidRequestValue)
	}

	if request.Category == RequestCategoryMessage {
		return fmtRequestURL(values.BaseURL, values.Version, values.PhoneNumberID, EndpointMessages)
	}

	if request.Category == RequestCategoryTemplates {
		return fmtRequestURL(values.BaseURL, values.Version, values.BusinessAccountID, EndpointTemplates)
	}

	elems := append([]string{values.BusinessAccountID}, request.Endpoints...)

	return fmtRequestURL(values.BaseURL, values.Version, elems...)
}

// fmtRequestURL returns the request url. It accepts base url, api version, endpoints.
func fmtRequestURL(baseURL, apiVersion string, endpoints ...string) (string, error) {
	elems := append([]string{apiVersion}, endpoints...)
	path, err := url.JoinPath(baseURL, elems...)
	if err != nil {
		return "", fmt.Errorf("failed to join url path: %w", err)
	}

	return path, nil
}

type RequestCategory string

const (
	RequestCategoryMessage      RequestCategory = "message"
	RequestCategoryPhoneNumbers RequestCategory = "phone numbers"
	RequestCategoryQRCodes      RequestCategory = "qr codes"
	RequestCategoryMedia        RequestCategory = "media"
	RequestCategoryWebhooks     RequestCategory = "webhooks"
	RequestCategoryVerification RequestCategory = "verification"
	RequestCategoryTemplates    RequestCategory = "templates"
)

type RequestAction string

const (
	RequestActionSend     RequestAction = "send"
	RequestActionList     RequestAction = "list"
	RequestActionCreate   RequestAction = "create"
	RequestActionDelete   RequestAction = "delete"
	RequestActionUpdate   RequestAction = "update"
	RequestActionGet      RequestAction = "get"
	RequestActionUpload   RequestAction = "upload"
	RequestActionDownload RequestAction = "download"
	RequestActionRead     RequestAction = "read"
	RequestActionVerify   RequestAction = "verify"
	RequestActionReact    RequestAction = "react"
)

type RequestName string

const (
	RequestNameTextMessage RequestName = "text"
	RequestNameLocation    RequestName = "location"
	RequestNameMedia       RequestName = "media"
	RequestNameTemplate    RequestName = "template"
	RequestNameReaction    RequestName = "reaction"
	RequestNameContacts    RequestName = "contacts"
	RequestNameInteractive RequestName = "interactive"
	RequestNameAudio       RequestName = "audio"
	RequestNameDocument    RequestName = "document"
	RequestNameImage       RequestName = "image"
	RequestNameVideo       RequestName = "video"
	RequestNameSticker     RequestName = "sticker"
)

type (
	/*	   CacheOptions contains the options on how to send a media message. You can specify either the
		   ID or the link of the media. Also, it allows you to specify caching options.

		   The Cloud API supports media http caching. If you are using a link (link) to a media asset on your
		   server (as opposed to the ID (id) of an asset you have uploaded to our servers),you can instruct us
		   to cache your asset for reuse with future messages by including the headers below
		   in your server Resp when we request the asset. If none of these headers are included, we will
		   not cache your asset.

		   	Cache-Control: <CACHE_CONTROL>
		   	Last-Modified: <LAST_MODIFIED>
		   	ETag: <ETAG>

		   # CacheControl

		   The Cache-Control header tells us how to handle asset caching. We support the following directives:

		   	max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
		   	messages until this time is exceeded, after which we will request the asset again, if needed.
		   	Example: Cache-Control: max-age=604800.

		   	no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
		   	is different from a previous Resp.Requires the Last-Modified header.
		   	Example: Cache-Control: no-cache.

		   	no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.

		   	private: Indicates that the asset is personalized for the recipient and should not be cached.

		   # LastModified

		   Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache. If the
		   goLast-Modified value
		   is different from a previous Resp and Cache-Control: no-cache is included in the Resp,
		   we will update our cached ApiVersion of the asset with the asset in the Resp.
		   Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

		   # ETag

		   The ETag header is a unique string that identifies a specific ApiVersion of an asset.
		   Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified headers
		   are not included in the Resp. In this case, we will cache the asset according to our own, internal
		   logic (which we do not disclose).
	*/
	CacheOptions struct {
		CacheControl string `json:"cache_control,omitempty"`
		LastModified string `json:"last_modified,omitempty"`
		ETag         string `json:"etag,omitempty"`
		Expires      int64  `json:"expires,omitempty"`
	}
	RequestContext struct {
		ID           string
		Action       RequestAction
		Category     RequestCategory
		Name         RequestName
		Endpoints    []string
		Metadata     map[string]string
		CacheOptions *CacheOptions
	}

	RequestContextOption func(*RequestContext)
)

// String returns the string representation of the request context.
func (requestContext *RequestContext) String() string {
	if requestContext == nil {
		return "request context: nil"
	}

	return fmt.Sprintf("request context: [id: %s, action: %s, category: %s, name: %s, endpoints: %v, metadata: %v]",
		requestContext.ID, requestContext.Action,
		requestContext.Category, requestContext.Name,
		requestContext.Endpoints, requestContext.Metadata)
}

// MakeRequestContext creates a new request context with the given options.
func MakeRequestContext(options ...RequestContextOption) *RequestContext {
	requestContext := &RequestContext{
		ID:        "",
		Action:    RequestActionSend,
		Category:  RequestCategoryMessage,
		Name:      RequestNameTextMessage,
		Endpoints: []string{"/messages"},
		Metadata:  nil,
	}
	for _, option := range options {
		option(requestContext)
	}

	return requestContext
}

func WithRequestContextCacheOptions(cacheOptions *CacheOptions) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.CacheOptions = cacheOptions
	}
}

func WithRequestContextID(id string) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.ID = id
	}
}

func WithRequestContextAction(action RequestAction) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.Action = action
	}
}

func WithRequestContextCategory(category RequestCategory) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.Category = category
	}
}

func WithRequestContextName(name RequestName) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.Name = name
	}
}

func WithRequestContextEndpoints(endpoints ...string) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.Endpoints = endpoints
	}
}

func WithRequestContextMetadata(metadata map[string]string) RequestContextOption {
	return func(requestContext *RequestContext) {
		requestContext.Metadata = metadata
	}
}

// requestContextKey is the key used to store the request context in the request context.
type requestContextKey string

// requestContextValue is the value used to store the request context in the request context.
const requestContextValue = "github.com/piusalfred/whatsapp/pkg/http/request_context"

// attachRequestContext takes a request context and a context and returns a new
// context with the request context.
func attachRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey(requestContextValue), reqCtx)
}

// RetrieveRequestContext returns the request context from the context.
func RetrieveRequestContext(ctx context.Context) *RequestContext {
	reqCtx, ok := ctx.Value(requestContextKey(requestContextValue)).(*RequestContext)
	if !ok {
		return nil
	}

	return reqCtx
}
