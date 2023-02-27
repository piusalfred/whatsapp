/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	werrors "github.com/piusalfred/whatsapp/errors"
	"github.com/piusalfred/whatsapp/models"
)

// PayloadMaxSize is the maximum size of the payload that can be sent to the webhook.
// Webhooks payloads can be up to 3MB.
const PayloadMaxSize = 3 * 1024 * 1024

const (
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusSent      MessageStatus = "sent"
)

const (
	AudioMessageType       MessageType = "audio"
	ButtonMessageType      MessageType = "button"
	DocumentMessageType    MessageType = "document"
	TextMessageType        MessageType = "text"
	ImageMessageType       MessageType = "image"
	InteractiveMessageType MessageType = "interactive"
	OrderMessageType       MessageType = "order"
	StickerMessageType     MessageType = "sticker"
	SystemMessageType      MessageType = "system"
	UnknownMessageType     MessageType = "unknown"
	VideoMessageType       MessageType = "video"
	LocationMessageType    MessageType = "location"
	ReactionMessageType    MessageType = "reaction"
	ContactMessageType     MessageType = "contacts"
)

const (
	InteractiveListReply   InteractiveReply = "list_reply"
	InteractiveButtonReply InteractiveReply = "button_reply"
)

type (

	// InteractiveReply is the type of interactive reply. It can be one of the following:
	// list_reply,or button_reply.
	InteractiveReply string

	// MessageType is type of message that has been received by the business that has subscribed
	// to Webhooks. Possible value can be one of the following: audio,button,document,text,image,
	// interactive,order,sticker,system – for customer number change messages,unknown and video
	// The documentation is not clear in case of location,reaction and contacts. They will be included
	// just in case.
	MessageType string

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

	// MessageContext is the context of a message contains information about the
	// message and the business that is subscribed to the Webhooks.
	// these are common fields to all type of messages.
	// From The customer's phone number who sent the message to the business.
	// ID The ID for the message that was received by the business. You could use messages
	// endpoint to mark this specific message as read.
	// Timestamp The timestamp for when the message was received by the business.
	// Type The type of message that was received by the business.
	// Ctx The context of the message. Only included when a user replies or interacts with one
	// of your messages.
	MessageContext struct {
		From      string
		ID        string
		Timestamp string
		Type      string
		Ctx       *Context
	}

	// NotificationHooks is a generic interface for all Hooks. It intends to have a dedicated hook  for each
	// notification type or scenario.
	//
	// All the Hooks takes a context.Context, a NotificationContext which is used to identify and
	// distinguish one notification to the rest. The Hooks that deals with messages like these
	// OnMessageErrors, OnMessageReceived, OnTextMessageReceived, OnReferralMessageReceived takes a
	// OnMessageStatusChange is a hook that is called when a message status changes.
	// Status change is triggered when a message is sent or delivered to a customer or
	// the customer reads the delivered message sent by a business that is subscribed
	// to the Webhooks.
	//
	// OnNotificationError is a hook that is called when a notification error occurs. Sometimes a
	// webhook notification being sent to a business contains errors. This hook is called when a
	// webhook notification contains errors. This hook is called when a webhook notification contains
	// errors.
	//
	// OnMessageReceived is a hook that is called when a message is received. This message can be a
	// text message, image, video, audio, document, location, vcard, template, sticker, or file. It can
	// be a reply to a message sent by the business. This is overridden by the more specific Hooks
	// like OnTextMessageReceived, OnReferralMessageReceived, OnImageReceived, and OnVideoReceived.
	//
	// OnMessageErrors is a hook that is called when the notification contains errors.
	//
	// OnTextMessageReceived is a hook that is called when a text message is received.
	//
	// OnReferralMessageReceived is a hook that is called when a referral message is received.
	// A referral message is a message is sent when a customer clicked an ad that redirects them
	// to WhatsApp.
	// Note that there is no message type for referral. According to documentation, it is included
	// when the type is set to text. So when the message type is set to text, this hook is called.
	// but when a condition that the message contains a referral object is met.
	//
	// OnCustomerIDChange is a hook that is called when a customer ID changes. Webhook is triggered
	// when a customer's phone number or profile information has been updated.
	//
	// OnSystemMessage is a hook that is called when a system message is received.When messages type
	// is set to system, a customer has updated their phone number or profile information, this object
	// is included in the messages object.
	//
	// OnButtonMessage is a hook that is called when a button message is received.
	// When your customer clicks on a quick reply button in an interactive message template,
	// a response is sent. This hook is responsible for handling that response
	//
	// OnLocationReceived is a hook that is called when a location is received. From documentation
	// there is no message type for location but it is included in the messages object when a customer
	// sends a location.
	// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
	//
	// OnContactsReceived is a hook that is called when a contact is received. From documentation
	// there is no message type for contact but it is included in the messages object when a customer
	// sends a contact.
	// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
	//
	// OnMessageReaction is a hook that is called when a message reaction is received. From documentation
	// there is no message type for reaction but it is included in the messages object when a customer
	// reacts to a message.
	// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
	//
	// OnProductEnquiry is a hook that is called when a product enquiry is received. A product enquiry is
	// a message that is sent when a customer clicks on a product in a catalog template.
	//
	// Snippet from documentation:
	// A Product Inquiry Message is received when a customer asks for more information about a product.
	// These can happen when:
	// - a customer replies to Single or Multi-Product Messages, or
	// - a customer accesses a business's catalog via another entry point, navigates to a Product Details page,
	//   and clicks Message Business about this Product.
	//
	// There is no message type for product enquiry. According to documentation, it is included as a text
	// Example:
	//
	//"messages": [
	// {
	// 	"from": "PHONE_NUMBER",
	// 	"id": "wamid.ID",
	// 	"text": {
	// 	  "body": "MESSAGE_TEXT"
	// 	},
	// 	"context": {
	// 	  "from": "PHONE_NUMBER",
	// 	  "id": "wamid.ID",
	// 	  "referred_product": {
	// 		"catalog_id": "CATALOG_ID",
	// 		"product_retailer_id": "PRODUCT_ID"
	// 	  }
	// 	},
	// 	"timestamp": "TIMESTAMP",
	// 	"type": "text"
	//   }
	// ]
	// Referred product is the product being enquired.
	//
	// OnInteractiveMessage is a hook that is called when an interactive message is received.
	// This can happen when a customer clicks on a button you sent them in a template message.
	// Or they can click a list item in a list template you sent them. In case of a list template
	// the reply will be of type list_reply and button_reply for a button template.
	//NotificationHooks interface {
	//	OnMessageStatusChange(ctx context.Context, nctx *NotificationContext, status *Status) error
	//	OnNotificationError(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error
	//	OnMessageReceived(ctx context.Context, nctx *NotificationContext, message *Message) error
	//}
	//
	//MessageHooks interface {
	//	OnMessageErrors(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	//	OnTextMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	//	OnReferralMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text, referral *Referral) error
	//	OnCustomerIdChangeMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, customerID *Identity) error
	//	OnSystemMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, system *System) error
	//	OnImageMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, image *models.MediaInfo) error
	//	OnAudioMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, audio *models.MediaInfo) error
	//	OnVideoMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, video *models.MediaInfo) error
	//	OnDocumentMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, document *models.MediaInfo) error
	//	OnStickerMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, sticker *models.MediaInfo) error
	//	OnOrderMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, order *Order) error
	//	OnButtonMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, button *Button) error
	//	OnLocationMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, location *models.Location) error
	//	OnContactsMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, contacts *models.Contacts) error
	//	OnMessageReaction(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, reaction *models.Reaction) error
	//	OnUnknownMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	//	OnProductEnquiry(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	//	OnInteractiveMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, interactive *Interactive) error
	//}

	OnOrderMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, order *Order) error
	OnButtonMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, button *Button) error
	OnLocationMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, location *models.Location) error
	OnContactsMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, contacts *models.Contacts) error
	OnMessageReactionHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, reaction *models.Reaction) error
	OnUnknownMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	OnProductEnquiryHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	OnInteractiveMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, interactive *Interactive) error

	OnMessageErrorsHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	OnTextMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	OnReferralMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text, referral *Referral) error
	OnCustomerIDChangeMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, customerID *Identity) error
	OnSystemMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, system *System) error

	// OnMediaMessageHook is a hook that is called when a media message is received. This is when Message.Type is
	// image, audio, video or document or sticker.
	OnMediaMessageHook func(ctx context.Context, nctx *NotificationContext, mctx *MessageContext,
		media *models.MediaInfo) error

	// OnNotificationErrorHook is a hook that is called when an error is received in a notification.
	// This is called when an error is received in a notification. This is not called when an error
	// is received in a message, that is handled by NotificationHooks.OnMessageErrors.
	OnNotificationErrorHook func(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

	// OnMessageStatusChangeHook is a hook that is called when a there is a notification about a message status change.
	// This is called when a message status changes. For example, when a message is delivered or read.
	OnMessageStatusChangeHook func(ctx context.Context, nctx *NotificationContext, status *Status) error

	// OnMessageReceivedHook is a hook that is called when a message is received. A notification
	// can contain a lot of things like errors status changes etc. This is called when a
	// notification contains a message. This work with the
	// Message in general. The Hooks for specific message types are called after this hook. They are all implemented
	// in the MessageHooks interface.
	OnMessageReceivedHook func(ctx context.Context, nctx *NotificationContext, message *Message) error

	// Hooks is a struct that contains all the hooks that can be attached to a notification.
	// OnNotificationErrorHook is the OnNotificationErrorHook called when an error is received
	// in a notification.
	//
	// OnMessageStatusChangeHook is the OnMessageStatusChangeHook called when a there is a
	// notification about a message status change.
	// M is the OnMessageReceivedHook called when a message is received.
	// H is the MessageHooks called when a message is received.
	Hooks struct {
		OnOrderMessageHook        OnOrderMessageHook
		OnButtonMessageHook       OnButtonMessageHook
		OnLocationMessageHook     OnLocationMessageHook
		OnContactsMessageHook     OnContactsMessageHook
		OnMessageReactionHook     OnMessageReactionHook
		OnUnknownMessageHook      OnUnknownMessageHook
		OnProductEnquiryHook      OnProductEnquiryHook
		OnInteractiveMessageHook  OnInteractiveMessageHook
		OnMessageErrorsHook       OnMessageErrorsHook
		OnTextMessageHook         OnTextMessageHook
		OnReferralMessageHook     OnReferralMessageHook
		OnCustomerIDChangeHook    OnCustomerIDChangeMessageHook
		OnSystemMessageHook       OnSystemMessageHook
		OnMediaMessageHook        OnMediaMessageHook
		OnNotificationErrorHook   OnNotificationErrorHook
		OnMessageStatusChangeHook OnMessageStatusChangeHook
		OnMessageReceivedHook     OnMessageReceivedHook
	}

	// MessageStatus is the status of a message.
	// delivered – A webhook is triggered when a message sent by a business has been delivered
	// read – A webhook is triggered when a message sent by a business has been read
	// sent – A webhook is triggered when a business sends a message to a customer.
	MessageStatus string
)

