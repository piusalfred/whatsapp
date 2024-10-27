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
		ViolationInfo                *ViolationInfo    `json:"violation_info,omitempty"`
	}

	BanInfo struct {
		WABABanState WABABanState `json:"waba_ban_state"` // e.g., DISABLE, REINSTATE, SCHEDULE_FOR_DISABLE
		WABABanDate  string       `json:"waba_ban_date"`  // Date of the ban
		CurrentLimit string       `json:"current_limit"`  // Current tier limit of the account
	}

	DisableInfo struct {
		DisableDate string `json:"disable_date"` // Date when the template was disabled
	}

	ViolationInfo struct {
		ViolationType string `json:"violation_type"` // e.g., "ACCOUNT_VIOLATION"
	}

	RestrictionInfo struct {
		RestrictionType RestrictionType `json:"restriction_type"` // e.g., "RESTRICTED_BIZ_INITIATED_MESSAGING"
		Expiration      string          `json:"expiration"`       // Expiration date of the restriction
	}
)

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

	EvenHandler interface {
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
