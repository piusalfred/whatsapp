//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/message/media"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

// Handler registers callbacks for every WhatsApp webhook event type. Each
// field holds an event-specific handler that defaults to a no-op. Use the
// On*/Set* method pairs to read or replace individual handlers.
//
// Registration follows a consistent pattern:
//
//	handler.OnTextMessage(...)   // returns current handler
//	handler.SetTextMessage(...)  // replaces with a new handler
//
// Fields are organized by event category: flow alerts, template updates,
// account notifications, message types (text, image, interactive, etc.),
// status changes, and group events.
type Handler struct {
	flows                   *FlowNotificationHandler
	business                *BusinessNotificationHandler
	buttonMessage           MessageHandler[Button]
	textMessage             MessageHandler[Text]
	orderMessage            MessageHandler[Order]
	locationMessage         MessageHandler[media.Location]
	contactsMessage         MessageHandler[message.Contacts]
	reactionMessage         MessageHandler[media.Reaction]
	productInquiry          MessageHandler[Text]
	interactiveMessage      MessageHandler[Interactive]
	buttonReplyMessage      MessageHandler[ButtonReply]
	listReplyMessage        MessageHandler[ListReply]
	flowCompletionUpdate    NativeFlowCompletionHandler
	addressSubmission       NativeFlowCompletionHandler
	referralMessage         MessageHandler[ReferralNotification]
	customerIDChange        MessageHandler[Identity]
	systemMessage           MessageHandler[System]
	requestWelcome          MessageHandler[Message]
	audioMessage            MediaMessageHandler
	videoMessage            MediaMessageHandler
	imageMessage            MediaMessageHandler
	documentMessage         MediaMessageHandler
	stickerMessage          MediaMessageHandler
	notificationErrors      MessageChangeValueHandler[werrors.Error]
	messageStatusChange     MessageChangeValueHandler[Status]
	revokeMessage           MessageHandler[Revoke]
	editMessage             MessageHandler[Edit]
	smbAppStateSync         MessageChangeValueHandler[SMBAppStateSync]
	smbMessageEcho          MessageHandler[Message]
	userPreferencesUpdate   MessageChangeValueHandler[UserPreference]
	groupLifecycleUpdate    MessageChangeValueHandler[Group]
	groupParticipantsUpdate MessageChangeValueHandler[Group]
	groupSettingsUpdate     MessageChangeValueHandler[Group]
	groupStatusUpdate       MessageChangeValueHandler[Group]
	errorMessage            MessageErrorsHandler
	unsupportedMessage      MessageErrorsHandler
	historySync             MessageChangeValueHandler[HistoryEntry]

	errorHandlerFunc func(ctx context.Context, err error) error

	// unrecognizedField is called for any change.Field not handled by the
	// dispatch switch. When nil (the default), unrecognized fields are
	// silently acknowledged with 200 to prevent WhatsApp retry storms.
	// Set via [Handler.OnUnrecognizedField] to handle future or custom
	// notification types without modifying the library.
	unrecognizedField func(ctx context.Context, notification *Notification, change Change, entry Entry) error
}

