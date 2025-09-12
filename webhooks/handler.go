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
	"strings"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/message"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
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
		Event                        string               `json:"event,omitempty"`
		MessageTemplateID            int64                `json:"message_template_id,omitempty"`
		MessageTemplateName          string               `json:"message_template_name,omitempty"`
		MessageTemplateLanguage      string               `json:"message_template_language,omitempty"`
		Reason                       *string              `json:"reason,omitempty"`
		PreviousCategory             string               `json:"previous_category,omitempty"`
		PreviousQualityScore         string               `json:"previous_quality_score,omitempty"`
		NewQualityScore              string               `json:"new_quality_score,omitempty"`
		NewCategory                  string               `json:"new_category,omitempty"`
		DisplayPhoneNumber           string               `json:"display_phone_number,omitempty"`
		PhoneNumber                  string               `json:"phone_number,omitempty"`
		CurrentLimit                 string               `json:"current_limit,omitempty"`
		MaxDailyConversationPerPhone int                  `json:"max_daily_conversation_per_phone,omitempty"`
		MaxPhoneNumbersPerBusiness   int                  `json:"max_phone_numbers_per_business,omitempty"`
		MaxPhoneNumbersPerWABA       int                  `json:"max_phone_numbers_per_waba,omitempty"`
		RejectionReason              string               `json:"rejection_reason,omitempty"`
		RequestedVerifiedName        string               `json:"requested_verified_name,omitempty"`
		RestrictionInfo              []RestrictionInfo    `json:"restriction_info,omitempty"`
		BanInfo                      *BanInfo             `json:"ban_info,omitempty"`
		Decision                     string               `json:"decision,omitempty"`
		DisableInfo                  *DisableInfo         `json:"disable_info,omitempty"`
		OtherInfo                    *OtherInfo           `json:"other_info,omitempty"`
		ViolationInfo                *ViolationInfo       `json:"violation_info,omitempty"`
		EntityType                   string               `json:"entity_type,omitempty"`
		EntityID                     string               `json:"entity_id,omitempty"`
		AlertSeverity                string               `json:"alert_severity,omitempty"`
		AlertStatus                  string               `json:"alert_status,omitempty"`
		AlertType                    string               `json:"alert_type,omitempty"`
		AlertDescription             string               `json:"alert_description,omitempty"`
		Message                      string               `json:"message"`                  // Descriptive message of the event
		FlowID                       string               `json:"flow_id"`                  // ID of the flow
		OldStatus                    string               `json:"old_status,omitempty"`     // Previous status of the flow (optional)
		NewStatus                    string               `json:"new_status,omitempty"`     // New status of the flow (optional)
		ErrorRate                    float64              `json:"error_rate,omitempty"`     // Overall error rate for the alert (optional)
		Threshold                    int                  `json:"threshold,omitempty"`      // Alert threshold that was reached or recovered from
		AlertState                   string               `json:"alert_state,omitempty"`    // Status of the alert, e.g., "ACTIVATED" or "DEACTIVATED"
		Errors                       []ErrorInfo          `json:"errors,omitempty"`         // List of errors describing the alert (optional)
		P50Latency                   int                  `json:"p50_latency,omitempty"`    // P50 latency of the endpoint requests (optional)
		P90Latency                   int                  `json:"p90_latency,omitempty"`    // P90 latency of the endpoint requests (optional)
		RequestsCount                int                  `json:"requests_count,omitempty"` // Number of requests used to calculate metric (optional)
		Availability                 int                  `json:"availability"`
		MessagingProduct             string               `json:"messaging_product,omitempty"`
		Metadata                     *Metadata            `json:"metadata,omitempty"`
		Contacts                     []*Contact           `json:"contacts,omitempty"`
		UserPreferences              []*UserPreference    `json:"user_preferences,omitempty"`
		Messages                     []*Message           `json:"messages,omitempty"`
		Statuses                     []*Status            `json:"statuses,omitempty"`
		PhoneNumberSettings          *PhoneNumberSettings `json:"phone_number_settings,omitempty"`
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
		Details    string             `json:"details,omitempty"`
		Href       string             `json:"href,omitempty"`
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
		Details:   info.Details,
		Href:      info.Href,
	}
}

func ErrorInfosAsErrors(infos []ErrorInfo) []*werrors.Error {
	errs := make([]*werrors.Error, len(infos))
	for i, info := range infos {
		errs[i] = info.Error()
	}

	return errs
}

