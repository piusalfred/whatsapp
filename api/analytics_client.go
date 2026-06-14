//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package api

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/business/analytics"
	"github.com/piusalfred/whatsapp/config"
)

func (c *Client) FetchMessagingAnalytics(
	ctx context.Context,
	req *analytics.MessagingRequest,
) (*analytics.MessagingResponse, error) {
	resp, err := c.sender.FetchMessagingAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch messaging analytics: %w", err)
	}
	return resp, nil
}

func (c *Client) FetchConversationAnalytics(
	ctx context.Context,
	req *analytics.ConversationalRequest,
) (*analytics.ConversationalResponse, error) {
	resp, err := c.sender.FetchConversationAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}
	return resp, nil
}

func (c *Client) FetchPricingAnalytics(
	ctx context.Context,
	req *analytics.PricingRequest,
) (*analytics.PricingResponse, error) {
	resp, err := c.sender.FetchPricingAnalytics(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch pricing analytics: %w", err)
	}
	return resp, nil
}

func (c *Client) EnableTemplatesAnalytics(ctx context.Context) (string, error) {
	id, err := c.sender.getAnalytics().Enable(ctx, c.config)
	if err != nil {
		return "", fmt.Errorf("enable templates analytics: %w", err)
	}
	return id, nil
}

func (c *Client) DisableButtonClickTracking(
	ctx context.Context,
	req *analytics.DisableButtonClickTrackingRequest,
) (*analytics.DisableButtonClickTrackingResponse, error) {
	resp, err := c.sender.getAnalytics().DisableButtonClickTracking(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("disable button click tracking: %w", err)
	}
	return resp, nil
}

func (c *Client) FetchTemplateAnalytics(
	ctx context.Context,
	req *analytics.TemplateAnalyticsRequest,
) (*analytics.TemplateAnalyticsResponse, error) {
	resp, err := c.sender.getAnalytics().Fetch(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("fetch template analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchMessagingAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.MessagingRequest,
) (*analytics.MessagingResponse, error) {
	resp, err := bc.getAnalytics().FetchGeneralAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch messaging analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchConversationAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.ConversationalRequest,
) (*analytics.ConversationalResponse, error) {
	resp, err := bc.getAnalytics().FetchConversationAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchPricingAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.PricingRequest,
) (*analytics.PricingResponse, error) {
	resp, err := bc.getAnalytics().FetchPricingAnalytics(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch pricing analytics: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) EnableTemplatesAnalytics(ctx context.Context, conf *config.Config) (string, error) {
	id, err := bc.getAnalytics().Enable(ctx, conf)
	if err != nil {
		return "", fmt.Errorf("enable templates analytics: %w", err)
	}
	return id, nil
}

func (bc *BaseClient) DisableButtonClickTracking(
	ctx context.Context,
	conf *config.Config,
	req *analytics.DisableButtonClickTrackingRequest,
) (*analytics.DisableButtonClickTrackingResponse, error) {
	resp, err := bc.getAnalytics().DisableButtonClickTracking(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("disable button click tracking: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) FetchTemplateAnalytics(
	ctx context.Context,
	conf *config.Config,
	req *analytics.TemplateAnalyticsRequest,
) (*analytics.TemplateAnalyticsResponse, error) {
	resp, err := bc.getAnalytics().Fetch(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("fetch template analytics: %w", err)
	}
	return resp, nil
}
