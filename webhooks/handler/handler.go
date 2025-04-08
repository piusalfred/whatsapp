package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/piusalfred/whatsapp/message"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	"github.com/piusalfred/whatsapp/webhooks"
)

type (
	Notification struct {
		Object string  `json:"object"`
		Entry  []Entry `json:"entry"`
	}

	Entry struct {
		ID      string   `json:"id"`
		Time    int64    `json:"time"`
		Changes []Change `json:"changes"`
	}

	Change struct {
		Field string `json:"field"`
		Value *Value `json:"value"`
	}

	Value struct {
		Event                        string            `json:"event,omitempty"`
		MessageTemplateID            int64             `json:"message_template_id,omitempty"`
		MessageTemplateName          string            `json:"message_template_name,omitempty"`
		MessageTemplateLanguage      string            `json:"message_template_language,omitempty"`
		Reason                       *string           `json:"reason,omitempty"`
		PreviousCategory             string            `json:"previous_category,omitempty"`
		PreviousQualityScore         string            `json:"previous_quality_score,omitempty"`
		NewQualityScore              string            `json:"new_quality_score,omitempty"`
		NewCategory                  string            `json:"new_category,omitempty"`
		DisplayPhoneNumber           string            `json:"display_phone_number,omitempty"`
		PhoneNumber                  string            `json:"phone_number,omitempty"`
		CurrentLimit                 string            `json:"current_limit,omitempty"`
		MaxDailyConversationPerPhone int               `json:"max_daily_conversation_per_phone,omitempty"`
		MaxPhoneNumbersPerBusiness   int               `json:"max_phone_numbers_per_business,omitempty"`
		MaxPhoneNumbersPerWABA       int               `json:"max_phone_numbers_per_waba,omitempty"`
		RejectionReason              string            `json:"rejection_reason,omitempty"`
		RequestedVerifiedName        string            `json:"requested_verified_name,omitempty"`
		RestrictionInfo              []RestrictionInfo `json:"restriction_info,omitempty"`
		BanInfo                      *BanInfo          `json:"ban_info,omitempty"`
		Decision                     string            `json:"decision,omitempty"`
		DisableInfo                  *DisableInfo      `json:"disable_info,omitempty"`
		OtherInfo                    *OtherInfo        `json:"other_info,omitempty"`
		ViolationInfo                *ViolationInfo    `json:"violation_info,omitempty"`
		EntityType                   string            `json:"entity_type,omitempty"`
		EntityID                     string            `json:"entity_id,omitempty"`
		AlertSeverity                string            `json:"alert_severity,omitempty"`
		AlertStatus                  string            `json:"alert_status,omitempty"`
		AlertType                    string            `json:"alert_type,omitempty"`
		AlertDescription             string            `json:"alert_description,omitempty"`
		Message                      string            `json:"message"`                  // Descriptive message of the event
		FlowID                       string            `json:"flow_id"`                  // EntryID of the flow
		OldStatus                    string            `json:"old_status,omitempty"`     // Previous status of the flow (optional)
		NewStatus                    string            `json:"new_status,omitempty"`     // New status of the flow (optional)
		ErrorRate                    float64           `json:"error_rate,omitempty"`     // Overall error rate for the alert (optional)
		Threshold                    int               `json:"threshold,omitempty"`      // Alert threshold that was reached or recovered from
		AlertState                   string            `json:"alert_state,omitempty"`    // Status of the alert, e.g., "ACTIVATED" or "DEACTIVATED"
		Errors                       []ErrorInfo       `json:"errors,omitempty"`         // List of errors describing the alert (optional)
		P50Latency                   int               `json:"p50_latency,omitempty"`    // P50 latency of the endpoint requests (optional)
		P90Latency                   int               `json:"p90_latency,omitempty"`    // P90 latency of the endpoint requests (optional)
		RequestsCount                int               `json:"requests_count,omitempty"` // Number of requests used to calculate metric (optional)
		Availability                 int               `json:"availability"`
		MessagingProduct             string            `json:"messaging_product,omitempty"`
		Metadata                     *Metadata         `json:"metadata,omitempty"`
		Contacts                     []*Contact        `json:"contacts,omitempty"`
		UserPreferences              []*UserPreference `json:"user_preferences,omitempty"`
		Messages                     []*Message        `json:"messages,omitempty"`
		Statuses                     []*Status         `json:"statuses,omitempty"`
	}

	ErrorInfo struct {
		ErrorType  string             `json:"error_type,omitempty"`
		ErrorRate  float64            `json:"error_rate,omitempty"`
		ErrorCount int64              `json:"error_count,omitempty"`
		Message    string             `json:"message,omitempty"`
		Type       string             `json:"type,omitempty"`
		Code       int                `json:"code,omitempty"`
		Data       *werrors.ErrorData `json:"error_data,omitempty"`
		Subcode    int                `json:"error_subcode,omitempty"`
		UserTitle  string             `json:"error_user_title,omitempty"`
		UserMsg    string             `json:"error_user_msg,omitempty"`
		FBTraceID  string             `json:"fbtrace_id,omitempty"`
	}

	BanInfo struct {
		WABABanState WABABanState `json:"waba_ban_state"` // e.g., DISABLE, REINSTATE, SCHEDULE_FOR_DISABLE
		WABABanDate  string       `json:"waba_ban_date"`  // Date of the ban
		CurrentLimit string       `json:"current_limit"`  // Current tier limit of the account
	}

	DisableInfo struct {
		DisableDate string `json:"disable_date"` // Date when the template was disabled
	}

	OtherInfo struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	ViolationInfo struct {
		ViolationType string `json:"violation_type"` // e.g., "ACCOUNT_VIOLATION"
	}

	RestrictionInfo struct {
		RestrictionType RestrictionType `json:"restriction_type"` // e.g., "RESTRICTED_BIZ_INITIATED_MESSAGING"
		Expiration      string          `json:"expiration"`       // Expiration date of the restriction
	}
)

func (info ErrorInfo) Error() *werrors.Error {
	return &werrors.Error{
		Message:   info.Message,
		Type:      info.Type,
		Code:      info.Code,
		Data:      info.Data,
		Subcode:   info.Subcode,
		UserTitle: info.UserTitle,
		UserMsg:   info.UserMsg,
		FBTraceID: info.FBTraceID,
	}
}

type MediaMessageHandler MessageHandler[message.MediaInfo]

