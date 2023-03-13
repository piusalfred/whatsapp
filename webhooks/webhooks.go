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
	// NotificationErrHandlerResponse is the response is returned by the NotificationErrorHandler instructing
	// how the http.Response sent to the whatsapp server should be.
	// Note that the NotificationErrorHandler can instruct the caller to ignore the error by setting the Skip
	// field to true. In this case the caller will just return http.StatusOK to whatsapp server.
	NotificationErrHandlerResponse struct {
		StatusCode int
		Headers    map[string]string
		Body       []byte
		Skip       bool
	}

	HooksErrorHandler func(err error) error
	// NotificationErrorHandler is a function that handles errors that occur when processing a notification.
	// The function returns a NotificationErrHandlerResponse that is sent to the whatsapp server.
	//
	// Note that retuning nil will make the default use http.StatusOK as the status code.
	//
	// Returning a status code that is not 200, will make a whatsapp server retry the notification. In some
	// cases this can lead to duplicate notifications. If your business logic is affected by this, you should
	// be careful when returning a non 200 status code.
	//
	// This is a snippet from the whatsapp documentation:
	//
	//		If we send a webhook request to your endpoint and your server responds with a http status code other
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
	// -  ErrOnGenericHandlerFunc when an error is received in the GenericHandlerFunc hook.
	NotificationErrorHandler func(context.Context, *http.Request, error) *NotificationErrHandlerResponse

	// BeforeFunc is a function that is called before a notification is processed. It receives the notification
	// and can return an error. If an error is returned, the notification is not processed and the error is
	// passed to the NotificationErrorHandler. A lot of use cases can be implemented using the BeforeFunc.
	// For example, you can use it to validate the notification, to check if it is a duplicate notification,
	// To check db availability etc.
	BeforeFunc func(ctx context.Context, notification *Notification) error

	// AfterFunc is a function that is called after a notification is processed. It also receives the error
	// that occurred during processing. There can be a number of use cases where the AfterFunc is useful.
	// For example, you can use it to log the error or send a notification to a monitoring service. Or have the
	// instrumentation logic put here.
	AfterFunc func(ctx context.Context, notification *Notification, err error)

	// HandlerOptions is a struct that contains the options that can be passed to the NotificationHandler. Note that
	// the options are optional. NotificationHandler can be used without any options set.
	HandlerOptions struct {
		BeforeFunc        BeforeFunc
		AfterFunc         AfterFunc
		ValidateSignature bool
		Secret            string
	}

	// VerificationRequest contains details sent by the whatsapp server during the verification process.
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

	// GlobalNotificationHandler is a function that handles all notifications. Use this function if you want to
	// create your own logic in handling different types of notifications. Because when this is used for receiving
	// notifications all types of notifications from Templates, Messages, Media, Contacts, etc. will be passed here,
	// and you can handle them as you wish.
	GlobalNotificationHandler func(context.Context, http.ResponseWriter, *Notification) error
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
func NoOpNotificationErrorHandler(_ context.Context, _ *http.Request, err error) *NotificationErrHandlerResponse {
	return &NotificationErrHandlerResponse{
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
	if hooks == nil || value == nil {
		return nil
	}

	notificationCtx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	// nonFatalErrors is a slice of non-fatal errors that are collected from the hooks.
	// can contain a maximum of 4 errors.
	nonFatalErrors := make([]error, 0, 4) //nolint:gomnd

	// call the Hooks
	if hooks.OnNotificationErrorHook != nil {
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

	if hooks.OnMessageStatusChangeHook != nil {
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

var ErrFailedToAttachHookToMessage = errors.New("could not attach hooks to message")

var errHooksOrMessageIsNil = fmt.Errorf("%w: hooks or message is nil", ErrFailedToAttachHookToMessage)

//nolint:cyclop
func attachHooksToMessage(ctx context.Context, nctx *NotificationContext, hooks *Hooks, message *Message) error {
	if hooks == nil || message == nil {
		return errHooksOrMessageIsNil
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

	case AudioMessageType, VideoMessageType, ImageMessageType, DocumentMessageType, StickerMessageType:
		return hooks.OnMediaMessageHook(ctx, nctx, mctx, message.Audio)

	case InteractiveMessageType:
		return hooks.OnInteractiveMessageHook(ctx, nctx, mctx, message.Interactive)

	case SystemMessageType:
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

		return ErrFailedToAttachHookToMessage
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

		if options != nil && options.BeforeFunc != nil {
			if bfe := options.BeforeFunc(ctx, notification); bfe != nil {
				err = fmt.Errorf("%w: %w", ErrOnBeforeFuncHook, bfe)
				if handleError(ctx, writer, request, neh, err) {
					return
				}
			}
		}

		if options != nil && options.ValidateSignature {
			signature, _ := ExtractSignatureFromHeader(request.Header)
			if !ValidateSignature(buff.Bytes(), signature, options.Secret) {
				if handleError(ctx, writer, request, neh, ErrInvalidSignature) {
					return
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
//	     - hub.verify_token A string that we grab from the SubscriptionVerificationHandler Token
//	       field in your app's App Dashboard.
//	       You will set this string when you complete the Webhooks configuration settings steps.
//
// Whenever your endpoint receives a verification request, it must:
//
//   - verify that the hub.verify_token value matches the string you set in the verification Token field
//     when you configure the Webhooks product in your App Dashboard.
//
//   - Respond with the hub.challenge value. If you are in your App Dashboard and configuring your Webhooks product
//     (and thus, triggering a Verification Request), the dashboard will indicate if your endpoint validated the request
//     correctly. If you are using the Graph APIs /app/subscriptions endpoint to configure the Webhooks product, the API
//     will indicate success or failure with a response.
func VerifySubscriptionHandler(verifier SubscriptionVerifier) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Retrieve the query parameters from the request.
		q := request.URL.Query()
		mode := q.Get("hub.mode")
		challenge := q.Get("hub.challenge")
		token := q.Get("hub.verify_token")
		if err := verifier(request.Context(), &VerificationRequest{
			Mode:      mode,
			Challenge: challenge,
			Token:     token,
		}); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(challenge))
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
	decodeSig, err := hex.DecodeString(signature)
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
	return hmac.Equal(decodeSig, expectedSignature)
}

var ErrSignatureNotFound = errors.New("signature not found")

// ExtractSignatureFromHeader extracts the signature from the header. A signature is a SHA256
// hash of the payload, encoded in hexadecimal and prefixed with sha256=. It is found in the
// X-Hub-Signature-256 header.
// The signature is used to verify the authenticity of the payload. This method is used to extract
// the actual signature from the header without the prefix.
func ExtractSignatureFromHeader(header http.Header) (string, error) {
	signature := header.Get(SignatureHeaderKey)
	if !strings.HasPrefix(signature, "sha256=") {
		return "",
			fmt.Errorf("signature is empty or does not have prefix \"sha256\" %w", ErrSignatureNotFound)
	}

	return signature[7:], nil
}
