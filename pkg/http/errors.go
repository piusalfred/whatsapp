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
	"fmt"
	"strings"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type ResponseError struct {
	Code int            `json:"code,omitempty"`
	Err  *werrors.Error `json:"error,omitempty"`
}

func (e *ResponseError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("whatsapp message error: http code: %d, <no details>", e.Code)
	}
	return fmt.Sprintf("whatsapp message error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

func (e *ResponseError) Unwrap() error {
	return e.Err
}

const (
	ErrNilRequest          = httpError("nil request provided")
	ErrNilResponse         = httpError("nil response provided")
	ErrEmptyResponseBody   = httpError("empty response body")
	ErrNilTarget           = httpError("nil value passed for decoding target")
	ErrRequestFailure      = httpError("request failed")
	ErrDecodeResponseBody  = httpError("failed to decode response body")
	ErrDecodeErrorResponse = httpError("failed to decode error response")
	ErrBodyTooLarge        = httpError("body exceeds maximum allowed size")
	ErrMultipleBodySources = httpError(
		"multiple body sources provided: only one of Message, Form, or BodyReader is allowed",
	)
)

type httpError string

func (e httpError) Error() string {
	return string(e)
}

type Paging struct {
	Cursors  *Cursors `json:"cursors,omitempty"`
	Previous string   `json:"previous,omitempty"`
	Next     string   `json:"next,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}
