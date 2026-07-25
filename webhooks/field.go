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

// ChangeField constants enumerate every webhook field supported by the
// WhatsApp Business API. Each constant documents the field purpose,
// trigger conditions, and processing nuances sourced from the official API
// reference.

package webhooks

// ChangeField identifies the type of webhook notification. The string value
// matches the WhatsApp API field name in the change object.
type ChangeField string

const (
	// ChangeFieldFlows corresponds to the "flows" webhook field. Notifies
	// you of flow status changes and endpoint-powered Flow performance
	// (client error rates, endpoint error rates, latency, availability).
	//
	// Flow response messages (user completing a flow) arrive through the
	// "messages" webhook as interactive nfm_reply messages, not here.
	//
	// Events delivered through this field:
	//
	//   FLOW_STATUS_CHANGE — flow is Published, Throttled, Blocked,
	//   sunset to DEPRECATED, or created as Draft. Also sent when a flow
	//   version is about to freeze with a migration warning.
	//
	//   CLIENT_ERROR_RATE — client-side screen navigation error rate
	//   crossed a threshold (5%, 10%, 50%) in the last 60 minutes.
	//
	//   ENDPOINT_ERROR_RATE — endpoint request error rate crossed a
	//   threshold (5%, 10%, 50%) in the last 30 minutes.
	//
	//   ENDPOINT_LATENCY — p90 endpoint latency crossed a threshold
	//   (1s, 5s, 7s) in the last 30 minutes.
	//
	//   ENDPOINT_AVAILABILITY — endpoint availability fell below 90%
	//   in the last 10 minutes.
	//
	//   FLOW_VERSION_EXPIRY_WARNING — a flow version will freeze within
	//   21 days; migrate to the recommended version.
	//
	// Each alert includes an alert_state (ACTIVATED or DEACTIVATED)
	// indicating whether the threshold was crossed or recovered from.
	ChangeFieldFlows ChangeField = "flows"

	// ChangeFieldAccountAlerts corresponds to the "account_alerts" webhook
	// field. Notifies you of changes to a business phone number's messaging
	// limit, business profile, and Official Business Account status.
	//
	// Triggers:
	//   - An increase to the messaging limit of all portfolio phone numbers
	//     is denied, deferred, or requires additional information.
	//   - A phone number's Official Business Account (green checkmark)
	//     status is approved or denied.
	//   - A phone number's business profile photo is deleted.
	ChangeFieldAccountAlerts ChangeField = "account_alerts"

	// ChangeFieldAccountReviewUpdate corresponds to the
	// "account_review_update" webhook field. Notifies you when a WhatsApp
	// Business Account (WABA) has been reviewed against Meta policy
	// guidelines.
	//
	// Triggers:
	//   - A WhatsApp Business account is approved following review.
	//   - A WhatsApp Business account is rejected due to policy violations.
	//   - A decision on approval has been deferred or is awaiting more
	//     information.
	ChangeFieldAccountReviewUpdate ChangeField = "account_review_update"

	// ChangeFieldAccountUpdate corresponds to the "account_update" webhook
	// field. Notifies you of changes to a WABA's partner-led business
	// verification submission, authentication-international rate eligibility,
	// primary business location, policy violations, offboarding, or deletion.
	//
	// Triggers:
	//   - Partner-led business verification is approved, rejected, or
	//     discarded.
	//   - The WABA is permanently deleted.
	//   - The WABA is shared ("installed") or unshared ("uninstalled") with
	//     a Solution Partner.
	//   - The WABA violates Meta policies or terms.
	//   - The WABA becomes eligible for authentication-international rates.
	//   - The primary business location is set or updated.
	//   - The partner gains explicit access to the WABA's ad accounts.
	//   - The WABA is restricted due to policy enforcement actions.
	//   - The business customer accepts the MM API for WhatsApp TOS.
	//   - App permissions are granted or revoked for the WABA.
	//   - Volume-based pricing tier is updated.
	//   - The WABA is offboarded due to a device change or phone number
	//     reregistration.
	//   - The WABA is reconnected after a device change or number
	//     reregistration.
	ChangeFieldAccountUpdate ChangeField = "account_update"

	// ChangeFieldAccountSettingsUpdate corresponds to the
	// "account_settings_update" webhook field. Notifies you of changes to
	// phone number settings such as calling configuration and callback
	// permissions.
	ChangeFieldAccountSettingsUpdate ChangeField = "account_settings_update"

	// ChangeFieldBusinessCapabilityUpdate corresponds to the
	// "business_capability_update" webhook field. Notifies you of WABA or
	// business portfolio capability changes (messaging limits, phone number
	// limits, etc.).
	//
	// Triggers:
	//   - A new WhatsApp Business account is provisioned and created.
	//   - Core capability metrics (messaging tier limits, maximum phone
	//     number limits) are increased or decreased.
	ChangeFieldBusinessCapabilityUpdate ChangeField = "business_capability_update"

	// ChangeFieldCalls corresponds to the "calls" webhook field. Notifies
	// you of calling events for phone numbers with calling enabled.
	//
	// Five event types are delivered through this field:
	//
	//   Call Connect (event: "connect") — a business-initiated WebRTC call
	//   is ready; contains the SDP Answer required to establish the media
	//   connection. Apply the SDP to your WebRTC stack.
	//
	//   Call Created (event: "call_created") — a SIP call was attempted
	//   (business or user initiated). No session object since signaling is
	//   handled via SIP rather than WebRTC.
	//
	//   Call Status (statuses array, type: "call") — the call is ringing,
	//   accepted, or rejected by the WhatsApp user. Delivered as a statuses
	//   array mirroring the message status webhook structure.
	//
	//   Call Terminate (event: "terminate") — the call ended for any
	//   reason (hangup, API terminate/reject). Includes start_time,
	//   end_time, duration (seconds), and status (COMPLETED or FAILED).
	//
	//   Call Permission Reply — user accepts or rejects a call permission
	//   request. Arrives through the "messages" webhook as an interactive
	//   call_permission_reply message, not through this field.
	//
	// Call direction is BUSINESS_INITIATED or USER_INITIATED.
	ChangeFieldCalls ChangeField = "calls"

	// ChangeFieldGroupLifecycleUpdate corresponds to the
	// "group_lifecycle_update" webhook field. Notifies you when a group is
	// created or deleted.
	//
	// Triggers:
	//   - A group is created (success or failure).
	//   - A group is deleted (success or failure).
	ChangeFieldGroupLifecycleUpdate ChangeField = "group_lifecycle_update"

	// ChangeFieldGroupParticipantsUpdate corresponds to the
	// "group_participants_update" webhook field. Notifies you when a
	// WhatsApp user joins a group via invite, requests to join, cancels
	// their request, or when join requests are approved; also participant
	// removal and departures.
	ChangeFieldGroupParticipantsUpdate ChangeField = "group_participants_update"

	// ChangeFieldGroupSettingsUpdate corresponds to the
	// "group_settings_update" webhook field. Notifies you of changes to
	// group settings (subject, description, profile picture) with per-field
	// success/failure reporting.
	ChangeFieldGroupSettingsUpdate ChangeField = "group_settings_update"

	// ChangeFieldGroupStatusUpdate corresponds to the
	// "group_status_update" webhook field. Notifies you when a group is
	// suspended or a suspension is cleared.
	ChangeFieldGroupStatusUpdate ChangeField = "group_status_update"

	// ChangeFieldHistory corresponds to the "history" webhook field.
	// Synchronizes the WhatsApp Business app chat history of a business
	// customer onboarded by a solution provider. History is delivered in
	// three phases spanning 180 days from onboarding.
	//
	// Two webhook shapes share this field:
	//   1. Thread messages — history entries with metadata (phase,
	//      chunk_order, progress) and threads of messages with delivery
	//      status. A single webhook may contain thousands of messages.
	//   2. Media messages — separate webhooks with actual media asset IDs
	//      for messages sent within 14 days of onboarding.
	//
	// Triggers:
	//   - A partner syncs chat history for an onboarded customer who
	//     approved sharing.
	//   - A partner syncs chat history for a customer who declined
	//     sharing (zero-message webhook with errors).
	//
	// Operational notes:
	//   - Phases: 0 (day 0–1), 1 (day 1–90), 2 (day 90–180).
	//   - Chunks may arrive out of order; use chunk_order to reassemble.
	//   - progress=100 means sync is complete.
	//   - Group chats are excluded.
	//   - Media messages initially appear as "media_placeholder" type;
	//     actual media content follows in a separate webhook.
	//
	// WARNING: A single webhook can contain thousands of messages. Do NOT
	// process synchronously — it will timeout. Capture and persist the
	// payload immediately, then process asynchronously.
	ChangeFieldHistory ChangeField = "history"

	// ChangeFieldMessages corresponds to the "messages" webhook field.
	// The main conversational webhook. Handles incoming customer messages
	// (text, image, audio, video, interactive, location, contacts, etc.)
	// and outgoing message delivery status updates (sent, delivered, read,
	// failed).
	//
	// Triggers:
	//   - An end-user sends a supported message type to the business.
	//   - Outbound message statuses change state as they transit Meta
	//     infrastructure.
	ChangeFieldMessages ChangeField = "messages"

	// ChangeFieldTemplateComponentsUpdate corresponds to the
	// "message_template_components_update" webhook field. Notifies you of
	// changes to a template's components.
	//
	// Triggers:
	//   - An existing template has its component layout, buttons,
	//     parameters, or media headers edited.
	ChangeFieldTemplateComponentsUpdate ChangeField = "message_template_components_update"

	// ChangeFieldTemplateCategoryUpdate corresponds to the
	// "template_category_update" webhook field. Notifies you of changes to
	// a template's category.
	//
	// Triggers:
	//   - The existing category of a template is changed by an automated
	//     process.
	//   - The category is changed manually.
	ChangeFieldTemplateCategoryUpdate ChangeField = "template_category_update"

	// ChangeFieldTemplateQualityUpdate corresponds to the
	// "message_template_quality_update" webhook field. Notifies you of
	// changes to a template's quality score.
	//
	// Triggers:
	//   - A template's quality score changes tiers (e.g., shifts between
	//     High, Medium, Low based on user blocks or reports).
	ChangeFieldTemplateQualityUpdate ChangeField = "message_template_quality_update"

	// ChangeFieldTemplateStatusUpdate corresponds to the
	// "message_template_status_update" webhook field. Notifies you of
	// changes to the status of an existing message template.
	//
	// Triggers:
	//   - A template is approved.
	//   - A template is rejected.
	//   - A template is disabled (e.g., due to low quality score).
	//   - A template is archived.
	//   - A template is unarchived.
	ChangeFieldTemplateStatusUpdate ChangeField = "message_template_status_update"

	// ChangeFieldPhoneNumberNameUpdate corresponds to the
	// "phone_number_name_update" webhook field. Notifies you of business
	// phone number display name verification outcomes.
	//
	// Triggers:
	//   - A newly created phone number's display name undergoes initial
	//     Meta automated/manual verification.
	//   - An already approved display name is edited and re-reviewed.
	ChangeFieldPhoneNumberNameUpdate ChangeField = "phone_number_name_update"

	// ChangeFieldPhoneNumberQualityUpdate corresponds to the
	// "phone_number_quality_update" webhook field. Notifies you of changes
	// to a phone number's messaging throughput level.
	//
	// Triggers:
	//   - A phone number's throughput level changes scale (e.g., moving
	//     between standard and upgraded messaging tiers).
	ChangeFieldPhoneNumberQualityUpdate ChangeField = "phone_number_quality_update"

	// ChangeFieldSecurity corresponds to the "security" webhook field.
	// Notifies you of changes to a business phone number's security
	// settings, specifically two-step verification.
	//
	// Triggers:
	//   - A Meta Business Suite user clicks "Turn off two-step
	//     verification" in WhatsApp Manager.
	//   - A user completes the Two-Step Verification Reset email to turn
	//     off two-step verification.
	//   - A user changes or enables the phone number PIN via WhatsApp
	//     Manager.
	ChangeFieldSecurity ChangeField = "security"

	// ChangeFieldSMBAppStateSync corresponds to the "smb_app_state_sync"
	// webhook field. Synchronizes contacts of WhatsApp Business app users
	// who have been onboarded via a solution provider.
	//
	// Triggers:
	//   - A partner initiates an initial forced contact database sync.
	//   - The onboarded business customer adds a new local device contact.
	//   - The business customer deletes a contact record.
	//   - The business customer edits an existing contact card.
	ChangeFieldSMBAppStateSync ChangeField = "smb_app_state_sync"

	// ChangeFieldSMBMessageEchoes corresponds to the "smb_message_echoes"
	// webhook field. Notifies you of messages sent via the WhatsApp Business
	// app or a companion ("linked") device by a business customer who has
	// been onboarded to Cloud API via a solution provider. The payload shape
	// is identical to incoming messages.
	//
	// Triggers:
	//   - An onboarded business customer sends an outbound message using
	//     the native app UI or a linked workspace device.
	//   - The business customer revokes (deletes) a previously sent
	//     message.
	//   - The business customer edits a previously sent message.
	ChangeFieldSMBMessageEchoes ChangeField = "smb_message_echoes"

	// ChangeFieldUserPreferences corresponds to the "user_preferences"
	// webhook field. Notifies you of changes to a WhatsApp user's marketing
	// message preferences.
	//
	// Triggers:
	//   - A WhatsApp user chooses to stop receiving marketing messages.
	//   - A WhatsApp user explicitly resumes marketing communications.
	//
	// Note: This webhook triggers exclusively on broad stop/resume
	// preferences. It does NOT fire when a customer interacts with
	// "Interested" or "Not interested" survey context within the native
	// "Offers and announcements" configuration screen.
	ChangeFieldUserPreferences ChangeField = "user_preferences"

	// ChangeFieldPartnerSolutions corresponds to the "partner_solutions"
	// webhook field. Describes changes to the status of a Multi-Partner
	// Solution.
	//
	// Triggers:
	//   - A multi-partner solution is saved as a draft
	//     (solution_status: DRAFT).
	//   - A solution request is sent to a partner
	//     (solution_status: INITIATED).
	//   - A partner accepts a solution request
	//     (solution_status: ACTIVE).
	//   - A partner rejects a solution request
	//     (solution_status: REJECTED).
	//   - A partner requests deactivation of a solution.
	//   - A solution is deactivated (solution_status: DEACTIVATED).
	//
	// Note: This webhook field is not yet implemented by the handler
	// dispatch. Received payloads are silently acknowledged (HTTP 200).
	ChangeFieldPartnerSolutions ChangeField = "partner_solutions"

	// ChangeFieldPaymentConfigUpdate corresponds to the
	// "payment_configuration_update" webhook field. Notifies you of changes
	// to payment configurations for Payments API India and Payments API
	// Brazil.
	//
	// Triggers:
	//   - The payment configuration is connected to a payment gateway
	//     account.
	//   - The payment configuration is disconnected from a payment
	//     gateway account.
	//   - The payment configuration is now active.
	//
	// Note: This webhook field is not yet implemented by the handler
	// dispatch. Received payloads are silently acknowledged (HTTP 200).
	ChangeFieldPaymentConfigUpdate ChangeField = "payment_configuration_update"

	// UnimplementedChangeField this is placeholder to categorize unimplemented webhook fields.
	UnimplementedChangeField ChangeField = "unimplemented_change_field"
)

