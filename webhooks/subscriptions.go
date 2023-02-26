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

package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (

	// SubscriptionRequest ....
	// ApiVersion: The API version of the subscription.
	// AccountID: The  WhatsApp Business account ID.
	SubscriptionRequest struct {
		ApiVersion string
		Token      string
		AccountID  string
	}

	// CreateSubscriptionResponse ....
	CreateSubscriptionResponse struct {
		Success bool
	}

	DeleteSubscriptionResponse struct {
		Success bool
	}

	SubscribedApp struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}

	// ListSubscriptionsResponse contains a list of subscribed apps.
	//{
	// "data": [
	// {
	//"name": "<APP_NAME>",
	//"id": "<APP_ID>" } ]
	//}
	ListSubscriptionsResponse struct {
		Data []*SubscribedApp `json:"data"`
	}
)

func CreateSubscription(ctx context.Context, client *http.Client, baseURL string, req *SubscriptionRequest) (*CreateSubscriptionResponse, error) {
	reqURL, err := url.JoinPath(baseURL, req.ApiVersion, req.AccountID, "subscribed_apps")
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to read subscription response body: %w", err)
	}

	var response CreateSubscriptionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("whatsapp: failed to unmarshal subscription response: %w", err)
	}

	return &response, nil
}

func ListSubscriptions(ctx context.Context, client *http.Client, baseURL string, req *SubscriptionRequest) (*ListSubscriptionsResponse, error) {
	reqURL, err := url.JoinPath(baseURL, req.ApiVersion, req.AccountID, "subscribed_apps")
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to read subscription response body: %w", err)
	}

	var response ListSubscriptionsResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("whatsapp: failed to unmarshal subscription response: %w", err)
	}

	return &response, nil
}

// DeleteSubscription ...
func DeleteSubscription(ctx context.Context, client *http.Client, baseURL string, req *SubscriptionRequest) (*DeleteSubscriptionResponse, error) {
	reqURL, err := url.JoinPath(baseURL, req.ApiVersion, req.AccountID, "subscribed_apps")
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("whatsapp: failed to create subscription: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to read subscription response body: %w", err)
	}

	var response DeleteSubscriptionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("whatsapp: failed to unmarshal subscription response: %w", err)
	}

	return &response, nil
}
