# IMPROVE.md — Webhooks Package Audit

Generated 2026-07-25. Covers `webhooks/`, `webhooks/router/`, and `webhooks/callbacks/`.

---

## 1. Bugs & Critical Defects

### 1.1 `html.EscapeString` breaks subscription verification

**File:** `listener.go:218`
**Severity:** CRITICAL

`HandleSubscriptionVerification` HTML-escapes the challenge before echoing it back:

```go

_, _ = writer.Write([]byte(html.EscapeString(challenge)))

```

WhatsApp expects the challenge returned **verbatim**. Any character that `html.EscapeString` transforms (`<`, `>`, `&`, `"`, `'`) will cause verification to fail. Replace with plain `writer.Write([]byte(challenge))`.

### 1.2 `SetSignatureVerifier(nil)` causes panic

**File:** `listener.go:191-193, 289-294`
**Severity:** CRITICAL

`SetSignatureVerifier` accepts nil with no guard. When `ParseNotification` is called with `VerifyPayloadSignature == true`, it dereferences the nil `ls.verifier` at line 290. Fix: reject nil in `SetSignatureVerifier` or fall back to the default verifier.

### 1.3 `Contact.SenderInfo()` nil-dereference

**File:** `webhooks.go:392-397`
**Severity:** CRITICAL

`SenderInfo()` dereferences `c.Profile.Name` without checking if `c` or `c.Profile` is nil. Both `MessageNotificationContext.AllSendersInfo` (line 725) and `SenderInfo` (line 732) call into it. Fix: add nil guards.

### 1.4 `handleSystem` passes nil `Identity` to handler

**File:** `message.go:735-742`
**Severity:** MEDIUM

When `message.System.Type == "user_changed_number"`, the code passes `message.Identity` to the handler without checking if it's nil. `Identity` is a separate field from `System` — a `user_changed_number` system message may not carry an identity block.

### 1.5 Dead code: `BusinessNotificationHandler.calls` never reached

**File:** `business.go:279` (field), `business.go:638` (IsEventHandlerImplemented)
**Severity:** MEDIUM

`ChangeFieldCalls` is routed to `CallsHandler` via `GetChangeFieldCategory` (field.go:399), not to `BusinessNotificationHandler`. The `calls` field, `OnCalls` setter, and the `IsEventHandlerImplemented` case for calls in BusinessNotificationHandler are **dead code**.

### 1.6 Error→HTTP 504 semantic mismatch in parallel dispatch

**File:** `handler.go:305`, `event.go:172`
**Severity:** MEDIUM

Parallel dispatch maps ANY `g.Wait()` error (including handler logic errors) to HTTP 504 (Gateway Timeout). Sequential dispatch maps the same errors to HTTP 500. Same error path, different status code depending on dispatch mode.

---

## 2. Dead Code & Unused Exports

### 2.1 Duplicate entries in `implemented` list

**File:** `handler.go:121-123, 137-139`

`ChangeFieldTemplateCategoryUpdate`, `ChangeFieldTemplateQualityUpdate`, `ChangeFieldTemplateComponentsUpdate` appear twice in the `implemented` slice passed to `initChangeFieldMap`. Harmless (map insertion is idempotent) but indicates copy-paste error.

### 2.2 `HistorySyncContext` never constructed

**File:** `history.go:41-44`

Defined but unused — `HistoryHandler.Handle` constructs `MessageNotificationContext` instead.

### 2.3 `CallStatusUpdate` effectively dead

**File:** `calls.go:306-323`, `notification.go:306-312`

Only reachable via the legacy `OnCallStatusUpdate` on `*Handler` (line 394), which routes to `BusinessNotificationHandler.calls` — itself dead code (see 1.5). Also: `MessagingProduct` field has no JSON tag.

### 2.4 14 `MessageType*` constants never dispatched

**File:** `message.go:54-64`

`MessageTypeGif`, `MessageTypeGroupInvite`, `MessageTypeHsm`, `MessageTypeKeepInChat`, `MessageTypeLinkPreview`, `MessageTypeList`, `MessageTypeMediaPlaceholder`, `MessageTypePin`, `MessageTypePollCreation`, `MessageTypePollUpdate`, `MessageTypeProduct`, etc. fall through to the `default` branch and are treated as unknown.

### 2.5 Duplicate `OnHandler` / `OnEcho` / `OnChange` / `OnSync`

**Files:** `smb.go:168,303`, `userpreferences.go:82,93`

`SMBMessageEchoesHandler` has both `OnEcho` and `OnHandler` doing the same thing. Same pattern on `SMBAppStateSyncsHandler` (`OnSync`/`OnHandler`) and `UserPreferencesHandler` (`OnChange`/`OnHandler`). Pick one and remove the duplicate.

