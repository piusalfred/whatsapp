# Unified `pkg/http` Review

**Date:** 2026-06-06  
**Scope:** `pkg/http/*.go` and all production callers (`message/`, `phonenumber/`, `media/`, `qrcode/`, `auth/`, `business/`, etc.)  
**Sources:** Merged from `http_findings.md` (Kimi) + `deepseek_http_findings.md` (DeepSeek)

---

## Executive Summary

The `pkg/http` package is a capable, generic HTTP abstraction, but it suffers from three categories of problems:

1. **Safety / Security** — unbounded memory reads, missing bounds checks, nil-pointer panics, and a default HTTP client with zero timeout.
2. **Concurrency** — `CoreClient` exposes mutable setters without synchronization, making it unsafe for shared or long-lived use.
3. **Ergonomics / Design Debt** — every request option carries a generic type parameter, forcing callers to repeat `[Message]` or `[any]` dozens of times across the codebase. There are also redundant constructors, dead code, and duplicated middleware wrapping in sub-packages.

To address the ergonomic pain-point **immediately**, a new **`RequestBuilder`** (`pkg/http/builder.go`) has been introduced. It lets callers configure the type-independent parts of a request without any generics, attaching the typed message body only at the final step.

---

## 1. Critical Issues

### 1.1 `RequestType.String()` — Out-of-Bounds Panic (DoS)
**Files:** `pkg/http/request_type.go:183-185`  
`RequestType` is `uint8`. `String()` indexes `requestTypeStrings[r]` with no bounds check. A value ≥ the sentinel (`requestTypeCount`) crashes the process.

**Fix:** Add a guard and return a safe fallback string.

### 1.2 `encodeFormData` — Nil `FormFile` Panic
**Files:** `pkg/http/encoder.go:103`  
If `Form` is non-nil but `FormFile` is nil, `os.Open(form.FormFile.Path)` dereferences a nil pointer.

**Fix:** Return a typed error (`ErrMissingFormFile`) when `form.FormFile == nil`.

### 1.3 `CoreClient` Setters Are Not Thread-Safe
**Files:** `pkg/http/http.go:59-83`  
`SetHTTPClient`, `SetRequestInterceptor`, `AppendMiddlewares`, `PrependMiddlewares`, and `SetBaseSender` mutate fields without a mutex. `Send` reads those same fields concurrently. This is a data race under any real concurrent load.

**Fix (preferred):** Make `CoreClient` immutable after construction — remove the setters and configure everything via functional options in `NewSender`.  
**Fix (alternative):** Protect the struct with `sync.RWMutex`.

### 1.4 Default `http.DefaultClient` Has No Timeout
**Files:** `pkg/http/http.go:111`  
`NewSender` defaults to `http.DefaultClient`, which has **zero** request, dial, and TLS-handshake timeouts. A slow or malicious peer can hang goroutines indefinitely.

**Fix:** Default to a fresh `*http.Client{Timeout: 30s, Transport: …}` instead of the global.

### 1.5 Unbounded `io.ReadAll` — Memory-Exhaustion DoS
**Files:**
- `pkg/http/http.go:174` (`SendFuncWithInterceptors`)
- `pkg/http/decoder.go:42,103,168`
- `pkg/http/encoder.go:127` (`encodeFormData`)

Entire request/response bodies (and uploaded files) are slurped into memory with no size cap.

**Fix:** Use `http.MaxBytesReader` or `io.LimitReader` with a configurable maximum. For multipart uploads, consider streaming via `io.Pipe`.