// NewHandler creates a Handler with all callbacks initialized to no-ops.
// Register handlers via the Set* methods before attaching to a Listener.
func NewHandler() *Handler {
	return &Handler{
		flows:                   &FlowNotificationHandler{},
		business:                &BusinessNotificationHandler{},
		buttonMessage:           newMessageEventHandler[Button](),
		textMessage:             newMessageEventHandler[Text](),
		orderMessage:            newMessageEventHandler[Order](),
		locationMessage:         newMessageEventHandler[media.Location](),
		contactsMessage:         newMessageEventHandler[message.Contacts](),
		reactionMessage:         newMessageEventHandler[media.Reaction](),
		productInquiry:          newMessageEventHandler[Text](),
		interactiveMessage:      newMessageEventHandler[Interactive](),
		buttonReplyMessage:      newMessageEventHandler[ButtonReply](),
		listReplyMessage:        newMessageEventHandler[ListReply](),
		flowCompletionUpdate:    newMessageEventHandler[NFMReply](),
		addressSubmission:       newMessageEventHandler[NFMReply](),
		referralMessage:         newMessageEventHandler[ReferralNotification](),
		customerIDChange:        newMessageEventHandler[Identity](),
		systemMessage:           newMessageEventHandler[System](),
		audioMessage:            newMessageEventHandler[media.Info](),
		videoMessage:            newMessageEventHandler[media.Info](),
		imageMessage:            newMessageEventHandler[media.Info](),
		documentMessage:         newMessageEventHandler[media.Info](),
		stickerMessage:          newMessageEventHandler[media.Info](),
		notificationErrors:      newMessageValueEventHandler[werrors.Error](),
		messageStatusChange:     newMessageValueEventHandler[Status](),
		revokeMessage:           newMessageEventHandler[Revoke](),
		editMessage:             newMessageEventHandler[Edit](),
		smbAppStateSync:         newMessageValueEventHandler[SMBAppStateSync](),
		smbMessageEcho:          newMessageEventHandler[Message](),
		userPreferencesUpdate:   newMessageValueEventHandler[UserPreference](),
		groupLifecycleUpdate:    newMessageValueEventHandler[Group](),
		groupParticipantsUpdate: newMessageValueEventHandler[Group](),
		groupSettingsUpdate:     newMessageValueEventHandler[Group](),
		groupStatusUpdate:       newMessageValueEventHandler[Group](),
		errorMessage:            NewNoOpMessageErrorsHandler(),
		unsupportedMessage:      NewNoOpMessageErrorsHandler(),
		requestWelcome:          newMessageEventHandler[Message](),
		historySync:             newMessageValueEventHandler[HistoryEntry](),
		errorHandlerFunc: func(_ context.Context, _ error) error {
			return nil
		},
	}
}

func newMessageEventHandler[T any]() MessageHandler[T] {
	return NewNoOpMessageHandler[T]()
}

func newMessageValueEventHandler[T any]() MessageChangeValueHandler[T] {
	return NewNoOpMessageChangeValueHandler[T]()
}

// OnError configures a callback which is invoked whenever an error occurs
// while processing the webhook payload. The callback can decide whether the
// error is "fatal" or "non-fatal":
//
//   - If the callback returns nil, processing continues for any remaining
//     changes and messages in the payload.
//
//   - If the callback returns a non-nil error, processing stops immediately,
//     and the Handler responds with an error status (e.g., 500).
//
// This allows you to handle partial failures gracefully in multi-change payloads
// (e.g., ignoring certain errors or logging them) without terminating the entire
// notification processing.
// If no callback is set, the default behavior is to ignore errors and continue
// processing the payload.
func (handler *Handler) OnError(f func(ctx context.Context, err error) error) {
	handler.errorHandlerFunc = f
}

// OnUnrecognizedField registers a handler for notification fields that the
// library does not yet recognise. By default unrecognized fields are silently
// acknowledged (200) to prevent WhatsApp from retrying. When a handler is set,
// it receives the full notification, change, and entry — use change.Field and
// change.Value to inspect the raw payload.
func (handler *Handler) OnUnrecognizedField(
	f func(ctx context.Context, notification *Notification, change Change, entry Entry) error,
) {
	handler.unrecognizedField = f
}

// HandleNotification processes with a single Notification containing one or more
// Entry objects. Each Entry can have multiple Changes, and each Change can
// contain multiple messages or event objects.
//
// For every error encountered in a Change or Message handler, the error is passed
// to the error handler function (registered via OnError). If that callback
// returns a non-nil error, processing halts immediately and this method returns
// an HTTP 500 response. Otherwise, it continues processing subsequent changes.
//
// If no fatal errors are encountered, it returns an HTTP 200 response, indicating
// all parts of the payload were processed successfully.
func (handler *Handler) HandleNotification(ctx context.Context, notification *Notification) *Response {
	for _, entry := range notification.Entry {
		for _, change := range entry.Changes {
			if err := handler.handleNotificationChange(ctx, notification, change, entry); err != nil {
				return &Response{StatusCode: http.StatusInternalServerError}
			}
		}
	}

	return &Response{StatusCode: http.StatusOK}
}

