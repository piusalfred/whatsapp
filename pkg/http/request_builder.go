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
	"io"

	"github.com/piusalfred/whatsapp/pkg/types"
)

// RequestBuilder provides a non-generic, fluent way to construct the
// type-independent parts of a Request. The generic type parameter is only
// needed when the message body is attached (see BuildRequest and
// BuildRequestWithMessage).
type RequestBuilder struct {
	method        string
	baseURL       string
	endpoints     []string
	headers       map[string]string
	queryParams   map[string]string
	bearer        string
	appSecret     string
	secure        bool
	reqType       RequestType
	debugLogLevel DebugLogLevel
	metadata      types.Metadata
	downloadURL   string
	form          *RequestForm
	bodyReader    io.Reader
}

// NewRequestBuilder starts building a new request with the given HTTP method
// and base URL.
func NewRequestBuilder(method, baseURL string) *RequestBuilder {
	return &RequestBuilder{
		method:        method,
		baseURL:       baseURL,
		headers:       make(map[string]string),
		queryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
	}
}

// WithRequestType sets the request type (e.g. send_message, upload_media).
func (b *RequestBuilder) WithRequestType(reqType RequestType) *RequestBuilder {
	b.reqType = reqType
	return b
}

// WithEndpoints appends URL path segments to the request.
func (b *RequestBuilder) WithEndpoints(endpoints ...string) *RequestBuilder {
	b.endpoints = endpoints
	return b
}

// WithBearer sets the Authorization bearer token.
func (b *RequestBuilder) WithBearer(bearer string) *RequestBuilder {
	b.bearer = bearer
	return b
}

// WithHeaders replaces the request headers with the provided map.
func (b *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder {
	b.headers = headers
	return b
}

// WithQueryParams replaces the query parameters with the provided map.
func (b *RequestBuilder) WithQueryParams(params map[string]string) *RequestBuilder {
	b.queryParams = params
	return b
}

// WithAppSecret configures the app secret and toggles secure requests.
func (b *RequestBuilder) WithAppSecret(secret string, secure bool) *RequestBuilder {
	b.appSecret = secret
	b.secure = secure
	return b
}

// WithSecured sets whether the request should be marked as secure.
func (b *RequestBuilder) WithSecured(secure bool) *RequestBuilder {
	b.secure = secure
	return b
}

// WithDebugLogLevel sets the debug logging level for the request.
func (b *RequestBuilder) WithDebugLogLevel(level DebugLogLevel) *RequestBuilder {
	b.debugLogLevel = level
	return b
}

// WithMetadata attaches metadata to the request context.
func (b *RequestBuilder) WithMetadata(metadata types.Metadata) *RequestBuilder {
	b.metadata = metadata
	return b
}

// WithDownloadURL marks this request as a download using the provided URL.
func (b *RequestBuilder) WithDownloadURL(url string) *RequestBuilder {
	b.downloadURL = url
	return b
}

// WithForm sets a multipart form body.
func (b *RequestBuilder) WithForm(form *RequestForm) *RequestBuilder {
	b.form = form
	return b
}

// WithBodyReader sets a raw body reader.
func (b *RequestBuilder) WithBodyReader(r io.Reader) *RequestBuilder {
	b.bodyReader = r
	return b
}

// BuildRequest creates a typed Request[T] from the builder configuration and
// the supplied message body. This is the only place a generic parameter is
// required.
func BuildRequest[T any](b *RequestBuilder, message *T) *Request[T] {
	return &Request[T]{
		Type:           b.reqType,
		Method:         b.method,
		BaseURL:        b.baseURL,
		Endpoints:      b.endpoints,
		Headers:        b.headers,
		QueryParams:    b.queryParams,
		Bearer:         b.bearer,
		AppSecret:      b.appSecret,
		SecureRequests: b.secure,
		debugLogLevel:  b.debugLogLevel,
		Metadata:       b.metadata,
		DownloadURL:    b.downloadURL,
		Message:        message,
		Form:           b.form,
		BodyReader:     b.bodyReader,
	}
}

// BuildAnyRequest creates a Request[any] from the builder configuration. Use
// this when no typed message body is needed (e.g. GET requests or downloads).
func BuildAnyRequest(b *RequestBuilder) *Request[any] {
	return &Request[any]{
		Type:           b.reqType,
		Method:         b.method,
		BaseURL:        b.baseURL,
		Endpoints:      b.endpoints,
		Headers:        b.headers,
		QueryParams:    b.queryParams,
		Bearer:         b.bearer,
		AppSecret:      b.appSecret,
		SecureRequests: b.secure,
		debugLogLevel:  b.debugLogLevel,
		Metadata:       b.metadata,
		DownloadURL:    b.downloadURL,
		Form:           b.form,
		BodyReader:     b.bodyReader,
	}
}