### 1.6 Multipart Header Escaping Is Incomplete
**Files:** `pkg/http/encoder.go:112-114`  
`Content-Disposition` is built with `fmt.Sprintf` and a naïve quote escaper that only handles `"` and `\`. It does **not** protect against CRLF injection or other control characters in filenames.

**Fix:** Use `multipart.Writer.CreateFormFile(name, filename)` or `mime.FormatMediaType` for the header value.

### 1.7 `ResponseError.Error` Can Panic on Nil Inner Error
**Files:** `pkg/http/errors.go:34-36`  
```go
return fmt.Sprintf("...%s", strings.ToLower(e.Err.Error()))
```
If `e.Err` is nil, this panics.

**Fix:** Guard the nil case before calling `Error()`.

---

## 2. Bugs & Reliability

### 2.1 `DecodeRequestJSON` Returns Wrong Sentinel
**Files:** `pkg/http/decoder.go:99-101`  
A nil `*http.Request` argument returns `ErrNilResponse` instead of the existing `ErrNilRequest`. This is a copy-paste bug.

### 2.2 File Handle Cleanliness on Encoding Errors
**Files:** `pkg/http/encoder.go:103-107`  
If `os.Open` succeeds but a later step (`CreatePart`, `io.Copy`) fails, the `defer file.Close()` handles the file, but the `multipart.Writer` may be left in an incomplete state. Very low severity, but worth ensuring the writer is closed or documented as best-effort.

### 2.3 `Content-Type` Is Always Set (Even for GET)
**Files:** `pkg/http/request.go:287`  
`RequestWithContext` unconditionally writes `Content-Type: application/json`. For GET requests with no body this is technically incorrect and can confuse WAFs or signature validators.

**Fix:** Only set `Content-Type` when a body is present.

### 2.4 Success Status-Code Range Includes 3xx
**Files:** `pkg/http/decoder.go:54`  
`isResponseOk` uses `< http.StatusPermanentRedirect` (308), which treats `301-307` as success. The standard client normally follows redirects, but this is an unusual choice.

**Fix:** Use the conventional `>= 200 && < 300`.

### 2.5 Query-Parameter Iteration Is Non-Deterministic
**Files:** `pkg/http/request.go:221-240`  
Map iteration randomizes query-string ordering, hurting cacheability and log comparability.

**Fix:** Sort keys before appending to `url.Values`.

---

## 3. Performance Findings

### 3.1 Response Body Is Buffered Multiple Times
**Files:** `pkg/http/http.go:174-183`  
When a response hook is active, the body is read once, then wrapped in `bytes.NewBuffer` twice. `bytes.NewBuffer` copies the slice; `bytes.NewReader` is zero-copy.

**Fix:** Replace both restorations with `io.NopCloser(bytes.NewReader(bodyBytes))`.

### 3.2 Multipart Uploads Buffer the Entire File
**Files:** `pkg/http/encoder.go:93-137`  
`encodeFormData` copies the whole file into a `bytes.Buffer`. For WhatsApp document uploads (up to 100 MB) this is significant memory pressure.

**Fix:** Stream via `io.Pipe` or, at minimum, stat the file and set `ContentLength` on a streaming reader.

### 3.3 `BodyReaderResponseDecoder` Buffers Everything
**Files:** `pkg/http/decoder.go:166-183`  
It reads the entire body before handing a reader to the callback. Large media downloads can OOM.

**Fix:** Pass `response.Body` directly to the callback (the caller can decide whether to buffer).

### 3.4 Middleware Chain Rebuilt on Every `Send`
**Files:** `pkg/http/http.go:195-199`  
`wrapMiddlewares` is called on every invocation. The overhead is small, but unnecessary for an immutable client.

**Fix:** Pre-compute the wrapped chain once during construction.

### 3.5 `Request.URL()` Allocates Heavily
**Files:** `pkg/http/request.go:206-243`  
URL construction goes through `url.JoinPath` → `url.Parse` → map iteration → `q.Encode()`. High-throughput callers could benefit from a `strings.Builder` path or URL caching.

### 3.6 JSON Encoder Adds Trailing Newline
**Files:** `pkg/http/encoder.go:74`  
`json.NewEncoder(buf).Encode(p)` appends a newline. It is harmless for most decoders but adds an unexpected byte for callers that sign or hash bodies.

**Fix:** Use `json.Marshal` instead.

---

## 4. API Ergonomics & Design Debt

### 4.1 Generic Option Noise *(Partially Addressed — See §5)*
**Files:** `pkg/http/request.go` (all `WithRequest*` options)  
Every option requires a type parameter, yet the parameter is almost never used in the function body. Callers are forced to write:

```go
whttp.WithRequestBearer[Message](token),
whttp.WithRequestAppSecret[Message](secret),
```

**Mitigation:** The new `RequestBuilder` eliminates generics for all type-independent configuration.

### 4.2 `NewAnySender` Is Redundant
**Files:** `pkg/http/http.go:125-139`  
It is a manual copy of `NewSender[any]`. Can be a one-line wrapper or removed entirely.

### 4.3 Exported `Options` Struct Is Unused
**Files:** `pkg/http/http.go:46-50`  
Dead code — never referenced anywhere in the project.

### 4.4 Mixed Configuration APIs (Options + Setters)
**Files:** `pkg/http/http.go`  
Callers can configure via functional options at construction **or** via mutable setters afterwards. The overlap is confusing and unsafe (see §1.3).

**Fix:** Remove setters; configure once at construction.

### 4.5 Type Parameter Proliferation
**Files:** `pkg/http/request.go:118+`  
Many option bodies do not touch `T` at all. A non-generic builder or plain functions for headers, bearer, query params, etc. would remove noise without losing type safety for the message body.

### 4.6 Duplicate Request Abstractions in Sub-Packages
**Observed in:** `message/`, `phonenumber/`, `qrcode/`, `media/`, etc.  
Each sub-package reinvents its own `BaseRequest`, `BaseSender`, and `SenderMiddleware`. They all do the same thing: read config → build `whttp.RequestOption[T]` → call `whttp.MakeRequest` → decode.

**Fix:** A higher-level `Do` helper or `FromConfig` builder inside `pkg/http` would collapse this duplication.

### 4.7 Duplicate Middleware Wrapping Logic
**Files:** `pkg/http/http.go:223-231` and `message/base_client.go:135-141`  
Both reverse-iterate a slice, skipping nil middlewares. The logic should live in one place.

### 4.8 `DecodeOptions` Lacks Presets
**Files:** `pkg/http/decoder.go:31-35`  
Every caller manually constructs the struct. Presets like `StrictDecodeOptions()` and `LenientDecodeOptions()` would reduce boilerplate.

### 4.9 `SetBaseSender` Exported Unnecessarily
**Files:** `pkg/http/http.go:69-71`  
Exposing this breaks encapsulation; external code can replace the internal sender and bypass interceptors.

**Fix:** Unexport or remove.

### 4.10 `EncodePayload` Type Switch Is Closed
**Files:** `pkg/http/encoder.go:40-83`  
Adding a new payload kind means editing the central switch. An interface like `PayloadEncoder` would be more extensible.

### 4.11 Request Body Precedence Is Implicit
**Files:** `pkg/http/request.go:259-280`  
`Message`, `Form`, and `BodyReader` are checked in independent `if` blocks, so later sources silently overwrite earlier ones.

**Fix:** Use `if/else if` or return an error when more than one source is supplied.

### 4.12 Decoder Must Always Be Assembled Manually
**Files:** `pkg/http/decoder.go`  
Every call site does `decoder := whttp.ResponseDecoderJSON(&resp, opts)` before `Send`. A first-class `SendJSON` convenience would remove this boilerplate.

---

## 5. What Was Implemented — `RequestBuilder`

**File:** `pkg/http/builder.go`  
**Tests:** `pkg/http/builder_test.go`

A new, **non-generic** fluent builder captures all type-independent request fields. The generic parameter is only needed when attaching the message body.

### Usage Example

```go
// Type-independent configuration — zero generics.
b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
    WithEndpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint).
    WithBearer(conf.AccessToken).
    WithAppSecret(conf.AppSecret, conf.SecureRequests).
    WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
    WithMetadata(request.Metadata)

