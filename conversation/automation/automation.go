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

	// BaseRequest is the wire-format payload for the Conversational Automation API.
	BaseRequest struct {
		EnableWelcomeMessage bool       `json:"enable_welcome_message"`
		Commands             []*Command `json:"commands,omitempty"`
		Prompts              []string   `json:"prompts,omitempty"`
	}

	// BaseResponse is the general response for the Conversational Automation API.
	BaseResponse struct {
		ID                       string   `json:"id,omitempty"`
		ConversationalAutomation *Request `json:"conversational_automation,omitempty"`
		Success                  bool     `json:"success,omitempty"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}
)

const Endpoint = "conversational_automation"

// Client orchestrates high-level Conversational Automation API operations.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// NewClient creates a high-level Client for the Conversational Automation API.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: NewBaseClient(options...),
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.SetRequestSender(sender)
}

// SetMiddlewares configures middlewares that wrap the underlying Sender.
// Middlewares are applied to the sender's Send method in the order provided.
// If a custom sender has been injected and does not support middleware
// configuration internally, the configuration is silently discarded.
// Apply middlewares to your custom sender before injecting it.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.SetMiddlewares(mws...)
}

func (c *Client) AddComponents(ctx context.Context, commands []*Command, prompts []string) (*SuccessResponse, error) {
	return c.sender.AddComponents(ctx, c.config, commands, prompts)
}

func (c *Client) UpdateWelcomeMessageStatus(ctx context.Context, shouldEnable bool) (*SuccessResponse, error) {
	return c.sender.UpdateWelcomeMessageStatus(ctx, c.config, shouldEnable)
}

func (c *Client) ListComponents(ctx context.Context) (*BaseResponse, error) {
	return c.sender.ListComponents(ctx, c.config)
}

// BaseClient is the low-level HTTP executor for the Conversational Automation API.
type BaseClient struct {
	sender whttp.Sender[BaseRequest]
}

// NewBaseClient creates a low-level BaseClient with optional whttp.CoreSenderOption.
func NewBaseClient(options ...whttp.CoreSenderOption) *BaseClient {
	return &BaseClient{sender: whttp.NewCoreClient[BaseRequest](options...)}
}

// SetRequestSender replaces the internal sender.
func (bc *BaseClient) SetRequestSender(sender whttp.Sender[BaseRequest]) {
	bc.sender = sender
}

// SetMiddlewares configures middlewares that wrap the underlying Sender.
// Middlewares are applied to the sender's Send method in the order provided.
// If a custom sender has been injected and does not support middleware
// configuration internally, the configuration is silently discarded.
// Apply middlewares to your custom sender before injecting it.
func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	if core, ok := bc.sender.(*whttp.CoreClient[BaseRequest]); ok {
		core.SetMiddlewares(mws...)
	}
}

func (bc *BaseClient) AddComponents(ctx context.Context, conf *config.Config,
	commands []*Command, prompts []string,
) (*SuccessResponse, error) {
	message := &BaseRequest{
		EnableWelcomeMessage: true,
		Commands:             commands,
		Prompts:              prompts,
	}

	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		WithBearer(conf.AccessToken).
		WithAppSecret(conf.AppSecret, conf.SecureRequests).
		WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		WithRequestType(whttp.RequestTypeUpdateConversationAutomationComponents).
		WithEndpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint)

	req := whttp.BuildRequest(b, message)

	response := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (bc *BaseClient) UpdateWelcomeMessageStatus(ctx context.Context, conf *config.Config,
	shouldEnable bool,
) (*SuccessResponse, error) {
	var requestType whttp.RequestType
	if shouldEnable {
		requestType = whttp.RequestTypeEnableWelcomeMessage
	} else {
		requestType = whttp.RequestTypeDisableWelcomeMessage
	}

	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		WithBearer(conf.AccessToken).
		WithAppSecret(conf.AppSecret, conf.SecureRequests).
		WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		WithRequestType(requestType).
		WithEndpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint).
		WithQueryParams(map[string]string{
			"enable_welcome_message": strconv.FormatBool(shouldEnable),
		})

	req := whttp.BuildRequest(b, (*BaseRequest)(nil))

	response := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (bc *BaseClient) ListComponents(ctx context.Context, conf *config.Config) (*BaseResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		WithBearer(conf.AccessToken).
		WithAppSecret(conf.AppSecret, conf.SecureRequests).
		WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		WithRequestType(whttp.RequestTypeGetConversationAutomationComponents).
		WithEndpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint)

	req := whttp.BuildRequest(b, (*BaseRequest)(nil))

	response := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}
