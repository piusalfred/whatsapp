//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http_test

import (
	"strings"
	"testing"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestResponseError(t *testing.T) {
	t.Parallel()

	t.Run("Error method", func(t *testing.T) {
		t.Parallel()

		respErr := &whttp.ResponseError{
			Code: 500,
			Err: &werrors.Error{
				Message: "Test error message",
				Type:    "TestException",
				Code:    12345,
			},
		}

		errMsg := respErr.Error()
		if !strings.Contains(errMsg, "500") {
			t.Errorf("expected error message to contain status code 500, got %s", errMsg)
		}
		if !strings.Contains(strings.ToLower(errMsg), "test error message") {
			t.Errorf("expected error message to contain 'test error message', got %s", errMsg)
		}
	})

	t.Run("Unwrap method", func(t *testing.T) {
		t.Parallel()

		innerErr := &werrors.Error{
			Message: "Inner error",
			Type:    "InnerException",
			Code:    999,
		}

		respErr := &whttp.ResponseError{
			Code: 400,
			Err:  innerErr,
		}

		unwrapped := respErr.Unwrap()
		if unwrapped != innerErr {
			t.Errorf("expected unwrapped error to be innerErr, got %v", unwrapped)
		}
	})
}

func TestHTTPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNilResponse",
			err:  whttp.ErrNilResponse,
			want: "nil response provided",
		},
		{
			name: "ErrEmptyResponseBody",
			err:  whttp.ErrEmptyResponseBody,
			want: "empty response body",
		},
		{
			name: "ErrNilTarget",
			err:  whttp.ErrNilTarget,
			want: "nil value passed for decoding target",
		},
		{
			name: "ErrRequestFailure",
			err:  whttp.ErrRequestFailure,
			want: "request failed",
		},
		{
			name: "ErrDecodeResponseBody",
			err:  whttp.ErrDecodeResponseBody,
			want: "failed to decode response body",
		},
		{
			name: "ErrDecodeErrorResponse",
			err:  whttp.ErrDecodeErrorResponse,
			want: "failed to decode error response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.err.Error() != tt.want {
				t.Errorf("expected error message %q, got %q", tt.want, tt.err.Error())
			}
		})
	}
}
