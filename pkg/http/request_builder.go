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
	stdhttp "net/http"

	"github.com/piusalfred/whatsapp/pkg/types"
)

// RequestBuilder constructs a [Request[T]] without requiring a type parameter
// until the message body is attached. This eliminates the repetitive [T]
// annotations on type-independent options like Bearer or Headers.
//
// Usage:
//
//	req := http.NewRequestBuilder(http.MethodPost, "https://api.example.com").
//	    Bearer("token").
//	    Endpoints("v20", phoneID).
//	    Message(&MyMessage{Text: "hello"}) // T inferred
type RequestBuilder struct {
	method        string
	baseURL       string
	reqType       RequestType
	bearer        string
	headers       map[string]string
	queryParams   map[string]string
	endpoints     []string
	metadata      types.Metadata
	appSecret     string
	secure        bool
	downloadURL   string
	bodyReader    io.Reader
	form          *RequestForm
	debugLogLevel DebugLogLevel
}

// NewRequestBuilder starts building a request with the given method and base URL.
func NewRequestBuilder(method, baseURL string) *RequestBuilder {
	return &RequestBuilder{
		method:        method,
		baseURL:       baseURL,
		headers:       make(map[string]string),
		queryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
	}
}

// NewDownloadBuilder starts building a download request from a pre-signed URL.
func NewDownloadBuilder(downloadURL string) *RequestBuilder {
	return &RequestBuilder{
		method:        stdhttp.MethodGet,
		downloadURL:   downloadURL,
		headers:       make(map[string]string),
		queryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
	}
}

func (b *RequestBuilder) Type(t RequestType) *RequestBuilder          { b.reqType = t; return b }
func (b *RequestBuilder) Bearer(token string) *RequestBuilder         { b.bearer = token; return b }
func (b *RequestBuilder) Headers(h map[string]string) *RequestBuilder { b.headers = h; return b }
func (b *RequestBuilder) QueryParams(q map[string]string) *RequestBuilder {
	b.queryParams = q
	return b
}
func (b *RequestBuilder) Endpoints(e ...string) *RequestBuilder     { b.endpoints = e; return b }
func (b *RequestBuilder) Metadata(m types.Metadata) *RequestBuilder { b.metadata = m; return b }
func (b *RequestBuilder) AppSecret(s string) *RequestBuilder        { b.appSecret = s; return b }
func (b *RequestBuilder) Secured(v bool) *RequestBuilder            { b.secure = v; return b }
func (b *RequestBuilder) BodyReader(r io.Reader) *RequestBuilder    { b.bodyReader = r; return b }
func (b *RequestBuilder) Form(f *RequestForm) *RequestBuilder       { b.form = f; return b }
func (b *RequestBuilder) DebugLogLevel(l DebugLogLevel) *RequestBuilder {
	b.debugLogLevel = l
	return b
}

func (b *RequestBuilder) DownloadURL(url string) *RequestBuilder { b.downloadURL = url; return b }

// Build creates a typed [Request[T]] from the builder and message body.
// This is the only point where a generic type parameter is required.
func Build[T any](b *RequestBuilder, msg *T) *Request[T] {
	return &Request[T]{
		Method:         b.method,
		BaseURL:        b.baseURL,
		Type:           b.reqType,
		Bearer:         b.bearer,
		Headers:        b.headers,
		QueryParams:    b.queryParams,
		Endpoints:      b.endpoints,
		Metadata:       b.metadata,
		AppSecret:      b.appSecret,
		SecureRequests: b.secure,
		DownloadURL:    b.downloadURL,
		BodyReader:     b.bodyReader,
		Form:           b.form,
		debugLogLevel:  b.debugLogLevel,
		Message:        msg,
	}
}

// BuildRequest is a package-level helper for callers that prefer the
// function-call style over the method-chain style.
func BuildRequest[T any](b *RequestBuilder, message *T) *Request[T] {
	return Build(b, message)
}

// BuildAnyRequest creates a [Request[any]] with no typed message body.
func BuildAnyRequest(b *RequestBuilder) *Request[any] {
	return Build[any](b, nil)
}
