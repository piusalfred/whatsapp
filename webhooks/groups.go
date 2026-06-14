//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks

import (
	"context"

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
		WaID  string `json:"wa_id"`
		Input string `json:"input"`
	}
	FailedGroupParticipant struct {
		Input  string          `json:"input"`
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

// OnGroupLifecycleUpdate registers a handler for group_lifecycle_update webhooks
// (group creation, deletion, subject/icon changes).
func (handler *Handler) OnGroupLifecycleUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupLifecycleUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// SetGroupLifecycleUpdateHandler sets the handler for group_lifecycle_update webhooks.
func (handler *Handler) SetGroupLifecycleUpdateHandler(
	h GroupLifecycleUpdateHandler,
) {
	handler.groupLifecycleUpdate = h
}

// OnGroupParticipantsUpdate registers a handler for group_participants_update webhooks
// (members added, removed, promoted to admin, or demoted).
func (handler *Handler) OnGroupParticipantsUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupParticipantsUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// SetGroupParticipantsUpdateHandler sets the handler for group_participants_update webhooks.
func (handler *Handler) SetGroupParticipantsUpdateHandler(
	h GroupParticipantsUpdateHandler,
) {
	handler.groupParticipantsUpdate = h
}

// OnGroupSettingsUpdate registers a handler for group_settings_update webhooks
// (group subject, description, or join approval mode changes).
func (handler *Handler) OnGroupSettingsUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupSettingsUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// SetGroupSettingsUpdateHandler sets the handler for group_settings_update webhooks.
func (handler *Handler) SetGroupSettingsUpdateHandler(
	h GroupSettingsUpdateHandler,
) {
	handler.groupSettingsUpdate = h
}

// OnGroupStatusUpdate registers a handler for group_status_update webhooks
// (group invite link or profile picture changes).
func (handler *Handler) OnGroupStatusUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupStatusUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

// SetGroupStatusUpdateHandler sets the handler for group_status_update webhooks.
func (handler *Handler) SetGroupStatusUpdateHandler(
	h GroupStatusUpdateHandler,
) {
	handler.groupStatusUpdate = h
}
