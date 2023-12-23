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

import "net/http"

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
	v interface{},
) error {
	return f(response)
}

// DecodeResponse calls f(ctx, response, v).
func (f ResponseDecoderFunc) DecodeResponse(response *http.Response,
	v interface{},
) error {
	return f(response, v)
}
