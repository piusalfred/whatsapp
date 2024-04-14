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

	"github.com/piusalfred/whatsapp/pkg/models"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type (

	// Pricing An object containing billing information. It contains the following fields:
	//
	//	Billable, boolean – Indicates if the given message or conversation is billable. Default is true
	//  for all conversations, including those inside your free tier limit, except those initiated from
	//  free entry points. Free entry point conversation are not billable, false. You will not be charged
	//  for free tier limit conversations, but they are considered billable and will be reflected on your
	//  invoice. Deprecated. Visit the WhatsApp Changelog for more information.
	//
	//	Category, string – Indicates the conversation pricing category, one of the following:
	//		- business_initiated – The business sent a message to a customer more than 24 hours after the last customer message
	//		- referral_conversion – The conversation originated from a free entry point.
	//		                        These conversations are always customer-initiated.
	//		- customer_initiated – The business replied to a customer message within 24 hours of the last customer message
	//
	//	PricingModel, string – Type of pricing model used by the business. Current supported value is CBP
	Pricing struct {
		Billable     bool   `json:"billable,omitempty"` // Deprecated
		Category     string `json:"category,omitempty"`
		PricingModel string `json:"pricing_model,omitempty"`
	}

	// ConversationOrigin represents the origin of a conversation. It can be either business_initiated,
	// customer_initiated or referral_conversion.
	// business_initiated – Indicates that the conversation started by a business sending the first message
	// to a customer. This applies any time it has been more than 24 hours since the last customer message.
	// customer_initiated – Indicates that the conversation started by a business replying to a customer
	// message. This applies only when the business reply is within 24 hours of the last customer message.
	// referral_conversion – Indicates that the conversation originated from a free entry point. These conversations
	// are always customer-initiated.
	ConversationOrigin struct {
		Type string `json:"type,omitempty"`
	}

	// Conversation represents information about the conversation. It has the following fields:
	// id – Represents the ID of the conversation the given status notification belongs to.
	// origin – Indicates who initiated the conversation
	// expiry – Indicates the time in seconds after which the conversation will expire.
	//
	// WhatsApp defines a conversation as a 24-hour session of messaging between a person and a business.
	// There is no limit on the number of messages that can be exchanged in the fixed 24-hour window.
	// The 24-hour conversation session begins when:
	//		- A business-initiated message is delivered to a customer
	//		- A business’ reply to a customer message is delivered
	//
	// The 24-hour conversation session is different from the 24-hour customer support window. The customer
	// support window is a rolling window that is refreshed when a customer-initiated message is delivered
	// to a business. Within the customer support window businesses can send free-form messages.
	// Any business-initiated message sent more than 24 hours after the last customer message must be a
	// template message.
	Conversation struct {
		ID     string              `json:"id,omitempty"`
		Origin *ConversationOrigin `json:"origin,omitempty"`
		Expiry int                 `json:"expiration_timestamp,omitempty"`
	}

	// Status contains information about the status of a message sent to a customer.
	// The Status object is nested within the Value object and is triggered when a message is sent or
	// delivered to a customer or the customer reads the delivered message sent by a business that is
	// subscribed to the Webhooks.
	//
	// The Status object has the following fields:
	//
	// ID, id string  The ID for the message that the business that is subscribed to the webhooks sent
	// to a customer
	//
	// RecipientID, recipient_id string The WhatsApp ID for the customer that the business, that is subscribed
	// to the webhooks, sent to the customer
	//
	// StatusValue, status string, status of the message. Which can be one of the following:
	//		- delivered – A webhook is triggered when a message sent by a business has been delivered
	//		- read – A webhook is triggered when a message sent by a business has been read
	//		- sent – A webhook is triggered when a business sends a message to a customer
	//
	// Timestamp , timestamp Unix timestamp Date for the status message.
	//
	// Conversation, conversation object that contains information about the conversation.
	//
	// Pricing, pricing object that contains information about the billing information.
	//
	// Errors, array of errors.Error objects that contains information about the errors that occurred.
	//
	// NOTE:
	//
	// For a status to be read, it must have been delivered. In some scenarios, such as when a user
	// is in the chat screen and a message arrives, the message is delivered and read almost
	// simultaneously. In this or other similar scenarios, the delivered notification will not be sent
	// back, as it is implied that a message has been delivered if it has been read. The reason for this
	// behavior is internal optimization.
	Status struct {
		ID           string           `json:"id,omitempty"`
		RecipientID  string           `json:"recipient_id,omitempty"`
		StatusValue  string           `json:"status,omitempty"`
		Timestamp    int              `json:"timestamp,omitempty"`
		Conversation *Conversation    `json:"conversation,omitempty"`
		Pricing      *Pricing         `json:"pricing,omitempty"`
		Errors       []*werrors.Error `json:"errors,omitempty"`
	}

	// Event is the type of event that occurred and leads to the notification being sent.
	// You get a webhooks' notification, When a customer performs one of the following an action
	//
	//  - Sends a text message to the business
	//  - Sends an image, video, audio, document, or sticker to the business
	//  - Sends contact information to the business
	//  - Sends location information to the business
	//  - Clicks a reply button set up by the business
	//  - Clicks a call-to-actions button on an Ad that Clicks to WhatsApp
	//  - Clicks an item on a business list
	//  - Updates their profile information such as their phone number
	//  - Asks for information about a specific product
	//  - Orders products being sold by the business
	Event string

	Metadata struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		PhoneNumberID      string `json:"phone_number_id,omitempty"`
	}

	Profile struct {
		Name string `json:"name,omitempty"`
	}

	Contact struct {
		Profile *Profile `json:"profile,omitempty"`
		WaID    string   `json:"wa_id,omitempty"`
	}

	// Message contains the information of a message. It is embedded in the Value object.
	Message struct {
		Audio       *models.MediaInfo `json:"audio,omitempty"`
		Button      *Button           `json:"button,omitempty"`
		Context     *Context          `json:"context,omitempty"`
		Document    *models.MediaInfo `json:"document,omitempty"`
		Errors      []*werrors.Error  `json:"werrors,omitempty"`
		From        string            `json:"from,omitempty"`
		ID          string            `json:"id,omitempty"`
		Identity    *Identity         `json:"identity,omitempty"`
		Image       *models.MediaInfo `json:"image,omitempty"`
		Interactive *Interactive      `json:"interactive,omitempty"`
		Order       *Order            `json:"order,omitempty"`
		Referral    *Referral         `json:"referral,omitempty"`
		Sticker     *models.MediaInfo `json:"sticker,omitempty"`
		System      *System           `json:"system,omitempty"`
		Text        *Text             `json:"text,omitempty"`
		Timestamp   string            `json:"timestamp,omitempty"`
		Type        string            `json:"type,omitempty"`
		Video       *models.MediaInfo `json:"video,omitempty"`
		Contacts    *models.Contacts  `json:"contacts,omitempty"`
		Location    *models.Location  `json:"location,omitempty"`
		Reaction    *models.Reaction  `json:"reaction,omitempty"`
	}

	// System When messages type is set to system, a customer has updated their phone number or profile information,
	// this object is included in the messages object. System objects have the following properties:
	//
	// Body - Describes the change to the customer's identity or phone number.
	// Identity - Hash for the identity fetched from server.
	// NewWaID - New WhatsApp ID for the customer when their phone number is updated.
	// Available on webhook versions v11.0 and earlier.
	// WaID New WhatsApp ID for the customer when their phone number is updated. Available on
	// webhook versions v12.0 and later.
	// Type – type of system update. Will be one of the following:.
	// 		- customer_changed_number – A customer changed their phone number.
	//		- customer_identity_changed – A customer changed their profile information.
	// Customer The WhatsApp ID for the customer prior to the update.
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

	// Interactive represent the interactive template.
	Interactive struct {
		Type *InteractiveType `json:"type,omitempty"`
	}

	// InteractiveType represent an item sent to user. It can be a reply button
	// (ButtonReply) or a list reply containing a list of items (ListReply).
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

	// ProductItem represents a product item, Whereas the ProductRetailerID is the unique identifier of
	// the product in a catalog. Quantity represents the number of items. ItemPrice represents the price
	// of a single item. Currency represents the price currency.
	ProductItem struct {
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
		Quantity          string `json:"quantity,omitempty"`
		ItemPrice         string `json:"item_price,omitempty"`
		Currency          string `json:"currency,omitempty"`
	}

	// Order have information about order created by the customer. Order objects have the following properties:
	// The CatalogID which is the ID for the catalog the ordered item belongs to. Text message from the user and
	// ProductItems which is an array of product item objects.
	Order struct {
		CatalogID    string         `json:"catalog_id,omitempty"`
		Text         string         `json:"text,omitempty"`
		ProductItems []*ProductItem `json:"product_items,omitempty"`
	}

	// Referral A customer clicked an ad that redirects them to WhatsApp, this object is included in
	// the Message object. Referral objects have the following properties:
	//
	// SourceURL – String. The Meta URL that leads to the ad or post clicked by the customer. Opening this
	// url takes you to the ad viewed by your customer.
	//
	// SourceType – String. The type of the ad’s source; ad or post.
	//
	// SourceID – String. Meta ID for an ad or a post.
	//
	// Headline – String. Headline used in the ad or post.
	//
	// Body – String. Body for the ad or post.
	//
	// MediaType – String. Media present in the ad or post; image or video.
	//
	// ImageURL – String. URL of the image, when media_type is an image.
	//
	// VideoURL – String. URL of the video, when media_type is a video.
	//
	// ThumbnailURL – String. URL for the thumbnail, when media_type is a video.
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
	}

	// Button embedded in the Message object. When the messages type field is set to button,
	// this object is included in the messages object:
	//
	// Payload, payload – String. The payload for a button set up by the business that a customer
	// clicked as part of an interactive message.
	//
	// Text, text — String. Button text.
	Button struct {
		Payload string `json:"payload,omitempty"`
		Text    string `json:"text,omitempty"`
	}

	// Identity Webhook is triggered when a customer's phone number or profile information has been updated.
	// See messages system identity. Identity objects can have the following properties:
	//
	// Acknowledged, acknowledged — State of acknowledgment for the messages system customer_identity_changed.
	//
	// CreatedTimestamp, created_timestamp — String. The time when the WhatsApp Business Management API detected
	// the customer may have changed their profile information.
	//
	// Hash, hash — String. The ID for the messages system customer_identity_changed.
	Identity struct {
		Acknowledged     bool   `json:"acknowledged,omitempty"`
		CreatedTimestamp int64  `json:"created_timestamp,omitempty"`
		Hash             string `json:"hash,omitempty"`
	}

	// Context object. Only included when a user replies or interacts with one of your messages. Context objects\
	// can have the following properties:
	//
	//	- Forwarded, forwarded — Boolean. Set to true if the message received by the business has been forwarded.
	//	- FrequentlyForwarded,frequently_forwarded — Boolean. Set to true if the message received by the business
	//	  has been forwarded more than 5 times.
	//	- From, from — String. The WhatsApp ID for the customer who replied to an inbound message.
	//	- ID, id — String. The message ID for the send message for an inbound reply.
	//	- ReferredProduct, referred_product — Object. Referred product object describing the product the user is
	//	  requesting information about. You must parse this value if you support Product Enquiry Messages. See
	//	  Receive NotificationErrHandlerResponse From Customers. Referred product objects have the following properties:
	//	  	- CatalogID, catalog_id — String. Unique identifier of the Meta catalog linked to the WhatsApp Business Account.
	//      - ProductRetailerID,product_retailer_id — String. Unique identifier of the product in a catalog.
	Context struct {
		Forwarded           bool   `json:"forwarded,omitempty"`
		FrequentlyForwarded bool   `json:"frequently_forwarded,omitempty"`
		From                string `json:"from,omitempty"`
		ID                  string `json:"id,omitempty"`
		ReferredProduct     *ReferredProduct
	}

	// ReferredProduct ,Referred product object describing the product the user is
	// requesting information about. You must parse this value if you support Product Enquiry Messages. See
	// Receive NotificationErrHandlerResponse From Customers. Referred product objects have the following properties:
	//
	// CatalogID, catalog_id — String. Unique identifier of the Meta catalog linked to the WhatsApp Business Account.
	//
	// ProductRetailerID,product_retailer_id — String. Unique identifier of the product in a catalog.
	ReferredProduct struct {
		CatalogID         string `json:"catalog_id,omitempty"`
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
	}

	// Value The value object contains details for the change that triggered the webhook. This object is nested
	// within the Change array of the Entry array.
	//
	//- Contacts, contacts — Array of Contact objects with information for the customer who sent
	//  a message to the business. Contact objects have the following properties:
	//
	//- Errors, errors — An array of error objects describing the error. Error objects have the
	//  properties, which map to their equivalent properties in API error response payloads.
	//
	//- MessagingProduct messaging_product (string) Product used to send the message.
	//  Value is always whatsapp.
	//
	//- Messages (array of objects) Information about a message received by
	//  the business that is subscribed to the webhook. See Message Object.
	//
	//- Metadata (object) A metadata object describing the business subscribed to
	//  the webhook. See Metadata Object.
	//
	//- Statuses (array of objects) Status object for a message that was sent by
	//  the business that is subscribed to the webhook. See Status Object.
	Value struct {
		MessagingProduct string           `json:"messaging_product,omitempty"`
		Metadata         *Metadata        `json:"metadata,omitempty"`
		Errors           []*werrors.Error `json:"errors,omitempty"`
		Contacts         []*Contact       `json:"contacts,omitempty"`
		Messages         []*Message       `json:"messages,omitempty"`
		Statuses         []*Status        `json:"statuses,omitempty"`
	}

	Change struct {
		Value *Value `json:"value,omitempty"`
		Field string `json:"field,omitempty"`
	}

	Entry struct {
		ID      string    `json:"id,omitempty"`
		Changes []*Change `json:"changes,omitempty"`
	}

	Notification struct {
		Object string   `json:"object,omitempty"`
		Entry  []*Entry `json:"entry,omitempty"`
	}
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
	MessageTypeAudio       MessageType = "audio"
	MessageTypeButton      MessageType = "button"
	MessageTypeDocument    MessageType = "document"
	MessageTypeText        MessageType = "text"
	MessageTypeImage       MessageType = "image"
	MessageTypeInteractive MessageType = "interactive"
	MessageTypeOrder       MessageType = "order"
	MessageTypeSticker     MessageType = "sticker"
	MessageTypeSystem      MessageType = "system"
	MessageTypeUnknown     MessageType = "unknown"
	MessageTypeVideo       MessageType = "video"
	MessageTypeLocation    MessageType = "location"
	MessageTypeReaction    MessageType = "reaction"
	MessageTypeContacts    MessageType = "contacts"
)

