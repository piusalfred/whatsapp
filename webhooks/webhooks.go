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
	"github.com/piusalfred/whatsapp/models"
	"io"
	"net/http"
	"strings"

	werrors "github.com/piusalfred/whatsapp/errors"
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

	// Hooks is a generic interface for all hooks. It intends to have a dedicated hook  for each
	// notification type or scenario.
	//
	// All the hooks takes a context.Context, a NotificationContext which is used to identify and
	// distinguish one notification to the rest. The hooks that deals with messages like these
	// OnMessageErrors, OnMessageReceived, OnTextMessageReceived, OnReferralMessageReceived takes a
	// a MessageContext which is used to identify and distinguish one message to the rest.
	//
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
	// be a reply to a message sent by the business. This is overridden by the more specific hooks
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
	Hooks interface {
		OnMessageStatusChange(ctx context.Context, nctx *NotificationContext, status *Status) error
		OnNotificationError(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error
		OnMessageReceived(ctx context.Context, nctx *NotificationContext, message *Message) error
		OnMessageErrors(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
		OnTextMessageReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
		OnReferralMessageReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text, referral *Referral) error
		OnCustomerIDChange(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, customerID *Identity) error
		OnSystemMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, system *System) error
		OnImageReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, image *models.MediaInfo) error
		OnAudioReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, audio *models.MediaInfo) error
		OnVideoReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, video *models.MediaInfo) error
		OnDocumentReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, document *models.MediaInfo) error
		OnStickerReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, sticker *models.MediaInfo) error
		OnOrderReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, order *Order) error
		OnButtonMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, button *Button) error
		OnLocationReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, location *models.Location) error
		OnContactsReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, contacts *models.Contacts) error
		OnMessageReaction(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, reaction *models.Reaction) error
		OnUnknownMessageReceived(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
		OnProductEnquiry(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
		OnInteractiveMessage(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, interactive *Interactive) error
	}

	// MessageStatus is the status of a message.
	// delivered – A webhook is triggered when a message sent by a business has been delivered
	// read – A webhook is triggered when a message sent by a business has been read
	// sent – A webhook is triggered when a business sends a message to a customer
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

type (
	HooksErrorHandler        func(err error) error
	NotificationErrorHandler func(context.Context, http.ResponseWriter, *http.Request, error) error
)

// NoOpHooksErrorHandler is a no-op hooks error handler. It just returns the error as is.
// It is applied by ApplyHooks if no hooks error handler is provided.
func NoOpHooksErrorHandler(err error) error {
	return err
}

// NoOpNotificationErrorHandler is a no-op notification error handler. It just returns the error as is.
// It is applied by ApplyHooks if no notification error handler is provided.
func NoOpNotificationErrorHandler(_ context.Context, _ http.ResponseWriter, _ *http.Request, err error) error {
	return err
}

// ApplyHooks applies the hooks to notification received. Sometimes the hooks can return
// errors. The errors are collected and returned as a single error. So in your implementation
// of Hooks, you can return a FatalError if you want to stop the processing of the notification.
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
func ApplyHooks(ctx context.Context, notification *Notification, hooks Hooks, eh HooksErrorHandler) error {
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

			return applyHooks(ctx, id, value, hooks, eh)
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
	var fatalErr *FatalError
	return errors.As(err, &fatalErr)
}

func applyHooks(ctx context.Context, id string, value *Value, hooks Hooks, ef HooksErrorHandler) error {
	if hooks == nil {
		return nil
	}

	var allErrors []error

	notificationCtx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	if ef == nil {
		ef = NoOpHooksErrorHandler
	}

	// call the hooks
	if value.Errors != nil {
		for _, ev := range value.Errors {
			ev := ev
			if err := hooks.OnNotificationError(ctx, notificationCtx, ev); err != nil {
				if IsFatalError(ef(err)) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Statuses != nil {
		for _, sv := range value.Statuses {
			sv := sv
			if err := hooks.OnMessageStatusChange(ctx, notificationCtx, sv); err != nil {
				if IsFatalError(ef(err)) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Messages != nil {
		for _, mv := range value.Messages {
			mv := mv
			if err := attachHooksToMessage(ctx, notificationCtx, hooks, mv); err != nil {
				if IsFatalError(ef(err)) {
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

var (
	ErrNilNotificationHook = errors.New("notification hook is nil")
)

const SignatureHeaderKey = "X-Hub-Signature-256"

// HandlerOptions are the options for the handler. They are used to configure the handler.
type HandlerOptions struct {
	ValidateSignature bool
	Secret            string
}

// NotificationHandler returns a http.Handler that can be used to handle the notification
// from the webhook. It calls ApplyHooks to apply the hooks to the notification.
// There are two ErrorHandlers, one for the notification and one for the hooks. The NotificationErrorHandler
// is expected to handle scenarios where the notification is malformed or the request is invalid.
// The HooksErrorHandler is expected to handle scenarios where the hooks after being applied return an error.
func NotificationHandler(
	hooks Hooks, nfh NotificationErrorHandler,
	eh HooksErrorHandler, options *HandlerOptions) http.Handler {
	if nfh == nil {
		nfh = NoOpNotificationErrorHandler
	}
	handler := func(writer http.ResponseWriter, request *http.Request) {
		if options != nil && options.ValidateSignature {
			var buff bytes.Buffer
			if _, err := io.Copy(&buff, request.Body); err != nil {
				if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				signature := request.Header.Get(SignatureHeaderKey)
				if !ValidateSignature(buff.Bytes(), signature, options.Secret) {
					if nErr := nfh(request.Context(), writer, request, ErrInvalidSignature); nErr != nil {
						writer.WriteHeader(http.StatusUnauthorized)
						return
					}
				}
			}
			request.Body = io.NopCloser(&buff)
		}

		if hooks == nil {
			if nErr := nfh(request.Context(), writer, request, ErrNilNotificationHook); nErr != nil {
				writer.WriteHeader(http.StatusAccepted)
				return
			}
		}

		// Construct the notification
		var notification Notification
		if err := json.NewDecoder(request.Body).Decode(&notification); err != nil {
			if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Apply the hooks
		if err := ApplyHooks(request.Context(), &notification, hooks, eh); err != nil {
			if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writer.WriteHeader(http.StatusAccepted)
	}
	return http.HandlerFunc(handler)
}

type (
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
)

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

type EventListener struct {
	h       Hooks
	hef     HooksErrorHandler
	neh     NotificationErrorHandler
	v       SubscriptionVerifier
	options *HandlerOptions
	g       GenericNotificationHandler
}

type ListenerOption func(*EventListener)

func NewEventListener(options ...ListenerOption) *EventListener {
	ls := &EventListener{
		options: &HandlerOptions{
			ValidateSignature: false,
		},
	}

	for _, option := range options {
		option(ls)
	}

	return ls
}

func WithGenericNotificationHandler(g GenericNotificationHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.g = g
	}
}

func WithHooks(hooks Hooks) ListenerOption {
	return func(ls *EventListener) {
		ls.h = hooks
	}
}

func WithHooksErrorHandler(hooksErrorHandler HooksErrorHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.hef = hooksErrorHandler
	}
}

func WithNotificationErrorHandler(notificationErrorHandler NotificationErrorHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.neh = notificationErrorHandler
	}
}

func WithSubscriptionVerifier(verifier SubscriptionVerifier) ListenerOption {
	return func(ls *EventListener) {
		ls.v = verifier
	}
}

func WithHandlerOptions(options *HandlerOptions) ListenerOption {
	return func(ls *EventListener) {
		ls.options = options
	}
}

// Handle returns a http.Handler that can be used to handle the notification
func (ls *EventListener) Handle() http.Handler {
	return NotificationHandler(ls.h, ls.neh, ls.hef, ls.options)
}

// GenericHandler returns a http.Handler that handles all type of notification in one function.
// It  calls GenericNotificationHandler. So before using this function, you should set GenericNotificationHandler
// with WithGenericNotificationHandler.
func (ls *EventListener) GenericHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var nfh NotificationErrorHandler
		if ls.neh == nil {
			nfh = NoOpNotificationErrorHandler
		} else {
			nfh = ls.neh
		}

		if ls.options != nil && ls.options.ValidateSignature {
			var buff bytes.Buffer
			if _, err := io.Copy(&buff, request.Body); err != nil {
				if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				signature := request.Header.Get(SignatureHeaderKey)
				if !ValidateSignature(buff.Bytes(), signature, ls.options.Secret) {
					if nErr := nfh(request.Context(), writer, request, ErrInvalidSignature); nErr != nil {
						writer.WriteHeader(http.StatusUnauthorized)
						return
					}
				}
			}
			request.Body = io.NopCloser(&buff)
		}

		// Construct the notification
		var notification Notification
		if err := json.NewDecoder(request.Body).Decode(&notification); err != nil {
			if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// call the generic handler
		if err := ls.g(request.Context(), &notification, nfh); err != nil {
			if nErr := nfh(request.Context(), writer, request, err); nErr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})
}

// Verify returns a http.Handler that can be used to verify the subscription
func (ls *EventListener) Verify() http.Handler {
	return VerifySubscriptionHandler(ls.v)
}

func attachHooksToMessage(ctx context.Context, nctx *NotificationContext, hooks Hooks, message *Message) error {
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
		return hooks.OnOrderReceived(ctx, nctx, mctx, message.Order)

	case ButtonMessageType:
		return hooks.OnButtonMessage(ctx, nctx, mctx, message.Button)

	case AudioMessageType:
		return hooks.OnAudioReceived(ctx, nctx, mctx, message.Audio)

	case VideoMessageType:
		return hooks.OnVideoReceived(ctx, nctx, mctx, message.Video)

	case ImageMessageType:
		return hooks.OnImageReceived(ctx, nctx, mctx, message.Image)

	case DocumentMessageType:
		return hooks.OnDocumentReceived(ctx, nctx, mctx, message.Document)

	case StickerMessageType:
		return hooks.OnStickerReceived(ctx, nctx, mctx, message.Sticker)

	case InteractiveMessageType:
		return hooks.OnInteractiveMessage(ctx, nctx, mctx, message.Interactive)

	case SystemMessageType:
		// TODO: documentation is not clear if the ID change will also be sent here:
		return hooks.OnSystemMessage(ctx, nctx, mctx, message.System)

	case UnknownMessageType:
		return hooks.OnMessageErrors(ctx, nctx, mctx, message.Errors)

	case TextMessageType:
		if message.Referral != nil {
			return hooks.OnReferralMessageReceived(ctx, nctx, mctx, message.Text, message.Referral)
		}

		// ProductEnquiry
		if mctx.Ctx != nil {
			return hooks.OnProductEnquiry(ctx, nctx, mctx, message.Text)
		}

		return hooks.OnTextMessageReceived(ctx, nctx, mctx, message.Text)

	case ReactionMessageType:
		return hooks.OnMessageReaction(ctx, nctx, mctx, message.Reaction)

	case LocationMessageType:
		return hooks.OnLocationReceived(ctx, nctx, mctx, message.Location)

	case ContactMessageType:
		return hooks.OnContactsReceived(ctx, nctx, mctx, message.Contacts)

	default:
		if message.Contacts != nil {
			if len(message.Contacts.Contacts) > 0 {
				return hooks.OnContactsReceived(ctx, nctx, mctx, message.Contacts)
			}
		}
		if message.Location != nil {
			return hooks.OnLocationReceived(ctx, nctx, mctx, message.Location)
		}

		if message.Identity != nil {
			return hooks.OnCustomerIDChange(ctx, nctx, mctx, message.Identity)
		}

		return fmt.Errorf("could not attach hook to this message")
	}

}

type GenericNotificationHandler func(context.Context, *Notification, NotificationErrorHandler) error
