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

package business

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	Endpoint = "whatsapp_business_profile"
)

type (
	Profile struct {
		About             string   `json:"about"`
		Address           string   `json:"address"`
		Description       string   `json:"description"`
		Email             string   `json:"email"`
		MessagingProduct  string   `json:"messaging_product"`
		ProfilePictureURL string   `json:"profile_picture_url"`
		Websites          []string `json:"websites"`
		Vertical          string   `json:"vertical"`
	}

	GetProfileRequest struct {
		Fields []string `json:"fields"`
	}

	UpdateProfileRequest struct {
		About                string   `json:"about,omitempty"`
		Address              string   `json:"address,omitempty"`
		Description          string   `json:"description,omitempty"`
		Email                string   `json:"email,omitempty"`
		MessagingProduct     string   `json:"messaging_product"`
		ProfilePictureHandle string   `json:"profile_picture_handle,omitempty"`
		Websites             []string `json:"websites,omitempty"`
		Vertical             string   `json:"vertical,omitempty"`
	}
)

type (
	// BaseRequest carries the details needed to execute a Business Profile API HTTP call.
	BaseRequest struct {
		Type        whttp.RequestType
		Method      string
		Payload     any
		QueryFields []string
	}

	// Response is the decoded response from the Business Profile API.
	Response struct {
		Profiles []*Profile `json:"data,omitempty"`
		Success  bool       `json:"success"`
	}
)

// Client orchestrates high-level Business Profile API operations.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// NewClient creates a high-level Client for the Business Profile API.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[any](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[any]) {
	c.sender.Sender = sender
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.Sender = whttp.WrapMiddlewareSender(c.sender.Sender, mws...)
}

// Send dispatches a raw BaseRequest through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *BaseRequest) (*Response, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("business client: %w", err)
	}

	return response, nil
}

// Get retrieves the business profile.
func (c *Client) Get(ctx context.Context, fields []string) ([]*Profile, error) {
	resp, err := c.Send(ctx, &BaseRequest{
		Type:        whttp.RequestTypeGetBusinessProfile,
		Method:      http.MethodGet,
		QueryFields: fields,
	})
	if err != nil {
		return nil, err
	}

	return resp.Profiles, nil
}

// Update updates the business profile.
func (c *Client) Update(ctx context.Context, request *UpdateProfileRequest) (bool, error) {
	resp, err := c.Send(ctx, &BaseRequest{
		Type:    whttp.RequestTypeUpdateBusinessProfile,
		Method:  http.MethodPost,
		Payload: request,
	})
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

// BaseClient is the low-level HTTP executor for the Business Profile API. It
// accepts a concrete [*config.Config] per request, making it suitable for
// multi-tenant SaaS scenarios. For a fixed-configuration client, use [Client].
type BaseClient struct {
	whttp.BaseClient[any]
}

// Send translates a high-level BaseRequest into an HTTP transaction and returns
// the decoded Response.
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error) {
	if request.QueryFields == nil {
		request.QueryFields = []string{}
	}

	params := map[string]string{}
	fields := strings.Join(request.QueryFields, ",")
	if fields != "" {
		params["fields"] = fields
	}

	b := whttp.NewRequestBuilder(request.Method, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.Type).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, Endpoint).
		QueryParams(params)

	var msg *any
	if request.Payload != nil {
		msg = &request.Payload
	}

	req := whttp.Build[any](b, msg)

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[any]) {
	bc.Sender = whttp.WrapMiddlewareSender(bc.Sender, mws...)
}

// Get retrieves the business profile for the given config.
func (bc *BaseClient) Get(ctx context.Context, conf *config.Config, fields []string) ([]*Profile, error) {
	resp, err := bc.Send(ctx, conf, &BaseRequest{
		Type:        whttp.RequestTypeGetBusinessProfile,
		Method:      http.MethodGet,
		QueryFields: fields,
	})
	if err != nil {
		return nil, err
	}

	return resp.Profiles, nil
}

// Update updates the business profile for the given config.
func (bc *BaseClient) Update(ctx context.Context, conf *config.Config, request *UpdateProfileRequest) (bool, error) {
	resp, err := bc.Send(ctx, conf, &BaseRequest{
		Type:    whttp.RequestTypeUpdateBusinessProfile,
		Method:  http.MethodPost,
		Payload: request,
	})
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

var _ = (*Client)(nil)
