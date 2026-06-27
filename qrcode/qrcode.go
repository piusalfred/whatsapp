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

// Package qrcode provides a client for the WhatsApp Business QR Code API.
//
// The QR Code API lets businesses create, retrieve, update, list, and delete
// QR codes that customers can scan to start a WhatsApp conversation with a
// pre-filled message.
//
// # Getting Started
//
// Create a [Client] using [NewClient] with a [config.Config] and optional sender options:
//
//	conf := &config.Config{
//	    BaseURL:       "https://graph.facebook.com",
//	    APIVersion:    "v22.0",
//	    AccessToken:   "YOUR_ACCESS_TOKEN",
//	    PhoneNumberID: "YOUR_PHONE_NUMBER_ID",
//	}
//
//	client := qrcode.NewClient(conf,
//	    whttp.WithSenderTimeout(30*time.Second),
//	    whttp.WithSenderMaxBodyBytes(10<<20),
//	)
//
// # Creating a QR Code
//
//	resp, err := client.Create(ctx, &qrcode.CreateRequest{
//	    PrefilledMessage: "Hello! I saw your ad",
//	    ImageFormat:      qrcode.ImageFormatPNG,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("QR Image URL:", resp.QRImageURL)
//
// # Listing QR Codes
//
//	listResp, err := client.List(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, qr := range listResp.Data {
//	    fmt.Println(qr.Code, qr.DeepLinkURL)
//	}
//
// # Updating and Deleting
//
//	updateResp, err := client.Update(ctx, &qrcode.UpdateRequest{
//	    QRCodeID:         resp.Code,
//	    PrefilledMessage: "Updated message",
//	    ImageFormat:      qrcode.ImageFormatSVG,
//	})
//
//	delResp, err := client.Delete(ctx, resp.Code)
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
//	client := qrcode.NewClient(conf)
//	client.SetBaseClient(mockSender)
package qrcode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	// ImageFormatPNG requests a PNG image for the QR code.
	ImageFormatPNG ImageFormat = "PNG"
	// ImageFormatSVG requests an SVG image for the QR code.
	ImageFormatSVG ImageFormat = "SVG"
)

const Endpoint = "message_qrdls"

// ImageFormat determines the output format of the generated QR code image.
type ImageFormat string

type (
	// CreateRequest carries the parameters for creating a new QR code.
	CreateRequest struct {
		PrefilledMessage string      `json:"prefilled_message"`
		ImageFormat      ImageFormat `json:"generate_qr_image"`
	}

	// UpdateRequest carries the parameters for updating an existing QR code.
	UpdateRequest struct {
		QRCodeID         string      `json:"-"`
		PrefilledMessage string      `json:"prefilled_message,omitempty"`
		ImageFormat      ImageFormat `json:"generate_qr_image,omitempty"`
	}

	// ListOptions provides pagination and filtering for list operations.
	ListOptions struct {
		Limit  *int
		After  *string
		Before *string
		Fields *string
		Code   *string
	}

	// CreateResponse is returned on a successful create operation.
	CreateResponse struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
		QRImageURL       string `json:"qr_image_url"`
	}

	// Information describes a single QR code without the image URL.
	Information struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
	}

	// ListResponse contains the collection of QR codes.
	ListResponse struct {
		Data []*Information `json:"data,omitempty"`
	}

	// SuccessResponse indicates whether a mutating operation succeeded.
	SuccessResponse struct {
		Success bool `json:"success"`
	}

	// Client is a high-level client bound to a fixed [config.Config].
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	// Request is an internal unified context data carrier mapping operation
	// metadata down to the HTTP executor.
	Request struct {
		Type             whttp.RequestType
		QRCodeID         string
		PrefilledMessage string
		GenerateQRImage  string
		ListOptions      *ListOptions
	}

	// BaseRequest is the JSON wire-format payload for create and update operations.
	BaseRequest struct {
		Code             string `json:"code,omitempty"`
		PrefilledMessage string `json:"prefilled_message,omitempty"`
		GenerateQRImage  string `json:"generate_qr_image,omitempty"`
	}

	// BaseResponse acts as a flexible intermediate data capture layer unmarshaling
	// varying response structures across disparate HTTP verbs.
	BaseResponse struct {
		Data             []*Information `json:"data,omitempty"`
		Success          bool           `json:"success"`
		Code             string         `json:"code"`
		PrefilledMessage string         `json:"prefilled_message"`
		DeepLinkURL      string         `json:"deep_link_url"`
		QRImageURL       string         `json:"qr_image_url"`
	}
)

