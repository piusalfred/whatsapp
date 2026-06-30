# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Lint, and Test Commands

```bash
make all       # Format (license + goimports + golines + lint fix), generate mocks, run tests with race detector
make fmt       # Add license headers, go mod tidy, go fix, golangci-lint fmt, lint fix, regenerate mocks
make check     # CI-safe: verify license headers and lint without modifying code
make test      # Run all tests with -race, -coverpkg ./..., -parallel=4; output piped through tparse
make mocks     # Regenerate all mock files (mockgen from source)
make update    # go get -u ./... + go mod tidy
make upgrade   # Match go.mod go/toolchain directives to installed Go version, then update deps
make help      # List all make targets
```

**Running a single test:**
```bash
go test -race -run TestName ./path/to/package/
```

**Running tests for a single package:**
```bash
go test -race ./webhooks/
```

**Lint a single file/package (after changes):**
```bash
gotools exec golangci-lint run ./path/to/package/
```

## Architecture

### Layered Structure

```
User code
   ↓
api.Client / message.Client / groups.Client  ...   ← domain packages (single-tenant)
   ↓
whttp.BaseClient[T]  →  CoreClient[T]              ← transport layer (pkg/http)
   ↓
net/http
```

### Key Design Patterns

**Two clients per domain package.** Every domain package exposes:
- `Client` — holds a fixed `*config.Config`; for single-tenant services
- `BaseClient` — accepts `*config.Config` per call; for multi-tenant / credential-rotation workloads

Both embed `whttp.BaseClient[T]` which wraps a typed `Sender[T]` interface for HTTP dispatch. The `Sender[T]` interface (`pkg/http/http.go`) is the mockable boundary — tests inject `mocks/http.MockSender[T]` via `SetBaseClient(sender)`.

**Domain model / wire format separation.** Domain types (e.g., `message.Message`) are separate from wire-format types (`message.BaseRequest`). The client maps between them in `sendMessage()` — this keeps builder methods and metadata off the JSON wire shape.

**HTTP layer is generic and typed.** `pkg/http` provides:
- `CoreClient[T]` — typed HTTP client with interceptors, middleware, limits
- `Sender[T]` — interface (`Send(ctx, *Request[T], ResponseDecoder) error`) that domain code calls and tests mock
- `Middleware[T]` — wraps `SenderFunc[T]`; applied inside-out so `middlewares[0]` runs outermost
- `Request[T]` — typed request carrying method, URL, headers, body, auth, metadata
- Request building: `RequestBuilder` (chained, preferred for many options) or `MakeRequest` (functional options, for simple cases)
- Always alias `pkg/http` as `whttp` to avoid shadowing `net/http`:
  ```go
  import whttp "github.com/piusalfred/whatsapp/pkg/http"
  ```

**Webhooks: listener → handler → sub-handler cascade.** `webhooks/webhooks.go` defines `Listener` (HTTP entry point, signature validation), which delegates to `Handler` (`webhooks/handler.go`). The `Handler` routes by `change.Field` to typed sub-handlers: `MessagesHandler`, `FlowNotificationHandler`, `BusinessNotificationHandler`, `GroupManagementHandler`, `HistoryHandler`, `CallsHandler`. Each sub-handler has its own typed callbacks and fallback. Registration is not goroutine-safe — register all handlers at startup before calling `HandleNotification`.

**Unified api.Client.** `api/` composes all 15 domain sub-clients behind one struct. Each sub-client is lazily initialized via `sync.Once`. Middleware setters (`SetCallsMiddlewares`, etc.) must complete before any API calls — they trigger `sync.Once` then call `SetMiddlewares` which is not goroutine-safe with respect to in-flight `Send`.

**Config flow.** `config.Config` has no defaults — every required field (`BaseURL`, `APIVersion`, `AccessToken`, `PhoneNumberID`) must be set. Use `config.Validate()` or `config.ReadValidate()` to fail fast at startup. `config.Reader` / `ReaderFunc` supports dynamic config sources. API version must be ≥ v20.0.

