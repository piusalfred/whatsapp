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

// Package api provides a unified client for the WhatsApp Cloud API.
//
// Usage:
//
//	client := api.NewClient(conf, whttp.WithSenderTimeout(30*time.Second))
//	client.SetCallsMiddlewares(logger)
//	resp, err := client.CheckPermission(ctx, req)
//	resp, err := client.CreateQR(ctx, &qrcode.CreateRequest{PrefilledMessage: "Hi"})
package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/piusalfred/whatsapp/auth"
	"github.com/piusalfred/whatsapp/business"
	"github.com/piusalfred/whatsapp/business/analytics"
	"github.com/piusalfred/whatsapp/calls"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/conversation/automation"
	"github.com/piusalfred/whatsapp/flow"
	"github.com/piusalfred/whatsapp/groups"
	"github.com/piusalfred/whatsapp/media"
	messagev2 "github.com/piusalfred/whatsapp/message/v2"
	"github.com/piusalfred/whatsapp/phonenumber"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/qrcode"
	"github.com/piusalfred/whatsapp/settings"
	"github.com/piusalfred/whatsapp/uploads"
	"github.com/piusalfred/whatsapp/user"
	"github.com/piusalfred/whatsapp/webhooks/callbacks"
)

// Client wraps BaseClient with a fixed configuration.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// BaseClient is the multi-tenant layer. Pass a *config.Config per call.
type BaseClient struct {
	calls     *calls.BaseClient
	users     *user.BlockBaseClient
	qrCode    *qrcode.BaseClient
	auto      *automation.BaseClient
	flows     *flow.BaseClient
	media     *media.BaseClient
	settings  *settings.BaseClient
	phone     *phonenumber.BaseClient
	groups    *groups.BaseClient
	biz       *business.BaseClient
	analytics *analytics.BaseClient
	uploads   *uploads.BaseClient
	auth      *auth.BaseClient
	callbacks *callbacks.BaseClient
	message   *messagev2.BaseClient
}

// NewClient creates a Client with the given fixed configuration.
func NewClient(conf *config.Config, opts ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: NewBaseClient(opts...),
		config: conf,
	}
}

// NewBaseClient creates a BaseClient with the given sender options.
func NewBaseClient(opts ...whttp.CoreSenderOption) *BaseClient {
	return &BaseClient{
		calls:     &calls.BaseClient{BaseClient: *whttp.NewBaseClient[calls.BaseRequest](opts...)},
		users:     &user.BlockBaseClient{BaseClient: *whttp.NewBaseClient[user.BlockBaseRequest](opts...)},
		qrCode:    &qrcode.BaseClient{BaseClient: *whttp.NewBaseClient[qrcode.BaseRequest](opts...)},
		auto:      &automation.BaseClient{BaseClient: *whttp.NewBaseClient[automation.BaseRequest](opts...)},
		flows:     &flow.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		media:     &media.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		settings:  &settings.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		phone:     &phonenumber.BaseClient{BaseClient: *whttp.NewBaseClient[phonenumber.BaseRequest](opts...)},
		groups:    &groups.BaseClient{BaseClient: *whttp.NewBaseClient[groups.BaseRequest](opts...)},
		biz:       &business.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		analytics: &analytics.BaseClient{BaseClient: *whttp.NewBaseClient[analytics.BaseRequest](opts...)},
		uploads:   &uploads.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		auth:      &auth.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		callbacks: &callbacks.BaseClient{BaseClient: *whttp.NewBaseClient[callbacks.BaseRequest](opts...)},
		message:   &messagev2.BaseClient{BaseClient: *whttp.NewBaseClient[messagev2.BaseRequest](opts...)},
	}
}

func (c *Client) SetCallsMiddlewares(mws ...whttp.Middleware[calls.BaseRequest]) {
	c.sender.calls.SetMiddlewares(mws...)
}

func (c *Client) SetUsersBlockMiddlewares(mws ...whttp.Middleware[user.BlockBaseRequest]) {
	c.sender.users.SetMiddlewares(mws...)
}

func (c *Client) SetQRCodesMiddlewares(mws ...whttp.Middleware[qrcode.BaseRequest]) {
	c.sender.qrCode.SetMiddlewares(mws...)
}

func (c *Client) SetAutomationMiddlewares(mws ...whttp.Middleware[automation.BaseRequest]) {
	c.sender.auto.SetMiddlewares(mws...)
}

func (c *Client) SetFlowsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.flows.SetMiddlewares(mws...)
}

func (c *Client) SetMediaMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.media.SetMiddlewares(mws...)
}

func (c *Client) SetSettingsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.settings.SetMiddlewares(mws...)
}