// handleNotificationChange routes each incoming webhook change to the correct
// handler based on change.Field. Unknown fields are short-circuited via
// [isKnownField] and routed to [Handler.OnUnrecognizedField] (if set) or
// silently acknowledged.
func (handler *Handler) handleNotificationChange(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	// Fields not recognized by the library route to the user-supplied
	// unrecognized-field handler (if set) or are silently acknowledged.
	if !isKnownField(change.Field) {
		if handler.unrecognizedField != nil {
			return handler.unrecognizedField(ctx, notification, change, entry)
		}
		return nil
	}

	switch change.Field {
	case ChangeFieldFlows.String():
		return handler.handleFlowsChange(ctx, notification, change, entry)
	case ChangeFieldAccountAlerts.String(),
		ChangeFieldTemplateStatusUpdate.String(),
		ChangeFieldTemplateCategoryUpdate.String(),
		ChangeFieldTemplateQualityUpdate.String(),
		ChangeFieldTemplateComponentsUpdate.String(),
		ChangeFieldPhoneNumberNameUpdate.String(),
		ChangeFieldPhoneNumberQualityUpdate.String(),
		ChangeFieldAccountUpdate.String(),
		ChangeFieldAccountReviewUpdate.String(),
		ChangeFieldCalls.String(),
		ChangeFieldBusinessCapabilityUpdate.String(),
		ChangeFieldAccountSettingsUpdate.String(),
		ChangeFieldSecurity.String():
		return handler.business.Handle(ctx, notification, change, entry, handler.errorHandlerFunc)
	case ChangeFieldUserPreferences.String():
		return handler.handleUserPreferencesChange(ctx, notification, change, entry)
	case ChangeFieldSMBAppStateSync.String():
		return handler.handleSMBAppStateSyncChange(ctx, notification, change, entry)
	case ChangeFieldMessages.String():
		return handler.handleNotificationMessageItem(ctx, entry, change)
	case ChangeFieldSMBMessageEchoes.String():
		return handler.handleSMBMessageEchoesChange(ctx, notification, change, entry)
	case ChangeFieldGroupLifecycleUpdate.String(),
		ChangeFieldGroupParticipantsUpdate.String(),
		ChangeFieldGroupSettingsUpdate.String(),
		ChangeFieldGroupStatusUpdate.String():
		return handler.handleGroupWebhooks(ctx, change, entry)
	case ChangeFieldHistory.String():
		return handler.handleHistoryChange(ctx, notification, change, entry)
	}

	return nil
}

func (handler *Handler) handleFlowsChange(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	notificationCtx := &FlowNotificationContext{
		NotificationObject: notification.Object,
		EntryID:            entry.ID,
		EntryTime:          entry.Time,
		ChangeField:        change.Field,
		EventName:          change.Value.Event,
		EventMessage:       change.Value.Message,
		FlowID:             change.Value.FlowID,
	}
	if err := handler.flows.Handle(ctx, notificationCtx, change.Value); err != nil {
		if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
			return handlerErr
		}
	}
	return nil
}

func (handler *Handler) handleUserPreferencesChange(
	ctx context.Context,
	_ *Notification,
	change Change,
	entry Entry,
) error {
	return handleMessageChangeNotification(
		ctx, handler, handler.userPreferencesUpdate, change,
		entry, change.Value.UserPreferences,
	)
}

func (handler *Handler) handleSMBAppStateSyncChange(
	ctx context.Context,
	_ *Notification,
	change Change,
	entry Entry,
) error {
	syncs := make([]*SMBAppStateSync, len(change.Value.StateSync))
	for i := range change.Value.StateSync {
		syncs[i] = &change.Value.StateSync[i]
	}
	return handleMessageChangeNotification(
		ctx, handler, handler.smbAppStateSync, change,
		entry, syncs,
	)
}