const (
	InteractiveReplyList   InteractiveReply = "list_reply"
	InteractiveReplyButton InteractiveReply = "button_reply"
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

	// MessageStatus is the status of a message.
	// delivered – A webhook is triggered when a message sent by a business has been delivered
	// read – A webhook is triggered when a message sent by a business has been read
	// sent – A webhook is triggered when a business sends a message to a customer.
	MessageStatus string
)

// ParseMessageType parses the message type from a string.
func ParseMessageType(s string) MessageType {
	msgMap := map[string]MessageType{
		"audio":       MessageTypeAudio,
		"button":      MessageTypeButton,
		"document":    MessageTypeDocument,
		"text":        MessageTypeText,
		"image":       MessageTypeImage,
		"interactive": MessageTypeInteractive,
		"order":       MessageTypeOrder,
		"sticker":     MessageTypeSticker,
		"system":      MessageTypeSystem,
		"unknown":     MessageTypeUnknown,
		"video":       MessageTypeVideo,
		"location":    MessageTypeLocation,
		"reaction":    MessageTypeReaction,
		"contacts":    MessageTypeContacts,
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
	// NotificationErrorHandler is expected at least to receive errors from HandleNotification these errors are
	//
	// -  ErrBeforeFunc when an error is received in the BeforeFunc hook
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

	// HandlerOptions is a struct that contains the options that can be passed to the HandleNotification. Note that
	// the options are optional. HandleNotification can be used without any options set.
	HandlerOptions struct {
		BeforeFunc            BeforeFunc
		AfterFunc             AfterFunc
		ShouldValidatePayload bool
		AppSecret             string
	}

	PayloadValidationOptions struct {
		AppSecret      string
		ShouldValidate bool
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

	// GeneralNotificationHandler is a function that handles all notifications. Use this function if you want to
	// create your own logic in handling different types of notifications. Because when this is used for receiving
	// notifications all types of notifications from Templates, Messages, Media, Contacts, etc. will be passed here,
	// and you can handle them as you wish.
	GeneralNotificationHandler func(context.Context, *Notification) *Response

	Response struct {
		StatusCode int
	}

	// Config contains the configuration for the webhook listener. It contains
	//
	// - AppSecret: the application secret used to validate the incoming requests.
	//
	// - ShouldValidate: a flag that determines if the incoming requests should be validated.
	//
	// - VerifyToken: the token used to verify the subscription request.
	Config struct {
		AppSecret      string
		ShouldValidate bool
		VerifyToken    string
	}

	ConfigReaderFunc func() (*Config, error)

	// NotificationListener contains nuts and bolts needed to craft a webhook listener.
	NotificationListener struct {
		handlers             *Handlers
		handlersErrorHandler func(ctx context.Context, notification *Notification, err error) *Response
		v                    SubscriptionVerifier
		after                AfterFunc
		before               BeforeFunc
		g                    GeneralNotificationHandler
		pv                   *PayloadValidationOptions
		subVerifyToken       string
	}

	ListenerOption func(*NotificationListener)
)

// NewWithConfigReader returns a new NotificationListener with the provided config reader.
func NewWithConfigReader(reader ConfigReaderFunc, options ...ListenerOption) (*NotificationListener, error) {
	config, err := reader()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return NewListener(config, options...), nil
}

// NewGeneralListener returns a new NotificationListener with the provided options.
func NewGeneralListener(config *Config, g GeneralNotificationHandler) *NotificationListener {
	return &NotificationListener{
		pv: &PayloadValidationOptions{
			AppSecret:      config.AppSecret,
			ShouldValidate: config.ShouldValidate,
		},
		g:              g,
		subVerifyToken: config.VerifyToken,
	}
}

func NewListener(config *Config, options ...ListenerOption) *NotificationListener {
	listener := &NotificationListener{
		handlers: nil,
		v:        nil,
		after:    nil,
		before:   nil,
		pv: &PayloadValidationOptions{
			AppSecret:      config.AppSecret,
			ShouldValidate: config.ShouldValidate,
		},
		g:              nil,
		subVerifyToken: config.VerifyToken,
	}

	for _, option := range options {
		if option != nil {
			option(listener)
		}
	}

	return listener
}

// ExtractAndValidateSignature extracts the signature from the header and validates it.
func (listener *NotificationListener) ExtractAndValidateSignature(header http.Header, body []byte) error {
	signature, err := ExtractSignatureFromHeader(header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	if err := ValidateSignature(body, signature, listener.pv.AppSecret); err != nil {
		return fmt.Errorf("%w: %w", ErrSignatureVerification, err)
	}

	return nil
}

// HandleNotificationX handles all the notification types.
func (listener *NotificationListener) HandleNotificationX(writer http.ResponseWriter, request *http.Request) {
	buff, err := readNotificationBuffer(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	var (
		notification Notification
		ctx          = request.Context()
	)

	if listener.pv.ShouldValidate {
		if err := listener.ExtractAndValidateSignature(request.Header, buff.Bytes()); err != nil {

			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}
	}

	if err := json.NewDecoder(buff).Decode(&notification); err != nil && !errors.Is(err, io.EOF) {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	response := listener.g(ctx, &notification)
	if response != nil {
		writer.WriteHeader(response.StatusCode)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// HandleNotification handles the notification.
func (listener *NotificationListener) HandleNotification(writer http.ResponseWriter, request *http.Request) {
	var (
		notification Notification
		ctx          = request.Context()
		err          error
	)

	defer func() {
		if listener.after != nil {
			listener.after(ctx, &notification, err)
		}
	}()

	buff, err := readNotificationBuffer(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	if err := json.NewDecoder(buff).Decode(&notification); err != nil && !errors.Is(err, io.EOF) {
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	if listener.before != nil {
		if bfe := listener.before(ctx, &notification); bfe != nil {
			err = fmt.Errorf("%w: %w", ErrBeforeFunc, bfe)

			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}

	if listener.pv.ShouldValidate {
		if err := listener.ExtractAndValidateSignature(request.Header, buff.Bytes()); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)

			return
		}
	}

	if err := listener.passNotificationToHandlers(ctx, &notification); err != nil {
		response := listener.handlersErrorHandler(ctx, &notification, err)

		if response != nil {
			writer.WriteHeader(response.StatusCode)

			return
		}

		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// HandleSubscriptionVerification verifies the subscription to the webhooks.
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
func (listener *NotificationListener) HandleSubscriptionVerification(writer http.ResponseWriter, request *http.Request) {
	q := request.URL.Query()
	mode := q.Get("hub.mode")
	challenge := q.Get("hub.challenge")
	token := q.Get("hub.verify_token")

	if token != listener.subVerifyToken || mode != "subscribe" {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(challenge))
}

// ValidateSignature validates the signature of the payload. All Event Notification payloads are signed
// with a SHA256 signature and include the signature in the request's X-Hub-Signature-256 header, preceded
// with sha256=. You don't have to validate the payload, but you should.
//
// To validate the payload:
//  1. Generate a SHA256 signature using the payload and your app's App AppSecret.
//  2. Compare your signature to the signature in the X-Hub-Signature-256 header (everything after sha256=).
//
// If the signatures match, the payload is genuine. Please note that we generate the signature using an escaped
// unicode version of the payload, with lowercase hex digits. If you just calculate against the decoded bytes,
// you will end up with a different signature.
// For example, the string äöå should be escaped to \u00e4\u00f6\u00e5.
func ValidateSignature(payload []byte, signature, secret string) error {
	decodeSig, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("error decoding signature: %w", err)
	}

	// Calculate the expected signature using the payload and secret
	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write(payload)
	if err != nil {
		return fmt.Errorf("error hashing payload: %w", err)
	}
	expectedSignature := mac.Sum(nil)

	// Compare the expected and actual signatures
	if !hmac.Equal(decodeSig, expectedSignature) {
		return ErrInvalidSignature
	}

	return nil
}

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

// readNotificationBuffer returns notification content as a bytes buffer.
func readNotificationBuffer(r *http.Request) (*bytes.Buffer, error) {
	var buff bytes.Buffer
	_, err := io.Copy(&buff, r.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	// restore the request body to its original state
	r.Body = io.NopCloser(&buff)

	return &buff, nil
}

// webhookError is a custom error type for webhook errors.
type webhookError string

func (e webhookError) Error() string {
	return string(e)
}

const (
	ErrBeforeFunc                         = webhookError("before func failed")
	ErrInvalidSignature                   = webhookError("signature is invalid")
	ErrBadRequest                         = webhookError("could not retrieve the notification content")
	ErrSignatureNotFound                  = webhookError("signature not found")
	ErrSignatureVerification              = webhookError("signature verification failed")
	ErrHandleMessage                      = webhookError("could not handle message")
	ErrNotificationErrorHandler           = webhookError("notification error handler failed")
	ErrOrderMessageHandler                = webhookError("order message handler failed")
	ErrButtonMessageHandler               = webhookError("button message handler failed")
	ErrMediaMessageHandler                = webhookError("media message handler failed")
	ErrTextMessageHandler                 = webhookError("text message handler failed")
	ErrInteractiveHandler                 = webhookError("interactive message handler failed")
	ErrSystemMessageHandler               = webhookError("system message handler failed")
	ErrReferralMessage                    = webhookError("referral message handler failed")
	ErrMessageReaction                    = webhookError("message reaction handler failed")
	ErrLocationMessage                    = webhookError("location message handler failed")
	ErrContactsMessage                    = webhookError("contacts message handler failed")
	ErrCustomerIDChange                   = webhookError("customer id change handler failed")
	ErrProductEnquiry                     = webhookError("product enquiry handler failed")
	ErrUnknownMessageHandler              = webhookError("unknown message handler failed")
	ErrContactsMessageHandler             = webhookError("contacts message handler failed")
	ErrMessageStatusChangeHandler         = webhookError("message status change handler failed")
	ErrMessageReceivedNotificationHandler = webhookError("message received notification handler failed")
)

// passNotificationToHandlers passes the notification to the handlers.
func (listener *NotificationListener) passNotificationToHandlers(ctx context.Context, notification *Notification) error {
	if notification == nil || listener.handlers == nil {
		return nil
	}

	for _, entry := range notification.Entry {
		if err := listener.passEntryToHandlers(ctx, entry); err != nil {
			return err
		}
	}

	return nil
}

// passEntryToHandlers passes the entry to the handlers.
func (listener *NotificationListener) passEntryToHandlers(ctx context.Context, entry *Entry) error {
	entryID := entry.ID
	changes := entry.Changes
	for _, change := range changes {
		value := change.Value
		if value == nil {
			continue
		}
		if err := listener.passValueToHandlers(ctx, entryID, value); err != nil {
			return err
		}
	}

	return nil
}

//nolint:cyclop
func (listener *NotificationListener) passValueToHandlers(ctx context.Context, id string, value *Value) error {
	handlers := listener.handlers

	if handlers == nil || value == nil {
		return nil
	}

	notificationCtx := &NotificationContext{
		ID:       id,
		Contacts: value.Contacts,
		Metadata: value.Metadata,
	}

	// call the Handlers
	if handlers.NotificationError != nil {
		for _, ev := range value.Errors {
			if err := handlers.NotificationError.Handle(ctx, notificationCtx, ev); err != nil {
				return fmt.Errorf("%w: %w", ErrNotificationErrorHandler, err)
			}
		}
	}

	if handlers.MessageStatusChange != nil {
		for _, sv := range value.Statuses {
			if err := handlers.MessageStatusChange.Handle(ctx, notificationCtx, sv); err != nil {
				return fmt.Errorf("%w: %w", ErrMessageStatusChangeHandler, err)
			}
		}
	}

	for _, mv := range value.Messages {
		if handlers.MessageReceived != nil {
			if err := handlers.MessageReceived.Handle(ctx, notificationCtx, mv); err != nil {
				return fmt.Errorf("%w: %w", ErrMessageReceivedNotificationHandler, err)
			}
		}

		if err := listener.passMessageToHandlers(ctx, notificationCtx, mv); err != nil {
			return err
		}
	}

	return nil
}

func (listener *NotificationListener) passMessageToHandlers(ctx context.Context, nctx *NotificationContext, message *Message) error {
	mctx := &MessageContext{
		From:      message.From,
		ID:        message.ID,
		Timestamp: message.Timestamp,
		Type:      message.Type,
		Ctx:       message.Context,
	}

	messageType := ParseMessageType(message.Type)
	switch messageType {
	case MessageTypeOrder:
		if err := listener.handlers.OrderMessage.Handle(ctx, nctx, mctx, message.Order); err != nil {
			return fmt.Errorf("%w: %w", ErrOrderMessageHandler, err)
		}

		return nil

	case MessageTypeButton:
		if err := listener.handlers.ButtonMessage.Handle(ctx, nctx, mctx, message.Button); err != nil {
			return fmt.Errorf("%w: %w", ErrButtonMessageHandler, err)
		}

		return nil

	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		if err := listener.handlers.MediaMessage.Handle(ctx, nctx, mctx, message.Audio); err != nil {
			return fmt.Errorf("%w: %w", ErrMediaMessageHandler, err)
		}

		return nil

	case MessageTypeInteractive:
		if err := listener.handlers.InteractiveMessage.Handle(ctx, nctx, mctx, message.Interactive); err != nil {
			return fmt.Errorf("%w: %w", ErrInteractiveHandler, err)
		}

		return nil

	case MessageTypeSystem:
		if err := listener.handlers.SystemMessage.Handle(ctx, nctx, mctx, message.System); err != nil {
			return fmt.Errorf("%w: %w", ErrSystemMessageHandler, err)
		}

		return nil

	case MessageTypeUnknown:
		if err := listener.handlers.MessageErrors.Handle(ctx, nctx, mctx, message.Errors); err != nil {
			return fmt.Errorf("%w: %w", ErrUnknownMessageHandler, err)
		}

		return nil

	case MessageTypeText:
		if message.Referral != nil {
			if err := listener.handlers.ReferralMessage.Handle(ctx, nctx, mctx, message.Text, message.Referral); err != nil {
				return fmt.Errorf("%w: %w", ErrReferralMessage, err)
			}

			return nil
		}

		if mctx.Ctx != nil {
			if err := listener.handlers.ProductEnquiry.Handle(ctx, nctx, mctx, message.Text); err != nil {
				return fmt.Errorf("%w: %w", ErrProductEnquiry, err)
			}

			return nil
		}

		if err := listener.handlers.TextMessage.Handle(ctx, nctx, mctx, message.Text); err != nil {
			return fmt.Errorf("%w: %w", ErrTextMessageHandler, err)
		}

		return nil

	case MessageTypeReaction:
		if err := listener.handlers.MessageReaction.Handle(ctx, nctx, mctx, message.Reaction); err != nil {
			return fmt.Errorf("%w: %w", ErrMessageReaction, err)
		}

		return nil

	case MessageTypeLocation:
		if err := listener.handlers.LocationMessage.Handle(ctx, nctx, mctx, message.Location); err != nil {
			return fmt.Errorf("%w: %w", ErrLocationMessage, err)
		}

		return nil

	case MessageTypeContacts:
		if err := listener.handlers.ContactsMessage.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
			return fmt.Errorf("%w: %w", ErrContactsMessage, err)
		}

		return nil

	default:
		if message.Contacts != nil {
			if err := listener.handlers.ContactsMessage.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
				return fmt.Errorf("%w: %w", ErrContactsMessageHandler, err)
			}

			return nil
		}
		if message.Location != nil {
			if err := listener.handlers.LocationMessage.Handle(ctx, nctx, mctx, message.Location); err != nil {
				return fmt.Errorf("%w: %w", ErrLocationMessage, err)
			}

			return nil
		}

		if message.Identity != nil {
			if err := listener.handlers.CustomerIDChange.Handle(ctx, nctx, mctx, message.Identity); err != nil {
				return fmt.Errorf("%w: %w", ErrCustomerIDChange, err)
			}

			return nil
		}

		return fmt.Errorf("%w: unsupported message type", ErrHandleMessage)
	}
}

type (
	// Handlers is a struct that contains all the hooks that can be attached to a notification.
	// OnNotificationErrorHook is the OnNotificationErrorHook called when an error is received
	// in a notification.
	Handlers struct {
		OrderMessage        OrderMessageHandler
		ButtonMessage       ButtonMessageHandler
		LocationMessage     LocationMessageHandler
		ContactsMessage     ContactsMessageHandler
		MessageReaction     MessageReactionHandler
		UnknownMessage      UnknownMessageHandler
		ProductEnquiry      ProductEnquiryHandler
		InteractiveMessage  InteractiveMessageHandler
		MessageErrors       MessageErrorsHandler
		TextMessage         TextMessageHandler
		ReferralMessage     ReferralMessageHandler
		CustomerIDChange    CustomerIDChangeMessageHandler
		SystemMessage       SystemMessageHandler
		MediaMessage        MediaMessageHandler
		NotificationError   MessageErrorHandler
		MessageStatusChange MessageStatusChangeHandler
		MessageReceived     MessageReceivedHandler
	}
)

type (
	OnLocationMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, location *models.Location) error

	LocationMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, location *models.Location) error
	}
)

func (h OnLocationMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	location *models.Location,
) error {
	return h(ctx, nctx, mctx, location)
}

type (
	OnTextMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error

	TextMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	}
)

func (h OnTextMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	text *Text,
) error {
	return h(ctx, nctx, mctx, text)
}

type (
	OnOrderMessageHook func(
		context.Context, *NotificationContext, *MessageContext, *Order) error

	OrderMessageHandler interface {
		Handle(
			ctx context.Context,
			nctx *NotificationContext,
			mctx *MessageContext,
			order *Order) error
	}
)

func (h OnOrderMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	order *Order,
) error {
	return h(ctx, nctx, mctx, order)
}

type (
	OnButtonMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, button *Button) error

	ButtonMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, button *Button) error
	}
)

func (h OnButtonMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	button *Button,
) error {
	return h(ctx, nctx, mctx, button)
}

type (
	OnContactsMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, contacts *models.Contacts) error

	ContactsMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, contacts *models.Contacts) error
	}
)

