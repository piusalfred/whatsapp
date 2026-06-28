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

// Message types, MessageHandler interface, MessagesHandler dispatch, MediaHandler,
// InteractiveHandler, type aliases, and all On/Set registration methods for
// message webhooks., the MessageHandler
// and MessageErrorsHandler interfaces, type aliases for per-message-type
// handlers, and all On/Set registration methods on Handler for message
// webhooks.

package webhooks

import (
	"context"
	"fmt"
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
// OnRequestWelcomeMessage registers a handler for request_welcome messages in the messages webhook.
func (handler *Handler) OnRequestWelcomeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *Message) error,
) {
	handler.messages.RequestWelcome = MessageHandlerFunc[Message](fn)
}

// SetRequestWelcomeMessageHandler sets the handler for request_welcome messages.
func (handler *Handler) SetRequestWelcomeMessageHandler(
	fn RequestWelcomeMessageHandler,
) {
	handler.messages.RequestWelcome = fn
}
// Each field is a handler for one message type or sub-type. Leave a field nil
// to silently skip that type during dispatch.
//
// Media types delegate to [MediaHandler]; interactive messages delegate to
// [InteractiveHandler]. Unknown types fall through to [Fallback].
//
// Usage:
//
//	mh := &MessagesHandler{}
//	mh.Media = &MediaHandler{}          // enable media dispatch
//	mh.Interactive = &InteractiveHandler{}
//	mh.Text = myTextHandler
//	mh.Fallback = catchAll              // handle anything not explicitly registered
//	h := webhooks.NewHandler()
//	h.SetMessagesHandler(mh)            // wire into Handler (future API)
type MessagesHandler struct {
	Media            *MediaHandler
	Interactive      *InteractiveHandler
	Text             MessageHandler[Text]
	Order            MessageHandler[Order]
	Button           MessageHandler[Button]
	System           MessageHandler[System]
	Reaction         MessageHandler[media.Reaction]
	Location         MessageHandler[media.Location]
	Contacts         MessageHandler[message.Contacts]
	Revoke           MessageHandler[Revoke]
	Edit             MessageHandler[Edit]
	RequestWelcome   MessageHandler[Message]
	Referral         MessageHandler[ReferralNotification]
	CustomerIDChange MessageHandler[Identity]
	Unknown          MessageErrorsHandler
	Unsupported      MessageErrorsHandler
	ProductInquiry   MessageHandler[Text]
	Fallback         MessageHandler[Message]
}

