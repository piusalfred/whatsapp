# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Session Rules

- **Never** add Claude co-author signatures (`Co-Authored-By: Claude ...`) to commits.
- Only commit when explicitly asked or given permission. If a commit seems necessary, ask first — do not commit unprompted.
- After making changes, verify that all examples under `_examples/` compile and run correctly. If an example fails to build, fix the issue before considering the work done.

## Build & Test Commands

```bash
make all          # format, lint, generate mocks, run tests with race detector
make fmt          # go mod tidy, go fix, addlicense, golangci-lint run --fix, regenerate mocks
make test         # go test -race -json -parallel=4 ./... (pipe through tparse)
make webhook-test # run only webhooks package tests
make coverage     # HTML coverage report with -coverpkg ./...
make bench        # run all benchmarks
make build        # compile all packages (go build -v ./...)
make mocks        # regenerate mocks only
make check        # pre-commit sanity checks (license, fmt, lint, mod-tidy, build, vuln)
make check-all    # CI checks (check + check-test)
```

To run a single test or specific package tests:

```bash
go test -race -v -run TestName ./path/to/package/...
```

Tools (`golangci-lint`, `addlicense`, `tparse`, `mockgen`) are invoked via `gotools exec` — a tool dependency manager configured in the project.

After making changes, run `make all` as a sanity check — it formats, lints, regenerates mocks, and runs tests with the race detector in one pass.

## Architecture

### Client Pattern (Domain Packages)

Every domain package (`message`, `media`, `groups`, `qrcode`, `phonenumber`, `auth`, `business`, `calls`, `flow`, `user`, `templates`, `settings`, `uploads`, `business/analytics`, `conversation/automation`) follows the same pattern:

1. **`NewClient(conf *config.Config)`** — single-tenant: holds a fixed `*config.Config`. Methods accept `ctx` and the domain payload, constructing the request internally.
2. **`NewBaseClient()`** — multi-tenant: accepts per-call config via `WithSenderConfig`. Ideal for dynamic credential rotation or multi-tenant workloads.

Both embed `http.BaseClient[RequestType]` internally. The `BaseClient` holds a `Sender[T]` interface; tests inject `mocks/http.MockSender[T]` via `SetSender()`.

### HTTP Transport Layer (`pkg/http`)

The HTTP layer is generic over request body type `T`:

- **`Request[T]`** — carries method, URL, headers, query params, body (JSON `Message *T`, multipart `Form`, or raw `BodyReader`), auth config, metadata.
- **`CoreClient[T]`** — the concrete `Sender[T]` implementation. Wraps `*http.Client` with request/response interceptors (`RequestInterceptorFunc`, `ResponseInterceptorFunc`) that snapshot and restore bodies so interceptors can read freely without consuming the stream. Supports functional options (`WithSenderHTTPClient`, `WithSenderTimeout`, etc.) and `Middleware[T]` wrapping.
- **`BaseClient[T]`** — generic building block embedded by domain clients. Delegates to a `Sender[T]`; provides `SetSender()` for mock injection and `SetMiddlewares()` for middleware composition.
- **`RequestBuilder`** — fluent API to build `Request[T]` values. Supports `Auth(authConfig)` to bridge `config.Config.AuthConfig()` to the transport layer.

Middleware is composed inside-out: `middlewares[0]` runs outermost (first on the way in, last on the way out). Use `WrapMiddlewares` or `WrapMiddlewareSender` for composition.

### Webhooks (`webhooks`)

The webhook system has a three-layer architecture:

1. **`Listener`** — HTTP entry point. Handles subscription verification (GET) and notification dispatch (POST). Validates `X-Hub-Signature-256` headers, decodes JSON, delegates to `NotificationHandler`. Supports `Middleware` wrapping (inside-out, like the HTTP layer). Config is resolved per-request via `ConfigReader(*http.Request)` to enable multi-tenancy.
2. **`Handler`** — central dispatch unit. Registers callbacks for every WhatsApp webhook event type via `On<T>Event` methods. Dispatches flattened `NotificationEvent` values to the correct domain handler based on `ChangeField`. Supports `Sequential` (default, first error stops) and `Parallel` (concurrent via `errgroup`) dispatch modes.
3. **Domain handlers** (`MessagesHandler`, `BusinessNotificationHandler`, `FlowNotificationHandler`, `GroupManagementHandler`, `CallsHandler`, `HistoryHandler`, `SMBMessageEchoesHandler`, `SMBAppStateSyncsHandler`, `UserPreferencesHandler`) — each owns its own typed dispatch logic, error handler cascade, and fallback. Sub-handlers (e.g., `MessagesHandler` has separate handlers for text, image, interactive, etc.) follow the same `On*` + `Func` adapter pattern.

Error handling cascade: domain handler → global error handler → fallback handler. Error handlers return `nil` to continue processing or a non-nil error to abort (HTTP 500, triggers WhatsApp retry).

### Configuration (`config`)

`config.Config` holds all API credentials (BaseURL, APIVersion, AccessToken, PhoneNumberID, BusinessAccountID, AppSecret, AppID, SystemUserID, SecureRequests, DebugLogLevel). `Validate()` checks required fields and API version minimum. `ReadValidate()` combines a `Reader` (dynamic config source) with validation.

### Type System

- **`Interface` + `Func` adapter pattern**: Every interface (`Sender`, `EventHandler`, `ErrorHandler`, etc.) has a corresponding `Func` type adapter, analogous to `http.HandlerFunc`. This is the primary extensibility mechanism.
- **`pkg/types`**: Shared types (`Metadata`, phone number types).
- **`pkg/errors`**: Shared error types used across domain packages.
- **`pkg/crypto`**: Crypto utilities used by the HTTP layer for request signing.

### Licensing

Every `.go` file must have the MIT license header. `make fmt` runs `addlicense` to add headers; `check-license` verifies them. Excludes: `tools/`, `_examples/`, `extras/`.

### Mock Generation

Mocks use `go.uber.org/mock` and are generated via `mockgen`:

- `mockgen -destination=./mocks/config/config_mock.go -package=config -source=./config/config.go`
- `mockgen -destination=./mocks/http/mock_http.go -package=http -source=./pkg/http/core_client.go`

Only these two mocks are auto-generated; other mocks were removed. If a new interface needs mocking, add the generation command to the `mocks` Make target.

### Module & Version

Module path: `github.com/piusalfred/whatsapp`. Go 1.26.5. Minimal dependencies (go-cmp, uber/mock, x/sync). API version floor: `v20.0` (checked by `whatsapp.IsCorrectAPIVersion`).
