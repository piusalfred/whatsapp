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

// Message types and message dispatch for WhatsApp Messages webhooks.
// Handles all incoming WhatsApp Business Platform message notifications.

package webhooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/media"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

const (
	MessageTypeAudio            = "audio"
	MessageTypeButton           = "button"
	MessageTypeDocument         = "document"
	MessageTypeText             = "text"
	MessageTypeImage            = "image"
	MessageTypeInteractive      = "interactive"
	MessageTypeOrder            = "order"
	MessageTypeSticker          = "sticker"
	MessageTypeSystem           = "system"
	MessageTypeUnknown          = "unknown"
	MessageTypeUnsupported      = "unsupported"
	MessageTypeVideo            = "video"
	MessageTypeLocation         = "location"
	MessageTypeReaction         = "reaction"
	MessageTypeContacts         = "contacts"
	MessageTypeRequestWelcome   = "request_welcome"
	MessageTypeRevoke           = "revoke"
	MessageTypeEdit             = "edit"
	MessageTypeErrors           = "errors"
	MessageTypeGif              = "gif"
	MessageTypeGroupInvite      = "group_invite"
	MessageTypeHsm              = "hsm"
	MessageTypeKeepInChat       = "keep_in_chat"
	MessageTypeLinkPreview      = "link_preview"
	MessageTypeList             = "list"
	MessageTypeMediaPlaceholder = "media_placeholder"
	MessageTypePin              = "pin"
	MessageTypePollCreation     = "poll_creation"
	MessageTypePollUpdate       = "poll_update"
	MessageTypeProduct          = "product"
)

// MessageType represents the type of a WhatsApp message.
type MessageType string

func (mm MessageType) String() string {
	return string(mm)
}

//nolint:gochecknoglobals // read-only lookup table for message type parsing
var parseMessageTypeMap = map[string]MessageType{
	MessageTypeAudio:            MessageTypeAudio,
	MessageTypeButton:           MessageTypeButton,
	MessageTypeDocument:         MessageTypeDocument,
	MessageTypeText:             MessageTypeText,
	MessageTypeImage:            MessageTypeImage,
	MessageTypeInteractive:      MessageTypeInteractive,
	MessageTypeOrder:            MessageTypeOrder,
	MessageTypeSticker:          MessageTypeSticker,
	MessageTypeSystem:           MessageTypeSystem,
	MessageTypeUnknown:          MessageTypeUnknown,
	MessageTypeUnsupported:      MessageTypeUnsupported,
	MessageTypeVideo:            MessageTypeVideo,
	MessageTypeLocation:         MessageTypeLocation,
	MessageTypeReaction:         MessageTypeReaction,
	MessageTypeContacts:         MessageTypeContacts,
	MessageTypeRequestWelcome:   MessageTypeRequestWelcome,
	MessageTypeRevoke:           MessageTypeRevoke,
	MessageTypeEdit:             MessageTypeEdit,
	MessageTypeErrors:           MessageTypeErrors,
	MessageTypeGif:              MessageTypeGif,
	MessageTypeGroupInvite:      MessageTypeGroupInvite,
	MessageTypeHsm:              MessageTypeHsm,
	MessageTypeKeepInChat:       MessageTypeKeepInChat,
	MessageTypeLinkPreview:      MessageTypeLinkPreview,
	MessageTypeList:             MessageTypeList,
	MessageTypeMediaPlaceholder: MessageTypeMediaPlaceholder,
	MessageTypePin:              MessageTypePin,
	MessageTypePollCreation:     MessageTypePollCreation,
	MessageTypePollUpdate:       MessageTypePollUpdate,
	MessageTypeProduct:          MessageTypeProduct,
}

// ParseMessageType parses a message type string into a MessageType.
// It trims whitespace and performs case-insensitive matching.
func ParseMessageType(s string) MessageType {
	key := strings.TrimSpace(strings.ToLower(s))
	return parseMessageTypeMap[key]
}

