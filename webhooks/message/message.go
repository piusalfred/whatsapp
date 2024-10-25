/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package message

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/message"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	"github.com/piusalfred/whatsapp/webhooks"
)

type (
	NotificationHandler     webhooks.NotificationHandler[Notification]
	NotificationHandlerFunc webhooks.NotificationHandlerFunc[Notification]
)

func (e NotificationHandlerFunc) HandleNotification(ctx context.Context,
	notification *Notification,
) *webhooks.Response {
	return e(ctx, notification)
}

type (
	Notification struct {
		Object string   `json:"object,omitempty"`
		Entry  []*Entry `json:"entry,omitempty"`
	}

	Entry struct {
		ID      string    `json:"id,omitempty"`
		Time    int64     `json:"time,omitempty"`
		Changes []*Change `json:"changes,omitempty"`
	}

	Change struct {
		Value *Value `json:"value,omitempty"`
		Field string `json:"field,omitempty"`
	}

	Value struct {
		MessagingProduct string           `json:"messaging_product,omitempty"`
		Metadata         *Metadata        `json:"metadata,omitempty"`
		Errors           []*werrors.Error `json:"errors,omitempty"`
		Contacts         []*Contact       `json:"contacts,omitempty"`
		Messages         []*Message       `json:"messages,omitempty"`
		Statuses         []*Status        `json:"statuses,omitempty"`
	}

	Contact struct {
		Profile *Profile `json:"profile,omitempty"`
		WaID    string   `json:"wa_id,omitempty"`
	}

	// Message contains the information of a message. It is embedded in the Value object.
	Message struct {
		Audio       *message.MediaInfo `json:"audio,omitempty"`
		Button      *Button            `json:"button,omitempty"`
		Context     *Context           `json:"context,omitempty"`
		Document    *message.MediaInfo `json:"document,omitempty"`
		Errors      []*werrors.Error   `json:"errors,omitempty"`
		From        string             `json:"from,omitempty"`
		ID          string             `json:"id,omitempty"`
		Identity    *Identity          `json:"identity,omitempty"`
		Image       *message.MediaInfo `json:"image,omitempty"`
		Interactive *Interactive       `json:"interactive,omitempty"`
		Order       *Order             `json:"order,omitempty"`
		Referral    *Referral          `json:"referral,omitempty"`
		Sticker     *message.MediaInfo `json:"sticker,omitempty"`
		System      *System            `json:"system,omitempty"`
		Text        *Text              `json:"text,omitempty"`
		Timestamp   string             `json:"timestamp,omitempty"`
		Type        string             `json:"type,omitempty"`
		Video       *message.MediaInfo `json:"video,omitempty"`
		Contacts    *message.Contacts  `json:"contacts,omitempty"`
		Location    *message.Location  `json:"location,omitempty"`
		Reaction    *message.Reaction  `json:"reaction,omitempty"`
	}

	Status struct {
		ID                    string           `json:"id,omitempty"`
		RecipientID           string           `json:"recipient_id,omitempty"`
		StatusValue           string           `json:"status,omitempty"`
		Timestamp             int64            `json:"timestamp,omitempty"`
		Conversation          *Conversation    `json:"conversation,omitempty"`
		Pricing               *Pricing         `json:"pricing,omitempty"`
		Errors                []*werrors.Error `json:"errors,omitempty"`
		BizOpaqueCallbackData string           `json:"biz_opaque_callback_data,omitempty"`
	}

	Metadata struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		PhoneNumberID      string `json:"phone_number_id,omitempty"`
	}

	ReferralNotification struct {
		Text     *Text
		Referral *Referral
	}

	Pricing struct {
		Billable     bool   `json:"billable,omitempty"` // Deprecated
		Category     string `json:"category,omitempty"`
		PricingModel string `json:"pricing_model,omitempty"`
	}

	ConversationOrigin struct {
		Type string `json:"type,omitempty"`
	}

	Conversation struct {
		ID     string              `json:"id,omitempty"`
		Origin *ConversationOrigin `json:"origin,omitempty"`
		Expiry int                 `json:"expiration_timestamp,omitempty"`
	}

	Profile struct {
		Name string `json:"name,omitempty"`
	}

	System struct {
		Body     string `json:"body,omitempty"`
		Identity string `json:"identity,omitempty"`
		NewWaID  string `json:"new_wa_id,omitempty"`
		Type     string `json:"type,omitempty"`
		WaID     string `json:"wa_id,omitempty"`
		Customer string `json:"customer,omitempty"`
	}

	Text struct {
		Body string `json:"body,omitempty"`
	}

	Interactive struct {
		Type        string       `json:"type,omitempty"`
		ButtonReply *ButtonReply `json:"button_reply,omitempty"`
		ListReply   *ListReply   `json:"list_reply,omitempty"`
		NFMReply    *NFMReply    `json:"nfm_reply,omitempty"`
	}

	NFMReply struct {
		Name         string          `json:"name"`          // Always "flow"
		Body         string          `json:"body"`          // Always "Sent"
		ResponseJSON json.RawMessage `json:"response_json"` // Flow-specific data (JSON string)
	}

	InteractiveType struct {
		ButtonReply *ButtonReply `json:"button_reply,omitempty"`
		ListReply   *ListReply   `json:"list_reply,omitempty"`
	}

	ButtonReply struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	ListReply struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	ProductItem struct {
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
		Quantity          string `json:"quantity,omitempty"`
		ItemPrice         string `json:"item_price,omitempty"`
		Currency          string `json:"currency,omitempty"`
	}

	Order struct {
		CatalogID    string         `json:"catalog_id,omitempty"`
		Text         string         `json:"text,omitempty"`
		ProductItems []*ProductItem `json:"product_items,omitempty"`
	}

	Referral struct {
		SourceURL    string `json:"source_url,omitempty"`
		SourceType   string `json:"source_type,omitempty"`
		SourceID     string `json:"source_id,omitempty"`
		Headline     string `json:"headline,omitempty"`
		Body         string `json:"body,omitempty"`
		MediaType    string `json:"media_type,omitempty"`
		ImageURL     string `json:"image_url,omitempty"`
		VideoURL     string `json:"video_url,omitempty"`
		ThumbnailURL string `json:"thumbnail_url,omitempty"`
		CtwaClid     string `json:"ctwa_clid,omitempty"`
	}

	Button struct {
		Payload string `json:"payload,omitempty"`
		Text    string `json:"text,omitempty"`
	}

	Identity struct {
		Acknowledged     bool   `json:"acknowledged,omitempty"`
		CreatedTimestamp int64  `json:"created_timestamp,omitempty"`
		Hash             string `json:"hash,omitempty"`
	}

	Context struct {
		Forwarded           bool   `json:"forwarded,omitempty"`
		FrequentlyForwarded bool   `json:"frequently_forwarded,omitempty"`
		From                string `json:"from,omitempty"`
		ID                  string `json:"id,omitempty"`
		ReferredProduct     *ReferredProduct
		Type                string `json:"type,omitempty"`
	}

	ReferredProduct struct {
		CatalogID         string `json:"catalog_id,omitempty"`
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
	}
)