// ParseMessageType parses the message type from a string.
func ParseMessageType(s string) MessageType {
	msgMap := map[string]MessageType{
		"audio":       AudioMessageType,
		"button":      ButtonMessageType,
		"document":    DocumentMessageType,
		"text":        TextMessageType,
		"image":       ImageMessageType,
		"interactive": InteractiveMessageType,
		"order":       OrderMessageType,
		"sticker":     StickerMessageType,
		"system":      SystemMessageType,
		"unknown":     UnknownMessageType,
		"video":       VideoMessageType,
		"location":    LocationMessageType,
		"reaction":    ReactionMessageType,
		"contacts":    ContactMessageType,
	}

	msgType, ok := msgMap[strings.TrimSpace(strings.ToLower(s))]
	if !ok {
		return ""
	}

	return msgType
}

const SignatureHeaderKey = "X-Hub-Signature-256"

type (
	Response struct {
		StatusCode int
		Headers    map[string]string
		Body       []byte
		Skip       bool
	}

	HooksErrorHandler func(err error) error
	// NotificationErrorHandler is a function that handles errors that occur when processing a notification.
	// The function returns a Response that is sent to the whatsapp server.
	//
	// Note that retuning nil will make the default use http.StatusOK as the status code.
	//
	// Returning a status code that is not 200, will make a whatsapp server retry the notification. In some
	// cases this can lead to duplicate notifications. If your business logic is affected by this, you should
	// be careful when returning a non 200 status code.
	//
	// This is a snippet from the whatsapp documentation:
	//
	//		If we send a webhook request to your endpoint and your server responds with an HTTP status code other
	//		than 200, or if we are unable to deliver the webhook for another reason, we will keep trying with
	//		decreasing frequency until the request succeeds, for up to 7 days.
	//
	//      Note that retries will be sent to all apps that have subscribed to webhooks (and their appropriate fields)
	//      for the WhatsApp Business Account. This can result in duplicate webhook notifications.
	//
	// NotificationErrorHandler is expected at least to receive errors from NotificationHandler these errors are
	//
	// -  ErrOnBeforeFuncHook when an error is received in the BeforeFunc hook
	// -  ErrOnAttachNotificationHooks when an error is received in the AttachNotificationHooks hook
	// -  ErrOnGenericHandlerFunc when an error is received in the GenericHandlerFunc hook
	NotificationErrorHandler func(context.Context, *http.Request, error) *Response
	BeforeFunc               func(ctx context.Context, notification *Notification) error
	AfterFunc                func(ctx context.Context, notification *Notification, err error)
	HandlerOptions           struct {
		BeforeFunc        BeforeFunc
		AfterFunc         AfterFunc
		ValidateSignature bool
		Secret            string
	}

	VerificationRequest struct {
		Mode      string `json:"hub.mode"`
		Challenge string `json:"hub.challenge"`
		Token     string `json:"hub.verify_token"`
	}

	// SubscriptionVerifier is a function that processes the verification request.
	// The function must return nil if the verification request is valid.
	// It mainly checks if hub.mode is set to subscribe and if the hub.verify_token matches
	// the one set in the App Dashboard.
	SubscriptionVerifier func(context.Context, *VerificationRequest) error

	GenericNotificationHandler func(context.Context, http.ResponseWriter, *Notification) error
)

