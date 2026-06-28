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

// Message types define the WhatsApp message type constants, the MessageHandler
// and MessageErrorsHandler interfaces, type aliases for per-message-type
// handlers, and all On/Set registration methods on Handler for message
// webhooks.

package webhooks

import (
	"context"
	"errors"
	"strings"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/media"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

const (
	MessageTypeAudio            MessageType = "audio"
	MessageTypeButton           MessageType = "button"
	MessageTypeDocument         MessageType = "document"
	MessageTypeText             MessageType = "text"
	MessageTypeImage            MessageType = "image"
	MessageTypeInteractive      MessageType = "interactive"
	MessageTypeOrder            MessageType = "order"
	MessageTypeSticker          MessageType = "sticker"
	MessageTypeSystem           MessageType = "system"
	MessageTypeUnknown          MessageType = "unknown"
	MessageTypeUnsupported      MessageType = "unsupported"
	MessageTypeVideo            MessageType = "video"
	MessageTypeLocation         MessageType = "location"
	MessageTypeReaction         MessageType = "reaction"
	MessageTypeContacts         MessageType = "contacts"
	MessageTypeRequestWelcome   MessageType = "request_welcome"
	MessageTypeRevoke           MessageType = "revoke"
	MessageTypeEdit             MessageType = "edit"
	MessageTypeErrors           MessageType = "errors"
	MessageTypeGif              MessageType = "gif"
	MessageTypeGroupInvite      MessageType = "group_invite"
	MessageTypeHsm              MessageType = "hsm"
	MessageTypeKeepInChat       MessageType = "keep_in_chat"
	MessageTypeLinkPreview      MessageType = "link_preview"
	MessageTypeList             MessageType = "list"
	MessageTypeMediaPlaceholder MessageType = "media_placeholder"
	MessageTypePin              MessageType = "pin"
	MessageTypePollCreation     MessageType = "poll_creation"
	MessageTypePollUpdate       MessageType = "poll_update"
	MessageTypeProduct          MessageType = "product"
)

// MessageType is a type of message that has been received by the business that has subscribed
// to Webhooks. Possible value can be one of the following: audio,button, document, text,image,
// interactive, order,sticker, system – for customer number change messages, unknown and video
// The documentation is not clear in case of location, reaction and contacts. They will be included
// just in case.
type MessageType string

func (mm MessageType) String() string {
	return string(mm)
}

// ParseMessageType parses the message type from a string.
func ParseMessageType(s string) MessageType {
	msgMap := map[string]MessageType{
		"audio":             MessageTypeAudio,
		"button":            MessageTypeButton,
		"document":          MessageTypeDocument,
		"text":              MessageTypeText,
		"image":             MessageTypeImage,
		"interactive":       MessageTypeInteractive,
		"order":             MessageTypeOrder,
		"sticker":           MessageTypeSticker,
		"system":            MessageTypeSystem,
		"unknown":           MessageTypeUnknown,
		"unsupported":       MessageTypeUnsupported,
		"video":             MessageTypeVideo,
		"location":          MessageTypeLocation,
		"reaction":          MessageTypeReaction,
		"contacts":          MessageTypeContacts,
		"request_welcome":   MessageTypeRequestWelcome,
		"revoke":            MessageTypeRevoke,
		"edit":              MessageTypeEdit,
		"errors":            MessageTypeErrors,
		"gif":               MessageTypeGif,
		"link_preview":      MessageTypeLinkPreview,
		"list":              MessageTypeList,
		"media_placeholder": MessageTypeMediaPlaceholder,
		"pin":               MessageTypePin,
		"poll_creation":     MessageTypePollCreation,
		"poll_update":       MessageTypePollUpdate,
		"product":           MessageTypeProduct,
		"hsm":               MessageTypeHsm,
		"keep_in_chat":      MessageTypeKeepInChat,
		"group_invite":      MessageTypeGroupInvite,
	}

	msgType, ok := msgMap[strings.TrimSpace(strings.ToLower(s))]
	if !ok {
		return ""
	}

	return msgType
}

var ErrUnrecognizedMessageType = errors.New("unrecognized message type")

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

type (
	MediaMessageHandler MessageHandler[media.Info]

	MessageHandlerFunc[T any] func(
		ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext,
			info *MessageInfo, message *T) error
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

func (fn MessageHandlerFunc[T]) Handle(ctx context.Context,
	notificationCtx *MessageNotificationContext,
	info *MessageInfo, message *T,
) error {
	return fn(ctx, notificationCtx, info, message)
}

func NewNoOpMessageHandler[T any]() MessageHandler[T] {
	return MessageHandlerFunc[T](func(_ context.Context, _ *MessageNotificationContext,
		_ *MessageInfo, _ *T,
	) error {
		return nil
	})
}

// OnTextMessage registers a handler for text messages in the messages webhook.
func (handler *Handler) OnTextMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, text *Text) error,
) {
	handler.messages.Text = MessageHandlerFunc[Text](fn)
}