// ErrUnrecognizedMessageType is returned when a message type is not recognized.
var ErrUnrecognizedMessageType = errors.New("unrecognized message type")

const (
	InteractiveTypeListReply           = "list_reply"
	InteractiveTypeButtonReply         = "button_reply"
	InteractiveTypeNFMReply            = "nfm_reply"
	InteractiveAddressSubmission       = "address_message"
	InteractiveTypeCallPermissionReply = "call_permission_reply"
)

type (
	MessageErrorsHandlerFunc func(ctx context.Context, req *MessageRequest[Message], errors []*werrors.Error) error
	MessageErrorsHandler     interface {
		Handle(ctx context.Context, req *MessageRequest[Message], errors []*werrors.Error) error
	}
)

func (fn MessageErrorsHandlerFunc) Handle(ctx context.Context,
	req *MessageRequest[Message], errors []*werrors.Error,
) error {
	return fn(ctx, req, errors)
}

func NewNoOpMessageErrorsHandler() MessageErrorsHandler {
	return MessageErrorsHandlerFunc(
		func(_ context.Context, _ *MessageRequest[Message], _ []*werrors.Error) error {
			return nil
		},
	)
}

type (
	MediaMessageHandler MessageHandler[media.Info]

	MessageHandlerFunc[T any] func(ctx context.Context, req *MessageRequest[T]) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, req *MessageRequest[T]) error
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
	CallPermissionReplyHandler   = MessageHandler[CallPermissionReply]
	ReferralMessageHandler       = MessageHandler[ReferralNotification]
	CustomerIDChangeHandler      = MessageHandler[Identity]
	SystemMessageHandler         = MessageHandler[System]
	RequestWelcomeMessageHandler = MessageHandler[Message]
)

func (fn MessageHandlerFunc[T]) Handle(ctx context.Context, req *MessageRequest[T]) error {
	return fn(ctx, req)
}

func NewNoOpMessageHandler[T any]() MessageHandler[T] {
	return MessageHandlerFunc[T](func(_ context.Context, _ *MessageRequest[T]) error {
		return nil
	})
}

// OnTextMessage registers a handler for text messages in the messages webhook.
func (handler *Handler) OnTextMessage(h MessageHandler[Text]) {
	handler.ensureMessages().Text = h
}

// OnButtonMessage registers a handler for button messages in the messages webhook.
func (handler *Handler) OnButtonMessage(h ButtonMessageHandler) {
	handler.ensureMessages().Button = h
}

// OnOrderMessage registers a handler for order messages in the messages webhook.
func (handler *Handler) OnOrderMessage(h OrderMessageHandler) {
	handler.ensureMessages().Order = h
}

// OnLocationMessage registers a handler for shared location messages in the
// messages webhook. The location includes latitude, longitude, name, address,
// and optionally a URL (usually only for business locations).
func (handler *Handler) OnLocationMessage(h LocationMessageHandler) {
	handler.ensureMessages().Location = h
}

// OnContactsMessage registers a handler for contacts messages in the messages
// webhook. Contact payloads contain name, phone, email, address, URL, and
// organization fields — all optional since the WhatsApp user chooses what to
// share. If the message came via a Click to WhatsApp ad, the referral data is
// delivered to [OnReferralMessage] instead.
func (handler *Handler) OnContactsMessage(h ContactsMessageHandler) {
	handler.ensureMessages().Contacts = h
}

// OnReactionMessage registers a handler for reaction messages in the messages webhook.
func (handler *Handler) OnReactionMessage(h ReactionHandler) {
	handler.ensureMessages().Reaction = h
}

// OnProductEnquiryMessage registers a handler for product enquiry messages in the messages webhook.
func (handler *Handler) OnProductEnquiryMessage(h ProductEnquiryHandler) {
	handler.ensureMessages().ProductInquiry = h
}

