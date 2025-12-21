/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	CoreClient[T any] struct {
		http        *http.Client
		reqHook     RequestInterceptorFunc
		resHook     ResponseInterceptorFunc
		middlewares []Middleware[T]
		sender      Sender[T]
	}

	CoreClientOption[T any] interface {
		apply(client *CoreClient[T])
	}

	CoreClientOptionFunc[T any] func(client *CoreClient[T])

	Options struct {
		HTTPClient   *http.Client
		RequestHook  RequestInterceptorFunc
		ResponseHook ResponseInterceptorFunc
	}
)

func (fn CoreClientOptionFunc[T]) apply(client *CoreClient[T]) {
	fn(client)
}

func (core *CoreClient[T]) SetHTTPClient(httpClient *http.Client) {
	if httpClient != nil {
		core.http = httpClient
	}
}

func (core *CoreClient[T]) SetRequestInterceptor(hook RequestInterceptorFunc) {
	core.reqHook = hook
}

func (core *CoreClient[T]) SetBaseSender(sender Sender[T]) {
	core.sender = sender
}

func (core *CoreClient[T]) SetResponseInterceptor(hook ResponseInterceptorFunc) {
	core.resHook = hook
}

func (core *CoreClient[T]) AppendMiddlewares(mws ...Middleware[T]) {
	core.middlewares = append(core.middlewares, mws...)
}

func (core *CoreClient[T]) PrependMiddlewares(mws ...Middleware[T]) {
	core.middlewares = append(mws, core.middlewares...)
}

func WithCoreClientHTTPClient[T any](httpClient *http.Client) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		client.http = httpClient
	})
}

func WithCoreClientRequestInterceptor[T any](hook RequestInterceptorFunc) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		client.reqHook = hook
	})
}

func WithCoreClientResponseInterceptor[T any](hook ResponseInterceptorFunc) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		client.resHook = hook
	})
}

func WithCoreClientMiddlewares[T any](mws ...Middleware[T]) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		client.middlewares = mws
	})
}

func NewSender[T any](options ...CoreClientOption[T]) *CoreClient[T] {
	core := &CoreClient[T]{
		http: http.DefaultClient,
	}

	core.sender = SenderFunc[T](core.send)

	for _, option := range options {
		if option != nil {
			option.apply(core)
		}
	}

	return core
}

func NewAnySender(options ...CoreClientOption[any]) *CoreClient[any] {
	core := &CoreClient[any]{
		http: http.DefaultClient,
	}

	core.sender = SenderFunc[any](core.send)

	for _, option := range options {
		if option != nil {
			option.apply(core)
		}
	}

	return core
}

func (core *CoreClient[T]) send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	if err := SendFuncWithInterceptors[T](core.http, core.reqHook, core.resHook)(ctx, request, decoder); err != nil {
		return err
	}

	return nil
}

func SendFuncWithInterceptors[T any](client *http.Client, reqHook RequestInterceptorFunc,
	resHook ResponseInterceptorFunc,
) SenderFunc[T] {
	fn := SenderFunc[T](func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
		req, err := RequestWithContext(ctx, request)
		if err != nil {
			return err
		}

		if reqHook != nil {
			if errHook := reqHook(ctx, req); errHook != nil {
				return errHook
			}
		}

		response, err := client.Do(req) //nolint:bodyclose // body closed
		if err != nil {
			return fmt.Errorf("send request: %w", err)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)

		if resHook != nil {
			bodyBytes, errRead := io.ReadAll(response.Body)
			if errRead != nil && !errors.Is(errRead, io.EOF) {
				return fmt.Errorf("read response body: %w", errRead)
			}
			response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if errHook := resHook.InterceptResponse(ctx, response); errHook != nil {
				return errHook
			}
			response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if err = decoder.Decode(ctx, response); err != nil {
			return fmt.Errorf("core send: decode: %w", err)
		}

		return nil
	})

	return fn
}

func (core *CoreClient[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	fn := wrapMiddlewares(core.sender.Send, core.middlewares)

	return fn(ctx, request, decoder)
}

type (
	Sender[T any] interface {
		Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error
	}

	SenderFunc[T any] func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error

	Middleware[T any] func(next SenderFunc[T]) SenderFunc[T]

	AnySender Sender[any]

	AnySenderFunc SenderFunc[any]
)

func (fn SenderFunc[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	return fn(ctx, request, decoder)
}

func (fn AnySenderFunc) Send(ctx context.Context, request *Request[any], decoder ResponseDecoder) error {
	return fn(ctx, request, decoder)
}

func wrapMiddlewares[T any](doFunc SenderFunc[T], middlewares []Middleware[T]) SenderFunc[T] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] != nil {
			doFunc = middlewares[i](doFunc)
		}
	}

	return doFunc
}
