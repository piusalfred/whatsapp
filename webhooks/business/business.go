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
	"slices"

	"github.com/piusalfred/whatsapp/webhooks"
)

const (
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
		PreviousQualityScore:    v.PreviousQualityScore,
		NewQualityScore:         v.NewQualityScore,
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
	NotificationContext struct {
		Object      string
		EntryID     string
		EntryTime   int64
		ChangeField string
	}

	EventHandler[T any] interface {
		HandleEvent(ctx context.Context, ntx *NotificationContext, message *T) error
	}

	EventHandlerFunc[T any] func(ctx context.Context, ntx *NotificationContext, message *T) error

	Handler struct {
		AlertsHandler                   EventHandler[AlertNotification]
		TemplateStatusHandler           EventHandler[TemplateStatusUpdateNotification]
		TemplateCategoryHandler         EventHandler[TemplateCategoryUpdateNotification]
		TemplateQualityHandler          EventHandler[TemplateQualityUpdateNotification]
		PhoneNumberNameHandler          EventHandler[PhoneNumberNameUpdate]
		CapabilityUpdateHandler         EventHandler[CapabilityUpdate]
		AccountUpdateHandler            EventHandler[AccountUpdate]
		PhoneNumberQualityUpdateHandler EventHandler[PhoneNumberQualityUpdate]
		AccountReviewUpdateHandler      EventHandler[AccountReviewUpdate]
		availableHandlers               []string
		errorHandler                    func(ctx context.Context, err error) error
	}
)

func (fn EventHandlerFunc[T]) HandleEvent(ctx context.Context, ntx *NotificationContext, message *T) error {
	return fn(ctx, ntx, message)
}

func (h *Handler) SetAlertsHandler(handler EventHandler[AlertNotification]) {
	h.AlertsHandler = handler
}

func (h *Handler) SetTemplateStatusHandler(handler EventHandler[TemplateStatusUpdateNotification]) {
	h.TemplateStatusHandler = handler
}

func (h *Handler) SetTemplateCategoryHandler(handler EventHandler[TemplateCategoryUpdateNotification]) {
	h.TemplateCategoryHandler = handler
}

func (h *Handler) SetTemplateQualityHandler(handler EventHandler[TemplateQualityUpdateNotification]) {
	h.TemplateQualityHandler = handler
}

func (h *Handler) SetPhoneNumberNameHandler(handler EventHandler[PhoneNumberNameUpdate]) {
	h.PhoneNumberNameHandler = handler
}

func (h *Handler) SetCapabilityUpdateHandler(handler EventHandler[CapabilityUpdate]) {
	h.CapabilityUpdateHandler = handler
}

func (h *Handler) SetAccountUpdateHandler(handler EventHandler[AccountUpdate]) {
	h.AccountUpdateHandler = handler
}

func (h *Handler) SetPhoneNumberQualityUpdateHandler(handler EventHandler[PhoneNumberQualityUpdate]) {
	h.PhoneNumberQualityUpdateHandler = handler
}

func (h *Handler) SetAccountReviewUpdateHandler(handler EventHandler[AccountReviewUpdate]) {
	h.AccountReviewUpdateHandler = handler
}

func (h *Handler) SetErrorHandler(handler func(ctx context.Context, err error) error) {
	h.errorHandler = handler
}

func (h *Handler) ensureHandlers() {
	if h.availableHandlers == nil {
		h.availableHandlers = make([]string, 0)
	}

	h.availableHandlers = h.availableHandlers[:0]

	if h.AlertsHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldAccountAlerts)
	}
	if h.TemplateStatusHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldTemplateStatusUpdate)
	}
	if h.TemplateCategoryHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldTemplateCategoryUpdate)
	}
	if h.TemplateQualityHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldTemplateQualityUpdate)
	}
	if h.PhoneNumberNameHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldPhoneNumberNameUpdate)
	}
	if h.CapabilityUpdateHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldBusinessCapabilityUpdate)
	}
	if h.AccountUpdateHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldAccountUpdate)
	}
	if h.AccountReviewUpdateHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldAccountReviewUpdate)
	}
	if h.PhoneNumberQualityUpdateHandler != nil {
		h.availableHandlers = append(h.availableHandlers, ChangeFieldPhoneNumberQualityUpdate)
	}
}

