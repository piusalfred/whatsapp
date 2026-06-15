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
	"fmt"
	"strings"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/media"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

const (
	MessageTypeAudio          MessageType = "audio"
	MessageTypeButton         MessageType = "button"
	MessageTypeDocument       MessageType = "document"
	MessageTypeText           MessageType = "text"
	MessageTypeImage          MessageType = "image"
	MessageTypeInteractive    MessageType = "interactive"
	MessageTypeOrder          MessageType = "order"
	MessageTypeSticker        MessageType = "sticker"
	MessageTypeSystem         MessageType = "system"
	MessageTypeUnknown        MessageType = "unknown"
	MessageTypeUnsupported    MessageType = "unsupported"
	MessageTypeVideo          MessageType = "video"
	MessageTypeLocation       MessageType = "location"
	MessageTypeReaction       MessageType = "reaction"
	MessageTypeContacts       MessageType = "contacts"
	MessageTypeRequestWelcome MessageType = "request_welcome"
	MessageTypeRevoke         MessageType = "revoke"
	MessageTypeEdit           MessageType = "edit"
)

// MessageType is a type of message that has been received by the business that has subscribed
// to Webhooks. Possible value can be one of the following: audio,button, document, text,image,
// interactive, order,sticker, system – for customer number change messages, unknown and video
// The documentation is not clear in case of location, reaction and contacts. They will be included
// just in case.
type MessageType string

// ParseMessageType parses the message type from a string.
func ParseMessageType(s string) MessageType {
	msgMap := map[string]MessageType{
		"audio":           MessageTypeAudio,
		"button":          MessageTypeButton,
		"document":        MessageTypeDocument,
		"text":            MessageTypeText,
		"image":           MessageTypeImage,
		"interactive":     MessageTypeInteractive,
		"order":           MessageTypeOrder,
		"sticker":         MessageTypeSticker,
		"system":          MessageTypeSystem,
		"unknown":         MessageTypeUnknown,
		"unsupported":     MessageTypeUnsupported,
		"video":           MessageTypeVideo,
		"location":        MessageTypeLocation,
		"reaction":        MessageTypeReaction,
		"contacts":        MessageTypeContacts,
		"request_welcome": MessageTypeRequestWelcome,
		"revoke":          MessageTypeRevoke,
		"edit":            MessageTypeEdit,
	}

	msgType, ok := msgMap[strings.TrimSpace(strings.ToLower(s))]
	if !ok {
		return ""
	}

	return msgType
}

const ErrUnrecognizedMessageType = whatsapp.Error("unrecognized message type")

const (
	InteractiveTypeListReply     = "list_reply"
	InteractiveTypeButtonReply   = "button_reply"
	InteractiveTypeNFMReply      = "nfm_reply"
	InteractiveAddressSubmission = "address_message"
)

