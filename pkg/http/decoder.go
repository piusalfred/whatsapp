/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
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
)

type JSONDecodeFlag uint8

const (
	// JSONDecodeDefault represents the zero-value configuration where no extra validation behavior is enabled.
	JSONDecodeDefault               JSONDecodeFlag = 0
	JSONDecodeDisallowUnknownFields JSONDecodeFlag = 1 << 0
	JSONDecodeDisallowEmptyResponse JSONDecodeFlag = 1 << 1
	JSONDecodeInspectResponseError  JSONDecodeFlag = 1 << 2
)

const (
	// JSONDecodeStrict combines all validation and inspection behaviors.
	JSONDecodeStrict JSONDecodeFlag = JSONDecodeDisallowUnknownFields |
		JSONDecodeDisallowEmptyResponse |
		JSONDecodeInspectResponseError

		// JSONDecodePermissive only inspects response errors.
	JSONDecodePermissive JSONDecodeFlag = JSONDecodeInspectResponseError
)

// DecodeOptionsStrict returns options that reject unknown JSON fields, require
// a non-empty response body, and parse error responses into structured types.
// Use this when you expect a well-known response schema and want to catch
// unexpected fields early.
func DecodeOptionsStrict() DecodeOptions {
	return DecodeOptions{Flags: JSONDecodeStrict}
}

// DecodeOptionsPermissive returns options that accept unknown fields, allow
// empty response bodies, and parse error responses into structured types.
// This is the safest default for most API calls — it won't reject responses
// due to unexpected fields or missing bodies, but will still surface API
// errors as typed [ResponseError] values.
func DecodeOptionsPermissive() DecodeOptions {
	return DecodeOptions{Flags: JSONDecodePermissive}
}

// DecodeOptionsNoOp returns options that accept unknown fields, allow empty
// bodies, and treat error responses as opaque failures. Use this when you
// don't need error inspection and want the raw JSON decode with no
// validation whatsoever.
func DecodeOptionsNoOp() DecodeOptions {
	return DecodeOptions{}
}

type DecodeOptions struct {
	Flags        JSONDecodeFlag
	MaxBodyBytes int64 // 0 = use package default (DefaultMaxBodyBytes)
}

func MakeDecodeOptions() DecodeOptions {
	return DecodeOptions{
		Flags:        JSONDecodeDefault,
		MaxBodyBytes: DefaultMaxBodyBytes,
	}
}

func (o *DecodeOptions) Has(f JSONDecodeFlag) bool {
	return (o.Flags & f) == f
}

func (o *DecodeOptions) Set(f JSONDecodeFlag) {
	o.Flags |= f
}

func (o *DecodeOptions) Unset(f JSONDecodeFlag) {
	o.Flags &^= f
}

func (o *DecodeOptions) SetMaxBodyBytes(maxBodyBytes int64) {
	o.MaxBodyBytes = maxBodyBytes
}

func DecodeResponseJSON[T any](response *http.Response, v *T, opts DecodeOptions) error {
	if response == nil {
		return ErrNilResponse
	}

	maxBytes := opts.MaxBodyBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodyBytes
	}

	body, err := readAllLimited(response.Body, maxBytes)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusPermanentRedirect

	if !isResponseOk { //nolint:nestif // no need to nest
		if opts.Has(JSONDecodeInspectResponseError) {
			if len(body) == 0 {
				return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
			}

			var errorResponse ResponseError
			if err = json.Unmarshal(body, &errorResponse); err != nil {
				return fmt.Errorf("%w: %w, status code: %d", ErrDecodeErrorResponse, err, response.StatusCode)
			}

			deliverDebugHeaders(response, &errorResponse)

			return &errorResponse
		}

		return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
	}

	if err = decodeJSONBody(body, v, opts); err != nil {
		return err
	}
	deliverDebugHeaders(response, any(v))

	return nil
}

// decodeJSONBody decodes a JSON byte slice into v using opts. It handles
// empty-body checks, nil-target rejection, unknown-field enforcement, and
// JSON unmarshalling. Used by both [DecodeResponseJSON] and [DecodeRequestJSON]
// to avoid duplicated decode logic.
func decodeJSONBody[T any](body []byte, v *T, opts DecodeOptions) error {
	if len(body) == 0 && opts.Has(JSONDecodeDisallowEmptyResponse) {
		return fmt.Errorf("%w: expected non-empty body", ErrEmptyResponseBody)
	}
	if len(body) == 0 {
		return nil
	}
	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	if opts.Has(JSONDecodeDisallowUnknownFields) {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, err)
	}

	return nil
}

// deliverDebugHeaders reads debug headers from the response and delivers them
// to target if it implements [DebugHeadersCapturer].
func deliverDebugHeaders(resp *http.Response, target any) {
	if capturer, ok := target.(DebugHeadersCapturer); ok {
		capturer.OnDebugHeaders(debugHeadersFromResponse(resp))
	}
}

func DecodeRequestJSON[T any](request *http.Request, v *T, opts DecodeOptions) error {
	if request == nil {
		return ErrNilRequest
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	return decodeJSONBody(body, v, opts)
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
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer func() {
			response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}()

		if err = fn(ctx, bytes.NewReader(responseBody)); err != nil {
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
		// Close the original response body after reading it to allow HTTP
		// connection reuse. The original body is backed by the TCP connection;
		// failing to close it prevents the transport from returning the
		// connection to the idle pool.
		closeErr := response.Body.Close()
		if err != nil {
			return fmt.Errorf("capture response body: %w", err)
		}
		if closeErr != nil {
			return fmt.Errorf("close original response body: %w", closeErr)
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
