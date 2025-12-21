/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"testing"
)

func TestRequestType_String(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		if values, names := len(requestTypeStrings), int(requestTypeCount); values != names {
			t.Errorf("Number of RequestType strings (%d) does not match number of RequestType constants (%d)", values, names)
		}
	})

	type testCase struct {
		name        string
		requestType RequestType
		want        string
	}

	tests := []testCase{
		{name: "SendMessage", requestType: RequestTypeSendMessage, want: "send_message"},
		{name: "UpdateStatus", requestType: RequestTypeUpdateStatus, want: "update_status"},
		{name: "CreateQR", requestType: RequestTypeCreateQR, want: "create_qr"},
		{name: "ListQR", requestType: RequestTypeListQR, want: "list_qr"},
		{name: "GetQR", requestType: RequestTypeGetQR, want: "get_qr"},
		{name: "UpdateQR", requestType: RequestTypeUpdateQR, want: "update_qr"},
		{name: "DeleteQR", requestType: RequestTypeDeleteQR, want: "delete_qr"},
		{name: "ListPhoneNumbers", requestType: RequestTypeListPhoneNumbers, want: "list_phone_numbers"},
		{name: "GetPhoneNumber", requestType: RequestTypeGetPhoneNumber, want: "get_phone_number"},
		{name: "DownloadMedia", requestType: RequestTypeDownloadMedia, want: "download_media"},
		{name: "UploadMedia", requestType: RequestTypeUploadMedia, want: "upload_media"},
		{name: "DeleteMedia", requestType: RequestTypeDeleteMedia, want: "delete_media"},
		{name: "GetMedia", requestType: RequestTypeGetMedia, want: "get_media"},
		{name: "UpdateBusinessProfile", requestType: RequestTypeUpdateBusinessProfile, want: "update_business_profile"},
		{name: "GetBusinessProfile", requestType: RequestTypeGetBusinessProfile, want: "get_business_profile"},
		{name: "RetrieveFlows", requestType: RequestTypeRetrieveFlows, want: "retrieve_flows"},
		{name: "RetrieveFlowDetails", requestType: RequestTypeRetrieveFlowDetails, want: "retrieve_flow_details"},
		{name: "RetrieveAssets", requestType: RequestTypeRetrieveAssets, want: "retrieve_assets"},
		{name: "PublishFlow", requestType: RequestTypePublishFlow, want: "publish_flow"},
		{name: "DeprecateFlow", requestType: RequestTypeDeprecateFlow, want: "deprecate_flow"},
		{name: "DeleteFlow", requestType: RequestTypeDeleteFlow, want: "delete_flow"},
		{name: "UpdateFlow", requestType: RequestTypeUpdateFlow, want: "update_flow"},
		{name: "CreateFlow", requestType: RequestTypeCreateFlow, want: "create_flow"},
		{name: "RetrieveFlowPreview", requestType: RequestTypeRetrieveFlowPreview, want: "retrieve_flow_preview"},
		{name: "GetFlowMetrics", requestType: RequestTypeGetFlowMetrics, want: "get_flow_metrics"},
		{name: "InstallApp", requestType: RequestTypeInstallApp, want: "install_app"},
		{name: "RefreshToken", requestType: RequestTypeRefreshToken, want: "refresh_token"},
		{name: "GenerateToken", requestType: RequestTypeGenerateToken, want: "generate_token"},
		{name: "RevokeToken", requestType: RequestTypeRevokeToken, want: "revoke_token"},
		{name: "TwoStepVerification", requestType: RequestTypeTwoStepVerification, want: "two_step_verification"},
		{name: "FetchMessagingAnalytics", requestType: RequestTypeFetchMessagingAnalytics, want: "fetch_messaging_analytics"},
		{name: "FetchTemplateAnalytics", requestType: RequestTypeFetchTemplateAnalytics, want: "fetch_template_analytics"},
		{name: "FetchPricingAnalytics", requestType: RequestTypeFetchPricingAnalytics, want: "fetch_pricing_analytics"},
		{name: "FetchConversationAnalytics", requestType: RequestTypeFetchConversationAnalytics, want: "fetch_conversation_analytics"},
		{name: "EnableTemplatesAnalytics", requestType: RequestTypeEnableTemplatesAnalytics, want: "enable_templates_analytics"},
		{name: "DisableButtonClickTracking", requestType: RequestTypeDisableButtonClickTracking, want: "disable_button_click_tracking"},
		{name: "BlockUsers", requestType: RequestTypeBlockUsers, want: "block_users"},
		{name: "UnblockUsers", requestType: RequestTypeUnblockUsers, want: "unblock_users"},
		{name: "ListBlockedUsers", requestType: RequestTypeListBlockedUsers, want: "list_blocked_users"},
		{name: "DisableWelcomeMessage", requestType: RequestTypeDisableWelcomeMessage, want: "disable_welcome_message"},
		{name: "EnableWelcomeMessage", requestType: RequestTypeEnableWelcomeMessage, want: "enable_welcome_message"},
		{name: "GetConversationAutomationComponents", requestType: RequestTypeGetConversationAutomationComponents, want: "get_conversation_automation_components"},
		{name: "UpdateConversationAutomationComponents", requestType: RequestTypeUpdateConversationAutomationComponents, want: "update_conversation_automation_components"},
		{name: "InitResumableUploadSession", requestType: RequestTypeInitResumableUploadSession, want: "init_resumable_upload_session"},
		{name: "GetResumableUploadSessionStatus", requestType: RequestTypeGetResumableUploadSessionStatus, want: "get_resumable_upload_session_status"},
		{name: "PerformResumableUpload", requestType: RequestTypePerformResumableUpload, want: "perform_resumable_upload"},
		{name: "SetWABAAlternateCallbackURI", requestType: RequestTypeSetWABAAlternateCallbackURI, want: "set_waba_alternate_callback_uri"},
		{name: "GetWABAAlternateCallbackURI", requestType: RequestTypeGetWABAAlternateCallbackURI, want: "get_waba_alternate_callback_uri"},
		{name: "DeleteWABAAlternateCallbackURI", requestType: RequestTypeDeleteWABAAlternateCallbackURI, want: "delete_waba_alternate_callback_uri"},
		{name: "SetPhoneNumberAlternateCallbackURI", requestType: RequestTypeSetPhoneNumberAlternateCallbackURI, want: "set_phonenumber_alternate_callback_uri"},
		{name: "GetPhoneNumberAlternateCallbackURI", requestType: RequestTypeGetPhoneNumberAlternateCallbackURI, want: "get_phonenumber_alternate_callback_uri"},
		{name: "DeletePhoneNumberAlternateCallbackURI", requestType: RequestTypeDeletePhoneNumberAlternateCallbackURI, want: "delete_phonenumber_alternate_callback_uri"},
		{name: "GetSettings", requestType: RequestTypeGetSettings, want: "get_settings"},
		{name: "UpdateSettings", requestType: RequestTypeUpdateSettings, want: "update_settings"},
		{name: "UpdateCallStatus", requestType: RequestTypeUpdateCallStatus, want: "update_call_status"},
		{name: "CreateGroup", requestType: RequestTypeCreateGroup, want: "create_group"},
		{name: "DeleteGroup", requestType: RequestTypeDeleteGroup, want: "delete_group"},
		{name: "GetGroupInviteLink", requestType: RequestTypeGetGroupInviteLink, want: "get_group_invite_link"},
		{name: "ResetGroupInviteLink", requestType: RequestTypeResetGroupInviteLink, want: "reset_group_invite_link"},
		{name: "SendGroupInviteLinkTemplateMessage", requestType: RequestTypeSendGroupInviteLinkTemplateMessage, want: "send_group_invite_link_template_message"},
		{name: "RemoveGroupParticipants", requestType: RequestTypeRemoveGroupParticipants, want: "remove_group_participants"},
		{name: "GetGroupInfo", requestType: RequestTypeGetGroupInfo, want: "get_group_info"},
		{name: "GetActiveGroups", requestType: RequestTypeGetActiveGroups, want: "get_active_groups"},
		{name: "UpdateGroupSettings", requestType: RequestTypeUpdateGroupSettings, want: "update_group_settings"},
		{name: "UpdateGroupCallStatus", requestType: RequestTypeUpdateGroupCallStatus, want: "update_group_call_status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.requestType.String()
			if got != tt.want {
				t.Errorf("RequestType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