type MediaMessageHandler MessageHandler[message.MediaInfo]

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
	phoneNumberNameUpdate    EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate]
	capabilityUpdate         EventHandler[BusinessNotificationContext, CapabilityUpdate]
	accountUpdate            EventHandler[BusinessNotificationContext, AccountUpdate]
	phoneSettingsUpdate      EventHandler[BusinessNotificationContext, PhoneNumberSettings]
	phoneNumberQualityUpdate EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate]
	accountReviewUpdate      EventHandler[BusinessNotificationContext, AccountReviewUpdate]
	buttonMessage            MessageHandler[Button]
	textMessage              MessageHandler[Text]
	orderMessage             MessageHandler[Order]
	locationMessage          MessageHandler[message.Location]
	contactsMessage          MessageHandler[message.Contacts]
	reactionMessage          MessageHandler[message.Reaction]
	productInquiry           MessageHandler[Text]
	interactiveMessage       MessageHandler[Interactive]
	buttonReplyMessage       MessageHandler[ButtonReply]
	listReplyMessage         MessageHandler[ListReply]
	flowCompletionUpdate     MessageHandler[NFMReply]
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
	userPreferencesUpdate    MessageChangeValueHandler[UserPreference]
	errorMessage             MessageErrorsHandler
	unsupportedMessage       MessageErrorsHandler

	errorHandlerFunc func(ctx context.Context, err error) error
}

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
		phoneNumberNameUpdate:    NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberNameUpdate](),
		capabilityUpdate:         NewNoOpEventHandler[BusinessNotificationContext, CapabilityUpdate](),
		accountUpdate:            NewNoOpEventHandler[BusinessNotificationContext, AccountUpdate](),
		phoneSettingsUpdate:      NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberSettings](),
		phoneNumberQualityUpdate: NewNoOpEventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate](),
		accountReviewUpdate:      NewNoOpEventHandler[BusinessNotificationContext, AccountReviewUpdate](),
		buttonMessage:            NewNoOpMessageHandler[Button](),
		textMessage:              NewNoOpMessageHandler[Text](),
		orderMessage:             NewNoOpMessageHandler[Order](),
		locationMessage:          NewNoOpMessageHandler[message.Location](),
		contactsMessage:          NewNoOpMessageHandler[message.Contacts](),
		reactionMessage:          NewNoOpMessageHandler[message.Reaction](),
		productInquiry:           NewNoOpMessageHandler[Text](),
		interactiveMessage:       NewNoOpMessageHandler[Interactive](),
		buttonReplyMessage:       NewNoOpMessageHandler[ButtonReply](),
		listReplyMessage:         NewNoOpMessageHandler[ListReply](),
		flowCompletionUpdate:     NewNoOpMessageHandler[NFMReply](),
		referralMessage:          NewNoOpMessageHandler[ReferralNotification](),
		customerIDChange:         NewNoOpMessageHandler[Identity](),
		systemMessage:            NewNoOpMessageHandler[System](),
		audioMessage:             NewNoOpMessageHandler[message.MediaInfo](),
		videoMessage:             NewNoOpMessageHandler[message.MediaInfo](),
		imageMessage:             NewNoOpMessageHandler[message.MediaInfo](),
		documentMessage:          NewNoOpMessageHandler[message.MediaInfo](),
		stickerMessage:           NewNoOpMessageHandler[message.MediaInfo](),
		notificationErrors:       NewNoOpMessageChangeValueHandler[werrors.Error](),
		messageStatusChange:      NewNoOpMessageChangeValueHandler[Status](),
		userPreferencesUpdate:    NewNoOpMessageChangeValueHandler[UserPreference](),
		errorMessage:             NewNoOpMessageErrorsHandler(),
		unsupportedMessage:       NewNoOpMessageErrorsHandler(),
		requestWelcome:           NewNoOpMessageHandler[Message](),
		errorHandlerFunc: func(_ context.Context, _ error) error {
			return nil
		},
	}
}

// OnError configures a callback that is invoked whenever an error occurs
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

// HandleNotification processes a single Notification containing one or more
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

func (handler *Handler) handleNotificationChange( //nolint: gocognit,gocyclo, cyclop, funlen // ok
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
		if err := handler.handleTemplateStatusUpdate(ctx, notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateCategoryUpdate.String():
		if err := handler.handleTemplateCategoryUpdate(ctx, notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldTemplateQualityUpdate.String():
		if err := handler.handleTemplateQualityUpdate(ctx, notification, change, entry); err != nil {
			return err
		}

	case ChangeFieldPhoneNumberNameUpdate.String():
		if handler.phoneNumberNameUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.phoneNumberNameUpdate.HandleEvent(ctx, notificationCtx, change.Value.PhoneNumberNameUpdate()); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldPhoneNumberQualityUpdate.String():
		if handler.phoneNumberQualityUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.phoneNumberQualityUpdate.HandleEvent(ctx, notificationCtx, change.Value.PhoneNumberQualityUpdate()); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldAccountUpdate.String():
		if handler.accountUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.accountUpdate.HandleEvent(ctx, notificationCtx, change.Value.AccountUpdate()); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldAccountReviewUpdate.String():
		if handler.accountReviewUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.accountReviewUpdate.HandleEvent(ctx, notificationCtx, change.Value.AccountReviewUpdate()); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldBusinessCapabilityUpdate.String():
		if handler.capabilityUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.capabilityUpdate.HandleEvent(ctx, notificationCtx, change.Value.CapabilityUpdate()); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldUserPreferences.String():
		if handler.userPreferencesUpdate != nil {
			notificationCtx := &MessageNotificationContext{
				EntryID:          entry.ID,
				MessagingProduct: change.Value.MessagingProduct,
				Contacts:         change.Value.Contacts,
				Metadata:         change.Value.Metadata,
			}
			if err := handler.userPreferencesUpdate.Handle(ctx, notificationCtx, change.Value.UserPreferences); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldAccountSettingsUpdate.String():
		if handler.accountUpdate != nil {
			notificationCtx := &BusinessNotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}
			if err := handler.phoneSettingsUpdate.HandleEvent(ctx, notificationCtx, change.Value.PhoneNumberSettings); err != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}

	case ChangeFieldMessages.String():
		return handler.handleNotificationMessageItem(ctx, entry, change)
	}

	return nil
}

func (handler *Handler) handleTemplateQualityUpdate(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	if handler.templateQuality != nil {
		notificationCtx := &BusinessNotificationContext{
			Object:      notification.Object,
			EntryID:     entry.ID,
			EntryTime:   entry.Time,
			ChangeField: change.Field,
		}

		if err := handler.templateQuality.HandleEvent(ctx, notificationCtx, change.Value.TemplateQualityUpdate()); err != nil {
			if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
				return handlerErr
			}
		}
	}
	return nil
}