type Handler struct {
	FlowStatusChangeHandler         EventHandler[FlowNotificationContext, StatusChangeDetails]
	FlowClientErrorRateHandler      EventHandler[FlowNotificationContext, ClientErrorRateDetails]
	FlowEndpointErrorRateHandler    EventHandler[FlowNotificationContext, EndpointErrorRateDetails]
	FlowEndpointLatencyHandler      EventHandler[FlowNotificationContext, EndpointLatencyDetails]
	FlowEndpointAvailabilityHandler EventHandler[FlowNotificationContext, EndpointAvailabilityDetails]
	AlertsHandler                   EventHandler[BusinessNotificationContext, AlertNotification]
	TemplateStatusHandler           EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification]
	TemplateCategoryHandler         EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification]
	TemplateQualityHandler          EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification]
	PhoneNumberNameHandler          EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate]
	CapabilityUpdateHandler         EventHandler[BusinessNotificationContext, CapabilityUpdate]
	AccountUpdateHandler            EventHandler[BusinessNotificationContext, AccountUpdate]
	PhoneNumberQualityUpdateHandler EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate]
	AccountReviewUpdateHandler      EventHandler[BusinessNotificationContext, AccountReviewUpdate]
	ButtonMessageHandler            MessageHandler[Button]
	TextMessageHandler              MessageHandler[Text]
	OrderMessageHandler             MessageHandler[Order]
	LocationMessageHandler          MessageHandler[message.Location]
	ContactsMessageHandler          MessageHandler[message.Contacts]
	ReactionHandler                 MessageHandler[message.Reaction]
	ProductEnquiryHandler           MessageHandler[Text]
	InteractiveMessageHandler       MessageHandler[Interactive]
	ButtonReplyMessageHandler       MessageHandler[ButtonReply]
	ListReplyMessageHandler         MessageHandler[ListReply]
	FlowCompletionMessageHandler    MessageHandler[NFMReply]
	ReferralMessageHandler          MessageHandler[ReferralNotification]
	CustomerIDChangeHandler         MessageHandler[Identity]
	SystemMessageHandler            MessageHandler[System]
	AudioMessageHandler             MediaMessageHandler
	VideoMessageHandler             MediaMessageHandler
	ImageMessageHandler             MediaMessageHandler
	DocumentMessageHandler          MediaMessageHandler
	StickerMessageHandler           MediaMessageHandler
	UserPreferencesUpdateHandler    func(ctx context.Context, notificationCtx *MessageNotificationContext, preferences []*UserPreference) error
	ErrorHandler                    func(ctx context.Context, err error) error
	MessageNotificationErrorHandler MessageChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler      MessageChangeValueHandler[Status]
	MessageReceivedHandler          MessageChangeValueHandler[Message]
	MessageUserPreferenceHandler    MessageChangeValueHandler[UserPreference]
	MessageErrorsHandler            MessageErrorsHandler
	RequestWelcome                  func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, message *Message) error

	availableHandlers []string
}

func (handler *Handler) ensureHandlers() {
	if handler.availableHandlers == nil {
		handler.availableHandlers = make([]string, 0)
	}

	if handler.FlowStatusChangeHandler != nil {
		handler.availableHandlers = append(handler.availableHandlers, EventFlowStatusChange)
	}

	if handler.FlowClientErrorRateHandler != nil {
		handler.availableHandlers = append(handler.availableHandlers, EventClientErrorRate)
	}

	if handler.FlowEndpointErrorRateHandler != nil {
		handler.availableHandlers = append(handler.availableHandlers, EventEndpointErrorRate)
	}

	if handler.FlowEndpointLatencyHandler != nil {
		handler.availableHandlers = append(handler.availableHandlers, EventEndpointLatency)
	}

	if handler.FlowEndpointAvailabilityHandler != nil {
		handler.availableHandlers = append(handler.availableHandlers, EventEndpointAvailability)
	}
}

func (handler *Handler) isHandlerSet(name string) bool {
	handler.ensureHandlers()

	return slices.Contains(handler.availableHandlers, name)
}