// SetTextMessageHandler sets the handler for text messages.
func (handler *Handler) SetTextMessageHandler(
	h TextMessageHandler,
) {
	handler.messages.Text = h
}

// OnButtonMessage registers a handler for button messages in the messages webhook.
func (handler *Handler) OnButtonMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, button *Button) error,
) {
	handler.messages.Button = MessageHandlerFunc[Button](fn)
}

// SetButtonMessageHandler sets the handler for button messages.
func (handler *Handler) SetButtonMessageHandler(
	h ButtonMessageHandler,
) {
	handler.messages.Button = h
}

// OnOrderMessage registers a handler for order messages in the messages webhook.
func (handler *Handler) OnOrderMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, order *Order) error,
) {
	handler.messages.Order = MessageHandlerFunc[Order](fn)
}

// SetOrderMessageHandler sets the handler for order messages.
func (handler *Handler) SetOrderMessageHandler(
	h OrderMessageHandler,
) {
	handler.messages.Order = h
}

// OnLocationMessage registers a handler for shared location messages in the
// messages webhook. The location includes latitude, longitude, name, address,
// and optionally a URL (usually only for business locations).
func (handler *Handler) OnLocationMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, loc *media.Location) error,
) {
	handler.messages.Location = MessageHandlerFunc[media.Location](fn)
}

// SetLocationMessageHandler sets the handler for location messages.
func (handler *Handler) SetLocationMessageHandler(
	h LocationMessageHandler,
) {
	handler.messages.Location = h
}

// OnContactsMessage registers a handler for contacts messages in the messages
// webhook. Contact payloads contain name, phone, email, address, URL, and
// organization fields — all optional since the WhatsApp user chooses what to
// share. If the message came via a Click to WhatsApp ad, the referral data is
// delivered to [OnReferralMessage] instead.
func (handler *Handler) OnContactsMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, contacts *message.Contacts) error,
) {
	handler.messages.Contacts = MessageHandlerFunc[message.Contacts](fn)
}

// SetContactsMessageHandler sets the handler for contacts messages.
func (handler *Handler) SetContactsMessageHandler(
	h ContactsMessageHandler,
) {
	handler.messages.Contacts = h
}

// OnReactionMessage registers a handler for reaction messages in the messages webhook.
func (handler *Handler) OnReactionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, reaction *media.Reaction) error,
) {
	handler.messages.Reaction = MessageHandlerFunc[media.Reaction](fn)
}

// SetReactionMessageHandler sets the handler for reaction messages.
func (handler *Handler) SetReactionMessageHandler(
	h ReactionHandler,
) {
	handler.messages.Reaction = h
}

// OnProductEnquiryMessage registers a handler for product enquiry messages in the messages webhook.
func (handler *Handler) OnProductEnquiryMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, txt *Text) error,
) {
	handler.messages.ProductInquiry = MessageHandlerFunc[Text](fn)
}