func (h OnContactsMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	contacts *models.Contacts,
) error {
	return h(ctx, nctx, mctx, contacts)
}

type (
	OnMessageReactionHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, reaction *models.Reaction) error

	MessageReactionHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, reaction *models.Reaction) error
	}
)

func (h OnMessageReactionHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	reaction *models.Reaction,
) error {
	return h(ctx, nctx, mctx, reaction)
}

type (
	OnUnknownMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error

	UnknownMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	}
)

func (h OnUnknownMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	errors []*werrors.Error,
) error {
	return h(ctx, nctx, mctx, errors)
}

type (
	OnProductEnquiryHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error

	ProductEnquiryHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text) error
	}
)

func (h OnProductEnquiryHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	text *Text,
) error {
	return h(ctx, nctx, mctx, text)
}

type (
	OnInteractiveMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, interactive *Interactive) error

	InteractiveMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, interactive *Interactive) error
	}
)

func (h OnInteractiveMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	interactive *Interactive,
) error {
	return h(ctx, nctx, mctx, interactive)
}

type (
	OnMessageErrorsHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error

	MessageErrorsHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, errors []*werrors.Error) error
	}
)

func (h OnMessageErrorsHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	errors []*werrors.Error,
) error {
	return h(ctx, nctx, mctx, errors)
}

