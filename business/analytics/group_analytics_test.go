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

func TestFetchGroupAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"data": []map[string]any{
			{
				"granularity": "DAILY",
				"data_points": []map[string]any{
					{
						"group_id":  "GROUP_ID",
						"start":     1685548801,
						"end":       1685635200,
						"sent":      100,
						"delivered": 250,
						"read":      200,
						"joined":    3,
						"left":      1,
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
	resp, err := client.FetchGroupAnalytics(context.Background(), &analytics.GroupAnalyticsRequest{
		Start:    1764662400,
		End:      1764921600,
		GroupIDs: []string{"GROUP_ID"},
		MetricTypes: []analytics.GroupAnalyticsMetricType{
			analytics.GroupMetricSent,
			analytics.GroupMetricDelivered,
			analytics.GroupMetricRead,
			analytics.GroupMetricParticipantsJoined,
			analytics.GroupMetricParticipantsLeft,
		},
	})
	test.AssertNoError(t, "FetchGroupAnalytics failed", err)

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(resp.Data))
	}
	if len(resp.Data[0].DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(resp.Data[0].DataPoints))
	}
	dp := resp.Data[0].DataPoints[0]
	if dp.GroupID != "GROUP_ID" {
		t.Errorf("expected group_id GROUP_ID, got %s", dp.GroupID)
	}
	if dp.Sent != 100 {
		t.Errorf("expected sent 100, got %d", dp.Sent)
	}
	if dp.Joined != 3 {
		t.Errorf("expected joined 3, got %d", dp.Joined)
	}
	if dp.Left != 1 {
		t.Errorf("expected left 1, got %d", dp.Left)
	}

	reqs := srv.GetRequests()
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
	wantPath := "/v25.0/102290129340398/group_analytics"
	if r.Path != wantPath {
		t.Errorf("expected path %s, got %s", wantPath, r.Path)
	}
}

func TestGroupAnalytics_InvalidGroupIDs(t *testing.T) {
	t.Parallel()

	client := analytics.NewClient(mockConfig("https://graph.facebook.com"))

	_, err := client.FetchGroupAnalytics(context.Background(), &analytics.GroupAnalyticsRequest{
		Start:    1764662400,
		End:      1764921600,
		GroupIDs: nil,
	})
	if err == nil {
		t.Fatal("expected error for empty group IDs, got nil")
	}

	_, err = client.FetchGroupAnalytics(context.Background(), &analytics.GroupAnalyticsRequest{
		Start:    1764662400,
		End:      1764921600,
		GroupIDs: []string{"ID1", "ID2"},
	})
	if err == nil {
		t.Fatal("expected error for multiple group IDs, got nil")
	}
}

func TestGroupAnalyticsDataPoint_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	dp := &analytics.GroupAnalyticsDataPoint{
		GroupID:   "GROUP_ID",
		Start:     1685548801,
		End:       1685635200,
		Sent:      100,
		Delivered: 250,
		Read:      200,
		Joined:    3,
		Left:      1,
	}
	test.AssertJSONRoundTrip(t, "group analytics data point round-trip", dp)
}

func TestGroupAnalytics_FromDocumentation(t *testing.T) {
	t.Parallel()

	const docJSON = `{
		"data": [
			{
				"granularity": "DAILY",
				"data_points": [
					{
						"group_id": "GROUP_ID",
						"start": 1685548801,
						"end": 1685635200,
						"sent": 100,
						"delivered": 250,
						"read": 200,
						"joined": 3,
						"left": 1
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

	var response analytics.GroupAnalyticsResponse
	test.AssertJSONUnmarshal(t, "group analytics from docs", docJSON, &response)

	if len(response.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(response.Data))
	}
	if len(response.Data[0].DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(response.Data[0].DataPoints))
	}
	dp := response.Data[0].DataPoints[0]
	if dp.Sent != 100 {
		t.Errorf("expected sent 100, got %d", dp.Sent)
	}
	if dp.Joined != 3 {
		t.Errorf("expected joined 3, got %d", dp.Joined)
	}
}