func (handler *Handler) handleSMBMessageEchoesChange(
	ctx context.Context,
	_ *Notification,
	change Change,
	entry Entry,
) error {
	for _, msg := range change.Value.MessageEchoes {
		if msg == nil {
			continue
		}
		notificationCtx := &MessageNotificationContext{
			EntryID:          entry.ID,
			MessagingProduct: change.Value.MessagingProduct,
			Metadata:         change.Value.Metadata,
			Contacts:         change.Value.Contacts,
		}
		if err := handler.handleNotificationMessage(ctx, notificationCtx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (handler *Handler) handleHistoryChange(
	ctx context.Context,
	_ *Notification,
	change Change,
	entry Entry,
) error {
	entries := make([]*HistoryEntry, len(change.Value.History))
	for i := range change.Value.History {
		entries[i] = &change.Value.History[i]
	}
	if len(entries) > 0 {
		if err := handler.historySync.Handle(ctx,
			&MessageNotificationContext{
				EntryID:          entry.ID,
				MessagingProduct: change.Value.MessagingProduct,
				Metadata:         change.Value.Metadata,
			}, entries); err != nil {
			return fmt.Errorf("history sync: %w", err)
		}
	}
	// Media content for history messages is delivered as a
	// separate webhook with messages in the history field.
	if len(change.Value.Messages) > 0 {
		return handler.handleNotificationMessageItem(ctx, entry, change)
	}
	return nil
}

// ChangeField identifies the type of webhook notification. The string value
// matches the WhatsApp API field name in the change object.
//
// Supported fields from the WhatsApp API:
//   - account_alerts, account_review_update, account_update
//   - business_capability_update, calls
//   - flows (flow status, error rates, latency, availability)
//   - messages (text, image, audio, video, document, sticker, interactive,
//     button, order, location, contacts, reaction, system, referral, request_welcome)
//   - message_template_status_update, template_category_update,
//     message_template_quality_update, message_template_components_update
//   - phone_number_name_update, phone_number_quality_update
//   - user_preferences, account_settings_update, security
//   - group_lifecycle_update, group_participants_update, group_settings_update,
//     group_status_update
//
// Not yet implemented (no-ops if received):
//
//	automatic_events, partner_solutions, payment_configuration_update
type ChangeField string

const (
	ChangeFieldFlows                    ChangeField = "flows"
	ChangeFieldAccountAlerts            ChangeField = "account_alerts"
	ChangeFieldTemplateStatusUpdate     ChangeField = "message_template_status_update"
	ChangeFieldTemplateCategoryUpdate   ChangeField = "template_category_update"
	ChangeFieldTemplateQualityUpdate    ChangeField = "message_template_quality_update"
	ChangeFieldPhoneNumberNameUpdate    ChangeField = "phone_number_name_update"
	ChangeFieldBusinessCapabilityUpdate ChangeField = "business_capability_update"
	ChangeFieldAccountUpdate            ChangeField = "account_update"
	ChangeFieldAccountReviewUpdate      ChangeField = "account_review_update"
	ChangeFieldPhoneNumberQualityUpdate ChangeField = "phone_number_quality_update"
	ChangeFieldMessages                 ChangeField = "messages"
	ChangeFieldUserPreferences          ChangeField = "user_preferences"
	ChangeFieldAccountSettingsUpdate    ChangeField = "account_settings_update"
	ChangeFieldCalls                    ChangeField = "calls"
	ChangeFieldGroupLifecycleUpdate     ChangeField = "group_lifecycle_update"
	ChangeFieldGroupParticipantsUpdate  ChangeField = "group_participants_update"
	ChangeFieldGroupSettingsUpdate      ChangeField = "group_settings_update"
	ChangeFieldGroupStatusUpdate        ChangeField = "group_status_update"
	ChangeFieldHistory                  ChangeField = "history"
	ChangeFieldSecurity                 ChangeField = "security"
	ChangeFieldTemplateComponentsUpdate ChangeField = "message_template_components_update"
	ChangeFieldSMBAppStateSync          ChangeField = "smb_app_state_sync"
	ChangeFieldSMBMessageEchoes         ChangeField = "smb_message_echoes"
)

const (
	EventFlowStatusChange     = "FLOW_STATUS_CHANGE"
	EventEndpointErrorRate    = "ENDPOINT_ERROR_RATE"
	EventEndpointLatency      = "ENDPOINT_LATENCY"
	EventEndpointAvailability = "ENDPOINT_AVAILABILITY"
	EventClientErrorRate      = "CLIENT_ERROR_RATE"
)

func (c ChangeField) String() string {
	return string(c)
}

// KnownChangeFields returns every [ChangeField] that the [Handler] dispatch
// switch recognises. Callers can compare this list against their registered
// handlers at startup to detect missing coverage.
func KnownChangeFields() []ChangeField {
	return []ChangeField{
		ChangeFieldFlows,
		ChangeFieldAccountAlerts,
		ChangeFieldTemplateStatusUpdate,
		ChangeFieldTemplateCategoryUpdate,
		ChangeFieldTemplateQualityUpdate,
		ChangeFieldTemplateComponentsUpdate,
		ChangeFieldPhoneNumberNameUpdate,
		ChangeFieldPhoneNumberQualityUpdate,
		ChangeFieldAccountUpdate,
		ChangeFieldAccountReviewUpdate,
		ChangeFieldCalls,
		ChangeFieldBusinessCapabilityUpdate,
		ChangeFieldUserPreferences,
		ChangeFieldSMBAppStateSync,
		ChangeFieldAccountSettingsUpdate,
		ChangeFieldSecurity,
		ChangeFieldMessages,
		ChangeFieldSMBMessageEchoes,
		ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate,
		ChangeFieldHistory,
	}
}

// isKnownField reports whether field is handled by the dispatch switch.
func isKnownField(field string) bool {
	for _, f := range KnownChangeFields() {
		if f.String() == field {
			return true
		}
	}
	return false
}

type (
	EventHandler[S any, T any] interface {
		HandleEvent(ctx context.Context, ntx *S, notification *T) error
	}

	EventHandlerFunc[S any, T any] func(ctx context.Context, ntx *S, notification *T) error
)

func (f EventHandlerFunc[S, T]) HandleEvent(ctx context.Context, ntx *S, notification *T) error {
	return f(ctx, ntx, notification)
}

func NewNoOpEventHandler[S, T any]() EventHandler[S, T] {
	return EventHandlerFunc[S, T](func(_ context.Context, _ *S, _ *T) error {
		return nil
	})
}

func (c *Contact) SenderInfo() *SenderInfo {
	return &SenderInfo{
		Name: c.Profile.Name,
		WaID: c.WaID,
	}
}

type (
	Contact struct {
		Profile         *Profile `json:"profile,omitempty"`
		WaID            string   `json:"wa_id,omitempty"`
		IdentityKeyHash string   `json:"identity_key_hash,omitempty"`
	}

	SenderInfo struct {
		Name string `json:"name,omitempty"`
		WaID string `json:"wa_id,omitempty"`
	}

	// Message contains the information of a message. It is embedded in the Value object.
	Message struct {
		Audio       *media.Info       `json:"audio,omitempty"`
		Button      *Button           `json:"button,omitempty"`
		Context     *Context          `json:"context,omitempty"`
		Document    *media.Info       `json:"document,omitempty"`
		Errors      []*werrors.Error  `json:"errors,omitempty"`
		From        string            `json:"from,omitempty"`
		ID          string            `json:"id,omitempty"`
		GroupID     string            `json:"group_id,omitempty"`
		Identity    *Identity         `json:"identity,omitempty"`
		Image       *media.Info       `json:"image,omitempty"`
		Interactive *Interactive      `json:"interactive,omitempty"`
		Order       *Order            `json:"order,omitempty"`
		Referral    *Referral         `json:"referral,omitempty"`
		Sticker     *media.Info       `json:"sticker,omitempty"`
		System      *System           `json:"system,omitempty"`
		Text        *Text             `json:"text,omitempty"`
		Timestamp   string            `json:"timestamp,omitempty"`
		Type        string            `json:"type,omitempty"`
		Video       *media.Info       `json:"video,omitempty"`
		Contacts    *message.Contacts `json:"contacts,omitempty"`
		Location    *media.Location   `json:"location,omitempty"`
		Reaction    *media.Reaction   `json:"reaction,omitempty"`
		Revoke      *Revoke           `json:"revoke,omitempty"`
		Edit        *Edit             `json:"edit,omitempty"`
	}

	// Revoke payload when a WhatsApp user deletes a previously sent message.
	// Triggers: user deletes a message within ~2 days of sending.
	// The original_message_id identifies the message that was revoked.
	Revoke struct {
		OriginalMessageID string `json:"original_message_id"`
	}

	// Edit payload when a WhatsApp user edits a previously sent text or media
	// caption message. The original_message_id identifies the edited message;
	// Message contains the full replacement content.
	// Triggers: user edits a message within 15 minutes of sending.
	// Note: edit messages are currently unsupported by WhatsApp and may arrive
	// as unsupported message type instead.
	Edit struct {
		OriginalMessageID string   `json:"original_message_id"`
		Message           *Message `json:"message"`
	}

	Metadata struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		PhoneNumberID      string `json:"phone_number_id,omitempty"`
	}

	ReferralNotification struct {
		Text     *Text
		Referral *Referral
	}

	// Pricing provides billing information for a message status event.
	// Present only on sent status and one of delivered or read.
	// Category values: authentication, authentication-international, marketing,
	// marketing_lite, referral_conversion, service, utility.
	// PricingModel is "CBP" (conversation-based) or "PMP" (per-message).
	Pricing struct {
		Billable     bool   `json:"billable,omitempty"` // Deprecated
		Category     string `json:"category,omitempty"`
		PricingModel string `json:"pricing_model,omitempty"`
	}

	// ConversationOrigin identifies how a conversation was started (e.g.,
	// "authentication", "marketing", "service").
	ConversationOrigin struct {
		Type string `json:"type,omitempty"`
	}

	// Conversation holds metadata about the conversation associated with
	// a message. Omitted for v24.0+ unless the message was sent within an
	// open free entry point window. The ID is unique per window.
	Conversation struct {
		ID     string              `json:"id,omitempty"`
		Origin *ConversationOrigin `json:"origin,omitempty"`
		Expiry string              `json:"expiration_timestamp,omitempty"`
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
		Name         string          `json:"name"`
		Body         string          `json:"body"`
		ResponseJSON json.RawMessage `json:"response_json"`
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
		Forwarded           bool             `json:"forwarded,omitempty"`
		FrequentlyForwarded bool             `json:"frequently_forwarded,omitempty"`
		From                string           `json:"from,omitempty"`
		ID                  string           `json:"id,omitempty"`
		ReferredProduct     *ReferredProduct `json:"referred_product,omitempty"`
		Type                string           `json:"type,omitempty"`
	}

	ReferredProduct struct {
		CatalogID         string `json:"catalog_id,omitempty"`
		ProductRetailerID string `json:"product_retailer_id,omitempty"`
	}

	// MessageNotificationContext is the context of a notification contains information about the
	// notification and the business that is subscribed to the Webhooks.
	// these are common fields to all notifications.
	// EntryID - The WhatsApp Business Account EntryID for the business that is subscribed to the webhook.
	// Contacts - Array of contact objects with information for the customer who sent a message
	// to the business
	// Metadata - A metadata object describing the business subscribed to the webhook.
	MessageNotificationContext struct {
		EntryID          string
		MessagingProduct string
		Contacts         []*Contact
		Metadata         *Metadata
	}

	// MessageInfo is the context of a message contains information about the
	// message and the business that is subscribed to the Webhooks.
	// these are common fields to all types of messages.
	// From The customer's phone number who sent the message to the business.
	// MessageID The MessageID for the message that was received by the business. You could use messages
	// endpoint to mark this specific message as read.
	// Timestamp The timestamp for when the message was received by the business.
	// Type The type of message that was received by the business.
	// Context The context of the message. Only included when a user replies or interacts with one
	// of your messages.
	// GroupID is set when the message was sent to or from a group chat.
	MessageInfo struct {
		From             string
		MessageID        string
		Timestamp        string
		Type             string
		GroupID          string
		Context          *Context
		IsAReply         bool
		IsForwarded      bool
		IsProductInquiry bool
		IsReferral       bool
	}
)

func (message *Message) IsAReply() bool {
	return message.Context != nil && message.Context.ReferredProduct == nil && !message.Context.Forwarded
}

// IsForwarded checks if the message is a forwarded message. It returns true if the message has a non-nil Context and the Forwarded field in the Context is true.
func (message *Message) IsForwarded() bool {
	return message.Context != nil && message.Context.Forwarded
}

// IsProductInquiry checks if the message is a product inquiry message.
func (message *Message) IsProductInquiry() bool {
	return message.Context != nil && message.Context.ReferredProduct != nil
}

func (message *Message) IsReferral() bool {
	return message.Referral != nil
}

// AllSendersInfo returns the sender information of all contacts in the notification context.
func (ctx *MessageNotificationContext) AllSendersInfo() []*SenderInfo {
	senders := make([]*SenderInfo, len(ctx.Contacts))
	for i, contact := range ctx.Contacts {
		senders[i] = contact.SenderInfo()
	}
	return senders
}

// SenderInfo returns the sender information of the first contact in the notification context.
func (ctx *MessageNotificationContext) SenderInfo() *SenderInfo {
	if len(ctx.Contacts) == 0 {
		return nil
	}
	return ctx.Contacts[0].SenderInfo()
}