func (handler *Handler) HandleNotification(ctx context.Context, notification *Notification) *webhooks.Response {
	for _, entry := range notification.Entry {
		for _, change := range entry.Changes {
			response, done := handler.handleNotificationChange(ctx, notification, change, entry)
			if done {
				return response
			}
		}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}

func (handler *Handler) handleNotificationChange(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) (*webhooks.Response, bool) {
	if change.Field == ChangeFieldFlows.String() {
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
			if handler.ErrorHandler != nil {
				if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
				}
			}
		}
	}

	if change.Field == ChangeFieldAccountAlerts.String() {
		if handler.AlertsHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.AlertsHandler.HandleEvent(ctx, notificationCtx, change.Value.AlertNotification()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldTemplateStatusUpdate.String() {
		if handler.TemplateStatusHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.TemplateStatusHandler.HandleEvent(ctx, notificationCtx, change.Value.TemplateStatusUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldTemplateCategoryUpdate.String() {
		if handler.TemplateCategoryHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.TemplateCategoryHandler.HandleEvent(ctx, notificationCtx, change.Value.TemplateCategoryUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldTemplateQualityUpdate.String() {
		if handler.TemplateQualityHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.TemplateQualityHandler.HandleEvent(ctx, notificationCtx, change.Value.TemplateQualityUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldPhoneNumberNameUpdate.String() {
		if handler.PhoneNumberNameHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.PhoneNumberNameHandler.HandleEvent(ctx, notificationCtx, change.Value.PhoneNumberNameUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldPhoneNumberQualityUpdate.String() {
		if handler.PhoneNumberQualityUpdateHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.PhoneNumberQualityUpdateHandler.HandleEvent(ctx, notificationCtx, change.Value.PhoneNumberQualityUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldAccountUpdate.String() {
		if handler.AccountUpdateHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.AccountUpdateHandler.HandleEvent(ctx, notificationCtx, change.Value.AccountUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldAccountReviewUpdate.String() {
		if handler.AccountReviewUpdateHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.AccountReviewUpdateHandler.HandleEvent(ctx, notificationCtx, change.Value.AccountReviewUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldBusinessCapabilityUpdate.String() {
		if handler.CapabilityUpdateHandler != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := handler.CapabilityUpdateHandler.HandleEvent(ctx, notificationCtx, change.Value.CapabilityUpdate()); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldUserPreferences.String() {
		if handler.UserPreferencesUpdateHandler != nil {
			notificationCtx := &MessageNotificationContext{
				EntryID:          entry.ID,
				MessagingProduct: change.Value.MessagingProduct,
				Contacts:         change.Value.Contacts,
				Metadata:         change.Value.Metadata,
			}

			if err := handler.UserPreferencesUpdateHandler(ctx, notificationCtx, change.Value.UserPreferences); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true
					}
				}
			}
		}
	}

	if change.Field == ChangeFieldMessages.String() {
		response, b, done := handler.handleNotificationMessageItem(ctx, entry, change)
		if done {
			return response, b
		}
	}
	return nil, false
}

func (handler *Handler) handleNotificationMessageItem(
	ctx context.Context,
	entry Entry,
	change Change,
) (*webhooks.Response, bool, bool) {
	notificationCtx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Contacts:         change.Value.Contacts,
		Metadata:         change.Value.Metadata,
	}

	for _, m := range change.Value.Messages {
		if err := handler.handleNotificationMessage(ctx, notificationCtx, m); err != nil {
			if handler.ErrorHandler != nil {
				if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true, true
				}
			}
		}
	}

	if handler.MessageNotificationErrorHandler != nil {
		for _, ev := range change.Value.Errors {
			if err := handler.MessageNotificationErrorHandler.Handle(ctx, notificationCtx, ev.Error()); err != nil {
				return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true, true
			}
		}
	}

	if handler.MessageStatusChangeHandler != nil {
		for _, sv := range change.Value.Statuses {
			if err := handler.MessageStatusChangeHandler.Handle(ctx, notificationCtx, sv); err != nil {
				if handler.ErrorHandler != nil {
					if handlerErr := handler.ErrorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}, true, true
					}
				}
			}
		}
	}
	return nil, false, false
}

// ChangeField represent the name of the field in which the webhook notification payload
// is embedded.
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

const (
	RestrictionTypeRestrictedAddPhoneNumber    RestrictionType = "RESTRICTED_ADD_PHONE_NUMBER_ACTION"
	RestrictionTypeRestrictedBizInitiated      RestrictionType = "RESTRICTED_BIZ_INITIATED_MESSAGING"
	RestrictionTypeRestrictedCustomerInitiated RestrictionType = "RESTRICTED_CUSTOMER_INITIATED_MESSAGING"

	WABABanStateDisable            WABABanState = "DISABLE"
	WABABanStateReinstate          WABABanState = "REINSTATE"
	WABABanStateScheduleForDisable WABABanState = "SCHEDULE_FOR_DISABLE"
)

type (
	BusinessNotificationContext struct {
		Object      string
		EntryID     string
		EntryTime   int64
		ChangeField string
	}

	WABABanState string

	RestrictionType string

	AlertNotification struct {
		EntityType       string `json:"entity_type,omitempty"`
		EntityID         string `json:"entity_id,omitempty"`
		AlertSeverity    string `json:"alert_severity,omitempty"`
		AlertStatus      string `json:"alert_status,omitempty"`
		AlertType        string `json:"alert_type,omitempty"`
		AlertDescription string `json:"alert_description,omitempty"`
	}

	TemplateStatusUpdateNotification struct {
		Event                   string       `json:"event,omitempty"`
		MessageTemplateID       int64        `json:"message_template_id,omitempty"`
		MessageTemplateName     string       `json:"message_template_name,omitempty"`
		MessageTemplateLanguage string       `json:"message_template_language,omitempty"`
		Reason                  string       `json:"reason,omitempty"`
		DisableInfo             *DisableInfo `json:"disable_info,omitempty"`
		OtherInfo               *OtherInfo   `json:"other_info,omitempty"`
	}

	TemplateCategoryUpdateNotification struct {
		MessageTemplateID       int64  `json:"message_template_id,omitempty"`
		MessageTemplateName     string `json:"message_template_name,omitempty"`
		MessageTemplateLanguage string `json:"message_template_language,omitempty"`
		PreviousCategory        string `json:"previous_category,omitempty"`
		NewCategory             string `json:"new_category,omitempty"`
	}

	TemplateQualityUpdateNotification struct {
		PreviousQualityScore    string `json:"previous_quality_score,omitempty"`
		NewQualityScore         string `json:"new_quality_score,omitempty"`
		MessageTemplateID       int64  `json:"message_template_id,omitempty"`
		MessageTemplateName     string `json:"message_template_name,omitempty"`
		MessageTemplateLanguage string `json:"message_template_language,omitempty"`
	}

	PhoneNumberNameUpdate struct {
		PhoneNumber           string `json:"display_phone_number"`
		Decision              string `json:"decision"`
		RequestedVerifiedName string `json:"requested_verified_name"`
		RejectionReason       string `json:"rejection_reason"`
	}

	PhoneNumberQualityUpdate struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		Event              string `json:"event,omitempty"`
		CurrentLimit       string `json:"current_limit,omitempty"`
	}

	AccountReviewUpdate struct {
		Decision string `json:"decision,omitempty"`
	}

	AccountUpdate struct {
		PhoneNumber     string            `json:"phone_number,omitempty"`
		Event           string            `json:"event,omitempty"`
		RestrictionInfo []RestrictionInfo `json:"restriction_info,omitempty"`
		BanInfo         *BanInfo          `json:"ban_info,omitempty"`
		ViolationInfo   *ViolationInfo    `json:"violation_info,omitempty"`
	}

	CapabilityUpdate struct {
		MaxDailyConversationPerPhone int `json:"max_daily_conversation_per_phone,omitempty"`
		MaxPhoneNumbersPerBusiness   int `json:"max_phone_numbers_per_business,omitempty"`
	}
)

func (value *Value) AlertNotification() *AlertNotification {
	return &AlertNotification{
		EntityType:       value.EntityType,
		EntityID:         value.EntityID,
		AlertSeverity:    value.AlertSeverity,
		AlertStatus:      value.AlertStatus,
		AlertType:        value.AlertType,
		AlertDescription: value.AlertDescription,
	}
}

func (value *Value) TemplateStatusUpdate() *TemplateStatusUpdateNotification {
	return &TemplateStatusUpdateNotification{
		Event:                   value.Event,
		MessageTemplateID:       value.MessageTemplateID,
		MessageTemplateName:     value.MessageTemplateName,
		MessageTemplateLanguage: value.MessageTemplateLanguage,
		Reason:                  *value.Reason,
		DisableInfo:             value.DisableInfo,
		OtherInfo:               value.OtherInfo,
	}
}

func (value *Value) TemplateCategoryUpdate() *TemplateCategoryUpdateNotification {
	return &TemplateCategoryUpdateNotification{
		MessageTemplateID:       value.MessageTemplateID,
		MessageTemplateName:     value.MessageTemplateName,
		MessageTemplateLanguage: value.MessageTemplateLanguage,
		PreviousCategory:        value.PreviousCategory,
		NewCategory:             value.NewCategory,
	}
}

func (value *Value) TemplateQualityUpdate() *TemplateQualityUpdateNotification {
	return &TemplateQualityUpdateNotification{
		PreviousQualityScore:    value.PreviousQualityScore,
		NewQualityScore:         value.NewQualityScore,
		MessageTemplateID:       value.MessageTemplateID,
		MessageTemplateName:     value.MessageTemplateName,
		MessageTemplateLanguage: value.MessageTemplateLanguage,
	}
}

func (value *Value) PhoneNumberNameUpdate() *PhoneNumberNameUpdate {
	return &PhoneNumberNameUpdate{
		PhoneNumber:           value.DisplayPhoneNumber,
		Decision:              value.Decision,
		RequestedVerifiedName: value.RequestedVerifiedName,
		RejectionReason:       value.RejectionReason,
	}
}

func (value *Value) PhoneNumberQualityUpdate() *PhoneNumberQualityUpdate {
	return &PhoneNumberQualityUpdate{
		DisplayPhoneNumber: value.DisplayPhoneNumber,
		Event:              value.Event,
		CurrentLimit:       value.CurrentLimit,
	}
}

func (value *Value) AccountReviewUpdate() *AccountReviewUpdate {
	return &AccountReviewUpdate{
		Decision: value.Decision,
	}
}

func (value *Value) AccountUpdate() *AccountUpdate {
	return &AccountUpdate{
		PhoneNumber:     value.PhoneNumber,
		Event:           value.Event,
		RestrictionInfo: value.RestrictionInfo,
		BanInfo:         value.BanInfo,
		ViolationInfo:   value.ViolationInfo,
	}
}

func (value *Value) CapabilityUpdate() *CapabilityUpdate {
	return &CapabilityUpdate{
		MaxDailyConversationPerPhone: value.MaxDailyConversationPerPhone,
		MaxPhoneNumbersPerBusiness:   value.MaxPhoneNumbersPerBusiness,
	}
}

func (handler *Handler) OnBusinessAlertNotification(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AlertNotification) error,
) {
	handler.AlertsHandler = EventHandlerFunc[BusinessNotificationContext, AlertNotification](fn)
}

func (handler *Handler) SetBusinessAlertNotificationHandler(
	fn EventHandler[BusinessNotificationContext, AlertNotification]) {
	handler.AlertsHandler = fn
}

func (handler *Handler) OnBusinessTemplateStatusUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateStatusUpdateNotification) error,
) {
	handler.TemplateStatusHandler = EventHandlerFunc[BusinessNotificationContext, TemplateStatusUpdateNotification](fn)
}

func (handler *Handler) SetBusinessTemplateStatusUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification]) {
	handler.TemplateStatusHandler = fn
}

func (handler *Handler) OnBusinessTemplateCategoryUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateCategoryUpdateNotification) error,
) {
	handler.TemplateCategoryHandler = EventHandlerFunc[BusinessNotificationContext, TemplateCategoryUpdateNotification](
		fn,
	)
}

func (handler *Handler) SetBusinessTemplateCategoryUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification]) {
	handler.TemplateCategoryHandler = fn
}

