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

package webhooks

import (
	"context"
)

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

		if err := handler.templateQuality.HandleEvent(
			ctx,
			notificationCtx,
			change.Value.TemplateQualityUpdate(),
		); err != nil {
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

		if err := handler.templateCategory.HandleEvent(
			ctx,
			notificationCtx,
			change.Value.TemplateCategoryUpdate(),
		); err != nil {
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

		if err := handler.templateStatus.HandleEvent(
			ctx,
			notificationCtx,
			change.Value.TemplateStatusUpdate(),
		); err != nil {
			if handler.errorHandlerFunc != nil {
				if handlerErr := handler.errorHandlerFunc(ctx, err); handlerErr != nil {
					return handlerErr
				}
			}
		}
	}
	return nil
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
		TimezoneID           string               `json:"timezone_id,omitempty"` // e.g. "Europe/Berlin" or provider's TZ id
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

// OnBusinessAlertNotification registers a handler for account_alerts webhooks
// (messaging limit changes, Official Business Account status, profile photo deletion).
func (handler *Handler) OnBusinessAlertNotification(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AlertNotification) error,
) {
	handler.alerts = EventHandlerFunc[BusinessNotificationContext, AlertNotification](fn)
}

// SetBusinessAlertNotificationHandler sets the handler for account_alerts webhooks.
func (handler *Handler) SetBusinessAlertNotificationHandler(
	fn EventHandler[BusinessNotificationContext, AlertNotification],
) {
	handler.alerts = fn
}

// OnBusinessTemplateStatusUpdate registers a handler for message_template_status_update webhooks
// (template approval, rejection, or suspension).
func (handler *Handler) OnBusinessTemplateStatusUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateStatusUpdateNotification) error,
) {
	handler.templateStatus = EventHandlerFunc[BusinessNotificationContext, TemplateStatusUpdateNotification](fn)
}

// SetBusinessTemplateStatusUpdateHandler sets the handler for message_template_status_update webhooks.
func (handler *Handler) SetBusinessTemplateStatusUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateStatusUpdateNotification],
) {
	handler.templateStatus = fn
}

// OnBusinessTemplateCategoryUpdate registers a handler for template_category_update webhooks
// (template category reclassification).
func (handler *Handler) OnBusinessTemplateCategoryUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateCategoryUpdateNotification) error,
) {
	handler.templateCategory = EventHandlerFunc[BusinessNotificationContext, TemplateCategoryUpdateNotification](
		fn,
	)
}

// SetBusinessTemplateCategoryUpdateHandler sets the handler for template_category_update webhooks.
func (handler *Handler) SetBusinessTemplateCategoryUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateCategoryUpdateNotification],
) {
	handler.templateCategory = fn
}

// OnBusinessTemplateQualityUpdate registers a handler for message_template_quality_update webhooks
// (template quality score changes).
func (handler *Handler) OnBusinessTemplateQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *TemplateQualityUpdateNotification) error,
) {
	handler.templateQuality = EventHandlerFunc[BusinessNotificationContext, TemplateQualityUpdateNotification](
		fn,
	)
}

// SetBusinessTemplateQualityUpdateHandler sets the handler for message_template_quality_update webhooks.
func (handler *Handler) SetBusinessTemplateQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateQualityUpdateNotification],
) {
	handler.templateQuality = fn
}

// OnTemplateComponentsUpdate registers a callback for template component
// change events. Triggers when a WhatsApp template is edited — the callback
// receives the updated header, body, footer, and button details.
func (handler *Handler) OnTemplateComponentsUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext,
		details *TemplateComponentsUpdateNotification) error,
) {
	handler.templateComponents = EventHandlerFunc[BusinessNotificationContext, TemplateComponentsUpdateNotification](fn)
}

// SetTemplateComponentsUpdateHandler sets the handler for message_template_components_update webhooks.
func (handler *Handler) SetTemplateComponentsUpdateHandler(
	fn EventHandler[BusinessNotificationContext, TemplateComponentsUpdateNotification],
) {
	handler.templateComponents = fn
}

