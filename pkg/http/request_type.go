package http

import (
	"strings"
)

const (
	RequestTypeSendMessage RequestType = iota
	RequestTypeUpdateStatus
	RequestTypeCreateQR
	RequestTypeListQR
	RequestTypeGetQR
	RequestTypeUpdateQR
	RequestTypeDeleteQR
	RequestTypeListPhoneNumbers
	RequestTypeGetPhoneNumber
	RequestTypeDownloadMedia
	RequestTypeUploadMedia
	RequestTypeDeleteMedia
	RequestTypeGetMedia
	RequestTypeUpdateBusinessProfile
	RequestTypeGetBusinessProfile
	RequestTypeRetrieveFlows
	RequestTypeRetrieveFlowDetails
	RequestTypeRetrieveAssets
	RequestTypePublishFlow
	RequestTypeDeprecateFlow
	RequestTypeDeleteFlow
	RequestTypeUpdateFlow
	RequestTypeCreateFlow
	RequestTypeRetrieveFlowPreview
	RequestTypeGetFlowMetrics
	RequestTypeInstallApp
	RequestTypeRefreshToken
	RequestTypeGenerateToken
	RequestTypeRevokeToken
	RequestTypeTwoStepVerification
	RequestTypeFetchMessagingAnalytics
	RequestTypeFetchTemplateAnalytics
	RequestTypeFetchPricingAnalytics
	RequestTypeFetchConversationAnalytics
	RequestTypeEnableTemplatesAnalytics
	RequestTypeDisableButtonClickTracking
	RequestTypeBlockUsers
	RequestTypeUnblockUsers
	RequestTypeListBlockedUsers
	RequestTypeDisableWelcomeMessage
	RequestTypeEnableWelcomeMessage
	RequestTypeGetConversationAutomationComponents
	RequestTypeUpdateConversationAutomationComponents
	RequestTypeInitResumableUploadSession
	RequestTypeGetResumableUploadSessionStatus
	RequestTypePerformResumableUpload
	RequestTypeSetWABAAlternateCallbackURI
	RequestTypeGetWABAAlternateCallbackURI
	RequestTypeDeleteWABAAlternateCallbackURI
	RequestTypeSetPhoneNumberAlternateCallbackURI
	RequestTypeGetPhoneNumberAlternateCallbackURI
	RequestTypeDeletePhoneNumberAlternateCallbackURI
	RequestTypeGetSettings
	RequestTypeUpdateSettings
	RequestTypeUpdateCallStatus
	RequestTypeCreateGroup
	RequestTypeDeleteGroup
	RequestTypeGetGroupInviteLink
	RequestTypeResetGroupInviteLink
	RequestTypeSendGroupInviteLinkTemplateMessage
	RequestTypeRemoveGroupParticipants
	RequestTypeGetGroupInfo
	RequestTypeGetActiveGroups
	RequestTypeUpdateGroupSettings
	RequestTypeUpdateGroupCallStatus

	// Sentinel (NOT a real request type): keep this last for testing purposes.
	requestTypeCount
)

type (
	RequestType uint8
)

// Single source of truth for String() mappings.
// Must have exactly RequestTypeCount entries, in the same order as the constants.
var requestTypeStrings = [...]string{
	"send_message",
	"update_status",
	"create_qr",
	"list_qr",
	"get_qr",
	"update_qr",
	"delete_qr",
	"list_phone_numbers",
	"get_phone_number",
	"download_media",
	"upload_media",
	"delete_media",
	"get_media",
	"update_business_profile",
	"get_business_profile",
	"retrieve_flows",
	"retrieve_flow_details",
	"retrieve_assets",
	"publish_flow",
	"deprecate_flow",
	"delete_flow",
	"update_flow",
	"create_flow",
	"retrieve_flow_preview",
	"get_flow_metrics",
	"install_app",
	"refresh_token",
	"generate_token",
	"revoke_token",
	"two_step_verification",
	"fetch_messaging_analytics",
	"fetch_template_analytics",
	"fetch_pricing_analytics",
	"fetch_conversation_analytics",
	"enable_templates_analytics",
	"disable_button_click_tracking",
	"block_users",
	"unblock_users",
	"list_blocked_users",
	"disable_welcome_message",
	"enable_welcome_message",
	"get_conversation_automation_components",
	"update_conversation_automation_components",
	"init_resumable_upload_session",
	"get_resumable_upload_session_status",
	"perform_resumable_upload",
	"set_waba_alternate_callback_uri",
	"get_waba_alternate_callback_uri",
	"delete_waba_alternate_callback_uri",
	"set_phonenumber_alternate_callback_uri",
	"get_phonenumber_alternate_callback_uri",
	"delete_phonenumber_alternate_callback_uri",
	"get_settings",
	"update_settings",
	"update_call_status",
	"create_group",
	"delete_group",
	"get_group_invite_link",
	"reset_group_invite_link",
	"send_group_invite_link_template_message",
	"remove_group_participants",
	"get_group_info",
	"get_active_groups",
	"update_group_settings",
	"update_group_call_status",
}

func (r RequestType) Name() string {
	return strings.ReplaceAll(r.String(), "_", " ")
}

func (r RequestType) String() string {
	return requestTypeStrings[r]
}