//nolint:cyclop,funlen,gocognit,gocyclo,wrapcheck // dispatch switch; user handlers own error context
func (mh *MessagesHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
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

	msgType := ParseMessageType(message.Type)
	switch msgType {
	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return mh.HandleMediaMessage(ctx, nctx, info, message)
	case MessageTypeInteractive:
		return mh.HandleInteractiveMessage(ctx, nctx, info, message)
	case MessageTypeText:
		return mh.handleText(ctx, nctx, info, message)
	case MessageTypeSystem:
		return mh.handleSystem(ctx, nctx, info, message)
	case MessageTypeOrder:
		if mh.Order == nil {
			return nil
		}
		if err := mh.Order.Handle(ctx, nctx, info, message.Order); err != nil {
			return fmt.Errorf("handle order message: %w", err)
		}
		return nil
	case MessageTypeButton:
		if mh.Button == nil {
			return nil
		}
		if err := mh.Button.Handle(ctx, nctx, info, message.Button); err != nil {
			return fmt.Errorf("handle button message: %w", err)
		}
		return nil
	case MessageTypeReaction:
		if mh.Reaction == nil {
			return nil
		}
		if err := mh.Reaction.Handle(ctx, nctx, info, message.Reaction); err != nil {
			return fmt.Errorf("handle reaction message: %w", err)
		}
		return nil
	case MessageTypeLocation:
		if mh.Location == nil {
			return nil
		}
		if err := mh.Location.Handle(ctx, nctx, info, message.Location); err != nil {
			return fmt.Errorf("handle location message: %w", err)
		}
		return nil
	case MessageTypeContacts:
		if mh.Contacts == nil {
			return nil
		}
		if err := mh.Contacts.Handle(ctx, nctx, info, message.Contacts); err != nil {
			return fmt.Errorf("handle contacts message: %w", err)
		}
		return nil
	case MessageTypeRevoke:
		if mh.Revoke == nil {
			return nil
		}
		if err := mh.Revoke.Handle(ctx, nctx, info, message.Revoke); err != nil {
			return fmt.Errorf("handle revoke message: %w", err)
		}
		return nil
	case MessageTypeEdit:
		if mh.Edit == nil {
			return nil
		}
		if err := mh.Edit.Handle(ctx, nctx, info, message.Edit); err != nil {
			return fmt.Errorf("handle edit message: %w", err)
		}
		return nil
	case MessageTypeRequestWelcome:
		if mh.RequestWelcome == nil {
			return nil
		}
		if err := mh.RequestWelcome.Handle(ctx, nctx, info, message); err != nil {
			return fmt.Errorf("handle request welcome: %w", err)
		}
		return nil
	case MessageTypeUnknown:
		if mh.Unknown == nil {
			return nil
		}
		if err := mh.Unknown.Handle(ctx, nctx, info, message.Errors); err != nil {
			return fmt.Errorf("handle error message: %w", err)
		}
		return nil
	case MessageTypeUnsupported:
		if mh.Unsupported == nil {
			return nil
		}
		if err := mh.Unsupported.Handle(ctx, nctx, info, message.Errors); err != nil {
			return fmt.Errorf("handle unsupported message: %w", err)
		}
		return nil
	default:
		if message.Contacts != nil && mh.Contacts != nil {
			if err := mh.Contacts.Handle(ctx, nctx, info, message.Contacts); err != nil {
				return fmt.Errorf("handle contacts message: %w", err)
			}
			return nil
		}
		if message.Location != nil && mh.Location != nil {
			if err := mh.Location.Handle(ctx, nctx, info, message.Location); err != nil {
				return fmt.Errorf("handle location message: %w", err)
			}
			return nil
		}
		if message.Identity != nil && mh.CustomerIDChange != nil {
			if err := mh.CustomerIDChange.Handle(ctx, nctx, info, message.Identity); err != nil {
				return fmt.Errorf("handle customer ID change: %w", err)
			}
			return nil
		}
		if mh.Fallback != nil {
			return mh.Fallback.Handle(ctx, nctx, info, message)
		}
		return nil
	}
}

func (mh *MessagesHandler) handleText(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	if info.IsReferral && mh.Referral != nil {
		ref := &ReferralNotification{Text: message.Text, Referral: message.Referral}
		if err := mh.Referral.Handle(ctx, nctx, info, ref); err != nil {
			return fmt.Errorf("handle referral message: %w", err)
		}
		return nil
	}
	if info.IsProductInquiry && mh.ProductInquiry != nil {
		if err := mh.ProductInquiry.Handle(ctx, nctx, info, message.Text); err != nil {
			return fmt.Errorf("handle product inquiry: %w", err)
		}
		return nil
	}
	if mh.Text == nil {
		return nil
	}
	if err := mh.Text.Handle(ctx, nctx, info, message.Text); err != nil {
		return fmt.Errorf("handle text message: %w", err)
	}
	return nil
}

func (mh *MessagesHandler) handleSystem(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	if message.System != nil && message.System.Type == "user_changed_number" && mh.CustomerIDChange != nil {
		if err := mh.CustomerIDChange.Handle(ctx, nctx, info, message.Identity); err != nil {
			return fmt.Errorf("handle customer ID change: %w", err)
		}
		return nil
	}
	if mh.System == nil {
		return nil
	}
	if err := mh.System.Handle(ctx, nctx, info, message.System); err != nil {
		return fmt.Errorf("handle system message: %w", err)
	}
	return nil
}