// ChangeFieldCategory defines semantic groups for WhatsApp webhook fields.
type ChangeFieldCategory int

const (
	ChangeFieldCategoryUnknown ChangeFieldCategory = iota
	ChangeFieldCategoryFlows
	ChangeFieldCategoryBusiness
	ChangeFieldCategoryCalls
	ChangeFieldCategoryUserPreferences
	ChangeFieldCategorySMBAppStateSync
	ChangeFieldCategoryMessages
	ChangeFieldCategorySMBMessageEchoes
	ChangeFieldCategoryGroups
	ChangeFieldCategoryHistory
)

// GetChangeFieldCategory maps a raw string field name to its broader operational group.
func GetChangeFieldCategory(f string) ChangeFieldCategory {
	switch ChangeField(f) {
	case ChangeFieldFlows:
		return ChangeFieldCategoryFlows

	case ChangeFieldCalls:
		return ChangeFieldCategoryCalls

	case ChangeFieldAccountAlerts,
		ChangeFieldTemplateStatusUpdate,
		ChangeFieldTemplateCategoryUpdate,
		ChangeFieldTemplateQualityUpdate,
		ChangeFieldTemplateComponentsUpdate,
		ChangeFieldPhoneNumberNameUpdate,
		ChangeFieldPhoneNumberQualityUpdate,
		ChangeFieldAccountUpdate,
		ChangeFieldAccountReviewUpdate,
		ChangeFieldBusinessCapabilityUpdate,
		ChangeFieldAccountSettingsUpdate,
		ChangeFieldSecurity:
		return ChangeFieldCategoryBusiness

	case ChangeFieldUserPreferences:
		return ChangeFieldCategoryUserPreferences

	case ChangeFieldSMBAppStateSync:
		return ChangeFieldCategorySMBAppStateSync

	case ChangeFieldMessages:
		return ChangeFieldCategoryMessages

	case ChangeFieldSMBMessageEchoes:
		return ChangeFieldCategorySMBMessageEchoes

	case ChangeFieldGroupLifecycleUpdate,
		ChangeFieldGroupParticipantsUpdate,
		ChangeFieldGroupSettingsUpdate,
		ChangeFieldGroupStatusUpdate:
		return ChangeFieldCategoryGroups

	case ChangeFieldHistory:
		return ChangeFieldCategoryHistory

	default:
		return ChangeFieldCategoryUnknown
	}
}

type changeFieldMap map[ChangeField]struct{}

func initChangeFieldMap(ff ...ChangeField) changeFieldMap {
	m := make(changeFieldMap, len(ff))
	for _, f := range ff {
		m[f] = struct{}{}
	}
	return m
}

func (m changeFieldMap) Add(f ChangeField) {
	m[f] = struct{}{}
}

func (m changeFieldMap) Check(f string) (ChangeField, bool) {
	ff := ChangeField(f)
	if _, ok := m[ff]; ok {
		return ff, true
	}
	return UnimplementedChangeField, false
}
