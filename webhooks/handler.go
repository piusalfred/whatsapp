//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Handler is the central webhook dispatch unit. It registers callbacks for
// every WhatsApp webhook event type and routes incoming notifications to the
// correct handler based on the change field. Composite handlers (Flows,
// Business, Messages, Groups) own their own typed dispatch logic.

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
// On* methods to register handlers.
//
// Registration follows a consistent pattern:
//
//	handler.OnTextMessage(h)   // registers the handler
//
// # Concurrency
//
// Handler is not safe for concurrent modification. Register all handlers
// before calling [Handler.HandleNotification] for the first time. Once the
// server starts, only HandleNotification may be called concurrently — it is
// safe for concurrent reads of the registered callbacks.
//
// The same constraint applies to all sub-handlers reachable through the
// Handler: [MessagesHandler], [FlowNotificationHandler],
// [BusinessNotificationHandler], [GroupManagementHandler], and
// [HistoryHandler]. Register their callbacks during initialisation and treat
// them as immutable afterward.
type Handler struct {
	flows    *FlowNotificationHandler
	business *BusinessNotificationHandler
	messages *MessagesHandler
	groups   *GroupManagementHandler

	notificationErrors    ChangeValueHandler[werrors.Error]
	messageStatusChange   ChangeValueHandler[Status]
	smbAppStateSync       ChangeValueHandler[SMBAppStateSync]
	userPreferencesUpdate ChangeValueHandler[UserPreference]
	historySync           ChangeValueHandler[HistoryEntry]

	errorHandler ErrorHandler

	fallback            FallbackHandler
	changeFieldHandlers changeFieldMap
}

// NewHandler creates a Handler with all callbacks initialized to no-ops.
// Register handlers via the Set* methods before attaching to a Listener.
func NewHandler() *Handler {
	implemented := []ChangeField{
		ChangeFieldAccountAlerts,
		ChangeFieldTemplateStatusUpdate,
		ChangeFieldTemplateCategoryUpdate,
		ChangeFieldTemplateQualityUpdate,
		ChangeFieldTemplateComponentsUpdate,
		ChangeFieldPhoneNumberNameUpdate,
		ChangeFieldPhoneNumberQualityUpdate,
		ChangeFieldAccountReviewUpdate,
		ChangeFieldAccountUpdate,
		ChangeFieldBusinessCapabilityUpdate,
		ChangeFieldAccountSettingsUpdate,
		ChangeFieldCalls,
		ChangeFieldSecurity,
		ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate,
		ChangeFieldHistory,
		ChangeFieldMessages,
		ChangeFieldTemplateComponentsUpdate,
		ChangeFieldTemplateCategoryUpdate,
		ChangeFieldTemplateQualityUpdate,
		ChangeFieldSMBAppStateSync,
		ChangeFieldUserPreferences,
		ChangeFieldSMBMessageEchoes,
	}

	passThroughOnError := ErrorHandlerFunc(
		func(_ context.Context, err error) error { return err },
	)

	fallback := FallbackHandlerFunc(
		func(_ context.Context, _ NotificationEntry, _ Change) error {
			return nil
		},
	)

	h := &Handler{
		flows:                 nil,
		business:              nil,
		messages:              nil,
		groups:                nil,
		notificationErrors:    nil,
		messageStatusChange:   nil,
		smbAppStateSync:       nil,
		userPreferencesUpdate: nil,
		historySync:           nil,
		errorHandler:          passThroughOnError,
		changeFieldHandlers:   initChangeFieldMap(implemented...),
		fallback:              fallback,
	}

	return h
}

// OnError configures a callback which is invoked whenever an error occurs
// while processing the webhook payload. The callback can decide whether the
// error is "fatal" or "non-fatal":
//
//   - If the callback returns nil, processing continues for any remaining
//     changes and messages in the payload.
//
//   - If the callback returns a non-nil error, processing stops immediately,
//     and an HTTP 500 is returned to WhatsApp (which may trigger a retry).
//
// Example:
//
//	handler.OnError(webhooks.ErrorHandlerFunc(func(ctx context.Context, err error) error {
//	    log.Printf("webhook error: %v", err)
//	    return nil // continue processing remaining messages
//	}))
func (handler *Handler) OnError(h ErrorHandler) {
	handler.errorHandler = h
}

