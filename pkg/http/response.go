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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type (
	Response interface {
		ResponseStatus | ResponseMessage
	}

	ResponseStatus struct {
		Success bool `json:"success,omitempty"`
	}

	ResponseMessage struct {
		Product  string             `json:"messaging_product,omitempty"`
		Contacts []*ResponseContact `json:"contacts,omitempty"`
		Messages []*MessageID       `json:"messages,omitempty"`
	}

	MessageID struct {
		ID string `json:"id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappID string `json:"wa_id"`
	}
)

// DecodeResponseJSON decodes the response body into the given interface.
func DecodeResponseJSON[T Response](response *http.Response, v *T) error {
	if v == nil || response == nil {
		return nil
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed

	if !isResponseOk {
		if len(responseBody) == 0 {
			return fmt.Errorf("%w: status code: %d", ErrRequestFailed, response.StatusCode)
		}

		var errorResponse ResponseError
		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
			return fmt.Errorf("error decoding response error body: %w", err)
		}

		return &errorResponse
	}

	if len(responseBody) != 0 {
		if err := json.Unmarshal(responseBody, v); err != nil {
			return fmt.Errorf("error decoding response body: %w", err)
		}
	}

	return nil
}

type ResponseError struct {
	Code int            `json:"code,omitempty"`
	Err  *werrors.Error `json:"error,omitempty"`
}

// Error returns the error message for ResponseError.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("whatsapp error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

// Unwrap returns the underlying error for ResponseError.
func (e *ResponseError) Unwrap() error {
	return e.Err
}

type (
	// ResponseDecoder decodes the response body into the given interface.
	ResponseDecoder interface {
		DecodeResponse(response *http.Response, v interface{}) error
	}

	// ResponseDecoderFunc is an adapter to allow the use of ordinary functions as
	// response decoders. If f is a function with the appropriate signature,
	// ResponseDecoderFunc(f) is a ResponseDecoder that calls f.
	ResponseDecoderFunc func(response *http.Response, v interface{}) error

	// RawResponseDecoder ...
	RawResponseDecoder func(response *http.Response) error
)

// DecodeResponse calls f(response, v).
func (f RawResponseDecoder) DecodeResponse(response *http.Response,
	_ interface{},
) error {
	return f(response)
}

// DecodeResponse calls f(ctx, response, v).
func (f ResponseDecoderFunc) DecodeResponse(response *http.Response,
	v interface{},
) error {
	return f(response, v)
}
