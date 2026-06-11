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

// Package http provides a type-safe HTTP client with middleware chains, request/response
// interceptors, and pluggable body encoding/decoding.
//
// # Quick Start
//
//	client := http.NewCoreClient[MyMessage](
//	    http.WithSenderHTTPClient(&http.Client{Timeout: 10 * time.Second}),
//	    http.WithSenderRequestInterceptor(logRequest),
//	)
//
//	req := http.MakeRequest(http.MethodPost, "https://api.example.com",
//	    http.WithRequestBearer[MyMessage]("token"),
//	    http.WithRequestMessage[MyMessage](&MyMessage{Text: "hello"}),
//	)
//
//	var resp MyResponse
//	err := client.Send(ctx, req, http.ResponseDecoderJSON(&resp, http.StrictDecode()))
//
// # Architecture
//
// The package is organized around three layers:
//
//  1. Request  — builds [net/http.Request] from typed domain payloads
//  2. Sender   — executes the request through middleware chains and interceptors
//  3. Decoder  — deserializes [net/http.Response] into typed domain values
//
// Middlewares wrap the sender's Send method (analogous to http.Handler middleware).
// Interceptors are lighter hooks that snapshot and restore the body so they can
// inspect request/response data without consuming it.
package http
