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
	"net/http"

	"github.com/piusalfred/whatsapp/message"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

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
	callStatusUpdate         EventHandler[BusinessNotificationContext, CallStatusUpdate]
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
	userPreferencesUpdate    MessageChangeValueHandler[UserPreference]
	groupLifecycleUpdate     MessageChangeValueHandler[Group]
	groupParticipantsUpdate  MessageChangeValueHandler[Group]
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
		callStatusUpdate:         NewNoOpEventHandler[BusinessNotificationContext, CallStatusUpdate](),
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
		addressSubmission:        NewNoOpMessageHandler[NFMReply](),
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
		groupLifecycleUpdate:     NewNoOpMessageChangeValueHandler[Group](),
		groupParticipantsUpdate:  NewNoOpMessageChangeValueHandler[Group](),
		errorMessage:             NewNoOpMessageErrorsHandler(),
		unsupportedMessage:       NewNoOpMessageErrorsHandler(),
		requestWelcome:           NewNoOpMessageHandler[Message](),
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

	case ChangeFieldAccountSettingsUpdate.String():
		return handleBusinessNotification(
			ctx, handler, notification, change, entry,
			handler.phoneSettingsUpdate, change.Value.PhoneNumberSettings,
		)

	case ChangeFieldMessages.String():
		return handler.handleNotificationMessageItem(ctx, entry, change)

	case ChangeFieldGroupLifecycleUpdate.String(),
		ChangeFieldGroupParticipantsUpdate.String():
		return handler.handleGroupWebhooks(ctx, change, entry)
	}

	return nil
}

func handleBusinessNotification[T any](
	ctx context.Context,
	handler *Handler,
	notification *Notification,
	change Change,
	entry Entry,
	eventHandler EventHandler[BusinessNotificationContext, T],
	eventData *T,
) error {
	if eventHandler == nil {
		return nil
	}

	notificationCtx := &BusinessNotificationContext{
		Object:      notification.Object,
		EntryID:     entry.ID,
		EntryTime:   entry.Time,
		ChangeField: change.Field,
	}

	if err := eventHandler.HandleEvent(ctx, notificationCtx, eventData); err != nil {
		if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
			return handlerErr
		}
	}

	return nil
}

func (handler *Handler) handleGroupWebhooks(ctx context.Context, change Change, entry Entry) error {
	switch change.Field {
	case ChangeFieldGroupLifecycleUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupLifecycleUpdate, change, entry, change.Value.Groups,
		)
	case ChangeFieldGroupParticipantsUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupParticipantsUpdate, change, entry, change.Value.Groups,
		)
	}
	return nil
}

func handleMessageChangeNotification[T any](
	ctx context.Context,
	handler *Handler,
	eventHandler MessageChangeValueHandler[T],
	change Change,
	entry Entry,
	events []*T,
) error {
	if eventHandler == nil {
		return nil
	}

	notificationCtx := &MessageNotificationContext{
		EntryID:          entry.ID,
		MessagingProduct: change.Value.MessagingProduct,
		Contacts:         change.Value.Contacts,
		Metadata:         change.Value.Metadata,
	}

	if err := eventHandler.Handle(ctx, notificationCtx, events); err != nil {
		if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
			return handlerErr
		}
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

	// handle notification errors do not terminate of its success, or if the error is not fatal
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
	ChangeFieldCalls                    ChangeField = "calls"
	ChangeFieldGroupLifecycleUpdate     ChangeField = "group_lifecycle_update"
	ChangeFieldGroupParticipantsUpdate  ChangeField = "group_participants_update"
	ChangeFieldGroupSettingsUpdate      ChangeField = "group_settings_update"
	ChangeFieldGroupStatusUpdate        ChangeField = "group_status_update"
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

func NewNoOpEventHandler[S any, T any]() EventHandler[S, T] {
	return EventHandlerFunc[S, T](func(_ context.Context, _ *S, _ *T) error {
		return nil
	})
}

const (
	WABABanStateDisable            WABABanState = "DISABLE"
	WABABanStateReinstate          WABABanState = "REINSTATE"
	WABABanStateScheduleForDisable WABABanState = "SCHEDULE_FOR_DISABLE"
)

type (
	WABABanState string

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
		GroupID     string             `json:"group_id,omitempty"`
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
		ID                     string             `json:"id,omitempty"`
		RecipientID            string             `json:"recipient_id,omitempty"`
		RecipientType          string             `json:"recipient_type,omitempty"`
		RecipientParticipantID string             `json:"recipient_participant_id,omitempty"`
		ParticipantRecipientID string             `json:"participant_recipient_id,omitempty"`
		StatusValue            string             `json:"status,omitempty"`
		Timestamp              string             `json:"timestamp,omitempty"`
		Conversation           *Conversation      `json:"conversation,omitempty"`
		Pricing                *Pricing           `json:"pricing,omitempty"`
		Errors                 []*werrors.Error   `json:"errors,omitempty"`
		BizOpaqueCallbackData  string             `json:"biz_opaque_callback_data,omitempty"`
		Message                *StatusMessageInfo `json:"message,omitempty"`
		Type                   string             `json:"type,omitempty"`
	}

	StatusMessageInfo struct {
		RecipientID string `json:"recipient_id,omitempty"`
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
	MessageChangeValueHandler[T any] interface {
		Handle(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
	}
	MessageChangeValueHandlerFunc[T any] func(ctx context.Context, notificationCtx *MessageNotificationContext, value []*T) error
)

type (
	UserPreferenceUpdateHandler    = MessageChangeValueHandler[UserPreference]
	NotificationErrorsHandler      = MessageChangeValueHandler[werrors.Error]
	MessageStatusChangeHandler     = MessageChangeValueHandler[Status]
	GroupLifecycleUpdateHandler    = MessageChangeValueHandler[Group]
	GroupParticipantsUpdateHandler = MessageChangeValueHandler[Group]
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
