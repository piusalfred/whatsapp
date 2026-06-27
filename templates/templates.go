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

// Package templates provides a client for the WhatsApp Business Template Management API.
//
// The Template Management API lets businesses create, retrieve, list, edit, and delete
// message templates that can be used to send structured messages to customers.
//
// # Getting Started
//
// Create a [Client] using [NewClient] with a [config.Config] and optional sender options:
//
//	conf := &config.Config{
//	    BaseURL:           "https://graph.facebook.com",
//	    APIVersion:        "v22.0",
//	    AccessToken:       "YOUR_ACCESS_TOKEN",
//	    PhoneNumberID:     "YOUR_PHONE_NUMBER_ID",
//	    BusinessAccountID: "YOUR_WABA_ID",
//	}
//
//	client := templates.NewClient(conf,
//	    whttp.WithSenderTimeout(30*time.Second),
//	    whttp.WithSenderMaxBodyBytes(10<<20),
//	)
//
// # Creating a Template
//
//	resp, err := client.Create(ctx, &templates.CreateRequest{
//	    Name:     "welcome_message",
//	    Language: "en",
//	    Category: "MARKETING",
//	    Components: []*templates.Component{{
//	        Type: "BODY",
//	        Text: "Welcome to our service!",
//	    }},
//	})
//
// # Listing Templates
//
//	listResp, err := client.List(ctx, &templates.ListRequest{
//	    Limit: 10,
//	})
//	for _, t := range listResp.Data {
//	    fmt.Println(t.Name, t.Status)
//	}
//
// # Editing and Deleting
//
//	editResp, err := client.Edit(ctx, &templates.EditRequest{
//	    TemplateID: "12345",
//	    Components: []*templates.Component{{Type: "BODY", Text: "Updated text"}},
//	})
//
//	delResp, err := client.Delete(ctx, &templates.DeleteRequest{Name: "welcome_message"})
//
// # Configuration Options
//
// [whttp.CoreSenderOption] functions customize the underlying HTTP transport:
//
//	whttp.WithSenderHTTPClient(customHTTPClient)
//	whttp.WithSenderRequestInterceptor(myRequestHook)
//	whttp.WithSenderResponseInterceptor(myResponseHook)
//	whttp.WithSenderTimeout(30 * time.Second)
//	whttp.WithSenderMaxBodyBytes(10 << 20)
//	whttp.WithSenderMaxHeaderBytes(1 << 20)
//
// # Testing
//
// For unit tests, inject a mock sender via [Client.SetBaseClient]:
//
//	client := templates.NewClient(conf)
//	client.SetBaseClient(mockSender)
package templates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const Endpoint = "message_templates"

var ErrTemplateNotFound = errors.New("template not found")

