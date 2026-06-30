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
	// through to [FallbackHandler].
	Handler UserPreferenceHandler

	// Fallback is called when Handler is nil. When nil, the event is silently
	// acknowledged (HTTP 200).
	Fallback FallbackHandler

	// ErrorHandler is called when Handler returns an error. When nil, the
	// error is returned as-is (passthrough).
	ErrorHandler ErrorHandler
}

// OnChange sets the handler for all user preference changes.
func (uh *UserPreferencesHandler) OnChange(h UserPreferenceHandler) {
	uh.Handler = h
}

// OnFallback sets the catch-all handler when [Handler] is nil.
func (uh *UserPreferencesHandler) OnFallback(h FallbackHandler) {
	uh.Fallback = h
}

func (uh *UserPreferencesHandler) handleError(ctx context.Context, err error) error {
	if uh.ErrorHandler == nil {
		return err
	}
	if handlerErr := uh.ErrorHandler.Handle(ctx, err); handlerErr != nil {
		return fmt.Errorf("error handler: %w", handlerErr)
	}
	return nil
}

func (uh *UserPreferencesHandler) executeFallback(ctx context.Context, ne NotificationEntry, change Change) error {
	if uh.Fallback == nil {
		return nil
	}
	if err := uh.Fallback.Handle(ctx, ne, change); err != nil {
		return fmt.Errorf("user preferences fallback: %w", err)
	}
	return nil
}

// Handle dispatches the user_preferences value to the handler.
//
//  1. If [Handler] is set, each entry in value.UserPreferences is passed
//     to it. Errors are routed through [ErrorHandler].
//  2. If [Handler] is nil, falls back to [Fallback].
//  3. If [Fallback] is also nil, the event is silently skipped (HTTP 200).
func (uh *UserPreferencesHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	if change.Value == nil {
		return nil
	}

	if uh.Handler == nil {
		return uh.executeFallback(ctx, ne, change)
	}

	nctx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Contacts:           change.Value.Contacts,
		Metadata:           change.Value.Metadata,
	}

	for _, pref := range change.Value.UserPreferences {
		if pref == nil {
			continue
		}
		if err := uh.Handler.Handle(ctx, nctx, pref); err != nil {
			return uh.handleError(ctx, fmt.Errorf("user preferences: %w", err))
		}
	}

	return nil
}

// OnUserPreferencesUpdate sets the handler for user preference changes. Each
// preference entry (stop or resume marketing messages) is delivered to the
// handler.
func (handler *Handler) OnUserPreferencesUpdate(h UserPreferenceHandler) {
	handler.ensureUserPrefs().OnChange(h)
}
