# Agent 3 – Correctness, Testability & Security Audit

**Verdict & Grade: B** (80/100)

Core dispatch logic is sound. The handler registration pattern is thread-safe for reads. Major gaps exist in nil-safety for sub-handler pointers, missing test coverage for several code paths, and no concurrency/race tests for handler dispatch.

---

## Finding 1 – Nil sub-handler dereference risk (Major)

**File:** `handler.go:56-61`, `handler.go:228-242`

```go
case ChangeFieldCategoryFlows:
    return handler.handleFlowsChange(ctx, ne, change)  // → handler.flows.Handle(...)
case ChangeFieldCategoryBusiness:
    return handler.business.Handle(ctx, ne, change)
case ChangeFieldCategoryGroups:
    return handler.groups.Handle(ctx, ne, change)
case ChangeFieldCategoryHistory:
    return handler.history.Handle(ctx, ne, change)
```

`handler.flows`, `handler.business`, `handler.groups`, `handler.history`, and `handler.messages` are all `nil` by default (from `NewHandler()`). If someone adds a flow field to the `implemented` list or the `changeFieldMap` dispatches to a category where the sub-handler is nil, it panics.

Defense: the `changeFieldHandlers.Check()` call precedes dispatch and routes to `handler.fallback` for unimplemented fields. But if a field is in the "implemented" set and its sub-handler is nil, there's no defense.

**Failing test:**
```go
func TestHandler_NilSubHandler_Panics(t *testing.T) {
    h := webhooks.NewHandler()
    h.OnBusinessAlertNotification(webhooks.BusinessEventHandlerFunc[...](func(...) error { return nil }))
    // handler.business is nil — if a business field arrives, it panics
}
```

**Recommendation:** `NewHandler()` should initialize sub-handlers to no-op instances, or the dispatch should guard against nil sub-handlers.

---

## Finding 2 – `handleSMBMessageEchoesChange` returns nil after routing error through handler (Correct)

**File:** `handler.go:290-312`

The function iterates messages, routes errors through `handler.handleError`, returns nil on the first fatal error. Correct behavior — confirms `handleError` semantics are properly applied.

---

## Finding 3 – Test Coverage Gaps (Medium)

**Missing tests:**
- History handler (now covered ✓ — added `history_test.go`)
- Error handler routing for flow events (now implicitly tested via `handleError` pass-through)
- `handleSMBAppStateSyncChange` — no dedicated test
- `OnSMBMessageEcho` — no test
- `OnMessageErrors` and `OnUnsupportedMessage` — only via unsupported path
- `OnMessageEdit` — no test
- `OnRevokeMessage` — no test
- `OnRequestWelcomeMessage` — no test
- `OnCustomerIDChangeMessage` — no test
- Media dispatch per-type (audio/video/image/document/sticker) — no unit test
- Concurrent access — no `-race` test for concurrent reads
- Context cancellation — `HandleNotification` checks `ctx.Done()` but no test verifies it

**Recommendation:** Add table-driven tests for all remaining `On*` registration paths.

---

## Finding 4 – No `-race` test in CI (Medium)

**File:** CI configuration (see Agent 6)

The test suite uses `t.Parallel()` extensively and the handlers are documented as safe for concurrent reads. But there's no race detector in CI. `go test -race` should be part of the standard test suite.

---

## Finding 5 – Input validation (Minor)

**File:** `handler.go:172-192`

`HandleNotification` iterates entries and changes. If `notification` were nil, it would panic. This is an internal function called by `Listener.HandleNotification`, which validates the payload earlier. Safe in practice, but a nil-guard would be defensive.

---

## Finding 6 – Error wrapping consistency (Clean)

All errors from user handlers are now routed through `handleError` (on both `Handler` and sub-handlers). Wrapping uses `fmt.Errorf("error handler: %w", ...)` consistently. The `%w` verb preserves original errors for `errors.Is`/`As` comparison.

---

## Finding 7 – No goroutine leaks (Clean)

`HandleNotification` is fully synchronous — no fire-and-forget goroutines, no channels. Each webhook is processed inline. This is correct for a webhook handler.

---

## Finding 8 – Security: TLS and secret handling (Minor)

**File:** `webhooks.go:354-372`

`ValidateSignature` uses HMAC-SHA256 to validate payload signatures. The implementation is correct. The `ConfigReader` pattern means secrets are fetched per-request rather than stored statically — good design. However, no connection-level TLS enforcement exists; that's the caller's responsibility.

---

## Prioritized Action List

| # | Finding | Severity | Effort |
|---|---|---|---|
| 1 | Nil sub-handler panic | Major | Medium |
| 2 | Missing test coverage (8+ handlers) | Medium | Medium |
| 3 | No race detector in test suite | Medium | Small |
| 4 | No concurrent-access test | Medium | Small |
