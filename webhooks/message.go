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

func (handler *Handler) handleNotificationMessage(ctx context.Context,
	notificationCtx *MessageNotificationContext, message *Message,
) error {
	info := &MessageInfo{
		From:             message.From,
		MessageID:        message.ID,
		Timestamp:        message.Timestamp,
		Type:             message.Type,
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
			return err
		}

		return nil

	case MessageTypeButton:
		if err := handler.buttonMessage.Handle(ctx, notificationCtx, info, message.Button); err != nil {
			return err
		}

		return nil

	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return handler.handleMediaMessage(ctx, notificationCtx, message, info)

	case MessageTypeInteractive:
		return handler.handleInteractiveNotification(ctx, notificationCtx, message, info)

	case MessageTypeSystem:
		if err := handler.systemMessage.Handle(ctx, notificationCtx, info, message.System); err != nil {
			return err
		}

		return nil

	case MessageTypeUnknown:
		if err := handler.errorMessage.Handle(ctx, notificationCtx, info, message.Errors); err != nil {
			return err
		}

		return nil

	case MessageTypeUnsupported:
		return handler.unsupportedMessage.Handle(ctx, notificationCtx, info, message.Errors)

	case MessageTypeText:
		return handler.handleTextNotification(ctx, notificationCtx, message, info)

	case MessageTypeRequestWelcome:
		if handler.requestWelcome != nil {
			if err := handler.requestWelcome.Handle(ctx, notificationCtx, info, message); err != nil {
				return err
			}
		}

		return nil

	case MessageTypeReaction:
		if err := handler.reactionMessage.Handle(ctx, notificationCtx, info, message.Reaction); err != nil {
			return err
		}

		return nil

	case MessageTypeLocation:
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return err
		}

		return nil

	case MessageTypeContacts:
		if err := handler.contactsMessage.Handle(ctx, notificationCtx, info, message.Contacts); err != nil {
			return err
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
			return err
		}

		return nil

	case MessageTypeVideo:
		if err := handler.videoMessage.Handle(ctx, notificationCtx, info, message.Video); err != nil {
			return err
		}

		return nil

	case MessageTypeImage:
		if err := handler.imageMessage.Handle(ctx, notificationCtx, info, message.Image); err != nil {
			return err
		}

		return nil

	case MessageTypeDocument:
		if err := handler.documentMessage.Handle(ctx, notificationCtx, info, message.Document); err != nil {
			return err
		}

		return nil

	case MessageTypeSticker:
		if err := handler.stickerMessage.Handle(ctx, notificationCtx, info, message.Sticker); err != nil {
			return err
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
			return err
		}

		return nil
	}

	if info.IsProductInquiry {
		if err := handler.productInquiry.Handle(ctx, notificationCtx, info, message.Text); err != nil {
			return err
		}

		return nil
	}

	if err := handler.textMessage.Handle(ctx, notificationCtx, info, message.Text); err != nil {
		return err
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
			return err
		}

		return nil
	}
	if message.Location != nil {
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return err
		}

		return nil
	}

	if message.Identity != nil {
		if err := handler.customerIDChange.Handle(ctx, notificationCtx, info, message.Identity); err != nil {
			return err
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
		if err := handler.listReplyMessage.Handle(ctx, notificationCtx, info, message.Interactive.ListReply); err != nil {
			return fmt.Errorf("handle list reply: %w", err)
		}

		return nil
	case InteractiveTypeButtonReply:
		if err := handler.buttonReplyMessage.Handle(ctx, notificationCtx, info, message.Interactive.ButtonReply); err != nil {
			return fmt.Errorf("handle button reply: %w", err)
		}

		return nil
	case InteractiveTypeNFMReply:
		if err := handler.flowCompletionUpdate.Handle(ctx, notificationCtx, info, message.Interactive.NFMReply); err != nil {
			return fmt.Errorf("handle flow completion update: %w", err)
		}

		return nil

	case InteractiveAddressSubmission:
		if err := handler.addressSubmission.Handle(ctx, notificationCtx, info, message.Interactive.NFMReply); err != nil {
			return fmt.Errorf("handle address submission: %w", err)
		}
		return nil
	default:
		if err := handler.interactiveMessage.Handle(ctx, notificationCtx, info, message.Interactive); err != nil {
			return err
		}

		return nil
	}
}

type (
	MediaMessageHandler MessageHandler[message.MediaInfo]

	MessageHandlerFunc[T any] func(
		ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error
	}

	ButtonMessageHandler         = MessageHandler[Button]
	TextMessageHandler           = MessageHandler[Text]
	OrderMessageHandler          = MessageHandler[Order]
	LocationMessageHandler       = MessageHandler[message.Location]
	ContactsMessageHandler       = MessageHandler[message.Contacts]
	ReactionHandler              = MessageHandler[message.Reaction]
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

func (handler *Handler) OnTextMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, text *Text) error,
) {
	handler.textMessage = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetTextMessageHandler(
	h TextMessageHandler,
) {
	handler.textMessage = h
}

func (handler *Handler) OnButtonMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, button *Button) error,
) {
	handler.buttonMessage = MessageHandlerFunc[Button](fn)
}

