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

// Group types, GroupManagementHandler, and Handler registration methods for
// WhatsApp Groups API webhooks. Covers group lifecycle (create/delete),
// participant updates (join/leave/approve), settings changes, and status
// (suspension/clearance).

package webhooks

import (
	"context"
	"fmt"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type (
	Group struct {
		Timestamp           string                   `json:"timestamp"`
		GroupID             string                   `json:"group_id"`
		Type                string                   `json:"type"`
		RequestID           string                   `json:"request_id"`
		Subject             string                   `json:"subject"`
		InviteLink          string                   `json:"invite_link"`
		JoinApprovalMode    string                   `json:"join_approval_mode"`
		Description         string                   `json:"description"`
		Errors              []werrors.Error          `json:"errors,omitempty"`
		Reason              string                   `json:"reason"`
		AddedParticipants   []GroupParticipant       `json:"added_participants"`
		RemovedParticipants []GroupParticipant       `json:"removed_participants"`
		FailedParticipants  []FailedGroupParticipant `json:"failed_participants"`
		JoinRequestID       string                   `json:"join_request_id"`
		WaID                string                   `json:"wa_id"`
		InitiatedBy         string                   `json:"initiated_by"`
		ProfilePicture      *GroupProfilePicture     `json:"profile_picture,omitempty"`
		GroupSubject        *GroupSettingText        `json:"group_subject,omitempty"`
		GroupDescription    *GroupSettingText        `json:"group_description,omitempty"`
	}
	GroupParticipant struct {
		WaID  string `json:"wa_id,omitempty"`
		Input string `json:"input,omitempty"`
	}
	FailedGroupParticipant struct {
		Input  string          `json:"input,omitempty"`
		Errors []werrors.Error `json:"errors,omitempty"`
	}
	GroupProfilePicture struct {
		MimeType         string          `json:"mime_type"`
		UpdateSuccessful bool            `json:"update_successful"`
		SHA256           string          `json:"sha256"`
		Errors           []werrors.Error `json:"errors,omitempty"`
	}
	GroupSettingText struct {
		Text             string          `json:"text"`
		UpdateSuccessful bool            `json:"update_successful"`
		Errors           []werrors.Error `json:"errors,omitempty"`
	}
)

func (handler *Handler) OnGroupLifecycleUpdate(h GroupLifecycleUpdateHandler) {
	handler.groups.LifecycleUpdate = h
}

func (handler *Handler) OnGroupParticipantsUpdate(h GroupParticipantsUpdateHandler) {
	handler.groups.ParticipantsUpdate = h
}

func (handler *Handler) OnGroupSettingsUpdate(h GroupSettingsUpdateHandler) {
	handler.groups.SettingsUpdate = h
}

func (handler *Handler) OnGroupStatusUpdate(h GroupStatusUpdateHandler) {
	handler.groups.StatusUpdate = h
}

// GroupManagementHandler groups all group webhook field handlers into a single
// dispatch unit. Each field accepts a [ChangeValueHandler[Group]] for one
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
	LifecycleUpdate    ChangeValueHandler[Group]
	ParticipantsUpdate ChangeValueHandler[Group]
	SettingsUpdate     ChangeValueHandler[Group]
	StatusUpdate       ChangeValueHandler[Group]

	Fallback     FallbackHandler
	ErrorHandler ErrorHandler
}

// OnLifecycleUpdate sets the handler for group_lifecycle_update webhooks
// (group creation and deletion, with success and failure variants).
func (gh *GroupManagementHandler) OnLifecycleUpdate(h ChangeValueHandler[Group]) {
	gh.LifecycleUpdate = h
}

// OnParticipantsUpdate sets the handler for group_participants_update webhooks
// (participants joining via invite, requesting to join, cancelling requests,
// join request approval, participant removal, and participant departures).
func (gh *GroupManagementHandler) OnParticipantsUpdate(h ChangeValueHandler[Group]) {
	gh.ParticipantsUpdate = h
}

// OnSettingsUpdate sets the handler for group_settings_update webhooks
// (group subject, description, and profile picture changes with per-field
// success/failure reporting).
func (gh *GroupManagementHandler) OnSettingsUpdate(h ChangeValueHandler[Group]) {
	gh.SettingsUpdate = h
}

// OnStatusUpdate sets the handler for group_status_update webhooks
// (group suspension and suspension clearance).
func (gh *GroupManagementHandler) OnStatusUpdate(h ChangeValueHandler[Group]) {
	gh.StatusUpdate = h
}

// OnFallback sets the catch-all handler for group events without a dedicated
// sub-category handler. When nil, the general [Handler] fallback is tried.
func (gh *GroupManagementHandler) OnFallback(h FallbackHandler) {
	gh.Fallback = h
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
// change.Field. Nil handlers are silently skipped (HTTP 200).
//
//nolint:gocognit // dispatch switch
func (gh *GroupManagementHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	nctx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Metadata:           change.Value.Metadata,
		Contacts:           change.Value.Contacts,
	}

	switch change.Field {
	case ChangeFieldGroupLifecycleUpdate.String():
		if gh.LifecycleUpdate == nil {
			if gh.Fallback != nil {
				if err := gh.Fallback.Handle(ctx, ne, change); err != nil {
					return fmt.Errorf("group fallback: %w", err)
				}
			}
			return nil
		}
		if err := gh.LifecycleUpdate.Handle(ctx, &ChangeValueRequest[Group]{Notification: nctx, Payload: change.Value.Groups}); err != nil {
			return fmt.Errorf("group lifecycle update: %w", err)
		}
		return nil

	case ChangeFieldGroupParticipantsUpdate.String():
		if gh.ParticipantsUpdate == nil {
			if gh.Fallback != nil {
				if err := gh.Fallback.Handle(ctx, ne, change); err != nil {
					return fmt.Errorf("group fallback: %w", err)
				}
			}
			return nil
		}
		if err := gh.ParticipantsUpdate.Handle(ctx, &ChangeValueRequest[Group]{Notification: nctx, Payload: change.Value.Groups}); err != nil {
			return fmt.Errorf("group participants update: %w", err)
		}
		return nil

	case ChangeFieldGroupSettingsUpdate.String():
		if gh.SettingsUpdate == nil {
			if gh.Fallback != nil {
				if err := gh.Fallback.Handle(ctx, ne, change); err != nil {
					return fmt.Errorf("group fallback: %w", err)
				}
			}
			return nil
		}
		if err := gh.SettingsUpdate.Handle(ctx, &ChangeValueRequest[Group]{Notification: nctx, Payload: change.Value.Groups}); err != nil {
			return fmt.Errorf("group settings update: %w", err)
		}
		return nil

	case ChangeFieldGroupStatusUpdate.String():
		if gh.StatusUpdate == nil {
			if gh.Fallback != nil {
				if err := gh.Fallback.Handle(ctx, ne, change); err != nil {
					return fmt.Errorf("group fallback: %w", err)
				}
			}
			return nil
		}
		if err := gh.StatusUpdate.Handle(ctx, &ChangeValueRequest[Group]{Notification: nctx, Payload: change.Value.Groups}); err != nil {
			return fmt.Errorf("group status update: %w", err)
		}
		return nil
	}

	return nil
}
