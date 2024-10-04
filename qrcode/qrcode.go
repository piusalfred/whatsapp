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

package qrcode

//go:generate mockgen -destination=../mocks/qrcode/mock_qrcode.go -package=qrcode -source=qrcode.go

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	ImageFormatPNG ImageFormat = "PNG"
	ImageFormatSVG ImageFormat = "SVG"
)

const Endpoint = "message_qrdls"

type (
	ImageFormat string

	CreateRequest struct {
		PrefilledMessage string      `json:"prefilled_message"`
		ImageFormat      ImageFormat `json:"generate_qr_image"`
	}

	UpdateRequest struct {
		QRCodeID         string      `json:"-"`
		PrefilledMessage string      `json:"prefilled_message,omitempty"`
		ImageFormat      ImageFormat `json:"generate_qr_image,omitempty"`
	}

	CreateResponse struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
		QRImageURL       string `json:"qr_image_url"`
	}

	Information struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
	}

	ListResponse struct {
		Data []*Information `json:"data,omitempty"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}

	BaseClient struct {
		Sender Sender
		Config config.Reader
	}
)

func NewBaseClient(s whttp.Sender[any], reader config.Reader, middlewares ...SenderMiddleware) *BaseClient {
	sender := &BaseSender{Sender: s}
	return &BaseClient{
		Sender: wrapMiddlewares(sender.Send, middlewares),
		Config: reader,
	}
}

func (c *BaseClient) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return Create(ctx, c.Sender, conf, req)
}

func (c *BaseClient) List(ctx context.Context) (*ListResponse, error) {
	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return List(ctx, c.Sender, conf)
}

func (c *BaseClient) Update(ctx context.Context, req *UpdateRequest) (*SuccessResponse, error) {
	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return Update(ctx, c.Sender, conf, req)
}

func (c *BaseClient) Delete(ctx context.Context, qrCodeID string) (*SuccessResponse, error) {
	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return Delete(ctx, c.Sender, conf, qrCodeID)
}

func (c *BaseClient) Get(ctx context.Context, qrCodeID string) (*Information, error) {
	conf, err := c.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return Get(ctx, c.Sender, conf, qrCodeID)
}

type Client struct {
	Config *config.Config
	Sender Sender
}

func NewClient(ctx context.Context, reader config.Reader, sender Sender, middlewares ...SenderMiddleware) (*Client, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Config: conf,
		Sender: wrapMiddlewares(sender.Send, middlewares),
	}

	return client, nil
}

func (c *Client) Get(ctx context.Context, qrCodeID string) (*Information, error) {
	return Get(ctx, c.Sender, c.Config, qrCodeID)
}

func (c *Client) List(ctx context.Context) (*ListResponse, error) {
	return List(ctx, c.Sender, c.Config)
}

func (c *Client) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	return Create(ctx, c.Sender, c.Config, req)
}

func (c *Client) Delete(ctx context.Context, qrCodeID string) (*SuccessResponse, error) {
	return Delete(ctx, c.Sender, c.Config, qrCodeID)
}

func (c *Client) Update(ctx context.Context, req *UpdateRequest) (*SuccessResponse, error) {
	return Update(ctx, c.Sender, c.Config, req)
}

func Create(ctx context.Context, sender Sender, conf *config.Config, req *CreateRequest) (*CreateResponse, error) {
	queryParams := map[string]string{
		"prefilled_message": req.PrefilledMessage,
		"generate_qr_image": string(req.ImageFormat),
	}

	request := &BaseRequest{
		Method:      http.MethodPost,
		Type:        whttp.RequestTypeCreateQR,
		QueryParams: queryParams,
	}

	response, err := sender.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return response.CreateResponse(), nil
}

func Get(ctx context.Context, sender Sender, conf *config.Config, qrCodeID string) (*Information, error) {
	request := &BaseRequest{
		Method:      http.MethodGet,
		Type:        whttp.RequestTypeGetQR,
		QRCodeID:    qrCodeID,
		QueryParams: map[string]string{},
	}

	response, err := sender.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("qr code not found")
	}

	return response.Data[0], nil
}

func List(ctx context.Context, sender Sender, conf *config.Config) (*ListResponse, error) {
	request := &BaseRequest{
		Method:      http.MethodGet,
		Type:        whttp.RequestTypeListQR,
		QueryParams: map[string]string{},
	}

	response, err := sender.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return response.ListResponse(), nil
}

func Delete(ctx context.Context, sender Sender, conf *config.Config, qrCodeID string) (*SuccessResponse, error) {
	request := &BaseRequest{
		Method:      http.MethodDelete,
		Type:        whttp.RequestTypeDeleteQR,
		QRCodeID:    qrCodeID,
		QueryParams: map[string]string{},
	}

	response, err := sender.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: response.Success}, nil
}

func Update(ctx context.Context, sender Sender, conf *config.Config, req *UpdateRequest) (*SuccessResponse, error) {
	queryParams := map[string]string{
		"prefilled_message": req.PrefilledMessage,
		"generate_qr_image": string(req.ImageFormat),
	}

	request := &BaseRequest{
		Method:      http.MethodPost,
		Type:        whttp.RequestTypeUpdateQR,
		QRCodeID:    req.QRCodeID,
		QueryParams: queryParams,
	}

	response, err := sender.Send(ctx, conf, request)
	if err != nil {
		return nil, err
	}

	return &SuccessResponse{Success: response.Success}, nil
}

type (
	BaseRequest struct {
		Method      string
		Type        whttp.RequestType
		QRCodeID    string
		QueryParams map[string]string
	}

	Response struct {
		Data             []*Information `json:"data,omitempty"`
		Success          bool           `json:"success"`
		Code             string         `json:"code"`
		PrefilledMessage string         `json:"prefilled_message"`
		DeepLinkURL      string         `json:"deep_link_url"`
		QRImageURL       string         `json:"qr_image_url"`
	}

	Sender interface {
		Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error)
	}

	SenderFunc func(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error)
)

func (fn SenderFunc) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error) {
	return fn(ctx, conf, req)
}

type SenderMiddleware func(senderFunc SenderFunc) SenderFunc

func wrapMiddlewares(next SenderFunc, middlewares []SenderMiddleware) SenderFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

func (response *Response) CreateResponse() *CreateResponse {
	return &CreateResponse{
		Code:             response.Code,
		PrefilledMessage: response.PrefilledMessage,
		DeepLinkURL:      response.DeepLinkURL,
		QRImageURL:       response.QRImageURL,
	}
}

func (response *Response) ListResponse() *ListResponse {
	return &ListResponse{Data: response.Data}
}

type BaseSender struct {
	Sender whttp.Sender[any]
}

func (sender *BaseSender) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*Response, error) {
	if req.QueryParams == nil {
		req.QueryParams = map[string]string{}
	}
	req.QueryParams["access_token"] = conf.AccessToken

	endpoints := []string{conf.APIVersion, conf.PhoneNumberID, Endpoint}
	if req.QRCodeID != "" {
		endpoints = append(endpoints, req.QRCodeID)
	}

	request := &whttp.Request[any]{
		Type:        req.Type,
		Method:      req.Method,
		BaseURL:     conf.BaseURL,
		Endpoints:   endpoints,
		QueryParams: req.QueryParams,
	}

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := sender.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type Service interface {
	Get(ctx context.Context, qrCodeID string) (*Information, error)
	List(ctx context.Context) (*ListResponse, error)
	Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
	Delete(ctx context.Context, qrCodeID string) (*SuccessResponse, error)
	Update(ctx context.Context, req *UpdateRequest) (*SuccessResponse, error)
}

var (
	_ Service = (*BaseClient)(nil)
	_ Service = (*Client)(nil)
)