func (handler *Handler) handleTemplateCategoryUpdate(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	if handler.templateCategory != nil {
		notificationCtx := &BusinessNotificationContext{
			Object:      notification.Object,
			EntryID:     entry.ID,
			EntryTime:   entry.Time,
			ChangeField: change.Field,
		}

		if err := handler.templateCategory.HandleEvent(ctx, notificationCtx, change.Value.TemplateCategoryUpdate()); err != nil {
			if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
				return handlerErr
			}
		}
	}
	return nil
}

func (handler *Handler) handleAccountAlerts(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	if handler.alerts != nil {
		notificationCtx := &BusinessNotificationContext{
			Object:      notification.Object,
			EntryID:     entry.ID,
			EntryTime:   entry.Time,
			ChangeField: change.Field,
		}

		if err := handler.alerts.HandleEvent(ctx, notificationCtx, change.Value.AlertNotification()); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}
	return nil
}

func (handler *Handler) handleTemplateStatusUpdate(
	ctx context.Context,
	notification *Notification,
	change Change,
	entry Entry,
) error {
	if handler.templateStatus != nil {
		notificationCtx := &BusinessNotificationContext{
			Object:      notification.Object,
			EntryID:     entry.ID,
			EntryTime:   entry.Time,
			ChangeField: change.Field,
		}

		if err := handler.templateStatus.HandleEvent(ctx, notificationCtx, change.Value.TemplateStatusUpdate()); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}
	return nil
}

func (handler *Handler) handleNotificationMessageItem( //nolint: gocognit // ok
	ctx context.Context,
	entry Entry,
	change Change,
) error {
	notificationCtx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Contacts:         change.Value.Contacts,
		Metadata:         change.Value.Metadata,
	}

	// handle notification errors do not terminate of its success or if the error is not fatal
	if handler.notificationErrors != nil {
		if err := handler.notificationErrors.Handle(
			ctx, notificationCtx, ErrorInfosAsErrors(change.Value.Errors),
		); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}

	if handler.messageStatusChange != nil {
		if err := handler.messageStatusChange.Handle(
			ctx, notificationCtx, change.Value.Statuses,
		); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}

	for _, m := range change.Value.Messages {
		if err := handler.handleNotificationMessage(ctx, notificationCtx, m); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}

	return nil
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
	ChangeFieldAccountSettingsUpdate    ChangeField = "account_settings_update"
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

	FlowStatusHandler               EventHandler[FlowNotificationContext, StatusChangeDetails]
	FlowClientErrorRateHandler      EventHandler[FlowNotificationContext, ClientErrorRateDetails]
	FlowEndpointErrorRateHandler    EventHandler[FlowNotificationContext, EndpointErrorRateDetails]
	FlowEndpointLatencyHandler      EventHandler[FlowNotificationContext, EndpointLatencyDetails]
	FlowEndpointAvailabilityHandler EventHandler[FlowNotificationContext, EndpointAvailabilityDetails]
	AlertsHandler                   EventHandler[BusinessNotificationContext, AlertNotification]
	TemplateStatusHandler           EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification]
	TemplateCategoryHandler         EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification]
	TemplateQualityHandler          EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification]
	PhoneNumberNameUpdateHandler    EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate]
	CapabilityUpdateHandler         EventHandler[BusinessNotificationContext, CapabilityUpdate]
	AccountUpdateHandler            EventHandler[BusinessNotificationContext, AccountUpdate]
	PhoneNumberQualityUpdateHandler EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate]
	AccountReviewUpdateHandler      EventHandler[BusinessNotificationContext, AccountReviewUpdate]
)

func (f EventHandlerFunc[S, T]) HandleEvent(ctx context.Context, ntx *S, notification *T) error {
	return f(ctx, ntx, notification)
}