// OnInteractiveMessage registers a handler for interactive messages in the messages webhook.
func (handler *Handler) OnInteractiveMessage(h InteractiveMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.Fallback = h
}

// OnButtonReplyMessage registers a handler for button reply messages in the messages webhook.
func (handler *Handler) OnButtonReplyMessage(h ButtonReplyMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.ButtonReply = h
}

// OnListReplyMessage registers a handler for list reply messages in the messages webhook.
func (handler *Handler) OnListReplyMessage(h ListReplyMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.ListReply = h
}

// OnFlowCompletionMessage registers a handler for flow completion messages in the messages webhook.
func (handler *Handler) OnFlowCompletionMessage(h NativeFlowCompletionHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.FlowCompletion = h
}

// OnAddressSubmissionMessage registers a handler for address submission messages in the messages webhook.
func (handler *Handler) OnAddressSubmissionMessage(h NativeFlowCompletionHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.AddressSubmission = h
}

// OnCallPermissionReply registers a handler for call permission reply messages
// in the messages webhook. Call permission replies arrive as interactive
// messages with type "call_permission_reply" when a WhatsApp user accepts or
// rejects a call permission request, or when permission is automatically
// granted by initiating a call.
func (handler *Handler) OnCallPermissionReply(h CallPermissionReplyHandler) {
	mh := handler.ensureMessages()
	if mh.Interactive == nil {
		mh.Interactive = &InteractiveHandler{}
	}
	mh.Interactive.CallPermissionReply = h
}

// OnReferralMessage registers a handler for messages originating from Click to
// WhatsApp ads. The referral object carries the ad ID, source URL, headline,
// body, media URLs, and click tracking ID (ctwa_clid). Present on any incoming
// message type (text, image, contacts, etc.) sent via a Click to WhatsApp ad.
func (handler *Handler) OnReferralMessage(h ReferralMessageHandler) {
	handler.ensureMessages().Referral = h
}

// OnCustomerIDChangeMessage registers a handler for customer identity change messages in the messages webhook.
func (handler *Handler) OnCustomerIDChangeMessage(h CustomerIDChangeHandler) {
	handler.ensureMessages().CustomerIDChange = h
}

// OnSystemMessage registers a handler for system messages in the messages webhook.
func (handler *Handler) OnSystemMessage(h SystemMessageHandler) {
	handler.ensureMessages().System = h
}

// OnAudioMessage registers a handler for audio messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and download URL.
// For voice messages, check the voice field via the Media API.
func (handler *Handler) OnAudioMessage(h MediaMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Media == nil {
		mh.Media = &MediaHandler{}
	}
	mh.Media.Audio = h
}

// OnVideoMessage registers a handler for video messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+). Use the ID with the Media API or the URL directly with
// your access token to retrieve the asset.
func (handler *Handler) OnVideoMessage(h MediaMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Media == nil {
		mh.Media = &MediaHandler{}
	}
	mh.Media.Video = h
}

// OnImageMessage registers a handler for image messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, caption, and download
// URL (v2025.11+).
func (handler *Handler) OnImageMessage(h MediaMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Media == nil {
		mh.Media = &MediaHandler{}
	}
	mh.Media.Image = h
}

// OnDocumentMessage registers a handler for document messages in the messages
// webhook. Metadata includes filename, MIME type, SHA-256 hash, caption, and
// download URL.
func (handler *Handler) OnDocumentMessage(h MediaMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Media == nil {
		mh.Media = &MediaHandler{}
	}
	mh.Media.Document = h
}

// OnStickerMessage registers a handler for sticker messages in the messages
// webhook. Metadata includes MIME type, SHA-256 hash, and an animated flag.
func (handler *Handler) OnStickerMessage(h MediaMessageHandler) {
	mh := handler.ensureMessages()
	if mh.Media == nil {
		mh.Media = &MediaHandler{}
	}
	mh.Media.Sticker = h
}