// SetOnNotificationErrorHook sets the OnNotificationErrorHook.
func (h *Hooks) SetOnNotificationErrorHook(f OnNotificationErrorHook) {
	h.OnNotificationErrorHook = f
}

// NoOpHooksErrorHandler is a no-op hooks error handler. It just returns the error as is.
// It is applied by AttachHooksToNotification if no Hooks error handler is provided.
func NoOpHooksErrorHandler(err error) error {
	return err
}

// NoOpNotificationErrorHandler is a no-op notification error handler. It ignores the error and
// returns a response with status code 200.
func NoOpNotificationErrorHandler(_ context.Context, _ *http.Request, err error) *Response {
	return &Response{
		StatusCode: http.StatusOK,
		Skip:       false,
	}
}

// AttachHooksToNotification applies the hooks to notification received. Sometimes the Hooks can return
// errors. The errors are collected and returned as a single error. So in your implementation
// of NotificationHooks, you can return a FatalError if you want to stop the processing of the notification.
// immediately. If you want to continue processing the notification, you can return a non-fatal
// error. The errors are collected and returned as a single error.
// Also since all hooks errors are passed to the HooksErrorHandler, you can decide to either
// escalate the non-fatal errors to fatal errors or just ignore them also you can decide to
// ignore the fatal errors.
//
// Example:
//
//	func ShouldIgnoreFatalErrors(ignore bool) hef{
//	    return func(err error) error {
//	        if IsFatalError(err) {
//	            if ignore {
//	                return fmt.Errorf("ignoring fatal error: %v", err)
//	            }
//	            return err
//	        }
//	        return err
//	    }
//	}
func AttachHooksToNotification(ctx context.Context, notification *Notification,
	hooks *Hooks, heh HooksErrorHandler,
) error {
	if notification == nil || hooks == nil {
		return nil
	}

	entries := notification.Entry
	for _, entry := range entries {
		entry := entry
		if err := attachHooksToEntry(ctx, entry, hooks, heh); err != nil {
			return err
		}
	}

	return nil
}