// SetProductEnquiryMessageHandler sets the handler for product enquiry messages.
func (handler *Handler) SetProductEnquiryMessageHandler(
	h ProductEnquiryHandler,
) {
	handler.messages.ProductInquiry = h
}

// OnInteractiveMessage registers a handler for interactive messages in the messages webhook.
func (handler *Handler) OnInteractiveMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, itv *Interactive) error,
) {
	handler.messages.Interactive.Fallback = MessageHandlerFunc[Interactive](fn)
}

// SetInteractiveMessageHandler sets the handler for interactive messages.
func (handler *Handler) SetInteractiveMessageHandler(
	h InteractiveMessageHandler,
) {
	handler.messages.Interactive.Fallback = h
}

// OnButtonReplyMessage registers a handler for button reply messages in the messages webhook.
func (handler *Handler) OnButtonReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, br *ButtonReply) error,
) {
	handler.messages.Interactive.ButtonReply = MessageHandlerFunc[ButtonReply](fn)
}

// SetButtonReplyMessageHandler sets the handler for button reply messages.
func (handler *Handler) SetButtonReplyMessageHandler(
	h ButtonReplyMessageHandler,
) {
	handler.messages.Interactive.ButtonReply = h
}

// OnListReplyMessage registers a handler for list reply messages in the messages webhook.
func (handler *Handler) OnListReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, lr *ListReply) error,
) {
	handler.messages.Interactive.ListReply = MessageHandlerFunc[ListReply](fn)
}

// SetListReplyMessageHandler sets the handler for list reply messages.
func (handler *Handler) SetListReplyMessageHandler(
	h ListReplyMessageHandler,
) {
	handler.messages.Interactive.ListReply = h
}

// OnFlowCompletionMessage registers a handler for flow completion messages in the messages webhook.
func (handler *Handler) OnFlowCompletionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, nfm *NFMReply) error,
) {
	handler.messages.Interactive.FlowCompletion = MessageHandlerFunc[NFMReply](fn)
}

// SetFlowCompletionMessageHandler sets the handler for flow completion messages.
func (handler *Handler) SetFlowCompletionMessageHandler(
	h NativeFlowCompletionHandler,
) {
	handler.messages.Interactive.FlowCompletion = h
}

// OnAddressSubmissionMessage registers a handler for address submission messages in the messages webhook.
func (handler *Handler) OnAddressSubmissionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, nfm *NFMReply) error,
) {
	handler.messages.Interactive.AddressSubmission = MessageHandlerFunc[NFMReply](fn)
}

// SetAddressSubmissionHandler sets the handler for address submission messages.
func (handler *Handler) SetAddressSubmissionHandler(
	h NativeFlowCompletionHandler,
) {
	handler.messages.Interactive.AddressSubmission = h
}

// OnReferralMessage registers a handler for messages originating from Click to
// WhatsApp ads. The referral object carries the ad ID, source URL, headline,
// body, media URLs, and click tracking ID (ctwa_clid). Present on any incoming
// message type (text, image, contacts, etc.) sent via a Click to WhatsApp ad.
func (handler *Handler) OnReferralMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, ref *ReferralNotification) error,
) {
	handler.messages.Referral = MessageHandlerFunc[ReferralNotification](fn)
}

// SetReferralMessageHandler sets the handler for referral messages.
func (handler *Handler) SetReferralMessageHandler(
	h ReferralMessageHandler,
) {
	handler.messages.Referral = h
}

// OnCustomerIDChangeMessage registers a handler for customer identity change messages in the messages webhook.
func (handler *Handler) OnCustomerIDChangeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, identity *Identity) error,
) {
	handler.messages.CustomerIDChange = MessageHandlerFunc[Identity](fn)
}

// SetCustomerIDChangeMessageHandler sets the handler for customer identity change messages.
func (handler *Handler) SetCustomerIDChangeMessageHandler(
	h CustomerIDChangeHandler,
) {
	handler.messages.CustomerIDChange = h
}

