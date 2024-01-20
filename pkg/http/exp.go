/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
)

// DoWithDecoder sends a http request to the server and returns the response, It accepts a context,
// a request, a pointer to a variable to decode the response into and a response decoder.
func (client *BaseClient) DoWithDecoder(ctx context.Context, r *Request, decoder ResponseDecoder, v any) error {
	request, err := client.prepareRequest(ctx, r)
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}

	response, err := client.http.Do(request) //nolint:bodyclose
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// send error to error channel
			client.errorChannel <- fmt.Errorf("closing response body: %w", err)
		}
	}(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading response body: %w", err)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err = client.runResponseHooks(ctx, response); err != nil {
		return fmt.Errorf("response hooks: %w", err)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	rawResponseDecoder, ok := decoder.(RawResponseDecoder)
	if ok {
		if err := rawResponseDecoder(response); err != nil {
			return fmt.Errorf("raw response decoder: %w", err)
		}

		return nil
	}

	if err := decoder.DecodeResponse(response, v); err != nil {
		return fmt.Errorf("response decoder: %w", err)
	}

	return nil
}

// Do send a http request to the server and returns the response, It accepts a context,
// a request and a pointer to a variable to decode the response into.
func (client *BaseClient) Do(ctx context.Context, r *Request, v any) error {
	request, err := client.prepareRequest(ctx, r)
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}

	response, err := client.http.Do(request)
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// send error to error channel
			client.errorChannel <- fmt.Errorf("closing response body: %w", err)
		}
	}(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading response body: %w", err)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err = client.runResponseHooks(ctx, response); err != nil {
		return fmt.Errorf("response hooks: %w", err)
	}
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return decodeResponseJSON(response, v)
}