type (
	MessageErrorsHandlerFunc func(
		ctx context.Context, notificationContext *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error
	MessageErrorsHandler interface {
		Handle(
			ctx context.Context,
			notificationContext *MessageNotificationContext,
			info *MessageInfo,
			errors []*werrors.Error,
		) error
	}
)

func (fn MessageErrorsHandlerFunc) Handle(ctx context.Context, notificationCtx *MessageNotificationContext,
	info *MessageInfo, errors []*werrors.Error,
) error {
	return fn(ctx, notificationCtx, info, errors)
}

func NewNoOpMessageErrorsHandler() MessageErrorsHandler {
	return MessageErrorsHandlerFunc(
		func(_ context.Context, _ *MessageNotificationContext, _ *MessageInfo, _ []*werrors.Error) error {
			return nil
		},
	)
}

func (handler *Handler) handleNotificationMessage( //nolint:gocognit,funlen // single dispatch switch
	ctx context.Context,
	notificationCtx *MessageNotificationContext,
	message *Message,
) error {
	info := &MessageInfo{
		From:             message.From,
		MessageID:        message.ID,
		Timestamp:        message.Timestamp,
		Type:             message.Type,
		GroupID:          message.GroupID,
		Context:          message.Context,
		IsAReply:         message.IsAReply(),
		IsForwarded:      message.IsForwarded(),
		IsProductInquiry: message.IsProductInquiry(),
		IsReferral:       message.IsReferral(),
	}

	messageType := ParseMessageType(message.Type)
	switch messageType {
	case MessageTypeOrder:
		if err := handler.orderMessage.Handle(ctx, notificationCtx, info, message.Order); err != nil {
			return fmt.Errorf("handle order message: %w", err)
		}

		return nil

	case MessageTypeButton:
		if err := handler.buttonMessage.Handle(ctx, notificationCtx, info, message.Button); err != nil {
			return fmt.Errorf("handle button message: %w", err)
		}

		return nil

	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return handler.handleMediaMessage(ctx, notificationCtx, message, info)

	case MessageTypeInteractive:
		return handler.handleInteractiveNotification(ctx, notificationCtx, message, info)

	case MessageTypeSystem:
		if err := handler.systemMessage.Handle(ctx, notificationCtx, info, message.System); err != nil {
			return fmt.Errorf("handle system message: %w", err)
		}

		return nil

	case MessageTypeUnknown:
		if err := handler.errorMessage.Handle(ctx, notificationCtx, info, message.Errors); err != nil {
			return fmt.Errorf("handle error message: %w", err)
		}

		return nil

	case MessageTypeUnsupported:
		if err := handler.unsupportedMessage.Handle(ctx, notificationCtx, info, message.Errors); err != nil {
			return fmt.Errorf("handle unsupported message: %w", err)
		}

		return nil

	case MessageTypeText:
		return handler.handleTextNotification(ctx, notificationCtx, message, info)

	case MessageTypeRequestWelcome:
		if handler.requestWelcome != nil {
			if err := handler.requestWelcome.Handle(ctx, notificationCtx, info, message); err != nil {
				return fmt.Errorf("handle request welcome: %w", err)
			}
		}

		return nil

	case MessageTypeReaction:
		if err := handler.reactionMessage.Handle(ctx, notificationCtx, info, message.Reaction); err != nil {
			return fmt.Errorf("handle reaction message: %w", err)
		}

		return nil

	case MessageTypeLocation:
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return fmt.Errorf("handle location message: %w", err)
		}

		return nil

	case MessageTypeContacts:
		if err := handler.contactsMessage.Handle(ctx, notificationCtx, info, message.Contacts); err != nil {
			return fmt.Errorf("handle contacts message: %w", err)
		}

		return nil

	case MessageTypeRevoke:
		if err := handler.revokeMessage.Handle(ctx, notificationCtx, info, message.Revoke); err != nil {
			return fmt.Errorf("handle revoke message: %w", err)
		}

		return nil

	case MessageTypeEdit:
		if err := handler.editMessage.Handle(ctx, notificationCtx, info, message.Edit); err != nil {
			return fmt.Errorf("handle edit message: %w", err)
		}

		return nil

	default:
		return handler.handleDefaultNotificationMessage(ctx, notificationCtx, message, info)
	}
}

func (handler *Handler) handleMediaMessage(ctx context.Context, notificationCtx *MessageNotificationContext,
	message *Message, info *MessageInfo,
) error {
	messageType := ParseMessageType(message.Type)
	switch messageType { //nolint:exhaustive // ok
	case MessageTypeAudio:
		if err := handler.audioMessage.Handle(ctx, notificationCtx, info, message.Audio); err != nil {
			return fmt.Errorf("handle audio message: %w", err)
		}

		return nil

	case MessageTypeVideo:
		if err := handler.videoMessage.Handle(ctx, notificationCtx, info, message.Video); err != nil {
			return fmt.Errorf("handle video message: %w", err)
		}

		return nil

	case MessageTypeImage:
		if err := handler.imageMessage.Handle(ctx, notificationCtx, info, message.Image); err != nil {
			return fmt.Errorf("handle image message: %w", err)
		}

		return nil

	case MessageTypeDocument:
		if err := handler.documentMessage.Handle(ctx, notificationCtx, info, message.Document); err != nil {
			return fmt.Errorf("handle document message: %w", err)
		}

		return nil

	case MessageTypeSticker:
		if err := handler.stickerMessage.Handle(ctx, notificationCtx, info, message.Sticker); err != nil {
			return fmt.Errorf("handle sticker message: %w", err)
		}

		return nil
	}

	return nil
}