var (
	ErrCreateQRCode = errors.New("failed to create qr code")
	ErrGetQRCode    = errors.New("failed to get qr code")
	ErrListQRCode   = errors.New("failed to list qr codes")
	ErrDeleteQRCode = errors.New("failed to delete qr code")
	ErrUpdateQRCode = errors.New("failed to update qr code")
)

// ToCreateResponse attempts to coerce a BaseResponse into a CreateResponse.
func (r *BaseResponse) ToCreateResponse() *CreateResponse {
	return &CreateResponse{
		Code:             r.Code,
		PrefilledMessage: r.PrefilledMessage,
		DeepLinkURL:      r.DeepLinkURL,
		QRImageURL:       r.QRImageURL,
	}
}

// ToListResponse attempts to coerce a BaseResponse into a ListResponse.
func (r *BaseResponse) ToListResponse() *ListResponse {
	return &ListResponse{Data: r.Data}
}

// NewClient creates a high-level [Client] with a fixed configuration.
// Optional [SenderOption] functions tune the underlying HTTP transport.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.SetSender(sender)
}

// SetMiddlewares configures middlewares that wrap the underlying Sender.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.SetMiddlewares(mws...)
}

// Create generates a new QR code for the given prefilled message and image format.
func (c *Client) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	request := &Request{
		Type:             whttp.RequestTypeCreateQR,
		PrefilledMessage: req.PrefilledMessage,
		GenerateQRImage:  string(req.ImageFormat),
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateQRCode, err)
	}
	return resp.ToCreateResponse(), nil
}

// Get retrieves metadata for a specific QR code by its code identifier.
func (c *Client) Get(ctx context.Context, qrCodeID string) (*Information, error) {
	request := &Request{
		Type:     whttp.RequestTypeGetQR,
		QRCodeID: qrCodeID,
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetQRCode, err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("%w: qr code not found", ErrGetQRCode)
	}
	return resp.Data[0], nil
}

// List returns all QR codes associated with the phone number.
func (c *Client) List(ctx context.Context, opts *ListOptions) (*ListResponse, error) {
	request := &Request{
		Type:        whttp.RequestTypeListQR,
		ListOptions: opts,
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrListQRCode, err)
	}
	return resp.ToListResponse(), nil
}

// Delete removes a QR code by its code identifier.
func (c *Client) Delete(ctx context.Context, qrCodeID string) (*SuccessResponse, error) {
	request := &Request{
		Type:     whttp.RequestTypeDeleteQR,
		QRCodeID: qrCodeID,
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeleteQRCode, err)
	}
	return &SuccessResponse{Success: resp.Success}, nil
}

// Update modifies the prefilled message and/or image format of an existing QR code.
func (c *Client) Update(ctx context.Context, req *UpdateRequest) (*SuccessResponse, error) {
	request := &Request{
		Type:             whttp.RequestTypeUpdateQR,
		QRCodeID:         req.QRCodeID,
		PrefilledMessage: req.PrefilledMessage,
		GenerateQRImage:  string(req.ImageFormat),
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpdateQRCode, err)
	}
	return &SuccessResponse{Success: resp.Success}, nil
}

// Send dispatches a raw [Request] through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return response, nil
}

// BaseClient is the low-level HTTP executor for the QR Code API. It accepts a
// concrete [*config.Config] per request, making it suitable for multi-tenant
// SaaS scenarios. For a fixed-configuration client, use [Client].
type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

