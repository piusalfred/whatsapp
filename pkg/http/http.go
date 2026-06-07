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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultHTTPClientTimeout = 30 * time.Second
	DefaultMaxBodyBytes      = 10 << 20 // 10 MB
	DefaultMaxHeaderBytes    = 1 << 20  // 1 MB
)

type (
	CoreClient[T any] struct {
		http           *http.Client
		reqHook        RequestInterceptorFunc
		resHook        ResponseInterceptorFunc
		sender         Sender[T]
		maxBodyBytes   int64
		maxHeaderBytes int64
	}

	CoreClientOption[T any] interface {
		apply(client *CoreClient[T])
	}

	CoreClientOptionFunc[T any] func(client *CoreClient[T])
)

func (fn CoreClientOptionFunc[T]) apply( //nolint:unused // implements CoreClientOption[T]
	client *CoreClient[T],
) {
	fn(client)
}

func WithCoreClientHTTPClient[T any](httpClient *http.Client) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		if httpClient != nil {
			client.http = httpClient
		}
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
		client.sender = WrapMiddlewares(SenderFunc[T](client.sender.Send), mws)
	})
}

func WithCoreClientMaxBodyBytes[T any](n int64) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		if n > 0 {
			client.maxBodyBytes = n
		}
	})
}

func WithCoreClientMaxHeaderBytes[T any](n int64) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		if n > 0 {
			client.maxHeaderBytes = n
		}
	})
}

func WithCoreClientSender[T any](sender Sender[T]) CoreClientOption[T] {
	return CoreClientOptionFunc[T](func(client *CoreClient[T]) {
		if sender != nil {
			client.sender = sender
		}
	})
}

func NewSender[T any](options ...CoreClientOption[T]) *CoreClient[T] {
	core := &CoreClient[T]{
		http: &http.Client{
			Timeout: DefaultHTTPClientTimeout,
			Transport: &http.Transport{
				MaxResponseHeaderBytes: DefaultMaxHeaderBytes,
			},
		},
		maxBodyBytes:   DefaultMaxBodyBytes,
		maxHeaderBytes: DefaultMaxHeaderBytes,
	}

	core.sender = SenderFunc[T](core.send)

	for _, option := range options {
		if option != nil {
			option.apply(core)
		}
	}

	return core
}

func (core *CoreClient[T]) send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	if err := SendFuncWithInterceptors[T](
		core.http, core.reqHook, core.resHook, core.maxBodyBytes,
	)(ctx, request, decoder); err != nil {
		return err
	}

	return nil
}

// SendFuncWithInterceptors returns a SenderFunc that applies request and response
// interceptors around the actual HTTP call.
//
// Both request and response bodies are snapshot before the interceptor runs and
// restored afterward, so interceptors can read the body freely without affecting the
// HTTP call or downstream decoding.
func SendFuncWithInterceptors[T any](client *http.Client, reqHook RequestInterceptorFunc,
	resHook ResponseInterceptorFunc,
	maxBodyBytes int64,
) SenderFunc[T] {
	return SenderFunc[T](func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
		req, err := RequestWithContext(ctx, request)
		if err != nil {
			return err
		}

		if err = applyRequestInterceptor(ctx, req, reqHook, maxBodyBytes); err != nil {
			return err
		}

		response, err := client.Do(req) //nolint:bodyclose // body closed
		if err != nil {
			return fmt.Errorf("send request: %w", err)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)

		if err = applyResponseInterceptor(ctx, response, resHook, maxBodyBytes); err != nil {
			return err
		}

		if err = decoder.Decode(ctx, response); err != nil {
			return fmt.Errorf("core send: decode: %w", err)
		}

		return nil
	})
}

func readAllLimited(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w: exceeded %d bytes", ErrBodyTooLarge, limit)
	}
	return data, nil
}

func applyRequestInterceptor(ctx context.Context, req *http.Request,
	hook RequestInterceptorFunc, maxBodyBytes int64,
) error {
	if hook == nil {
		return nil
	}

	var reqBodyBytes []byte
	if req.Body != nil {
		var err error
		reqBodyBytes, err = readAllLimited(req.Body, maxBodyBytes)
		if err != nil {
			return fmt.Errorf("read request body for interceptor: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
	}

	if err := hook(ctx, req); err != nil {
		return err
	}

	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
	}

	return nil
}

func applyResponseInterceptor(
	ctx context.Context,
	resp *http.Response,
	hook ResponseInterceptorFunc,
	maxBodyBytes int64,
) error {
	if hook == nil {
		return nil
	}

	bodyBytes, errRead := readAllLimited(resp.Body, maxBodyBytes)
	if errRead != nil && !errors.Is(errRead, io.EOF) {
		return fmt.Errorf("read response body: %w", errRead)
	}

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err := hook.InterceptResponse(ctx, resp); err != nil {
		return err
	}

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return nil
}

func (core *CoreClient[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	if err := core.sender.Send(ctx, request, decoder); err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
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

func WrapMiddlewares[T any](doFunc SenderFunc[T], middlewares []Middleware[T]) SenderFunc[T] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] != nil {
			doFunc = middlewares[i](doFunc)
		}
	}

	return doFunc
}
