package webhooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	werrors "github.com/piusalfred/whatsapp/errors"
	"github.com/piusalfred/whatsapp/models"
)

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

type (

	// MessageType is atype of message that has been received by the business that has subscribed
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

	// MessageHooks is a generic interface for all message hooks.
	MessageHooks interface {
		// OnMessageError is a hook that is called when a message error occurs.
		// Sometimes a message being sent to a customer contains errors.
		// This hook is called when a message contains errors.
		OnMessageErrors(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, errors []*werrors.Error) error

		// OnTextMessageReceived is a hook that is called when a text message is received.
		OnTextMessageReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, text *Text) error

		// OnReferralMessageReceived is a hook that is called when a referral message is received.
		// A referral message is a message is sent when a customer clicked an ad that redirects them
		// to WhatsApp.
		// Note that there is no message type for referral. According to documentation, it is included
		// when the type is set to text. So when the message type is set to text, this hook is called.
		// but when a condition that the message contains a referral object is met.
		OnReferralMessageReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, text *Text, referral *Referral) error

		// OnCustomerIDChange is a hook that is called when a customer ID changes. Webhook is triggered
		// when a customer's phone number or profile information has been updated.
		OnCustomerIDChange(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, customerID *Identity) error

		// OnSystemMessage is a hook that is called when a system message is received.When messages type
		// is set to system, a customer has updated their phone number or profile information, this object
		// is included in the messages object.
		OnSystemMessage(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, system *System) error

		// OnImageReceived is a hook that is called when an image is received.
		OnImageReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, image *models.MediaInfo) error

		// OnAudioReceived is a hook that is called when an audio is received.
		OnAudioReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, audio *models.MediaInfo) error

		// OnVideoReceived is a hook that is called when a video is received.
		OnVideoReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, video *models.MediaInfo) error

		// OnDocumentReceived is a hook that is called when a document is received.
		OnDocumentReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, document *models.MediaInfo) error

		// OnStickerReceived is a hook that is called when a sticker is received.
		OnStickerReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, sticker *models.MediaInfo) error

		// OnOrderReceived is a hook that is called when an order is received. Included in the messages object when
		//a customer has placed an order.
		OnOrderReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, order *Order) error

		// OnInteractiveMessage is a hook that is called when an interactive message is received.

		// OnButtonMessage is a hook that is called when a button message is received.
		// When your customer clicks on a quick reply button in an interactive message template,
		// a response is sent. This hook is responsible for handling that response.
		OnButtonMessage(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, button *Button) error

		// OnLocationReceived is a hook that is called when a location is received. From documentation
		// there is no message type for location but it is included in the messages object when a customer
		// sends a location.
		// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
		OnLocationReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, location *models.Location) error

		// OnContactsReceived is a hook that is called when a contact is received. From documentation
		// there is no message type for contact but it is included in the messages object when a customer
		// sends a contact.
		// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
		OnContactsReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, contacts models.Contacts) error

		// OnMessageReaction is a hook that is called when a message reaction is received. From documentation
		// there is no message type for reaction but it is included in the messages object when a customer
		// reacts to a message.
		// Example of that payload can be found here https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
		OnMessageReaction(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, reaction *models.Reaction) error

		// OnUnknownMessageReceived is a hook that is called when an unknown message is received. A message type
		// that is not supported. It includes errors.
		OnUnknownMessageReceived(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, errors []*werrors.Error) error

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
		// Reffered product is the product being enquired.
		OnProductEnquiry(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error

		// OnInteractiveMessage is a hook that is called when an interactive message is received.
		// This can happen when a customer clicks on a button you sent them in a template message.
		// Or they can click a list item in a list template you sent them. In case of of a list template
		// the reply will be of type list_reply and button_reply for a button template.
		OnInteractiveMessage(ctx context.Context, nctx *NotificationContext,
			mctx *MessageContext, interactive *Interactive) error
	}

	// Hooks is a generic interface for all hooks.
	Hooks interface {
		// OnMessageStatusChange is a hook that is called when a message status changes.
		// Status change is triggered when a message is sent or delivered to a customer or
		// the customer reads the delivered message sent by a business that is subscribed
		// to the Webhooks.
		OnMessageStatusChange(ctx context.Context, nctx *NotificationContext, status *Status) error

		// OnNotificationError is a hook that is called when a notification error occurs.
		// Sometimes a webhook notification being sent to a business contains errors.
		// This hook is called when a webhook notification contains errors.
		OnNotificationError(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

		// OnMessageReceived is a hook that is called when a message is received.
		// This message can be a text message, image, video, audio, document, location,
		// vcard, template, sticker, or file. It can be a reply to a message sent by the
		// business or a new message.
		OnMessageReceived(ctx context.Context, nctx *NotificationContext, message *Message, hooks MessageHooks) error
	}

	// MessageStatus is the status of a message.
	// delivered – A webhook is triggered when a message sent by a business has been delivered
	// read – A webhook is triggered when a message sent by a business has been read
	// sent – A webhook is triggered when a business sends a message to a customer
	MessageStatus string

	// OnMessageStatusChange is a hook that is called when a message status changes.
	// Status change is triggered when a message is sent or delivered to a customer or
	// the customer reads the delivered message sent by a business that is subscribed
	// to the Webhooks.
	OnMessageStatusChange func(ctx context.Context, nctx *NotificationContext, status *Status) error

	// OnNotificationError is a hook that is called when a notification error occurs.
	// Sometimes a webhook notification being sent to a business contains errors.
	// This hook is called when a webhook notification contains errors.
	// waba is the Whatsapp Business Account ID
	OnNotificationError func(
		ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

	// OnMessageReceived is a hook that is called when a message is received.
	// This message can be a text message, image, video, audio, document, location,
	// vcard, template, sticker, or file. It can be a reply to a message sent by the
	// business or a new message.
	OnMessageReceived func(ctx context.Context, nctx *NotificationContext,
		messages *Message, hooks MessageHooks) error
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

type ErrorHandler func(err error) error

// ApplyHooks applies the hooks to notification received. Sometimes the hooks can return
// errors. The errors are collected and returned as a single error. So in your implementation
// of Hooks, you can return a FatalError if you want to stop the processing of the notification.
// immediately. If you want to continue processing the notification, you can return a non-fatal
// error. The errors are collected and returned as a single error.
// Also since all hooks errors are passed to the ErrorHandler, you can decide to either
// escalate the non-fatal errors to fatal errors or just ignore them also you can decide to
// ignore the fatal errors.
//
// Example:
//
//	func ShouldIgnoreFatalErrors(ignore bool) ErrorHandler{
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
func ApplyHooks(ctx context.Context, notification *Notification, hooks Hooks,
	mh MessageHooks, eh ErrorHandler) error {
	if notification == nil || hooks == nil {
		return nil
	}

	entries := notification.Entry
	for _, entry := range entries {
		entry := entry
		changes := entry.Changes
		for _, change := range changes {
			change := change
			value := change.Value
			if value == nil {
				continue
			}
			id := entry.ID

			return applyHooks(ctx, id, value, hooks, mh, eh)
		}
	}

	return nil
}

type FatalError struct {
	Err  error
	Desc string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Desc, e.Err.Error())
}

func IsFatalError(err error) bool {
	_, ok := err.(*FatalError)
	return ok
}

func applyHooks(ctx context.Context, id string, value *Value, hooks Hooks, mh MessageHooks, ef ErrorHandler) error {
	if hooks == nil {
		return nil
	}

	var allErrors []error

	nctx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	// call the hooks
	if value.Errors != nil {
		for _, ev := range value.Errors {
			ev := ev
			if err := hooks.OnNotificationError(ctx, nctx, ev); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Statuses != nil {
		for _, sv := range value.Statuses {
			sv := sv
			if err := hooks.OnMessageStatusChange(ctx, nctx, sv); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Messages != nil {
		for _, mv := range value.Messages {
			mv := mv
			if err := hooks.OnMessageReceived(ctx, nctx, mv, mh); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}