func buildListQueryParams(opts *ListOptions) map[string]string {
	if opts == nil {
		return nil
	}

	params := map[string]string{}
	if opts.Limit != nil {
		params["limit"] = strconv.Itoa(*opts.Limit)
	}
	if opts.After != nil {
		params["after"] = *opts.After
	}
	if opts.Before != nil {
		params["before"] = *opts.Before
	}
	if opts.Fields != nil {
		params["fields"] = *opts.Fields
	}
	if opts.Code != nil {
		params["code"] = *opts.Code
	}

	return params
}

// Send translates a high-level [Request] into an HTTP transaction and returns
// the decoded [BaseResponse].
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	var (
		method      string
		message     *BaseRequest
		queryParams map[string]string
	)

	switch request.Type {
	case whttp.RequestTypeCreateQR:
		method = http.MethodPost
		message = &BaseRequest{
			PrefilledMessage: request.PrefilledMessage,
			GenerateQRImage:  request.GenerateQRImage,
		}
	case whttp.RequestTypeUpdateQR:
		method = http.MethodPost
		message = &BaseRequest{
			Code:             request.QRCodeID,
			PrefilledMessage: request.PrefilledMessage,
			GenerateQRImage:  request.GenerateQRImage,
		}
	case whttp.RequestTypeGetQR:
		method = http.MethodGet
	case whttp.RequestTypeListQR:
		method = http.MethodGet
		queryParams = buildListQueryParams(request.ListOptions)
	case whttp.RequestTypeDeleteQR:
		method = http.MethodDelete
	}

	endpoints := []string{conf.APIVersion, conf.PhoneNumberID, Endpoint}
	if request.QRCodeID != "" && request.Type != whttp.RequestTypeUpdateQR {
		endpoints = append(endpoints, request.QRCodeID)
	}

	bld := whttp.NewRequestBuilder(method, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(request.Type).
		Endpoints(endpoints...)

	if len(queryParams) > 0 {
		bld = bld.QueryParams(queryParams)
	}

	req := whttp.Build(bld, message)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptionsPermissive())

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// Create generates a new QR code.
func (bc *BaseClient) Create(ctx context.Context, conf *config.Config, req *CreateRequest) (*CreateResponse, error) {
	request := &Request{
		Type:             whttp.RequestTypeCreateQR,
		PrefilledMessage: req.PrefilledMessage,
		GenerateQRImage:  string(req.ImageFormat),
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateQRCode, err)
	}

	return resp.ToCreateResponse(), nil
}

// Get retrieves a single QR code by ID.
func (bc *BaseClient) Get(ctx context.Context, conf *config.Config, qrCodeID string) (*Information, error) {
	request := &Request{
		Type:     whttp.RequestTypeGetQR,
		QRCodeID: qrCodeID,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetQRCode, err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("%w: qr code not found", ErrGetQRCode)
	}

	return resp.Data[0], nil
}

// List returns all QR codes for the phone number.
func (bc *BaseClient) List(ctx context.Context, conf *config.Config, opts *ListOptions) (*ListResponse, error) {
	request := &Request{
		Type:        whttp.RequestTypeListQR,
		ListOptions: opts,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrListQRCode, err)
	}

	return resp.ToListResponse(), nil
}

// Delete removes a QR code by ID.
func (bc *BaseClient) Delete(ctx context.Context, conf *config.Config, qrCodeID string) (*SuccessResponse, error) {
	request := &Request{
		Type:     whttp.RequestTypeDeleteQR,
		QRCodeID: qrCodeID,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeleteQRCode, err)
	}

	return &SuccessResponse{Success: resp.Success}, nil
}

// Update modifies an existing QR code.
func (bc *BaseClient) Update(ctx context.Context, conf *config.Config, req *UpdateRequest) (*SuccessResponse, error) {
	request := &Request{
		Type:             whttp.RequestTypeUpdateQR,
		QRCodeID:         req.QRCodeID,
		PrefilledMessage: req.PrefilledMessage,
		GenerateQRImage:  string(req.ImageFormat),
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpdateQRCode, err)
	}

	return &SuccessResponse{Success: resp.Success}, nil
}
