package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/piusalfred/whatsapp/errors"
	"github.com/piusalfred/whatsapp/models"
	"net/http"
)

var (
	ErrSignatureValidationFailed = fmt.Errorf("signature validation failed")
)

const (
	ReceivedTextMessage  EventType = "received_text_message"
	ReceivedImage        EventType = "received_image"
	ReceivedVideo        EventType = "received_video"
	ReceivedAudio        EventType = "received_audio"
	ReceivedDocument     EventType = "received_document"
	ReceivedSticker      EventType = "received_sticker"
	ReceivedLocation     EventType = "received_location"
	ReceivedContact      EventType = "received_contact"
	ReplyButtonClicked   EventType = "reply_button_clicked"
	CallToActionClicked  EventType = "call_to_action_clicked"
	ProfileUpdated       EventType = "profile_updated"
	BussinessItemClicked EventType = "business_item_clicked"
	ProductQuery         EventType = "product_query"
	ProductOrder         EventType = "product_order"
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
		ID           string          `json:"id,omitempty"`
		RecipientID  string          `json:"recipient_id,omitempty"`
		StatusValue  string          `json:"status,omitempty"`
		Timestamp    int             `json:"timestamp,omitempty"`
		Conversation *Conversation   `json:"conversation,omitempty"`
		Pricing      *Pricing        `json:"pricing,omitempty"`
		Errors       []*errors.Error `json:"errors,omitempty"`
	}

	// EventType is the type of event that occurred and leads to the notification being sent.
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
	EventType string
	Metadata  struct {
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

	Message struct {
		From      string            `json:"from,omitempty"`
		ID        string            `json:"id,omitempty"`
		Timestamp int64             `json:"timestamp,omitempty"`
		Type      string            `json:"type,omitempty"`
		Errors    []*errors.Error   `json:"errors,omitempty"`
		Text      *models.Text      `json:"text,omitempty"`
		Location  *models.Location  `json:"location,omitempty"`
		Recation  *models.Reaction  `json:"reaction,omitempty"`
		Sticker   *models.MediaInfo `json:"sticker,omitempty"`
		Image     *models.MediaInfo `json:"image,omitempty"`
		Video     *models.MediaInfo `json:"video,omitempty"`
		Audio     *models.MediaInfo `json:"audio,omitempty"`
		Document  *models.MediaInfo `json:"document,omitempty"`
		Contacts  *models.Contacts  `json:"contacts,omitempty"`
	}

	Value struct {
		MessagingProduct string          `json:"messaging_product,omitempty"`
		Metadata         *Metadata       `json:"metadata,omitempty"`
		Errors           []*errors.Error `json:"errors,omitempty"`
		Contacts         []*Contact      `json:"contacts,omitempty"`
		Messages         []*Message      `json:"messages,omitempty"`
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

	// EventListener is a function that processes the event notification. It is an extension of the
	// http.HandlerFunc type. It accepts a context, http.ResponseWriter, *http.Request, NotificationType,
	// ErrorHandlerFunc func(err error) error and returns an error.
	EventListener func(context.Context, http.ResponseWriter, *http.Request,
		EventType, NotificationHandler, ErrorHandlerFunc) error

	// ErrorHandlerFunc is a function that processes the error returned by the EventListener.
	ErrorHandlerFunc func(err error) error

	NotificationHandler func(context.Context, *Notification) error
)

// Make EventListener a http.HandlerFunc
func (el EventListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//nCtx, cancel := context.WithCancel(r.Context())
	//defer cancel()

	el(context.Background(), w, r, "", nil, nil)
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