// OnRevokeMessage registers a callback for message deletion events. Triggers
// when a WhatsApp user deletes a previously sent message (within ~2 days of
// sending). The callback receives the original message ID that was revoked.
func (handler *Handler) OnRevokeMessage(h MessageHandler[Revoke]) {
	handler.ensureMessages().Revoke = h
}

// OnMessageEdit registers a callback for edit events. Triggers when a WhatsApp
// user edits a previously sent message (text or media caption) within 15 minutes
// of sending. The callback receives the original message ID and the replacement
// content.
//
// Note: edit webhooks are temporarily unsupported by WhatsApp. Edited messages
// may currently arrive as unsupported message type instead of edit type.
func (handler *Handler) OnMessageEdit(h MessageHandler[Edit]) {
	handler.ensureMessages().Edit = h
}

// OnRequestWelcomeMessage registers a handler for request_welcome messages in the messages webhook.
func (handler *Handler) OnRequestWelcomeMessage(h RequestWelcomeMessageHandler) {
	handler.ensureMessages().RequestWelcome = h
}

// MessagesHandler is the single entry point for all WhatsApp message dispatch.
// Each field accepts a [MessageHandler] for one message type or sub-type. Leave
// a field nil to silently skip that type during dispatch.
//
// Media types delegate to [MediaHandler]; interactive messages delegate to
// [InteractiveHandler]. Unknown types fall through to [Fallback].
//
// Usage:
//
//	mh := &MessagesHandler{}
//	mh.Media = &MediaHandler{}
//	mh.Interactive = &InteractiveHandler{}
//	mh.Text = myTextHandler
//	mh.Fallback = catchAll
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
	Fallback         FallbackHandler
}

// newMessageInfo extracts MessageInfo from a raw Message. All boolean
// flags are computed via Message methods (IsAReply, IsForwarded, etc.).
func newMessageInfo(msg *Message) *MessageInfo {
	return &MessageInfo{
		From:             msg.From,
		MessageID:        msg.ID,
		Timestamp:        msg.Timestamp,
		Type:             msg.Type,
		GroupID:          msg.GroupID,
		Context:          msg.Context,
		IsAReply:         msg.IsAReply(),
		IsForwarded:      msg.IsForwarded(),
		IsProductInquiry: msg.IsProductInquiry(),
		IsReferral:       msg.IsReferral(),
	}
}

// newMessageRequest is a helper to build a MessageRequest for dispatch.
func newMessageRequest[T any](nctx *MessageNotificationContext, info *MessageInfo, payload *T) *MessageRequest[T] {
	return &MessageRequest[T]{Notification: nctx, Info: info, Payload: payload}
}

// Handle dispatches every message in the change to the correct typed handler.
// It matches the unified sub-handler signature used by Groups, Business,
// Flows, and History.
func (mh *MessagesHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	nctx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Contacts:           change.Value.Contacts,
		Metadata:           change.Value.Metadata,
	}

	for _, msg := range change.Value.Messages {
		if msg == nil {
			continue
		}
		if err := mh.handleOne(ctx, nctx, msg); err != nil {
			return err
		}
	}

	return nil
}

