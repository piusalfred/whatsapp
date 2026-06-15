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
	flowStatus               EventHandler[FlowNotificationContext, StatusChangeDetails]
	flowClientErrorRate      EventHandler[FlowNotificationContext, ClientErrorRateDetails]
	flowEndpointErrorRate    EventHandler[FlowNotificationContext, EndpointErrorRateDetails]
	flowEndpointLatency      EventHandler[FlowNotificationContext, EndpointLatencyDetails]
	flowEndpointAvailability EventHandler[FlowNotificationContext, EndpointAvailabilityDetails]
	alerts                   EventHandler[BusinessNotificationContext, AlertNotification]
	templateStatus           EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification]
	templateCategory         EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification]
	templateQuality          EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification]
	templateComponents       EventHandler[BusinessNotificationContext, TemplateComponentsUpdateNotification]
	phoneNumberNameUpdate    EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate]
	capabilityUpdate         EventHandler[BusinessNotificationContext, CapabilityUpdate]
	accountUpdate            EventHandler[BusinessNotificationContext, AccountUpdate]
	phoneSettingsUpdate      EventHandler[BusinessNotificationContext, PhoneNumberSettings]
	phoneNumberQualityUpdate EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate]
	accountReviewUpdate      EventHandler[BusinessNotificationContext, AccountReviewUpdate]
	callStatusUpdate         EventHandler[BusinessNotificationContext, CallStatusUpdate]
	securityUpdate           EventHandler[BusinessNotificationContext, SecurityNotification]
	buttonMessage            MessageHandler[Button]
	textMessage              MessageHandler[Text]
	orderMessage             MessageHandler[Order]
	locationMessage          MessageHandler[media.Location]
	contactsMessage          MessageHandler[message.Contacts]
	reactionMessage          MessageHandler[media.Reaction]
	productInquiry           MessageHandler[Text]
	interactiveMessage       MessageHandler[Interactive]
	buttonReplyMessage       MessageHandler[ButtonReply]
	listReplyMessage         MessageHandler[ListReply]
	flowCompletionUpdate     NativeFlowCompletionHandler
	addressSubmission        NativeFlowCompletionHandler
	referralMessage          MessageHandler[ReferralNotification]
	customerIDChange         MessageHandler[Identity]
	systemMessage            MessageHandler[System]
	requestWelcome           MessageHandler[Message]
	audioMessage             MediaMessageHandler
	videoMessage             MediaMessageHandler
	imageMessage             MediaMessageHandler
	documentMessage          MediaMessageHandler
	stickerMessage           MediaMessageHandler
	notificationErrors       MessageChangeValueHandler[werrors.Error]
	messageStatusChange      MessageChangeValueHandler[Status]
	revokeMessage            MessageHandler[Revoke]
	editMessage              MessageHandler[Edit]
	smbAppStateSync          MessageChangeValueHandler[SMBAppStateSync]
	smbMessageEcho           MessageHandler[Message]
	userPreferencesUpdate    MessageChangeValueHandler[UserPreference]
	groupLifecycleUpdate     MessageChangeValueHandler[Group]
	groupParticipantsUpdate  MessageChangeValueHandler[Group]
	groupSettingsUpdate      MessageChangeValueHandler[Group]
	groupStatusUpdate        MessageChangeValueHandler[Group]
	errorMessage             MessageErrorsHandler
	unsupportedMessage       MessageErrorsHandler
	historySync              MessageChangeValueHandler[HistoryEntry]

	errorHandlerFunc func(ctx context.Context, err error) error
}