// PayloadMaxSize is the maximum size of the payload that can be sent to the webhook.
// Webhooks payloads can be up to 3MB.
const PayloadMaxSize = 3 * 1024 * 1024

type (
	// NotificationContext is the context of a notification contains information about the
	// notification and the business that is subscribed to the Webhooks.
	// these are common fields to all notifications.
	// ID - The WhatsApp Business Account ID for the business that is subscribed to the webhook.
	// Contacts - Array of contact objects with information for the customer who sent a message
	// to the business
	// Metadata - A metadata object describing the business subscribed to the webhook.
	NotificationContext struct {
		ID       string
		Contacts []*Contact
		Metadata *Metadata
	}

	// Info is the context of a message contains information about the
	// message and the business that is subscribed to the Webhooks.
	// these are common fields to all type of messages.
	// From The customer's phone number who sent the message to the business.
	// ID The ID for the message that was received by the business. You could use messages
	// endpoint to mark this specific message as read.
	// Timestamp The timestamp for when the message was received by the business.
	// Type The type of message that was received by the business.
	// Context The context of the message. Only included when a user replies or interacts with one
	// of your messages.
	Info struct {
		From      string
		ID        string
		Timestamp string
		Type      string
		Context   *Context
	}
)

var _ webhooks.NotificationHandler[Notification] = (*Handlers)(nil)