func (handler *Handler) OnBusinessTemplateQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateQualityUpdateNotification) error,
) {
	handler.TemplateQualityHandler = EventHandlerFunc[BusinessNotificationContext, TemplateQualityUpdateNotification](
		fn,
	)
}

func (handler *Handler) SetBusinessTemplateQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification]) {
	handler.TemplateQualityHandler = fn
}

func (handler *Handler) OnBusinessPhoneNumberNameUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberNameUpdate) error,
) {
	handler.PhoneNumberNameHandler = EventHandlerFunc[BusinessNotificationContext, PhoneNumberNameUpdate](fn)
}

func (handler *Handler) SetBusinessPhoneNumberNameUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate]) {
	handler.PhoneNumberNameHandler = fn
}

func (handler *Handler) OnBusinessPhoneNumberQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberQualityUpdate) error,
) {
	handler.PhoneNumberQualityUpdateHandler = EventHandlerFunc[BusinessNotificationContext, PhoneNumberQualityUpdate](
		fn,
	)
}

func (handler *Handler) SetBusinessPhoneNumberQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate]) {
	handler.PhoneNumberQualityUpdateHandler = fn
}

func (handler *Handler) OnBusinessAccountReviewUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountReviewUpdate) error,
) {
	handler.AccountReviewUpdateHandler = EventHandlerFunc[BusinessNotificationContext, AccountReviewUpdate](fn)
}

func (handler *Handler) SetBusinessAccountReviewUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountReviewUpdate]) {
	handler.AccountReviewUpdateHandler = fn
}

func (handler *Handler) OnBusinessAccountUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountUpdate) error,
) {
	handler.AccountUpdateHandler = EventHandlerFunc[BusinessNotificationContext, AccountUpdate](fn)
}

func (handler *Handler) SetBusinessAccountUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountUpdate]) {
	handler.AccountUpdateHandler = fn
}

func (handler *Handler) OnBusinessCapabilityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *CapabilityUpdate) error,
) {
	handler.CapabilityUpdateHandler = EventHandlerFunc[BusinessNotificationContext, CapabilityUpdate](fn)
}

func (handler *Handler) SetBusinessCapabilityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, CapabilityUpdate]) {
	handler.CapabilityUpdateHandler = fn
}

