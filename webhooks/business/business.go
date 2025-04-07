/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package business

import (
	"context"
	"net/http"

	"github.com/piusalfred/whatsapp/webhooks"
)

const (
	NotificationObjectBusinessAccount   = "whatsapp_business_account"
	ChangeFieldAccountAlerts            = "account_alerts"
	ChangeFieldTemplateStatusUpdate     = "message_template_status_update"
	ChangeFieldTemplateCategoryUpdate   = "template_category_update"
	ChangeFieldTemplateQualityUpdate    = "message_template_quality_update"
	ChangeFieldPhoneNumberNameUpdate    = "phone_number_name_update"
	ChangeFieldBusinessCapabilityUpdate = "business_capability_update"
	ChangeFieldAccountUpdate            = "account_update"
	ChangeFieldAccountReviewUpdate      = "account_review_update"
	ChangeFieldPhoneNumberQualityUpdate = "phone_number_quality_update"
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
		Field string `json:"field"` // e.g., "message_template_status_update"
		Value *Value `json:"value"`
	}

	// Value holds details related to message templates, account limits, and status updates.
	// - Event: e.g., "APPROVED", "FLAGGED", etc.
	// - MessageTemplateID: ID of the message template.
	// - MessageTemplateName: Name of the message template.
	// - MessageTemplateLanguage: Language and locale code, e.g., "en_US".
	// - Reason: Nullable reason for template rejection or status.
	// - PreviousCategory: Previous template category.
	// - NewCategory: New template category.
	// - DisplayPhoneNumber: Display phone number related to the account.
	// - CurrentLimit: Current messaging limit tier.
	// - MaxDailyConversationPerPhone: Max daily conversations allowed.
	// - MaxPhoneNumbersPerBusiness: Max phone numbers per business.
	// - MaxPhoneNumbersPerWABA: Max phone numbers per WABA.
	// - RejectionReason: Reason for template rejection.
	// - RequestedVerifiedName: Verified name request.
	// - RestrictionInfo: Information related to account restrictions.
	// - BanInfo: Information when an account is banned.
	// - Decision: Decision made regarding the account or phone number.
	// - DisableInfo: Information about a template being disabled.
	// - ViolationInfo: Details when the account violates policy.
	Value struct {
		Event                        string            `json:"event,omitempty"`
		MessageTemplateID            int64             `json:"message_template_id,omitempty"`
		MessageTemplateName          string            `json:"message_template_name,omitempty"`
		MessageTemplateLanguage      string            `json:"message_template_language,omitempty"`
		Reason                       *string           `json:"reason,omitempty"`
		PreviousCategory             string            `json:"previous_category,omitempty"`
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

type AlertNotification struct {
	EntityType       string `json:"entity_type,omitempty"`
	EntityID         string `json:"entity_id,omitempty"`
	AlertSeverity    string `json:"alert_severity,omitempty"`
	AlertStatus      string `json:"alert_status,omitempty"`
	AlertType        string `json:"alert_type,omitempty"`
	AlertDescription string `json:"alert_description,omitempty"`
}

func (v *Value) AlertNotification() *AlertNotification {
	return &AlertNotification{
		EntityType:       v.EntityType,
		EntityID:         v.EntityID,
		AlertSeverity:    v.AlertSeverity,
		AlertStatus:      v.AlertStatus,
		AlertType:        v.AlertType,
		AlertDescription: v.AlertDescription,
	}
}

type TemplateStatusUpdateNotification struct {
	Event                   string       `json:"event,omitempty"`
	MessageTemplateID       int64        `json:"message_template_id,omitempty"`
	MessageTemplateName     string       `json:"message_template_name,omitempty"`
	MessageTemplateLanguage string       `json:"message_template_language,omitempty"`
	Reason                  string       `json:"reason,omitempty"`
	DisableInfo             *DisableInfo `json:"disable_info,omitempty"`
	OtherInfo               *OtherInfo   `json:"other_info,omitempty"`
}

func (v *Value) TemplateStatusUpdate() *TemplateStatusUpdateNotification {
	return &TemplateStatusUpdateNotification{
		Event:                   v.Event,
		MessageTemplateID:       v.MessageTemplateID,
		MessageTemplateName:     v.MessageTemplateName,
		MessageTemplateLanguage: v.MessageTemplateLanguage,
		Reason:                  *v.Reason,
		DisableInfo:             v.DisableInfo,
		OtherInfo:               v.OtherInfo,
	}
}

type TemplateCategoryUpdateNotification struct {
	MessageTemplateID       int64  `json:"message_template_id,omitempty"`
	MessageTemplateName     string `json:"message_template_name,omitempty"`
	MessageTemplateLanguage string `json:"message_template_language,omitempty"`
	PreviousCategory        string `json:"previous_category,omitempty"`
	NewCategory             string `json:"new_category,omitempty"`
}

func (v *Value) TemplateCategoryUpdate() *TemplateCategoryUpdateNotification {
	return &TemplateCategoryUpdateNotification{
		MessageTemplateID:       v.MessageTemplateID,
		MessageTemplateName:     v.MessageTemplateName,
		MessageTemplateLanguage: v.MessageTemplateLanguage,
		PreviousCategory:        v.PreviousCategory,
		NewCategory:             v.NewCategory,
	}
}

type TemplateQualityUpdateNotification struct {
	PreviousQualityScore    string `json:"previous_quality_score,omitempty"`
	NewQualityScore         string `json:"new_quality_score,omitempty"`
	MessageTemplateID       int64  `json:"message_template_id,omitempty"`
	MessageTemplateName     string `json:"message_template_name,omitempty"`
	MessageTemplateLanguage string `json:"message_template_language,omitempty"`
}

func (v *Value) TemplateQualityUpdate() *TemplateQualityUpdateNotification {
	return &TemplateQualityUpdateNotification{
		PreviousQualityScore:    v.PreviousCategory,
		NewQualityScore:         v.NewCategory,
		MessageTemplateID:       v.MessageTemplateID,
		MessageTemplateName:     v.MessageTemplateName,
		MessageTemplateLanguage: v.MessageTemplateLanguage,
	}
}

type PhoneNumberNameUpdate struct {
	PhoneNumber           string `json:"display_phone_number"`
	Decision              string `json:"decision"`
	RequestedVerifiedName string `json:"requested_verified_name"`
	RejectionReason       string `json:"rejection_reason"`
}

func (v *Value) PhoneNumberNameUpdate() *PhoneNumberNameUpdate {
	return &PhoneNumberNameUpdate{
		PhoneNumber:           v.DisplayPhoneNumber,
		Decision:              v.Decision,
		RequestedVerifiedName: v.RequestedVerifiedName,
		RejectionReason:       v.RejectionReason,
	}
}

type PhoneNumberQualityUpdate struct {
	DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
	Event              string `json:"event,omitempty"`
	CurrentLimit       string `json:"current_limit,omitempty"`
}

func (v *Value) PhoneNumberQualityUpdate() *PhoneNumberQualityUpdate {
	return &PhoneNumberQualityUpdate{
		DisplayPhoneNumber: v.DisplayPhoneNumber,
		Event:              v.Event,
		CurrentLimit:       v.CurrentLimit,
	}
}

type AccountReviewUpdate struct {
	Decision string `json:"decision,omitempty"`
}

func (v *Value) AccountReviewUpdate() *AccountReviewUpdate {
	return &AccountReviewUpdate{
		Decision: v.Decision,
	}
}

type AccountUpdate struct {
	PhoneNumber     string            `json:"phone_number,omitempty"`
	Event           string            `json:"event,omitempty"`
	RestrictionInfo []RestrictionInfo `json:"restriction_info,omitempty"`
	BanInfo         *BanInfo          `json:"ban_info,omitempty"`
	ViolationInfo   *ViolationInfo    `json:"violation_info,omitempty"`
}

func (v *Value) AccountUpdate() *AccountUpdate {
	return &AccountUpdate{
		PhoneNumber:     v.PhoneNumber,
		Event:           v.Event,
		RestrictionInfo: v.RestrictionInfo,
		BanInfo:         v.BanInfo,
		ViolationInfo:   v.ViolationInfo,
	}
}

type CapabilityUpdate struct {
	MaxDailyConversationPerPhone int `json:"max_daily_conversation_per_phone,omitempty"`
	MaxPhoneNumbersPerBusiness   int `json:"max_phone_numbers_per_business,omitempty"`
}

func (v *Value) CapabilityUpdate() *CapabilityUpdate {
	return &CapabilityUpdate{
		MaxDailyConversationPerPhone: v.MaxDailyConversationPerPhone,
		MaxPhoneNumbersPerBusiness:   v.MaxPhoneNumbersPerBusiness,
	}
}

type WABABanState string

const (
	WABABanStateDisable            WABABanState = "DISABLE"
	WABABanStateReinstate          WABABanState = "REINSTATE"
	WABABanStateScheduleForDisable WABABanState = "SCHEDULE_FOR_DISABLE"
)

type RestrictionType string

const (
	RestrictionTypeRestrictedAddPhoneNumber    RestrictionType = "RESTRICTED_ADD_PHONE_NUMBER_ACTION"
	RestrictionTypeRestrictedBizInitiated      RestrictionType = "RESTRICTED_BIZ_INITIATED_MESSAGING"
	RestrictionTypeRestrictedCustomerInitiated RestrictionType = "RESTRICTED_CUSTOMER_INITIATED_MESSAGING"
)

type TemplateRejectionReason string

const (
	TemplateRejectionReasonAbusiveContent    TemplateRejectionReason = "ABUSIVE_CONTENT"
	TemplateRejectionReasonIncorrectCategory TemplateRejectionReason = "INCORRECT_CATEGORY"
	TemplateRejectionReasonInvalidFormat     TemplateRejectionReason = "INVALID_FORMAT"
	TemplateRejectionReasonScam              TemplateRejectionReason = "SCAM"
	TemplateRejectionReasonNone              TemplateRejectionReason = "NONE"
)

type (
	NotificationHandler     webhooks.NotificationHandler[Notification]
	NotificationHandlerFunc webhooks.NotificationHandlerFunc[Notification]
)

func (e NotificationHandlerFunc) HandleNotification(ctx context.Context,
	notification *Notification,
) *webhooks.Response {
	return e(ctx, notification)
}

type (
	// NotificationContext holds the context of the webhook event. Information includes the object type, ID,
	// and timestamp. that helps to identify the webhook event.
	NotificationContext struct {
		Object      string
		EntryID     string
		EntryTime   int64
		ChangeField string
	}

	EventHandler interface {
		HandleEvent(ctx context.Context, ntx *NotificationContext, notification *Value) error
	}

	EventHandlerFunc func(ctx context.Context, ntx *NotificationContext, notification *Value) error

	HandleEventFunc func(ctx context.Context, event *NotificationContext, value *Value) error
)

func (fn EventHandlerFunc) HandleEvent(ctx context.Context, ntx *NotificationContext, notification *Value) error {
	return fn(ctx, ntx, notification)
}

func (fn HandleEventFunc) HandleNotification(ctx context.Context,
	notification *Notification,
) *webhooks.Response {
	ec := &NotificationContext{}
	for _, entry := range notification.Entry {
		ec.Object = notification.Object
		ec.EntryID = entry.ID
		ec.EntryTime = entry.Time
		for _, change := range entry.Changes {
			ec.ChangeField = change.Field
			if err := fn(ctx, ec, change.Value); err != nil {
				return &webhooks.Response{StatusCode: http.StatusInternalServerError}
			}
		}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}

type (
	MessageHandlerFunc[T any] func(ctx context.Context, ntx *NotificationContext, message *T) error

	Handlers struct {
		AlertsHandler                   MessageHandlerFunc[AlertNotification]
		TemplateStatusHandler           MessageHandlerFunc[TemplateStatusUpdateNotification]
		TemplateCategoryHandler         MessageHandlerFunc[TemplateCategoryUpdateNotification]
		TemplateQualityHandler          MessageHandlerFunc[TemplateQualityUpdateNotification]
		PhoneNumberNameHandler          MessageHandlerFunc[PhoneNumberNameUpdate]
		CapabilityUpdateHandler         MessageHandlerFunc[CapabilityUpdate]
		AccountUpdateHandler            MessageHandlerFunc[AccountUpdate]
		PhoneNumberQualityUpdateHandler MessageHandlerFunc[PhoneNumberQualityUpdate]
		AccountReviewUpdateHandler      MessageHandlerFunc[AccountReviewUpdate]
	}
)

func (h *Handlers) HandleNotification(ctx context.Context,
	notification *Notification) *webhooks.Response {
	ec := &NotificationContext{}
	for _, entry := range notification.Entry {
		ec.Object = notification.Object
		ec.EntryID = entry.ID
		ec.EntryTime = entry.Time
		for _, change := range entry.Changes {
			ec.ChangeField = change.Field
			switch change.Field {
			case ChangeFieldAccountAlerts:
				alert := change.Value.AlertNotification()
				if err := h.AlertsHandler(ctx, ec, alert); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}
			case ChangeFieldTemplateStatusUpdate:
				status := change.Value.TemplateStatusUpdate()
				if err := h.TemplateStatusHandler(ctx, ec, status); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}
			case ChangeFieldTemplateCategoryUpdate:
				category := change.Value.TemplateCategoryUpdate()
				if err := h.TemplateCategoryHandler(ctx, ec, category); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}
			case ChangeFieldTemplateQualityUpdate:
				quality := change.Value.TemplateQualityUpdate()
				if err := h.TemplateQualityHandler(ctx, ec, quality); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}

			case ChangeFieldPhoneNumberNameUpdate:
				name := change.Value.PhoneNumberNameUpdate()
				if err := h.PhoneNumberNameHandler(ctx, ec, name); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}

			case ChangeFieldBusinessCapabilityUpdate:
				capability := change.Value.CapabilityUpdate()
				if err := h.CapabilityUpdateHandler(ctx, ec, capability); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}

			case ChangeFieldAccountUpdate:
				account := change.Value.AccountUpdate()
				if err := h.AccountUpdateHandler(ctx, ec, account); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}

			case ChangeFieldPhoneNumberQualityUpdate:
				quality := change.Value.PhoneNumberQualityUpdate()
				if err := h.PhoneNumberQualityUpdateHandler(ctx, ec, quality); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}

			case ChangeFieldAccountReviewUpdate:
				account := change.Value.AccountReviewUpdate()
				if err := h.AccountReviewUpdateHandler(ctx, ec, account); err != nil {
					return &webhooks.Response{StatusCode: http.StatusInternalServerError}
				}
			}
		}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}