// HandleMediaMessage dispatches the media message to [MediaHandler] if set.
// Returns nil when mh.Media is nil.
func (mh *MessagesHandler) HandleMediaMessage(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, msg *Message,
) error {
	if mh.Media == nil {
		return nil
	}

	if err := mh.Media.Handle(ctx, nctx, info, msg); err != nil {
		return fmt.Errorf("handle media message: %w", err)
	}

	return nil
}

// HandleInteractiveMessage dispatches the interactive message to
// [InteractiveHandler] if set. Returns nil when mh.Interactive is nil.
func (mh *MessagesHandler) HandleInteractiveMessage(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, msg *Message,
) error {
	if mh.Interactive == nil {
		return nil
	}

	if err := mh.Interactive.Handle(ctx, nctx, info, msg); err != nil {
		return fmt.Errorf("handle interactive message: %w", err)
	}

	return nil
}

// MediaHandler groups handlers for media message types (audio, video, image,
// document, sticker). All fields are [MessageHandler[media.Info]]. Leave a
// field nil to silently skip that media type during dispatch.
//
// MediaHandler has no fallback — unhandled media types return an error.
// A catch-all for unrecognized message types belongs in the message-level
// handler, not here.
//
// Usage:
//
//	mh := &MediaHandler{}
//	mh.OnAudio(myAudioHandler)
type MediaHandler struct {
	Audio    MessageHandler[media.Info]
	Video    MessageHandler[media.Info]
	Image    MessageHandler[media.Info]
	Document MessageHandler[media.Info]
	Sticker  MessageHandler[media.Info]
}

// OnAudio sets the handler for audio messages.
func (mh *MediaHandler) OnAudio(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Audio = MessageHandlerFunc[media.Info](fn)
}

// OnVideo sets the handler for video messages.
func (mh *MediaHandler) OnVideo(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Video = MessageHandlerFunc[media.Info](fn)
}

// OnImage sets the handler for image messages.
func (mh *MediaHandler) OnImage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Image = MessageHandlerFunc[media.Info](fn)
}

// OnDocument sets the handler for document messages.
func (mh *MediaHandler) OnDocument(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Document = MessageHandlerFunc[media.Info](fn)
}

// OnSticker sets the handler for sticker messages.
func (mh *MediaHandler) OnSticker(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *media.Info) error,
) {
	mh.Sticker = MessageHandlerFunc[media.Info](fn)
}

// Handle dispatches the media message to the correct handler based on msg.Type.
//
// If the dedicated handler is nil, the message is silently skipped. Unknown
// media types return an error — they should have been caught earlier by the
// message dispatch switch.
//
//nolint:gocognit // typed dispatch switch
func (mh *MediaHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	info *MessageInfo,
	msg *Message,
) error {
	mt := ParseMessageType(msg.Type)
	switch mt {
	case MessageTypeAudio:
		if mh.Audio == nil {
			return nil
		}
		if err := mh.Audio.Handle(ctx, nctx, info, msg.Audio); err != nil {
			return fmt.Errorf("handle audio message: %w", err)
		}

	case MessageTypeVideo:
		if mh.Video == nil {
			return nil
		}
		if err := mh.Video.Handle(ctx, nctx, info, msg.Video); err != nil {
			return fmt.Errorf("handle video message: %w", err)
		}

	case MessageTypeImage:
		if mh.Image == nil {
			return nil
		}
		if err := mh.Image.Handle(ctx, nctx, info, msg.Image); err != nil {
			return fmt.Errorf("handle image message: %w", err)
		}

	case MessageTypeDocument:
		if mh.Document == nil {
			return nil
		}
		if err := mh.Document.Handle(ctx, nctx, info, msg.Document); err != nil {
			return fmt.Errorf("handle document message: %w", err)
		}

	case MessageTypeSticker:
		if mh.Sticker == nil {
			return nil
		}
		if err := mh.Sticker.Handle(ctx, nctx, info, msg.Sticker); err != nil {
			return fmt.Errorf("handle sticker message: %w", err)
		}

	default:
		return fmt.Errorf("media message handler: %w: %s", ErrUnrecognizedMessageType, mt)
	}

	return nil
}

