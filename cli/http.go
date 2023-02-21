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

package cli

import (
	"io"
	"net/http"
	"time"
)

const (
	JsonOutPutFormat  OutPutFormat = "json"
	TextOutPutFormat  OutPutFormat = "text"
	BytesOutPutFormat OutPutFormat = "bytes"
)

type (
	OutPutFormat string
	// HttpClient is a struct that wraps the *http.Client and hence it can be used to
	// make http requests by the cli, with more fine grained control like logging,
	// middlewares, retries, etc
	HttpClient struct {
		http       *http.Client
		Debug      bool
		Logger     io.Writer
		MaxRetries int
		Jitter     time.Duration
		Timeout    time.Duration
		Out        OutPutFormat
	}
)
