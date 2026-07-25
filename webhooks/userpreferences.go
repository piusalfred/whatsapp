//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// UserPreferencesHandler for the "user_preferences" webhook field.
// Handles marketing message opt-out/opt-in preference changes.

package webhooks

import (
	"context"
	"fmt"
)

// UserPreference represents a single user preference change.
// Each entry in value.UserPreferences carries one of these.
type UserPreference struct {
	WaID      string `json:"wa_id"`
	Detail    string `json:"detail"`
	Category  string `json:"category"` // always "marketing_messages"
	Value     string `json:"value"`    // can be "stop" or "resume"
	Timestamp string `json:"timestamp"`
}

// UserPreferenceHandler is the interface for handling a single user
// preference change. Each entry in value.UserPreferences is delivered
// individually.
type UserPreferenceHandler interface {
	Handle(ctx context.Context, nctx *MessageNotificationContext, pref *UserPreference) error
}

// UserPreferenceHandlerFunc adapts a bare function to the
// UserPreferenceHandler interface.
type UserPreferenceHandlerFunc func(ctx context.Context, nctx *MessageNotificationContext, pref *UserPreference) error

// Handle implements [UserPreferenceHandler] by calling the underlying function.
func (f UserPreferenceHandlerFunc) Handle(
	ctx context.Context, nctx *MessageNotificationContext, pref *UserPreference,
) error {
	return f(ctx, nctx, pref)
}

// UserPreferencesHandler handles the "user_preferences" webhook field.
//
// This webhook notifies of changes to a WhatsApp user's marketing message
// preferences — when a user stops or resumes marketing messages. Each
// preference entry includes the user's wa_id, the preference value
// ("stop" or "resume"), the category ("marketing_messages"), and a
// timestamp.
//
// Dispatch is intentionally simple: one handler for all preference
// entries. There is no per-value sub-dispatch (stop vs. resume).
//
// Usage:
//
//	uh := &UserPreferencesHandler{}
//	uh.OnChange(webhooks.UserPreferenceHandlerFunc(
//	    func(ctx context.Context, nctx *webhooks.MessageNotificationContext, p *webhooks.UserPreference) error {
//	        log.Printf("user %s: marketing_messages %s", p.WaID, p.Value)
//	        return nil
//	    },
//	))
type UserPreferencesHandler struct {
	// Handler receives every user preference change. When nil, entries fall
	// through to the [Fallback] method.
	handler UserPreferenceHandler

	// Fallback is called when Handler is nil. When nil, the event is silently
	// acknowledged (HTTP 200).
	fallback FallbackHandler

	// ErrorHandler is called when Handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	errorHandler ErrorHandler
}

// OnHandler sets the handler for all user preference changes.
func (uh *UserPreferencesHandler) OnHandler(h UserPreferenceHandler) {
	uh.handler = h
}

// OnError sets the error handler for this domain handler. When nil, errors
// bubble up to the general error handler configured on [Handler].
func (uh *UserPreferencesHandler) OnError(h ErrorHandler) {
	uh.errorHandler = h
}

// OnChange sets the handler for all user preference changes.
func (uh *UserPreferencesHandler) OnChange(h UserPreferenceHandler) {
	uh.handler = h
}

// OnFallback sets the catch-all handler when [Handler] is nil.
func (uh *UserPreferencesHandler) OnFallback(h FallbackHandler) {
	uh.fallback = h
}

func (uh *UserPreferencesHandler) HandleError(ctx context.Context, err error) error {
	return execErrorHandler(ctx, uh.errorHandler, err)
}

func (uh *UserPreferencesHandler) Fallback(ctx context.Context, event NotificationEvent) error {
	if uh.fallback == nil {
		return nil
	}
	if err := uh.fallback.Handle(ctx, event); err != nil {
		return fmt.Errorf("user preferences fallback: %w", err)
	}
	return nil
}

// Handle dispatches the user_preferences value to the handler.
//
//  1. If [Handler] is set, each entry in value.UserPreferences is passed
//     to it. Errors are routed through [ErrorHandler].
//  2. If [Handler] is nil, returns [ErrEventNotHandled].
//  3. If [Fallback] is also nil, the event is silently skipped (HTTP 200).
func (uh *UserPreferencesHandler) Handle(
	ctx context.Context,
	event NotificationEvent,
) error {
	if event.Value == nil {
		return nil
	}

	if uh.handler == nil {
		return ErrEventNotHandled
	}

	nctx := &MessageNotificationContext{
		EntryID:            event.EntryID,
		EntryTime:          event.Time,
		NotificationObject: event.Object,
		MessagingProduct:   event.Value.MessagingProduct,
		Contacts:           event.Value.Contacts,
		Metadata:           event.Value.Metadata,
	}

	for _, pref := range event.Value.UserPreferences {
		if pref == nil {
			continue
		}
		if err := uh.handler.Handle(ctx, nctx, pref); err != nil {
			return uh.HandleError(ctx, fmt.Errorf("user preferences: %w", err))
		}
	}

	return nil
}

// CanHandleEvent reports whether a handler is registered for user preference changes.
// Returns true when the handler is non-nil.
func (uh *UserPreferencesHandler) CanHandleEvent(_ NotificationEvent) bool {
	return uh.handler != nil
}

var _ EventHandler = (*UserPreferencesHandler)(nil)

// OnUserPreferencesUpdate sets the handler for user preference changes. Each
// preference entry (stop or resume marketing messages) is delivered to the
// handler.
func (handler *Handler) OnUserPreferencesUpdate(h UserPreferenceHandler) {
	handler.userPrefs.OnChange(h)
}