type (
	// CreateRequest carries the parameters for creating a new message template.
	CreateRequest struct {
		Name                     string
		Language                 string
		Category                 string
		ParameterFormat          string
		Components               []*Component
		AllowCategoryChange      bool
		CTAURLLinkTrackingOptOut bool
		MessageSendTTLSeconds    int64
		SubCategory              string
		DisplayFormat            string
		LibraryTemplateName      string
		IsPrimaryDeviceOnly      bool
		SendType                 string
	}

	// CreateResponse contains the ID, status, and category of a newly created template.
	CreateResponse struct {
		ID       string
		Status   string
		Category string
	}

	// ListRequest carries pagination and field filtering for listing templates.
	ListRequest struct {
		Fields []string
		Limit  int
		After  string
		Before string
	}

	// ListResponse contains a page of templates.
	ListResponse struct {
		Data   []*TemplateInfo
		Paging *whttp.Paging
	}

	// GetRequest carries the template ID and optional field filter.
	GetRequest struct {
		TemplateID string
		Fields     []string
	}

	// EditRequest carries the parameters for updating an existing template.
	EditRequest struct {
		TemplateID               string
		Components               []*Component
		Category                 string
		ParameterFormat          string
		AllowCategoryChange      bool
		CTAURLLinkTrackingOptOut bool
		MessageSendTTLSeconds    int64
		SubCategory              string
		DisplayFormat            string
		IsPrimaryDeviceOnly      bool
	}

	// DeleteRequest carries identifiers for template deletion.
	DeleteRequest struct {
		Name   string
		HSMID  string
		HSMIDs []string
	}

	// Request is the domain model that holds all fields needed for any
	// template management operation.
	Request struct {
		RequestType whttp.RequestType `json:"-"`
		TemplateID  string            `json:"-"`
		Name        string            `json:"-"`
		HSMID       string            `json:"-"`
		HSMIDs      []string          `json:"-"`
		Fields      []string          `json:"-"`
		Limit       int               `json:"-"`
		After       string            `json:"-"`
		Before      string            `json:"-"`
		Message     *BaseRequest      `json:"-"`
	}

	// BaseRequest is the wire-format payload sent to the Templates API.
	BaseRequest struct {
		Name                     string       `json:"name,omitempty"`
		Language                 string       `json:"language,omitempty"`
		Category                 string       `json:"category,omitempty"`
		ParameterFormat          string       `json:"parameter_format,omitempty"`
		Components               []*Component `json:"components,omitempty"`
		AllowCategoryChange      bool         `json:"allow_category_change,omitempty"`
		CTAURLLinkTrackingOptOut bool         `json:"cta_url_link_tracking_opted_out,omitempty"`
		MessageSendTTLSeconds    int64        `json:"message_send_ttl_seconds,omitempty"`
		SubCategory              string       `json:"sub_category,omitempty"`
		DisplayFormat            string       `json:"display_format,omitempty"`
		LibraryTemplateName      string       `json:"library_template_name,omitempty"`
		IsPrimaryDeviceOnly      bool         `json:"is_primary_device_delivery_only,omitempty"`
		SendType                 string       `json:"send_type,omitempty"`
	}

	// BaseResponse is the general response struct that captures all possible
	// fields returned by the Templates API. Use the typed helper methods to
	// extract operation-specific responses.
	BaseResponse struct {
		Success  bool            `json:"success,omitempty"`
		ID       string          `json:"id,omitempty"`
		Status   string          `json:"status,omitempty"`
		Category string          `json:"category,omitempty"`
		Data     []*TemplateInfo `json:"data,omitempty"`
		Paging   *whttp.Paging   `json:"paging,omitempty"`
	}

	// TemplateInfo describes a single message template.
	TemplateInfo struct {
		ID                       string          `json:"id,omitempty"`
		Name                     string          `json:"name,omitempty"`
		Language                 string          `json:"language,omitempty"`
		Category                 string          `json:"category,omitempty"`
		Status                   string          `json:"status,omitempty"`
		QualityScore             string          `json:"quality_score,omitempty"`
		ParameterFormat          string          `json:"parameter_format,omitempty"`
		Components               json.RawMessage `json:"components,omitempty"`
		CTAURLLinkTrackingOptOut bool            `json:"cta_url_link_tracking_opted_out,omitempty"`
		LastUpdatedTime          int64           `json:"last_updated_time,omitempty"`
		MessageSendTTLSeconds    int64           `json:"message_send_ttl_seconds,omitempty"`
		RejectedReason           string          `json:"rejected_reason,omitempty"`
	}

	// Component describes a single template component (HEADER, BODY, FOOTER, BUTTONS).
	Component struct {
		Type                      string    `json:"type"`
		Text                      string    `json:"text,omitempty"`
		Format                    string    `json:"format,omitempty"`
		Buttons                   []*Button `json:"buttons,omitempty"`
		Example                   *Example  `json:"example,omitempty"`
		AddSecurityRecommendation bool      `json:"add_security_recommendation,omitempty"`
		CodeExpirationMinutes     int       `json:"code_expiration_minutes,omitempty"`
	}

	// Button describes a button within a BUTTONS component.
	Button struct {
		Type           string `json:"type"`
		Text           string `json:"text,omitempty"`
		URL            string `json:"url,omitempty"`
		PhoneNumber    string `json:"phone_number,omitempty"`
		OTPType        string `json:"otp_type,omitempty"`
		AutofillText   string `json:"autofill_text,omitempty"`
		PackageName    string `json:"package_name,omitempty"`
		SignatureHash  string `json:"signature_hash,omitempty"`
		FlowID         string `json:"flow_id,omitempty"`
		FlowAction     string `json:"flow_action,omitempty"`
		NavigateScreen string `json:"navigate_screen,omitempty"`
	}

	// Example provides sample values for template variables.
	Example struct {
		HeaderText   []string   `json:"header_text,omitempty"`
		HeaderHandle []string   `json:"header_handle,omitempty"`
		BodyText     [][]string `json:"body_text,omitempty"`
	}

	// Client is a high-level client bound to a fixed [config.Config].
	Client struct {
		sender *BaseClient
		config *config.Config
	}
)

