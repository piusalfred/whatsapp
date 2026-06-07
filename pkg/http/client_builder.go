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
	"net/http"
)

// CoreClientBuilder provides a non-generic, fluent way to configure the
// type-independent parts of a CoreClient. The generic type parameter is only
// needed when the client is built (see BuildSender and BuildAnySender).
type CoreClientBuilder struct {
	httpClient     *http.Client
	reqHook        RequestInterceptorFunc
	resHook        ResponseInterceptorFunc
	maxBodyBytes   int64
	maxHeaderBytes int64
}

// NewCoreClientBuilder starts building a new CoreClient with sensible defaults.
func NewCoreClientBuilder() *CoreClientBuilder {
	core := &CoreClientBuilder{
		httpClient: &http.Client{
			Timeout: DefaultHTTPClientTimeout,
			Transport: &http.Transport{
				MaxResponseHeaderBytes: DefaultMaxHeaderBytes,
			},
		},
		maxBodyBytes:   DefaultMaxBodyBytes,
		maxHeaderBytes: DefaultMaxHeaderBytes,
		reqHook:        nil,
		resHook:        nil,
	}

	return core
}

// WithHTTPClient sets the underlying *http.Client.
func (b *CoreClientBuilder) WithHTTPClient(client *http.Client) *CoreClientBuilder {
	if client != nil {
		b.httpClient = client
	}
	return b
}

// WithRequestInterceptor sets the request interceptor.
func (b *CoreClientBuilder) WithRequestInterceptor(hook RequestInterceptorFunc) *CoreClientBuilder {
	b.reqHook = hook
	return b
}

// WithResponseInterceptor sets the response interceptor.
func (b *CoreClientBuilder) WithResponseInterceptor(hook ResponseInterceptorFunc) *CoreClientBuilder {
	b.resHook = hook
	return b
}

// WithMaxBodyBytes sets the maximum allowed body size in bytes.
func (b *CoreClientBuilder) WithMaxBodyBytes(n int64) *CoreClientBuilder {
	if n > 0 {
		b.maxBodyBytes = n
	}
	return b
}

// WithMaxHeaderBytes sets the maximum allowed response header size in bytes.
func (b *CoreClientBuilder) WithMaxHeaderBytes(n int64) *CoreClientBuilder {
	if n > 0 {
		b.maxHeaderBytes = n
	}
	return b
}

// BuildSender creates a typed CoreClient[T] from the builder configuration.
// Middlewares (which are generic) are supplied at this final step.
func BuildSender[T any](b *CoreClientBuilder, middlewares ...Middleware[T]) *CoreClient[T] {
	opts := []CoreClientOption[T]{
		CoreClientOptionFunc[T](func(client *CoreClient[T]) {
			client.maxBodyBytes = b.maxBodyBytes
			client.maxHeaderBytes = b.maxHeaderBytes
			client.reqHook = b.reqHook
			client.resHook = b.resHook
			if b.httpClient != nil {
				client.http = b.httpClient
			}
		}),
	}

	if len(middlewares) > 0 {
		opts = append(opts, WithCoreClientMiddlewares[T](middlewares...))
	}

	return NewSender[T](opts...)
}

// BuildAnySender creates a CoreClient[any] from the builder configuration.
// This is a convenience wrapper around BuildSender[any].
func BuildAnySender(b *CoreClientBuilder, middlewares ...Middleware[any]) *CoreClient[any] {
	return BuildSender[any](b, middlewares...)
}