// handleOne dispatches a single message to the correct typed handler based
// on msg.Type. Media types delegate to MediaHandler; interactive messages
// delegate to InteractiveHandler. Unknown types fall through to Fallback.
//
//nolint:cyclop,funlen,gocognit,gocyclo // dispatch switch
func (mh *MessagesHandler) handleOne(
	ctx context.Context,
	nctx *MessageNotificationContext,
	message *Message,
) error {
	info := newMessageInfo(message)

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
		if err := mh.Order.Handle(ctx, newMessageRequest(nctx, info, message.Order)); err != nil {
			return fmt.Errorf("handle order message: %w", err)
		}
		return nil
	case MessageTypeButton:
		if mh.Button == nil {
			return nil
		}
		if err := mh.Button.Handle(ctx, newMessageRequest(nctx, info, message.Button)); err != nil {
			return fmt.Errorf("handle button message: %w", err)
		}
		return nil
	case MessageTypeReaction:
		if mh.Reaction == nil {
			return nil
		}
		if err := mh.Reaction.Handle(ctx, newMessageRequest(nctx, info, message.Reaction)); err != nil {
			return fmt.Errorf("handle reaction message: %w", err)
		}
		return nil
	case MessageTypeLocation:
		if mh.Location == nil {
			return nil
		}
		if err := mh.Location.Handle(ctx, newMessageRequest(nctx, info, message.Location)); err != nil {
			return fmt.Errorf("handle location message: %w", err)
		}
		return nil
	case MessageTypeContacts:
		if mh.Contacts == nil {
			return nil
		}
		if err := mh.Contacts.Handle(ctx, newMessageRequest(nctx, info, message.Contacts)); err != nil {
			return fmt.Errorf("handle contacts message: %w", err)
		}
		return nil
	case MessageTypeRevoke:
		if mh.Revoke == nil {
			return nil
		}
		if err := mh.Revoke.Handle(ctx, newMessageRequest(nctx, info, message.Revoke)); err != nil {
			return fmt.Errorf("handle revoke message: %w", err)
		}
		return nil
	case MessageTypeEdit:
		if mh.Edit == nil {
			return nil
		}
		if err := mh.Edit.Handle(ctx, newMessageRequest(nctx, info, message.Edit)); err != nil {
			return fmt.Errorf("handle edit message: %w", err)
		}
		return nil
	case MessageTypeRequestWelcome:
		if mh.RequestWelcome == nil {
			return nil
		}
		if err := mh.RequestWelcome.Handle(ctx, newMessageRequest(nctx, info, message)); err != nil {
			return fmt.Errorf("handle request welcome: %w", err)
		}
		return nil
	case MessageTypeUnknown:
		if mh.Unknown == nil {
			return nil
		}
		req := &MessageRequest[Message]{
			Notification: nctx,
			Info:         info,
			Payload:      message,
		}
		if err := mh.Unknown.Handle(ctx, req, message.Errors); err != nil {
			return fmt.Errorf("handle error message: %w", err)
		}
		return nil
	case MessageTypeUnsupported:
		if mh.Unsupported == nil {
			return nil
		}
		req := &MessageRequest[Message]{
			Notification: nctx,
			Info:         info,
			Payload:      message,
		}
		if err := mh.Unsupported.Handle(ctx, req, message.Errors); err != nil {
			return fmt.Errorf("handle unsupported message: %w", err)
		}
		return nil
	default:
		// Messages without a recognized type field are probed for known
		// payload fields. Contacts, location, and identity historically
		// arrive without an explicit type in some WhatsApp webhook payloads.
		if message.Contacts != nil && mh.Contacts != nil {
			if err := mh.Contacts.Handle(ctx, newMessageRequest(nctx, info, message.Contacts)); err != nil {
				return fmt.Errorf("handle contacts message: %w", err)
			}
			return nil
		}
		if message.Location != nil && mh.Location != nil {
			if err := mh.Location.Handle(ctx, newMessageRequest(nctx, info, message.Location)); err != nil {
				return fmt.Errorf("handle location message: %w", err)
			}
			return nil
		}
		if message.Identity != nil && mh.CustomerIDChange != nil {
			if err := mh.CustomerIDChange.Handle(ctx, newMessageRequest(nctx, info, message.Identity)); err != nil {
				return fmt.Errorf("handle customer ID change: %w", err)
			}
			return nil
		}
		if mh.Fallback != nil {
			ne := NotificationEntry{
				Object: nctx.NotificationObject,
				ID:     nctx.EntryID,
				Time:   nctx.EntryTime,
			}
			change := Change{
				Field: "messages",
				Value: &Value{
					MessagingProduct: nctx.MessagingProduct,
					Metadata:         nctx.Metadata,
					Contacts:         nctx.Contacts,
					Messages:         []*Message{message},
				},
			}
			if err := mh.Fallback.Handle(ctx, ne, change); err != nil {
				return fmt.Errorf("handle fallback: %w", err)
			}
		}
		return nil
	}
}