// NewHandler creates a Handler with all callbacks initialized to no-ops.
// Register handlers via the Set* methods before attaching to a Listener.
func NewHandler() *Handler {
	return &Handler{
		flowStatus:               NewNoOpEventHandler[FlowNotificationContext, StatusChangeDetails](),
		flowClientErrorRate:      NewNoOpEventHandler[FlowNotificationContext, ClientErrorRateDetails](),
		flowEndpointErrorRate:    NewNoOpEventHandler[FlowNotificationContext, EndpointErrorRateDetails](),
		flowEndpointLatency:      NewNoOpEventHandler[FlowNotificationContext, EndpointLatencyDetails](),
		flowEndpointAvailability: NewNoOpEventHandler[FlowNotificationContext, EndpointAvailabilityDetails](),
		alerts:                   NewNoOpEventHandler[BusinessNotificationContext, AlertNotification](),
		templateStatus:           NewNoOpEventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification](),
		templateCategory:         NewNoOpEventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification](),
		templateQuality:          NewNoOpEventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification](),
		templateComponents:       NewNoOpEventHandler[BusinessNotificationContext, TemplateComponentsUpdateNotification](),
		phoneNumberNameUpdate:    NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberNameUpdate](),
		capabilityUpdate:         NewNoOpEventHandler[BusinessNotificationContext, CapabilityUpdate](),
		accountUpdate:            NewNoOpEventHandler[BusinessNotificationContext, AccountUpdate](),
		phoneSettingsUpdate:      NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberSettings](),
		phoneNumberQualityUpdate: NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate](),
		accountReviewUpdate:      NewNoOpEventHandler[BusinessNotificationContext, AccountReviewUpdate](),
		callStatusUpdate:         NewNoOpEventHandler[BusinessNotificationContext, CallStatusUpdate](),
		securityUpdate:           NewNoOpEventHandler[BusinessNotificationContext, SecurityNotification](),
		buttonMessage:            NewNoOpMessageHandler[Button](),
		textMessage:              NewNoOpMessageHandler[Text](),
		orderMessage:             NewNoOpMessageHandler[Order](),
		locationMessage:          NewNoOpMessageHandler[media.Location](),
		contactsMessage:          NewNoOpMessageHandler[message.Contacts](),
		reactionMessage:          NewNoOpMessageHandler[media.Reaction](),
		productInquiry:           NewNoOpMessageHandler[Text](),
		interactiveMessage:       NewNoOpMessageHandler[Interactive](),
		buttonReplyMessage:       NewNoOpMessageHandler[ButtonReply](),
		listReplyMessage:         NewNoOpMessageHandler[ListReply](),
		flowCompletionUpdate:     NewNoOpMessageHandler[NFMReply](),
		addressSubmission:        NewNoOpMessageHandler[NFMReply](),
		referralMessage:          NewNoOpMessageHandler[ReferralNotification](),
		customerIDChange:         NewNoOpMessageHandler[Identity](),
		systemMessage:            NewNoOpMessageHandler[System](),
		audioMessage:             NewNoOpMessageHandler[media.Info](),
		videoMessage:             NewNoOpMessageHandler[media.Info](),
		imageMessage:             NewNoOpMessageHandler[media.Info](),
		documentMessage:          NewNoOpMessageHandler[media.Info](),
		stickerMessage:           NewNoOpMessageHandler[media.Info](),
		notificationErrors:       NewNoOpMessageChangeValueHandler[werrors.Error](),
		messageStatusChange:      NewNoOpMessageChangeValueHandler[Status](),
		revokeMessage:            NewNoOpMessageHandler[Revoke](),
		editMessage:              NewNoOpMessageHandler[Edit](),
		smbAppStateSync:          NewNoOpMessageChangeValueHandler[SMBAppStateSync](),
		smbMessageEcho:           NewNoOpMessageHandler[Message](),
		userPreferencesUpdate:    NewNoOpMessageChangeValueHandler[UserPreference](),
		groupLifecycleUpdate:     NewNoOpMessageChangeValueHandler[Group](),
		groupParticipantsUpdate:  NewNoOpMessageChangeValueHandler[Group](),
		groupSettingsUpdate:      NewNoOpMessageChangeValueHandler[Group](),
		groupStatusUpdate:        NewNoOpMessageChangeValueHandler[Group](),
		errorMessage:             NewNoOpMessageErrorsHandler(),
		unsupportedMessage:       NewNoOpMessageErrorsHandler(),
		requestWelcome:           NewNoOpMessageHandler[Message](),
		historySync:              NewNoOpMessageChangeValueHandler[HistoryEntry](),
		errorHandlerFunc: func(_ context.Context, _ error) error {
			return nil
		},
	}
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