type (
	OnReferralMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text, referral *Referral) error

	ReferralMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, text *Text, referral *Referral) error
	}
)

func (h OnReferralMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	text *Text,
	referral *Referral,
) error {
	return h(ctx, nctx, mctx, text, referral)
}

type (
	OnCustomerIDChangeMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, customerID *Identity) error

	CustomerIDChangeMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, customerID *Identity) error
	}
)

func (h OnCustomerIDChangeMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	customerID *Identity,
) error {
	return h(ctx, nctx, mctx, customerID)
}

type (
	OnSystemMessageHook func(
		ctx context.Context, nctx *NotificationContext, mctx *MessageContext, system *System) error

	SystemMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, system *System) error
	}
)

func (h OnSystemMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	system *System,
) error {
	return h(ctx, nctx, mctx, system)
}

type (
	// OnMediaMessageHook is a hook that is called when a media message is received. This is when Message.Type is
	// image, audio, video or document or sticker.
	OnMediaMessageHook func(ctx context.Context, nctx *NotificationContext, mctx *MessageContext,
		media *models.MediaInfo) error

	MediaMessageHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, mctx *MessageContext, media *models.MediaInfo) error
	}
)

func (h OnMediaMessageHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	mctx *MessageContext,
	media *models.MediaInfo,
) error {
	return h(ctx, nctx, mctx, media)
}