func attachHooksToEntry(ctx context.Context, entry *Entry, hooks *Hooks, heh HooksErrorHandler) error {
	eid := entry.ID
	changes := entry.Changes
	for _, change := range changes {
		change := change
		value := change.Value
		if value == nil {
			continue
		}

		if err := attachHooksToValue(ctx, eid, value, hooks, heh); err != nil {
			return err
		}
	}

	return nil
}

var (
	ErrOnMessageStatusChangeHook = errors.New("on message status change hook error")
	ErrOnMessageHooks            = errors.New("on specific message hooks error")
	ErrOnNotificationErrorHook   = errors.New("on notification error hook error")
	ErrOnGlobalMessageHook       = errors.New("on global message hook error")
)

func attachHooksToValue(ctx context.Context, id string, value *Value, hooks *Hooks,
	hooksErrorHandler HooksErrorHandler,
) error {
	if hooks == nil {
		return nil
	}

	notificationCtx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	// nonFatalErrors is a slice of non-fatal errors that are collected from the hooks.
	// can contain a maximum of 4 errors.
	nonFatalErrors := make([]error, 0, 4)

	// call the Hooks
	if value.Errors != nil && hooks.OnNotificationErrorHook != nil {
		for _, ev := range value.Errors {
			ev := ev
			if err := hooks.OnNotificationErrorHook(ctx, notificationCtx, ev); err != nil {
				if IsFatalError(hooksErrorHandler(err)) {
					return err
				}
				nonFatalErrors = append(nonFatalErrors, ErrOnNotificationErrorHook)
			}
		}
	}

	if value.Statuses != nil && hooks.OnMessageStatusChangeHook != nil {
		for _, sv := range value.Statuses {
			sv := sv
			if err := hooks.OnMessageStatusChangeHook(ctx, notificationCtx, sv); err != nil {
				if IsFatalError(hooksErrorHandler(err)) {
					return err
				}
				nonFatalErrors = append(nonFatalErrors, ErrOnMessageStatusChangeHook)
			}
		}
	}

	if value.Messages != nil {
		for _, mv := range value.Messages {
			mv := mv
			if hooks.OnMessageReceivedHook != nil {
				if err := hooks.OnMessageReceivedHook(ctx, notificationCtx, mv); err != nil {
					if IsFatalError(hooksErrorHandler(err)) {
						return err
					}
					nonFatalErrors = append(nonFatalErrors, ErrOnGlobalMessageHook)
				}
			}

			if err := attachHooksToMessage(ctx, notificationCtx, hooks, mv); err != nil {
				if IsFatalError(hooksErrorHandler(err)) {
					return err
				}
				nonFatalErrors = append(nonFatalErrors, ErrOnMessageHooks)
			}
		}
	}

	return getEncounteredError(nonFatalErrors)
}

