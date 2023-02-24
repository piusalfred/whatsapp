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

const (
	ImageFormatPNG ImageFormat = "PNG"
	ImageFormatSVG ImageFormat = "SVG"
)

type (
	ImageFormat string

	CreateRequest struct {
		BaseURL          string      `json:"-"`
		PhoneID          string      `json:"-"`
		ApiVersion       string      `json:"-"`
		AccessToken      string      `json:"-"`
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

func Create(ctx context.Context, client *http.Client, req *CreateRequest) (*CreateResponse, error) {
	queryParams := map[string]string{
		"prefilled_message": req.PrefilledMessage,
		"generate_qr_image": string(req.ImageFormat),
		"access_token":      req.AccessToken,
	}
	params := &whttp.RequestParams{
		Method:     http.MethodPost,
		SenderID:   req.PhoneID,
		ApiVersion: req.ApiVersion,
		//Bearer:     req.AccessToken, // token is passed as a query param
		BaseURL:   req.BaseURL,
		Endpoints: []string{"message_qrdls"},
		Query:     queryParams,
	}

	response, err := whttp.Send(ctx, client, params, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	resp := CreateResponse{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func List(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken string) (*ListResponse, error) {
	var (
		resp     ListResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
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

func Get(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken, qrCodeID string) (*Information, error) {
	var (
		list     ListResponse
		resp     Information
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
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

func Update(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken, qrCodeID string, req *CreateRequest) (*SuccessResponse, error) {
	var (
		resp     SuccessResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("prefilled_message", req.PrefilledMessage)
	q.Add("generate_qr_image", string(req.ImageFormat))
	q.Add("access_token", accessToken)
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

func Delete(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken, qrCodeID string) (*SuccessResponse, error) {
	var (
		resp     SuccessResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
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
