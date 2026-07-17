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
	"context"
	"fmt"
)

// BaseClient is a generic building block for domain-specific HTTP clients. It holds
// an unexported sender that domain code calls via the [BaseClient.Send] method.
// Tests inject mocks via [BaseClient.SetSender].
//
// [SetMiddlewares] and [SetSender] must complete before any concurrent calls to
// [BaseClient.Send]. These methods are not goroutine-safe with respect to Send.
//
// Example:
//
//	type MyClient struct {
//	    http.BaseClient[MyRequest]
//	}
//
//	client := MyClient{BaseClient: http.NewBaseClient[MyRequest](opts...)}
//	client.Send(ctx, req, decoder) // direct dispatch
type BaseClient[T any] struct {
	sender Sender[T]
}

// NewBaseClient creates a BaseClient[T] backed by a [CoreClient] configured with opts.
// The returned BaseClient is safe for concurrent use as long as [SetMiddlewares] and
// [SetSender] are not called concurrently with [Send].
func NewBaseClient[T any](opts ...CoreSenderOption) *BaseClient[T] {
	return &BaseClient[T]{sender: NewCoreClient[T](opts...)}
}

// Send dispatches the request through the underlying [Sender].
func (bc *BaseClient[T]) Send(ctx context.Context, req *Request[T], decoder ResponseDecoder) error {
	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

// SetMiddlewares wraps the sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
//
// Must be called before any concurrent [Send] calls. Not goroutine-safe
// with respect to Send.
func (bc *BaseClient[T]) SetMiddlewares(mws ...Middleware[T]) {
	bc.sender = WrapMiddlewareSender(bc.sender, mws...)
}

// Sender returns the underlying [Sender], useful for tests that need to
// inspect or wrap the current sender before calling [SetSender].
func (bc *BaseClient[T]) Sender() Sender[T] {
	return bc.sender
}

// SetSender replaces the sender, useful for injecting mocks in tests.
//
// Must be called before any concurrent [Send] calls. Not goroutine-safe
// with respect to Send.
func (bc *BaseClient[T]) SetSender(sender Sender[T]) {
	bc.sender = sender
}

// CloseIdleConnections closes idle connections in the underlying HTTP transport,
// if the sender supports it. Call during graceful shutdown to drain the
// connection pool and avoid leaking sockets.
func (bc *BaseClient[T]) CloseIdleConnections() {
	type closeIdler interface{ CloseIdleConnections() }
	if c, ok := bc.sender.(closeIdler); ok {
		c.CloseIdleConnections()
	}
}

// WrapMiddlewareSender wraps a [Sender] with the provided middlewares and returns
// a new [Sender]. Middlewares are applied in the order provided.
func WrapMiddlewareSender[T any](sender Sender[T], mws ...Middleware[T]) Sender[T] {
	return WrapMiddlewares(
		SenderFunc[T](func(ctx context.Context, req *Request[T], decoder ResponseDecoder) error {
			return sender.Send(ctx, req, decoder)
		}),
		mws,
	)
}
