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
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestRequestInterceptorFunc(t *testing.T) {
	t.Parallel()

	intercepted := false
	interceptor := whttp.RequestInterceptorFunc(func(ctx context.Context, req *http.Request) error {
		intercepted = true
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	err := interceptor.InterceptRequest(context.Background(), req)

	test.AssertNoError(t, "InterceptRequest", err)
	if !intercepted {
		t.Error("expected interceptor to be called")
	}
}
func TestResponseInterceptorFunc(t *testing.T) {
	t.Parallel()

	intercepted := false
	interceptor := whttp.ResponseInterceptorFunc(func(ctx context.Context, resp *http.Response) error {
		intercepted = true
		return nil
	})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("test")),
	}

	err := interceptor.InterceptResponse(context.Background(), resp)

	test.AssertNoError(t, "InterceptResponse", err)
	if !intercepted {
		t.Error("expected interceptor to be called")
	}
}
