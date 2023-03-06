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

package qrcodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	whttp "github.com/piusalfred/whatsapp/http"
)

var ErrUnexpectedResponseCode = fmt.Errorf("unexpected response code")

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

func Create(ctx context.Context, client *http.Client, rtx *RequestContext,
	req *CreateRequest,
) (*CreateResponse, error) {
	queryParams := map[string]string{
		"prefilled_message": req.PrefilledMessage,
		"generate_qr_image": string(req.ImageFormat),
		"access_token":      rtx.AccessToken,
	}
	params := &whttp.RequestParams{
		Query: queryParams,
	}

	reqURL, err := whttp.CreateRequestURL(rtx.BaseURL, rtx.ApiVersion, rtx.PhoneID, "message_qrdls")
	if err != nil {
		return nil, fmt.Errorf("qr code create: %w", err)
	}

	response, err := whttp.Send(ctx, client, http.MethodPost, reqURL, params, nil)
	if err != nil {
		return nil, fmt.Errorf("qr code create: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qr code create: %w: status code: %d", ErrUnexpectedResponseCode, response.StatusCode)
	}

	resp := CreateResponse{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("qr code create: %w", err)
	}

	return &resp, nil
}

func List(ctx context.Context, client *http.Client, rctx *RequestContext) (*ListResponse, error) {
	var (
		resp     ListResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(rctx.BaseURL, rctx.ApiVersion, rctx.PhoneID, "message_qrdls")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", rctx.AccessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type RequestContext struct {
	BaseURL     string `json:"-"`
	PhoneID     string `json:"-"`
	ApiVersion  string `json:"-"`
	AccessToken string `json:"-"`
}

func Get(ctx context.Context, client *http.Client, reqCtx *RequestContext, qrCodeID string) (*Information, error) {
	var (
		list     ListResponse
		resp     Information
		respBody []byte
	)
	requestURL, err := url.JoinPath(reqCtx.BaseURL, reqCtx.ApiVersion, reqCtx.PhoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", reqCtx.AccessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(respBody, &list)
	if err != nil {
		return nil, err
	}

	if len(list.Data) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	resp = *list.Data[0]

	return &resp, nil
}

func Update(ctx context.Context, client *http.Client, rtx *RequestContext, qrCodeID string, req *CreateRequest) (
	*SuccessResponse, error,
) {
	var (
		resp     SuccessResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(rtx.BaseURL, rtx.ApiVersion, rtx.PhoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("qr code update: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("qr code update: %w", err)
	}

	q := request.URL.Query()
	q.Add("prefilled_message", req.PrefilledMessage)
	q.Add("generate_qr_image", string(req.ImageFormat))
	q.Add("access_token", rtx.AccessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("qr code update: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code: %d", ErrUnexpectedResponseCode, response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("qr code update: %w", err)
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("qr code update: %w", err)
	}

	return &resp, nil
}

func Delete(ctx context.Context, client *http.Client, rtx *RequestContext, qrCodeID string) (*SuccessResponse, error) {
	var (
		resp     SuccessResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(rtx.BaseURL, rtx.ApiVersion, rtx.PhoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("qr code delete: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("qr code delete: %w", err)
	}

	q := request.URL.Query()
	q.Add("access_token", rtx.AccessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("qr code delete: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code: %d", ErrUnexpectedResponseCode, response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("qr code delete: %w", err)
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("qr code delete: %w", err)
	}

	return &resp, nil
}
