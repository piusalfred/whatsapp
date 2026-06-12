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

	// Request is an internal unified context data carrier mapping domain
	// operations down to the HTTP executor. Fields tagged `json:"-"` are
	// routing metadata and are not serialized.
	Request struct {
		EnableWelcomeMessage bool              `json:"enable_welcome_message,omitempty"`
		Commands             []*Command        `json:"commands,omitempty"`
		Prompts              []string          `json:"prompts,omitempty"`
		RequestType          whttp.RequestType `json:"-"`
	}

	// BaseRequest is the wire-format payload sent to the Conversational Automation API.
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
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.Sender = sender
}

// SetMiddlewares configures middlewares that wrap the underlying Sender.
// Middlewares are applied to the sender's Send method in the order provided.
// If a custom sender has been injected and does not support middleware
// configuration internally, the configuration is silently discarded.
// Apply middlewares to your custom sender before injecting it.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.Sender = whttp.WrapMiddlewareSender(c.sender.Sender, mws...)
}

// Send dispatches a Request through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}

func (c *Client) AddComponents(ctx context.Context, commands []*Command, prompts []string) (*SuccessResponse, error) {
	request := &Request{
		EnableWelcomeMessage: true,
		Commands:             commands,
		Prompts:              prompts,
		RequestType:          whttp.RequestTypeUpdateConversationAutomationComponents,
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: resp.Success}, nil
}

func (c *Client) UpdateWelcomeMessageStatus(ctx context.Context, shouldEnable bool) (*SuccessResponse, error) {
	var requestType whttp.RequestType
	if shouldEnable {
		requestType = whttp.RequestTypeEnableWelcomeMessage
	} else {
		requestType = whttp.RequestTypeDisableWelcomeMessage
	}

	request := &Request{
		EnableWelcomeMessage: shouldEnable,
		RequestType:          requestType,
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: resp.Success}, nil
}

func (c *Client) ListComponents(ctx context.Context) (*BaseResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetConversationAutomationComponents,
	}

	return c.Send(ctx, request)
}

// BaseClient is the low-level HTTP executor for the Conversational Automation API.
type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

// Send translates a Request into an HTTP transaction and returns the decoded BaseResponse.
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	var (
		method      string
		queryParams map[string]string
		message     *BaseRequest
	)

	switch request.RequestType {
	case whttp.RequestTypeUpdateConversationAutomationComponents:
		method = http.MethodPost
		message = &BaseRequest{
			EnableWelcomeMessage: request.EnableWelcomeMessage,
			Commands:             request.Commands,
			Prompts:              request.Prompts,
		}
	case whttp.RequestTypeEnableWelcomeMessage, whttp.RequestTypeDisableWelcomeMessage:
		method = http.MethodPost
		queryParams = map[string]string{
			"enable_welcome_message": strconv.FormatBool(request.EnableWelcomeMessage),
		}
	case whttp.RequestTypeGetConversationAutomationComponents:
		method = http.MethodGet
	default:
		return nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}

	b := whttp.NewRequestBuilder(method, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.RequestType).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint)

	if len(queryParams) > 0 {
		b = b.QueryParams(queryParams)
	}

	req := whttp.BuildRequest(b, message)

	response := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (bc *BaseClient) AddComponents(ctx context.Context, conf *config.Config,
	commands []*Command, prompts []string,
) (*SuccessResponse, error) {
	request := &Request{
		EnableWelcomeMessage: true,
		Commands:             commands,
		Prompts:              prompts,
		RequestType:          whttp.RequestTypeUpdateConversationAutomationComponents,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: resp.Success}, nil
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

	request := &Request{
		EnableWelcomeMessage: shouldEnable,
		RequestType:          requestType,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: resp.Success}, nil
}

func (bc *BaseClient) ListComponents(ctx context.Context, conf *config.Config) (*BaseResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetConversationAutomationComponents,
	}

	return bc.Send(ctx, conf, request)
}