func NewNoOpEventHandler[S any, T any]() EventHandler[S, T] {
	return EventHandlerFunc[S, T](func(_ context.Context, _ *S, _ *T) error {
		return nil
	})
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

	PhoneNumberSettings struct {
		PhoneNumberID string                      `json:"phone_number_id,omitempty"`
		Callings      *PhoneNumberSettingsCalling `json:"callings,omitempty"`
	}

	PhoneNumberSettingsCalling struct {
		Status                   string     `json:"status,omitempty"`
		CallIconVisibility       string     `json:"call_icon_visibility,omitempty"`
		CallbackPermissionStatus string     `json:"callback_permission_status,omitempty"`
		CallHours                *CallHours `json:"call_hours,omitempty"`
		SIP                      *SIP       `json:"sip,omitempty"`
	}

	CallHours struct {
		Status               string               `json:"status,omitempty"`
		TimezoneID           string               `json:"timezone_id,omitempty"` // e.g. "Europe/Berlin" or provider’s TZ id
		WeeklyOperatingHours []WeeklyOperatingDay `json:"weekly_operating_hours,omitempty"`
		HolidaySchedule      []Holiday            `json:"holiday_schedule,omitempty"`
	}

	WeeklyOperatingDay struct {
		DayOfWeek string `json:"day_of_week,omitempty"` // "MONDAY", ...
		OpenTime  string `json:"open_time,omitempty"`   // "HHMM" e.g., "0400"
		CloseTime string `json:"close_time,omitempty"`  // "HHMM" e.g., "1020"
	}

	Holiday struct {
		Date      string `json:"date,omitempty"`       // "YYYY-MM-DD"
		StartTime string `json:"start_time,omitempty"` // "HHMM"
		EndTime   string `json:"end_time,omitempty"`   // "HHMM"
	}

	SIP struct {
		Status  string      `json:"status,omitempty"`
		Servers []SIPServer `json:"servers,omitempty"`
	}

	SIPServer struct {
		Hostname        string `json:"hostname,omitempty"`
		SIPUserPassword string `json:"sip_user_password,omitempty"`
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
	handler.alerts = EventHandlerFunc[BusinessNotificationContext, AlertNotification](fn)
}

func (handler *Handler) SetBusinessAlertNotificationHandler(
	fn EventHandler[BusinessNotificationContext, AlertNotification],
) {
	handler.alerts = fn
}

func (handler *Handler) OnBusinessTemplateStatusUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateStatusUpdateNotification) error,
) {
	handler.templateStatus = EventHandlerFunc[BusinessNotificationContext, TemplateStatusUpdateNotification](fn)
}

func (handler *Handler) SetBusinessTemplateStatusUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification],
) {
	handler.templateStatus = fn
}

func (handler *Handler) OnBusinessTemplateCategoryUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateCategoryUpdateNotification) error,
) {
	handler.templateCategory = EventHandlerFunc[BusinessNotificationContext, TemplateCategoryUpdateNotification](
		fn,
	)
}

func (handler *Handler) SetBusinessTemplateCategoryUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification],
) {
	handler.templateCategory = fn
}

func (handler *Handler) OnBusinessTemplateQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateQualityUpdateNotification) error,
) {
	handler.templateQuality = EventHandlerFunc[BusinessNotificationContext, TemplateQualityUpdateNotification](
		fn,
	)
}

func (handler *Handler) SetBusinessTemplateQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification],
) {
	handler.templateQuality = fn
}

func (handler *Handler) OnBusinessPhoneNumberNameUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberNameUpdate) error,
) {
	handler.phoneNumberNameUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberNameUpdate](fn)
}

func (handler *Handler) SetBusinessPhoneNumberNameUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate],
) {
	handler.phoneNumberNameUpdate = fn
}

func (handler *Handler) OnBusinessPhoneNumberQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberQualityUpdate) error,
) {
	handler.phoneNumberQualityUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberQualityUpdate](
		fn,
	)
}

func (handler *Handler) SetBusinessPhoneNumberQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate],
) {
	handler.phoneNumberQualityUpdate = fn
}

func (handler *Handler) OnBusinessAccountReviewUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountReviewUpdate) error,
) {
	handler.accountReviewUpdate = EventHandlerFunc[BusinessNotificationContext, AccountReviewUpdate](fn)
}

func (handler *Handler) SetBusinessAccountReviewUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountReviewUpdate],
) {
	handler.accountReviewUpdate = fn
}

func (handler *Handler) OnBusinessAccountUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountUpdate) error,
) {
	handler.accountUpdate = EventHandlerFunc[BusinessNotificationContext, AccountUpdate](fn)
}

func (handler *Handler) OnPhoneSettingsUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberSettings) error,
) {
	handler.phoneSettingsUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberSettings](fn)
}

func (handler *Handler) SetBusinessAccountUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountUpdate],
) {
	handler.accountUpdate = fn
}

func (handler *Handler) OnBusinessCapabilityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *CapabilityUpdate) error,
) {
	handler.capabilityUpdate = EventHandlerFunc[BusinessNotificationContext, CapabilityUpdate](fn)
}