### 2.6 `NewWebhookRouter` always returns nil error

**File:** `router/router.go:135`

The function signature returns `error` but never returns a non-nil error. Either remove the return or add validation (e.g., reject nil listener).

---

## 3. Naming & Semantic Incoherence

### 3.1 `OnMessageErrors` maps to `handler.messages.Unknown`

**File:** `status.go:113`

The name "MessageErrors" doesn't convey that these go to the **unknown** handler. This is confusing — errors can accompany known message types too.

### 3.2 `OnTemplateComponentsUpdate` missing `Business` prefix

**File:** `business.go:664`

All other business `On*` methods on `*Handler` use `OnBusiness*` — this one doesn't. Should be `OnBusinessTemplateComponentsUpdate`.

### 3.3 `ErrorInfo.Error()` name misleading

**File:** `notification.go:148`

Named `Error()` but returns `*werrors.Error`, not `string`. Does NOT implement the `error` interface. Should be named `ToError()` or `AsError()`.

### 3.4 `verifySubscriptionRequest` returns `ErrInvalidSignature`

**File:** `listener.go:394`

Token mismatch or wrong mode returns `ErrInvalidSignature` — but it's not a signature failure. Should use a dedicated `ErrSubscriptionVerification`.

### 3.5 `HandleNotificationEvents` ignores `handler.mode`

**File:** `event.go:152`

Always dispatches in parallel regardless of `handler.mode`. By design per doc, but inconsistent with `HandleNotification` which respects the mode.

---

## 4. JSON Tag Issues

| File | Line | Field | Issue |
|------|------|-------|-------|
| notification.go | 100 | `Message` | Missing `omitempty` — always serialized as empty string |
| notification.go | 101 | `FlowID` | Missing `omitempty` |
| notification.go | 111 | `Availability` | Missing `omitempty` — serialized as 0 when absent |
| calls.go | 306 | `CallStatusUpdate.MessagingProduct` | Missing JSON tag entirely |

---

## 5. Code Structure & DRY Violations

### 5.1 `handleOne` in message.go: 13 near-identical switch branches

**File:** `message.go:564-703`
**Effort:** LOW, high payoff

Every message type branch follows: nil-check handler → build request → call Handle → wrap error. A generic dispatch helper with a type-switch table would reduce ~140 lines to ~40.

### 5.2 13 identical switch cases in BusinessNotificationHandler

**File:** `business.go:398-521` (Handle), `business.go:615-643` (IsEventHandlerImplemented)
**Effort:** LOW

Same pattern repeated. Could use a registry map keyed by field name.

### 5.3 SMB + UserPreferences triple structural redundancy

**Files:** `smb.go`, `userpreferences.go`
**Effort:** MEDIUM

`SMBMessageEchoesHandler`, `SMBAppStateSyncsHandler`, and `UserPreferencesHandler` are structurally identical (handler, fallback, errorHandler fields + same method set). Consider extracting a generic `SingleDispatchHandler[T]` base.

### 5.4 `callbacks/` Client and BaseClient duplicate request-building logic

**File:** `callbacks/callbacks.go:284,314` vs `181,205`
**Effort:** LOW

The high-level `Client` methods copy the `BaseClient` methods verbatim, only omitting the `config` parameter. Have `Client` delegate to `BaseClient`.

### 5.5 `ClientErrorRateDetails` and `EndpointErrorRateDetails` identical

**File:** `flows.go:57-69`
**Effort:** LOW

Both structs have exactly the same fields. Merge into one type or use a shared embedded struct.

---

## 6. Missing Functionality & Improvement Opportunities

### 6.1 `callbacks/` package has zero tests

**File:** `callbacks/callbacks.go` (335 lines of untested code)
**Priority:** HIGH

A package that constructs HTTP requests with 6 request types, complex endpoint routing, and auth has no tests. At minimum: MockServer tests for `BaseClient.Send`, `SetAlternativeCallback`, `DeleteAlternativeCallback`.

### 6.2 No `CloseIdleConnections` on any webhooks client

**Priority:** LOW

Unlike the `calls` domain client (which now exposes it), no webhook-related client exposes `CloseIdleConnections`. The `callbacks.Client` and `callbacks.BaseClient` should expose it.

### 6.3 No panic recovery in Listener middleware chain

**File:** `listener.go` (documented as `TestDefect002`)
**Priority:** MEDIUM

A panic in middleware brings down the server. Add `recover()` around the middleware chain invocation.

### 6.4 `BaseClient.Send` in callbacks doesn't validate RequestType

**File:** `callbacks/callbacks.go:108`
**Priority:** LOW

If `request.RequestType` is zero-value, the switch falls through without setting `method`. Add a default case returning an error.

### 6.5 `html.EscapeString` test documents the bug but doesn't fix it

