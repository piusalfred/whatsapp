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
	"log/slog"
	"net/url"
)

func (request *Request) LogValue() slog.Value {
	if request == nil {
		return slog.StringValue("nil")
	}
	var reqURL string
	if request.Context != nil {
		reqURL, _ = url.JoinPath(request.Context.BaseURL, request.Context.Endpoints...)
	}

	var metadataAttr []any

	for key, value := range request.Metadata {
		metadataAttr = append(metadataAttr, slog.String(key, value))
	}

	var headersAttr []any

	for key, value := range request.Headers {
		headersAttr = append(headersAttr, slog.String(key, value))
	}

	var queryAttr []any

	for key, value := range request.Query {
		queryAttr = append(queryAttr, slog.String(key, value))
	}

	value := slog.GroupValue(
		slog.String("type", request.Context.RequestType.String()),
		slog.String("method", request.Method),
		slog.String("url", reqURL),
		slog.Group("metadata", metadataAttr...),
		slog.Group("headers", headersAttr...),
		slog.Group("query", queryAttr...),
	)

	return value
}
