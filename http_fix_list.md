# `pkg/http` Findings Validation & PR Plan

> **Source:** `findings.md` (merged from `http_findings.md` + `deepseek_http_findings.md`)  
> **Validated against:** `pkg/http/*.go` and production callers (`message/`, `phonenumber/`, `media/`, `qrcode/`, `auth/`, `business/`, etc.)

---

## Validation Summary

| # | Finding | Status | Notes |
|---|---------|--------|-------|
| **1.1** | `RequestType.String()` out-of-bounds panic | **FIXED** | Guard added at `request_type.go:185` |
| **1.2** | `encodeFormData` nil `FormFile` panic | **FIXED** | Returns `ErrMissingFormFile` at `encoder.go` |
| **1.3** | `CoreClient` setters are racy | **FIXED** | Setters removed; `CoreClient` is now immutable after construction |
| **1.4** | Default `http.DefaultClient` has zero timeout | **FIXED** | Defaults to `&http.Client{Timeout: 30s}` |
| **1.5** | Unbounded `io.ReadAll` on bodies | **FIXED** | `readAllLimited` with configurable `MaxBodyBytes` (default 10MB) |
| **1.6** | Multipart header escaping incomplete | **FIXED** | Uses `mime.FormatMediaType` for safe header construction |
| **1.7** | `ResponseError.Error` nil panic | **FIXED** | Nil guard added at `errors.go` |
| **2.1** | `DecodeRequestJSON` wrong sentinel | **FIXED** | Returns `ErrNilRequest` |
| **2.2** | File handle cleanliness on encoding errors | **FIXED** | `multipartWriter` closed via `defer` in goroutine; pipe closed on error |
| **2.3** | `Content-Type` set on GET | **FIXED** | Only set when `body != nil` |
| **2.4** | Success range includes 3xx | **FIXED** | Now uses `>= 200 && < 300` |
| **2.5** | Query-param iteration non-deterministic | **FIXED** | Keys sorted before encoding |
| **3.1** | Response body double-buffered | **FIXED** | Uses `bytes.NewReader` (zero-copy) |
| **3.2** | Multipart uploads fully buffered | **FIXED** | Streamed via `io.Pipe` |
| **3.3** | `BodyReaderResponseDecoder` buffers everything | **FIXED** | Passes `response.Body` directly to callback |
| **3.4** | Middleware chain rebuilt every `Send` | **FIXED** | Pre-computed once during `NewSender` construction |
| **3.5** | `Request.URL()` heavy allocations | **OPEN** | `url.JoinPath` → `url.Parse` → `q.Encode()` |
| **3.6** | JSON encoder trailing newline | **FIXED** | Uses `json.Marshal` instead of `json.NewEncoder` |
| **4.1** | Generic option noise | **PARTIALLY ADDRESSED** | `RequestBuilder` introduced in `request_builder.go` |
| **4.2** | `NewAnySender` redundant | **FIXED** | Removed; callers use `NewSender[any]` |
| **4.3** | Exported `Options` struct unused | **FIXED** | Removed from `http.go` |
| **4.4** | Mixed config APIs (options + setters) | **FIXED** | Setters removed; only functional options remain |
| **4.5** | Type parameter proliferation | **PARTIALLY ADDRESSED** | `RequestBuilder` removes generics for type-independent config |
| **4.6** | Duplicate request abstractions in sub-packages | **OPEN** | Affects `message/`, `phonenumber/`, `qrcode/`, `media/`, `flow/`, `business/` |
| **4.7** | Duplicate middleware wrapping logic | **OPEN** | `qrcode/qrcode.go:306` copies `pkg/http` logic |
| **4.8** | Missing `DecodeOptions` presets | **OPEN** | No `StrictDecodeOptions()` / `LenientDecodeOptions()` |
| **4.9** | `SetBaseSender` exported unnecessarily | **FIXED** | Removed; use `WithCoreClientSender` option instead |
| **4.10** | `EncodePayload` closed type switch | **OPEN** | No `PayloadEncoder` interface |
| **4.11** | Body source precedence ambiguity | **FIXED** | Returns `ErrMultipleBodySources` when more than one source is supplied |
| **4.12** | Manual decoder assembly | **OPEN** | No `SendJSON` convenience exists |

**Already fixed:** 21 findings (1.1–1.7, 2.1–2.5, 3.1–3.4, 3.6, 4.2–4.4, 4.9, 4.11)  
**Partially addressed:** 2 findings (4.1, 4.5) via `RequestBuilder`  
**Still open:** 5 findings (3.5, 4.6, 4.7, 4.8, 4.10, 4.12)

---

## PR Plan (1 PR per item)