func (c *Client) SetPhoneNumbersMiddlewares(mws ...whttp.Middleware[phonenumber.BaseRequest]) {
	c.sender.phone.SetMiddlewares(mws...)
}

func (c *Client) SetGroupsMiddlewares(mws ...whttp.Middleware[groups.BaseRequest]) {
	c.sender.groups.SetMiddlewares(mws...)
}

func (c *Client) SetBusinessMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.biz.SetMiddlewares(mws...)
}

func (c *Client) SetAnalyticsMiddlewares(mws ...whttp.Middleware[analytics.BaseRequest]) {
	c.sender.analytics.SetMiddlewares(mws...)
}

func (c *Client) SetUploadsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.uploads.SetMiddlewares(mws...)
}

func (c *Client) SetAuthMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.auth.SetMiddlewares(mws...)
}

func (c *Client) SetCallbacksMiddlewares(mws ...whttp.Middleware[callbacks.BaseRequest]) {
	c.sender.callbacks.SetMiddlewares(mws...)
}

func (c *Client) SetMessagesMiddlewares(mws ...whttp.Middleware[messagev2.BaseRequest]) {
	c.sender.message.SetMiddlewares(mws...)
}

func (c *Client) CheckCallingPermission(
	ctx context.Context,
	req *calls.CheckPermissionRequest,
) (*calls.CallPermissionCheckResponse, error) {
	resp, err := c.sender.CheckCallingPermission(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("check permission: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateCallStatus(
	ctx context.Context,
	req *calls.CallUpdateStatusRequest,
) (*calls.CallUpdateStatusResponse, error) {
	resp, err := c.sender.UpdateCallStatus(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update call status: %w", err)
	}
	return resp, nil
}

func (c *Client) BlockUsers(ctx context.Context, req *user.BlockRequest) (*user.BlockResponse, error) {
	resp, err := c.sender.BlockUsers(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("block: %w", err)
	}
	return resp, nil
}

func (c *Client) UnblockUsers(ctx context.Context, req *user.UnblockRequest) (*user.BlockResponse, error) {
	resp, err := c.sender.UnblockUsers(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("unblock: %w", err)
	}
	return resp, nil
}

func (c *Client) ListBlockedUsers(
	ctx context.Context,
	opts *user.ListBlockedUsersOptions,
) (*user.ListBlockedUsersResponse, error) {
	resp, err := c.sender.ListBlockedUsers(ctx, c.config, opts)
	if err != nil {
		return nil, fmt.Errorf("list blocked: %w", err)
	}
	return resp, nil
}

func (c *Client) CreateQR(ctx context.Context, req *qrcode.CreateRequest) (*qrcode.CreateResponse, error) {
	resp, err := c.sender.CreateQR(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create qr: %w", err)
	}
	return resp, nil
}

func (c *Client) GetQR(ctx context.Context, qrCodeID string) (*qrcode.Information, error) {
	resp, err := c.sender.GetQR(ctx, c.config, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}
	return resp, nil
}

func (c *Client) ListQR(
	ctx context.Context,
	opts *qrcode.ListOptions,
) (*qrcode.ListResponse, error) {
	resp, err := c.sender.ListQR(ctx, c.config, opts)
	if err != nil {
		return nil, fmt.Errorf("list qr: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteQR(ctx context.Context, qrCodeID string) (*qrcode.SuccessResponse, error) {
	resp, err := c.sender.DeleteQR(ctx, c.config, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("delete qr: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateQR(ctx context.Context, req *qrcode.UpdateRequest) (*qrcode.SuccessResponse, error) {
	resp, err := c.sender.UpdateQR(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update qr: %w", err)
	}
	return resp, nil
}

func (c *Client) AddConversationComponents(
	ctx context.Context,
	commands []*automation.Command,
	prompts []string,
) (*automation.SuccessResponse, error) {
	resp, err := c.sender.AddConversationComponents(ctx, c.config, commands, prompts)
	if err != nil {
		return nil, fmt.Errorf("add components: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateWelcomeMessageStatus(
	ctx context.Context,
	shouldEnable bool,
) (*automation.SuccessResponse, error) {
	resp, err := c.sender.UpdateWelcomeMessageStatus(ctx, c.config, shouldEnable)
	if err != nil {
		return nil, fmt.Errorf("update welcome message status: %w", err)
	}
	return resp, nil
}

func (c *Client) ListConversationComponents(ctx context.Context) (*automation.BaseResponse, error) {
	resp, err := c.sender.ListConversationComponents(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}
	return resp, nil
}

func (c *Client) GetBotDetails(ctx context.Context, request *automation.BotRequest) (*automation.Bot, error) {
	resp, err := c.sender.GetBotDetails(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("get bot details: %w", err)
	}
	return resp, nil
}

func (c *Client) GetSettings(ctx context.Context, req *settings.GetSettingsRequest) (*settings.Settings, error) {
	resp, err := c.sender.GetSettings(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateSettings(ctx context.Context, s *settings.Settings) (*settings.SuccessResponse, error) {
	resp, err := c.sender.UpdateSettings(ctx, c.config, s)
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}
	return resp, nil
}

func (c *Client) ListPhoneNumbers(ctx context.Context) (*phonenumber.ListResponse, error) {
	resp, err := c.sender.ListPhoneNumbers(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list phone numbers: %w", err)
	}
	return resp, nil
}

func (c *Client) GetPhoneNumber(ctx context.Context, req *phonenumber.GetRequest) (*phonenumber.PhoneNumber, error) {
	resp, err := c.sender.GetPhoneNumber(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}
	return resp, nil
}

func (c *Client) CreateFlow(ctx context.Context, req flow.CreateRequest) (*flow.CreateResponse, error) {
	resp, err := c.sender.CreateFlow(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create flow: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateFlow(ctx context.Context, id string, req flow.UpdateRequest) (*flow.UpdateResponse, error) {
	resp, err := c.sender.UpdateFlow(ctx, c.config, id, req)
	if err != nil {
		return nil, fmt.Errorf("update flow: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateFlowJSON(
	ctx context.Context,
	req *flow.UpdateFlowJSONRequest,
) (*flow.UpdateFlowJSONResponse, error) {
	resp, err := c.sender.UpdateFlowJSON(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update flow json: %w", err)
	}
	return resp, nil
}

func (c *Client) ListFlows(ctx context.Context) (*flow.ListResponse, error) {
	resp, err := c.sender.ListFlows(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	return resp, nil
}

func (c *Client) ListFlowAssets(ctx context.Context, id string) (*flow.RetrieveAssetsResponse, error) {
	resp, err := c.sender.ListFlowAssets(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("list flow assets: %w", err)
	}
	return resp, nil
}

func (c *Client) PublishFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.PublishFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("publish flow: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.DeleteFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("delete flow: %w", err)
	}
	return resp, nil
}

func (c *Client) DeprecateFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.DeprecateFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("deprecate flow: %w", err)
	}
	return resp, nil
}

func (c *Client) GetFlow(ctx context.Context, req *flow.GetRequest) (*flow.SingleFlowResponse, error) {
	resp, err := c.sender.GetFlow(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}
	return resp, nil
}

func (c *Client) GenerateFlowPreview(ctx context.Context, req *flow.PreviewRequest) (*flow.PreviewResponse, error) {
	resp, err := c.sender.GenerateFlowPreview(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("generate flow preview: %w", err)
	}
	return resp, nil
}

func (c *Client) UploadMedia(ctx context.Context, req *media.UploadRequest) (*media.UploadMediaResponse, error) {
	resp, err := c.sender.UploadMedia(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}
	return resp, nil
}

func (c *Client) GetMediaInfo(ctx context.Context, req *media.BaseRequest) (*media.Information, error) {
	resp, err := c.sender.GetMediaInfo(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get media info: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteMedia(ctx context.Context, req *media.BaseRequest) (*media.DeleteMediaResponse, error) {
	resp, err := c.sender.DeleteMedia(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}
	return resp, nil
}

func (c *Client) DownloadMedia(ctx context.Context, req *media.DownloadRequest, decoder whttp.ResponseDecoder) error {
	return c.sender.DownloadMedia(ctx, c.config, req, decoder)
}

func (c *Client) DownloadMediaByID(
	ctx context.Context,
	req *media.BaseRequest,
	decoder whttp.ResponseDecoder,
	opts ...media.DownloadOptionFunc,
) error {
	return c.sender.DownloadMediaByID(ctx, c.config, req, decoder, opts...)
}

// --- Groups ---

func (c *Client) CreateGroup(ctx context.Context, req *groups.CreateGroupRequest) (*groups.BaseResponse, error) {
	resp, err := c.sender.CreateGroup(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteGroup(ctx context.Context, req *groups.DeleteGroupRequest) (*groups.BaseResponse, error) {
	resp, err := c.sender.DeleteGroup(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("delete group: %w", err)
	}
	return resp, nil
}

func (c *Client) GetGroupInfo(ctx context.Context, req *groups.GetGroupInfoRequest) (*groups.GroupInfoResponse, error) {
	resp, err := c.sender.GetGroupInfo(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get group info: %w", err)
	}
	return resp, nil
}

func (c *Client) GetGroupInviteLink(
	ctx context.Context,
	req *groups.GetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := c.sender.GetGroupInviteLink(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get group invite link: %w", err)
	}
	return resp, nil
}

func (c *Client) ResetGroupInviteLink(
	ctx context.Context,
	req *groups.ResetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := c.sender.ResetGroupInviteLink(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("reset group invite link: %w", err)
	}
	return resp, nil
}

func (c *Client) RemoveGroupParticipants(
	ctx context.Context,
	req *groups.RemoveGroupParticipantsRequest,
) (*groups.BaseResponse, error) {
	resp, err := c.sender.RemoveGroupParticipants(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("remove group participants: %w", err)
	}
	return resp, nil
}

func (c *Client) ListActiveGroups(
	ctx context.Context,
	req *groups.GetActiveGroupsRequest,
) (*groups.ActiveGroupsResponse, error) {
	resp, err := c.sender.ListActiveGroups(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateGroupSettings(
	ctx context.Context,
	req *groups.UpdateGroupSettingsRequest,
) (*groups.BaseResponse, error) {
	resp, err := c.sender.UpdateGroupSettings(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update group settings: %w", err)
	}
	return resp, nil
}

func (c *Client) ListJoinRequests(
	ctx context.Context,
	req *groups.GetJoinRequestsRequest,
) (*groups.JoinRequestsResponse, error) {
	resp, err := c.sender.ListJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return resp, nil
}

func (c *Client) ApproveJoinRequests(
	ctx context.Context,
	req *groups.ApproveJoinRequestsRequest,
) (*groups.ApproveJoinRequestsResponse, error) {
	resp, err := c.sender.ApproveJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("approve join requests: %w", err)
	}
	return resp, nil
}

func (c *Client) RejectJoinRequests(
	ctx context.Context,
	req *groups.RejectJoinRequestsRequest,
) (*groups.RejectJoinRequestsResponse, error) {
	resp, err := c.sender.RejectJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("reject join requests: %w", err)
	}
	return resp, nil
}

// --- Business ---

func (c *Client) GetBusinessProfile(ctx context.Context, fields []string) ([]*business.Profile, error) {
	resp, err := c.sender.GetBusinessProfile(ctx, c.config, fields)
	if err != nil {
		return nil, fmt.Errorf("get business profile: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateBusinessProfile(ctx context.Context, req *business.UpdateProfileRequest) (bool, error) {
	resp, err := c.sender.UpdateBusinessProfile(ctx, c.config, req)
	if err != nil {
		return false, fmt.Errorf("update business profile: %w", err)
	}
	return resp, nil
}

// --- Analytics ---

func (c *Client) FetchMessagingAnalytics(
	ctx context.Context,
	req *analytics.MessagingRequest,
) (*analytics.MessagingResponse, error) {
	resp, err := c.sender.FetchMessagingAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch messaging analytics: %w", err)
	}
	return resp, nil
}

func (c *Client) FetchConversationAnalytics(
	ctx context.Context,
	req *analytics.ConversationalRequest,
) (*analytics.ConversationalResponse, error) {
	resp, err := c.sender.FetchConversationAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}
	return resp, nil
}

func (c *Client) FetchPricingAnalytics(
	ctx context.Context,
	req *analytics.PricingRequest,
) (*analytics.PricingResponse, error) {
	resp, err := c.sender.FetchPricingAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch pricing analytics: %w", err)
	}
	return resp, nil
}

// --- Uploads ---

func (c *Client) InitUploadSession(
	ctx context.Context,
	req *uploads.InitUploadSessionRequest,
) (*uploads.InitUploadSessionResponse, error) {
	resp, err := c.sender.InitUploadSession(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("init upload session: %w", err)
	}
	return resp, nil
}

func (c *Client) UploadChunk(
	ctx context.Context,
	req *uploads.UploadChunkRequest,
) (*uploads.UploadChunkResponse, error) {
	resp, err := c.sender.UploadChunk(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("upload chunk: %w", err)
	}
	return resp, nil
}

func (c *Client) GetUploadStatus(ctx context.Context, uploadSessionID string) (*uploads.UploadStatusResponse, error) {
	resp, err := c.sender.GetUploadStatus(ctx, c.config, uploadSessionID)
	if err != nil {
		return nil, fmt.Errorf("get upload status: %w", err)
	}
	return resp, nil
}

// --- Auth (System User) ---

// InstallApp installs an app for a system user.
func (c *Client) InstallApp(ctx context.Context, params *auth.InstallAppParams) error {
	return c.sender.InstallApp(ctx, c.config, params)
}

// GenerateSystemUserToken generates a persistent access token for a system user.
func (c *Client) GenerateSystemUserToken(
	ctx context.Context,
	params *auth.GenerateAccessTokenParams,
) (*auth.GenerateAccessTokenResponse, error) {
	resp, err := c.sender.GenerateAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("generate system user token: %w", err)
	}
	return resp, nil
}

// RevokeSystemUserToken revokes a system user access token.
func (c *Client) RevokeSystemUserToken(
	ctx context.Context,
	params *auth.RevokeAccessTokenParams,
) (*auth.RevokeAccessTokenResponse, error) {
	resp, err := c.sender.RevokeAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("revoke system user token: %w", err)
	}
	return resp, nil
}

// RefreshSystemUserToken refreshes an expiring system user access token.
func (c *Client) RefreshSystemUserToken(
	ctx context.Context,
	params *auth.RefreshAccessTokenParams,
) (*auth.RefreshAccessTokenResponse, error) {
	resp, err := c.sender.RefreshAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("refresh system user token: %w", err)
	}
	return resp, nil
}

// TwoStepVerification sets up two-step verification for a WhatsApp Business API phone number.
func (c *Client) TwoStepVerification(
	ctx context.Context,
	request *auth.TwoStepVerificationRequest,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.TwoStepVerification(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("two step verification: %w", err)
	}
	return resp, nil
}

// CreateSystemUser creates a system user in a business manager.
func (c *Client) CreateSystemUser(
	ctx context.Context,
	req *auth.CreateSystemUserRequest,
) (*auth.CreateSystemUserResponse, error) {
	resp, err := c.sender.CreateSystemUser(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}
	return resp, nil
}

// ListSystemUsers retrieves all system users in a business manager.
func (c *Client) ListSystemUsers(ctx context.Context) (*auth.ListSystemUsersResponse, error) {
	resp, err := c.sender.ListSystemUsers(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}
	return resp, nil
}

// UpdateSystemUser updates the name of an existing system user.
func (c *Client) UpdateSystemUser(
	ctx context.Context,
	req *auth.UpdateSystemUserRequest,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.UpdateSystemUser(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}
	return resp, nil
}

// InvalidateSystemUserTokens invalidates all access tokens for a system user.
func (c *Client) InvalidateSystemUserTokens(
	ctx context.Context,
	systemUserID string,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.InvalidateSystemUserTokens(ctx, c.config, systemUserID)
	if err != nil {
		return nil, fmt.Errorf("invalidate system user tokens: %w", err)
	}
	return resp, nil
}

// SetAlternativeCallback configures an alternate callback URL for a WABA or
// phone number. After setting, supported webhook fields route to the alternate
// URL instead of the app's default callback.
func (c *Client) SetAlternativeCallback(
	ctx context.Context,
	request *callbacks.SetAlternativeCallbackRequest,
) (*callbacks.SuccessResponse, error) {
	resp, err := c.sender.SetAlternativeCallback(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("set alternative callback: %w", err)
	}
	return resp, nil
}

// DeleteAlternativeCallback removes the alternate callback URL for the given
// OverrideType. After deletion, webhooks fall back to the next priority level
// (phone number → WABA → app default).
func (c *Client) DeleteAlternativeCallback(
	ctx context.Context,
	overrideType callbacks.OverrideType,
) (*callbacks.SuccessResponse, error) {
	resp, err := c.sender.DeleteAlternativeCallback(ctx, c.config, overrideType)
	if err != nil {
		return nil, fmt.Errorf("delete alternative callback: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CheckCallingPermission(
	ctx context.Context,
	conf *config.Config,
	req *calls.CheckPermissionRequest,
) (*calls.CallPermissionCheckResponse, error) {
	resp, err := bc.calls.CheckPermission(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("check permission: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateCallStatus(
	ctx context.Context,
	conf *config.Config,
	req *calls.CallUpdateStatusRequest,
) (*calls.CallUpdateStatusResponse, error) {
	resp, err := bc.calls.UpdateCallStatus(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update call status: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) BlockUsers(
	ctx context.Context,
	conf *config.Config,
	req *user.BlockRequest,
) (*user.BlockResponse, error) {
	resp, err := bc.users.Block(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("block: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UnblockUsers(
	ctx context.Context,
	conf *config.Config,
	req *user.UnblockRequest,
) (*user.BlockResponse, error) {
	resp, err := bc.users.Unblock(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("unblock: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListBlockedUsers(
	ctx context.Context,
	conf *config.Config,
	opts *user.ListBlockedUsersOptions,
) (*user.ListBlockedUsersResponse, error) {
	resp, err := bc.users.ListBlocked(ctx, conf, opts)
	if err != nil {
		return nil, fmt.Errorf("list blocked: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateQR(
	ctx context.Context,
	conf *config.Config,
	req *qrcode.CreateRequest,
) (*qrcode.CreateResponse, error) {
	resp, err := bc.qrCode.Create(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetQR(ctx context.Context, conf *config.Config, qrCodeID string) (*qrcode.Information, error) {
	resp, err := bc.qrCode.Get(ctx, conf, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListQR(
	ctx context.Context,
	conf *config.Config,
	opts *qrcode.ListOptions,
) (*qrcode.ListResponse, error) {
	resp, err := bc.qrCode.List(ctx, conf, opts)
	if err != nil {
		return nil, fmt.Errorf("list qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteQR(
	ctx context.Context,
	conf *config.Config,
	qrCodeID string,
) (*qrcode.SuccessResponse, error) {
	resp, err := bc.qrCode.Delete(ctx, conf, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("delete qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateQR(
	ctx context.Context,
	conf *config.Config,
	req *qrcode.UpdateRequest,
) (*qrcode.SuccessResponse, error) {
	resp, err := bc.qrCode.Update(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) AddConversationComponents(ctx context.Context, conf *config.Config,
	commands []*automation.Command, prompts []string,
) (*automation.SuccessResponse, error) {
	resp, err := bc.auto.AddConversationComponents(ctx, conf, commands, prompts)
	if err != nil {
		return nil, fmt.Errorf("add components: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateWelcomeMessageStatus(ctx context.Context, conf *config.Config,
	shouldEnable bool,
) (*automation.SuccessResponse, error) {
	resp, err := bc.auto.UpdateWelcomeMessageStatus(ctx, conf, shouldEnable)
	if err != nil {
		return nil, fmt.Errorf("update welcome message status: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListConversationComponents(
	ctx context.Context,
	conf *config.Config,
) (*automation.BaseResponse, error) {
	resp, err := bc.auto.ListConversationComponents(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetBotDetails(
	ctx context.Context,
	conf *config.Config,
	request *automation.BotRequest,
) (*automation.Bot, error) {
	resp, err := bc.auto.GetBotDetails(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get bot details: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetSettings(
	ctx context.Context,
	conf *config.Config,
	req *settings.GetSettingsRequest,
) (*settings.Settings, error) {
	resp, err := bc.settings.GetSettings(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateSettings(
	ctx context.Context,
	conf *config.Config,
	s *settings.Settings,
) (*settings.SuccessResponse, error) {
	resp, err := bc.settings.UpdateSettings(ctx, conf, s)
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListPhoneNumbers(ctx context.Context, conf *config.Config) (*phonenumber.ListResponse, error) {
	req := &phonenumber.Request{
		RequestType: whttp.RequestTypeListPhoneNumbers,
		QueryParams: map[string]string{},
	}
	resp, err := bc.phone.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list phone numbers: %w", err)
	}
	return resp.ListPhoneNumbersResponse(), nil
}

func (bc *BaseClient) GetPhoneNumber(
	ctx context.Context,
	conf *config.Config,
	req *phonenumber.GetRequest,
) (*phonenumber.PhoneNumber, error) {
	request := &phonenumber.Request{
		RequestType: whttp.RequestTypeGetPhoneNumber,
		QueryParams: map[string]string{
			"fields": strings.Join(req.Fields, ";"),
		},
	}
	resp, err := bc.phone.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}
	return resp.PhoneNumber(), nil
}

func (bc *BaseClient) CreateFlow(
	ctx context.Context,
	conf *config.Config,
	req flow.CreateRequest,
) (*flow.CreateResponse, error) {
	resp, err := bc.flows.Create(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateFlow(
	ctx context.Context,
	conf *config.Config,
	id string,
	req flow.UpdateRequest,
) (*flow.UpdateResponse, error) {
	resp, err := bc.flows.Update(ctx, conf, id, req)
	if err != nil {
		return nil, fmt.Errorf("update flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateFlowJSON(
	ctx context.Context,
	conf *config.Config,
	req *flow.UpdateFlowJSONRequest,
) (*flow.UpdateFlowJSONResponse, error) {
	resp, err := bc.flows.UpdateFlowJSON(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update flow json: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListFlows(ctx context.Context, conf *config.Config) (*flow.ListResponse, error) {
	resp, err := bc.flows.ListAll(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListFlowAssets(
	ctx context.Context,
	conf *config.Config,
	id string,
) (*flow.RetrieveAssetsResponse, error) {
	resp, err := bc.flows.ListAssets(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("list flow assets: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) PublishFlow(ctx context.Context, conf *config.Config, id string) (*flow.SuccessResponse, error) {
	resp, err := bc.flows.Publish(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("publish flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteFlow(ctx context.Context, conf *config.Config, id string) (*flow.SuccessResponse, error) {
	resp, err := bc.flows.Delete(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("delete flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeprecateFlow(
	ctx context.Context,
	conf *config.Config,
	id string,
) (*flow.SuccessResponse, error) {
	resp, err := bc.flows.Deprecate(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("deprecate flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetFlow(
	ctx context.Context,
	conf *config.Config,
	req *flow.GetRequest,
) (*flow.SingleFlowResponse, error) {
	resp, err := bc.flows.Get(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GenerateFlowPreview(
	ctx context.Context,
	conf *config.Config,
	req *flow.PreviewRequest,
) (*flow.PreviewResponse, error) {
	resp, err := bc.flows.GeneratePreview(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("generate flow preview: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UploadMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.UploadRequest,
) (*media.UploadMediaResponse, error) {
	resp, err := bc.media.Upload(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetMediaInfo(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
) (*media.Information, error) {
	resp, err := bc.media.GetInfo(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get media info: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
) (*media.DeleteMediaResponse, error) {
	resp, err := bc.media.Delete(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DownloadMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.DownloadRequest,
	decoder whttp.ResponseDecoder,
) error {
	if err := bc.media.Download(ctx, conf, req, decoder); err != nil {
		return fmt.Errorf("download media: %w", err)
	}
	return nil
}

func (bc *BaseClient) DownloadMediaByID(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
	decoder whttp.ResponseDecoder,
	opts ...media.DownloadOptionFunc,
) error {
	if err := bc.media.DownloadByMediaID(ctx, conf, req, decoder, opts...); err != nil {
		return fmt.Errorf("download media by id: %w", err)
	}
	return nil
}

// --- Groups BaseClient ---

func (bc *BaseClient) CreateGroup(
	ctx context.Context,
	conf *config.Config,
	req *groups.CreateGroupRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.groups.CreateGroup(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteGroup(
	ctx context.Context,
	conf *config.Config,
	req *groups.DeleteGroupRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.groups.DeleteGroup(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete group: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetGroupInfo(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetGroupInfoRequest,
) (*groups.GroupInfoResponse, error) {
	resp, err := bc.groups.GetGroupInfo(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get group info: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := bc.groups.GetGroupInviteLink(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get group invite link: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ResetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *groups.ResetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := bc.groups.ResetGroupInviteLink(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("reset group invite link: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RemoveGroupParticipants(
	ctx context.Context,
	conf *config.Config,
	req *groups.RemoveGroupParticipantsRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.groups.RemoveGroupParticipants(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("remove group participants: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListActiveGroups(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetActiveGroupsRequest,
) (*groups.ActiveGroupsResponse, error) {
	resp, err := bc.groups.GetActiveGroups(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateGroupSettings(
	ctx context.Context,
	conf *config.Config,
	req *groups.UpdateGroupSettingsRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.groups.UpdateGroupSettings(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update group settings: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetJoinRequestsRequest,
) (*groups.JoinRequestsResponse, error) {
	resp, err := bc.groups.GetJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ApproveJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.ApproveJoinRequestsRequest,
) (*groups.ApproveJoinRequestsResponse, error) {
	resp, err := bc.groups.ApproveJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("approve join requests: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RejectJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.RejectJoinRequestsRequest,
) (*groups.RejectJoinRequestsResponse, error) {
	resp, err := bc.groups.RejectJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("reject join requests: %w", err)
	}
	return resp, nil
}

// --- Business BaseClient ---

func (bc *BaseClient) GetBusinessProfile(
	ctx context.Context,
	conf *config.Config,
	fields []string,
) ([]*business.Profile, error) {
	resp, err := bc.biz.Get(ctx, conf, fields)
	if err != nil {
		return nil, fmt.Errorf("get business profile: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateBusinessProfile(
	ctx context.Context,
	conf *config.Config,
	req *business.UpdateProfileRequest,
) (bool, error) {
	resp, err := bc.biz.Update(ctx, conf, req)
	if err != nil {
		return false, fmt.Errorf("update business profile: %w", err)
	}
	return resp, nil
}

// --- Analytics BaseClient ---

func (bc *BaseClient) FetchMessagingAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.MessagingRequest,
) (*analytics.MessagingResponse, error) {
	resp, err := bc.analytics.FetchGeneralAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch messaging analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchConversationAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.ConversationalRequest,
) (*analytics.ConversationalResponse, error) {
	resp, err := bc.analytics.FetchConversationAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchPricingAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.PricingRequest,
) (*analytics.PricingResponse, error) {
	resp, err := bc.analytics.FetchPricingAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch pricing analytics: %w", err)
	}
	return resp, nil
}

// --- Uploads BaseClient ---

func (bc *BaseClient) InitUploadSession(
	ctx context.Context,
	conf *config.Config,
	req *uploads.InitUploadSessionRequest,
) (*uploads.InitUploadSessionResponse, error) {
	resp, err := bc.uploads.InitUploadSession(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("init upload session: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UploadChunk(
	ctx context.Context,
	conf *config.Config,
	req *uploads.UploadChunkRequest,
) (*uploads.UploadChunkResponse, error) {
	resp, err := bc.uploads.UploadChunk(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("upload chunk: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetUploadStatus(
	ctx context.Context,
	conf *config.Config,
	uploadSessionID string,
) (*uploads.UploadStatusResponse, error) {
	resp, err := bc.uploads.GetUploadStatus(ctx, conf, uploadSessionID)
	if err != nil {
		return nil, fmt.Errorf("get upload status: %w", err)
	}
	return resp, nil
}

// --- Auth BaseClient ---

// InstallApp installs an app for a system user.
func (bc *BaseClient) InstallApp(
	ctx context.Context,
	conf *config.Config,
	params *auth.InstallAppParams,
) error {
	if err := bc.auth.InstallApp(ctx, conf, params); err != nil {
		return fmt.Errorf("install app: %w", err)
	}
	return nil
}

// GenerateAccessToken generates a persistent access token for a system user.
func (bc *BaseClient) GenerateAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.GenerateAccessTokenParams,
) (*auth.GenerateAccessTokenResponse, error) {
	resp, err := bc.auth.GenerateAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	return resp, nil
}

// RevokeAccessToken revokes a system user access token.
func (bc *BaseClient) RevokeAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.RevokeAccessTokenParams,
) (*auth.RevokeAccessTokenResponse, error) {
	resp, err := bc.auth.RevokeAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("revoke access token: %w", err)
	}
	return resp, nil
}

// RefreshAccessToken refreshes an expiring system user access token.
func (bc *BaseClient) RefreshAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.RefreshAccessTokenParams,
) (*auth.RefreshAccessTokenResponse, error) {
	resp, err := bc.auth.RefreshAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}
	return resp, nil
}

// TwoStepVerification sets up two-step verification for a WhatsApp Business API phone number.
func (bc *BaseClient) TwoStepVerification(
	ctx context.Context,
	conf *config.Config,
	request *auth.TwoStepVerificationRequest,
) (*auth.SuccessResponse, error) {
	resp, err := bc.auth.TwoStepVerification(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("two step verification: %w", err)
	}
	return resp, nil
}

// CreateSystemUser creates a system user in a business manager.
func (bc *BaseClient) CreateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *auth.CreateSystemUserRequest,
) (*auth.CreateSystemUserResponse, error) {
	resp, err := bc.auth.CreateSystemUser(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}
	return resp, nil
}

// ListSystemUsers retrieves all system users in a business manager.
func (bc *BaseClient) ListSystemUsers(
	ctx context.Context,
	conf *config.Config,
) (*auth.ListSystemUsersResponse, error) {
	resp, err := bc.auth.ListSystemUsers(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}
	return resp, nil
}

// UpdateSystemUser updates the name of an existing system user.
func (bc *BaseClient) UpdateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *auth.UpdateSystemUserRequest,
) (*auth.SuccessResponse, error) {
	resp, err := bc.auth.UpdateSystemUser(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}
	return resp, nil
}

// InvalidateSystemUserTokens invalidates all access tokens for a system user.
func (bc *BaseClient) InvalidateSystemUserTokens(
	ctx context.Context,
	conf *config.Config,
	systemUserID string,
) (*auth.SuccessResponse, error) {
	resp, err := bc.auth.InvalidateSystemUserTokens(ctx, conf, systemUserID)
	if err != nil {
		return nil, fmt.Errorf("invalidate system user tokens: %w", err)
	}
	return resp, nil
}

// SetAlternativeCallback configures an alternate callback URL for a WABA or
// phone number.
func (bc *BaseClient) SetAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	request *callbacks.SetAlternativeCallbackRequest,
) (*callbacks.SuccessResponse, error) {
	resp, err := bc.callbacks.SetAlternativeCallback(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("set alternative callback: %w", err)
	}
	return resp, nil
}

// DeleteAlternativeCallback removes the alternate callback URL for the given
// OverrideType.
func (bc *BaseClient) DeleteAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	overrideType callbacks.OverrideType,
) (*callbacks.SuccessResponse, error) {
	resp, err := bc.callbacks.DeleteAlternativeCallback(ctx, conf, overrideType)
	if err != nil {
		return nil, fmt.Errorf("delete alternative callback: %w", err)
	}
	return resp, nil
}
