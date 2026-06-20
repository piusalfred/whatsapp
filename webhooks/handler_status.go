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

package webhooks

import (
	"context"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

func handleMessageChangeNotification[T any](
	ctx context.Context,
	handler *Handler,
	eventHandler MessageChangeValueHandler[T],
	change Change,
	entry Entry,
	events []*T,
) error {
	if eventHandler == nil {
		return nil
	}

	notificationCtx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Contacts:         change.Value.Contacts,
		Metadata:         change.Value.Metadata,
	}

	if err := eventHandler.Handle(ctx, notificationCtx, events); err != nil {
		if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
			return handlerErr
		}
	}

	return nil
}

func (handler *Handler) handleNotificationMessageItem( //nolint: gocognit // ok
	ctx context.Context,
	entry Entry,
	change Change,
) error {
	notificationCtx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Contacts:         change.Value.Contacts,
		Metadata:         change.Value.Metadata,
	}

	// handle notification errors do not terminate of its success, or if the error is not fatal
	if handler.notificationErrors != nil {
		if err := handler.notificationErrors.Handle(
			ctx, notificationCtx, ErrorInfosAsErrors(change.Value.Errors),
		); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}

	if handler.messageStatusChange != nil {
		if err := handler.messageStatusChange.Handle(
			ctx, notificationCtx, change.Value.Statuses,
		); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}

	for _, m := range change.Value.Messages {
		if err := handler.handleNotificationMessage(ctx, notificationCtx, m); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
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
	MessageChangeValueHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
	}
	MessageChangeValueHandlerFunc[T any] func(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
)

type (
	UserPreferenceUpdateHandler    = MessageChangeValueHandler[UserPreference]
	NotificationErrorsHandler      = MessageChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler     = MessageChangeValueHandler[Status]
	GroupLifecycleUpdateHandler    = MessageChangeValueHandler[Group]
	GroupParticipantsUpdateHandler = MessageChangeValueHandler[Group]
	GroupSettingsUpdateHandler     = MessageChangeValueHandler[Group]
	GroupStatusUpdateHandler       = MessageChangeValueHandler[Group]
	HistorySyncHandler             = MessageChangeValueHandler[HistoryEntry]
	SMBAppStateSyncHandler         = MessageChangeValueHandler[SMBAppStateSync]
)

func (f MessageChangeValueHandlerFunc[T]) Handle(
	ctx context.Context,
	notificationCtx *MessageNotificationContext,
	values []*T,
) error {
	return f(ctx, notificationCtx, values)
}

func NewNoOpMessageChangeValueHandler[T any]() MessageChangeValueHandler[T] {
	return MessageChangeValueHandlerFunc[T](func(_ context.Context, _ *MessageNotificationContext, _ []*T) error {
		return nil
	})
}

// OnUserPreferencesUpdate registers a handler for user_preferences webhooks
// (WhatsApp user marketing message opt-in/out changes).
func (handler *Handler) OnUserPreferencesUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, prefs []*UserPreference) error,
) {
	handler.userPreferencesUpdate = MessageChangeValueHandlerFunc[UserPreference](fn)
}

// SetUserPreferencesUpdateHandler sets the handler for user_preferences webhooks.
func (handler *Handler) SetUserPreferencesUpdateHandler(
	h UserPreferenceUpdateHandler,
) {
	handler.userPreferencesUpdate = h
}

// OnSMBAppStateSync registers a callback for SMB contact sync events.
// Triggers when a solution provider syncs contacts for an onboarded business,
// or when the business customer adds, edits, or removes a contact in their
// WhatsApp Business app address book.
func (handler *Handler) OnSMBAppStateSync(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, syncs []*SMBAppStateSync) error,
) {
	handler.smbAppStateSync = MessageChangeValueHandlerFunc[SMBAppStateSync](fn)
}

// SetSMBAppStateSyncHandler sets the handler for smb_app_state_sync webhooks.
func (handler *Handler) SetSMBAppStateSyncHandler(h SMBAppStateSyncHandler) {
	handler.smbAppStateSync = h
}

// OnSMBMessageEcho registers a callback for messages sent by an onboarded
// business customer via their WhatsApp Business app or companion device.
// Triggers: business sends a message, revokes a message, or edits a message
// using the WhatsApp Business app.
// The payload shape is identical to incoming messages — use the same Message
// handlers you use for regular messages.
func (handler *Handler) OnSMBMessageEcho(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, msg *Message) error,
) {
	handler.smbMessageEcho = MessageHandlerFunc[Message](fn)
}

// SetSMBMessageEchoHandler sets the handler for smb_message_echoes webhooks.
func (handler *Handler) SetSMBMessageEchoHandler(h MessageHandler[Message]) {
	handler.smbMessageEcho = h
}

// OnNotificationErrors registers a handler for notification-level errors in the messages webhook.
func (handler *Handler) OnNotificationErrors(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, errors []*werrors.Error) error,
) {
	handler.notificationErrors = MessageChangeValueHandlerFunc[werrors.Error](fn)
}

// SetNotificationErrorsHandler sets the handler for notification-level errors.
func (handler *Handler) SetNotificationErrorsHandler(
	fn NotificationErrorsHandler,
) {
	handler.notificationErrors = fn
}

// OnMessageStatusChange registers a handler for message delivery status updates
// (sent, delivered, read, played, failed). A single outgoing message may trigger
// up to three status webhooks. The "delivered" webhook may be skipped if the
// message is read immediately.
func (handler *Handler) OnMessageStatusChange(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, status []*Status) error,
) {
	handler.messageStatusChange = MessageChangeValueHandlerFunc[Status](fn)
}

// SetMessageStatusChangeHandler sets the handler for message status changes.
func (handler *Handler) SetMessageStatusChangeHandler(
	fn MessageStatusChangeHandler,
) {
	handler.messageStatusChange = fn
}

// OnMessageErrors registers a handler for message-level errors in the messages webhook.
func (handler *Handler) OnMessageErrors(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error,
) {
	handler.errorMessage = MessageErrorsHandlerFunc(fn)
}

// SetMessageErrorsHandler sets the handler for message-level errors.
func (handler *Handler) SetMessageErrorsHandler(
	fn MessageErrorsHandler,
) {
	handler.errorMessage = fn
}

// OnUnsupportedMessage registers a handler for unsupported messages in the messages webhook.
func (handler *Handler) OnUnsupportedMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error,
) {
	handler.unsupportedMessage = MessageErrorsHandlerFunc(fn)
}

// SetUnsupportedMessageHandler sets the handler for unsupported messages.
func (handler *Handler) SetUnsupportedMessageHandler(
	fn MessageErrorsHandler,
) {
	handler.unsupportedMessage = fn
}
