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

var (
	ErrSignatureValidationFailed = fmt.Errorf("signature validation failed")
	ErrBodyReadFailed            = fmt.Errorf("failed to read request body")
	ErrBodyNil                   = fmt.Errorf("request body is nil")
	ErrBodyUnmarshalFailed       = fmt.Errorf("failed to unmarshal request body")
)

const (
	ReceivedTextMessage  Event = "received_text_message"
	ReceivedImage        Event = "received_image"
	ReceivedVideo        Event = "received_video"
	ReceivedAudio        Event = "received_audio"
	ReceivedDocument     Event = "received_document"
	ReceivedSticker      Event = "received_sticker"
	ReceivedLocation     Event = "received_location"
	ReceivedContact      Event = "received_contact"
	ReceivedReaction     Event = "received_reaction"
	ReplyButtonClicked   Event = "reply_button_clicked"
	CallToActionClicked  Event = "call_to_action_clicked"
	ProfileUpdated       Event = "profile_updated"
	BussinessItemClicked Event = "business_item_clicked"
	ProductQuery         Event = "product_query"
	ProductOrder         Event = "product_order"
	UnknownEvent         Event = "unknown"
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
		ID      string              `json:"id,omitempty"`
		Origin  *ConversationOrigin `json:"origin,omitempty"`
		Exipiry int                 `json:"expiration_timestamp,omitempty"`
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
		ButtonReply *InteractiveReplyType `json:"button_reply,omitempty"` // sent when a customer clicks a button
		ListReply   *InteractiveReplyType `json:"list_reply,omitempty"`   // sent when a customer selects an item from a list
	}

	// InteractiveType ...
	InteractiveReplyType struct {
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
		ReferredProduct     *ReferedProduct
	}

	// ReferredProduct,Referred product object describing the product the user is
	// requesting information about. You must parse this value if you support Product Enquiry Messages. See
	// Receive Response From Customers. Referred product objects have the following properties:
	//
	// CatalogID, catalog_id — String. Unique identifier of the Meta catalog linked to the WhatsApp Business Account.
	//
	// ProductRetailerID,product_retailer_id — String. Unique identifier of the product in a catalog.
	ReferedProduct struct {
		CatalogID         string `json:"catalog_id,omitempty"`
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
	}

	Value struct {
		MessagingProduct string           `json:"messaging_product,omitempty"`
		Metadata         *Metadata        `json:"metadata,omitempty"`
		Errors           []*werrors.Error `json:"werrors,omitempty"`
		Contacts         []*Contact       `json:"contacts,omitempty"`
		Messages         []*Message       `json:"messages,omitempty"`
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

	EventListener struct {
		h EventHandler
	}

	EventHandler interface {
		HandleError(context.Context, http.ResponseWriter, *http.Request, error) error
		HandleEvent(context.Context, http.ResponseWriter, *http.Request, *Notification) error
	}
)

func NewEventListener(h EventHandler) *EventListener {
	return &EventListener{
		h: h,
	}
}

// Make EventListener a http.Handler
func (el *EventListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	if r.Body == nil {
		nErr := el.h.HandleError(r.Context(), w, r, ErrBodyNil)
		if nErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	bodyBytes, readErr := io.ReadAll(r.Body)
	defer io.NopCloser(bytes.NewBuffer(bodyBytes))
	if readErr != nil {
		nErr := el.h.HandleError(r.Context(), w, r, errors.Join(readErr, ErrBodyReadFailed))
		if nErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	unmarshalErr := json.Unmarshal(bodyBytes, &notification)
	if unmarshalErr != nil {
		nErr := el.h.HandleError(r.Context(), w, r, errors.Join(unmarshalErr, ErrBodyUnmarshalFailed))
		if nErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	// pass the notification to the handler
	notifErr := el.h.HandleEvent(r.Context(), w, r, &notification)
	if notifErr != nil {
		nErr := el.h.HandleError(r.Context(), w, r, notifErr)
		if nErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
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
//	     - hub.verify_token A string that that we grab from the Verify Token field in your app's App Dashboard.
//	       You will set this string when you complete the Webhooks configuration settings steps.
//
// Whenever your endpoint receives a verification request, it must:
//
//   - Verify that the hub.verify_token value matches the string you set in the Verify Token field when you configure
//     the Webhooks product in your App Dashboard (you haven't set up this token string yet).
//
//   - Respond with the hub.challenge value. If you are in your App Dashboard and configuring your Webhooks product
//     (and thus, triggering a Verification Request), the dashboard will indicate if your endpoint validated the request
//     correctly. If you are using the Graph API's /app/subscriptions endpoint to configure the Webhooks product, the API
//     will indicate success or failure with a response.
func VerifySubscriptionHandler(verifier SubscriptionVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte(challenge))
	}
}

// ValidateSignature validates the signature of the payload. all Event Notification payloads are signed
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
func ValidateSignature(payload []byte, signature string, secret string) bool {

	// TODO: fix this
	// change the payload to escaped unicode version with lowercase hex digits
	// escaped := strconv.QuoteToASCII(string(payload))
	// // remove the quotes
	// unquoted, err := strconv.Unquote(escaped)
	// if err != nil {
	// 	return false
	// }
	// payload = []byte(unquoted)

	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(payload)
	sig := hex.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(signature))
}

// CategorizeEvent categorizes the event notification from message type. This determines the type of event that occurred.
// Note:
//   - The type of message that has been received by the business that has subscribed to Webhooks has
//     these possible values: audio, button, document, text, image, interactive, order, sticker, system
//     and unknown.
//     System messages are sent when a customer number changes and UnknownEvent messages are sent when the message type is not
//     recognized.
//
// For more info -> https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components#messages-object
func CategorizeEvent(messageType string) Event {
	switch strings.ToLower(messageType) {
	case "text":
		return ReceivedTextMessage
	case "image":
		return ReceivedImage
	case "audio":
		return ReceivedAudio
	case "document":
		return ReceivedDocument
	case "sticker":
		return ReceivedSticker
	case "video":
		return ReceivedVideo
	case "location":
		return ReceivedLocation
	case "contacts":
		return ReceivedContact
	default:
		return UnknownEvent
	}
}
