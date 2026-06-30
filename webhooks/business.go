// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Business notification types and BusinessNotificationHandler for WhatsApp
// account webhooks. Covers alerts, template updates (status, category,
// quality, components), phone number name/quality updates, account review,
// capability changes, security, calls, and account settings.

package webhooks

import (
	"context"
	"fmt"
)

// BusinessEventHandler is a shorthand for EventHandler[BusinessNotificationContext, T].
type (
	BusinessEventHandler[T any] EventHandler[BusinessNotificationContext, T]

	// BusinessEventHandlerFunc is a shorthand for EventHandlerFunc[BusinessNotificationContext, T].
	BusinessEventHandlerFunc[T any] EventHandlerFunc[BusinessNotificationContext, T]
)

func (f BusinessEventHandlerFunc[T]) Handle(ctx context.Context, req *BusinessRequest[T]) error {
	return f(ctx, req)
}

type (
	BusinessNotificationContext struct {
		Object      string
		EntryID     string
		EntryTime   int64
		ChangeField string
	}

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

	DisableInfo struct {
		DisableDate string `json:"disable_date"`
	}

	OtherInfo struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
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

	CapabilityUpdate struct {
		MaxDailyConversationPerPhone int `json:"max_daily_conversation_per_phone,omitempty"`
		MaxPhoneNumbersPerBusiness   int `json:"max_phone_numbers_per_business,omitempty"`
	}

	AccountUpdate struct {
		PhoneNumber     string            `json:"phone_number,omitempty"`
		Event           string            `json:"event,omitempty"`
		RestrictionInfo []RestrictionInfo `json:"restriction_info,omitempty"`
		BanInfo         *BanInfo          `json:"ban_info,omitempty"`
		ViolationInfo   *ViolationInfo    `json:"violation_info,omitempty"`
	}

	SecurityNotification struct {
		Event              string `json:"event"`
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		Requester          string `json:"requester,omitempty"`
	}

	TemplateComponentsUpdateNotification struct {
		MessageTemplateID       int64                     `json:"message_template_id"`
		MessageTemplateName     string                    `json:"message_template_name,omitempty"`
		MessageTemplateLanguage string                    `json:"message_template_language,omitempty"`
		Title                   string                    `json:"message_template_title,omitempty"`
		Element                 string                    `json:"message_template_element,omitempty"`
		Footer                  string                    `json:"message_template_footer,omitempty"`
		Buttons                 []TemplateComponentButton `json:"message_template_buttons,omitempty"`
	}

	TemplateComponentButton struct {
		Type        string `json:"message_template_button_type"`
		Text        string `json:"message_template_button_text"`
		URL         string `json:"message_template_button_url,omitempty"`
		PhoneNumber string `json:"message_template_button_phone_number,omitempty"`
	}

	SMBAppStateSync struct {
		Type     string          `json:"type"`
		Contact  *SMBContactSync `json:"contact,omitempty"`
		Action   string          `json:"action"`
		Metadata *SMBMetadata    `json:"metadata,omitempty"`
	}

	SMBContactSync struct {
		FullName    string `json:"full_name,omitempty"`
		FirstName   string `json:"first_name,omitempty"`
		PhoneNumber string `json:"phone_number"`
	}

	SMBMetadata struct {
		Timestamp int64 `json:"timestamp"`
	}

	ViolationInfo struct {
		ViolationType string `json:"violation_type"`
	}

	BanInfo struct {
		WABABanState WABABanState `json:"waba_ban_state"`
		WABABanDate  string       `json:"waba_ban_date"`
		CurrentLimit string       `json:"current_limit"`
	}

	RestrictionInfo struct {
		RestrictionType RestrictionType `json:"restriction_type"`
		Expiration      string          `json:"expiration"`
	}

	RestrictionType string
)

type (
	AlertsHandler                   = BusinessEventHandler[AlertNotification]
	TemplateStatusHandler           = BusinessEventHandler[TemplateStatusUpdateNotification]
	TemplateCategoryHandler         = BusinessEventHandler[TemplateCategoryUpdateNotification]
	TemplateQualityHandler          = BusinessEventHandler[TemplateQualityUpdateNotification]
	PhoneNumberNameUpdateHandler    = BusinessEventHandler[PhoneNumberNameUpdate]
	CapabilityUpdateHandler         = BusinessEventHandler[CapabilityUpdate]
	AccountUpdateHandler            = BusinessEventHandler[AccountUpdate]
	PhoneNumberQualityUpdateHandler = BusinessEventHandler[PhoneNumberQualityUpdate]
	AccountReviewUpdateHandler      = BusinessEventHandler[AccountReviewUpdate]
	TemplateComponentsHandler       = BusinessEventHandler[TemplateComponentsUpdateNotification]
	CallsHandler                    = BusinessEventHandler[CallStatusUpdate]
	SecurityHandler                 = BusinessEventHandler[SecurityNotification]
	PhoneSettingsHandler            = BusinessEventHandler[PhoneNumberSettings]
)