// NewClient creates a high-level [Client] with a fixed configuration.
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
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.SetMiddlewares(mws...)
}

// Send dispatches a raw [Request] through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return response, nil
}

// Create creates a new message template.
func (c *Client) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	resp, err := c.sender.Create(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("client: create template: %w", err)
	}
	return resp, nil
}

// List returns a paginated list of message templates.
func (c *Client) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	resp, err := c.sender.List(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("client: list templates: %w", err)
	}
	return resp, nil
}

// Get retrieves a single message template by ID.
func (c *Client) Get(ctx context.Context, req *GetRequest) (*TemplateInfo, error) {
	resp, err := c.sender.Get(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("client: get template: %w", err)
	}
	return resp, nil
}

// Edit updates an existing message template.
func (c *Client) Edit(ctx context.Context, req *EditRequest) (*CreateResponse, error) {
	resp, err := c.sender.Edit(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("client: edit template: %w", err)
	}
	return resp, nil
}

// Delete removes a message template.
func (c *Client) Delete(ctx context.Context, req *DeleteRequest) (*BaseResponse, error) {
	resp, err := c.sender.Delete(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("client: delete template: %w", err)
	}
	return resp, nil
}

// BaseClient is the low-level HTTP executor for the Templates API.
type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	bc.BaseClient.SetMiddlewares(mws...)
}

// buildRequestParams maps a [Request] to HTTP method, endpoints, query params,
// and body payload.
func buildRequestParams(
	conf *config.Config,
	request *Request,
) (string, []string, map[string]string, *BaseRequest, error) {
	var (
		method      string
		endpoints   []string
		queryParams = map[string]string{}
		message     *BaseRequest
	)

	switch request.RequestType {
	case whttp.RequestTypeCreateTemplate:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, conf.PhoneNumberID, Endpoint}
		message = request.Message

	case whttp.RequestTypeListTemplates:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, conf.BusinessAccountID, Endpoint}
		if len(request.Fields) > 0 {
			queryParams["fields"] = strings.Join(request.Fields, ",")
		}
		if request.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(request.Limit)
		}
		if request.After != "" {
			queryParams["after"] = request.After
		}
		if request.Before != "" {
			queryParams["before"] = request.Before
		}

	case whttp.RequestTypeGetTemplate:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, request.TemplateID}
		if len(request.Fields) > 0 {
			queryParams["fields"] = strings.Join(request.Fields, ",")
		}

	case whttp.RequestTypeEditTemplate:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, request.TemplateID}
		message = request.Message

	case whttp.RequestTypeDeleteTemplate:
		method = http.MethodDelete
		endpoints = []string{conf.APIVersion, conf.BusinessAccountID, Endpoint}
		if request.Name != "" {
			queryParams["name"] = request.Name
		}
		if request.HSMID != "" {
			queryParams["hsm_id"] = request.HSMID
		}
		if len(request.HSMIDs) > 0 {
			queryParams["hsm_ids"] = strings.Join(request.HSMIDs, ",")
		}

	default:
		return "", nil, nil, nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}

	return method, endpoints, queryParams, message, nil
}