type (
	// OnNotificationErrorHook is a hook that is called when an error is received in a notification.
	// This is called when an error is received in a notification. This is not called when an error
	// is received in a message, that is handled by NotificationHooks.Handle.
	OnNotificationErrorHook func(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

	// MessageErrorHandler is an interface that is implemented by a type that wants to handle notification errors.
	MessageErrorHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error
	}
)

// Handle implements the MessageErrorHandler interface.
func (h OnNotificationErrorHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	errors *werrors.Error,
) error {
	return h(ctx, nctx, errors)
}

type (

	// OnMessageStatusChangeHook is a hook that is called when a there is a notification about a message status change.
	// This is called when a message status changes. For example, when a message is delivered or read.
	OnMessageStatusChangeHook func(ctx context.Context, nctx *NotificationContext, status *Status) error

	MessageStatusChangeHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, status *Status) error
	}
)

// Handle implements the MessageStatusChangeHandler interface.
func (h OnMessageStatusChangeHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	status *Status,
) error {
	return h(ctx, nctx, status)
}

type (
	// OnMessageReceivedHook is a hook that is called when a message is received. A notification
	// can contain a lot of things like errors status changes etc. This is called when a
	// notification contains a message. This work with the
	// Message in general. The Handlers for specific message types are called after this hook. They are all implemented
	// in the MessageHooks interface.
	OnMessageReceivedHook func(ctx context.Context, nctx *NotificationContext, message *Message) error

	MessageReceivedHandler interface {
		Handle(ctx context.Context, nctx *NotificationContext, message *Message) error
	}
)