func getEncounteredError(errs []error) error {
	var finalErr error
	for i := 0; i < len(errs); i++ {
		if finalErr == nil {
			finalErr = errs[i]

			continue
		}
		finalErr = fmt.Errorf("%w, %w", finalErr, errs[i])
	}

	return finalErr
}

func attachHooksToMessage(ctx context.Context, nctx *NotificationContext, hooks *Hooks, message *Message) error {
	if hooks == nil || message == nil {
		return fmt.Errorf("hooks or message is nil")
	}
	mctx := &MessageContext{
		From:      message.From,
		ID:        message.ID,
		Timestamp: message.Timestamp,
		Type:      message.Type,
		Ctx:       message.Context,
	}
	messageType := ParseMessageType(message.Type)
	switch messageType {
	case OrderMessageType:
		return hooks.OnOrderMessageHook(ctx, nctx, mctx, message.Order)

	case ButtonMessageType:
		return hooks.OnButtonMessageHook(ctx, nctx, mctx, message.Button)

	case AudioMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Audio)

	case VideoMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Video)

	case ImageMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Image)

	case DocumentMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Document)

	case StickerMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Sticker)

	case InteractiveMessageType:
		return hooks.OnInteractiveMessageHook(ctx, nctx, mctx, message.Interactive)

	case SystemMessageType:
		// TODO: documentation is not clear if the ID change will also be sent here:
		return hooks.OnSystemMessageHook(ctx, nctx, mctx, message.System)

	case UnknownMessageType:
		return hooks.OnMessageErrorsHook(ctx, nctx, mctx, message.Errors)

	case TextMessageType:
		if message.Referral != nil {
			return hooks.OnReferralMessageHook(ctx, nctx, mctx, message.Text, message.Referral)
		}
		if mctx.Ctx != nil {
			return hooks.OnProductEnquiryHook(ctx, nctx, mctx, message.Text)
		}

		return hooks.OnTextMessageHook(ctx, nctx, mctx, message.Text)

	case ReactionMessageType:
		return hooks.OnMessageReactionHook(ctx, nctx, mctx, message.Reaction)

	case LocationMessageType:
		return hooks.OnLocationMessageHook(ctx, nctx, mctx, message.Location)

	case ContactMessageType:
		return hooks.OnContactsMessageHook(ctx, nctx, mctx, message.Contacts)

	default:
		if message.Contacts != nil {
			if len(message.Contacts.Contacts) > 0 {
				return hooks.OnContactsMessageHook(ctx, nctx, mctx, message.Contacts)
			}
		}
		if message.Location != nil {
			return hooks.OnLocationMessageHook(ctx, nctx, mctx, message.Location)
		}

		if message.Identity != nil {
			return hooks.OnCustomerIDChangeHook(ctx, nctx, mctx, message.Identity)
		}

		return fmt.Errorf("could not attach hook to this message")
	}
}

