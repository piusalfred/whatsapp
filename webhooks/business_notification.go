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
		DisableDate string `json:"disable_date"` // Date when the template was disabled
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

	// SecurityNotification describes a security-related event on a business
	// phone number. Events include PIN changes and two-step verification resets.
	// Requester is the Meta Business Suite user ID (only for PIN reset requests).
	SecurityNotification struct {
		Event              string `json:"event"` // PIN_CHANGED, PIN_RESET_REQUEST, PIN_REQUEST_SUCCESS
		DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
		Requester          string `json:"requester,omitempty"` // MBS user ID, PIN reset only
	}

	// TemplateComponentsUpdateNotification describes changes to a template's
	// components (header, body, footer, buttons). Triggers when a template is
	// edited.
	TemplateComponentsUpdateNotification struct {
		MessageTemplateID       int64                     `json:"message_template_id"`
		MessageTemplateName     string                    `json:"message_template_name,omitempty"`
		MessageTemplateLanguage string                    `json:"message_template_language,omitempty"`
		Title                   string                    `json:"message_template_title,omitempty"`
		Element                 string                    `json:"message_template_element,omitempty"`
		Footer                  string                    `json:"message_template_footer,omitempty"`
		Buttons                 []TemplateComponentButton `json:"message_template_buttons,omitempty"`
	}

	// TemplateComponentButton describes a single button within a template
	// component update notification.
	TemplateComponentButton struct {
		Type        string `json:"message_template_button_type"`
		Text        string `json:"message_template_button_text"`
		URL         string `json:"message_template_button_url,omitempty"`
		PhoneNumber string `json:"message_template_button_phone_number,omitempty"`
	}

	// SMBAppStateSync describes a contact sync event from a WhatsApp Business
	// app user onboarded via a solution provider.
	// Triggers: provider syncs contacts, or the business customer adds, edits,
	// or removes a contact in their WhatsApp Business app address book.
	SMBAppStateSync struct {
		Type     string          `json:"type"` // always "contact"
		Contact  *SMBContactSync `json:"contact,omitempty"`
		Action   string          `json:"action"` // "add" or "remove"
		Metadata *SMBMetadata    `json:"metadata,omitempty"`
	}

	// SMBContactSync is the contact information within an SMB state sync event.
	// Fields are omitted when the action is "remove".
	SMBContactSync struct {
		FullName    string `json:"full_name,omitempty"`
		FirstName   string `json:"first_name,omitempty"`
		PhoneNumber string `json:"phone_number"`
	}

	// SMBMetadata carries the timestamp of the state sync event.
	SMBMetadata struct {
		Timestamp int64 `json:"timestamp"`
	}

	ViolationInfo struct {
		ViolationType string `json:"violation_type"` // e.g., "ACCOUNT_VIOLATION"
	}

	BanInfo struct {
		WABABanState WABABanState `json:"waba_ban_state"` // e.g., DISABLE, REINSTATE, SCHEDULE_FOR_DISABLE
		WABABanDate  string       `json:"waba_ban_date"`  // Date of the ban
		CurrentLimit string       `json:"current_limit"`  // Current tier limit of the account
	}

	RestrictionInfo struct {
		RestrictionType RestrictionType `json:"restriction_type"` // e.g., "RESTRICTED_BIZ_INITIATED_MESSAGING"
		Expiration      string          `json:"expiration"`       // Expiration date of the restriction
	}

	RestrictionType string

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

const (
	RestrictionTypeRestrictedAddPhoneNumber    RestrictionType = "RESTRICTED_ADD_PHONE_NUMBER_ACTION"
	RestrictionTypeRestrictedBizInitiated      RestrictionType = "RESTRICTED_BIZ_INITIATED_MESSAGING"
	RestrictionTypeRestrictedCustomerInitiated RestrictionType = "RESTRICTED_CUSTOMER_INITIATED_MESSAGING"
)