**File:** `listener_test.go:663`
**Priority:** MEDIUM

There's a test `TestListener_HandleSubscriptionVerification_HTMLEscapesChallenge` that verifies the current (buggy) behavior — it asserts that HTML entities are escaped. This test should be updated to expect verbatim echo.

---

## 7. Test Strategy Assessment & Recommendations

### Current state

- **~5,900 lines of test code** across 6 test files + 1 router test
- Main pattern: `httptest.Server` + inline JSON payload + boolean `handled` flag + handler closure
- Secondary pattern: direct struct creation + boolean flag + method call (in dedicated test files)

### Problems

1. **61 near-identical fallback tests** consuming ~1,500 lines. Nine domains x 6–8 sub-tests each, all following the same pattern: NoSubHandler_Silent200, NoSubHandler_GeneralFallbackFires, DedicatedHandlerFires, SubFallbackFires, NoSubFallback_Silent200, OnFallbackPropagates. These could be reduced to a single table-driven helper.

2. **Inline JSON payloads are verbose.** Each test inlines a 20–40 line JSON string. Replace with shared test fixtures (constants or `testdata/` files).

3. **Boolean `handled` flags** in closures make tests fragile — if the handler is called twice, the flag is still true. Use a `[]string` call log or `atomic.Int32` counter instead.

4. **No error propagation tests** for most domain handlers. Only messages and history test that handler errors propagate. Business, flows, groups, calls, SMB, and userprefs don't test the error path.

5. **No mock-based unit tests.** The only testing approach is integration through `httptest.Server`. Domain handler logic (dispatch, error handling, fallback) could be tested at the unit level with mock `EventHandler` implementations.

6. **`callbacks/` has zero tests.**

### Recommended test strategy

**Layer 1 — Unit tests (new):** Test each domain handler's `Handle`, `HandleError`, `Fallback`, and `IsEventHandlerImplemented` methods directly. Use hand-rolled mock `EventHandler` implementations (following the `Func` adapter pattern already in the codebase). No HTTP server needed.

**Layer 2 — Integration tests (improve existing):** Keep `httptest.Server` tests but:

- Extract shared JSON fixtures to `testdata/` files
- Replace 61 fallback tests with a single table-driven helper
- Replace `bool handled` flags with call-log slices

**Layer 3 — Router tests (existing):** Already well-structured. Add nil-listener test and concurrent access test.

**Layer 4 — callbacks/ tests (new):** Follow the `calls/calls_test.go` pattern — `internal/test.MockServer` for HTTP integration, `mocks/http.MockSender` for unit tests at the `Sender` boundary.

### Estimated impact

| Change | Lines saved | Coverage gain |
|--------|-------------|---------------|
| Table-drive fallback tests | ~1,300 | None (same coverage) |
| Shared JSON fixtures | ~800 | None |
| Table-drive handler dispatch tests | ~400 (net add) | +15% branch coverage |
| callbacks/ tests | ~500 (net add) | New package covered |
| Unit tests for error paths | ~300 (net add) | +10% branch coverage |

---

## 8. Summary by Severity

### Must fix (bugs)

| # | Issue | File |
|---|-------|------|
| 1.1 | `html.EscapeString` breaks verification | listener.go:218 |
| 1.2 | `SetSignatureVerifier(nil)` panics | listener.go:191 |
| 1.3 | `Contact.SenderInfo()` nil-deref | webhooks.go:392 |
| 1.4 | `handleSystem` nil Identity | message.go:735 |

### Should fix (dead code / incoherence)

| # | Issue | File |
|---|-------|------|
| 1.5 | BusinessNotificationHandler.calls dead code | business.go |
| 1.6 | Error→504 vs 500 mismatch | handler.go:305 |
| 2.5 | Duplicate OnHandler / OnEcho / OnSync | smb.go, userpreferences.go |
| 3.1 | OnMessageErrors→Unknown semantic | status.go:113 |
| 3.3 | ErrorInfo.Error() naming | notification.go:148 |
| 5.3 | SMB+UserPrefs triple redundancy | smb.go, userpreferences.go |

### Nice to fix (cleanup / improvement)

| # | Issue | File |
|---|-------|------|
| 2.1 | Duplicate `implemented` entries | handler.go |
| 2.2 | HistorySyncContext unused | history.go |
| 2.4 | 14 unused MessageType constants | message.go |
| 3.2 | Naming inconsistency | business.go:664 |
| 5.1 | handleOne repetition | message.go:564 |
| 5.2 | Business switch repetition | business.go:398 |
| 5.4 | callbacks Client/BaseClient DRY | callbacks/callbacks.go |
| 6.3 | No panic recovery in middleware | listener.go |
| 6.1 | callbacks/ zero tests | callbacks/ |
