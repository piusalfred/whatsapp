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
	"fmt"
	werrors "github.com/piusalfred/whatsapp/errors"
	"github.com/piusalfred/whatsapp/models"
)

var (
	ErrInvalidSignature    = fmt.Errorf("signature validation failed")
	ErrBodyReadFailed      = fmt.Errorf("failed to read request body")
	ErrBodyNil             = fmt.Errorf("request body is nil")
	ErrBodyUnmarshalFailed = fmt.Errorf("failed to unmarshal request body")
)

const (
	TextMessageEvent       Event = "text"
	ImageMessageEvent      Event = "image"
	VideoMessageEvent      Event = "video"
	AudioMessageEvent      Event = "audio"
	DocumentMessageEvent   Event = "document"
	StickerMessageEvent    Event = "sticker"
	LocationMessageEvent   Event = "location"
	ContactMessageEvent    Event = "contact"
	ReactionMessageEvent   Event = "reaction"
	ReplyButtonClickEvent  Event = "reply_button"
	CallToActionClickEvent Event = "call_to_action"
	ProfileUpdateEvent     Event = "profile_update"
	BusinessItemClickEvent Event = "business_item"
	ProductQueryEvent      Event = "product_query"
	ProductOrderEvent      Event = "product_order"
	StatusChangeEvent      Event = "status_change"
	UnknownEvent           Event = "unknown"
)

type (

	// Pricing An object containing billing information. It contains the following fields:
	//
	//	Billable, boolean – Indicates if the given message or conversation is billable. Default is true
	//  for all conversations, including those inside your free tier limit, except those initiated from
	//  free entry points. Free entry point conversatsion are not billable, false. You will not be charged
	//  for free tier limit conversations, but they are considered billable and will be reflected on your
	//  invoice. Deprecated. Visit the WhatsApp Changelog for more information.
	//
	//	Category, string – Indicates the conversation pricing category, one of the following:
	//		- business_initiated – The business sent a message to a customer more than 24 hours after the last customer message
	//		- referral_conversion – The conversation originated from a free entry point. These conversations are always customer-initiated.
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
	// Errors, array of errors.Error objects that contains information about the werrors that occurred.
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
		Errors       []*werrors.Error `json:"werrors,omitempty"`
	}

	// Event is the type of event that occurred and leads to the notification being sent.
	// You get a webhooks notification, When a customer performs one of the following an action
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

	// NotificationMessageType is the type of message that was sent to the webhook.
	// This is a filed in Message object. It can take the following values:
	// - audio
	// - button
	// - document
	// - text
	// - image
	// - interactive
	// - order
	// - sticker
	// - system – for customer number change messages
	// - unknown
	// - video
	// For interactive messages, there are two scenarios: when a user has
	// clicked a button and when a user has selected an item from a list.
	// The information of these Scenarios are ButtonReply and ListReply respectively.
	NotificationMessageType string

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
	// NewWaID - New WhatsApp ID for the customer when their phone number is updated. Available on webhook versions v11.0 and earlier.
	// WaID New WhatsApp ID for the customer when their phone number is updated. Available on webhook versions v12.0 and later.
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

	// Interactive ...
	Interactive struct {
		Type *InteractiveType `json:"type,omitempty"`
	}

	// InteractiveType ...
	// ButtonReply, sent when a customer clicks a button
	// ListReply, sent when a customer selects an item from a list
	InteractiveType struct {
		ButtonReply *ButtonReply `json:"button_reply,omitempty"` // sent when a customer clicks a button
		ListReply   *ListReply   `json:"list_reply,omitempty"`   // sent when a customer selects an item from a list
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

	// ProductItem ...
	// ProductRetailerID Unique identifier of the product in a catalog.
	// Quantity Number of items.
	// ItemPrice Price of each item.
	// Currency — Price currency.
	ProductItem struct {
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
		Quantity          string `json:"quantity,omitempty"`
		ItemPrice         string `json:"item_price,omitempty"`
		Currency          string `json:"currency,omitempty"`
	}

	// Order ...
	// CatalogID ID for the catalog the ordered item belongs to.
	// Text message from the user sent along with the order.
	// ProductItems Array of product item objects
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
	// Hash, hash — String. The ID for the messages system customer_identity_changed
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
	//	- ID, id — String. The message ID for the sent message for an inbound reply.
	//	- ReferredProduct, referred_product — Object. Referred product object describing the product the user is
	//	  requesting information about. You must parse this value if you support Product Enquiry Messages. See
	//	  Receive Response From Customers. Referred product objects have the following properties:
	//	  	- CatalogID, catalog_id — String. Unique identifier of the Meta catalog linked to the WhatsApp Business Account.
	//      - ProductRelailerID,product_retailer_id — String. Unique identifier of the product in a catalog.
	Context struct {
		Forwarded           bool   `json:"forwarded,omitempty"`
		FrequentlyForwarded bool   `json:"frequently_forwarded,omitempty"`
		From                string `json:"from,omitempty"`
		ID                  string `json:"id,omitempty"`
		ReferredProduct     *ReferredProduct
	}

	// ReferredProduct ,Referred product object describing the product the user is
	// requesting information about. You must parse this value if you support Product Enquiry Messages. See
	// Receive Response From Customers. Referred product objects have the following properties:
	//
	// CatalogID, catalog_id — String. Unique identifier of the Meta catalog linked to the WhatsApp Business Account.
	//
	// ProductRetailerID,product_retailer_id — String. Unique identifier of the product in a catalog.
	ReferredProduct struct {
		CatalogID         string `json:"catalog_id,omitempty"`
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
	}

	/*
		Value The value object contains details for the change that triggered the webhook. This object is nested
		within the Change array of the Entry array.

		- Contacts, contacts — Array of Contact objects with information for the customer who sent
		  a message to the business. Contact objects have the following properties:

		- Errors, errors — An array of error objects describing the error. Error objects have the
		  properties, which map to their equivalent properties in API error response payloads.

		- MessagingProduct messaging_product (string) Product used to send the message.
		  Value is always whatsapp.

		- Messages messages (array of objects) Information about a message received by
		  the business that is subscribed to the webhook. See Message Object.

		- Metadata metadata (object) A metadata object describing the business subscribed to
		  the webhook. See Metadata Object.

		- Statuses statuses (array of objects) Status object for a message that was sent by
		  the business that is subscribed to the webhook. See Status Object.
	*/
	Value struct {
		MessagingProduct string           `json:"messaging_product,omitempty"`
		Metadata         *Metadata        `json:"metadata,omitempty"`
		Errors           []*werrors.Error `json:"werrors,omitempty"`
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