### PR 1: Fix Unbounded Memory Reads & Streaming
**Findings:** 1.5, 3.3  
**Priority:** P0  
**Status:** ✅ **DONE**  
**Files:** `pkg/http/http.go`, `pkg/http/decoder.go`  
**Work:**
- Added `DefaultMaxBodyBytes` (10MB) and `DefaultMaxHeaderBytes` (1MB) constants.
- Added `maxBodyBytes` / `maxHeaderBytes` fields to `CoreClient` with `WithCoreClientMaxBodyBytes` / `WithCoreClientMaxHeaderBytes` options.
- Added `readAllLimited` helper that errors with `ErrBodyTooLarge` when limit is exceeded.
- Wrapped all `io.ReadAll` calls in interceptors and decoders with `readAllLimited`.
- Changed `BodyReaderResponseDecoder` to pass `response.Body` directly to the callback instead of buffering.

---

### PR 2: Harden & Stream Payload Encoder
**Findings:** 1.6, 2.2, 3.2, 3.6  
**Priority:** P0  
**Status:** ✅ **DONE**  
**Files:** `pkg/http/encoder.go`  
**Work:**
- Replaced naïve `fmt.Sprintf` header construction with `mime.FormatMediaType` for safe `Content-Disposition` escaping.
- Streamed multipart uploads via `io.Pipe` instead of buffering into `bytes.Buffer`.
- Ensured `multipartWriter.Close()` is called via `defer` in the goroutine; pipe is closed with error on failure.
- Replaced `json.NewEncoder(buf).Encode(p)` with `json.Marshal` to eliminate the trailing newline.

---

### PR 3: Make CoreClient Immutable & Thread-Safe
**Findings:** 1.3, 3.4, 4.4, 4.9  
**Priority:** P0  
**Status:** ✅ **DONE**  
**Files:** `pkg/http/http.go`, `pkg/http/http_test.go`, `README.md`, `_examples/auth/main.go`, `_examples/qr/main.go`  
**Work:**
- Removed all mutable setters (`SetHTTPClient`, `SetRequestInterceptor`, `SetResponseInterceptor`, `AppendMiddlewares`, `PrependMiddlewares`, `SetBaseSender`).
- Added `WithCoreClientSender` functional option for injecting a custom sender.
- Configured everything via functional options in `NewSender`.
- Pre-computed the wrapped middleware chain once during construction instead of in every `Send`.
- Updated tests and examples to use constructor options instead of setters.

---

### PR 4: Fix HTTP Semantics & Body Precedence
**Findings:** 2.3, 2.4, 4.11  
**Priority:** P1  
**Status:** ✅ **DONE**  
**Files:** `pkg/http/request.go`, `pkg/http/decoder.go`, `pkg/http/errors.go`  
**Work:**
- Only set `Content-Type` when `body != nil` (skip for GET with no body).
- Changed `isResponseOk` from `< 308` to `>= 200 && < 300`.
- Enforced mutual exclusivity for `Message`, `Form`, and `BodyReader` — returns `ErrMultipleBodySources` when more than one is supplied.

---

### PR 5: Decoder Ergonomics & Presets
**Findings:** 4.8, 4.12  
**Priority:** P2  
**Files:** `pkg/http/decoder.go`, callers in sub-packages  
**Work:**
- Add `StrictDecodeOptions()` and `LenientDecodeOptions()` presets.
- Add a first-class `SendJSON` convenience on `CoreClient` that assembles `ResponseDecoderJSON` automatically.
- Update representative call sites (e.g. `message/base_client.go`) to use presets.

---

### PR 6: Deduplicate Middleware & Remove Redundancies
**Findings:** 4.2, 4.7  
**Priority:** P2  
**Files:** `pkg/http/http.go`, `qrcode/qrcode.go`, `message/base_client.go`, `phonenumber/phonenumber.go`, etc.  
**Work:**
- `NewAnySender` already removed in PR 3.
- Export `WrapMiddlewares[T]` from `pkg/http` (already exported) and replace duplicate implementations in `qrcode/`, `message/`, and any other sub-packages.

---

### PR 7: Unify Sub-Package Request Abstractions
**Finding:** 4.6  
**Priority:** P3  
**Files:** `message/`, `phonenumber/`, `qrcode/`, `media/`, `flow/`, `business/`, etc.  
**Work:**
- Introduce a generic high-level `Do` helper or `FromConfig` builder in `pkg/http` that collapses the repeated pattern: read config → build options → call `whttp.MakeRequest` → decode.
- Migrate one sub-package at a time to prove the pattern, then roll out to the rest.

---

### PR 8: Encoder Extensibility & URL Allocations
**Findings:** 3.5, 4.10  
**Priority:** P3  
**Files:** `pkg/http/encoder.go`, `pkg/http/request.go`  
**Work:**
- Introduce a `PayloadEncoder` interface so new payload kinds don't require editing `EncodePayload`.
- Optimize `Request.URL()` to reduce allocations (e.g. `strings.Builder` for path segments, cache parsed URL).

---

*End of PR plan.*
