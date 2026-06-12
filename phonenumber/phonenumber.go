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

package phonenumber

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	NameStatusApproved               = "APPROVED"
	NameStatusAvailableWithoutReview = "AVAILABLE_WITHOUT_REVIEW"
	NameStatusDeclined               = "DECLINED"
	NameStatusExpired                = "EXPIRED"
	NameStatusPendingReview          = "PENDING_REVIEW"
	NameStatusNone                   = "NONE"
)

type (
	PhoneNumber struct {
		ID                     string `json:"id"`
		DisplayPhoneNumber     string `json:"display_phone_number"`
		VerifiedName           string `json:"verified_name"`
		QualityRating          string `json:"quality_rating"`
		CodeVerificationStatus string `json:"code_verification_status,omitempty"`
		NameStatus             string `json:"name_status,omitempty"`
	}

	ListResponse struct {
		Data   []*PhoneNumber `json:"data,omitempty"`
		Paging *whttp.Paging  `json:"paging,omitempty"`
	}

	Response struct {
		Data                   []*PhoneNumber `json:"data,omitempty"`
		CodeVerificationStatus string         `json:"code_verification_status,omitempty"`
		DisplayPhoneNumber     string         `json:"display_phone_number,omitempty"`
		ID                     string         `json:"id,omitempty"`
		QualityRating          string         `json:"quality_rating,omitempty"`
		VerifiedName           string         `json:"verified_name,omitempty"`
		Paging                 *whttp.Paging  `json:"paging,omitempty"`
		NameStatus             string         `json:"name_status,omitempty"`
	}

	Request struct {
		RequestType whttp.RequestType
		QueryParams map[string]string
	}

	BaseRequest struct{}

	GetRequest struct {
		ID     string
		Fields []string
	}
)

func (response *Response) ListPhoneNumbersResponse() *ListResponse {
	return &ListResponse{
		Data:   response.Data,
		Paging: response.Paging,
	}
}

func (response *Response) PhoneNumber() *PhoneNumber {
	return &PhoneNumber{
		ID:                     response.ID,
		DisplayPhoneNumber:     response.DisplayPhoneNumber,
		VerifiedName:           response.VerifiedName,
		QualityRating:          response.QualityRating,
		CodeVerificationStatus: response.CodeVerificationStatus,
		NameStatus:             response.NameStatus,
	}
}

// Client orchestrates high-level Phone Number API operations.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// List retrieves all phone numbers associated with the business account.
func (c *Client) List(ctx context.Context) (*ListResponse, error) {
	req := &Request{
		RequestType: whttp.RequestTypeListPhoneNumbers,
		QueryParams: map[string]string{},
	}

	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list phone numbers: %w", err)
	}

	return resp.ListPhoneNumbersResponse(), nil
}

// Get retrieves details for a specific phone number.
func (c *Client) Get(ctx context.Context, req *GetRequest) (*PhoneNumber, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetPhoneNumber,
		QueryParams: map[string]string{
			"fields": strings.Join(req.Fields, ";"),
		},
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}

	return resp.PhoneNumber(), nil
}

// Send dispatches a raw Request through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*Response, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

// NewClient creates a high-level Client for the Phone Number API. The conf
// argument provides endpoint and credential configuration. Optional SenderOption
// functions tune the underlying HTTP transport.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock whttp.Sender and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.Sender = sender
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.Sender = whttp.WrapMiddlewareSender(c.sender.Sender, mws...)
}

// BaseClient is the low-level HTTP executor for the Phone Number API. It
// converts domain Request values into HTTP traffic and decodes JSON responses.
type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

// Send translates a high-level Request into an HTTP transaction and returns
// the decoded Response.
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*Response, error) {
	var endpoints []string

	switch request.RequestType {
	case whttp.RequestTypeListPhoneNumbers:
		endpoints = []string{conf.APIVersion, conf.BusinessAccountID, "phone_numbers"}

	case whttp.RequestTypeGetPhoneNumber:
		endpoints = []string{conf.APIVersion, conf.PhoneNumberID}

	default:
		return nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}

	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.RequestType).
		Endpoints(endpoints...)

	if len(request.QueryParams) > 0 {
		b = b.QueryParams(request.QueryParams)
	}

	req := whttp.Build[BaseRequest](b, nil)

	resp := &Response{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return resp, nil
}

func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	bc.Sender = whttp.WrapMiddlewareSender(bc.Sender, mws...)
}

var _ = (*Client)(nil)