type Handlers struct {
	OrderMessage        OrderMessageHandler
	ButtonMessage       ButtonMessageHandler
	LocationMessage     LocationMessageHandler
	ContactsMessage     ContactsMessageHandler
	MessageReaction     ReactionHandler
	UnknownMessage      ErrorsHandler
	ProductEnquiry      ProductEnquiryHandler
	InteractiveMessage  InteractiveMessageHandler
	ButtonReply         ButtonReplyMessageHandler
	ListReply           ListReplyMessageHandler
	FlowReply           FlowCompletionMessageHandler
	MessageErrors       ErrorsHandler
	TextMessage         TextMessageHandler
	ReferralMessage     ReferralMessageHandler
	CustomerIDChange    CustomerIDChangeHandler
	SystemMessage       SystemMessageHandler
	AudioMessage        MediaMessageHandler
	VideoMessage        MediaMessageHandler
	ImageMessage        MediaMessageHandler
	DocumentMessage     MediaMessageHandler
	StickerMessage      MediaMessageHandler
	NotificationError   ErrorHandler
	MessageStatusChange StatusChangeHandler
	MessageReceived     ReceivedHandler
}

// SetOrderMessageHandler sets the order message handler.
func (handler *Handlers) SetOrderMessageHandler(h OrderMessageHandler) {
	handler.OrderMessage = h
}

// SetButtonMessageHandler sets the button message handler.
func (handler *Handlers) SetButtonMessageHandler(h ButtonMessageHandler) {
	handler.ButtonMessage = h
}

// SetLocationMessageHandler sets the location message handler.
func (handler *Handlers) SetLocationMessageHandler(h LocationMessageHandler) {
	handler.LocationMessage = h
}

// SetContactsMessageHandler sets the contacts message handler.
func (handler *Handlers) SetContactsMessageHandler(h ContactsMessageHandler) {
	handler.ContactsMessage = h
}

// SetMessageReactionHandler sets the message reaction handler.
func (handler *Handlers) SetMessageReactionHandler(h ReactionHandler) {
	handler.MessageReaction = h
}

// SetUnknownMessageHandler sets the unknown message handler.
func (handler *Handlers) SetUnknownMessageHandler(h ErrorsHandler) {
	handler.UnknownMessage = h
}

// SetProductEnquiryHandler sets the product enquiry handler.
func (handler *Handlers) SetProductEnquiryHandler(h ProductEnquiryHandler) {
	handler.ProductEnquiry = h
}

// SetInteractiveMessageHandler sets the interactive message handler.
func (handler *Handlers) SetInteractiveMessageHandler(h InteractiveMessageHandler) {
	handler.InteractiveMessage = h
}

// SetButtonReplyMessageHandler sets the button reply message handler.
func (handler *Handlers) SetButtonReplyMessageHandler(h ButtonReplyMessageHandler) {
	handler.ButtonReply = h
}

// SetListReplyMessageHandler sets the list reply message handler.
func (handler *Handlers) SetListReplyMessageHandler(h ListReplyMessageHandler) {
	handler.ListReply = h
}

// SetFlowCompletionMessageHandler sets the flow completion message handler.
func (handler *Handlers) SetFlowCompletionMessageHandler(h FlowCompletionMessageHandler) {
	handler.FlowReply = h
}

// SetMessageErrorsHandler sets the message errors handler.
func (handler *Handlers) SetMessageErrorsHandler(h ErrorsHandler) {
	handler.MessageErrors = h
}

// SetTextMessageHandler sets the text message handler.
func (handler *Handlers) SetTextMessageHandler(h TextMessageHandler) {
	handler.TextMessage = h
}

// SetReferralMessageHandler sets the referral message handler.
func (handler *Handlers) SetReferralMessageHandler(h ReferralMessageHandler) {
	handler.ReferralMessage = h
}