func (handler *Handler) handleTextNotification(ctx context.Context, notificationCtx *MessageNotificationContext,
	message *Message, info *MessageInfo,
) error {
	if info.IsReferral {
		referral := &ReferralNotification{
			Text:     message.Text,
			Referral: message.Referral,
		}

		if err := handler.referralMessage.Handle(ctx, notificationCtx, info, referral); err != nil {
			return fmt.Errorf("handle referral message: %w", err)
		}

		return nil
	}

	if info.IsProductInquiry {
		if err := handler.productInquiry.Handle(ctx, notificationCtx, info, message.Text); err != nil {
			return fmt.Errorf("handle product inquiry: %w", err)
		}

		return nil
	}

	if err := handler.textMessage.Handle(ctx, notificationCtx, info, message.Text); err != nil {
		return fmt.Errorf("handle text message: %w", err)
	}

	return nil
}

func (handler *Handler) handleDefaultNotificationMessage(
	ctx context.Context,
	notificationCtx *MessageNotificationContext,
	message *Message,
	info *MessageInfo,
) error {
	if message.Contacts != nil {
		if err := handler.contactsMessage.Handle(ctx, notificationCtx, info, message.Contacts); err != nil {
			return fmt.Errorf("handle contacts message: %w", err)
		}

		return nil
	}
	if message.Location != nil {
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return fmt.Errorf("handle location message: %w", err)
		}

		return nil
	}

	if message.Identity != nil {
		if err := handler.customerIDChange.Handle(ctx, notificationCtx, info, message.Identity); err != nil {
			return fmt.Errorf("handle customer ID change: %w", err)
		}

		return nil
	}

	return ErrUnrecognizedMessageType
}

func (handler *Handler) handleInteractiveNotification(ctx context.Context,
	notificationCtx *MessageNotificationContext, message *Message, info *MessageInfo,
) error {
	switch message.Interactive.Type {
	case InteractiveTypeListReply:
		if err := handler.listReplyMessage.Handle(
			ctx,
			notificationCtx,
			info,
			message.Interactive.ListReply,
		); err != nil {
			return fmt.Errorf("handle list reply: %w", err)
		}

		return nil
	case InteractiveTypeButtonReply:
		if err := handler.buttonReplyMessage.Handle(
			ctx,
			notificationCtx,
			info,
			message.Interactive.ButtonReply,
		); err != nil {
			return fmt.Errorf("handle button reply: %w", err)
		}

		return nil
	case InteractiveTypeNFMReply:
		if err := handler.flowCompletionUpdate.Handle(
			ctx,
			notificationCtx,
			info,
			message.Interactive.NFMReply,
		); err != nil {
			return fmt.Errorf("handle flow completion update: %w", err)
		}

		return nil

	case InteractiveAddressSubmission:
		if err := handler.addressSubmission.Handle(
			ctx,
			notificationCtx,
			info,
			message.Interactive.NFMReply,
		); err != nil {
			return fmt.Errorf("handle address submission: %w", err)
		}
		return nil
	default:
		if err := handler.interactiveMessage.Handle(ctx, notificationCtx, info, message.Interactive); err != nil {
			return fmt.Errorf("handle interactive message: %w", err)
		}

		return nil
	}
}

