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
//	builder := http.NewRequestBuilder(http.MethodPost, "https://api.example.com").
//	    Bearer("token")
//	req := http.Build(builder, &MyMessage{Text: "hello"})
//
//	var resp MyResponse
//	err := client.Send(ctx, req, http.ResponseDecoderJSON(&resp, http.DecodeOptions{Flags: http.JSONDecodeStrict}))
//
// # Architecture
//
// The package is organized around three layers:
//
//  1. Request    — builds [net/http.Request] from typed domain payloads
//  2. Sender     — executes the request through middleware chains and interceptors
//  3. ResponseDecoder    — deserializes [net/http.Response] into typed domain values
//  4. BaseClient — embeddable building block that domain packages compose into
//     their own client types, exposing a [Sender] for direct dispatch
//
// Middlewares wrap the sender's Send method (analogous to http.Handler middleware).
// Interceptors are lighter hooks that snapshot and restore the body so they can
// inspect request/response data without consuming it.
//
// # Request Construction: RequestBuilder vs MakeRequest
//
// Two APIs exist for building requests:
//
//	RequestBuilder (preferred for most callers):
//	    req := NewRequestBuilder(method, baseURL).Bearer("t").Type(rt).Endpoints("v1", "x")
//	    typedReq := Build(builder, &body)
//
//	MakeRequest (preferred for simple requests with few options):
//	    req := MakeRequest[T](method, baseURL, WithRequestBearer("t"), WithRequestEndpoints("v1"))
//
// Prefer [RequestBuilder] when building requests with many options — the chained
// API avoids repeating the type parameter [T] on every option. Prefer [MakeRequest]
// for requests with a single option where the functional-options style reads
// more naturally.
//
// # Middlewares vs Interceptors
//
// Middlewares and interceptors are two distinct cross-cutting concern mechanisms
// that operate at different layers:
//
//	Middlewares ([Middleware]) wrap the [SenderFunc] — they see the typed [Request]
//	and [ResponseDecoder] before the HTTP call, and the typed response value after
//	decoding. Analogous to HTTP handler middleware: they wrap the entire
//	send-decode pipeline.
//
//	Interceptors ([RequestInterceptor], [ResponseInterceptor]) wrap the raw HTTP
//	call (the [net/http.Client.Do]). They see the raw [*net/http.Request] and
//	[*net/http.Response] and work by snapshotting and restoring the request/response
//	body so the hook can inspect data without consuming it.
//
// Middlewares are applied in reverse slice order via [WrapMiddlewares] so that
// middlewares[0] runs outermost (the standard onion pattern). Interceptors are
// called linearly: request interceptors before the HTTP call, response interceptors
// after.
//
// # Option Patterns
//
// The package uses several option patterns, each suited to a different layer:
//
//	[CoreSenderOption] — interface-based (unexported apply method). Used by
//	[NewCoreClient] for sender-level configuration. Only [CoreSenderOptionFunc]
//	satisfies this interface; external types cannot implement it.
//
//	[RequestOption[T]] — generic functional option (a function). Used by
//	[MakeRequest] for request-level configuration.
//
//	[DecodeOptions] — plain struct. Used by response decoders. Pre-built
//	instances available via [DecodeOptionsStrict], [DecodeOptionsPermissive],
//	and [DecodeOptionsNoOp].
//
//	[RequestBuilder] — non-generic fluent builder that becomes generic at the
//	final [Build] call. Preferred for constructing requests with many options.
//
//	[CoreSenderConfigBuilder] — non-generic fluent builder for sender
//	configuration. Alternative to the functional-options path; call
//	[CoreSenderConfigBuilder.BuildSender] to produce a [Sender].
//
// # Context Metadata
//
// [InjectMessageMetadata] and [RetrieveMessageMetadata] store and retrieve
// [types.Metadata] through the Go context. The context key is an unexported
// typed string, preventing external packages from constructing an equivalent key
// and accidentally colliding with this value.
//
// Metadata is injected automatically when calling [RequestWithContext] (and
// therefore flows through [BaseClient.Send]). Server-side handlers can call
// [RetrieveMessageMetadata] to extract it:
//
//	ctx := http.InjectMessageMetadata(ctx, metadata)
//	// ... later, in a handler:
//	md := http.RetrieveMessageMetadata(ctx)
//
// # Response Capturer & ResponseDecoder2
//
// [ResponseCapturer] wraps a [ResponseDecoder] and transparently captures the raw
// HTTP response status code, headers, and body before delegating to the inner
// decoder. It is also the canonical implementation of the [ResponseDecoder2]
// interface:
//
//	capturer := http.NewResponseCapturer(http.ResponseDecoderJSON(&resp, opts))
//	err := client.Send(ctx, req, capturer)
//	// Individual accessors:
//	status := capturer.StatusCode()
//	headers := capturer.Header()
//	rawBody := capturer.Body()
//	// Or as a single snapshot:
//	dump := capturer.DumpResponse()
//	fmt.Println(dump.StatusCode, dump.Header, string(dump.Body))
//
// Use the individual getters ([ResponseCapturer.StatusCode], [ResponseCapturer.Header],
// [ResponseCapturer.Body]) when you need specific fields. Use
// [ResponseCapturer.DumpResponse] when you need a complete [ResponseDump].
//
// [ResponseDecoder2] is the interface form — depend on it in function signatures
// when you want to accept any decoder that can dump response metadata:
//
//	func MySender(decoder http.ResponseDecoder2) { ... }
//
// [ResponseCapturer.Reset] clears all captured data so the capturer can be reused,
// e.g., from a sync.Pool.
//
// # Debug Headers
//
// The WhatsApp API returns debug headers ([DebugHeaders]) on responses. These are
// delivered automatically to any decoded response type that implements the
// [DebugHeadersCapturer] interface:
//
//	type MyResponse struct {
//	    Data string `json:"data"`
//	}
//	func (r *MyResponse) OnDebugHeaders(h http.DebugHeaders) {
//	    // h contains Facebook-Api-Version, X-Fb-Trace-Id, X-Fb-Rev, X-Fb-Debug
//	}
//
// [*ResponseError] already implements [DebugHeadersCapturer], so errors
// automatically carry debug headers. Custom types must implement the interface
// explicitly to receive them on success responses.
//
// The debug header delivery happens inside the JSON decoder via a runtime type
// assertion — there is no compile-time registration step. Simply implement the
// interface on your response type.
//
// # Debug Logging
//
// Requests can be tagged with a debug log level that appends a ?debug=<level>
// query parameter to the request URL (see [Request.URL]). Set it via:
//
//	// functional option:
//	req := http.MakeRequest[MyMsg](method, url, http.WithRequestDebugLogLevel(http.DebugLogLevelAll))
//
//	// builder:
//	req := http.NewRequestBuilder(method, url).DebugLogLevel(http.DebugLogLevelAll)
//
// Available levels: [DebugLogLevelInfo], [DebugLogLevelAll], [DebugLogLevelWarning],
// [DebugLogLevelNone].
//
// [ParseDebugLogLevel] provides case-insensitive parsing from a string. Unknown
// values silently fall back to [DebugLogLevelNone] — use [ValidDebugLogLevel] to
// validate before parsing if you need to reject unknown inputs.
//
// # Error Handling
//
// All sentinel errors (e.g., [ErrNilRequest], [ErrBodyTooLarge]) use an unexported
// type, so callers must use [errors.Is] to check for them — type assertions or
// type switches will not work.
//
// [ResponseError] is the structured error type for API error responses. It
// implements:
//   - error (formatted string combining the HTTP code and inner error)
//   - [errors.Unwrap] (returns the inner [werrors.Error] for [errors.As] extraction)
//   - [DebugHeadersCapturer] (automatically populated by the decoder)
//
// # Body Source Validation
//
// A [Request] carries exactly one body source: [Request.Message] (JSON/encoded),
// [Request.Form] (multipart), or [Request.BodyReader] (raw). Setting more than one
// produces [ErrMultipleBodySources] during [RequestWithContext], before any network
// call occurs. [Request.DownloadURL] is mutually exclusive with all body sources.
//
// # App Secret Proof
//
// Requests can include an appsecret_proof query parameter for secure API calls.
// Set both the app secret and the secured flag:
//
//	// functional option:
//	req := http.MakeRequest[MyMsg](method, url,
//	    http.WithRequestAppSecret("mysecret"),
//	    http.WithRequestSecured(true),
//	)
//
//	// builder:
//	req := http.NewRequestBuilder(method, url).AppSecret("mysecret").Secured(true)
//
//	// or in one call:
//	req := http.NewRequestBuilder(method, url).Auth(http.AuthConfig{
//	    AccessToken: "token",
//	    AppSecret:   "mysecret",
//	    Secure:      true,
//	})
//
// The proof is only generated when both [Request.SecureRequests] is true and
// [Request.AppSecret] is non-empty (see [Request.URL]).
//
// # Media Downloads
//
// For downloading media (e.g., images, audio), use the download-specific APIs
// that bypass the base URL and body encoding:
//
//	// functional option:
//	req := http.MakeDownloadRequest[MyMsg](downloadURL, http.WithRequestBearer("token"))
//
//	// builder:
//	req := http.NewDownloadBuilder(downloadURL).Bearer("token")
//
// For requests with no typed body at all, use [BuildAnyRequest]:
//
//	req := http.BuildAnyRequest(http.NewRequestBuilder(http.MethodGet, url).Bearer("token"))
//
// # Custom Payload Encoding
//
// Types that implement the [PayloadEncoder] interface bypass the built-in encoding
// dispatch (JSON, form, raw bytes/string). This allows a custom type to control
// how it is serialized into an HTTP request body without modifying the encoding
// layer:
//
//	type CustomPayload struct { ... }
//	func (p *CustomPayload) EncodePayload(ctx context.Context, w io.Writer) error { ... }
//
// The dispatch order in [EncodePayloadWithContext] is: [PayloadEncoder] →
// [*RequestForm] (multipart) → [io.Reader] → [[]byte] → [string] → JSON.
//
// # Raw Body Decoding
//
// For non-JSON response parsing (e.g., binary data, XML), use
// [BodyReaderResponseDecoder] which converts a raw reader function into a
// [ResponseDecoder]:
//
//	decoder := http.BodyReaderResponseDecoder(func(ctx context.Context, r io.Reader) error {
//	    data, err := io.ReadAll(r)
//	    // parse data as needed
//	    return err
//	})
//
// # Request URL Determinism
//
// Query parameters are sorted alphabetically when building the URL (see
// [Request.URL]), guaranteeing deterministic output for testing and caching. The
// appsecret_proof and debug parameters are appended after sorting.
//
// # Sender Flexibility
//
// [BaseClient.SetSender] replaces the underlying [Sender] at runtime, and
// [BaseClient.SetMiddlewares] dynamically swaps the middleware chain. Both are
// not goroutine-safe with concurrent [BaseClient.Send] calls — use them during
// setup, not during active request processing.
//
// [BaseClient.CloseIdleConnections] uses an inline interface check against the
// sender, so it works for any sender that implements CloseIdleConnections,
// not just [CoreClient].
package http