type (
	FlowNotificationContext struct {
		NotificationObject string // Corresponds to the 'object' field
		EntryID            string // Corresponds to the 'id' field in Entry
		EntryTime          int64  // Corresponds to the 'time' field in Entry
		ChangeField        string // Corresponds to the 'field' in Changes
		EventName          string // Corresponds to 'event' field in Value
		EventMessage       string // Corresponds to 'message' field in Value
		FlowID             string // Corresponds to 'flow_id' field in Value
	}

	StatusChangeDetails struct {
		OldStatus string `json:"old_status,omitempty"`
		NewStatus string `json:"new_status,omitempty"`
	}

	ClientErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointErrorRateDetails struct {
		ErrorRate  float64     `json:"error_rate,omitempty"`
		Threshold  int         `json:"threshold,omitempty"`
		AlertState string      `json:"alert_state,omitempty"`
		Errors     []ErrorInfo `json:"errors,omitempty"`
	}

	EndpointLatencyDetails struct {
		P50Latency    int    `json:"p50_latency,omitempty"`
		P90Latency    int    `json:"p90_latency,omitempty"`
		RequestsCount int    `json:"requests_count,omitempty"`
		Threshold     int    `json:"threshold,omitempty"`
		AlertState    string `json:"alert_state,omitempty"`
	}

	EndpointAvailabilityDetails struct {
		Availability int    `json:"availability,omitempty"`
		Threshold    int    `json:"threshold,omitempty"`
		AlertState   string `json:"alert_state,omitempty"`
	}
)

func (value *Value) FlowStatusChange() *StatusChangeDetails {
	return &StatusChangeDetails{
		OldStatus: value.OldStatus,
		NewStatus: value.NewStatus,
	}
}

func (value *Value) FlowClientErrorRate() *ClientErrorRateDetails {
	return &ClientErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

func (value *Value) FlowEndpointErrorRate() *EndpointErrorRateDetails {
	return &EndpointErrorRateDetails{
		ErrorRate:  value.ErrorRate,
		Threshold:  value.Threshold,
		AlertState: value.AlertState,
		Errors:     value.Errors,
	}
}

func (value *Value) FlowEndpointLatency() *EndpointLatencyDetails {
	return &EndpointLatencyDetails{
		P50Latency:    value.P50Latency,
		P90Latency:    value.P90Latency,
		RequestsCount: value.RequestsCount,
		Threshold:     value.Threshold,
		AlertState:    value.AlertState,
	}
}

func (value *Value) FlowEndpointAvailability() *EndpointAvailabilityDetails {
	return &EndpointAvailabilityDetails{
		Availability: value.Availability,
		Threshold:    value.Threshold,
		AlertState:   value.AlertState,
	}
}

func (handler *Handler) handleFlowNotification(
	ctx context.Context,
	notificationContext *FlowNotificationContext,
	value *Value,
) error {
	if handler.isHandlerSet(value.Event) {
		switch value.Event {
		case EventFlowStatusChange:
			details := value.FlowStatusChange()
			if err := handler.FlowStatusChangeHandler.HandleEvent(ctx, notificationContext, details); err != nil {
				return fmt.Errorf("handle flow status change event: %w", err)
			}

		case EventClientErrorRate:
			details := value.FlowClientErrorRate()
			if err := handler.FlowClientErrorRateHandler.HandleEvent(ctx, notificationContext, details); err != nil {
				return fmt.Errorf("handle client error rate event: %w", err)
			}

		case EventEndpointErrorRate:
			details := value.FlowEndpointErrorRate()
			if err := handler.FlowEndpointErrorRateHandler.HandleEvent(ctx, notificationContext, details); err != nil {
				return fmt.Errorf("handle endpoint error rate event: %w", err)
			}

		case EventEndpointLatency:
			details := value.FlowEndpointLatency()
			if err := handler.FlowEndpointLatencyHandler.HandleEvent(ctx, notificationContext, details); err != nil {
				return fmt.Errorf("handle endpoint latency event: %w", err)
			}

		case EventEndpointAvailability:
			details := value.FlowEndpointAvailability()
			if err := handler.FlowEndpointAvailabilityHandler.HandleEvent(ctx, notificationContext, details); err != nil {
				return fmt.Errorf("handle endpoint availability event: %w", err)
			}
		}
	}

	return nil
}

func (handler *Handler) OnFlowStatusChange(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *StatusChangeDetails) error,
) {
	handler.FlowStatusChangeHandler = EventHandlerFunc[FlowNotificationContext, StatusChangeDetails](fn)
}

func (handler *Handler) SetFlowStatusChangeHandler(fn EventHandler[FlowNotificationContext, StatusChangeDetails]) {
	handler.FlowStatusChangeHandler = fn
}

func (handler *Handler) OnFlowClientErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *ClientErrorRateDetails) error,
) {
	handler.FlowClientErrorRateHandler = EventHandlerFunc[FlowNotificationContext, ClientErrorRateDetails](fn)
}

func (handler *Handler) SetFlowClientErrorRateHandler(
	fn EventHandler[FlowNotificationContext, ClientErrorRateDetails],
) {
	handler.FlowClientErrorRateHandler = fn
}

func (handler *Handler) OnFlowEndpointErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointErrorRateDetails) error,
) {
	handler.FlowEndpointErrorRateHandler = EventHandlerFunc[FlowNotificationContext, EndpointErrorRateDetails](fn)
}

func (handler *Handler) SetFlowEndpointErrorRateHandler(
	fn EventHandler[FlowNotificationContext, EndpointErrorRateDetails],
) {
	handler.FlowEndpointErrorRateHandler = fn
}

func (handler *Handler) OnFlowEndpointLatency(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointLatencyDetails) error,
) {
	handler.FlowEndpointLatencyHandler = EventHandlerFunc[FlowNotificationContext, EndpointLatencyDetails](fn)
}

func (handler *Handler) SetFlowEndpointLatencyHandler(
	fn EventHandler[FlowNotificationContext, EndpointLatencyDetails],
) {
	handler.FlowEndpointLatencyHandler = fn
}

func (handler *Handler) OnFlowEndpointAvailability(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointAvailabilityDetails) error,
) {
	handler.FlowEndpointAvailabilityHandler = EventHandlerFunc[FlowNotificationContext, EndpointAvailabilityDetails](fn)
}

func (handler *Handler) SetFlowEndpointAvailabilityHandler(
	fn EventHandler[FlowNotificationContext, EndpointAvailabilityDetails],
) {
	handler.FlowEndpointAvailabilityHandler = fn
}

