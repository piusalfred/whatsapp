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

// ErrUnknownRequestType is returned when a request carries an unsupported RequestType.
var ErrUnknownRequestType = errors.New("unknown request type")

const (
	DefaultHTTPClientTimeout = 30 * time.Second
	DefaultMaxBodyBytes      = 10 << 20 // 10 MB
	DefaultMaxHeaderBytes    = 1 << 20  // 1 MB
)

type (
	// CoreClient is a type-safe HTTP client parameterized by the request body type.
	// It is immutable after construction via [NewCoreClient] and safe for concurrent use.
	// Fields are unexported; configure behavior through [CoreSenderOption] values passed
	// to [NewCoreClient].
	CoreClient[T any] struct {
		http           *http.Client
		reqHook        RequestInterceptorFunc
		resHook        ResponseInterceptorFunc
		maxBodyBytes   int64
		maxHeaderBytes int64
		timeout        time.Duration
		sender         Sender[T]
	}

	// CoreSenderConfig bundles HTTP-level settings (client, hooks, limits) that are
	// independent of the generic type parameter. [CoreSenderOption] implementations
	// mutate this config, and [NewCoreClient] materializes it into a typed [CoreClient].
	CoreSenderConfig struct {
		httpClient        *http.Client
		requestHook       RequestInterceptorFunc
		responseHook      ResponseInterceptorFunc
		httpClientTimeout time.Duration
		maxBodyBytes      int64
		maxHeaderBytes    int64
	}
)

// NewCoreClient creates a CoreClient[T] with the provided options. The returned
// client is safe for concurrent use and should not be mutated after construction.
func NewCoreClient[T any](options ...CoreSenderOption) *CoreClient[T] {
	config := defaultCoreSenderConfig()
	for _, option := range options {
		if option != nil {
			option.apply(&config)
		}
	}
	return newCoreClientFromConfig[T](config)
}

// defaultCoreSenderConfig returns a [CoreSenderConfig] with safe defaults:
// 30-second timeout, 10 MB body limit, and 1 MB header limit.
func defaultCoreSenderConfig() CoreSenderConfig {
	return CoreSenderConfig{
		httpClient: &http.Client{
			Timeout:   DefaultHTTPClientTimeout,
			Transport: &http.Transport{MaxResponseHeaderBytes: DefaultMaxHeaderBytes},
		},
		httpClientTimeout: DefaultHTTPClientTimeout,
		maxBodyBytes:      DefaultMaxBodyBytes,
		maxHeaderBytes:    DefaultMaxHeaderBytes,
	}
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
// Both request and response bodies are snapshotted before the interceptor runs and
// restored afterward, so interceptors can read the body freely without affecting the
// HTTP call or downstream decoding.
func SendFuncWithInterceptors[T any](client *http.Client, reqHook RequestInterceptorFunc,
	resHook ResponseInterceptorFunc,
	maxBodyBytes int64,
) SenderFunc[T] {
	return SenderFunc[T](func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
		return sendWithInterceptors(ctx, client, reqHook, resHook, maxBodyBytes, request, decoder)
	})
}

