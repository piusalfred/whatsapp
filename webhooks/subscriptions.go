package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

	SubcribedApp struct {
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
		Data []*SubcribedApp `json:"data"`
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
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: failed to read subscription response body: %w", err)
	}

	var response CreateSubscriptionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("whatsapp: failed to unmarshal subscription response: %w", err)
	}

	return &response, nil
}

// ListSubscriptions
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
	bodyBytes, err := ioutil.ReadAll(resp.Body)
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