func (mh *MessagesHandler) handleText(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	// Referrals take priority over product inquiries because a single text
	// message can carry both a referral (from an ad) and a referred_product
	// context. Check referral first — if the message came from a Click-to-
	// WhatsApp ad, the referral handler receives the full text+referral.
	if info.IsReferral && mh.Referral != nil {
		ref := &ReferralNotification{Text: message.Text, Referral: message.Referral}
		if err := mh.Referral.Handle(ctx, newMessageRequest(nctx, info, ref)); err != nil {
			return fmt.Errorf("handle referral message: %w", err)
		}
		return nil
	}
	if info.IsProductInquiry && mh.ProductInquiry != nil {
		if err := mh.ProductInquiry.Handle(ctx, newMessageRequest(nctx, info, message.Text)); err != nil {
			return fmt.Errorf("handle product inquiry: %w", err)
		}
		return nil
	}
	if mh.Text == nil {
		return nil
	}
	if err := mh.Text.Handle(ctx, newMessageRequest(nctx, info, message.Text)); err != nil {
		return fmt.Errorf("handle text message: %w", err)
	}
	return nil
}

func (mh *MessagesHandler) handleSystem(ctx context.Context, nctx *MessageNotificationContext,
	info *MessageInfo, message *Message,
) error {
	if message.System != nil && message.System.Type == "user_changed_number" && mh.CustomerIDChange != nil {
		if err := mh.CustomerIDChange.Handle(ctx, newMessageRequest(nctx, info, message.Identity)); err != nil {
			return fmt.Errorf("handle customer ID change: %w", err)
		}
		return nil
	}
	if mh.System == nil {
		return nil
	}
	if err := mh.System.Handle(ctx, newMessageRequest(nctx, info, message.System)); err != nil {
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
// Metadata includes MIME type, SHA-256 hash, and a download URL.
// For voice messages, check the voice field via the Media API.
func (mh *MediaHandler) OnAudio(h MediaMessageHandler) {
	mh.Audio = h
}

// OnVideo sets the handler for video messages.
// Metadata includes MIME type, SHA-256 hash, caption, and download URL.
func (mh *MediaHandler) OnVideo(h MediaMessageHandler) {
	mh.Video = h
}

// OnImage sets the handler for image messages.
// Metadata includes MIME type, SHA-256 hash, caption, and download URL.
func (mh *MediaHandler) OnImage(h MediaMessageHandler) {
	mh.Image = h
}

// OnDocument sets the handler for document messages.
// Metadata includes filename, MIME type, SHA-256 hash, caption, and download URL.
func (mh *MediaHandler) OnDocument(h MediaMessageHandler) {
	mh.Document = h
}

// OnSticker sets the handler for sticker messages.
// Metadata includes MIME type, SHA-256 hash, and an animated flag.
func (mh *MediaHandler) OnSticker(h MediaMessageHandler) {
	mh.Sticker = h
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
		if err := mh.Audio.Handle(ctx, newMessageRequest(nctx, info, msg.Audio)); err != nil {
			return fmt.Errorf("handle audio message: %w", err)
		}

	case MessageTypeVideo:
		if mh.Video == nil {
			return nil
		}
		if err := mh.Video.Handle(ctx, newMessageRequest(nctx, info, msg.Video)); err != nil {
			return fmt.Errorf("handle video message: %w", err)
		}

	case MessageTypeImage:
		if mh.Image == nil {
			return nil
		}
		if err := mh.Image.Handle(ctx, newMessageRequest(nctx, info, msg.Image)); err != nil {
			return fmt.Errorf("handle image message: %w", err)
		}

	case MessageTypeDocument:
		if mh.Document == nil {
			return nil
		}
		if err := mh.Document.Handle(ctx, newMessageRequest(nctx, info, msg.Document)); err != nil {
			return fmt.Errorf("handle document message: %w", err)
		}

	case MessageTypeSticker:
		if mh.Sticker == nil {
			return nil
		}
		if err := mh.Sticker.Handle(ctx, newMessageRequest(nctx, info, msg.Sticker)); err != nil {
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
	ListReply           MessageHandler[ListReply]
	ButtonReply         MessageHandler[ButtonReply]
	FlowCompletion      MessageHandler[NFMReply]
	AddressSubmission   MessageHandler[NFMReply]
	CallPermissionReply MessageHandler[CallPermissionReply]
	Fallback            MessageHandler[Interactive]
}

// OnListReply sets the handler for list reply interactive messages.
func (ih *InteractiveHandler) OnListReply(h ListReplyMessageHandler) {
	ih.ListReply = h
}

// OnButtonReply sets the handler for button reply interactive messages.
func (ih *InteractiveHandler) OnButtonReply(h ButtonReplyMessageHandler) {
	ih.ButtonReply = h
}

// OnFlowCompletion sets the handler for flow completion (nfm_reply) messages.
func (ih *InteractiveHandler) OnFlowCompletion(h NativeFlowCompletionHandler) {
	ih.FlowCompletion = h
}

// OnAddressSubmission sets the handler for address submission messages.
func (ih *InteractiveHandler) OnAddressSubmission(h NativeFlowCompletionHandler) {
	ih.AddressSubmission = h
}

// OnCallPermissionReply sets the handler for call permission reply messages.
func (ih *InteractiveHandler) OnCallPermissionReply(h CallPermissionReplyHandler) {
	ih.CallPermissionReply = h
}

// OnFallback sets the catch-all handler for interactive messages without
// a dedicated subtype handler.
func (ih *InteractiveHandler) OnFallback(h InteractiveMessageHandler) {
	ih.Fallback = h
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
			if err := ih.ListReply.Handle(ctx, newMessageRequest(nctx, info, msg.Interactive.ListReply)); err != nil {
				return fmt.Errorf("handle list reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeButtonReply:
		if ih.ButtonReply != nil {
			if err := ih.ButtonReply.Handle(
				ctx,
				newMessageRequest(nctx, info, msg.Interactive.ButtonReply),
			); err != nil {
				return fmt.Errorf("handle button reply: %w", err)
			}
			return nil
		}

	case InteractiveTypeNFMReply:
		if ih.FlowCompletion != nil {
			if err := ih.FlowCompletion.Handle(
				ctx,
				newMessageRequest(nctx, info, msg.Interactive.NFMReply),
			); err != nil {
				return fmt.Errorf("handle flow completion: %w", err)
			}
			return nil
		}

	case InteractiveAddressSubmission:
		if ih.AddressSubmission != nil {
			if err := ih.AddressSubmission.Handle(
				ctx,
				newMessageRequest(nctx, info, msg.Interactive.NFMReply),
			); err != nil {
				return fmt.Errorf("handle address submission: %w", err)
			}
			return nil
		}

	case InteractiveTypeCallPermissionReply:
		if ih.CallPermissionReply != nil {
			if err := ih.CallPermissionReply.Handle(
				ctx,
				newMessageRequest(nctx, info, msg.Interactive.CallPermissionReply),
			); err != nil {
				return fmt.Errorf("handle call permission reply: %w", err)
			}
			return nil
		}
	}

	if ih.Fallback != nil {
		if err := ih.Fallback.Handle(ctx, newMessageRequest(nctx, info, msg.Interactive)); err != nil {
			return fmt.Errorf("handle interactive message: %w", err)
		}
	}

	return nil
}
