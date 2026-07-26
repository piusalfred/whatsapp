//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package analytics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	// TemplateGroupAnalyticsRequest is the user-facing request for template group analytics.
	TemplateGroupAnalyticsRequest struct {
		Start            int64
		End              int64
		TemplateGroupIDs []string
		MetricTypes      []string
	}

	// TemplateGroupAnalyticsResponse contains the template group analytics data returned by the API.
	TemplateGroupAnalyticsResponse struct {
		Data   []TemplateGroupAnalyticsData `json:"data,omitempty"`
		Paging *whttp.Paging                `json:"paging,omitempty"`
	}

	// TemplateGroupAnalyticsData holds a set of data points for a given granularity.
	TemplateGroupAnalyticsData struct {
		Granularity string                            `json:"granularity,omitempty"`
		DataPoints  []TemplateGroupAnalyticsDataPoint `json:"data_points,omitempty"`
	}

	// TemplateGroupAnalyticsDataPoint represents a single data point in template group analytics.
	TemplateGroupAnalyticsDataPoint struct {
		TemplateGroupID string               `json:"template_group_id,omitempty"`
		Start           int64                `json:"start,omitempty"`
		End             int64                `json:"end,omitempty"`
		Sent            int64                `json:"sent,omitempty"`
		Delivered       int64                `json:"delivered,omitempty"`
		Read            int64                `json:"read,omitempty"`
		Clicked         []TemplateClicked    `json:"clicked,omitempty"`
		Cost            []TemplateCostMetric `json:"cost,omitempty"`
	}
)

// ErrInvalidTemplateGroupIDs is returned when the template group IDs slice is empty
// or exceeds the maximum of 10.
var ErrInvalidTemplateGroupIDs = errors.New("invalid number of template group IDs")

// FetchTemplateGroupAnalytics retrieves template group analytics for the specified
// template groups within the given date range.
func (c *Client) FetchTemplateGroupAnalytics(ctx context.Context,
	params *TemplateGroupAnalyticsRequest,
) (*TemplateGroupAnalyticsResponse, error) {
	return c.sender.FetchTemplateGroupAnalytics(ctx, c.config, params)
}

// FetchTemplateGroupAnalytics retrieves template group analytics for the specified
// template groups within the given date range.
// This is the multi-tenant variant that accepts a per-call config.
func (bc *BaseClient) FetchTemplateGroupAnalytics(ctx context.Context, conf *config.Config,
	params *TemplateGroupAnalyticsRequest,
) (*TemplateGroupAnalyticsResponse, error) {
	queryParams := map[string]string{}

	queryParams["start"] = strconv.FormatInt(params.Start, 10)
	queryParams["end"] = strconv.FormatInt(params.End, 10)
	queryParams["granularity"] = GranularityDaily.String()

	if len(params.TemplateGroupIDs) > 0 && len(params.TemplateGroupIDs) <= 10 {
		queryParams["template_group_ids"] = "[" + strings.Join(params.TemplateGroupIDs, ",") + "]"
	} else {
		return nil, fmt.Errorf("%w: count: %d: should be >0 and < 11",
			ErrInvalidTemplateGroupIDs, len(params.TemplateGroupIDs))
	}

	if len(params.MetricTypes) > 0 {
		queryParams["metric_types"] = strings.Join(params.MetricTypes, ",")
	}

	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeFetchTemplateGroupAnalytics).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "template_group_analytics").
		QueryParams(queryParams)

	request := whttp.BuildRequest(b, (*BaseRequest)(nil))

	response := &TemplateGroupAnalyticsResponse{}
	decoder := whttp.NewResponseCapturer(whttp.ResponseDecoderJSON(response, whttp.DecodeOptionsStrict()))

	if err := bc.BaseClient.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("fetch template group analytics: %w", err)
	}

	return response, nil
}