var (
	ErrOnBeforeFuncHook          = errors.New("error on before func hook")
	ErrOnAttachNotificationHooks = errors.New("error during attaching hooks to a notification")
	ErrOnGenericHandlerFunc      = errors.New("error on generic handler func")
)

// NotificationHandler takes Hooks, NotificationErrorHandler,HooksErrorHandler and HandlerOptions
// and returns http.Handler
//
// It firstly decodes the request body into Notification struct and then calls BeforeFunc if it is
// not nil if HandlerOptions.ValidateSignature is true, it will validate the signature.
//
// All the errors returned from reading the body, running the BeforeFunc and validating the signature
// are passed to the NotificationErrorHandler, if neh returns true, the request is aborted and the
// response status code is set to http.StatusInternalServerError.
func NotificationHandler(
	hooks *Hooks, neh NotificationErrorHandler, heh HooksErrorHandler, options *HandlerOptions,
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var (
			buff         bytes.Buffer
			err          error
			notification = &Notification{}
		)
		ctx := request.Context()

		defer func() {
			buff.Reset()
			if options != nil {
				if options.AfterFunc != nil {
					options.AfterFunc(ctx, notification, err)
				}
			}
		}()

		if _, err = io.Copy(&buff, request.Body); err != nil && !errors.Is(err, io.EOF) {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}
		request.Body = io.NopCloser(&buff)

		if err = json.NewDecoder(&buff).Decode(notification); err != nil && !errors.Is(err, io.EOF) {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		if options != nil {
			// check if before func is set and call it
			if options.BeforeFunc != nil {
				if bfe := options.BeforeFunc(ctx, notification); bfe != nil {
					err = fmt.Errorf("%w: %w", ErrOnBeforeFuncHook, bfe)
					if handleError(ctx, writer, request, neh, err) {
						return
					}
				}
			}

			if options.ValidateSignature {
				signature := request.Header.Get(SignatureHeaderKey)
				if !ValidateSignature(buff.Bytes(), signature, options.Secret) {
					if handleError(ctx, writer, request, neh, ErrInvalidSignature) {
						return
					}
				}
			}
		}
		// Apply the Hooks
		if err = AttachHooksToNotification(ctx, notification, hooks, heh); err != nil {
			err = fmt.Errorf("%w: %w", ErrOnAttachNotificationHooks, err)
			if handleError(ctx, writer, request, neh, err) {
				return
			}
		}

		writer.WriteHeader(http.StatusOK)
	})
}