// Send translates a high-level [Request] into an HTTP transaction and returns
// the decoded [BaseResponse].
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	method, endpoints, queryParams, message, err := buildRequestParams(conf, request)
	if err != nil {
		return nil, err
	}

	b := whttp.NewRequestBuilder(method, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(request.RequestType).
		Endpoints(endpoints...)

	if len(queryParams) > 0 {
		b = b.QueryParams(queryParams)
	}

	req := whttp.Build(b, message)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptionsPermissive())

	if err = bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return resp, nil
}

// Create creates a new message template.
func (bc *BaseClient) Create(ctx context.Context, conf *config.Config, req *CreateRequest) (*CreateResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeCreateTemplate,
		Message: &BaseRequest{
			Name:                     req.Name,
			Language:                 req.Language,
			Category:                 req.Category,
			ParameterFormat:          req.ParameterFormat,
			Components:               req.Components,
			AllowCategoryChange:      req.AllowCategoryChange,
			CTAURLLinkTrackingOptOut: req.CTAURLLinkTrackingOptOut,
			MessageSendTTLSeconds:    req.MessageSendTTLSeconds,
			SubCategory:              req.SubCategory,
			DisplayFormat:            req.DisplayFormat,
			LibraryTemplateName:      req.LibraryTemplateName,
			IsPrimaryDeviceOnly:      req.IsPrimaryDeviceOnly,
			SendType:                 req.SendType,
		},
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: create template: %w", err)
	}

	return &CreateResponse{
		ID:       resp.ID,
		Status:   resp.Status,
		Category: resp.Category,
	}, nil
}

// List returns a paginated list of message templates.
func (bc *BaseClient) List(ctx context.Context, conf *config.Config, req *ListRequest) (*ListResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeListTemplates,
		Fields:      req.Fields,
		Limit:       req.Limit,
		After:       req.After,
		Before:      req.Before,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: list templates: %w", err)
	}

	return &ListResponse{
		Data:   resp.Data,
		Paging: resp.Paging,
	}, nil
}

// Get retrieves a single message template by ID.
func (bc *BaseClient) Get(ctx context.Context, conf *config.Config, req *GetRequest) (*TemplateInfo, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetTemplate,
		TemplateID:  req.TemplateID,
		Fields:      req.Fields,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: get template: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, ErrTemplateNotFound
	}

	return resp.Data[0], nil
}

// Edit updates an existing message template.
func (bc *BaseClient) Edit(ctx context.Context, conf *config.Config, req *EditRequest) (*CreateResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeEditTemplate,
		TemplateID:  req.TemplateID,
		Message: &BaseRequest{
			Components:               req.Components,
			Category:                 req.Category,
			ParameterFormat:          req.ParameterFormat,
			AllowCategoryChange:      req.AllowCategoryChange,
			CTAURLLinkTrackingOptOut: req.CTAURLLinkTrackingOptOut,
			MessageSendTTLSeconds:    req.MessageSendTTLSeconds,
			SubCategory:              req.SubCategory,
			DisplayFormat:            req.DisplayFormat,
			IsPrimaryDeviceOnly:      req.IsPrimaryDeviceOnly,
		},
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: edit template: %w", err)
	}

	return &CreateResponse{
		ID:       resp.ID,
		Status:   resp.Status,
		Category: resp.Category,
	}, nil
}

// Delete removes a message template.
func (bc *BaseClient) Delete(ctx context.Context, conf *config.Config, req *DeleteRequest) (*BaseResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeDeleteTemplate,
		Name:        req.Name,
		HSMID:       req.HSMID,
		HSMIDs:      req.HSMIDs,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: delete template: %w", err)
	}

	return resp, nil
}
