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

## Feature Categories

This SDK covers two distinct feature areas. Know which one you're working on before starting.

### 1. API Calling (domain packages)

Packages under root (`calls/`, `message/`, `media/`, `groups/`, `qrcode/`, `phonenumber/`, `auth/`, `business/`, `flow/`, `user/`, `templates/`, `settings/`, `uploads/`, `business/analytics/`, `conversation/automation/`) send requests TO the WhatsApp Cloud API. Each package wraps an HTTP endpoint: `POST /<PhoneNumberID>/messages`, `GET /<PhoneNumberID>/call_permissions`, etc.

Pattern: `NewClient(conf)` → build request → `BaseClient.Send()` → decode response. Tests use `internal/test.MockServer` for HTTP integration and `mocks/http.MockSender` for unit tests.

### 2. Webhook Handling (`webhooks/`)

The webhooks package receives inbound notifications FROM WhatsApp. It does NOT make API calls — it parses incoming JSON, routes to typed handlers, and lets user code react. Think of it as an event bus for WhatsApp events.

The package has its own sub-packages: `webhooks/callbacks/` (managing alternate callback URLs — an API-calling package) and `webhooks/router/` (HTTP mux integration).

---

## Adding an API Calling Feature — Checklist

When adding a new API endpoint to a domain package (e.g., enabling call recording on `POST /calls`, sending a new message type on `POST /messages`), use this checklist. Every item is required.

### Phase 1: Routing

