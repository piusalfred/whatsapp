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

package phonenumber

//go:generate mockgen -destination=../mocks/phonenumber/mock_phonenumber.go -package=phonenumber -source=phonenumber.go

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
		Paging *Paging        `json:"paging,omitempty"`
	}

	Paging struct {
		Cursors *Cursors `json:"cursors"`
	}

	Cursors struct {
		Before string `json:"before"`
		After  string `json:"after"`
	}

	Response struct {
		Data                   []*PhoneNumber `json:"data,omitempty"`
		CodeVerificationStatus string         `json:"code_verification_status,omitempty"`
		DisplayPhoneNumber     string         `json:"display_phone_number,omitempty"`
		ID                     string         `json:"id,omitempty"`
		QualityRating          string         `json:"quality_rating,omitempty"`
		VerifiedName           string         `json:"verified_name,omitempty"`
		Paging                 *Paging        `json:"paging,omitempty"`
		NameStatus             string         `json:"name_status,omitempty"`
	}

	BaseClient struct {
		Sender Sender
		Config config.Reader
	}

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

func NewBaseClient(reader config.Reader, sender Sender, middlewares ...SenderMiddleware) (*BaseClient, error) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		if mw != nil {
			sender = mw(sender.Send)
		}
	}

	client := &BaseClient{
		Config: reader,
		Sender: sender,
	}

	return client, nil
}

func (c *BaseClient) List(ctx context.Context) (*ListResponse, error) {
	req := &BaseRequest{
		Type:        whttp.RequestTypeListPhoneNumbers,
		Method:      http.MethodGet,
		QueryParams: map[string]string{},
	}

	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	response, err := c.Sender.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	return response.ListPhoneNumbersResponse(), nil
}

func (c *BaseClient) Get(ctx context.Context, req *GetRequest) (*PhoneNumber, error) {
	request := &BaseRequest{
		Type:   whttp.RequestTypeGetPhoneNumber,
		Method: http.MethodGet,
		QueryParams: map[string]string{
			"fields": strings.Join(req.Fields, ";"),
		},
	}

	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	response, err := c.Sender.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}

	return response.PhoneNumber(), nil
}

type (
	BaseRequest struct {
		Type        whttp.RequestType
		Method      string
		QueryParams map[string]string
	}

	BaseSender struct {
		Sender whttp.Sender[any]
	}

	Sender interface {
		Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error)
	}

	SenderFunc func(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error)

	SenderMiddleware func(sender SenderFunc) SenderFunc
)

func (fn SenderFunc) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error) {
	return fn(ctx, conf, req)
}

func (c *BaseSender) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error) {
	if req.QueryParams == nil {
		req.QueryParams = map[string]string{}
	}
	req.QueryParams["access_token"] = conf.AccessToken

	endpoints := []string{conf.APIVersion}
	if req.Type == whttp.RequestTypeListPhoneNumbers {
		endpoints = append(endpoints, conf.BusinessAccountID, "phone_numbers")
	} else {
		endpoints = append(endpoints, conf.PhoneNumberID)
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestEndpoints[any](endpoints...),
		whttp.WithRequestQueryParams[any](req.QueryParams),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestType[any](req.Type),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestSecured[any](conf.SecureRequests),
	}

	request := whttp.MakeRequest(req.Method, conf.BaseURL, opts...)

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})
	if err := c.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type Client struct {
	Config *config.Config
	Sender Sender
}

func NewClient(ctx context.Context, reader config.Reader, sender Sender,
	middlewares ...SenderMiddleware,
) (*Client, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		if mw != nil {
			sender = mw(sender.Send)
		}
	}

	client := &Client{
		Config: conf,
		Sender: sender,
	}

	return client, nil
}

func (c *Client) List(ctx context.Context) (*ListResponse, error) {
	req := &BaseRequest{
		Type:        whttp.RequestTypeListPhoneNumbers,
		Method:      http.MethodGet,
		QueryParams: map[string]string{},
	}

	response, err := c.Sender.Send(ctx, c.Config, req)
	if err != nil {
		return nil, fmt.Errorf("list phone numbers: %w", err)
	}

	return response.ListPhoneNumbersResponse(), nil
}

func (c *Client) Get(ctx context.Context, req *GetRequest) (*PhoneNumber, error) {
	request := &BaseRequest{
		Type:   whttp.RequestTypeGetPhoneNumber,
		Method: http.MethodGet,
		QueryParams: map[string]string{
			"fields": strings.Join(req.Fields, ";"),
		},
	}

	response, err := c.Sender.Send(ctx, c.Config, request)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}

	return response.PhoneNumber(), nil
}

type Service interface {
	List(ctx context.Context) (*ListResponse, error)
	Get(ctx context.Context, request *GetRequest) (*PhoneNumber, error)
}

var (
	_ Service = (*Client)(nil)
	_ Service = (*BaseClient)(nil)
)