func (handler *Handler) SetBusinessCapabilityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, CapabilityUpdate],
) {
	handler.capabilityUpdate = fn
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
	switch value.Event {
	case EventFlowStatusChange:
		details := value.FlowStatusChange()
		if err := handler.flowStatus.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle flow status change event: %w", err)
		}

	case EventClientErrorRate:
		details := value.FlowClientErrorRate()
		if err := handler.flowClientErrorRate.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle client error rate event: %w", err)
		}

	case EventEndpointErrorRate:
		details := value.FlowEndpointErrorRate()
		if err := handler.flowEndpointErrorRate.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint error rate event: %w", err)
		}

	case EventEndpointLatency:
		details := value.FlowEndpointLatency()
		if err := handler.flowEndpointLatency.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint latency event: %w", err)
		}

	case EventEndpointAvailability:
		details := value.FlowEndpointAvailability()
		if err := handler.flowEndpointAvailability.HandleEvent(ctx, notificationContext, details); err != nil {
			return fmt.Errorf("handle endpoint availability event: %w", err)
		}
	}

	return nil
}

func (handler *Handler) OnFlowStatusChange(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *StatusChangeDetails) error,
) {
	handler.flowStatus = EventHandlerFunc[FlowNotificationContext, StatusChangeDetails](fn)
}

func (handler *Handler) SetFlowStatusChangeHandler(fn FlowStatusHandler) {
	handler.flowStatus = fn
}

func (handler *Handler) OnFlowClientErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *ClientErrorRateDetails) error,
) {
	handler.flowClientErrorRate = EventHandlerFunc[FlowNotificationContext, ClientErrorRateDetails](fn)
}

func (handler *Handler) SetFlowClientErrorRateHandler(
	fn FlowClientErrorRateHandler,
) {
	handler.flowClientErrorRate = fn
}

func (handler *Handler) OnFlowEndpointErrorRate(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointErrorRateDetails) error,
) {
	handler.flowEndpointErrorRate = EventHandlerFunc[FlowNotificationContext, EndpointErrorRateDetails](fn)
}

func (handler *Handler) SetFlowEndpointErrorRateHandler(
	fn FlowEndpointErrorRateHandler,
) {
	handler.flowEndpointErrorRate = fn
}

func (handler *Handler) OnFlowEndpointLatency(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointLatencyDetails) error,
) {
	handler.flowEndpointLatency = EventHandlerFunc[FlowNotificationContext, EndpointLatencyDetails](fn)
}

func (handler *Handler) SetFlowEndpointLatencyHandler(
	fn FlowEndpointLatencyHandler,
) {
	handler.flowEndpointLatency = fn
}

func (handler *Handler) OnFlowEndpointAvailability(
	fn func(ctx context.Context, notificationContext *FlowNotificationContext, details *EndpointAvailabilityDetails) error,
) {
	handler.flowEndpointAvailability = EventHandlerFunc[FlowNotificationContext, EndpointAvailabilityDetails](fn)
}

