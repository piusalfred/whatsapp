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
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type (
	RequestHook  func(ctx context.Context, request *http.Request) error
	ResponseHook func(ctx context.Context, response *http.Response) error
)

func LogRequestHook(logger *slog.Logger) RequestHook {
	return func(ctx context.Context, request *http.Request) error {
		name := RequestNameFromContext(ctx)
		reader, err := request.GetBody()
		if err != nil {
			return fmt.Errorf("log request hook: %w", err)
		}
		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(reader); err != nil {
			return fmt.Errorf("log request hook: %w", err)
		}

		hb := &strings.Builder{}
		hb.WriteString("[")
		for k, v := range request.Header {
			hb.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		}
		hb.WriteString("]")

		logger.LogAttrs(ctx, slog.LevelDebug, "request", slog.String("name", name),
			slog.String("body", buf.String()), slog.String("headers", hb.String()))

		return nil
	}
}

func LogResponseHook(logger *slog.Logger) ResponseHook {
	return func(ctx context.Context, response *http.Response) error {
		name := RequestNameFromContext(ctx)
		reader := response.Body
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(reader); err != nil {
			return fmt.Errorf("log response hook: %w", err)
		}

		hb := &strings.Builder{}
		hb.WriteString("[")
		for k, v := range response.Header {
			hb.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		}
		hb.WriteString("]")

		logger.LogAttrs(ctx, slog.LevelDebug, "response", slog.String("name", name),
			slog.Int("status", response.StatusCode),
			slog.String("status_text", response.Status),
			slog.String("headers", hb.String()),
			slog.String("body", buf.String()),
		)

		return nil
	}
}
