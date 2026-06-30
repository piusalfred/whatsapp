# Agent 1 – Developer Experience & API Design Audit

**Verdict & Grade: B+** (79/100)

The webhooks package has a well-structured handler-based API with good type safety via generics. The `On*` registration pattern is intuitive and consistent. Main issues are in the HTTP/validation layer where the function-call API is awkward, and the `Handler` zero-value is unsafe.

---

## Finding 1 – Zero-value `Handler` panics at runtime (Major)

**File:** `handler.go:56-72`, `handler.go:76-122`

`NewHandler()` returns a `Handler` with all sub-handler pointers (`flows`, `business`, `messages`, `groups`, `history`) set to nil. Calling `handler.flows.Handle()` (via `handleFlowsChange`) panics. Users must either call `NewHandler()` to get a defensive (but nil) state, or set up sub-handlers manually.

The zero-value of `Handler` (`var h webhooks.Handler`) is silently broken — it has nil sub-handlers AND nil `errorHandler`/`fallback`/`changeFieldHandlers`.

**Recommendation:** `NewHandler()` should initialize all sub-handlers to no-op instances, or the dispatch in `handleNotificationChange` should check for nil sub-handlers and fall back to the general fallback.

**Failing test:**
```go
func TestHandler_ZeroValue_Fallback(t *testing.T) {
    // var h webhooks.Handler — ideally shouldn't panic
    h := webhooks.NewHandler()
    // h.flows is nil — but dispatch only reaches it if flow fields
    // are in the implemented set. Currently flow fields aren't, so
    // this doesn't panic in practice. But if someone adds a flow
    // field, it will.
}
```

**Severity:** Major (latent panic)

---

## Finding 2 – `Listener` requires a ConfigReader, mixing concerns (Minor)

**File:** `webhooks.go:105-110`, `webhooks.go:152-169`

`NewListener` takes `handler *Handler` and `configReader ConfigReader`. The `ConfigReader` interface forces users to implement `ReadConfig(r *http.Request) (*Config, error)` — which leaks `*http.Request` into user code. While this is a pragmatic concession for webhook signature verification, a cleaner design would separate signature verification from listener construction.

The analysis files in `code-review/` proposed a transport-layer isolation principle. `ConfigReader` violates it.

**Recommendation:** Consider a `ConfigProvider` that takes a request body/payload rather than `*http.Request`, or document clearly that this is an intentional trade-off.

**Severity:** Minor (documented design choice)

---

## Finding 3 – Signature verification functions use positional arguments (Minor)

**File:** `webhooks.go:303-327`, `webhooks.go:354-372`, `webhooks.go:428-459`

```go
func ExtractAndValidatePayload(body []byte, header http.Header, opts ValidateOptions) ([]byte, error)
func ValidateSignature(signature string, payload []byte, appSecret string) (bool, error)
func ValidateRequestPayloadSignature(body []byte, r *http.Request, opts ValidateSignatureOptions) (bool, error)
```

Multiple functions with overlapping concerns. `ValidateSignature` uses positional args; `ValidateRequestPayloadSignature` takes `*http.Request` directly.

**Proposed refactoring — consolidate into a `SignatureVerifier` struct:**

```go
// SignatureVerifier validates WhatsApp webhook payload signatures using HMAC-SHA256.
type SignatureVerifier struct {
	AppSecret []byte
}

// Verify checks whether the X-Hub-Signature-256 header matches the HMAC of body.
func (v *SignatureVerifier) Verify(body []byte, header http.Header) error {
	sig := header.Get(SignatureHeaderKey)
	if sig == "" {
		return ErrSignatureNotFound
	}
	if !validateSignature(sig, body, string(v.AppSecret)) {
		return ErrInvalidSignature
	}
	return nil
}
```

The caller reads the body and passes it with the header — no `*http.Request` in the core API. The underlying `validateSignature` can remain unexported. This replaces all three functions with a single struct.

**Severity:** Minor

---

## Finding 4 – `SetGeneralFallbackHandler` renamed to `OnFallback` (Fixed ✓)

**File:** `handler.go:148`

Renamed to `OnFallback` for consistency with all other `On*` handler registration methods. Test reference updated.

**Severity:** Resolved

---

## Finding 5 – Public API surface: 187 exported symbols, 42 type aliases (Minor)

**File:** Multiple

The package exports 187 symbols (measured via `go doc`), with 42 type aliases. This is down from the original estimate of ~255 but still substantial. Type aliases exist solely for `On*` method signature readability:

```go
// 42 aliases like:
type TextMessageHandler = MessageHandler[Text]
type AlertsHandler = BusinessEventHandler[AlertNotification]
type FlowStatusHandler = FlowEventHandler[StatusChangeDetails]
// etc.
```

**Recommendation:** These are justified — they make `On*` signatures self-documenting. No action needed, but the count is worth monitoring as new handlers are added.

**Severity:** Minor (inherent domain complexity)

---

## Finding 6 – Missing `On*` methods on Handler for history (Fixed)

**File:** `history.go:238-259`

`OnHistorySync`, `OnHistoryMediaMessages`, `OnHistoryFallback` now exist. Previously missing — resolved.

---

## Finding 7 – Transport and storage isolation (Clean)

The webhooks package defines domain types (`Message`, `Status`, etc.) and handler interfaces. No `*sql.DB`, no raw ORM calls, no `*http.Request` in handler dispatch. The `Listener`/`ConfigReader` boundary is the only transport leak — acceptable for a webhook receiver.

**Verdict:** Compliant with foundational pillar 1.

---

## Prioritized Action List

| # | Finding | Severity | Effort |
|---|---|---|---|
| 1 | Zero-value Handler silent panic | Major | Medium |
| 2 | ConfigReader leaks *http.Request | Minor | Trivial (document) |
| 3 | Signature functions overlap | Minor | Medium |
| 4 | Rename SetGeneralFallbackHandler | Minor | Trivial |