// InteractiveHandler groups handlers for interactive message sub-types
// (list reply, button reply, flow completion, address submission) plus a
// general handler for unrecognized interactive types.
//
// Leave a field nil to skip that subtype. The general [Interactive] handler
// catches any interactive type without a dedicated handler.
//
// Usage:
//
//	ih := &InteractiveHandler{}
//	ih.OnButtonReply(myButtonReplyHandler)
//	ih.OnFallback(myGeneralHandler) // catch-all for unrecognized types
type InteractiveHandler struct {
	ListReply         MessageHandler[ListReply]
	ButtonReply       MessageHandler[ButtonReply]
	FlowCompletion    MessageHandler[NFMReply]
	AddressSubmission MessageHandler[NFMReply]
	Fallback          MessageHandler[Interactive]
}

// OnListReply sets the handler for list reply interactive messages.
func (ih *InteractiveHandler) OnListReply(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *ListReply) error,
) {
	ih.ListReply = MessageHandlerFunc[ListReply](fn)
}

// OnButtonReply sets the handler for button reply interactive messages.
func (ih *InteractiveHandler) OnButtonReply(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *ButtonReply) error,
) {
	ih.ButtonReply = MessageHandlerFunc[ButtonReply](fn)
}

// OnFlowCompletion sets the handler for flow completion (nfm_reply) messages.
func (ih *InteractiveHandler) OnFlowCompletion(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *NFMReply) error,
) {
	ih.FlowCompletion = MessageHandlerFunc[NFMReply](fn)
}

// OnAddressSubmission sets the handler for address submission messages.
func (ih *InteractiveHandler) OnAddressSubmission(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *NFMReply) error,
) {
	ih.AddressSubmission = MessageHandlerFunc[NFMReply](fn)
}

// OnFallback sets the catch-all handler for interactive messages without
// a dedicated subtype handler.
func (ih *InteractiveHandler) OnFallback(
	fn func(ctx context.Context, nctx *MessageNotificationContext, info *MessageInfo, data *Interactive) error,
) {
	ih.Fallback = MessageHandlerFunc[Interactive](fn)
}

// Handle dispatches an interactive message to the correct handler based on
// msg.Interactive.Type. Subtypes without a dedicated handler fall through to
// the general [Interactive] handler. Messages with a nil Interactive payload
// are silently skipped.
//
//nolint:gocognit // typed dispatch switch
func (ih *InteractiveHandler) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	info *MessageInfo,
	msg *Message,
) error {
	if msg.Interactive == nil {
		return nil
	}

	switch msg.Interactive.Type {
	case InteractiveTypeListReply:
		if ih.ListReply != nil {
			if err := ih.ListReply.Handle(ctx, nctx, info, msg.Interactive.ListReply); err != nil {
				return fmt.Errorf("handle list reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeButtonReply:
		if ih.ButtonReply != nil {
			if err := ih.ButtonReply.Handle(ctx, nctx, info, msg.Interactive.ButtonReply); err != nil {
				return fmt.Errorf("handle button reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeNFMReply:
		if ih.FlowCompletion != nil {
			if err := ih.FlowCompletion.Handle(ctx, nctx, info, msg.Interactive.NFMReply); err != nil {
				return fmt.Errorf("handle flow completion: %w", err)
			}
			return nil
		}

	case InteractiveAddressSubmission:
		if ih.AddressSubmission != nil {
			if err := ih.AddressSubmission.Handle(ctx, nctx, info, msg.Interactive.NFMReply); err != nil {
				return fmt.Errorf("handle address submission: %w", err)
			}
			return nil
		}
	}

	if ih.Fallback != nil {
		if err := ih.Fallback.Handle(ctx, nctx, info, msg.Interactive); err != nil {
			return fmt.Errorf("handle interactive message: %w", err)
		}
	}

	return nil
}
