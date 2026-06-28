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

// GroupManagementHandler implementation and ChangeField helpers for group
// webhook fields. Provides IsGroupManagementWebhook and GroupChangeFields
// for querying group-related ChangeField values.

package webhooks

import (
	"context"
	"fmt"
)

// GroupManagementHandler groups all group webhook field handlers into a single
// dispatch unit. Each field accepts a [MessageChangeValueHandler[Group]] for one
// WhatsApp group notification type. Leave a field nil to silently skip that
// notification type (HTTP 200).
//
// Group lifecycle events cover create/delete (success and failure variants).
// Participant events cover joins via invite, join requests, request
// cancellation, join approval, removal, and departures — each with per-user
// success/failure arrays. Settings events cover subject, description, and
// profile picture updates — each with per-field success/failure flags. Status
// events cover group suspension and clearance.
//
// Usage:
//
//	gh := &GroupManagementHandler{}
//	gh.OnLifecycleUpdate(myLifecycleHandler)
//	gh.OnParticipantsUpdate(myParticipantsHandler)
type GroupManagementHandler struct {
	LifecycleUpdate    MessageChangeValueHandler[Group]
	ParticipantsUpdate MessageChangeValueHandler[Group]
	SettingsUpdate     MessageChangeValueHandler[Group]
	StatusUpdate       MessageChangeValueHandler[Group]
}

// OnLifecycleUpdate sets the handler for group_lifecycle_update webhooks
// (group creation and deletion, with success and failure variants).
func (gh *GroupManagementHandler) OnLifecycleUpdate(
	fn func(ctx context.Context, nctx *MessageNotificationContext, groups []*Group) error,
) {
	gh.LifecycleUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// OnParticipantsUpdate sets the handler for group_participants_update webhooks
// (participants joining via invite, requesting to join, cancelling requests,
// join request approval, participant removal, and participant departures).
func (gh *GroupManagementHandler) OnParticipantsUpdate(
	fn func(ctx context.Context, nctx *MessageNotificationContext, groups []*Group) error,
) {
	gh.ParticipantsUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// OnSettingsUpdate sets the handler for group_settings_update webhooks
// (group subject, description, and profile picture changes with per-field
// success/failure reporting).
func (gh *GroupManagementHandler) OnSettingsUpdate(
	fn func(ctx context.Context, nctx *MessageNotificationContext, groups []*Group) error,
) {
	gh.SettingsUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// OnStatusUpdate sets the handler for group_status_update webhooks
// (group suspension and suspension clearance).
func (gh *GroupManagementHandler) OnStatusUpdate(
	fn func(ctx context.Context, nctx *MessageNotificationContext, groups []*Group) error,
) {
	gh.StatusUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// IsGroupManagementWebhook reports whether field is one of the four group
// management webhook fields (lifecycle, participants, settings, status).
func (f ChangeField) IsGroupManagementWebhook() bool {
	switch f {
	case ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate:
		return true
	default:
		return false
	}
}

// GroupChangeFields returns every [ChangeField] that triggers a group
// management webhook notification.
func GroupChangeFields() []ChangeField {
	return []ChangeField{
		ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate,
	}
}

// Handle dispatches the group notification to the correct handler based on
// change.Field. Nil handlers are silently skipped (HTTP 200). Errors from user
// handlers are routed through onError; if onError returns a non-nil error,
// processing stops.
//
//nolint:gocognit // dispatch switch
func (gh *GroupManagementHandler) Handle(
	ctx context.Context,
	change Change,
	entry Entry,
	onError func(ctx context.Context, err error) error,
) error {
	nctx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Metadata:         change.Value.Metadata,
		Contacts:         change.Value.Contacts,
	}

	switch change.Field {
	case ChangeFieldGroupLifecycleUpdate.String():
		if gh.LifecycleUpdate == nil {
			return nil
		}
		if err := gh.LifecycleUpdate.Handle(ctx, nctx, change.Value.Groups); err != nil {
			if onError != nil {
				if handlerErr := onError(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}

	case ChangeFieldGroupParticipantsUpdate.String():
		if gh.ParticipantsUpdate == nil {
			return nil
		}
		if err := gh.ParticipantsUpdate.Handle(ctx, nctx, change.Value.Groups); err != nil {
			if onError != nil {
				if handlerErr := onError(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}

	case ChangeFieldGroupSettingsUpdate.String():
		if gh.SettingsUpdate == nil {
			return nil
		}
		if err := gh.SettingsUpdate.Handle(ctx, nctx, change.Value.Groups); err != nil {
			if onError != nil {
				if handlerErr := onError(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}

	case ChangeFieldGroupStatusUpdate.String():
		if gh.StatusUpdate == nil {
			return nil
		}
		if err := gh.StatusUpdate.Handle(ctx, nctx, change.Value.Groups); err != nil {
			if onError != nil {
				if handlerErr := onError(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}

	return nil
}