func (handler *Handler) SetFlowEndpointAvailabilityHandler(
	fn FlowEndpointAvailabilityHandler,
) {
	handler.flowEndpointAvailability = fn
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
	MessageTypeUnsupported    MessageType = "unsupported"
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
		"unsupported":     MessageTypeUnsupported,
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

func (c *Contact) SenderInfo() *SenderInfo {
	return &SenderInfo{
		Name: c.Profile.Name,
		WaID: c.WaID,
	}
}

type (
	Contact struct {
		Profile *Profile `json:"profile,omitempty"`
		WaID    string   `json:"wa_id,omitempty"`
	}

	SenderInfo struct {
		Name string `json:"name,omitempty"`
		WaID string `json:"wa_id,omitempty"`
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
	// these are common fields to all type of messages.
	// From The customer's phone number who sent the message to the business.
	// MessageID The MessageID for the message that was received by the business. You could use messages
	// endpoint to mark this specific message as read.
	// Timestamp The timestamp for when the message was received by the business.
	// Type The type of message that was received by the business.
	// Context The context of the message. Only included when a user replies or interacts with one
	// of your messages.
	MessageInfo struct {
		From             string
		MessageID        string
		Timestamp        string
		Type             string
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

type (
	MessageHandlerFunc[T any] func(
		ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error

	MessageHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, message *T) error
	}

	ButtonMessageHandler         = MessageHandler[Button]
	TextMessageHandler           = MessageHandler[Text]
	OrderMessageHandler          = MessageHandler[Order]
	LocationMessageHandler       = MessageHandler[message.Location]
	ContactsMessageHandler       = MessageHandler[message.Contacts]
	ReactionHandler              = MessageHandler[message.Reaction]
	ProductEnquiryHandler        = MessageHandler[Text]
	InteractiveMessageHandler    = MessageHandler[Interactive]
	ButtonReplyMessageHandler    = MessageHandler[ButtonReply]
	ListReplyMessageHandler      = MessageHandler[ListReply]
	FlowCompletionMessageHandler = MessageHandler[NFMReply]
	ReferralMessageHandler       = MessageHandler[ReferralNotification]
	CustomerIDChangeHandler      = MessageHandler[Identity]
	SystemMessageHandler         = MessageHandler[System]
	RequestWelcomeMessageHandler = MessageHandler[Message]
)

func (fn MessageHandlerFunc[T]) Handle(ctx context.Context, notificationCtx *MessageNotificationContext,
	info *MessageInfo, message *T,
) error {
	return fn(ctx, notificationCtx, info, message)
}

func NewNoOpMessageHandler[T any]() MessageHandler[T] {
	return MessageHandlerFunc[T](func(_ context.Context, _ *MessageNotificationContext, _ *MessageInfo, _ *T) error {
		return nil
	})
}

type (
	MessageChangeValueHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
	}

	MessageChangeValueHandlerFunc[T any] func(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
)

type (
	UserPreferenceUpdateHandler = MessageChangeValueHandler[UserPreference]
	NotificationErrorsHandler   = MessageChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler  = MessageChangeValueHandler[Status]
)

func (f MessageChangeValueHandlerFunc[T]) Handle(
	ctx context.Context,
	notificationCtx *MessageNotificationContext,
	values []*T,
) error {
	return f(ctx, notificationCtx, values)
}

func NewNoOpMessageChangeValueHandler[T any]() MessageChangeValueHandler[T] {
	return MessageChangeValueHandlerFunc[T](func(_ context.Context, _ *MessageNotificationContext, _ []*T) error {
		return nil
	})
}

const (
	InteractiveTypeListReply   = "list_reply"
	InteractiveTypeButtonReply = "button_reply"
	InteractiveTypeNFMReply    = "nfm_reply"
)

type (
	MessageErrorsHandlerFunc func(
		ctx context.Context, notificationContext *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error
	MessageErrorsHandler interface {
		Handle(
			ctx context.Context,
			notificationContext *MessageNotificationContext,
			info *MessageInfo,
			errors []*werrors.Error,
		) error
	}
)

func (fn MessageErrorsHandlerFunc) Handle(ctx context.Context, notificationCtx *MessageNotificationContext,
	info *MessageInfo, errors []*werrors.Error,
) error {
	return fn(ctx, notificationCtx, info, errors)
}

func NewNoOpMessageErrorsHandler() MessageErrorsHandler {
	return MessageErrorsHandlerFunc(
		func(_ context.Context, _ *MessageNotificationContext, _ *MessageInfo, _ []*werrors.Error) error {
			return nil
		},
	)
}

func (handler *Handler) handleNotificationMessage(ctx context.Context,
	notificationCtx *MessageNotificationContext, message *Message,
) error {
	info := &MessageInfo{
		From:             message.From,
		MessageID:        message.ID,
		Timestamp:        message.Timestamp,
		Type:             message.Type,
		Context:          message.Context,
		IsAReply:         message.IsAReply(),
		IsForwarded:      message.IsForwarded(),
		IsProductInquiry: message.IsProductInquiry(),
		IsReferral:       message.IsReferral(),
	}

	messageType := ParseMessageType(message.Type)
	switch messageType {
	case MessageTypeOrder:
		if err := handler.orderMessage.Handle(ctx, notificationCtx, info, message.Order); err != nil {
			return err
		}

		return nil

	case MessageTypeButton:
		if err := handler.buttonMessage.Handle(ctx, notificationCtx, info, message.Button); err != nil {
			return err
		}

		return nil

	case MessageTypeAudio, MessageTypeVideo, MessageTypeImage, MessageTypeDocument, MessageTypeSticker:
		return handler.handleMediaMessage(ctx, notificationCtx, message, info)

	case MessageTypeInteractive:
		return handler.handleInteractiveNotification(ctx, notificationCtx, message, info)

	case MessageTypeSystem:
		if err := handler.systemMessage.Handle(ctx, notificationCtx, info, message.System); err != nil {
			return err
		}

		return nil

	case MessageTypeUnknown:
		if err := handler.errorMessage.Handle(ctx, notificationCtx, info, message.Errors); err != nil {
			return err
		}

		return nil

	case MessageTypeUnsupported:
		return handler.unsupportedMessage.Handle(ctx, notificationCtx, info, message.Errors)

	case MessageTypeText:
		return handler.handleTextNotification(ctx, notificationCtx, message, info)

	case MessageTypeRequestWelcome:
		if handler.requestWelcome != nil {
			if err := handler.requestWelcome.Handle(ctx, notificationCtx, info, message); err != nil {
				return err
			}
		}

		return nil

	case MessageTypeReaction:
		if err := handler.reactionMessage.Handle(ctx, notificationCtx, info, message.Reaction); err != nil {
			return err
		}

		return nil

	case MessageTypeLocation:
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return err
		}

		return nil

	case MessageTypeContacts:
		if err := handler.contactsMessage.Handle(ctx, notificationCtx, info, message.Contacts); err != nil {
			return err
		}

		return nil

	default:
		return handler.handleDefaultNotificationMessage(ctx, notificationCtx, message, info)
	}
}

func (handler *Handler) handleMediaMessage(ctx context.Context, notificationCtx *MessageNotificationContext,
	message *Message, info *MessageInfo,
) error {
	messageType := ParseMessageType(message.Type)
	switch messageType { //nolint:exhaustive // ok
	case MessageTypeAudio:
		if err := handler.audioMessage.Handle(ctx, notificationCtx, info, message.Audio); err != nil {
			return err
		}

		return nil

	case MessageTypeVideo:
		if err := handler.videoMessage.Handle(ctx, notificationCtx, info, message.Video); err != nil {
			return err
		}

		return nil

	case MessageTypeImage:
		if err := handler.imageMessage.Handle(ctx, notificationCtx, info, message.Image); err != nil {
			return err
		}

		return nil

	case MessageTypeDocument:
		if err := handler.documentMessage.Handle(ctx, notificationCtx, info, message.Document); err != nil {
			return err
		}

		return nil

	case MessageTypeSticker:
		if err := handler.stickerMessage.Handle(ctx, notificationCtx, info, message.Sticker); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (handler *Handler) handleTextNotification(ctx context.Context, notificationCtx *MessageNotificationContext,
	message *Message, info *MessageInfo,
) error {
	if info.IsReferral {
		referral := &ReferralNotification{
			Text:     message.Text,
			Referral: message.Referral,
		}

		if err := handler.referralMessage.Handle(ctx, notificationCtx, info, referral); err != nil {
			return err
		}

		return nil
	}

	if info.IsProductInquiry {
		if err := handler.productInquiry.Handle(ctx, notificationCtx, info, message.Text); err != nil {
			return err
		}

		return nil
	}

	if err := handler.textMessage.Handle(ctx, notificationCtx, info, message.Text); err != nil {
		return err
	}

	return nil
}

func (handler *Handler) handleDefaultNotificationMessage(
	ctx context.Context,
	notificationCtx *MessageNotificationContext,
	message *Message,
	info *MessageInfo,
) error {
	if message.Contacts != nil {
		if err := handler.contactsMessage.Handle(ctx, notificationCtx, info, message.Contacts); err != nil {
			return err
		}

		return nil
	}
	if message.Location != nil {
		if err := handler.locationMessage.Handle(ctx, notificationCtx, info, message.Location); err != nil {
			return err
		}

		return nil
	}

	if message.Identity != nil {
		if err := handler.customerIDChange.Handle(ctx, notificationCtx, info, message.Identity); err != nil {
			return err
		}

		return nil
	}

	return ErrUnrecognizedMessageType
}

const ErrUnrecognizedMessageType = whatsapp.Error("unrecognized message type")

func (handler *Handler) handleInteractiveNotification(ctx context.Context,
	notificationCtx *MessageNotificationContext, message *Message, info *MessageInfo,
) error {
	switch message.Interactive.Type {
	case InteractiveTypeListReply:
		if err := handler.listReplyMessage.Handle(ctx, notificationCtx, info, message.Interactive.ListReply); err != nil {
			return fmt.Errorf("handle list reply: %w", err)
		}

		return nil
	case InteractiveTypeButtonReply:
		if err := handler.buttonReplyMessage.Handle(ctx, notificationCtx, info, message.Interactive.ButtonReply); err != nil {
			return fmt.Errorf("handle button reply: %w", err)
		}

		return nil
	case InteractiveTypeNFMReply:
		if err := handler.flowCompletionUpdate.Handle(ctx, notificationCtx, info, message.Interactive.NFMReply); err != nil {
			return err
		}

		return nil
	default:
		if err := handler.interactiveMessage.Handle(ctx, notificationCtx, info, message.Interactive); err != nil {
			return err
		}

		return nil
	}
}

func (handler *Handler) OnTextMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, text *Text) error,
) {
	handler.textMessage = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetTextMessageHandler(
	h TextMessageHandler,
) {
	handler.textMessage = h
}

func (handler *Handler) OnButtonMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, button *Button) error,
) {
	handler.buttonMessage = MessageHandlerFunc[Button](fn)
}

func (handler *Handler) SetButtonMessageHandler(
	h ButtonMessageHandler,
) {
	handler.buttonMessage = h
}

func (handler *Handler) OnOrderMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, order *Order) error,
) {
	handler.orderMessage = MessageHandlerFunc[Order](fn)
}