### Package Map

| Layer | Package | Role |
|-------|---------|------|
| Transport | `pkg/http` | Generic typed HTTP client, middleware, interceptors, encoding/decoding |
| Crypto | `pkg/crypto` | HMAC-based appsecret_proof generation |
| Errors | `pkg/errors` | WhatsApp API error types |
| Types | `pkg/types` | Shared types (`Metadata`) |
| Config | `config` | Configuration struct, validation, Reader interface |
| Foundation | `whatsapp` (root) | API version parsing, base URL constant, sentinel errors |
| Outbound APIs | `message/`, `groups/`, `media/`, `auth/`, `qrcode/`, `flow/`, `calls/`, `business/`, `user/`, `phonenumber/`, `settings/`, `uploads/`, `templates/`, `conversation/automation/`, `business/analytics/` | Domain-specific API clients |
| Unified | `api/` | Composes all 15 domain clients; single-tenant `Client` + multi-tenant `BaseClient` |
| Webhooks | `webhooks/` | Inbound listener + typed handler dispatch |
| Webhook callbacks | `webhooks/callbacks/` | Outbound client for alternate callback URL management |
| Test support | `internal/test/` | Assertion helpers (`AssertNoError`, `AssertJSONRoundTrip`, etc.) and `MockServer` |
| Mocks | `mocks/` | Generated gomock mocks (`mocks/http/MockSender[T]`, `mocks/webhooks/`, etc.) |

### Mock Injection Pattern

Generated mocks use `go.uber.org/mock` (uber-go fork of gomock). The standard injection pattern:
```go
ctrl := gomock.NewController(t)
mockSender := mockhttp.NewMockSender[message.BaseRequest](ctrl)
mockSender.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

client := message.NewClient(conf)
client.SetBaseClient(mockSender)
```

### Test Helpers (`internal/test/`)

- `AssertNoError`, `AssertError`, `AssertErrorIs`, `AssertErrorAs` — error assertions
- `AssertJSONMarshal` — marshal + cmp.Diff against expected JSON
- `AssertJSONUnmarshal` — unmarshal JSON string into typed value
- `AssertJSONRoundTrip` — marshal → unmarshal → re-marshal symmetry check
- `MockServer` — httptest.Server wrapper that records all requests; configure response via `MockBehavior`

Tests use `github.com/google/go-cmp` for comparisons.

### Important Gotchas

- **Webhook error handling:** Only return errors from webhook handlers for transient failures. WhatsApp retries non-200 responses for up to 7 days. For permanent failures (unsupported message types, invalid payloads), log and return nil.
- **Webhook context lifetime:** The `ctx` passed to webhook handlers is cancelled after the HTTP response is written. Spawn background work with `context.Background()`, not the webhook ctx.
- **Zero-value clients panic:** Always construct with `NewClient` or `NewBaseClient`. A zero-value `Client` has a nil sender.
- **Middleware registration order:** Call `SetMiddlewares` / `SetSender` at startup before any goroutines start calling `Send`. These methods are not goroutine-safe.
- **`sync.Once` fragility:** If an `api.BaseClient` sub-client init panics inside `sync.Once`, that sub-client is permanently nil — future calls also panic. Validate options before creating.
- **Webhook payload limit:** 4 MB (`MaxPayloadBytes`). WhatsApp's documented limit is 3 MB.
- **Unrecognized webhook fields:** Silently acknowledged with 200 by default. Register an `OnFallback` handler to log or handle future notification types.

### Tools (Go Module Proxies)

The `tools/` directory contains isolated `go.mod` files for development tools: `addlicense`, `golangci-lint`, `mockgen`, `tparse`. These are invoked through a `gotools` wrapper (configured via `.gotools.env`). Run `gotools exec <tool>` to invoke any tool at its pinned version.
