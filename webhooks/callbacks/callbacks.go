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

package callbacks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const EndpointSubscribedApps = "/subscribed_apps"

type OverrideType string

const (
	OverrideTypeWABA        OverrideType = "waba"
	OverrideTypePhoneNumber OverrideType = "phone_number"
)

type (
	BaseClient struct {
		BaseSender whttp.Sender[BaseRequest]
	}

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

func (bs *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*BaseResponse, error) {
	req := &whttp.Request[BaseRequest]{}
	switch request.RequestType { //nolint: exhaustive // we have only 4 request types
	case whttp.RequestTypeSetWABAAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodPost, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestHeaders[BaseRequest](map[string]string{
				"Content-Type": "application/json",
			}),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.BusinessAccountID, EndpointSubscribedApps),
			whttp.WithRequestMessage(&BaseRequest{
				WebhookConfiguration: &WebhookConfiguration{
					OverrideCallbackURI: request.OverrideCallbackURI,
					VerifyToken:         request.VerifyToken,
				},
			}),
		)

	case whttp.RequestTypeSetPhoneNumberAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodPost, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestHeaders[BaseRequest](map[string]string{
				"Content-Type": "application/json",
			}),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.BusinessAccountID, EndpointSubscribedApps),
			whttp.WithRequestMessage(&BaseRequest{
				OverrideCallbackURI: request.OverrideCallbackURI,
				VerifyToken:         request.VerifyToken,
			}),
		)

	case whttp.RequestTypeGetWABAAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodGet, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.BusinessAccountID, EndpointSubscribedApps),
		)

	case whttp.RequestTypeGetPhoneNumberAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodGet, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.PhoneNumberID, EndpointSubscribedApps),
			whttp.WithRequestQueryParams[BaseRequest](map[string]string{
				"fields": "webhook_configuration",
			}),
		)

	case whttp.RequestTypeDeleteWABAAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodPost, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.BusinessAccountID, EndpointSubscribedApps),
		)

	case whttp.RequestTypeDeletePhoneNumberAlternateCallbackURI:
		req = whttp.MakeRequest(http.MethodPost, conf.BaseURL,
			whttp.WithRequestType[BaseRequest](request.RequestType),
			whttp.WithRequestAppSecret[BaseRequest](conf.AppSecret),
			whttp.WithRequestSecured[BaseRequest](conf.SecureRequests),
			whttp.WithRequestEndpoints[BaseRequest](conf.APIVersion, conf.PhoneNumberID, EndpointSubscribedApps),
		)
	}

	response := &BaseResponse{}
	err := bs.BaseSender.Send(ctx, req, whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	}))
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
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
		BaseClient *BaseClient
		Config     *config.Config
	}

	SetAlternativeCallbackRequest struct {
		OverrideCallbackURI string `json:"override_callback_uri"`
		VerifyToken         string `json:"verify_token"`
		OverrideType        OverrideType
	}
)

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

	res, err := c.BaseClient.Send(ctx, c.Config, req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res.SuccessResponse(), nil
}

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

	res, err := c.BaseClient.Send(ctx, c.Config, req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res.SuccessResponse(), nil
}