const (
	MessageTypeAudio          MessageType = "audio"
	MessageTypeButton         MessageType = "button"
	MessageTypeDocument       MessageType = "document"
	MessageTypeText           MessageType = "text"
	MessageTypeImage          MessageType = "image"
	MessageTypeInteractive    MessageType = "interactive"
	MessageTypeOrder          MessageType = "order"
	MessageTypeSticker        MessageType = "sticker"
	MessageTypeSystem         MessageType = "system"
	MessageTypeUnknown        MessageType = "unknown"
	MessageTypeVideo          MessageType = "video"
	MessageTypeLocation       MessageType = "location"
	MessageTypeReaction       MessageType = "reaction"
	MessageTypeContacts       MessageType = "contacts"
	MessageTypeRequestWelcome MessageType = "request_welcome"
)

// MessageType is type of message that has been received by the business that has subscribed
// to Webhooks. Possible value can be one of the following: audio,button,document,text,image,
// interactive,order,sticker,system – for customer number change messages,unknown and video
// The documentation is not clear in case of location,reaction and contacts. They will be included
// just in case.
type MessageType string

// ParseMessageType parses the message type from a string.
func ParseMessageType(s string) MessageType {
	msgMap := map[string]MessageType{
		"audio":           MessageTypeAudio,
		"button":          MessageTypeButton,
		"document":        MessageTypeDocument,
		"text":            MessageTypeText,
		"image":           MessageTypeImage,
		"interactive":     MessageTypeInteractive,
		"order":           MessageTypeOrder,
		"sticker":         MessageTypeSticker,
		"system":          MessageTypeSystem,
		"unknown":         MessageTypeUnknown,
		"video":           MessageTypeVideo,
		"location":        MessageTypeLocation,
		"reaction":        MessageTypeReaction,
		"contacts":        MessageTypeContacts,
		"request_welcome": MessageTypeRequestWelcome,
	}

	msgType, ok := msgMap[strings.TrimSpace(strings.ToLower(s))]
	if !ok {
		return ""
	}

	return msgType
}

type (
	Contact struct {
		Profile *Profile `json:"profile,omitempty"`
		WaID    string   `json:"wa_id,omitempty"`
	}

	// Message contains the information of a message. It is embedded in the Value object.
	Message struct {
		Audio       *message.MediaInfo `json:"audio,omitempty"`
		Button      *Button            `json:"button,omitempty"`
		Context     *Context           `json:"context,omitempty"`
		Document    *message.MediaInfo `json:"document,omitempty"`
		Errors      []*werrors.Error   `json:"errors,omitempty"`
		From        string             `json:"from,omitempty"`
		ID          string             `json:"id,omitempty"`
		Identity    *Identity          `json:"identity,omitempty"`
		Image       *message.MediaInfo `json:"image,omitempty"`
		Interactive *Interactive       `json:"interactive,omitempty"`
		Order       *Order             `json:"order,omitempty"`
		Referral    *Referral          `json:"referral,omitempty"`
		Sticker     *message.MediaInfo `json:"sticker,omitempty"`
		System      *System            `json:"system,omitempty"`
		Text        *Text              `json:"text,omitempty"`
		Timestamp   string             `json:"timestamp,omitempty"`
		Type        string             `json:"type,omitempty"`
		Video       *message.MediaInfo `json:"video,omitempty"`
		Contacts    *message.Contacts  `json:"contacts,omitempty"`
		Location    *message.Location  `json:"location,omitempty"`
		Reaction    *message.Reaction  `json:"reaction,omitempty"`
	}

	Status struct {
		ID                    string           `json:"id,omitempty"`
		RecipientID           string           `json:"recipient_id,omitempty"`
		StatusValue           string           `json:"status,omitempty"`
		Timestamp             string           `json:"timestamp,omitempty"`
		Conversation          *Conversation    `json:"conversation,omitempty"`
		Pricing               *Pricing         `json:"pricing,omitempty"`
		Errors                []*werrors.Error `json:"errors,omitempty"`
		BizOpaqueCallbackData string           `json:"biz_opaque_callback_data,omitempty"`
	}

	UserPreference struct {
		WaID      string `json:"wa_id"`
		Detail    string `json:"detail"`
		Category  string `json:"category"` // always "marketing_messages"
		Value     string `json:"value"`    // can be "stop" or "resume"
		Timestamp string `json:"timestamp"`
	}

	Metadata struct {
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		PhoneNumberID      string `json:"phone_number_id,omitempty"`
	}

	ReferralNotification struct {
		Text     *Text
		Referral *Referral
	}

	Pricing struct {
		Billable     bool   `json:"billable,omitempty"` // Deprecated
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
		Type        string       `json:"type,omitempty"`
		ButtonReply *ButtonReply `json:"button_reply,omitempty"`
		ListReply   *ListReply   `json:"list_reply,omitempty"`
		NFMReply    *NFMReply    `json:"nfm_reply,omitempty"`
	}

	NFMReply struct {
		Name         string          `json:"name"`          // Always "flow"
		Body         string          `json:"body"`          // Always "Sent"
		ResponseJSON json.RawMessage `json:"response_json"` // Flow-specific data (JSON string)
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
		Forwarded           bool   `json:"forwarded,omitempty"`
		FrequentlyForwarded bool   `json:"frequently_forwarded,omitempty"`
		From                string `json:"from,omitempty"`
		ID                  string `json:"id,omitempty"`
		ReferredProduct     *ReferredProduct
		Type                string `json:"type,omitempty"`
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
	// these are common fields to all type of messages.
	// From The customer's phone number who sent the message to the business.
	// MessageID The MessageID for the message that was received by the business. You could use messages
	// endpoint to mark this specific message as read.
	// Timestamp The timestamp for when the message was received by the business.
	// Type The type of message that was received by the business.
	// Context The context of the message. Only included when a user replies or interacts with one
	// of your messages.
	MessageInfo struct {
		From      string
		MessageID string
		Timestamp string
		Type      string
		Context   *Context
	}
)

type (
	MessageHandlerFunc[T any] func(
		ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, message *T) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, message *T) error
	}
)

func (fn MessageHandlerFunc[T]) Handle(ctx context.Context, nctx *MessageNotificationContext,
	mctx *MessageInfo, message *T) error {
	return fn(ctx, nctx, mctx, message)
}

