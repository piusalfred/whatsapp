# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **webhooks**: Renamed `UnknownFieldHandler` to `FallbackHandler` (breaking).
- **webhooks**: Renamed `UnknownHandlerFunc` to `FallbackHandlerFunc` (breaking).
- **webhooks**: Removed `SetFlow*Handler` methods; use `OnFlow*` methods instead (breaking).
- **webhooks**: Renamed `SetGeneralFallbackHandler` to `OnFallback` (breaking).
- **webhooks**: Sub-handler `Handle` signatures unified to `(ctx, NotificationEntry, change)` (breaking).

### Added

- **webhooks**: `handleError` and `executeFallback` helpers on all sub-handlers to eliminate boilerplate.
- **webhooks**: `HistoryHandler` wired as a proper sub-handler with `OnHistorySync`, `OnHistoryMediaMessages`, `OnHistoryFallback`.
- **webhooks**: Media messages from history webhooks are now routed through `HistoryHandler` instead of the general `MessagesHandler`.
- **webhooks**: `newMessageInfo` helper for consistent `MessageInfo` construction.

### Fixed

- **webhooks**: Flow handler errors now route through `ErrorHandler` (was bypassed).
- **webhooks**: Business handler errors now route through `ErrorHandler`.
- **webhooks**: Group handler errors now route through `ErrorHandler`.
