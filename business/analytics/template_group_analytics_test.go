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

package analytics_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/business/analytics"
	"github.com/piusalfred/whatsapp/internal/test"
)

func TestFetchTemplateGroupAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"data": []map[string]any{
			{
				"granularity": "DAILY",
				"data_points": []map[string]any{
					{
						"template_group_id": "1044106240855852",
						"start":             1739491200,
						"end":               1739577600,
						"sent":              1460,
						"delivered":         1460,
						"read":              1399,
					},
				},
			},
		},
		"paging": map[string]any{
			"cursors": map[string]string{
				"before": "MAZDZD",
				"after":  "MjQZD",
			},
		},
	})

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    payload,
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.FetchTemplateGroupAnalytics(context.Background(),
		&analytics.TemplateGroupAnalyticsRequest{
			Start:            1738465116,
			End:              1739559516,
			TemplateGroupIDs: []string{"1044106240855852"},
			MetricTypes:      []string{"sent", "delivered", "read"},
		})
	test.AssertNoError(t, "FetchTemplateGroupAnalytics failed", err)

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(resp.Data))
	}
	if len(resp.Data[0].DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(resp.Data[0].DataPoints))
	}
	dp := resp.Data[0].DataPoints[0]
	if dp.TemplateGroupID != "1044106240855852" {
		t.Errorf("expected template_group_id 1044106240855852, got %s", dp.TemplateGroupID)
	}
	if dp.Sent != 1460 {
		t.Errorf("expected sent 1460, got %d", dp.Sent)
	}

	reqs := srv.GetRequests()
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
	wantPath := "/v25.0/102290129340398/template_group_analytics"
	if r.Path != wantPath {
		t.Errorf("expected path %s, got %s", wantPath, r.Path)
	}
}

func TestTemplateGroupAnalytics_InvalidGroupIDs(t *testing.T) {
	t.Parallel()

	client := analytics.NewClient(mockConfig("https://graph.facebook.com"))

	_, err := client.FetchTemplateGroupAnalytics(context.Background(),
		&analytics.TemplateGroupAnalyticsRequest{
			Start:            1738465116,
			End:              1739559516,
			TemplateGroupIDs: nil,
		})
	if err == nil {
		t.Fatal("expected error for empty template group IDs, got nil")
	}
}

func TestTemplateGroupAnalyticsDataPoint_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	dp := &analytics.TemplateGroupAnalyticsDataPoint{
		TemplateGroupID: "1044106240855852",
		Start:           1739491200,
		End:             1739577600,
		Sent:            1460,
		Delivered:       1460,
		Read:            1399,
	}
	test.AssertJSONRoundTrip(t, "template group analytics data point round-trip", dp)
}

func TestTemplateGroupAnalytics_FromDocumentation(t *testing.T) {
	t.Parallel()

	const docJSON = `{
		"data": [
			{
				"granularity": "DAILY",
				"data_points": [
					{
						"template_group_id": "1044106240855852",
						"start": 1739491200,
						"end": 1739577600,
						"sent": 1460,
						"delivered": 1460,
						"read": 1399
					}
				]
			}
		],
		"paging": {
			"cursors": {
				"before": "MAZDZD",
				"after": "MjQZD"
			}
		}
	}`

	var response analytics.TemplateGroupAnalyticsResponse
	test.AssertJSONUnmarshal(t, "template group analytics from docs", docJSON, &response)

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(response.Data))
	}
	if len(response.Data[0].DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(response.Data[0].DataPoints))
	}
	dp := response.Data[0].DataPoints[0]
	if dp.TemplateGroupID != "1044106240855852" {
		t.Errorf("expected template_group_id 1044106240855852, got %s", dp.TemplateGroupID)
	}
	if dp.Sent != 1460 {
		t.Errorf("expected sent 1460, got %d", dp.Sent)
	}
}

func TestTemplateGroupAnalytics_WithCostAndClicks(t *testing.T) {
	t.Parallel()

	const docJSON = `{
		"data": [
			{
				"granularity": "DAILY",
				"data_points": [
					{
						"template_group_id": "1044106240855852",
						"start": 1739491200,
						"end": 1739577600,
						"sent": 100,
						"delivered": 95,
						"read": 80,
						"clicked": [
							{
								"type": "url_button",
								"button_content": "Learn More",
								"count": 50
							}
						],
						"cost": [
							{
								"type": "amount_spent",
								"value": 5.25
							}
						]
					}
				]
			}
		]
	}`

	var response analytics.TemplateGroupAnalyticsResponse
	test.AssertJSONUnmarshal(t, "template group analytics with cost and clicks", docJSON, &response)

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(response.Data))
	}
	dp := response.Data[0].DataPoints[0]
	if len(dp.Clicked) != 1 {
		t.Fatalf("expected 1 clicked entry, got %d", len(dp.Clicked))
	}
	if dp.Clicked[0].Count != 50 {
		t.Errorf("expected clicked count 50, got %d", dp.Clicked[0].Count)
	}
	if len(dp.Cost) != 1 {
		t.Fatalf("expected 1 cost entry, got %d", len(dp.Cost))
	}
	if dp.Cost[0].Value != 5.25 {
		t.Errorf("expected cost value 5.25, got %f", dp.Cost[0].Value)
	}
}
