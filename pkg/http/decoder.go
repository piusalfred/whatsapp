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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DecodeOptions struct {
	DisallowUnknownFields bool
	DisallowEmptyResponse bool
	InspectResponseError  bool
	MaxBodyBytes          int64 // 0 = use package default (DefaultMaxBodyBytes)
}

// DecodeOptionsStrict returns options that reject unknown JSON fields, require
// a non-empty response body, and parse error responses into structured types.
// Use this when you expect a well-known response schema and want to catch
// unexpected fields early.
func DecodeOptionsStrict() DecodeOptions {
	return DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	}
}

// DecodeOptionsPermissive returns options that accept unknown fields, allow
// empty response bodies, and parse error responses into structured types.
// This is the safest default for most API calls — it won't reject responses
// due to unexpected fields or missing bodies, but will still surface API
// errors as typed [ResponseError] values.
func DecodeOptionsPermissive() DecodeOptions {
	return DecodeOptions{
		InspectResponseError: true,
	}
}

// DecodeOptionsNoOp returns options that accept unknown fields, allow empty
// bodies, and treat error responses as opaque failures. Use this when you
// don't need error inspection and want the raw JSON decode with no
// validation whatsoever.
func DecodeOptionsNoOp() DecodeOptions {
	return DecodeOptions{}
}

// SetMaxBodyBytesSize sets the maximum body size in bytes for decoding.
// When 0 or unset, [DefaultMaxBodyBytes] is used. Returns the receiver
// for chaining at the call site.
func (opts *DecodeOptions) SetMaxBodyBytesSize(size int64) *DecodeOptions {
	opts.MaxBodyBytes = size
	return opts
}

func effectiveMaxBodyBytes(opts DecodeOptions) int64 {
	if opts.MaxBodyBytes > 0 {
		return opts.MaxBodyBytes
	}
	return DefaultMaxBodyBytes
}

func DecodeResponseJSON[T any](response *http.Response, v *T, opts DecodeOptions) error {
	if response == nil {
		return ErrNilResponse
	}

	limit := effectiveMaxBodyBytes(opts)
	responseBody, err := readAllLimited(response.Body, limit)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer func() {
		response.Body = io.NopCloser(bytes.NewReader(responseBody))
	}()

	if len(responseBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty response body", ErrEmptyResponseBody)
	}

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices

	if !isResponseOk { //nolint:nestif // no need to nest
		if opts.InspectResponseError {
			if len(responseBody) == 0 {
				return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
			}

			var errorResponse ResponseError
			if err = json.Unmarshal(responseBody, &errorResponse); err != nil {
				return fmt.Errorf("%w: %w, status code: %d", ErrDecodeErrorResponse, err, response.StatusCode)
			}

			return &errorResponse
		}

		return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
	}

	//	At this point, we know the response is 2xx
	//	If the body is empty here, that means empty bodies are allowed, so we return early.
	//	If the body is not empty but `v` is nil, we return an error as there is no target to decode into.
	//	Otherwise, we proceed with decoding the body into `v` if the body is not empty and `v` is provided.

	if len(responseBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err = decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, err)
	}

	return nil
}

func DecodeRequestJSON[T any](request *http.Request, v *T, opts DecodeOptions) error {
	if request == nil {
		return ErrNilRequest
	}

	limit := effectiveMaxBodyBytes(opts)
	requestBody, err := readAllLimited(request.Body, limit)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}()

	if len(requestBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty request body", ErrEmptyResponseBody)
	}

	//	At this point, we know:
	//	- If the body is empty, it's allowed based on `DisallowEmptyResponse`.
	//	- If the body is not empty and `v == nil`, return an error as there's no target to decode into.
	//	- Otherwise, proceed with decoding into `v` if the body is not empty and `v` is provided.

	if len(requestBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(requestBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if decodeErr := decoder.Decode(v); decodeErr != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, decodeErr)
	}

	return nil
}

type (
	ResponseDecoderFunc func(ctx context.Context, response *http.Response) error

	ResponseDecoder interface {
		Decode(ctx context.Context, response *http.Response) error
	}

	ResponseBodyReaderFunc func(ctx context.Context, reader io.Reader) error
)

func (decoder ResponseDecoderFunc) Decode(ctx context.Context, response *http.Response) error {
	return decoder(ctx, response)
}

func ResponseDecoderJSON[T any](v *T, options DecodeOptions) ResponseDecoderFunc {
	fn := ResponseDecoderFunc(func(_ context.Context, response *http.Response) error {
		if err := DecodeResponseJSON(response, v, options); err != nil {
			return fmt.Errorf("decode json: %w", err)
		}

		return nil
	})

	return fn
}

func BodyReaderResponseDecoder(fn ResponseBodyReaderFunc) ResponseDecoderFunc {
	return func(ctx context.Context, response *http.Response) error {
		if err := fn(ctx, response.Body); err != nil {
			return err
		}

		return nil
	}
}

// ResponseCapturer wraps a ResponseDecoder and retains a copy of the response
// body, status code, and headers so callers can inspect the raw HTTP response
// after decoding completes.
type ResponseCapturer struct {
	inner      ResponseDecoder
	Body       []byte
	StatusCode int
	Header     http.Header
}

// NewResponseCapturer returns a ResponseCapturer that delegates to inner after
// capturing the response details.
func NewResponseCapturer(inner ResponseDecoder) *ResponseCapturer {
	return &ResponseCapturer{inner: inner}
}

// Decode captures status code, headers, and a copy of the body, then delegates
// to the inner decoder. The response body is restored so the inner decoder sees
// the original content.
func (c *ResponseCapturer) Decode(ctx context.Context, response *http.Response) error {
	c.StatusCode = response.StatusCode
	c.Header = response.Header.Clone()

	if response.Body != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("capture response body: %w", err)
		}
		c.Body = body
		response.Body = io.NopCloser(bytes.NewReader(body))
	}

	if c.inner != nil {
		if err := c.inner.Decode(ctx, response); err != nil {
			return fmt.Errorf("inner decoder: %w", err)
		}
	}

	return nil
}

// Reset clears captured data so the capturer can be reused.
func (c *ResponseCapturer) Reset() {
	c.Body = nil
	c.StatusCode = 0
	c.Header = nil
}
