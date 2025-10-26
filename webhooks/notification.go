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

import werrors "github.com/piusalfred/whatsapp/pkg/errors"

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
		Calls                        []*Call              `json:"calls,omitempty"`
		Groups                       []*Group             `json:"groups,omitempty"`
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
