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
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type DebugFunc func(io.Writer) Hook

// DebugHook is a hook that prints the request and response to the writer, Internally it uses
// httputil.DumpRequestOut and httputil.DumpResponse to dump the request and response. It also
// retrieves the request name from the context using RequestNameFromContext and print it with
// the request and response.
func DebugHook(writer io.Writer) Hook {
	return func(ctx context.Context, req *http.Request, resp *http.Response) {
		buff := &strings.Builder{}
		name := strings.ToUpper(RequestNameFromContext(ctx))
		buff.WriteString(name)
		buff.WriteString("\n\n")

		dumpRequest(buff, req)
		dumpResponse(buff, resp)
		//
		//if writer == nil {
		//	_, _ = os.Stdout.Write(buff.Bytes())
		//} else {
		//	_, _ = writer.Write(buff.Bytes())
		//}
	}
}

func dumpRequest(buff *strings.Builder, req *http.Request) {
	if req != nil {
		b, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			buff.Write(b)
			buff.WriteString("\n\n")
		}
	}
}

func dumpResponse(buff *strings.Builder, resp *http.Response) {
	if resp != nil {
		b, err := httputil.DumpResponse(resp, true)
		if err == nil {
			buff.Write(b)
			buff.WriteString("\n\n")
		}
	}
}