type (
	MediaMessageHandler MessageHandler[media.Info]

	MessageHandlerFunc[T any] func(
		ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error
	}

	ButtonMessageHandler         = MessageHandler[Button]
	TextMessageHandler           = MessageHandler[Text]
	OrderMessageHandler          = MessageHandler[Order]
	LocationMessageHandler       = MessageHandler[media.Location]
	ContactsMessageHandler       = MessageHandler[message.Contacts]
	ReactionHandler              = MessageHandler[media.Reaction]
	ProductEnquiryHandler        = MessageHandler[Text]
	InteractiveMessageHandler    = MessageHandler[Interactive]
	ButtonReplyMessageHandler    = MessageHandler[ButtonReply]
	ListReplyMessageHandler      = MessageHandler[ListReply]
	NativeFlowCompletionHandler  = MessageHandler[NFMReply]
	ReferralMessageHandler       = MessageHandler[ReferralNotification]
	CustomerIDChangeHandler      = MessageHandler[Identity]
	SystemMessageHandler         = MessageHandler[System]
	RequestWelcomeMessageHandler = MessageHandler[Message]
)

func (fn MessageHandlerFunc[T]) Handle(ctx context.Context, notificationCtx *MessageNotificationContext,
	info *MessageInfo, message *T,
) error {
	return fn(ctx, notificationCtx, info, message)
}

func NewNoOpMessageHandler[T any]() MessageHandler[T] {
	return MessageHandlerFunc[T](func(_ context.Context, _ *MessageNotificationContext, _ *MessageInfo, _ *T) error {
		return nil
	})
}

// OnTextMessage registers a handler for text messages in the messages webhook.
func (handler *Handler) OnTextMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, text *Text) error,
) {
	handler.textMessage = MessageHandlerFunc[Text](fn)
}

// SetTextMessageHandler sets the handler for text messages.
func (handler *Handler) SetTextMessageHandler(
	h TextMessageHandler,
) {
	handler.textMessage = h
}

// OnButtonMessage registers a handler for button messages in the messages webhook.
func (handler *Handler) OnButtonMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, button *Button) error,
) {
	handler.buttonMessage = MessageHandlerFunc[Button](fn)
}

// SetButtonMessageHandler sets the handler for button messages.
func (handler *Handler) SetButtonMessageHandler(
	h ButtonMessageHandler,
) {
	handler.buttonMessage = h
}

// OnOrderMessage registers a handler for order messages in the messages webhook.
func (handler *Handler) OnOrderMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, order *Order) error,
) {
	handler.orderMessage = MessageHandlerFunc[Order](fn)
}

// SetOrderMessageHandler sets the handler for order messages.
func (handler *Handler) SetOrderMessageHandler(
	h OrderMessageHandler,
) {
	handler.orderMessage = h
}

// OnLocationMessage registers a handler for shared location messages in the
// messages webhook. The location includes latitude, longitude, name, address,
// and optionally a URL (usually only for business locations).
func (handler *Handler) OnLocationMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, loc *media.Location) error,
) {
	handler.locationMessage = MessageHandlerFunc[media.Location](fn)
}

// SetLocationMessageHandler sets the handler for location messages.
func (handler *Handler) SetLocationMessageHandler(
	h LocationMessageHandler,
) {
	handler.locationMessage = h
}

// OnContactsMessage registers a handler for contacts messages in the messages
// webhook. Contact payloads contain name, phone, email, address, URL, and
// organization fields — all optional since the WhatsApp user chooses what to
// share. If the message came via a Click to WhatsApp ad, the referral data is
// delivered to [OnReferralMessage] instead.
func (handler *Handler) OnContactsMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, contacts *message.Contacts) error,
) {
	handler.contactsMessage = MessageHandlerFunc[message.Contacts](fn)
}

// SetContactsMessageHandler sets the handler for contacts messages.
func (handler *Handler) SetContactsMessageHandler(
	h ContactsMessageHandler,
) {
	handler.contactsMessage = h
}

// OnReactionMessage registers a handler for reaction messages in the messages webhook.
func (handler *Handler) OnReactionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, reaction *media.Reaction) error,
) {
	handler.reactionMessage = MessageHandlerFunc[media.Reaction](fn)
}

// SetReactionMessageHandler sets the handler for reaction messages.
func (handler *Handler) SetReactionMessageHandler(
	h ReactionHandler,
) {
	handler.reactionMessage = h
}