func handleError(ctx context.Context, writer http.ResponseWriter, request *http.Request,
	neh NotificationErrorHandler, err error,
) bool {
	res := neh(ctx, request, err)
	if !res.Skip {
		code, headers, message := res.StatusCode, res.Headers, res.Body
		for k, v := range headers {
			writer.Header().Set(k, v)
		}
		writer.WriteHeader(code)
		_, _ = writer.Write(message)

		return true
	}

	return false
}

// VerifySubscriptionHandler verifies the subscription to the webhooks.
// Your endpoint must be able to process two types of HTTPS requests: Verification Requests and Event Notifications.
// Since both requests use HTTPs, your server must have a valid TLS or SSL certificate correctly configured and
// installed. Self-signed certificates are not supported.
//
// Anytime you configure the Webhooks product in your App Dashboard, we'll send a GET request to your endpoint URL.
// Verification requests include the following query string parameters, appended to the end of your endpoint URL.
//
// They will look something like this:
//
//			GET https://www.your-clever-domain-name.com/webhooks?
//					hub.mode=subscribe&
//					hub.challenge=1158201444&
//					hub.verify_token=meatyhamhock
//
//	     - hub.mode This value will always be set to subscribe.
//	     - hub.challenge An int you must pass back to us.
//	     - hub.verify_token A string that we grab from the Verify Token field in your app's App Dashboard.
//	       You will set this string when you complete the Webhooks configuration settings steps.
//
// Whenever your endpoint receives a verification request, it must:
//
//   - Verify that the hub.verify_token value matches the string you set in the Verify Token field when you configure
//     the Webhooks product in your App Dashboard (you haven't set up this token string yet).
//
//   - Respond with the hub.challenge value. If you are in your App Dashboard and configuring your Webhooks product
//     (and thus, triggering a Verification Request), the dashboard will indicate if your endpoint validated the request
//     correctly. If you are using the Graph APIs /app/subscriptions endpoint to configure the Webhooks product, the API
//     will indicate success or failure with a response.
func VerifySubscriptionHandler(verifier SubscriptionVerifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the query parameters from the request.
		q := r.URL.Query()
		mode := q.Get("hub.mode")
		challenge := q.Get("hub.challenge")
		token := q.Get("hub.verify_token")
		if err := verifier(r.Context(), &VerificationRequest{
			Mode:      mode,
			Challenge: challenge,
			Token:     token,
		}); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(challenge))
	})
}

// ValidateSignature validates the signature of the payload. All Event Notification payloads are signed
// with a SHA256 signature and include the signature in the request's X-Hub-Signature-256 header, preceded
// with sha256=. You don't have to validate the payload, but you should.
//
// To validate the payload:
//  1. Generate a SHA256 signature using the payload and your app's App Secret.
//  2. Compare your signature to the signature in the X-Hub-Signature-256 header (everything after sha256=).
//
// If the signatures match, the payload is genuine. Please note that we generate the signature using an escaped
// unicode version of the payload, with lowercase hex digits. If you just calculate against the decoded bytes,
// you will end up with a different signature.
// For example, the string äöå should be escaped to \u00e4\u00f6\u00e5.
func ValidateSignature(payload []byte, signature, secret string) bool {
	// Extract the actual signature from the header
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualSignature, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}

	// Calculate the expected signature using the payload and secret
	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write(payload)
	if err != nil {
		return false
	}
	expectedSignature := mac.Sum(nil)

	// Compare the expected and actual signatures
	return hmac.Equal(actualSignature, expectedSignature)
}