func (h OnMessageReceivedHook) Handle(
	ctx context.Context,
	nctx *NotificationContext,
	message *Message,
) error {
	return h(ctx, nctx, message)
}

func (listener *NotificationListener) GenericNotificationHandler(handler GeneralNotificationHandler) {
	listener.g = handler
}

func (listener *NotificationListener) SubscriptionVerifier(verifier SubscriptionVerifier) {
	listener.v = verifier
}

func (listener *NotificationListener) OnOrderMessage(hook OnOrderMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.OrderMessage = hook
}

func (listener *NotificationListener) OnButtonMessage(hook OnButtonMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.ButtonMessage = hook
}

func (listener *NotificationListener) OnLocationMessage(hook OnLocationMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.LocationMessage = hook
}

func (listener *NotificationListener) OnContactsMessage(hook OnContactsMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.ContactsMessage = hook
}

func (listener *NotificationListener) OnMessageReaction(hook OnMessageReactionHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.MessageReaction = hook
}

func (listener *NotificationListener) OnUnknownMessage(hook OnUnknownMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.UnknownMessage = hook
}

func (listener *NotificationListener) OnProductEnquiry(hook OnProductEnquiryHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.ProductEnquiry = hook
}

func (listener *NotificationListener) OnInteractiveMessage(hook OnInteractiveMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.InteractiveMessage = hook
}