// OnProductEnquiryMessage registers a handler for product enquiry messages in the messages webhook.
func (handler *Handler) OnProductEnquiryMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, txt *Text) error,
) {
	handler.productInquiry = MessageHandlerFunc[Text](fn)
}

// SetProductEnquiryMessageHandler sets the handler for product enquiry messages.
func (handler *Handler) SetProductEnquiryMessageHandler(
	h ProductEnquiryHandler,
) {
	handler.productInquiry = h
}

// OnInteractiveMessage registers a handler for interactive messages in the messages webhook.
func (handler *Handler) OnInteractiveMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, itv *Interactive) error,
) {
	handler.interactiveMessage = MessageHandlerFunc[Interactive](fn)
}

// SetInteractiveMessageHandler sets the handler for interactive messages.
func (handler *Handler) SetInteractiveMessageHandler(
	h InteractiveMessageHandler,
) {
	handler.interactiveMessage = h
}

// OnButtonReplyMessage registers a handler for button reply messages in the messages webhook.
func (handler *Handler) OnButtonReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, br *ButtonReply) error,
) {
	handler.buttonReplyMessage = MessageHandlerFunc[ButtonReply](fn)
}

// SetButtonReplyMessageHandler sets the handler for button reply messages.
func (handler *Handler) SetButtonReplyMessageHandler(
	h ButtonReplyMessageHandler,
) {
	handler.buttonReplyMessage = h
}

// OnListReplyMessage registers a handler for list reply messages in the messages webhook.
func (handler *Handler) OnListReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, lr *ListReply) error,
) {
	handler.listReplyMessage = MessageHandlerFunc[ListReply](fn)
}

// SetListReplyMessageHandler sets the handler for list reply messages.
func (handler *Handler) SetListReplyMessageHandler(
	h ListReplyMessageHandler,
) {
	handler.listReplyMessage = h
}

// OnFlowCompletionMessage registers a handler for flow completion messages in the messages webhook.
func (handler *Handler) OnFlowCompletionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, nfm *NFMReply) error,
) {
	handler.flowCompletionUpdate = MessageHandlerFunc[NFMReply](fn)
}

// SetFlowCompletionMessageHandler sets the handler for flow completion messages.
func (handler *Handler) SetFlowCompletionMessageHandler(
	h NativeFlowCompletionHandler,
) {
	handler.flowCompletionUpdate = h
}

// OnAddressSubmissionMessage registers a handler for address submission messages in the messages webhook.
func (handler *Handler) OnAddressSubmissionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, nfm *NFMReply) error,
) {
	handler.addressSubmission = MessageHandlerFunc[NFMReply](fn)
}

// SetAddressSubmissionHandler sets the handler for address submission messages.
func (handler *Handler) SetAddressSubmissionHandler(
	h NativeFlowCompletionHandler,
) {
	handler.addressSubmission = h
}

// OnReferralMessage registers a handler for messages originating from Click to
// WhatsApp ads. The referral object carries the ad ID, source URL, headline,
// body, media URLs, and click tracking ID (ctwa_clid). Present on any incoming
// message type (text, image, contacts, etc.) sent via a Click to WhatsApp ad.
func (handler *Handler) OnReferralMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, ref *ReferralNotification) error,
) {
	handler.referralMessage = MessageHandlerFunc[ReferralNotification](fn)
}

// SetReferralMessageHandler sets the handler for referral messages.
func (handler *Handler) SetReferralMessageHandler(
	h ReferralMessageHandler,
) {
	handler.referralMessage = h
}

// OnCustomerIDChangeMessage registers a handler for customer identity change messages in the messages webhook.
func (handler *Handler) OnCustomerIDChangeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, identity *Identity) error,
) {
	handler.customerIDChange = MessageHandlerFunc[Identity](fn)
}

// SetCustomerIDChangeMessageHandler sets the handler for customer identity change messages.
func (handler *Handler) SetCustomerIDChangeMessageHandler(
	h CustomerIDChangeHandler,
) {
	handler.customerIDChange = h
}