func sendWithInterceptors[T any](
	ctx context.Context,
	client *http.Client,
	reqHook RequestInterceptorFunc,
	resHook ResponseInterceptorFunc,
	maxBodyBytes int64,
	request *Request[T],
	decoder ResponseDecoder,
) error {
	req, err := RequestWithContext(ctx, request)
	if err != nil {
		return err
	}

	if err = applyRequestInterceptor(ctx, req, reqHook, maxBodyBytes); err != nil {
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		if response != nil {
			_ = response.Body.Close()
		}
		return fmt.Errorf("send request: %w", err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if err = applyResponseInterceptor(ctx, response, resHook, maxBodyBytes); err != nil {
		return err
	}

	if err = decoder.Decode(ctx, response); err != nil {
		return fmt.Errorf("core send: decode: %w", err)
	}

	return nil
}

// readAllLimited reads up to limit bytes from r.
//
// The LimitReader is set to limit+1 to distinguish between "exactly at the limit"
// (allowed) and "over the limit" (rejected). This prevents rejecting a valid
// response whose body is exactly at the boundary.
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
		// Close the original body after reading it. This is necessary because
		// the body is a ReadCloser backed by the HTTP transport; failing to
		// close it prevents connection reuse.
		closeErr := req.Body.Close()
		if err != nil {
			return fmt.Errorf("read request body for interceptor: %w", err)
		}
		if closeErr != nil {
			return fmt.Errorf("close original request body: %w", closeErr)
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
	// Close the original response body after reading it. Failing to close the
	// underlying TCP-reader body prevents the HTTP transport from reusing the
	// connection, causing connection pool exhaustion under load.
	closeErr := resp.Body.Close()
	if errRead != nil && !errors.Is(errRead, io.EOF) {
		return fmt.Errorf("read response body: %w", errRead)
	}
	if closeErr != nil {
		return fmt.Errorf("close original response body: %w", closeErr)
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
	// Sender abstracts HTTP request execution. [CoreClient] implements it;
	// callers can provide their own implementation to mock HTTP calls in tests.
	Sender[T any] interface {
		Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error
	}

	// SenderFunc is a function adapter that implements [Sender] by calling itself,
	// analogous to [http.HandlerFunc].
	SenderFunc[T any] func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error

	// Middleware wraps a [SenderFunc] to add cross-cutting behavior (logging, tracing,
	// metrics, etc.). When composed via [WrapMiddlewares], middlewares are applied
	// inside-out so that middlewares[0] runs outermost.
	Middleware[T any] func(next SenderFunc[T]) SenderFunc[T]
)

func (fn SenderFunc[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	return fn(ctx, request, decoder)
}

// WrapMiddlewares composes middlewares into a single SenderFunc.
//
// Middlewares are applied in slice order: middlewares[0] is the outermost
// wrapper (runs first on the way in, last on the way out). The loop iterates
// in reverse to build the chain inside-out — each middleware wraps the
// accumulated result from the previous iteration.
func WrapMiddlewares[T any](doFunc SenderFunc[T], middlewares []Middleware[T]) SenderFunc[T] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] != nil {
			doFunc = middlewares[i](doFunc)
		}
	}

	return doFunc
}

// CoreSenderOption is a functional option for [NewCoreClient]. It is implemented
// by [CoreSenderOptionFunc] and the "WithSender*" helpers returned by
// [WithSenderHTTPClient], [WithSenderRequestInterceptor], etc.
type CoreSenderOption interface {
	apply(client *CoreSenderConfig)
}

// CoreSenderOptionFunc configures the underlying [BaseClient] HTTP transport.
type CoreSenderOptionFunc func(*CoreSenderConfig)

func (c CoreSenderOptionFunc) apply(client *CoreSenderConfig) {
	c(client)
}

// WithSenderHTTPClient replaces the default [http.Client] used by the sender.
// A nil client is ignored.
func WithSenderHTTPClient(hc *http.Client) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if hc != nil {
			cfg.httpClient = hc
		}
	}
}

// WithSenderRequestInterceptor registers a hook that inspects or mutates every
// outgoing [http.Request] before it is transmitted. A nil hook is ignored.
func WithSenderRequestInterceptor(hook RequestInterceptorFunc) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if hook != nil {
			cfg.requestHook = hook
		}
	}
}

// WithSenderResponseInterceptor registers a hook that inspects or mutates every
// incoming [http.Response] before it is decoded. A nil hook is ignored.
func WithSenderResponseInterceptor(hook ResponseInterceptorFunc) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if hook != nil {
			cfg.responseHook = hook
		}
	}
}

// WithSenderMaxBodyBytes sets the maximum allowable body size for request/response
// interceptors. Values less than or equal to zero are ignored.
func WithSenderMaxBodyBytes(n int64) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if n > 0 {
			cfg.maxBodyBytes = n
		}
	}
}

// WithSenderMaxHeaderBytes sets the maximum response header size. Values less than or
// equal to zero are ignored.
func WithSenderMaxHeaderBytes(n int64) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if n > 0 {
			cfg.maxHeaderBytes = n
		}
	}
}

// WithSenderTimeout sets the HTTP client timeout. Values less than or equal to zero
// are ignored.
func WithSenderTimeout(timeout time.Duration) CoreSenderOptionFunc {
	return func(cfg *CoreSenderConfig) {
		if timeout > 0 {
			cfg.httpClientTimeout = timeout
		}
	}
}