// SetCustomerIDChangeHandler sets the customer ID change handler.
func (handler *Handlers) SetCustomerIDChangeHandler(h CustomerIDChangeHandler) {
	handler.CustomerIDChange = h
}

// SetSystemMessageHandler sets the system message handler.
func (handler *Handlers) SetSystemMessageHandler(h SystemMessageHandler) {
	handler.SystemMessage = h
}

// SetAudioMessageHandler sets the audio message handler.
func (handler *Handlers) SetAudioMessageHandler(h MediaMessageHandler) {
	handler.AudioMessage = h
}

// SetVideoMessageHandler sets the video message handler.
func (handler *Handlers) SetVideoMessageHandler(h MediaMessageHandler) {
	handler.VideoMessage = h
}

// SetImageMessageHandler sets the image message handler.
func (handler *Handlers) SetImageMessageHandler(h MediaMessageHandler) {
	handler.ImageMessage = h
}

// SetDocumentMessageHandler sets the document message handler.
func (handler *Handlers) SetDocumentMessageHandler(h MediaMessageHandler) {
	handler.DocumentMessage = h
}

// SetStickerMessageHandler sets the sticker message handler.
func (handler *Handlers) SetStickerMessageHandler(h MediaMessageHandler) {
	handler.StickerMessage = h
}

// SetNotificationErrorHandler sets the notification error handler.
func (handler *Handlers) SetNotificationErrorHandler(h ErrorHandler) {
	handler.NotificationError = h
}

// SetMessageStatusChangeHandler sets the message status change handler.
func (handler *Handlers) SetMessageStatusChangeHandler(h StatusChangeHandler) {
	handler.MessageStatusChange = h
}

// SetMessageReceivedHandler sets the message received handler.
func (handler *Handlers) SetMessageReceivedHandler(h ReceivedHandler) {
	handler.MessageReceived = h
}