type (
	MessageChangeValueHandler[T any] interface {
		Handle(ctx context.Context, nctx *MessageNotificationContext, value *T) error
	}

	MessageChangeValueHandlerFunc[T any] func(ctx context.Context, nctx *MessageNotificationContext, value *T) error
)

func (f MessageChangeValueHandlerFunc[T]) Handle(
	ctx context.Context,
	nctx *MessageNotificationContext,
	value *T,
) error {
	return f(ctx, nctx, value)
}

const (
	InteractiveTypeListReply   = "list_reply"
	InteractiveTypeButtonReply = "button_reply"
	InteractiveTypeNFMReply    = "nfm_reply"
)

type (
	MessageErrorsHandlerFunc func(
		ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, errors []*werrors.Error) error
	MessageErrorsHandler interface {
		Handle(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, errors []*werrors.Error) error
	}
)

func (handler *Handler) handleNotificationMessage(ctx context.Context,
	nctx *MessageNotificationContext, message *Message,
) error {
	mctx := &MessageInfo{
		From:      message.From,
		MessageID: message.ID,
		Timestamp: message.Timestamp,
		Type:      message.Type,
		Context:   message.Context,
	}

	messageType := ParseMessageType(message.Type)
	switch messageType {
	case MessageTypeOrder:
		if err := handler.OrderMessageHandler.Handle(ctx, nctx, mctx, message.Order); err != nil {
			return err
		}

		return nil

	case MessageTypeButton:
		if err := handler.ButtonMessageHandler.Handle(ctx, nctx, mctx, message.Button); err != nil {
			return err
		}

		return nil

	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return handler.handleMediaMessage(ctx, nctx, message, mctx)

	case MessageTypeInteractive:
		return handler.handleInteractiveNotification(ctx, nctx, message, mctx)

	case MessageTypeSystem:
		if err := handler.SystemMessageHandler.Handle(ctx, nctx, mctx, message.System); err != nil {
			return err
		}

		return nil

	case MessageTypeUnknown:
		if err := handler.MessageErrorsHandler.Handle(ctx, nctx, mctx, message.Errors); err != nil {
			return err
		}

		return nil

	case MessageTypeText:
		return handler.handleTextNotification(ctx, nctx, message, mctx)

	case MessageTypeRequestWelcome:
		if handler.RequestWelcome != nil {
			if err := handler.RequestWelcome(ctx, nctx, mctx, message); err != nil {
				return err
			}
		}

		return nil

	case MessageTypeReaction:
		if err := handler.ReactionHandler.Handle(ctx, nctx, mctx, message.Reaction); err != nil {
			return err
		}

		return nil

	case MessageTypeLocation:
		if err := handler.LocationMessageHandler.Handle(ctx, nctx, mctx, message.Location); err != nil {
			return err
		}

		return nil

	case MessageTypeContacts:
		if err := handler.ContactsMessageHandler.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
			return err
		}

		return nil

	default:
		return handler.handleDefaultNotificationMessage(ctx, nctx, message, mctx)
	}
}

func (handler *Handler) handleMediaMessage(ctx context.Context, nctx *MessageNotificationContext,
	message *Message, mctx *MessageInfo,
) error {
	messageType := ParseMessageType(message.Type)
	switch messageType { //nolint:exhaustive // ok
	case MessageTypeAudio:
		if err := handler.AudioMessageHandler.Handle(ctx, nctx, mctx, message.Audio); err != nil {
			return err
		}

		return nil

	case MessageTypeVideo:
		if err := handler.VideoMessageHandler.Handle(ctx, nctx, mctx, message.Video); err != nil {
			return err
		}

		return nil

	case MessageTypeImage:
		if err := handler.ImageMessageHandler.Handle(ctx, nctx, mctx, message.Image); err != nil {
			return err
		}

		return nil

	case MessageTypeDocument:
		if err := handler.DocumentMessageHandler.Handle(ctx, nctx, mctx, message.Document); err != nil {
			return err
		}

		return nil

	case MessageTypeSticker:
		if err := handler.StickerMessageHandler.Handle(ctx, nctx, mctx, message.Sticker); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (handler *Handler) handleTextNotification(ctx context.Context, nctx *MessageNotificationContext,
	message *Message, mctx *MessageInfo,
) error {
	if message.Referral != nil {
		reff := &ReferralNotification{
			Text:     message.Text,
			Referral: message.Referral,
		}
		if err := handler.ReferralMessageHandler.Handle(ctx, nctx, mctx, reff); err != nil {
			return err
		}

		return nil
	}

	if mctx.Context != nil {
		if err := handler.ProductEnquiryHandler.Handle(ctx, nctx, mctx, message.Text); err != nil {
			return err
		}

		return nil
	}

	if err := handler.TextMessageHandler.Handle(ctx, nctx, mctx, message.Text); err != nil {
		return err
	}

	return nil
}

func (handler *Handler) handleDefaultNotificationMessage(ctx context.Context, nctx *MessageNotificationContext,
	message *Message, mctx *MessageInfo,
) error {
	if message.Contacts != nil {
		if err := handler.ContactsMessageHandler.Handle(ctx, nctx, mctx, message.Contacts); err != nil {
			return err
		}

		return nil
	}
	if message.Location != nil {
		if err := handler.LocationMessageHandler.Handle(ctx, nctx, mctx, message.Location); err != nil {
			return err
		}

		return nil
	}

	if message.Identity != nil {
		if err := handler.CustomerIDChangeHandler.Handle(ctx, nctx, mctx, message.Identity); err != nil {
			return err
		}

		return nil
	}

	return errors.New("unsupported message type")
}

func (handler *Handler) handleInteractiveNotification(ctx context.Context,
	nctx *MessageNotificationContext, message *Message, mctx *MessageInfo,
) error {
	switch message.Interactive.Type {
	case InteractiveTypeListReply:
		if err := handler.ListReplyMessageHandler.Handle(ctx, nctx, mctx, message.Interactive.ListReply); err != nil {
			return fmt.Errorf("handle list reply: %w", err)
		}

		return nil
	case InteractiveTypeButtonReply:
		if err := handler.ButtonReplyMessageHandler.Handle(ctx, nctx, mctx, message.Interactive.ButtonReply); err != nil {
			return fmt.Errorf("handle button reply: %w", err)
		}

		return nil
	case InteractiveTypeNFMReply:
		if err := handler.FlowCompletionMessageHandler.Handle(ctx, nctx, mctx, message.Interactive.NFMReply); err != nil {
			return err
		}

		return nil
	default:
		if err := handler.InteractiveMessageHandler.Handle(ctx, nctx, mctx, message.Interactive); err != nil {
			return err
		}

		return nil
	}
}

func (handler *Handler) OnTextMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, text *Text) error,
) {
	handler.TextMessageHandler = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetTextMessageHandler(
	h MessageHandler[Text],
) {
	handler.TextMessageHandler = h
}

func (handler *Handler) OnButtonMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, button *Button) error,
) {
	handler.ButtonMessageHandler = MessageHandlerFunc[Button](fn)
}