const (
	RestrictionTypeRestrictedAddPhoneNumber    RestrictionType = "RESTRICTED_ADD_PHONE_NUMBER_ACTION"
	RestrictionTypeRestrictedBizInitiated      RestrictionType = "RESTRICTED_BIZ_INITIATED_MESSAGING"
	RestrictionTypeRestrictedCustomerInitiated RestrictionType = "RESTRICTED_CUSTOMER_INITIATED_MESSAGING"

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
		TimezoneID           string               `json:"timezone_id,omitempty"`
		WeeklyOperatingHours []WeeklyOperatingDay `json:"weekly_operating_hours,omitempty"`
		HolidaySchedule      []Holiday            `json:"holiday_schedule,omitempty"`
	}

	WeeklyOperatingDay struct {
		DayOfWeek string `json:"day_of_week,omitempty"`
		OpenTime  string `json:"open_time,omitempty"`
		CloseTime string `json:"close_time,omitempty"`
	}

	Holiday struct {
		Date      string `json:"date,omitempty"`
		StartTime string `json:"start_time,omitempty"`
		EndTime   string `json:"end_time,omitempty"`
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

// BusinessNotificationHandler groups all per-field-type handlers for business
// account webhooks (alerts, templates, phone numbers, calls, security).
//
// Each exported field accepts a BusinessEventHandler[T] for one WhatsApp
// business notification field type. Leave a field nil to silently skip that
// notification type (HTTP 200).
//
// Usage:
//
//	bh := &BusinessNotificationHandler{}
//	bh.OnAlerts(myAlertHandler)
type BusinessNotificationHandler struct {
	Alerts             BusinessEventHandler[AlertNotification]
	TemplateStatus     BusinessEventHandler[TemplateStatusUpdateNotification]
	TemplateCategory   BusinessEventHandler[TemplateCategoryUpdateNotification]
	TemplateQuality    BusinessEventHandler[TemplateQualityUpdateNotification]
	TemplateComponents BusinessEventHandler[TemplateComponentsUpdateNotification]
	PhoneNumberName    BusinessEventHandler[PhoneNumberNameUpdate]
	PhoneNumberQuality BusinessEventHandler[PhoneNumberQualityUpdate]
	AccountReview      BusinessEventHandler[AccountReviewUpdate]
	Account            BusinessEventHandler[AccountUpdate]
	Capability         BusinessEventHandler[CapabilityUpdate]
	PhoneSettings      BusinessEventHandler[PhoneNumberSettings]
	Calls              BusinessEventHandler[CallStatusUpdate]
	Security           BusinessEventHandler[SecurityNotification]
	Fallback           FallbackHandler
	ErrorHandler       ErrorHandler
}

// OnAlerts sets the handler for account_alerts events.
func (bh *BusinessNotificationHandler) OnAlerts(h BusinessEventHandler[AlertNotification]) {
	bh.Alerts = h
}

// OnTemplateStatus sets the handler for message_template_status_update events.
func (bh *BusinessNotificationHandler) OnTemplateStatus(h BusinessEventHandler[TemplateStatusUpdateNotification]) {
	bh.TemplateStatus = h
}

// OnTemplateCategory sets the handler for message_template_category_update events.
func (bh *BusinessNotificationHandler) OnTemplateCategory(h BusinessEventHandler[TemplateCategoryUpdateNotification]) {
	bh.TemplateCategory = h
}

// OnTemplateQuality sets the handler for message_template_quality_update events.
func (bh *BusinessNotificationHandler) OnTemplateQuality(h BusinessEventHandler[TemplateQualityUpdateNotification]) {
	bh.TemplateQuality = h
}

// OnTemplateComponents sets the handler for message_template_components_update events.
func (bh *BusinessNotificationHandler) OnTemplateComponents(
	h BusinessEventHandler[TemplateComponentsUpdateNotification],
) {
	bh.TemplateComponents = h
}

// OnPhoneNumberName sets the handler for phone_number_name_update events.
func (bh *BusinessNotificationHandler) OnPhoneNumberName(h BusinessEventHandler[PhoneNumberNameUpdate]) {
	bh.PhoneNumberName = h
}

// OnPhoneNumberQuality sets the handler for phone_number_quality_update events.
func (bh *BusinessNotificationHandler) OnPhoneNumberQuality(h BusinessEventHandler[PhoneNumberQualityUpdate]) {
	bh.PhoneNumberQuality = h
}

// OnAccountReview sets the handler for account_review_update events.
func (bh *BusinessNotificationHandler) OnAccountReview(h BusinessEventHandler[AccountReviewUpdate]) {
	bh.AccountReview = h
}

// OnAccount sets the handler for account_update events.
func (bh *BusinessNotificationHandler) OnAccount(h BusinessEventHandler[AccountUpdate]) {
	bh.Account = h
}

// OnCapability sets the handler for business_capability_update events.
func (bh *BusinessNotificationHandler) OnCapability(h BusinessEventHandler[CapabilityUpdate]) {
	bh.Capability = h
}

// OnPhoneSettings sets the handler for account_settings_update events.
func (bh *BusinessNotificationHandler) OnPhoneSettings(h BusinessEventHandler[PhoneNumberSettings]) {
	bh.PhoneSettings = h
}

// OnCalls sets the handler for calls events.
func (bh *BusinessNotificationHandler) OnCalls(h BusinessEventHandler[CallStatusUpdate]) {
	bh.Calls = h
}

// OnSecurity sets the handler for security events.
func (bh *BusinessNotificationHandler) OnSecurity(h BusinessEventHandler[SecurityNotification]) {
	bh.Security = h
}

// OnFallback sets the catch-all handler for business events without a dedicated
// sub-category handler.
func (bh *BusinessNotificationHandler) OnFallback(h FallbackHandler) {
	bh.Fallback = h
}

// Handle dispatches the business notification change to the correct handler
// based on change.Field. Nil handlers route to Fallback if set, otherwise
// silently skip.
//
//nolint:funlen,gocognit,gocyclo // dispatch switch
func (bh *BusinessNotificationHandler) Handle(
	ctx context.Context,
	ne NotificationEntry,
	change Change,
) error {
	nctx := &BusinessNotificationContext{
		Object:      ne.Object,
		EntryID:     ne.ID,
		EntryTime:   ne.Time,
		ChangeField: change.Field,
	}

	switch change.Field {
	case ChangeFieldAccountAlerts.String():
		if bh.Alerts != nil {
			if err := bh.Alerts.Handle(ctx, &BusinessRequest[AlertNotification]{Context: nctx, Payload: change.Value.AlertNotification()}); err != nil {
				return fmt.Errorf("business alerts: %w", err)
			}
			return nil
		}
	case ChangeFieldTemplateStatusUpdate.String():
		if bh.TemplateStatus != nil {
			if err := bh.TemplateStatus.Handle(ctx, &BusinessRequest[TemplateStatusUpdateNotification]{Context: nctx, Payload: change.Value.TemplateStatusUpdate()}); err != nil {
				return fmt.Errorf("business template status: %w", err)
			}
			return nil
		}
	case ChangeFieldTemplateCategoryUpdate.String():
		if bh.TemplateCategory != nil {
			if err := bh.TemplateCategory.Handle(ctx, &BusinessRequest[TemplateCategoryUpdateNotification]{Context: nctx, Payload: change.Value.TemplateCategoryUpdate()}); err != nil {
				return fmt.Errorf("business template category: %w", err)
			}
			return nil
		}
	case ChangeFieldTemplateQualityUpdate.String():
		if bh.TemplateQuality != nil {
			if err := bh.TemplateQuality.Handle(ctx, &BusinessRequest[TemplateQualityUpdateNotification]{Context: nctx, Payload: change.Value.TemplateQualityUpdate()}); err != nil {
				return fmt.Errorf("business template quality: %w", err)
			}
			return nil
		}
	case ChangeFieldTemplateComponentsUpdate.String():
		if bh.TemplateComponents != nil {
			data := &TemplateComponentsUpdateNotification{
				MessageTemplateID:       change.Value.MessageTemplateID,
				MessageTemplateName:     change.Value.MessageTemplateName,
				MessageTemplateLanguage: change.Value.MessageTemplateLanguage,
				Title:                   change.Value.MessageTemplateTitle,
				Element:                 change.Value.MessageTemplateElement,
				Footer:                  change.Value.MessageTemplateFooter,
				Buttons:                 change.Value.MessageTemplateButtons,
			}
			if err := bh.TemplateComponents.Handle(ctx, &BusinessRequest[TemplateComponentsUpdateNotification]{Context: nctx, Payload: data}); err != nil {
				return fmt.Errorf("business template components: %w", err)
			}
			return nil
		}
	case ChangeFieldPhoneNumberNameUpdate.String():
		if bh.PhoneNumberName != nil {
			if err := bh.PhoneNumberName.Handle(ctx, &BusinessRequest[PhoneNumberNameUpdate]{Context: nctx, Payload: change.Value.PhoneNumberNameUpdate()}); err != nil {
				return fmt.Errorf("business phone name: %w", err)
			}
			return nil
		}
	case ChangeFieldPhoneNumberQualityUpdate.String():
		if bh.PhoneNumberQuality != nil {
			if err := bh.PhoneNumberQuality.Handle(ctx, &BusinessRequest[PhoneNumberQualityUpdate]{Context: nctx, Payload: change.Value.PhoneNumberQualityUpdate()}); err != nil {
				return fmt.Errorf("business phone quality: %w", err)
			}
			return nil
		}
	case ChangeFieldAccountReviewUpdate.String():
		if bh.AccountReview != nil {
			if err := bh.AccountReview.Handle(ctx, &BusinessRequest[AccountReviewUpdate]{Context: nctx, Payload: change.Value.AccountReviewUpdate()}); err != nil {
				return fmt.Errorf("business account review: %w", err)
			}
			return nil
		}
	case ChangeFieldAccountUpdate.String():
		if bh.Account != nil {
			if err := bh.Account.Handle(ctx, &BusinessRequest[AccountUpdate]{Context: nctx, Payload: change.Value.AccountUpdate()}); err != nil {
				return fmt.Errorf("business account: %w", err)
			}
			return nil
		}
	case ChangeFieldCalls.String():
		if bh.Calls != nil {
			if err := bh.Calls.Handle(ctx, &BusinessRequest[CallStatusUpdate]{Context: nctx, Payload: change.Value.CallStatusUpdate()}); err != nil {
				return fmt.Errorf("business calls: %w", err)
			}
			return nil
		}
	case ChangeFieldBusinessCapabilityUpdate.String():
		if bh.Capability != nil {
			if err := bh.Capability.Handle(ctx, &BusinessRequest[CapabilityUpdate]{Context: nctx, Payload: change.Value.CapabilityUpdate()}); err != nil {
				return fmt.Errorf("business capability: %w", err)
			}
			return nil
		}
	case ChangeFieldAccountSettingsUpdate.String():
		if bh.PhoneSettings != nil {
			if err := bh.PhoneSettings.Handle(ctx, &BusinessRequest[PhoneNumberSettings]{Context: nctx, Payload: change.Value.PhoneNumberSettings}); err != nil {
				return fmt.Errorf("business phone settings: %w", err)
			}
			return nil
		}
	case ChangeFieldSecurity.String():
		if bh.Security != nil {
			data := &SecurityNotification{
				Event:              change.Value.Event,
				DisplayPhoneNumber: change.Value.DisplayPhoneNumber,
				Requester:          change.Value.Requester,
			}
			if err := bh.Security.Handle(ctx, &BusinessRequest[SecurityNotification]{Context: nctx, Payload: data}); err != nil {
				return fmt.Errorf("business security: %w", err)
			}
			return nil
		}
	}

	if bh.Fallback != nil {
		if err := bh.Fallback.Handle(ctx, ne, change); err != nil {
			return fmt.Errorf("business fallback: %w", err)
		}
	}
	return nil
}

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
	return &AccountReviewUpdate{Decision: value.Decision}
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

func (handler *Handler) OnBusinessAlertNotification(h AlertsHandler) {
	handler.business.Alerts = h
}

func (handler *Handler) OnBusinessTemplateStatusUpdate(h TemplateStatusHandler) {
	handler.business.TemplateStatus = h
}

func (handler *Handler) OnBusinessTemplateCategoryUpdate(h TemplateCategoryHandler) {
	handler.business.TemplateCategory = h
}

func (handler *Handler) OnBusinessTemplateQualityUpdate(h TemplateQualityHandler) {
	handler.business.TemplateQuality = h
}

func (handler *Handler) OnTemplateComponentsUpdate(h TemplateComponentsHandler) {
	handler.business.TemplateComponents = h
}

func (handler *Handler) OnBusinessPhoneNumberNameUpdate(h PhoneNumberNameUpdateHandler) {
	handler.business.PhoneNumberName = h
}

func (handler *Handler) OnBusinessPhoneNumberQualityUpdate(h PhoneNumberQualityUpdateHandler) {
	handler.business.PhoneNumberQuality = h
}

func (handler *Handler) OnBusinessAccountReviewUpdate(h AccountReviewUpdateHandler) {
	handler.business.AccountReview = h
}

func (handler *Handler) OnBusinessAccountUpdate(h AccountUpdateHandler) {
	handler.business.Account = h
}

func (handler *Handler) OnPhoneSettingsUpdate(h PhoneSettingsHandler) {
	handler.business.PhoneSettings = h
}

func (handler *Handler) OnBusinessCapabilityUpdate(h CapabilityUpdateHandler) {
	handler.business.Capability = h
}

func (handler *Handler) OnSecurityUpdate(h SecurityHandler) {
	handler.business.Security = h
}