// SetGeneralFallbackHandler sets the general fallback handler for change.Field
// values not in the known set. It also propagates to all non-nil sub-handlers
// (Flows, Business, Groups, Messages) that don't already have a dedicated
// fallback set.
func (handler *Handler) SetGeneralFallbackHandler(h FallbackHandler) {
	handler.fallback = h

	if handler.business != nil && handler.business.Fallback == nil {
		handler.business.Fallback = h
	}
	if handler.groups != nil && handler.groups.Fallback == nil {
		handler.groups.Fallback = h
	}
}

// HandleNotification processes an incoming WhatsApp webhook notification.
// It iterates over every entry and change in the payload, dispatching each
// to the correct handler based on change.Field. Returns a Response indicating
// success (200), gateway timeout (504), or internal server error (500).
func (handler *Handler) HandleNotification(ctx context.Context, notification *Notification) *Response {
	for _, entry := range notification.Entry {
		select {
		case <-ctx.Done():
			return &Response{StatusCode: http.StatusGatewayTimeout}
		default:
		}
		for _, change := range entry.Changes {
			select {
			case <-ctx.Done():
				return &Response{StatusCode: http.StatusGatewayTimeout}
			default:
			}
			if err := handler.handleNotificationChange(ctx, notification, change, entry); err != nil {
				return &Response{StatusCode: http.StatusInternalServerError}
			}
		}
	}

	return &Response{StatusCode: http.StatusOK}
}

// handleNotificationChange routes each incoming webhook change to the correct
// handler based on change.Field. Unknown fields are short-circuited via
// [isKnownField] and routed to the general fallback (if set) or silently
// acknowledged.
func (handler *Handler) handleNotificationChange(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	if change.Value == nil {
		return nil
	}

	ne := NotificationEntry{
		Object:       notification.Object,
		ID:           entry.ID,
		Time:         entry.Time,
		EntryCount:   len(notification.Entry),
		ChangesCount: len(entry.Changes),
	}

	_, isImplemented := handler.changeFieldHandlers.Check(change.Field)
	if !isImplemented {
		if handler.fallback != nil {
			return fmt.Errorf("general fallback: %w", handler.fallback.Handle(ctx, ne, change))
		}
		return nil
	}

	cfc := GetChangeFieldCategory(change.Field)

	switch cfc {
	case ChangeFieldCategoryFlows:
		return handler.handleFlowsChange(ctx, ne, change)
	case ChangeFieldCategoryBusiness:
		return handler.business.Handle(ctx, ne, change)
	case ChangeFieldCategoryUserPreferences:
		return handler.handleUserPreferencesChange(ctx, ne, change)
	case ChangeFieldCategorySMBAppStateSync:
		return handler.handleSMBAppStateSyncChange(ctx, ne, change)
	case ChangeFieldCategoryMessages:
		return handler.handleNotificationMessageItem(ctx, ne, change)
	case ChangeFieldCategorySMBMessageEchoes:
		return handler.handleSMBMessageEchoesChange(ctx, ne, change)
	case ChangeFieldCategoryGroups:
		return handler.groups.Handle(ctx, ne, change)
	case ChangeFieldCategoryHistory:
		return handler.handleHistoryChange(ctx, ne, change)

	default:
		if handler.fallback != nil {
			return fmt.Errorf("fallback: %w", handler.fallback.Handle(ctx, ne, change))
		}
		return nil
	}
}

func (handler *Handler) handleFlowsChange(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	return handler.flows.Handle(ctx, ne, change)
}

func (handler *Handler) handleUserPreferencesChange(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	return handleMessageChangeNotification(
		ctx, handler, handler.userPreferencesUpdate, ne, change,
		change.Value.UserPreferences,
	)
}

