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

// Status types and change-value handler registration for WhatsApp message
// status webhooks. Includes delivery statuses, user preferences, SMB app
// state sync, SMB message echoes, notification errors, and message errors.

package webhooks

import (
	"context"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type (
	// Status represents a delivery status event for an outgoing message.
	//
	// StatusValue indicates the current state:
	//   sent      — message accepted by WhatsApp servers (one checkmark).
	//   delivered — message reached the user's device (two checkmarks).
	//              May be skipped if the message is read immediately.
	//   read      — message displayed in an open chat (two blue checkmarks).
	//   failed    — message could not be sent or delivered.
	//   played    — voice message was played by recipient (blue microphone).
	//
	// Conditional fields:
	//   Conversation — included with sent and one of delivered/read.
	//                  Omitted for v24.0+ unless sent within a free entry
	//                  point window.
	//   Pricing      — included with sent and one of delivered/read.
	//   Errors       — present only on failed status.
	//   RecipientType, RecipientParticipantID — present for group messages.
	//   BizOpaqueCallbackData — present if set when sending the message.
	Status struct {
		ID                     string             `json:"id,omitempty"`
		RecipientID            string             `json:"recipient_id,omitempty"`
		RecipientType          string             `json:"recipient_type,omitempty"`
		RecipientParticipantID string             `json:"recipient_participant_id,omitempty"`
		ParticipantRecipientID string             `json:"participant_recipient_id,omitempty"`
		StatusValue            string             `json:"status,omitempty"`
		Timestamp              string             `json:"timestamp,omitempty"`
		Conversation           *Conversation      `json:"conversation,omitempty"`
		Pricing                *Pricing           `json:"pricing,omitempty"`
		Errors                 []*werrors.Error   `json:"errors,omitempty"`
		BizOpaqueCallbackData  string             `json:"biz_opaque_callback_data,omitempty"`
		Message                *StatusMessageInfo `json:"message,omitempty"`
		Type                   string             `json:"type,omitempty"`
	}

	StatusMessageInfo struct {
		RecipientID string `json:"recipient_id,omitempty"`
	}

	UserPreference struct {
		WaID      string `json:"wa_id"`
		Detail    string `json:"detail"`
		Category  string `json:"category"` // always "marketing_messages"
		Value     string `json:"value"`    // can be "stop" or "resume"
		Timestamp string `json:"timestamp"`
	}
)

type (
	ChangeValueHandler[T any] interface {
		Handle(ctx context.Context, req *ChangeValueRequest[T]) error
	}
	ChangeValueHandlerFunc[T any] func(ctx context.Context, req *ChangeValueRequest[T]) error
)

type (
	NotificationErrorsHandler      = ChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler     = ChangeValueHandler[Status]
	GroupLifecycleUpdateHandler    = ChangeValueHandler[Group]
	GroupParticipantsUpdateHandler = ChangeValueHandler[Group]
	GroupSettingsUpdateHandler     = ChangeValueHandler[Group]
	GroupStatusUpdateHandler       = ChangeValueHandler[Group]
	HistorySyncHandler             = ChangeValueHandler[HistoryEntry]
)

func (f ChangeValueHandlerFunc[T]) Handle(ctx context.Context, req *ChangeValueRequest[T]) error {
	return f(ctx, req)
}

func NewNoOpChangeValueHandler[T any]() ChangeValueHandler[T] {
	return ChangeValueHandlerFunc[T](func(_ context.Context, _ *ChangeValueRequest[T]) error {
		return nil
	})
}

// OnNotificationErrors sets the handler for notification-level errors on the
// "messages" field. This is a convenience wrapper around
// [MessagesHandler.OnNotificationErrors].
func (handler *Handler) OnNotificationErrors(h NotificationErrorsHandler) {
	handler.ensureMessages().OnNotificationErrors(h)
}

// OnMessageStatusChange sets the handler for message delivery status updates
// (sent, delivered, read, failed). This is a convenience wrapper around
// [MessagesHandler.OnStatusChange].
func (handler *Handler) OnMessageStatusChange(h MessageStatusChangeHandler) {
	handler.ensureMessages().OnStatusChange(h)
}

func (handler *Handler) OnMessageErrors(h MessageErrorsHandler) {
	handler.ensureMessages().Unknown = h
}

func (handler *Handler) OnUnsupportedMessage(h MessageErrorsHandler) {
	handler.ensureMessages().Unsupported = h
}