func (handler *Handlers) HandleNotification(ctx context.Context, notification *Notification) *webhooks.Response {
	if err := handler.handleNotification(ctx, notification); err != nil {
		return &webhooks.Response{StatusCode: http.StatusInternalServerError}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}

func (handler *Handlers) handleNotification(ctx context.Context,
	notification *Notification,
) error {
	if notification == nil {
		return nil
	}

	for _, entry := range notification.Entry {
		if err := handler.handleNotificationEntry(ctx, entry); err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handlers) handleNotificationEntry(ctx context.Context, entry *Entry) error {
	entryID := entry.ID
	changes := entry.Changes
	for _, change := range changes {
		value := change.Value
		if value == nil {
			continue
		}
		if err := handler.handleNotificationChangeValue(ctx, entryID, value); err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handlers) handleNotificationChangeValue(ctx context.Context,
	id string, value *Value,
) error {
	notificationCtx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	if handler.NotificationError != nil {
		for _, ev := range value.Errors {
			if err := handler.NotificationError.Handle(ctx, notificationCtx, ev); err != nil {
				return fmt.Errorf("%w: %w", ErrNotificationErrorHandler, err)
			}
		}
	}

	if handler.MessageStatusChange != nil {
		for _, sv := range value.Statuses {
			if err := handler.MessageStatusChange.Handle(ctx, notificationCtx, sv); err != nil {
				return fmt.Errorf("%w: %w", ErrMessageStatusChangeHandler, err)
			}
		}
	}

	for _, mv := range value.Messages {
		if handler.MessageReceived != nil {
			if err := handler.MessageReceived.Handle(ctx, notificationCtx, mv); err != nil {
				return fmt.Errorf("%w: %w", ErrMessageReceivedNotificationHandler, err)
			}
		}

		if err := handler.handleNotificationMessage(ctx, notificationCtx, mv); err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handlers) handleNotificationMessage(ctx context.Context,
	nctx *NotificationContext, message *Message,
) error {
	mctx := &Info{
		From:      message.From,
		ID:        message.ID,
		Timestamp: message.Timestamp,
		Type:      message.Type,
		Context:   message.Context,
	}

	messageType := ParseType(message.Type)
	switch messageType {
	case TypeOrder:
		if err := handler.OrderMessage.Handle(ctx, nctx, mctx, message.Order); err != nil {
			return fmt.Errorf("%w: %w", ErrOrderMessageHandler, err)
		}

		return nil

	case TypeButton:
		if err := handler.ButtonMessage.Handle(ctx, nctx, mctx, message.Button); err != nil {
			return fmt.Errorf("%w: %w", ErrButtonMessageHandler, err)
		}

		return nil

	case TypeAudio, TypeVideo, TypeImage, TypeDocument, TypeSticker:
		return handler.handleMediaMessage(ctx, nctx, message, mctx)

	case TypeInteractive:
		return handler.handleInteractiveNotification(ctx, nctx, message, mctx)

	case TypeSystem:
		if err := handler.SystemMessage.Handle(ctx, nctx, mctx, message.System); err != nil {
			return fmt.Errorf("%w: %w", ErrSystemMessageHandler, err)
		}

		return nil

	case TypeUnknown:
		if err := handler.MessageErrors.Handle(ctx, nctx, mctx, message.Errors); err != nil {
			return fmt.Errorf("%w: %w", ErrUnknownMessageHandler, err)
		}

		return nil

	case TypeText:
		return handler.handleTextNotification(ctx, nctx, message, mctx)

	case TypeReaction:
		if err := handler.MessageReaction.Handle(ctx, nctx, mctx, message.Reaction); err != nil {
			return fmt.Errorf("%w: %w", ErrMessageReaction, err)
		}

		return nil

	case TypeLocation:
		if err := handler.LocationMessage.Handle(ctx, nctx, mctx, message.Location); err != nil {
			return fmt.Errorf("%w: %w", ErrLocationMessage, err)
		}

		return nil

	case TypeContacts:
		if err := handler.ContactsMessage.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
			return fmt.Errorf("%w: %w", ErrContactsMessage, err)
		}

		return nil

	default:
		return handler.handleDefaultNotificationMessage(ctx, nctx, message, mctx)
	}
}

func (handler *Handlers) handleMediaMessage(ctx context.Context, nctx *NotificationContext,
	message *Message, mctx *Info,
) error {
	messageType := ParseType(message.Type)
	switch messageType { //nolint:exhaustive
	case TypeAudio:
		if err := handler.AudioMessage.Handle(ctx, nctx, mctx, message.Audio); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil

	case TypeVideo:
		if err := handler.VideoMessage.Handle(ctx, nctx, mctx, message.Video); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil

	case TypeImage:
		if err := handler.ImageMessage.Handle(ctx, nctx, mctx, message.Image); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil

	case TypeDocument:
		if err := handler.DocumentMessage.Handle(ctx, nctx, mctx, message.Document); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil

	case TypeSticker:
		if err := handler.StickerMessage.Handle(ctx, nctx, mctx, message.Sticker); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil
	}

	return nil
}

func (handler *Handlers) handleTextNotification(ctx context.Context, nctx *NotificationContext,
	message *Message, mctx *Info,
) error {
	if message.Referral != nil {
		reff := &ReferralNotification{
			Text:     message.Text,
			Referral: message.Referral,
		}
		if err := handler.ReferralMessage.Handle(ctx, nctx, mctx, reff); err != nil {
			return fmt.Errorf("%w: %w", ErrReferralMessage, err)
		}

		return nil
	}

	if mctx.Context != nil {
		if err := handler.ProductEnquiry.Handle(ctx, nctx, mctx, message.Text); err != nil {
			return fmt.Errorf("%w: %w", ErrProductEnquiry, err)
		}

		return nil
	}

	if err := handler.TextMessage.Handle(ctx, nctx, mctx, message.Text); err != nil {
		return fmt.Errorf("%w: %w", ErrTextMessageHandler, err)
	}

	return nil
}

func (handler *Handlers) handleDefaultNotificationMessage(ctx context.Context, nctx *NotificationContext,
	message *Message, mctx *Info,
) error {
	if message.Contacts != nil {
		if err := handler.ContactsMessage.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
			return fmt.Errorf("%w: %w", ErrContactsMessageHandler, err)
		}

		return nil
	}
	if message.Location != nil {
		if err := handler.LocationMessage.Handle(ctx, nctx, mctx, message.Location); err != nil {
			return fmt.Errorf("%w: %w", ErrLocationMessage, err)
		}

		return nil
	}

	if message.Identity != nil {
		if err := handler.CustomerIDChange.Handle(ctx, nctx, mctx, message.Identity); err != nil {
			return fmt.Errorf("%w: %w", ErrCustomerIDChange, err)
		}

		return nil
	}

	return fmt.Errorf("%w: unsupported message type", ErrHandleMessage)
}

func (handler *Handlers) handleInteractiveNotification(ctx context.Context,
	nctx *NotificationContext, message *Message, mctx *Info,
) error {
	switch message.Interactive.Type {
	case InteractiveTypeListReply:
		if err := handler.ListReply.Handle(ctx, nctx, mctx, message.Interactive.ListReply); err != nil {
			return fmt.Errorf("handle list reply: %w", err)
		}

		return nil
	case InteractiveTypeButtonReply:
		if err := handler.ButtonReply.Handle(ctx, nctx, mctx, message.Interactive.ButtonReply); err != nil {
			return fmt.Errorf("handle button reply: %w", err)
		}

		return nil
	case InteractiveTypeNFMReply:
		if err := handler.FlowReply.Handle(ctx, nctx, mctx, message.Interactive.NFMReply); err != nil {
			return fmt.Errorf("handle flow reply: %w", err)
		}

		return nil
	default:
		if err := handler.InteractiveMessage.Handle(ctx, nctx, mctx, message.Interactive); err != nil {
			return fmt.Errorf("handle interactive message: %w", err)
		}

		return nil
	}
}

type (
	ButtonMessageHandler         = Handler[Button]
	TextMessageHandler           = Handler[Text]
	OrderMessageHandler          = Handler[Order]
	LocationMessageHandler       = Handler[message.Location]
	ContactsMessageHandler       = Handler[message.Contacts]
	ReactionHandler              = Handler[message.Reaction]
	ProductEnquiryHandler        = Handler[Text]
	InteractiveMessageHandler    = Handler[Interactive]
	ButtonReplyMessageHandler    = Handler[ButtonReply]
	ListReplyMessageHandler      = Handler[ListReply]
	FlowCompletionMessageHandler = Handler[NFMReply]
	ReferralMessageHandler       = Handler[ReferralNotification]
	CustomerIDChangeHandler      = Handler[Identity]
	SystemMessageHandler         = Handler[System]
	MediaMessageHandler          = Handler[message.MediaInfo]
	ErrorHandler                 = ChangeValueHandler[werrors.Error]
	StatusChangeHandler          = ChangeValueHandler[Status]
	ReceivedHandler              = ChangeValueHandler[Message]
	OnButtonMessageHook          = HandlerFunc[Button]
	OnTextMessageHook            = HandlerFunc[Text]
	OnOrderMessageHook           = HandlerFunc[Order]
	OnLocationMessageHook        = HandlerFunc[message.Location]
	OnContactsMessageHook        = HandlerFunc[message.Contacts]
	OnMessageReactionHook        = HandlerFunc[message.Reaction]
	OnProductEnquiryHook         = HandlerFunc[Text]
	OnInteractiveMessageHook     = HandlerFunc[Interactive]
	OnButtonReplyMessageHook     = HandlerFunc[ButtonReply]
	OnListReplyMessageHook       = HandlerFunc[ListReply]
	OnFlowCompletionMessageHook  = HandlerFunc[NFMReply]
	OnReferralMessageHook        = HandlerFunc[ReferralNotification]
	OnCustomerIDChangeHook       = HandlerFunc[Identity]
	OnSystemMessageHook          = HandlerFunc[System]
	OnMediaMessageHook           = HandlerFunc[message.MediaInfo]
	OnNotificationErrorHook      = ChangeValueHandlerFunc[werrors.Error]
	OnMessageStatusChangeHook    = ChangeValueHandlerFunc[Status]
	OnMessageReceivedHook        = ChangeValueHandlerFunc[Message]
)

type (
	HandlerFunc[T any] func(
		ctx context.Context, nctx *NotificationContext, mctx *Info, message *T) error

	Handler[T any] interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *Info, message *T) error
	}
)