// OnSystemMessage registers a handler for system messages in the messages webhook.
func (handler *Handler) OnSystemMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, sys *System) error,
) {
	handler.messages.System = MessageHandlerFunc[System](fn)
}

// SetSystemMessageHandler sets the handler for system messages.
func (handler *Handler) SetSystemMessageHandler(
	h SystemMessageHandler,
) {
	handler.messages.System = h
}

// OnAudioMessage registers a handler for audio messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and download URL.
// For voice messages, check the voice field via the Media API.
func (handler *Handler) OnAudioMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *media.Info) error,
) {
	handler.messages.Media.Audio = MessageHandlerFunc[media.Info](fn)
}

// SetAudioMessageHandler sets the handler for audio messages.
func (handler *Handler) SetAudioMessageHandler(
	h MediaMessageHandler,
) {
	handler.messages.Media.Audio = h
}

// OnVideoMessage registers a handler for video messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+). Use the ID with the Media API or the URL directly with
// your access token to retrieve the asset.
func (handler *Handler) OnVideoMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *media.Info) error,
) {
	handler.messages.Media.Video = MessageHandlerFunc[media.Info](fn)
}

// SetVideoMessageHandler sets the handler for video messages.
func (handler *Handler) SetVideoMessageHandler(
	h MediaMessageHandler,
) {
	handler.messages.Media.Video = h
}

// OnImageMessage registers a handler for image messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+).
func (handler *Handler) OnImageMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *media.Info) error,
) {
	handler.messages.Media.Image = MessageHandlerFunc[media.Info](fn)
}

// SetImageMessageHandler sets the handler for image messages.
func (handler *Handler) SetImageMessageHandler(
	h MediaMessageHandler,
) {
	handler.messages.Media.Image = h
}

// OnDocumentMessage registers a handler for document messages in the messages
// webhook. Metadata includes filename, MIME type, SHA-256 hash, caption, and
// download URL.
func (handler *Handler) OnDocumentMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *media.Info) error,
) {
	handler.messages.Media.Document = MessageHandlerFunc[media.Info](fn)
}

// SetDocumentMessageHandler sets the handler for document messages.
func (handler *Handler) SetDocumentMessageHandler(
	h MediaMessageHandler,
) {
	handler.messages.Media.Document = h
}

// OnStickerMessage registers a handler for sticker messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and an animated flag.
func (handler *Handler) OnStickerMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *media.Info) error,
) {
	handler.messages.Media.Sticker = MessageHandlerFunc[media.Info](fn)
}

// SetStickerMessageHandler sets the handler for sticker messages.
func (handler *Handler) SetStickerMessageHandler(
	h MediaMessageHandler,
) {
	handler.messages.Media.Sticker = h
}

// OnRevokeMessage registers a callback for message deletion events. Triggers
// when a WhatsApp user deletes a previously sent message (within ~2 days of
// sending). The callback receives the original message ID that was revoked.
func (handler *Handler) OnRevokeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, revoke *Revoke) error,
) {
	handler.messages.Revoke = MessageHandlerFunc[Revoke](fn)
}

// SetRevokeMessageHandler sets the handler for message revoke events.
func (handler *Handler) SetRevokeMessageHandler(h MessageHandler[Revoke]) {
	handler.messages.Revoke = h
}

// OnMessageEdit registers a callback for edit events. Triggers when a WhatsApp
// user edits a previously sent message (text or media caption) within 15 minutes
// of sending. The callback receives the original message ID and the replacement
// content.
//
// Note: edit webhooks are temporarily unsupported by WhatsApp. Edited messages
// may currently arrive as unsupported message type instead of edit type.
func (handler *Handler) OnMessageEdit(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, edit *Edit) error,
) {
	handler.messages.Edit = MessageHandlerFunc[Edit](fn)
}

// SetMessageEditHandler sets the handler for message edit events.
func (handler *Handler) SetMessageEditHandler(h MessageHandler[Edit]) {
	handler.messages.Edit = h
}