func (handler *Handler) SetOrderMessageHandler(
	h OrderMessageHandler,
) {
	handler.orderMessage = h
}

func (handler *Handler) OnLocationMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, loc *message.Location) error,
) {
	handler.locationMessage = MessageHandlerFunc[message.Location](fn)
}

func (handler *Handler) SetLocationMessageHandler(
	h LocationMessageHandler,
) {
	handler.locationMessage = h
}

func (handler *Handler) OnContactsMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, contacts *message.Contacts) error,
) {
	handler.contactsMessage = MessageHandlerFunc[message.Contacts](fn)
}

func (handler *Handler) SetContactsMessageHandler(
	h ContactsMessageHandler,
) {
	handler.contactsMessage = h
}

func (handler *Handler) OnReactionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, reaction *message.Reaction) error,
) {
	handler.reactionMessage = MessageHandlerFunc[message.Reaction](fn)
}

func (handler *Handler) SetReactionMessageHandler(
	h ReactionHandler,
) {
	handler.reactionMessage = h
}

func (handler *Handler) OnProductEnquiryMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, txt *Text) error,
) {
	handler.productInquiry = MessageHandlerFunc[Text](fn)
}

func (handler *Handler) SetProductEnquiryMessageHandler(
	h ProductEnquiryHandler,
) {
	handler.productInquiry = h
}

func (handler *Handler) OnInteractiveMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, itv *Interactive) error,
) {
	handler.interactiveMessage = MessageHandlerFunc[Interactive](fn)
}