func (handler *Handler) SetButtonMessageHandler(
	h ButtonMessageHandler,
) {
	handler.buttonMessage = h
}

func (handler *Handler) OnOrderMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, order *Order) error,
) {
	handler.orderMessage = MessageHandlerFunc[Order](fn)
}

func (handler *Handler) SetOrderMessageHandler(
	h OrderMessageHandler,
) {
	handler.orderMessage = h
}

func (handler *Handler) OnLocationMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, loc *message.Location) error,
) {
	handler.locationMessage = MessageHandlerFunc[message.Location](fn)
}

func (handler *Handler) SetLocationMessageHandler(
	h LocationMessageHandler,
) {
	handler.locationMessage = h
}

func (handler *Handler) OnContactsMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, contacts *message.Contacts) error,
) {
	handler.contactsMessage = MessageHandlerFunc[message.Contacts](fn)
}

func (handler *Handler) SetContactsMessageHandler(
	h ContactsMessageHandler,
) {
	handler.contactsMessage = h
}

func (handler *Handler) OnReactionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, reaction *message.Reaction) error,
) {
	handler.reactionMessage = MessageHandlerFunc[message.Reaction](fn)
}

func (handler *Handler) SetReactionMessageHandler(
	h ReactionHandler,
) {
	handler.reactionMessage = h
}

func (handler *Handler) OnProductEnquiryMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, txt *Text) error,
) {
	handler.productInquiry = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetProductEnquiryMessageHandler(
	h ProductEnquiryHandler,
) {
	handler.productInquiry = h
}

func (handler *Handler) OnInteractiveMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, itv *Interactive) error,
) {
	handler.interactiveMessage = MessageHandlerFunc[Interactive](fn)
}

func (handler *Handler) SetInteractiveMessageHandler(
	h InteractiveMessageHandler,
) {
	handler.interactiveMessage = h
}

func (handler *Handler) OnButtonReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, br *ButtonReply) error,
) {
	handler.buttonReplyMessage = MessageHandlerFunc[ButtonReply](fn)
}

func (handler *Handler) SetButtonReplyMessageHandler(
	h ButtonReplyMessageHandler,
) {
	handler.buttonReplyMessage = h
}

func (handler *Handler) OnListReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, lr *ListReply) error,
) {
	handler.listReplyMessage = MessageHandlerFunc[ListReply](fn)
}

func (handler *Handler) SetListReplyMessageHandler(
	h ListReplyMessageHandler,
) {
	handler.listReplyMessage = h
}

func (handler *Handler) OnFlowCompletionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, nfm *NFMReply) error,
) {
	handler.flowCompletionUpdate = MessageHandlerFunc[NFMReply](fn)
}

func (handler *Handler) SetFlowCompletionMessageHandler(
	h NativeFlowCompletionHandler,
) {
	handler.flowCompletionUpdate = h
}

func (handler *Handler) OnAddressSubmissionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, nfm *NFMReply) error,
) {
	handler.addressSubmission = MessageHandlerFunc[NFMReply](fn)
}

func (handler *Handler) SetAddressSubmissionHandler(
	h NativeFlowCompletionHandler,
) {
	handler.addressSubmission = h
}

func (handler *Handler) OnReferralMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, ref *ReferralNotification) error,
) {
	handler.referralMessage = MessageHandlerFunc[ReferralNotification](fn)
}

func (handler *Handler) SetReferralMessageHandler(
	h ReferralMessageHandler,
) {
	handler.referralMessage = h
}

func (handler *Handler) OnCustomerIDChangeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, identity *Identity) error,
) {
	handler.customerIDChange = MessageHandlerFunc[Identity](fn)
}

func (handler *Handler) SetCustomerIDChangeMessageHandler(
	h CustomerIDChangeHandler,
) {
	handler.customerIDChange = h
}

func (handler *Handler) OnSystemMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, sys *System) error,
) {
	handler.systemMessage = MessageHandlerFunc[System](fn)
}

func (handler *Handler) SetSystemMessageHandler(
	h SystemMessageHandler,
) {
	handler.systemMessage = h
}

func (handler *Handler) OnAudioMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.audioMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetAudioMessageHandler(
	h MediaMessageHandler,
) {
	handler.audioMessage = h
}

func (handler *Handler) OnVideoMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.videoMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetVideoMessageHandler(
	h MediaMessageHandler,
) {
	handler.videoMessage = h
}

func (handler *Handler) OnImageMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.imageMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetImageMessageHandler(
	h MediaMessageHandler,
) {
	handler.imageMessage = h
}

func (handler *Handler) OnDocumentMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.documentMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetDocumentMessageHandler(
	h MediaMessageHandler,
) {
	handler.documentMessage = h
}

func (handler *Handler) OnStickerMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.stickerMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetStickerMessageHandler(
	h MediaMessageHandler,
) {
	handler.stickerMessage = h
}