- **[ ] Add a RequestType.** If this is a new endpoint (not a new field on an existing one), add a constant to `pkg/http/request_type.go` (e.g., `RequestTypeUpdateCallStatus`). The `RequestType` controls method, endpoint path, and body construction in `BaseClient.Send()`.
- **[ ] Wire the RequestType in BaseClient.Send().** In the domain package's `BaseClient.Send()` method, add a `case` for the new `RequestType`. Set `method` (GET/POST/DELETE), `endpoint`, any `queryParams`, and construct the `message *BaseRequest` if it carries a body.
- **[ ] Add new fields to BaseRequest.** If the API endpoint accepts new optional fields (e.g., `recording` on connect/accept), add them to the wire-format `BaseRequest` struct with the correct JSON tag.
- **[ ] Pass through from Request.** The internal `Request` struct carries fields from the high-level API to `BaseClient.Send()`. Any new field needed for request construction goes here (with `json:"-"` — it's never serialized directly).
- **[ ] Pass through from the public request type.** The user-facing request struct (e.g., `CallUpdateStatusRequest`) is what callers construct. Add the field here with the correct JSON tag, then pass it through to `Request` in the `Client` method.

### Phase 2: Types

- **[ ] Define the public request struct (or add fields to an existing one).** This is the user-facing type. Every field that the WhatsApp API accepts must be represented. Use `omitempty` on optional fields to avoid sending zero values.
- **[ ] Define the response struct (or add fields).** If the API returns new fields in its response, add them to the response type. Match JSON tags exactly to the API response.
- **[ ] Add convenience constructors/helpers.** If the new feature is enabled by setting a sub-object (like `Recording`), add a `Set*` method (e.g., `SetRecording(r *Recording)`) rather than making callers construct the nested struct manually.
- **[ ] Add any new constants.** Status enums, action types, language codes — define them as typed string constants (e.g., `RecordingStatus`, `RecordingEnabled`).

### Phase 3: Client methods

- **[ ] High-level Client method.** The `Client` struct's method is the public API. It constructs the internal `Request`, calls `c.Send()`, and converts the response. If the new feature is a field on an existing endpoint (like `recording` on `UpdateCallStatus`), update the existing method to pass the field through — no new method needed.
- **[ ] BaseClient method (multi-tenant).** If adding a new endpoint (not just a field), add a `BaseClient` method for the multi-tenant path. The `Client` method can delegate to it (DRY), passing `c.config`.
- **[ ] Wire CloseIdleConnections.** The `Client` should expose `CloseIdleConnections()` delegating to `c.sender.CloseIdleConnections()`.

### Phase 4: Tests

- **[ ] JSON round-trip tests.** Every new request/response struct must have `test.AssertJSONRoundTrip` covering all variants (with and without optional fields).
- **[ ] JSON marshal test.** Verify the struct produces the exact wire format using `test.AssertJSONMarshal`.
- **[ ] Constructor/helper tests.** Test that `Set*` helpers and convenience constructors populate fields correctly.
- **[ ] MockServer HTTP integration test.** Use `internal/test.MockServer` to verify: HTTP method, URL path, query parameters, request body fields, response decoding. This is the most important test — it proves the full chain from `Client` → `BaseClient.Send` → HTTP → response works.
- **[ ] MockSender unit test.** Use `mocks/http.MockSender` for error propagation tests: verify that sender errors are wrapped and returned correctly, and that successful responses are decoded properly (use `DoAndReturn` to simulate responses).
- **[ ] Error response test.** Verify the client handles non-200 status codes and returns errors with the correct API error code.

### Phase 5: Documentation

- **[ ] Package doc.go.** Update the package-level comment (in the domain package's main file) to mention the new feature. Show a usage example in the comment — this is what `go doc` displays.
- **[ ] Type/function doc comments.** Every exported type, function, constant, and method gets a doc comment. Follow the existing style: start with the type/method name, describe what it does, mention required vs optional fields, note any WhatsApp API constraints (max length, required when status is X, etc.).
- **[ ] README update.** If the domain package has a README, add a usage example for the new feature.
- **[ ] Document associated webhook events.** If the API call triggers webhook events (e.g., enabling recording on connect causes a `call_recording_available` webhook later), document this in BOTH the API package doc comments AND the webhook handler doc comments. Cross-reference them: the API doc should say "after the call ends, you receive a call_recording_available webhook"; the webhook doc should say "sent when a call with recording enabled finishes". This bidirectional linking is critical for users to understand the full flow.
- **[ ] Document limits and constraints.** If the WhatsApp API imposes limits (e.g., purpose max 250 characters, announcement language must be from a specific set, 100 connected calls per 24h), document them in the type/function comments. Users should not need to read Facebook's docs for parameter constraints.

### Phase 6: Final verification

- **[ ] `make all` passes.** Zero lint issues, zero test failures, race detector clean.
- **[ ] Examples compile.** Both `_examples/message` and `_examples/webhooks` must build and run.
- **[ ] Associated webhook handler exists.** If the API call produces webhook events, verify the webhook handler (see Webhook Handler Checklist above) is implemented and tested. The API call and its webhook are two halves of the same feature — both must work.

---

## Adding a Webhook Handler — Checklist

When a new WhatsApp webhook event type needs handling (e.g., call recording, message edits, template updates), use this checklist. Every item is required; skip nothing.

### Phase 1: Categorization

- **[ ] Identify the ChangeField.** Find the exact `field` string WhatsApp sends in the webhook JSON (e.g., `"calls"`, `"messages"`, `"account_review_update"`). Add it as a `ChangeField` constant in `field.go` if it doesn't exist.
- **[ ] Map to a ChangeFieldCategory.** Add the field→category mapping in `GetChangeFieldCategory()` in `field.go`. If no suitable category exists, add a new `ChangeFieldCategory*` constant and a new handler field on the `Handler` struct.
- **[ ] Register as implemented.** Add the `ChangeField` to the `implemented` slice in `NewHandler()` in `handler.go`. This ensures `dispatchEvent` routes it instead of sending it to the general fallback. If a new category was added, wire the handler field in `changeHandler()` in `event.go`.
- **[ ] Wire defaults.** If a new handler struct was added, initialize it in `NewHandler()`, propagate `OnError`/`OnFallback`, and add accessor method (`Handler.Calls()`, `Handler.Flows()`, etc.).

### Phase 2: Type definitions

- **[ ] Define the payload struct.** The new type goes in the relevant domain file (e.g., `CallRecording` lives in `calls.go` inside `webhooks/`). Every field must have the exact JSON tag matching the WhatsApp webhook payload — copy field names verbatim from the API documentation.
- **[ ] Add the field to the carrier struct.** The payload needs a home: add a field to the existing carrier type. For a new call event, that's `Call` in `calls.go`. For a message sub-type, that's `Message` in `handler.go` (the types file). Use `omitempty` on optional fields.
- **[ ] Reuse existing types where possible.** If the payload carries a media asset, use `media.Info`. If it carries errors, use the existing `ErrorInfo`/`werrors.Error` chain. Don't create duplicate types for identical wire formats.
- **[ ] Use the `Interface` + `Func` adapter pattern.** Every handler interface (e.g., `CallsEventHandler[T]`) needs a corresponding `Func` adapter (e.g., `CallsEventHandlerFunc[T]`) and a type alias for each event sub-type (e.g., `CallRecordingAvailableHandler = CallsEventHandler[Call]`). This is non-negotiable — it's the extension mechanism for user code.

### Phase 3: Handler registration

- **[ ] Add handler field.** Add an unexported field to the domain handler struct (e.g., `recordingAvailable CallsEventHandler[Call]`).
- **[ ] Add `On*` setter.** Add an exported method (e.g., `OnCallRecordingAvailable(h CallRecordingAvailableHandler)`) that sets the field. Follow the naming convention of existing setters.
- **[ ] Add dispatch case.** In the domain handler's `Handle` method, add a `case` matching the event identifier (string from the webhook JSON). Follow the pattern: nil-check handler → build request → call Handle → wrap error through HandleError → continue.
- **[ ] Add CanHandleEvent case.** In the domain handler's `CanHandleEvent` method, add a matching case that returns `true` when the handler field is non-nil. This is what `HandleEvent` checks before calling `Handle`.
- **[ ] Add convenience method on `*Handler`.** Add an exported method (e.g., `func (handler *Handler) OnCallRecordingAvailable(...)`) that delegates to the domain handler's setter. Follow the existing naming convention.

### Phase 4: Tests

- **[ ] JSON round-trip test.** Verify the new payload struct marshals and unmarshals symmetrically using `test.AssertJSONRoundTrip`. Include all variants (e.g., with and without optional fields).
- **[ ] JSON marshal test.** Verify the struct produces the exact JSON format expected by the WhatsApp API using `test.AssertJSONMarshal`.
- **[ ] Exact API payload parsing test.** Copy the example JSON from WhatsApp's documentation verbatim into a test. Parse it with `test.AssertJSONUnmarshal` into the notification chain (`Notification` → `Entry` → `Change` → `Value` → your type). Assert every field value.
- **[ ] Handler dispatch test.** Register a handler via the `On*` convenience method, feed the parsed `NotificationEvent` through `HandleNotificationEvent`, and assert the handler was called with the correct payload.
- **[ ] HTTP integration test (API calling only).** For domain packages that send requests, add a `MockServer` test verifying the new field appears in the HTTP request body.
- **[ ] Nil/error handling test.** Verify the code does not panic when: the Value is nil, the carrier slice is nil or empty, the handler is nil, the handler returns an error (error propagates correctly), the fallback is nil (silent skip).

### Phase 5: Documentation

- **[ ] Package doc.go.** Update the comment in the relevant domain file (e.g., `calls.go`'s "Call types and CallsHandler..." header comment) to list the new event type. This comment is the package-level reference for supported events.
- **[ ] Type/function comments.** Every exported type, function, and method must have a Go doc comment. Follow the existing style: start with the name, describe what it does, mention the payload type, note any non-obvious behavior.
- **[ ] Method doc comments.** Document which WhatsApp event triggers this handler, what the payload carries, and any constraints (e.g., "purpose max 250 characters", "URL valid for 5 minutes").
- **[ ] README update.** If the feature is user-facing, add a brief entry in the relevant README (webhooks or domain package) showing a usage example.
- **[ ] Trigger documentation.** In a doc comment on the handler type or method, document:
  - What WhatsApp action triggers this webhook (e.g., "sent when a recorded call finishes post-processing")
  - Timing guarantees (e.g., "typically delivered within 1 minute of call end")
  - Retention/expiry (e.g., "recording available for 7 days")
  - Non-obvious behavior (e.g., "also fires for user-initiated calls if recording was enabled at accept time")
  - Relationship to other webhooks (e.g., "independent of call_transcription_available — enabling both sends both webhooks")
  - Error scenarios (e.g., "not sent if recording fails; no error webhook is delivered")

### Phase 6: Final verification

- **[ ] `make all` passes.** Zero lint issues, zero test failures, race detector clean.
- **[ ] Examples compile.** Both `_examples/message` and `_examples/webhooks` must build and run.
- **[ ] No new linter directives without justification.** If you add a `//nolint`, explain why in a comment.

### Module & Version

Module path: `github.com/piusalfred/whatsapp`. Go 1.26.5. Minimal dependencies (go-cmp, uber/mock, x/sync). API version floor: `v20.0` (checked by `whatsapp.IsCorrectAPIVersion`).