// OnBusinessPhoneNumberNameUpdate registers a handler for phone_number_name_update webhooks
// (display name verification outcomes).
func (handler *Handler) OnBusinessPhoneNumberNameUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberNameUpdate) error,
) {
	handler.phoneNumberNameUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberNameUpdate](fn)
}

// SetBusinessPhoneNumberNameUpdateHandler sets the handler for phone_number_name_update webhooks.
func (handler *Handler) SetBusinessPhoneNumberNameUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberNameUpdate],
) {
	handler.phoneNumberNameUpdate = fn
}

// OnBusinessPhoneNumberQualityUpdate registers a handler for phone_number_quality_update webhooks
// (phone number throughput level changes / quality rating).
func (handler *Handler) OnBusinessPhoneNumberQualityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberQualityUpdate) error,
) {
	handler.phoneNumberQualityUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberQualityUpdate](
		fn,
	)
}

// SetBusinessPhoneNumberQualityUpdateHandler sets the handler for phone_number_quality_update webhooks.
func (handler *Handler) SetBusinessPhoneNumberQualityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, PhoneNumberQualityUpdate],
) {
	handler.phoneNumberQualityUpdate = fn
}

// OnBusinessAccountReviewUpdate registers a handler for account_review_update webhooks
// (WABA policy review outcomes).
func (handler *Handler) OnBusinessAccountReviewUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountReviewUpdate) error,
) {
	handler.accountReviewUpdate = EventHandlerFunc[BusinessNotificationContext, AccountReviewUpdate](fn)
}

// SetBusinessAccountReviewUpdateHandler sets the handler for account_review_update webhooks.
func (handler *Handler) SetBusinessAccountReviewUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountReviewUpdate],
) {
	handler.accountReviewUpdate = fn
}

// OnBusinessAccountUpdate registers a handler for account_update webhooks
// (WABA changes: verification status, violations, deletion, reconnection).
func (handler *Handler) OnBusinessAccountUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *AccountUpdate) error,
) {
	handler.accountUpdate = EventHandlerFunc[BusinessNotificationContext, AccountUpdate](fn)
}

// OnPhoneSettingsUpdate registers a handler for account_settings_update webhooks
// (calling configuration, SIP settings).
func (handler *Handler) OnPhoneSettingsUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *PhoneNumberSettings) error,
) {
	handler.phoneSettingsUpdate = EventHandlerFunc[BusinessNotificationContext, PhoneNumberSettings](fn)
}

// SetBusinessAccountUpdateHandler sets the handler for account_update webhooks.
func (handler *Handler) SetBusinessAccountUpdateHandler(
	fn EventHandler[BusinessNotificationContext, AccountUpdate],
) {
	handler.accountUpdate = fn
}

// OnBusinessCapabilityUpdate registers a handler for business_capability_update webhooks
// (messaging throughput and phone number capability changes).
func (handler *Handler) OnBusinessCapabilityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *CapabilityUpdate) error,
) {
	handler.capabilityUpdate = EventHandlerFunc[BusinessNotificationContext, CapabilityUpdate](fn)
}

// SetBusinessCapabilityUpdateHandler sets the handler for business_capability_update webhooks.
func (handler *Handler) SetBusinessCapabilityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, CapabilityUpdate],
) {
	handler.capabilityUpdate = fn
}

// OnSecurityUpdate registers a callback for security-related phone number
// events. Triggers when a Meta Business Suite user changes the PIN, requests
// a PIN reset, or completes a two-step verification reset.
func (handler *Handler) OnSecurityUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *SecurityNotification) error,
) {
	handler.securityUpdate = EventHandlerFunc[BusinessNotificationContext, SecurityNotification](fn)
}

func (handler *Handler) SetSecurityUpdateHandler(
	fn EventHandler[BusinessNotificationContext, SecurityNotification],
) {
	handler.securityUpdate = fn
}