func (handler *Handler) SetInteractiveMessageHandler(
	h InteractiveMessageHandler,
) {
	handler.interactiveMessage = h
}

func (handler *Handler) OnButtonReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, br *ButtonReply) error,
) {
	handler.buttonReplyMessage = MessageHandlerFunc[ButtonReply](fn)
}

func (handler *Handler) SetButtonReplyMessageHandler(
	h ButtonReplyMessageHandler,
) {
	handler.buttonReplyMessage = h
}

func (handler *Handler) OnListReplyMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, lr *ListReply) error,
) {
	handler.listReplyMessage = MessageHandlerFunc[ListReply](fn)
}

func (handler *Handler) SetListReplyMessageHandler(
	h ListReplyMessageHandler,
) {
	handler.listReplyMessage = h
}

func (handler *Handler) OnFlowCompletionMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, nfm *NFMReply) error,
) {
	handler.flowCompletionUpdate = MessageHandlerFunc[NFMReply](fn)
}

func (handler *Handler) SetFlowCompletionMessageHandler(
	h FlowCompletionMessageHandler,
) {
	handler.flowCompletionUpdate = h
}

func (handler *Handler) OnReferralMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, ref *ReferralNotification) error,
) {
	handler.referralMessage = MessageHandlerFunc[ReferralNotification](fn)
}

func (handler *Handler) SetReferralMessageHandler(
	h ReferralMessageHandler,
) {
	handler.referralMessage = h
}

func (handler *Handler) OnCustomerIDChangeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, identity *Identity) error,
) {
	handler.customerIDChange = MessageHandlerFunc[Identity](fn)
}

func (handler *Handler) SetCustomerIDChangeMessageHandler(
	h CustomerIDChangeHandler,
) {
	handler.customerIDChange = h
}

func (handler *Handler) OnSystemMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, sys *System) error,
) {
	handler.systemMessage = MessageHandlerFunc[System](fn)
}

func (handler *Handler) SetSystemMessageHandler(
	h SystemMessageHandler,
) {
	handler.systemMessage = h
}

func (handler *Handler) OnAudioMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.audioMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetAudioMessageHandler(
	h MediaMessageHandler,
) {
	handler.audioMessage = h
}

func (handler *Handler) OnVideoMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.videoMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetVideoMessageHandler(
	h MediaMessageHandler,
) {
	handler.videoMessage = h
}

func (handler *Handler) OnImageMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.imageMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetImageMessageHandler(
	h MediaMessageHandler,
) {
	handler.imageMessage = h
}

func (handler *Handler) OnDocumentMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.documentMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetDocumentMessageHandler(
	h MediaMessageHandler,
) {
	handler.documentMessage = h
}

func (handler *Handler) OnStickerMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, media *message.MediaInfo) error,
) {
	handler.stickerMessage = MessageHandlerFunc[message.MediaInfo](fn)
}

func (handler *Handler) SetStickerMessageHandler(
	h MediaMessageHandler,
) {
	handler.stickerMessage = h
}

func (handler *Handler) OnUserPreferencesUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, prefs []*UserPreference) error,
) {
	handler.userPreferencesUpdate = MessageChangeValueHandlerFunc[UserPreference](fn)
}

func (handler *Handler) SetUserPreferencesUpdateHandler(
	h UserPreferenceUpdateHandler,
) {
	handler.userPreferencesUpdate = h
}

func (handler *Handler) OnRequestWelcomeMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext,
		info *MessageInfo, media *Message) error,
) {
	handler.requestWelcome = MessageHandlerFunc[Message](fn)
}

func (handler *Handler) SetRequestWelcomeMessageHandler(
	fn RequestWelcomeMessageHandler,
) {
	handler.requestWelcome = fn
}

func (handler *Handler) OnNotificationErrors(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, errors []*werrors.Error) error,
) {
	handler.notificationErrors = MessageChangeValueHandlerFunc[werrors.Error](fn)
}

func (handler *Handler) SetNotificationErrorsHandler(
	fn NotificationErrorsHandler,
) {
	handler.notificationErrors = fn
}

func (handler *Handler) OnMessageStatusChange(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, status []*Status) error,
) {
	handler.messageStatusChange = MessageChangeValueHandlerFunc[Status](fn)
}

func (handler *Handler) SetMessageStatusChangeHandler(
	fn MessageStatusChangeHandler,
) {
	handler.messageStatusChange = fn
}

func (handler *Handler) OnMessageErrors(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error,
) {
	handler.errorMessage = MessageErrorsHandlerFunc(fn)
}

func (handler *Handler) SetMessageErrorsHandler(
	fn MessageErrorsHandler,
) {
	handler.errorMessage = fn
}

func (handler *Handler) OnUnsupportedMessage(
	fn func(ctx context.Context, notificationCtx *MessageNotificationContext, info *MessageInfo, errors []*werrors.Error) error,
) {
	handler.unsupportedMessage = MessageErrorsHandlerFunc(fn)
}

func (handler *Handler) SetUnsupportedMessageHandler(
	fn MessageErrorsHandler,
) {
	handler.unsupportedMessage = fn
}
