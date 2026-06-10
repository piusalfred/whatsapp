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

// CoreSenderConfigBuilder provides a non-generic, fluent way to configure a
// CoreSenderConfig. The generic type parameter is only needed when the client
// is built (see BuildSender and BuildAnySender).
type CoreSenderConfigBuilder struct {
	cfg CoreSenderConfig
}

// NewCoreSenderConfigBuilder starts building a new CoreSenderConfig with sensible defaults.
func NewCoreSenderConfigBuilder() *CoreSenderConfigBuilder {
	return &CoreSenderConfigBuilder{
		cfg: CoreSenderConfig{
			httpClient: &http.Client{
				Timeout: DefaultHTTPClientTimeout,
				Transport: &http.Transport{
					MaxResponseHeaderBytes: DefaultMaxHeaderBytes,
				},
			},
			httpClientTimeout: DefaultHTTPClientTimeout,
			maxBodyBytes:      DefaultMaxBodyBytes,
			maxHeaderBytes:    DefaultMaxHeaderBytes,
		},
	}
}

// WithHTTPClient sets the underlying *http.Client.
func (b *CoreSenderConfigBuilder) WithHTTPClient(client *http.Client) *CoreSenderConfigBuilder {
	if client != nil {
		b.cfg.httpClient = client
	}
	return b
}

// WithRequestInterceptor sets the request interceptor.
func (b *CoreSenderConfigBuilder) WithRequestInterceptor(hook RequestInterceptorFunc) *CoreSenderConfigBuilder {
	b.cfg.requestHook = hook
	return b
}

// WithResponseInterceptor sets the response interceptor.
func (b *CoreSenderConfigBuilder) WithResponseInterceptor(hook ResponseInterceptorFunc) *CoreSenderConfigBuilder {
	b.cfg.responseHook = hook
	return b
}

// WithMaxBodyBytes sets the maximum allowed body size in bytes.
func (b *CoreSenderConfigBuilder) WithMaxBodyBytes(n int64) *CoreSenderConfigBuilder {
	if n > 0 {
		b.cfg.maxBodyBytes = n
	}
	return b
}

// WithMaxHeaderBytes sets the maximum allowed response header size in bytes.
func (b *CoreSenderConfigBuilder) WithMaxHeaderBytes(n int64) *CoreSenderConfigBuilder {
	if n > 0 {
		b.cfg.maxHeaderBytes = n
	}
	return b
}

// Build creates a CoreSenderConfig from the builder. This is useful when you
// want to build the config and pass it around before creating a CoreClient.
func (b *CoreSenderConfigBuilder) Build() CoreSenderConfig {
	return b.cfg
}

// BuildSender creates a typed CoreClient[T] from the builder configuration.
// Middlewares (which are generic) are supplied at this final step.
func BuildSender[T any](b *CoreSenderConfigBuilder, middlewares ...Middleware[T]) *CoreClient[T] {
	core := newCoreClientFromConfig[T](b.cfg)
	if len(middlewares) > 0 {
		core.SetMiddlewares(middlewares...)
	}
	return core
}

// BuildAnySender creates a CoreClient[any] from the builder configuration.
// This is a convenience wrapper around BuildSender[any].
func BuildAnySender(b *CoreSenderConfigBuilder, middlewares ...Middleware[any]) *CoreClient[any] {
	return BuildSender[any](b, middlewares...)
}

// newCoreClientFromConfig creates a CoreClient from a CoreSenderConfig.
func newCoreClientFromConfig[T any](cfg CoreSenderConfig) *CoreClient[T] {
	core := &CoreClient[T]{
		http:           cfg.httpClient,
		maxBodyBytes:   cfg.maxBodyBytes,
		maxHeaderBytes: cfg.maxHeaderBytes,
		reqHook:        cfg.requestHook,
		resHook:        cfg.responseHook,
		timeout:        cfg.httpClientTimeout,
	}
	core.sender = SenderFunc[T](core.send)
	return core
}