func (listener *NotificationListener) OnMessageErrors(hook OnMessageErrorsHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.MessageErrors = hook
}

func (listener *NotificationListener) OnTextMessage(hook OnTextMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.TextMessage = hook
}

func (listener *NotificationListener) OnReferralMessage(hook OnReferralMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.ReferralMessage = hook
}

func (listener *NotificationListener) OnCustomerIDChange(hook OnCustomerIDChangeMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.CustomerIDChange = hook
}

func (listener *NotificationListener) OnSystemMessage(hook OnSystemMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.SystemMessage = hook
}

func (listener *NotificationListener) OnMediaMessage(hook OnMediaMessageHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.MediaMessage = hook
}

func (listener *NotificationListener) OnNotificationError(hook OnNotificationErrorHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.NotificationError = hook
}

func (listener *NotificationListener) OnMessageStatusChange(hook OnMessageStatusChangeHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.MessageStatusChange = hook
}

func (listener *NotificationListener) OnMessageReceived(hook OnMessageReceivedHook) {
	if listener.handlers == nil {
		listener.handlers = &Handlers{}
	}
	listener.handlers.MessageReceived = hook
}

func WithGlobalNotificationHandler(g GeneralNotificationHandler) ListenerOption {
	return func(ls *NotificationListener) {
		ls.g = g
	}
}

func WithHooks(hooks *Handlers) ListenerOption {
	return func(ls *NotificationListener) {
		ls.handlers = hooks
	}
}

func WithSubscriptionVerifier(verifier SubscriptionVerifier) ListenerOption {
	return func(ls *NotificationListener) {
		ls.v = verifier
	}
}

func WithBeforeFunc(beforeFunc BeforeFunc) ListenerOption {
	return func(ls *NotificationListener) {
		ls.before = beforeFunc
	}
}

func WithAfterFunc(afterFunc AfterFunc) ListenerOption {
	return func(ls *NotificationListener) {
		ls.after = afterFunc
	}
}