func (h *Handler) isHandlerSet(name string) bool {
	h.ensureHandlers()

	return slices.Contains(h.availableHandlers, name)
}

// HandleNotification processes the notification and calls the appropriate handler based on the change field.
// if the handler for a specific change field is not set, it will skip that notification. if it is available and
// returns an error, it will call the error handler if set. If the error handler also returns an error,
// the execution will stop and return an internal server error response. if the handlers or error handler do not
// return an error, it will return a success response.
func (h *Handler) HandleNotification(ctx context.Context,
	notification *Notification) *webhooks.Response {
	for _, entry := range notification.Entry {
		for _, change := range entry.Changes {
			notificationCtx := &NotificationContext{
				Object:      notification.Object,
				EntryID:     entry.ID,
				EntryTime:   entry.Time,
				ChangeField: change.Field,
			}

			if err := h.handleChangeValue(ctx, notificationCtx, change.Value); err != nil {
				if h.errorHandler != nil {
					if handlerErr := h.errorHandler(ctx, err); handlerErr != nil {
						return &webhooks.Response{StatusCode: http.StatusInternalServerError}
					}
				}
			}
		}
	}

	return &webhooks.Response{StatusCode: http.StatusOK}
}

func (h *Handler) handleChangeValue( //nolint: gocognit // ok
	ctx context.Context,
	ntx *NotificationContext,
	value *Value,
) error {
	if h.isHandlerSet(ntx.ChangeField) { //nolint: nestif // ok
		switch ntx.ChangeField {
		case ChangeFieldAccountAlerts:
			alert := value.AlertNotification()
			if err := h.AlertsHandler.HandleEvent(ctx, ntx, alert); err != nil {
				return err
			}

		case ChangeFieldTemplateStatusUpdate:
			status := value.TemplateStatusUpdate()
			if err := h.TemplateStatusHandler.HandleEvent(ctx, ntx, status); err != nil {
				return err
			}

		case ChangeFieldTemplateCategoryUpdate:
			category := value.TemplateCategoryUpdate()
			if err := h.TemplateCategoryHandler.HandleEvent(ctx, ntx, category); err != nil {
				return err
			}

		case ChangeFieldTemplateQualityUpdate:
			quality := value.TemplateQualityUpdate()
			if err := h.TemplateQualityHandler.HandleEvent(ctx, ntx, quality); err != nil {
				return err
			}

		case ChangeFieldPhoneNumberNameUpdate:
			name := value.PhoneNumberNameUpdate()
			if err := h.PhoneNumberNameHandler.HandleEvent(ctx, ntx, name); err != nil {
				return err
			}

		case ChangeFieldBusinessCapabilityUpdate:
			capability := value.CapabilityUpdate()
			if err := h.CapabilityUpdateHandler.HandleEvent(ctx, ntx, capability); err != nil {
				return err
			}

		case ChangeFieldAccountUpdate:
			account := value.AccountUpdate()
			if err := h.AccountUpdateHandler.HandleEvent(ctx, ntx, account); err != nil {
				return err
			}

		case ChangeFieldPhoneNumberQualityUpdate:
			quality := value.PhoneNumberQualityUpdate()
			if err := h.PhoneNumberQualityUpdateHandler.HandleEvent(ctx, ntx, quality); err != nil {
				return err
			}

		case ChangeFieldAccountReviewUpdate:
			account := value.AccountReviewUpdate()
			if err := h.AccountReviewUpdateHandler.HandleEvent(ctx, ntx, account); err != nil {
				return err
			}
		}
	}

	return nil
}