func (h HandlerFunc[T]) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *Info,
	message *T,
) error {
	return h(ctx, nctx, mctx, message)
}

type ChangeValueHandler[T any] interface {
	Handle(ctx context.Context, nctx *NotificationContext, value *T) error
}

type ChangeValueHandlerFunc[T any] func(ctx context.Context, nctx *NotificationContext, value *T) error

func (f ChangeValueHandlerFunc[T]) Handle(ctx context.Context, nctx *NotificationContext, value *T) error {
	return f(ctx, nctx, value)
}

type (
	ErrorsHandlerFunc func(
		ctx context.Context, nctx *NotificationContext, mctx *Info, errors []*werrors.Error) error
	ErrorsHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *Info, errors []*werrors.Error) error
	}
)

func (h ErrorsHandlerFunc) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *Info,
	errors []*werrors.Error,
) error {
	return h(ctx, nctx, mctx, errors)
}

// messageError is a custom error type for webhook errors.
type messageError string

func (e messageError) Error() string {
	return string(e)
}

const (
	ErrHandleMessage                      = messageError("could not handle message")
	ErrNotificationErrorHandler           = messageError("notification error handler failed")
	ErrOrderMessageHandler                = messageError("order message handler failed")
	ErrButtonMessageHandler               = messageError("button message handler failed")
	ErrMediaMessageHandler                = messageError("media message handler failed")
	ErrTextMessageHandler                 = messageError("text message handler failed")
	ErrSystemMessageHandler               = messageError("system message handler failed")
	ErrReferralMessage                    = messageError("referral message handler failed")
	ErrMessageReaction                    = messageError("message reaction handler failed")
	ErrLocationMessage                    = messageError("location message handler failed")
	ErrContactsMessage                    = messageError("contacts message handler failed")
	ErrCustomerIDChange                   = messageError("customer id change handler failed")
	ErrProductEnquiry                     = messageError("product enquiry handler failed")
	ErrUnknownMessageHandler              = messageError("unknown message handler failed")
	ErrContactsMessageHandler             = messageError("contacts message handler failed")
	ErrMessageStatusChangeHandler         = messageError("message status change handler failed")
	ErrMessageReceivedNotificationHandler = messageError("message received notification handler failed")
)

