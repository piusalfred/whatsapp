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

	"github.com/piusalfred/whatsapp/calls"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/conversation/automation"
	"github.com/piusalfred/whatsapp/phonenumber"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/qrcode"
	"github.com/piusalfred/whatsapp/settings"
	"github.com/piusalfred/whatsapp/user"
)

// Client wraps BaseClient with a fixed configuration.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// BaseClient is the multi-tenant layer. Pass a *config.Config per call.
type BaseClient struct {
	calls    *calls.BaseClient
	users    *user.BlockBaseClient
	qrCode   *qrcode.BaseClient
	auto     *automation.BaseClient
	settings *settings.BaseClient
	phone    *phonenumber.BaseClient
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
		calls:    &calls.BaseClient{BaseClient: *whttp.NewBaseClient[calls.BaseRequest](opts...)},
		users:    &user.BlockBaseClient{BaseClient: *whttp.NewBaseClient[user.BlockBaseRequest](opts...)},
		qrCode:   &qrcode.BaseClient{BaseClient: *whttp.NewBaseClient[qrcode.BaseRequest](opts...)},
		auto:     &automation.BaseClient{BaseClient: *whttp.NewBaseClient[automation.BaseRequest](opts...)},
		settings: &settings.BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		phone:    &phonenumber.BaseClient{BaseClient: *whttp.NewBaseClient[phonenumber.BaseRequest](opts...)},
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

func (c *Client) SetSettingsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.settings.SetMiddlewares(mws...)
}

func (c *Client) SetPhoneNumbersMiddlewares(mws ...whttp.Middleware[phonenumber.BaseRequest]) {
	c.sender.phone.SetMiddlewares(mws...)
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

func (c *Client) UpdateSettings(ctx context.Context, settings *settings.Settings) (*settings.SuccessResponse, error) {
	resp, err := c.sender.UpdateSettings(ctx, c.config, settings)
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