// Attach the typed message at the very end.
req := whttp.BuildRequest(b, request.Message)
```

For requests that do not need a typed body (e.g. GET or downloads):

```go
req := whttp.BuildAnyRequest(b)
```

### API Surface

```go
func NewRequestBuilder(method, baseURL string) *RequestBuilder
func (b *RequestBuilder) WithEndpoints(endpoints ...string) *RequestBuilder
func (b *RequestBuilder) WithBearer(bearer string) *RequestBuilder
func (b *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder
func (b *RequestBuilder) WithQueryParams(params map[string]string) *RequestBuilder
func (b *RequestBuilder) WithAppSecret(secret string, secure bool) *RequestBuilder
func (b *RequestBuilder) WithDebugLogLevel(level DebugLogLevel) *RequestBuilder
func (b *RequestBuilder) WithMetadata(metadata types.Metadata) *RequestBuilder
func (b *RequestBuilder) WithDownloadURL(url string) *RequestBuilder
func (b *RequestBuilder) WithForm(form *RequestForm) *RequestBuilder
func (b *RequestBuilder) WithBodyReader(r io.Reader) *RequestBuilder

func BuildRequest[T any](b *RequestBuilder, message *T) *Request[T]
func BuildAnyRequest(b *RequestBuilder) *Request[any]
```

This keeps the existing `MakeRequest` / option API intact (no breaking changes) while providing an ergonomic, generic-free path for the common case.

---

## 6. Unified Prioritized Recommendations

### P0 — Fix Immediately

| # | Issue | Impact |
|---|-------|--------|
| 1.1 | `RequestType.String()` out-of-bounds panic | **Process crash / DoS** |
| 1.2 | `encodeFormData` nil `FormFile` dereference | **Process crash** |
| 1.3 | `CoreClient` setters are racy | **Data corruption / crashes under concurrency** |
| 1.4 | `http.DefaultClient` has zero timeout | **Goroutine leak / DoS** |
| 1.5 | Unbounded `io.ReadAll` on bodies | **Memory exhaustion** |
| 1.6 | Multipart header escaping incomplete | **Header injection** |
| 1.7 | `ResponseError.Error` nil panic | **Process crash on malformed error** |

### P1 — High Priority

| # | Issue | Impact |
|---|-------|--------|
| 2.1 | `DecodeRequestJSON` wrong sentinel | Misleading errors |
| 3.1 | Response body double-buffered (`bytes.NewBuffer`) | Wasted memory |
| 3.2 | Multipart uploads fully buffered | OOM on large files |
| 3.3 | `BodyReaderResponseDecoder` buffers everything | OOM on media downloads |
| 4.4 | Mixed config APIs (options + setters) | API ambiguity & races |

### P2 — Medium Priority

| # | Issue | Impact |
|---|-------|--------|
| 2.3 | `Content-Type` set on GET | Spec non-compliance |
| 2.4 | Success range includes 3xx | Unusual semantics |
| 2.5 | Query param ordering random | Cacheability / log noise |
| 3.4 | Middleware chain rebuilt every `Send` | Minor overhead |
| 3.6 | JSON encoder trailing newline | Signature / hashing surprise |
| 4.2 | `NewAnySender` redundant | Code bloat |
| 4.3 | Unused `Options` struct | Dead code |
| 4.7 | Duplicate middleware wrapping | Maintenance burden |
| 4.8 | Missing `DecodeOptions` presets | Boilerplate |
| 4.11 | Body source precedence ambiguity | Silent mis-encoding |

### P3 — Nice to Have

| # | Issue | Impact |
|---|-------|--------|
| 3.5 | `Request.URL()` heavy allocations | Micro-optimization |
| 4.6 | Duplicate request abstractions in sub-packages | Architectural debt |
| 4.9 | `SetBaseSender` exported unnecessarily | API surface bloat |
| 4.10 | `EncodePayload` closed type switch | Extensibility |
| 4.12 | Manual decoder assembly | Boilerplate |

---

*End of unified report.*
