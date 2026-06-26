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

// Package callbacks manages alternate webhook callback URLs for WhatsApp
// Business Accounts and phone numbers.
//
// WhatsApp supports a three-tier callback routing priority:
//  1. Phone number alternate URL (if set)
//  2. WABA alternate URL (if set)
//  3. App's default callback URL
//
// Not all webhook fields support overrides. The following are supported:
// messages, message_echoes, calls, consumer_profile, messaging_handovers,
// group_* updates, smb_message_echoes, smb_app_state_sync, history,
// account_settings_update.
//
// Template webhooks (message_template_status_update, template_category_update,
// etc.) and account-level webhooks (account_update, account_review_update,
// account_alerts) do NOT support overrides and are always delivered to the
// app's default callback URL.
package callbacks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const EndpointSubscribedApps = "/subscribed_apps"

// OverrideType specifies whether an alternate callback applies to a WABA or a phone number.
type OverrideType string

const (
	OverrideTypeWABA        OverrideType = "waba"
	OverrideTypePhoneNumber OverrideType = "phone_number"
)

type (
	BaseClient struct {
		whttp.BaseClient[BaseRequest]
	}

	// WebhookConfiguration holds the callback URL and verification token for an
	// override. When retrieving the current configuration, PhoneNumber and
	// WhatsappBusinessAccount indicate which level has an override set.
	WebhookConfiguration struct {
		OverrideCallbackURI     string `json:"override_callback_uri"`
		VerifyToken             string `json:"verify_token"`
		PhoneNumber             string `json:"phone_number"`
		WhatsappBusinessAccount string `json:"whatsapp_business_account"`
		Application             string `json:"application"`
	}

	BaseRequest struct {
		Method               string                `json:"-"`
		Type                 OverrideType          `json:"-"`
		OverrideCallbackURI  string                `json:"override_callback_uri"`
		VerifyToken          string                `json:"verify_token"`
		WebhookConfiguration *WebhookConfiguration `json:"webhook_configuration"`
		RequestType          whttp.RequestType     `json:"-"`
	}

	WhatsappBusinessAPIData struct {
		ID   string `json:"id"`
		Link string `json:"link"`
		Name string `json:"name"`
	}

	DataItem struct {
		OverrideCallbackURI     string                  `json:"override_callback_uri"`
		WhatsappBusinessAPIData WhatsappBusinessAPIData `json:"whatsapp_business_api_data"`
	}

	BaseResponse struct {
		ID                   string                `json:"id"`
		Success              bool                  `json:"success"`
		WebhookConfiguration *WebhookConfiguration `json:"webhook_configuration"`
		Data                 []DataItem            `json:"data"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}

	ListSubscribedAppsResponse struct {
		Data []DataItem `json:"data"`
	}
)

func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*BaseResponse, error) {
	var (
		method      string
		message     *BaseRequest
		queryParams map[string]string
		phoneNumber bool
	)

	switch request.RequestType { //nolint: exhaustive // we have only 4 request types
	case whttp.RequestTypeSetWABAAlternateCallbackURI:
		method = http.MethodPost
		message = &BaseRequest{
			WebhookConfiguration: &WebhookConfiguration{
				OverrideCallbackURI: request.OverrideCallbackURI,
				VerifyToken:         request.VerifyToken,
			},
		}

	case whttp.RequestTypeSetPhoneNumberAlternateCallbackURI:
		method = http.MethodPost
		message = &BaseRequest{
			OverrideCallbackURI: request.OverrideCallbackURI,
			VerifyToken:         request.VerifyToken,
		}

	case whttp.RequestTypeGetWABAAlternateCallbackURI:
		method = http.MethodGet

	case whttp.RequestTypeGetPhoneNumberAlternateCallbackURI:
		method = http.MethodGet
		queryParams = map[string]string{"fields": "webhook_configuration"}
		phoneNumber = true

	case whttp.RequestTypeDeleteWABAAlternateCallbackURI:
		method = http.MethodPost

	case whttp.RequestTypeDeletePhoneNumberAlternateCallbackURI:
		method = http.MethodPost
		phoneNumber = true
	}

	b := whttp.NewRequestBuilder(method, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.RequestType)

	if message != nil {
		b.Headers(map[string]string{"Content-Type": "application/json"})
	}

	if phoneNumber {
		b.Endpoints(conf.APIVersion, conf.PhoneNumberID, EndpointSubscribedApps)
	} else {
		b.Endpoints(conf.APIVersion, conf.BusinessAccountID, EndpointSubscribedApps)
	}

	if len(queryParams) > 0 {
		b.QueryParams(queryParams)
	}

	req := whttp.BuildRequest(b, message)

	response := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		Flags: whttp.JSONDecodeDisallowEmptyResponse | whttp.JSONDecodeInspectResponseError,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

func (bc *BaseClient) SetAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	request *SetAlternativeCallbackRequest,
) (*SuccessResponse, error) {
	req := &BaseRequest{
		Method:              http.MethodPost,
		Type:                request.OverrideType,
		VerifyToken:         request.VerifyToken,
		OverrideCallbackURI: request.OverrideCallbackURI,
	}
	if request.OverrideType == OverrideTypeWABA {
		req.RequestType = whttp.RequestTypeSetWABAAlternateCallbackURI
	}
	if request.OverrideType == OverrideTypePhoneNumber {
		req.RequestType = whttp.RequestTypeSetPhoneNumberAlternateCallbackURI
	}
	resp, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("set alternative callback: %w", err)
	}
	return resp.SuccessResponse(), nil
}

func (bc *BaseClient) DeleteAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	overrideType OverrideType,
) (*SuccessResponse, error) {
	req := &BaseRequest{
		Method: http.MethodPost,
		Type:   overrideType,
	}
	if overrideType == OverrideTypeWABA {
		req.RequestType = whttp.RequestTypeDeleteWABAAlternateCallbackURI
	}
	if overrideType == OverrideTypePhoneNumber {
		req.RequestType = whttp.RequestTypeDeletePhoneNumberAlternateCallbackURI
	}
	resp, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete alternative callback: %w", err)
	}
	return resp.SuccessResponse(), nil
}

func (res *BaseResponse) SuccessResponse() *SuccessResponse {
	return &SuccessResponse{
		Success: res.Success,
	}
}

func (res *BaseResponse) ListSubscribedAppsResponse() *ListSubscribedAppsResponse {
	return &ListSubscribedAppsResponse{
		Data: res.Data,
	}
}

type (
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	SetAlternativeCallbackRequest struct {
		OverrideCallbackURI string `json:"override_callback_uri"`
		VerifyToken         string `json:"verify_token"`
		OverrideType        OverrideType
	}
)

// NewClient creates a high-level [Client] for the Callbacks API.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.SetSender(sender)
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.SetMiddlewares(mws...)
}

// Send dispatches a raw [BaseRequest] through the underlying [BaseClient].
func (c *Client) Send(ctx context.Context, request *BaseRequest) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}

// SetAlternativeCallback configures an alternate callback URL for a WABA or
// phone number. The OverrideType determines which endpoint is used.
// After setting, supported webhook fields will be routed to the alternate URL
// instead of the app's default callback.
func (c *Client) SetAlternativeCallback(
	ctx context.Context,
	request *SetAlternativeCallbackRequest,
) (*SuccessResponse, error) {
	req := &BaseRequest{
		Method:              http.MethodPost,
		Type:                request.OverrideType,
		VerifyToken:         request.VerifyToken,
		OverrideCallbackURI: request.OverrideCallbackURI,
	}

	if request.OverrideType == OverrideTypeWABA {
		req.RequestType = whttp.RequestTypeSetWABAAlternateCallbackURI
	}

	if request.OverrideType == OverrideTypePhoneNumber {
		req.RequestType = whttp.RequestTypeSetPhoneNumberAlternateCallbackURI
	}

	res, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res.SuccessResponse(), nil
}

// DeleteAlternativeCallback removes the alternate callback URL for the given
// OverrideType. After deletion, webhooks fall back to the next priority level
// (phone number → WABA → app default).
func (c *Client) DeleteAlternativeCallback(ctx context.Context, overrideType OverrideType) (*SuccessResponse, error) {
	req := &BaseRequest{
		Method: http.MethodPost,
		Type:   overrideType,
	}

	if overrideType == OverrideTypeWABA {
		req.RequestType = whttp.RequestTypeDeleteWABAAlternateCallbackURI
	}

	if overrideType == OverrideTypePhoneNumber {
		req.RequestType = whttp.RequestTypeDeletePhoneNumberAlternateCallbackURI
	}

	res, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res.SuccessResponse(), nil
}
