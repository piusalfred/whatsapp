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

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

// GroupAnalyticsMetricType represents the type of metric for group analytics.
type GroupAnalyticsMetricType string

const (
	GroupMetricSent               GroupAnalyticsMetricType = "SENT"
	GroupMetricDelivered          GroupAnalyticsMetricType = "DELIVERED"
	GroupMetricRead               GroupAnalyticsMetricType = "READ"
	GroupMetricParticipantsJoined GroupAnalyticsMetricType = "PARTICIPANTS_JOINED"
	GroupMetricParticipantsLeft   GroupAnalyticsMetricType = "PARTICIPANTS_LEFT"
)

type (
	// GroupAnalyticsRequest is the user-facing request for group analytics.
	GroupAnalyticsRequest struct {
		Start       int64
		End         int64
		GroupIDs    []string
		MetricTypes []GroupAnalyticsMetricType
	}

	// GroupAnalyticsResponse contains the group analytics data returned by the API.
	GroupAnalyticsResponse struct {
		Data   []GroupAnalyticsData `json:"data,omitempty"`
		Paging *whttp.Paging        `json:"paging,omitempty"`
	}

	// GroupAnalyticsData holds a set of data points for a given granularity.
	GroupAnalyticsData struct {
		Granularity string                    `json:"granularity,omitempty"`
		DataPoints  []GroupAnalyticsDataPoint `json:"data_points,omitempty"`
	}

	// GroupAnalyticsDataPoint represents a single data point in group analytics.
	GroupAnalyticsDataPoint struct {
		GroupID   string `json:"group_id,omitempty"`
		Start     int64  `json:"start,omitempty"`
		End       int64  `json:"end,omitempty"`
		Sent      int64  `json:"sent,omitempty"`
		Delivered int64  `json:"delivered,omitempty"`
		Read      int64  `json:"read,omitempty"`
		Joined    int64  `json:"joined,omitempty"`
		Left      int64  `json:"left,omitempty"`
	}
)

// ErrInvalidGroupIDs is returned when the group IDs slice is empty or exceeds the maximum.
var ErrInvalidGroupIDs = errors.New("invalid number of group IDs")

// FetchGroupAnalytics retrieves group analytics for the specified groups within the given date range.
func (c *Client) FetchGroupAnalytics(ctx context.Context, params *GroupAnalyticsRequest) (
	*GroupAnalyticsResponse, error,
) {
	return c.sender.FetchGroupAnalytics(ctx, c.config, params)
}

// FetchGroupAnalytics retrieves group analytics for the specified groups within the given date range.
// This is the multi-tenant variant that accepts a per-call config.
func (bc *BaseClient) FetchGroupAnalytics(ctx context.Context, conf *config.Config,
	params *GroupAnalyticsRequest,
) (*GroupAnalyticsResponse, error) {
	queryParams := map[string]string{}

	queryParams["start"] = strconv.FormatInt(params.Start, 10)
	queryParams["end"] = strconv.FormatInt(params.End, 10)
	queryParams["granularity"] = GranularityDaily.String()

	if len(params.GroupIDs) == 0 || len(params.GroupIDs) > 1 {
		return nil, fmt.Errorf("%w: count: %d: must be exactly 1",
			ErrInvalidGroupIDs, len(params.GroupIDs))
	}
	queryParams["group_ids"] = formatArray(params.GroupIDs, quoteString)

	if len(params.MetricTypes) > 0 {
		metricTypes := make([]string, len(params.MetricTypes))
		for i, m := range params.MetricTypes {
			metricTypes[i] = string(m)
		}
		queryParams["metric_types"] = formatArray(metricTypes, quoteString)
	}

	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeFetchGroupAnalytics).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "group_analytics").
		QueryParams(queryParams)

	request := whttp.BuildRequest(b, (*BaseRequest)(nil))

	response := &GroupAnalyticsResponse{}
	decoder := whttp.NewResponseCapturer(whttp.ResponseDecoderJSON(response, whttp.DecodeOptionsStrict()))

	if err := bc.BaseClient.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("fetch group analytics: %w", err)
	}

	return response, nil
}
