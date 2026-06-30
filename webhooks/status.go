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
	"fmt"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

func handleMessageChangeNotification[T any](
	ctx context.Context,
	handler *Handler,
	eventHandler ChangeValueHandler[T],
	ne NotificationEntry,
	change Change,
	events []*T,
) error {
	if eventHandler == nil {
		return nil
	}

	req := &ChangeValueRequest[T]{
		Notification: &MessageNotificationContext{
			EntryID:            ne.ID,
			EntryTime:          ne.Time,
			NotificationObject: ne.Object,
			MessagingProduct:   change.Value.MessagingProduct,
			Contacts:           change.Value.Contacts,
			Metadata:           change.Value.Metadata,
		},
		Payload: events,
	}

	if err := eventHandler.Handle(ctx, req); err != nil {
		if handler.errorHandler != nil {
			if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
				return fmt.Errorf("error handler: %w", handlerErr)
			}
		}
	}

	return nil
}

func (handler *Handler) handleNotificationMessageItem( //nolint: gocognit // ok
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	notificationCtx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Contacts:           change.Value.Contacts,
		Metadata:           change.Value.Metadata,
	}

	// handle notification errors do not terminate of its success, or if the error is not fatal
	if handler.notificationErrors != nil {
		req := &ChangeValueRequest[werrors.Error]{
			Notification: notificationCtx,
			Payload:      ErrorInfosAsErrors(change.Value.Errors),
		}
		if err := handler.notificationErrors.Handle(ctx, req); err != nil {
			if handler.errorHandler != nil {
				if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}

	if handler.messageStatusChange != nil {
		req := &ChangeValueRequest[Status]{
			Notification: notificationCtx,
			Payload:      change.Value.Statuses,
		}
		if err := handler.messageStatusChange.Handle(ctx, req); err != nil {
			if handler.errorHandler != nil {
				if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}

	for _, m := range change.Value.Messages {
		if err := handler.messages.Handle(ctx, notificationCtx, m); err != nil {
			if handler.errorHandler != nil {
				if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}

	return nil
}

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
	UserPreferenceUpdateHandler    = ChangeValueHandler[UserPreference]
	NotificationErrorsHandler      = ChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler     = ChangeValueHandler[Status]
	GroupLifecycleUpdateHandler    = ChangeValueHandler[Group]
	GroupParticipantsUpdateHandler = ChangeValueHandler[Group]
	GroupSettingsUpdateHandler     = ChangeValueHandler[Group]
	GroupStatusUpdateHandler       = ChangeValueHandler[Group]
	HistorySyncHandler             = ChangeValueHandler[HistoryEntry]
	SMBAppStateSyncHandler         = ChangeValueHandler[SMBAppStateSync]
)

func (f ChangeValueHandlerFunc[T]) Handle(ctx context.Context, req *ChangeValueRequest[T]) error {
	return f(ctx, req)
}

func NewNoOpChangeValueHandler[T any]() ChangeValueHandler[T] {
	return ChangeValueHandlerFunc[T](func(_ context.Context, _ *ChangeValueRequest[T]) error {
		return nil
	})
}

func (handler *Handler) OnUserPreferencesUpdate(h UserPreferenceUpdateHandler) {
	handler.userPreferencesUpdate = h
}

func (handler *Handler) OnSMBAppStateSync(h SMBAppStateSyncHandler) {
	handler.smbAppStateSync = h
}

func (handler *Handler) OnSMBMessageEcho(h MessageHandler[Message]) {
	handler.messages.Fallback = h
}

func (handler *Handler) OnNotificationErrors(h NotificationErrorsHandler) {
	handler.notificationErrors = h
}

func (handler *Handler) OnMessageStatusChange(h MessageStatusChangeHandler) {
	handler.messageStatusChange = h
}

func (handler *Handler) OnMessageErrors(h MessageErrorsHandler) {
	handler.messages.Unknown = h
}

func (handler *Handler) OnUnsupportedMessage(h MessageErrorsHandler) {
	handler.messages.Unsupported = h
}