// OnSystemMessage registers a handler for system messages in the messages webhook.
func (handler *Handler) OnSystemMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, sys *System) error,
) {
	handler.systemMessage = MessageHandlerFunc[System](fn)
}

// SetSystemMessageHandler sets the handler for system messages.
func (handler *Handler) SetSystemMessageHandler(
	h SystemMessageHandler,
) {
	handler.systemMessage = h
}

// OnAudioMessage registers a handler for audio messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and download URL.
// For voice messages, check the voice field via the Media API.
func (handler *Handler) OnAudioMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *media.Info) error,
) {
	handler.audioMessage = MessageHandlerFunc[media.Info](fn)
}

// SetAudioMessageHandler sets the handler for audio messages.
func (handler *Handler) SetAudioMessageHandler(
	h MediaMessageHandler,
) {
	handler.audioMessage = h
}

// OnVideoMessage registers a handler for video messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+). Use the ID with the Media API or the URL directly with
// your access token to retrieve the asset.
func (handler *Handler) OnVideoMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *media.Info) error,
) {
	handler.videoMessage = MessageHandlerFunc[media.Info](fn)
}

// SetVideoMessageHandler sets the handler for video messages.
func (handler *Handler) SetVideoMessageHandler(
	h MediaMessageHandler,
) {
	handler.videoMessage = h
}

// OnImageMessage registers a handler for image messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+).
func (handler *Handler) OnImageMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *media.Info) error,
) {
	handler.imageMessage = MessageHandlerFunc[media.Info](fn)
}

// SetImageMessageHandler sets the handler for image messages.
func (handler *Handler) SetImageMessageHandler(
	h MediaMessageHandler,
) {
	handler.imageMessage = h
}

// OnDocumentMessage registers a handler for document messages in the messages
// webhook. Metadata includes filename, MIME type, SHA-256 hash, caption, and
// download URL.
func (handler *Handler) OnDocumentMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *media.Info) error,
) {
	handler.documentMessage = MessageHandlerFunc[media.Info](fn)
}

// SetDocumentMessageHandler sets the handler for document messages.
func (handler *Handler) SetDocumentMessageHandler(
	h MediaMessageHandler,
) {
	handler.documentMessage = h
}

// OnStickerMessage registers a handler for sticker messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and an animated flag.
func (handler *Handler) OnStickerMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *media.Info) error,
) {
	handler.stickerMessage = MessageHandlerFunc[media.Info](fn)
}

// SetStickerMessageHandler sets the handler for sticker messages.
func (handler *Handler) SetStickerMessageHandler(
	h MediaMessageHandler,
) {
	handler.stickerMessage = h
}

// OnRevokeMessage registers a callback for message deletion events. Triggers
// when a WhatsApp user deletes a previously sent message (within ~2 days of
// sending). The callback receives the original message ID that was revoked.
func (handler *Handler) OnRevokeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, revoke *Revoke) error,
) {
	handler.revokeMessage = MessageHandlerFunc[Revoke](fn)
}

// SetRevokeMessageHandler sets the handler for message revoke events.
func (handler *Handler) SetRevokeMessageHandler(h MessageHandler[Revoke]) {
	handler.revokeMessage = h
}

// OnMessageEdit registers a callback for edit events. Triggers when a WhatsApp
// user edits a previously sent message (text or media caption) within 15 minutes
// of sending. The callback receives the original message ID and the replacement
// content.
//
// Note: edit webhooks are temporarily unsupported by WhatsApp. Edited messages
// may currently arrive as unsupported message type instead of edit type.
func (handler *Handler) OnMessageEdit(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, edit *Edit) error,
) {
	handler.editMessage = MessageHandlerFunc[Edit](fn)
}

// SetMessageEditHandler sets the handler for message edit events.
func (handler *Handler) SetMessageEditHandler(h MessageHandler[Edit]) {
	handler.editMessage = h
}