func (handler *Handler) handleSMBAppStateSyncChange(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	syncs := make([]*SMBAppStateSync, len(change.Value.StateSync))
	for i := range change.Value.StateSync {
		syncs[i] = &change.Value.StateSync[i]
	}
	return handleMessageChangeNotification(
		ctx, handler, handler.smbAppStateSync, ne, change,
		syncs,
	)
}

func (handler *Handler) handleSMBMessageEchoesChange(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	notificationCtx := &MessageNotificationContext{
		EntryID:            ne.ID,
		EntryTime:          ne.Time,
		NotificationObject: ne.Object,
		MessagingProduct:   change.Value.MessagingProduct,
		Metadata:           change.Value.Metadata,
		Contacts:           change.Value.Contacts,
	}
	for _, msg := range change.Value.MessageEchoes {
		if msg == nil {
			continue
		}
		if err := handler.messages.Handle(ctx, notificationCtx, msg); err != nil {
			if handler.errorHandler != nil {
				if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}
	return nil
}

func (handler *Handler) handleHistoryChange(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	entries := make([]*HistoryEntry, len(change.Value.History))
	for i := range change.Value.History {
		entries[i] = &change.Value.History[i]
	}
	if len(entries) > 0 {
		req := &ChangeValueRequest[HistoryEntry]{
			Notification: &MessageNotificationContext{
				EntryID:            ne.ID,
				EntryTime:          ne.Time,
				NotificationObject: ne.Object,
				MessagingProduct:   change.Value.MessagingProduct,
				Metadata:           change.Value.Metadata,
			},
			Payload: entries,
		}
		if err := handler.historySync.Handle(ctx, req); err != nil {
			if handler.errorHandler != nil {
				if handlerErr := handler.errorHandler.Handle(ctx, err); handlerErr != nil {
					return fmt.Errorf("error handler: %w", handlerErr)
				}
			}
		}
	}
	// Media content for history messages is delivered as a
	// separate webhook with messages in the history field.
	if len(change.Value.Messages) > 0 {
		return handler.handleNotificationMessageItem(ctx, ne, change)
	}
	return nil
}

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

// KnownChangeFields returns all ChangeField values recognized by the
// library. Unknown fields are routed to the fallback handler (if set)
// or silently acknowledged.
//

func KnownChangeFields() []ChangeField {
	return []ChangeField{
		ChangeFieldAccountAlerts,
		ChangeFieldTemplateStatusUpdate,
		ChangeFieldTemplateCategoryUpdate,
		ChangeFieldTemplateQualityUpdate,
		ChangeFieldTemplateComponentsUpdate,
		ChangeFieldPhoneNumberNameUpdate,
		ChangeFieldPhoneNumberQualityUpdate,
		ChangeFieldAccountReviewUpdate,
		ChangeFieldAccountUpdate,
		ChangeFieldBusinessCapabilityUpdate,
		ChangeFieldAccountSettingsUpdate,
		ChangeFieldCalls,
		ChangeFieldSecurity,
		ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate,
		ChangeFieldHistory,
		ChangeFieldMessages,
		ChangeFieldFlows,
		ChangeFieldSMBAppStateSync,
		ChangeFieldUserPreferences,
		ChangeFieldSMBMessageEchoes,
	}
}

type (
	// FlowEventHandler is the interface for handling typed flow webhook events.
	// It receives a [FlowRequest] carrying the notification context and the typed
	// event payload.
	FlowEventHandler[T any] interface {
		Handle(ctx context.Context, req *FlowRequest[T]) error
	}

	// FlowEventHandlerFunc is an adapter that allows a plain function with the
	// (ctx, *FlowRequest[T]) signature to be used as a [FlowEventHandler].
	FlowEventHandlerFunc[T any] func(ctx context.Context, req *FlowRequest[T]) error
)

func (f FlowEventHandlerFunc[T]) Handle(ctx context.Context, req *FlowRequest[T]) error {
	return f(ctx, req)
}

func NewNoOpFlowEventHandler[T any]() FlowEventHandler[T] {
	return FlowEventHandlerFunc[T](func(_ context.Context, _ *FlowRequest[T]) error {
		return nil
	})
}

type (
	// EventHandler is the generic interface for handling typed webhook events.
	// CtxT is the context type carried in the BusinessRequest (e.g.,
	// BusinessNotificationContext), and T is the typed payload.
	EventHandler[CtxT any, T any] interface {
		Handle(ctx context.Context, req *BusinessRequest[T]) error
	}

	// EventHandlerFunc adapts a bare function to the EventHandler interface.
	EventHandlerFunc[CtxT any, T any] func(ctx context.Context, req *BusinessRequest[T]) error
)

func (f EventHandlerFunc[CtxT, T]) Handle(ctx context.Context, req *BusinessRequest[T]) error {
	return f(ctx, req)
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
		Name string
		WaID string
	}
)

type (
	Message struct {
		Audio          *media.Info       `json:"audio,omitempty"`
		Button         *Button           `json:"button,omitempty"`
		Context        *Context          `json:"context,omitempty"`
		Document       *media.Info       `json:"document,omitempty"`
		Errors         []*werrors.Error  `json:"errors,omitempty"`
		From           string            `json:"from,omitempty"`
		ID             string            `json:"id,omitempty"`
		GroupID        string            `json:"group_id,omitempty"`
		Identity       *Identity         `json:"identity,omitempty"`
		Image          *media.Info       `json:"image,omitempty"`
		Interactive    *Interactive      `json:"interactive,omitempty"`
		Order          *Order            `json:"order,omitempty"`
		Referral       *Referral         `json:"referral,omitempty"`
		Sticker        *media.Info       `json:"sticker,omitempty"`
		System         *System           `json:"system,omitempty"`
		Text           *Text             `json:"text,omitempty"`
		Timestamp      string            `json:"timestamp,omitempty"`
		To             string            `json:"to,omitempty"`
		Type           string            `json:"type,omitempty"`
		Video          *media.Info       `json:"video,omitempty"`
		Contacts       *message.Contacts `json:"contacts,omitempty"`
		Location       *media.Location   `json:"location,omitempty"`
		Reaction       *media.Reaction   `json:"reaction,omitempty"`
		Revoke         *Revoke           `json:"revoke,omitempty"`
		Edit           *Edit             `json:"edit,omitempty"`
		HistoryContext *HistoryContext   `json:"history_context,omitempty"`
	}

	Unsupported struct {
		Type string `json:"type,omitempty"`
	}

	// Revoke represents a message revocation sent by a WhatsApp user
	// (deletion within ~2 days of sending).
	Revoke struct {
		OriginalMessageID string `json:"original_message_id"`
	}

	// Edit represents a message edit sent by a WhatsApp user
	// (within 15 minutes of sending, text or media caption).
	Edit struct {
		OriginalMessageID string `json:"original_message_id"`
		Message           *Edit  `json:"message,omitempty"`
	}

	Metadata struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		PhoneNumberID      string `json:"phone_number_id,omitempty"`
	}

	ReferralNotification struct {
		Text     *Text     `json:"text,omitempty"`
		Referral *Referral `json:"referral,omitempty"`
	}

	Pricing struct {
		Billable     bool   `json:"billable,omitempty"`
		Category     string `json:"category,omitempty"`
		PricingModel string `json:"pricing_model,omitempty"`
	}

	ConversationOrigin struct {
		Type string `json:"type,omitempty"`
	}

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
		Type        InteractiveType `json:"type,omitempty"`
		ButtonReply *ButtonReply    `json:"button_reply,omitempty"`
		ListReply   *ListReply      `json:"list_reply,omitempty"`
		NFMReply    *NFMReply       `json:"nfm_reply,omitempty"`
	}

	NFMReply struct {
		Name         string          `json:"name,omitempty"`
		Body         string          `json:"body,omitempty"`
		ResponseJSON json.RawMessage `json:"response_json,omitempty"`
	}

	InteractiveType string

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
		Quantity          int    `json:"quantity,omitempty"`
		ItemPrice         int    `json:"item_price,omitempty"`
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

	// MessageNotificationContext carries the notification-level metadata from
	// a WhatsApp webhook entry. It identifies the WABA that received the event,
	// the messaging product, and the sender contacts.
	MessageNotificationContext struct {
		NotificationObject string
		EntryID            string
		EntryTime          int64
		MessagingProduct   string
		Contacts           []*Contact
		Metadata           *Metadata
	}

	// NotificationEntry carries the minimal notification envelope fields
	// needed by sub-handlers. Pass this instead of the full *Notification +
	// Entry to avoid leaking the entire payload tree.
	NotificationEntry struct {
		Object       string
		ID           string
		Time         int64
		EntryCount   int
		ChangesCount int
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

// MessageRequest carries all context for a single message webhook event.
// It groups notification-level metadata, message-level metadata extracted
// by the library, and the typed message payload into one struct.
//
// Fields are populated from the incoming JSON webhook before the callback
// is invoked; nil pointers mean the field was absent in the payload.
type MessageRequest[T any] struct {
	// Notification identifies the WhatsApp Business Account that received
	// the message and the messaging product (always "whatsapp"). Contacts
	// carries the sender profile(s).
	Notification *MessageNotificationContext
	// Info is the library-extracted message envelope: sender phone number,
	// message ID (use with the messages endpoint to mark as read),
	// timestamp, type string, and computed flags. IsAReply is true when
	// the message has a context ID and is neither forwarded nor a product
	// inquiry. GroupID is set for group messages; empty otherwise.
	Info *MessageInfo
	// Payload is the typed message body. *Text with a Body field for text
	// messages; *media.Info for media; *Interactive for interactive replies.
	Payload *T
}

// FlowRequest carries context and payload for a flows webhook event.
// Flows are structured data-collection forms; webhooks notify of status
// changes (draft→published) and endpoint performance alerts.
type FlowRequest[T any] struct {
	// Context identifies the source WABA, the flow ID, the event name
	// (e.g. FLOW_STATUS_CHANGE), and a human-readable message.
	Context *FlowNotificationContext
	// Payload is the typed event details: *StatusChangeDetails for status
	// transitions, *ClientErrorRateDetails for client-side error thresholds.
	Payload *T
}

// ChangeValueRequest carries notification context and a batch of typed
// events from a single change.value in the webhook payload. WhatsApp
// delivers these as arrays — a single POST can contain multiple group
// changes, status updates, or user preference changes.
type ChangeValueRequest[T any] struct {
	// Notification identifies the WABA and phone number that received
	// the webhook event.
	Notification *MessageNotificationContext
	// Payload is the array of typed events from change.value: []*Group,
	// []*Status, []*UserPreference, etc.
	Payload []*T
}

// BusinessRequest carries context and payload for a business-account
// webhook event: template lifecycle changes, phone number verification,
// account reviews, security alerts, and call events.
type BusinessRequest[T any] struct {
	// Context identifies the WABA, the change field that triggered the
	// event, and the entry timestamp from the webhook envelope.
	Context *BusinessNotificationContext
	// Payload is the typed event data: *AlertNotification for messaging
	// limit alerts, *TemplateStatusUpdateNotification for template
	// approval/rejection, *SecurityNotification for PIN changes, etc.
	Payload *T
}

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

// IsHistoryMessage reports whether this message originated from a history
// sync webhook. History messages carry a [HistoryContext] with delivery status.
func (message *Message) IsHistoryMessage() bool {
	return message.HistoryContext != nil
}

// IsMediaPlaceholder reports whether this is a history thread message whose
// media content has not yet been delivered. The actual media asset arrives
// in a separate webhook.
func (message *Message) IsMediaPlaceholder() bool {
	return message.Type == MessageTypeMediaPlaceholder
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
