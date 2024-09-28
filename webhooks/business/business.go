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
		Value Value  `json:"value"`
	}

	Value struct {
		Event                        string            `json:"event,omitempty"`                            // e.g., "APPROVED", "FLAGGED", etc.
		MessageTemplateID            int64             `json:"message_template_id,omitempty"`              // ID of the message template
		MessageTemplateName          string            `json:"message_template_name,omitempty"`            // Name of the message template
		MessageTemplateLanguage      string            `json:"message_template_language,omitempty"`        // Language and locale code, e.g., "en_US"
		Reason                       *string           `json:"reason,omitempty"`                           // Reason for template rejection or status, nullable
		PreviousCategory             string            `json:"previous_category,omitempty"`                // Previous template category
		NewCategory                  string            `json:"new_category,omitempty"`                     // New template category
		DisplayPhoneNumber           string            `json:"display_phone_number,omitempty"`             // Display phone number related to the account
		CurrentLimit                 string            `json:"current_limit,omitempty"`                    // Current messaging limit tier
		MaxDailyConversationPerPhone int               `json:"max_daily_conversation_per_phone,omitempty"` // Max daily conversations allowed
		MaxPhoneNumbersPerBusiness   int               `json:"max_phone_numbers_per_business,omitempty"`   // Max phone numbers per business
		MaxPhoneNumbersPerWABA       int               `json:"max_phone_numbers_per_waba,omitempty"`       // Max phone numbers per WABA
		RejectionReason              string            `json:"rejection_reason,omitempty"`                 // Reason for template rejection
		RequestedVerifiedName        string            `json:"requested_verified_name,omitempty"`          // Verified name request
		RestrictionInfo              []RestrictionInfo `json:"restriction_info,omitempty"`                 // Info related to account restrictions
		BanInfo                      *BanInfo          `json:"ban_info,omitempty"`                         // Information when an account is banned
		Decision                     string            `json:"decision,omitempty"`                         // Decision made regarding the account or phone number
		DisableInfo                  *DisableInfo      `json:"disable_info,omitempty"`                     // Information about a template being disabled
		ViolationInfo                *ViolationInfo    `json:"violation_info,omitempty"`                   // Details when account violates policy
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

func (e NotificationHandlerFunc) HandleNotification(ctx context.Context, notification *Notification) *webhooks.Response {
	return e(ctx, notification)
}