const (
	TypeAudio       Type = "audio"
	TypeButton      Type = "button"
	TypeDocument    Type = "document"
	TypeText        Type = "text"
	TypeImage       Type = "image"
	TypeInteractive Type = "interactive"
	TypeOrder       Type = "order"
	TypeSticker     Type = "sticker"
	TypeSystem      Type = "system"
	TypeUnknown     Type = "unknown"
	TypeVideo       Type = "video"
	TypeLocation    Type = "location"
	TypeReaction    Type = "reaction"
	TypeContacts    Type = "contacts"
)

// Type is type of message that has been received by the business that has subscribed
// to Webhooks. Possible value can be one of the following: audio,button,document,text,image,
// interactive,order,sticker,system – for customer number change messages,unknown and video
// The documentation is not clear in case of location,reaction and contacts. They will be included
// just in case.
type Type string

// ParseType parses the message type from a string.
func ParseType(s string) Type {
	msgMap := map[string]Type{
		"audio":       TypeAudio,
		"button":      TypeButton,
		"document":    TypeDocument,
		"text":        TypeText,
		"image":       TypeImage,
		"interactive": TypeInteractive,
		"order":       TypeOrder,
		"sticker":     TypeSticker,
		"system":      TypeSystem,
		"unknown":     TypeUnknown,
		"video":       TypeVideo,
		"location":    TypeLocation,
		"reaction":    TypeReaction,
		"contacts":    TypeContacts,
	}

	msgType, ok := msgMap[strings.TrimSpace(strings.ToLower(s))]
	if !ok {
		return ""
	}

	return msgType
}

const (
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusRead      DeliveryStatus = "read"
	DeliveryStatusSent      DeliveryStatus = "sent"
)

type DeliveryStatus string

const (
	InteractiveTypeListReply   = "list_reply"
	InteractiveTypeButtonReply = "button_reply"
	InteractiveTypeNFMReply    = "nfm_reply"
)
