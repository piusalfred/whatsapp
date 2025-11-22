/*
 * MIT License
 *
 * Copyright (c) 2025 Pius Alfred
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package automation

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	Command struct {
		CommandName        string `json:"command_name,omitempty"`
		CommandDescription string `json:"command_description,omitempty"`
	}

	CommandParams struct {
		Name        string
		Description string
		Prompt      string
	}

	Request struct {
		EnableWelcomeMessage bool       `json:"enable_welcome_message"`
		Commands             []*Command `json:"commands,omitempty"`
		Prompts              []string   `json:"prompts,omitempty"`
	}

	Response struct {
		ID                       string   `json:"id,omitempty"`
		ConversationalAutomation *Request `json:"conversational_automation,omitempty"`
		Success                  bool     `json:"success,omitempty"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}
)

func NewRequest(commands []*Command, prompts []string) *Request {
	req := &Request{
		EnableWelcomeMessage: true,
		Commands:             commands,
		Prompts:              prompts,
	}

	return req
}

const Endpoint = "conversational_automation"

type BaseClient struct {
	Sender whttp.Sender[Request]
}

func (client *BaseClient) AddComponents(ctx context.Context, reader config.Reader,
	commands []*Command, prompts []string,
) (*SuccessResponse, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	message := NewRequest(commands, prompts)

	options := []whttp.RequestOption[Request]{
		whttp.WithRequestEndpoints[Request](conf.APIVersion, conf.PhoneNumberID, Endpoint),
		whttp.WithRequestBearer[Request](conf.AccessToken),
		whttp.WithRequestAppSecret[Request](conf.AppSecret),
		whttp.WithRequestSecured[Request](conf.SecureRequests),
		whttp.WithRequestMessage(message),
		whttp.WithRequestType[Request](whttp.RequestTypeUpdateConversationAutomationComponents),
		whttp.WithRequestDebugLogLevel[Request](whttp.ParseDebugLogLevel(conf.DebugLogLevel)),
	}

	request := whttp.MakeRequest(http.MethodPost, conf.BaseURL, options...)

	response := &SuccessResponse{}

	if err = client.Sender.Send(ctx, request, whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (client *BaseClient) UpdateWelcomeMessageStatus(ctx context.Context, reader config.Reader,
	shouldEnable bool,
) (*SuccessResponse, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var requestTypeOption whttp.RequestOption[Request]

	if shouldEnable {
		requestTypeOption = whttp.WithRequestType[Request](whttp.RequestTypeEnableWelcomeMessage)
	} else {
		requestTypeOption = whttp.WithRequestType[Request](whttp.RequestTypeDisableWelcomeMessage)
	}

	options := []whttp.RequestOption[Request]{
		whttp.WithRequestEndpoints[Request](conf.APIVersion, conf.PhoneNumberID, Endpoint),
		whttp.WithRequestBearer[Request](conf.AccessToken),
		whttp.WithRequestAppSecret[Request](conf.AppSecret),
		whttp.WithRequestSecured[Request](conf.SecureRequests),
		whttp.WithRequestQueryParams[Request](map[string]string{
			"enable_welcome_message": strconv.FormatBool(shouldEnable),
		}),
		whttp.WithRequestDebugLogLevel[Request](whttp.ParseDebugLogLevel(conf.DebugLogLevel)),
		requestTypeOption,
	}

	request := whttp.MakeRequest(http.MethodPost, conf.BaseURL, options...)

	response := &SuccessResponse{}

	if err = client.Sender.Send(ctx, request, whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (client *BaseClient) ListComponents(ctx context.Context, reader config.Reader) (*Response, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	options := []whttp.RequestOption[Request]{
		whttp.WithRequestEndpoints[Request](conf.APIVersion, conf.PhoneNumberID, Endpoint),
		whttp.WithRequestBearer[Request](conf.AccessToken),
		whttp.WithRequestAppSecret[Request](conf.AppSecret),
		whttp.WithRequestSecured[Request](conf.SecureRequests),
		whttp.WithRequestDebugLogLevel[Request](whttp.ParseDebugLogLevel(conf.DebugLogLevel)),
		whttp.WithRequestType[Request](whttp.RequestTypeGetConversationAutomationComponents),
	}

	request := whttp.MakeRequest(http.MethodGet, conf.BaseURL, options...)

	response := &Response{}

	if err = client.Sender.Send(ctx, request, whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

type Client struct {
	Reader config.Reader
	Base   *BaseClient
}

func (c *Client) ListComponents(ctx context.Context) (*Response, error) {
	return c.Base.ListComponents(ctx, c.Reader)
}

func (c *Client) UpdateWelcomeMessageStatus(ctx context.Context, shouldEnable bool) (*SuccessResponse, error) {
	return c.Base.UpdateWelcomeMessageStatus(ctx, c.Reader, shouldEnable)
}

func (c *Client) AddComponents(ctx context.Context, commands []*Command, prompts []string) (*SuccessResponse, error) {
	return c.Base.AddComponents(ctx, c.Reader, commands, prompts)
}

type Service interface {
	ListComponents(ctx context.Context) (*Response, error)
	UpdateWelcomeMessageStatus(ctx context.Context, shouldEnable bool) (*SuccessResponse, error)
	AddComponents(ctx context.Context, commands []*Command, prompts []string) (*SuccessResponse, error)
}