func (handler *Handler) handleNotificationChange( //nolint:funlen,gocognit,gocyclo,cyclop // complex notification routing
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	switch change.Field {
	case ChangeFieldFlows.String():
		notificationCtx := &FlowNotificationContext{
			NotificationObject: notification.Object,
			EntryID:            entry.ID,
			EntryTime:          entry.Time,
			ChangeField:        change.Field,
			EventName:          change.Value.Event,
			EventMessage:       change.Value.Message,
			FlowID:             change.Value.FlowID,
		}
		if err := handler.handleFlowNotification(ctx, notificationCtx, change.Value); err != nil {
			if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
				return handlerErr
			}
		}

	case ChangeFieldAccountAlerts.String():
		if err := handler.handleAccountAlerts(ctx, notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateStatusUpdate.String():
		if err := handler.handleTemplateStatusUpdate(ctx,
			notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateCategoryUpdate.String():
		if err := handler.handleTemplateCategoryUpdate(ctx,
			notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateQualityUpdate.String():
		if err := handler.handleTemplateQualityUpdate(
			ctx, notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateComponentsUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.templateComponents, &TemplateComponentsUpdateNotification{
				MessageTemplateID:       change.Value.MessageTemplateID,
				MessageTemplateName:     change.Value.MessageTemplateName,
				MessageTemplateLanguage: change.Value.MessageTemplateLanguage,
				Title:                   change.Value.MessageTemplateTitle,
				Element:                 change.Value.MessageTemplateElement,
				Footer:                  change.Value.MessageTemplateFooter,
				Buttons:                 change.Value.MessageTemplateButtons,
			},
		)

	case ChangeFieldPhoneNumberNameUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.phoneNumberNameUpdate, change.Value.PhoneNumberNameUpdate(),
		)

	case ChangeFieldPhoneNumberQualityUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification,
			change, entry, handler.phoneNumberQualityUpdate,
			change.Value.PhoneNumberQualityUpdate(),
		)

	case ChangeFieldAccountUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.accountUpdate, change.Value.AccountUpdate(),
		)

	case ChangeFieldAccountReviewUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.accountReviewUpdate, change.Value.AccountReviewUpdate(),
		)

	case ChangeFieldCalls.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.callStatusUpdate, change.Value.CallStatusUpdate(),
		)

	case ChangeFieldBusinessCapabilityUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.capabilityUpdate, change.Value.CapabilityUpdate(),
		)

	case ChangeFieldUserPreferences.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.userPreferencesUpdate, change,
			entry, change.Value.UserPreferences,
		)

	case ChangeFieldSMBAppStateSync.String():
		syncs := make([]*SMBAppStateSync, len(change.Value.StateSync))
		for i := range change.Value.StateSync {
			syncs[i] = &change.Value.StateSync[i]
		}
		return handleMessageChangeNotification(
			ctx, handler, handler.smbAppStateSync, change,
			entry, syncs,
		)

	case ChangeFieldAccountSettingsUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.phoneSettingsUpdate, change.Value.PhoneNumberSettings,
		)

	case ChangeFieldSecurity.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.securityUpdate, &SecurityNotification{
				Event:              change.Value.Event,
				DisplayPhoneNumber: change.Value.DisplayPhoneNumber,
				Requester:          change.Value.Requester,
			},
		)

	case ChangeFieldMessages.String():
		return handler.handleNotificationMessageItem(ctx, entry, change)

	case ChangeFieldSMBMessageEchoes.String():
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

	case ChangeFieldGroupLifecycleUpdate.String(),
		ChangeFieldGroupParticipantsUpdate.String(),
		ChangeFieldGroupSettingsUpdate.String(),
		ChangeFieldGroupStatusUpdate.String():
		return handler.handleGroupWebhooks(ctx, change, entry)

	case ChangeFieldHistory.String():
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

	// Unrecognized fields are silently dropped. This includes official
	// WhatsApp fields not yet implemented (see ChangeField docs). Dropping
	// is intentional — we return nil so WhatsApp receives a 200 and does
	// not retry.
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
