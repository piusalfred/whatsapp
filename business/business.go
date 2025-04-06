/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
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

//go:generate go tool mockgen -destination=../mocks/business/mock_business.go -package=business -source=business.go

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
	BaseRequest struct {
		Type        whttp.RequestType
		Method      string
		Payload     any
		QueryFields []string
	}

	BaseSender struct {
		Sender whttp.AnySender
	}

	Response struct {
		Profiles []*Profile `json:"data,omitempty"`
		Success  bool       `json:"success"`
	}

	Service interface {
		Get(ctx context.Context, fields []string) ([]*Profile, error)
		Update(ctx context.Context, request *UpdateProfileRequest) (bool, error)
	}
)

func (s *BaseSender) Send(ctx context.Context, config *config.Config, request *BaseRequest) (*Response, error) {
	if request.QueryFields == nil {
		request.QueryFields = []string{}
	}

	params := map[string]string{}
	fields := strings.Join(request.QueryFields, ",")
	if fields != "" {
		params["fields"] = fields
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestEndpoints[any](config.APIVersion, config.PhoneNumberID, Endpoint),
		whttp.WithRequestQueryParams[any](params),
		whttp.WithRequestBearer[any](config.AccessToken),
		whttp.WithRequestType[any](request.Type),
		whttp.WithRequestAppSecret[any](config.AppSecret),
		whttp.WithRequestSecured[any](config.SecureRequests),
	}

	if request.Payload != nil {
		opts = append(opts, whttp.WithRequestMessage[any](&request.Payload))
	}

	req := whttp.MakeRequest[any](request.Method, config.BaseURL, opts...)

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := s.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type BaseClient struct {
	Sender Sender
	Config config.Reader
}

func NewBaseClient(sender Sender, config config.Reader) *BaseClient {
	return &BaseClient{Sender: sender, Config: config}
}

func (b *BaseClient) Get(ctx context.Context, fields []string) ([]*Profile, error) {
	conf, err := b.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	req := &BaseRequest{
		Type:        whttp.RequestTypeGetBusinessProfile,
		Method:      http.MethodGet,
		QueryFields: fields,
	}

	response, err := b.Send(ctx, conf, req)
	if err != nil {
		return nil, err
	}

	return response.Profiles, nil
}

func (b *BaseClient) Update(ctx context.Context, request *UpdateProfileRequest) (bool, error) {
	conf, err := b.Config.Read(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to read config: %w", err)
	}

	req := &BaseRequest{
		Type:    whttp.RequestTypeUpdateBusinessProfile,
		Method:  http.MethodPost,
		Payload: request,
	}

	response, err := b.Send(ctx, conf, req)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

func (b *BaseClient) Send(ctx context.Context, config *config.Config, request *BaseRequest) (*Response, error) {
	response, err := b.Sender.Send(ctx, config, request)
	if err != nil {
		return nil, fmt.Errorf("business client: %w", err)
	}

	return response, nil
}

var (
	_ Sender  = (*BaseClient)(nil)
	_ Service = (*BaseClient)(nil)
)

type Client struct {
	Sender Sender
	Config *config.Config
}

func NewClient(ctx context.Context, reader config.Reader, sender Sender) (*Client, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	c := &Client{
		Sender: sender,
		Config: conf,
	}

	return c, nil
}

func (c *Client) Get(ctx context.Context, fields []string) ([]*Profile, error) {
	req := &BaseRequest{
		Type:        whttp.RequestTypeGetBusinessProfile,
		Method:      http.MethodGet,
		QueryFields: fields,
	}

	response, err := c.Send(ctx, c.Config, req)
	if err != nil {
		return nil, err
	}

	return response.Profiles, nil
}

func (c *Client) Update(ctx context.Context, request *UpdateProfileRequest) (bool, error) {
	req := &BaseRequest{
		Type:    whttp.RequestTypeUpdateBusinessProfile,
		Method:  http.MethodPost,
		Payload: request,
	}

	response, err := c.Send(ctx, c.Config, req)
	if err != nil {
		return false, err
	}

	return response.Success, nil
}

func (c *Client) Send(ctx context.Context, config *config.Config, request *BaseRequest) (*Response, error) {
	response, err := c.Sender.Send(ctx, config, request)
	if err != nil {
		return nil, fmt.Errorf("business client: %w", err)
	}

	return response, nil
}

var (
	_ Sender  = (*Client)(nil)
	_ Service = (*Client)(nil)
)

type (
	Sender interface {
		Send(ctx context.Context, config *config.Config, request *BaseRequest) (*Response, error)
	}
)
