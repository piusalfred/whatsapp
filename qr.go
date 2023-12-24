/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package whatsapp

import (
	"context"
	"fmt"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	ImageFormatPNG ImageFormat = "PNG"
	ImageFormatSVG ImageFormat = "SVG"
)

type (
	ImageFormat string

	CreateRequest struct {
		PrefilledMessage string      `json:"prefilled_message"`
		ImageFormat      ImageFormat `json:"generate_qr_image"`
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
)

func (c *BaseClient) CreateQR(ctx context.Context, rtx *whttp.RequestContext,
	req *CreateRequest,
) (*CreateResponse, error) {
	queryParams := map[string]string{
		"prefilled_message": req.PrefilledMessage,
		"generate_qr_image": string(req.ImageFormat),
		"access_token":      rtx.Bearer,
	}
	reqCtx := &whttp.RequestContext{
		Name:          "create qr code",
		BaseURL:       rtx.BaseURL,
		ApiVersion:    rtx.ApiVersion,
		PhoneNumberID: rtx.PhoneNumberID,
		Endpoints:     []string{"message_qrdls"},
	}
	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Query:   queryParams,
	}

	var response CreateResponse

	err := c.base.Do(ctx, params, &response)
	if err != nil {
		return nil, fmt.Errorf("qr code create: %w", err)
	}

	return &response, nil
}

func (c *BaseClient) ListQR(ctx context.Context, request *RequestContext) (*ListResponse, error) {
	reqCtx := &whttp.RequestContext{
		Name:          "list qr codes",
		BaseURL:       request.BaseURL,
		ApiVersion:    request.ApiVersion,
		PhoneNumberID: request.PhoneID,
		Endpoints:     []string{"message_qrdls"},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Query:   map[string]string{"access_token": request.AccessToken},
	}

	var response ListResponse
	err := c.base.Do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("qr code list: %w", err)
	}

	return &response, nil
}

type RequestContext struct {
	BaseURL     string `json:"-"`
	PhoneID     string `json:"-"`
	ApiVersion  string `json:"-"` //nolint: revive,stylecheck
	AccessToken string `json:"-"`
}

var ErrNoDataFound = fmt.Errorf("no data found")

func (c *BaseClient) Get(ctx context.Context, request *whttp.RequestContext, qrCodeID string,
) (*Information, error) {
	var (
		list ListResponse
		resp Information
	)
	reqCtx := &whttp.RequestContext{
		Name:          "get qr code",
		BaseURL:       request.BaseURL,
		ApiVersion:    request.ApiVersion,
		PhoneNumberID: request.PhoneNumberID,
		Endpoints:     []string{"message_qrdls", qrCodeID},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Query:   map[string]string{"access_token": request.Bearer},
	}

	err := c.base.Do(ctx, req, &list)
	if err != nil {
		return nil, fmt.Errorf("qr code get: %w", err)
	}

	if len(list.Data) == 0 {
		return nil, fmt.Errorf("qr code get: %w", ErrNoDataFound)
	}

	resp = *list.Data[0]

	return &resp, nil
}

func (c *BaseClient) UpdateQR(ctx context.Context, rtx *whttp.RequestContext, qrCodeID string,
	req *CreateRequest) (*SuccessResponse, error,
) {
	reqCtx := &whttp.RequestContext{
		Name:          "update qr code",
		BaseURL:       rtx.BaseURL,
		ApiVersion:    rtx.ApiVersion,
		PhoneNumberID: rtx.PhoneNumberID,
		Endpoints:     []string{"message_qrdls", qrCodeID},
	}

	request := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Query: map[string]string{
			"prefilled_message": req.PrefilledMessage,
			"generate_qr_image": string(req.ImageFormat),
			"access_token":      rtx.Bearer,
		},
	}

	var resp SuccessResponse
	err := c.base.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("qr code update (%s): %w", qrCodeID, err)
	}

	return &resp, nil
}

func (c *BaseClient) DeleteQR(ctx context.Context, rtx *whttp.RequestContext, qrCodeID string,
) (*SuccessResponse, error) {
	reqCtx := &whttp.RequestContext{
		Name:          "delete qr code",
		BaseURL:       rtx.BaseURL,
		ApiVersion:    rtx.ApiVersion,
		PhoneNumberID: rtx.PhoneNumberID,
		Endpoints:     []string{"message_qrdls", qrCodeID},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodDelete,
		Query:   map[string]string{"access_token": rtx.Bearer},
	}
	var resp SuccessResponse
	err := c.base.Do(ctx, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("qr code delete: %w", err)
	}

	return &resp, nil
}