func (handler *Handler) SetButtonMessageHandler(
	h MessageHandler[Button],
) {
	handler.ButtonMessageHandler = h
}

func (handler *Handler) OnOrderMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, order *Order) error,
) {
	handler.OrderMessageHandler = MessageHandlerFunc[Order](fn)
}

func (handler *Handler) SetOrderMessageHandler(
	h MessageHandler[Order],
) {
	handler.OrderMessageHandler = h
}

func (handler *Handler) OnLocationMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, loc *message.Location) error,
) {
	handler.LocationMessageHandler = MessageHandlerFunc[message.Location](fn)
}

func (handler *Handler) SetLocationMessageHandler(
	h MessageHandler[message.Location],
) {
	handler.LocationMessageHandler = h
}

func (handler *Handler) OnContactsMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, contacts *message.Contacts) error,
) {
	handler.ContactsMessageHandler = MessageHandlerFunc[message.Contacts](fn)
}

func (handler *Handler) SetContactsMessageHandler(
	h MessageHandler[message.Contacts],
) {
	handler.ContactsMessageHandler = h
}

func (handler *Handler) OnReactionMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, reaction *message.Reaction) error,
) {
	handler.ReactionHandler = MessageHandlerFunc[message.Reaction](fn)
}

func (handler *Handler) SetReactionMessageHandler(
	h MessageHandler[message.Reaction],
) {
	handler.ReactionHandler = h
}

func (handler *Handler) OnProductEnquiryMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, txt *Text) error,
) {
	handler.ProductEnquiryHandler = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetProductEnquiryMessageHandler(
	h MessageHandler[Text],
) {
	handler.ProductEnquiryHandler = h
}

func (handler *Handler) OnInteractiveMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, itv *Interactive) error,
) {
	handler.InteractiveMessageHandler = MessageHandlerFunc[Interactive](fn)
}

func (handler *Handler) SetInteractiveMessageHandler(
	h MessageHandler[Interactive],
) {
	handler.InteractiveMessageHandler = h
}

func (handler *Handler) OnButtonReplyMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, br *ButtonReply) error,
) {
	handler.ButtonReplyMessageHandler = MessageHandlerFunc[ButtonReply](fn)
}

func (handler *Handler) SetButtonReplyMessageHandler(
	h MessageHandler[ButtonReply],
) {
	handler.ButtonReplyMessageHandler = h
}

func (handler *Handler) OnListReplyMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, lr *ListReply) error,
) {
	handler.ListReplyMessageHandler = MessageHandlerFunc[ListReply](fn)
}

func (handler *Handler) SetListReplyMessageHandler(
	h MessageHandler[ListReply],
) {
	handler.ListReplyMessageHandler = h
}

func (handler *Handler) OnFlowCompletionMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, nfm *NFMReply) error,
) {
	handler.FlowCompletionMessageHandler = MessageHandlerFunc[NFMReply](fn)
}

func (handler *Handler) SetFlowCompletionMessageHandler(
	h MessageHandler[NFMReply],
) {
	handler.FlowCompletionMessageHandler = h
}

func (handler *Handler) OnReferralMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, ref *ReferralNotification) error,
) {
	handler.ReferralMessageHandler = MessageHandlerFunc[ReferralNotification](fn)
}

func (handler *Handler) SetReferralMessageHandler(
	h MessageHandler[ReferralNotification],
) {
	handler.ReferralMessageHandler = h
}

func (handler *Handler) OnCustomerIDChangeMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, identity *Identity) error,
) {
	handler.CustomerIDChangeHandler = MessageHandlerFunc[Identity](fn)
}

func (handler *Handler) SetCustomerIDChangeMessageHandler(
	h MessageHandler[Identity],
) {
	handler.CustomerIDChangeHandler = h
}

func (handler *Handler) OnSystemMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, sys *System) error,
) {
	handler.SystemMessageHandler = MessageHandlerFunc[System](fn)
}

func (handler *Handler) SetSystemMessageHandler(
	h MessageHandler[System],
) {
	handler.SystemMessageHandler = h
}

func (handler *Handler) OnAudioMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, media *message.MediaInfo) error,
) {
	handler.AudioMessageHandler = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetAudioMessageHandler(
	h MessageHandler[message.MediaInfo],
) {
	handler.AudioMessageHandler = h
}

func (handler *Handler) OnVideoMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, media *message.MediaInfo) error,
) {
	handler.VideoMessageHandler = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetVideoMessageHandler(
	h MessageHandler[message.MediaInfo],
) {
	handler.VideoMessageHandler = h
}

func (handler *Handler) OnImageMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, media *message.MediaInfo) error,
) {
	handler.ImageMessageHandler = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetImageMessageHandler(
	h MessageHandler[message.MediaInfo],
) {
	handler.ImageMessageHandler = h
}

func (handler *Handler) OnDocumentMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, media *message.MediaInfo) error,
) {
	handler.DocumentMessageHandler = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetDocumentMessageHandler(
	h MessageHandler[message.MediaInfo],
) {
	handler.DocumentMessageHandler = h
}

func (handler *Handler) OnStickerMessage(
	fn func(ctx context.Context, nctx *MessageNotificationContext, mctx *MessageInfo, media *message.MediaInfo) error,
) {
	handler.StickerMessageHandler = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetStickerMessageHandler(
	h MessageHandler[message.MediaInfo],
) {
	handler.StickerMessageHandler = h
}

func (handler *Handler) OnUserPreferencesUpdate(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, prefs []*UserPreference) error,
) {
	handler.UserPreferencesUpdateHandler = fn
}

func (handler *Handler) SetUserPreferencesUpdateHandler(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, prefs []*UserPreference) error,
) {
	handler.UserPreferencesUpdateHandler = fn
}
