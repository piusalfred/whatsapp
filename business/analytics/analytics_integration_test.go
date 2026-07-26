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
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/internal/test"
)

func mockConfig(baseURL string) *config.Config {
	return &config.Config{
		BaseURL:           baseURL,
		APIVersion:        "v25.0",
		BusinessAccountID: "102290129340398",
		AccessToken:       "EAAJB...",
	}
}

func TestFetchGeneralAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"analytics": map[string]any{
			"phone_numbers": []string{"16505550111"},
			"country_codes": []string{"US"},
			"granularity":   "DAY",
			"data_points": []map[string]any{
				{
					"start":     1543543200,
					"end":       1543629600,
					"sent":      196093,
					"delivered": 179715,
				},
			},
		},
		"id": "102290129340398",
	})

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    payload,
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.FetchGeneralAnalytics(context.Background(), &analytics.MessagingRequest{
		Start:       1543543200,
		End:         1544148000,
		Granularity: analytics.GranularityDay,
	})
	test.AssertNoError(t, "FetchGeneralAnalytics failed", err)

	if resp.Analytics == nil {
		t.Fatal("expected analytics data, got nil")
	}
	if len(resp.Analytics.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(resp.Analytics.DataPoints))
	}

	reqs := srv.GetRequests()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
	wantPath := "/v25.0/102290129340398"
	if r.Path != wantPath {
		t.Errorf("expected path %s, got %s", wantPath, r.Path)
	}
	if fields := r.QueryParams.Get("fields"); fields == "" {
		t.Error("expected fields query param, got empty")
	}
}

func TestFetchConversationAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"conversation_analytics": map[string]any{
			"data": []map[string]any{
				{
					"data_points": []map[string]any{
						{
							"start":                  1685602800,
							"end":                    1688194800,
							"conversation":           1558,
							"phone_number":           "15550458206",
							"country":                "US",
							"conversation_type":      "REGULAR",
							"conversation_direction": "UNKNOWN",
							"conversation_category":  "AUTHENTICATION",
							"cost":                   15.58,
						},
					},
				},
			},
		},
		"id": "102290129340398",
	})

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    payload,
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.FetchConversationAnalytics(context.Background(), &analytics.ConversationalRequest{
		Start:       1685602800,
		End:         1688194800,
		Granularity: analytics.GranularityMonthly,
		Options: []analytics.ConversationalQueryParamsOption{
			analytics.WithConversationalDimensions(
				analytics.DimensionConversationCategory,
				analytics.DimensionConversationType,
				analytics.DimensionCountry,
				analytics.DimensionPhone,
			),
		},
	})
	test.AssertNoError(t, "FetchConversationAnalytics failed", err)

	if resp.ConversationAnalytics == nil {
		t.Fatal("expected conversation analytics data, got nil")
	}

	reqs := srv.GetRequests()
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
}

func TestFetchPricingAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"pricing_analytics": map[string]any{
			"data": []map[string]any{
				{
					"data_points": []map[string]any{
						{
							"start":            1748761200,
							"end":              1748847600,
							"country":          "US",
							"pricing_type":     "REGULAR",
							"pricing_category": "MARKETING",
							"tier":             "0:MAX",
							"volume":           4,
							"cost":             40,
						},
					},
				},
			},
		},
		"id": "102290129340398",
	})

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    payload,
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.FetchPricingAnalytics(context.Background(), &analytics.PricingRequest{
		Start:       1748761200,
		End:         1749687703,
		Granularity: analytics.GranularityDaily,
	})
	test.AssertNoError(t, "FetchPricingAnalytics failed", err)

	if resp.PricingAnalytics == nil {
		t.Fatal("expected pricing analytics data, got nil")
	}

	reqs := srv.GetRequests()
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
}

func TestFetchCallAnalytics_MockServer(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"call_analytics": map[string]any{
			"granularity": "DAILY",
			"directions":  "USER_INITIATED",
			"data_points": []map[string]any{
				{
					"start":            1765958400,
					"end":              1766044800,
					"cost":             0.47795,
					"count":            35,
					"average_duration": 106,
				},
			},
		},
		"id": "102290129340398",
	})

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    payload,
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.FetchCallAnalytics(context.Background(), &analytics.CallAnalyticsRequest{
		Start:       1759302000,
		End:         1767168000,
		Granularity: analytics.GranularityDaily,
		Options: []analytics.CallAnalyticsQueryParamsOption{
			analytics.WithCallDirections(analytics.CallDirectionUserInitiated),
		},
	})
	test.AssertNoError(t, "FetchCallAnalytics failed", err)

	if resp.CallAnalytics == nil {
		t.Fatal("expected call analytics data, got nil")
	}
	if len(resp.CallAnalytics.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(resp.CallAnalytics.DataPoints))
	}
	if resp.CallAnalytics.DataPoints[0].Count != 35 {
		t.Errorf("expected count 35, got %d", resp.CallAnalytics.DataPoints[0].Count)
	}

	reqs := srv.GetRequests()
	r := reqs[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
	wantPath := "/v25.0/102290129340398"
	if r.Path != wantPath {
		t.Errorf("expected path %s, got %s", wantPath, r.Path)
	}
	if fields := r.QueryParams.Get("fields"); fields == "" {
		t.Error("expected fields query param, got empty")
	}
}

func TestBaseResponse_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("MessagingAnalytics", func(t *testing.T) {
		t.Parallel()
		response := &analytics.BaseResponse{
			Messaging: &analytics.MessagingAnalytics{
				DataPoints: []*analytics.DataPoint{
					{Start: 1543543200, End: 1543629600, Sent: 196093, Delivered: 179715},
				},
			},
			ID: "102290129340398",
		}
		test.AssertJSONRoundTrip(t, "messaging analytics round-trip", response)
	})

	t.Run("ConversationAnalytics", func(t *testing.T) {
		t.Parallel()
		response := &analytics.BaseResponse{
			Conversational: &analytics.ConversationAnalytics{
				Data: []*analytics.Data{
					{
						DataPoints: []*analytics.DataPoint{
							{
								Start: 1685602800, End: 1688194800,
								Conversation: 1558, PhoneNumber: "15550458206",
								Country: "US", Cost: 15.58,
								ConversationCategory: "AUTHENTICATION",
							},
						},
					},
				},
			},
			ID: "102290129340398",
		}
		test.AssertJSONRoundTrip(t, "conversation analytics round-trip", response)
	})

	t.Run("PricingAnalytics", func(t *testing.T) {
		t.Parallel()
		response := &analytics.BaseResponse{
			Pricing: &analytics.PricingAnalytics{
				Data: []*analytics.Data{
					{
						DataPoints: []*analytics.DataPoint{
							{
								Start: 1748761200, End: 1748847600,
								Country: "US", Tier: "0:MAX",
								PricingType: "REGULAR", PricingCategory: "MARKETING",
								Volume: 4, Cost: 40,
							},
						},
					},
				},
			},
			ID: "102290129340398",
		}
		test.AssertJSONRoundTrip(t, "pricing analytics round-trip", response)
	})

	t.Run("CallAnalytics", func(t *testing.T) {
		t.Parallel()
		response := &analytics.BaseResponse{
			Call: &analytics.CallAnalytics{
				Granularity: "DAILY",
				Directions:  "USER_INITIATED",
				DataPoints: []*analytics.CallAnalyticsDataPoint{
					{
						Start: 1765958400, End: 1766044800,
						Cost: 0.47795, Count: 35, AverageDuration: 106,
					},
				},
			},
			ID: "102290129340398",
		}
		test.AssertJSONRoundTrip(t, "call analytics round-trip", response)
	})
}

func TestCallAnalyticsDataPoint_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	dp := &analytics.CallAnalyticsDataPoint{
		Start:           1765958400,
		End:             1766044800,
		Cost:            0.47795,
		Count:           35,
		AverageDuration: 106,
	}
	test.AssertJSONRoundTrip(t, "call analytics data point round-trip", dp)
}

func TestCallAnalytics_FromDocumentation(t *testing.T) {
	t.Parallel()

	const docJSON = `{
		"call_analytics": {
			"granularity": "DAILY",
			"directions": "USER_INITIATED",
			"data_points": [
				{
					"start": 1765958400,
					"end": 1766044800,
					"cost": 0.47795,
					"count": 35,
					"average_duration": 106
				}
			]
		},
		"id": "102290129340398"
	}`

	var response analytics.BaseResponse
	test.AssertJSONUnmarshal(t, "call analytics from docs", docJSON, &response)

	if response.Call == nil {
		t.Fatal("expected call analytics, got nil")
	}
	if len(response.Call.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(response.Call.DataPoints))
	}
	dp := response.Call.DataPoints[0]
	if dp.Count != 35 {
		t.Errorf("expected count 35, got %d", dp.Count)
	}
	if dp.AverageDuration != 106 {
		t.Errorf("expected average_duration 106, got %d", dp.AverageDuration)
	}
	if dp.Cost != 0.47795 {
		t.Errorf("expected cost 0.47795, got %f", dp.Cost)
	}
	if response.ID != "102290129340398" {
		t.Errorf("expected ID 102290129340398, got %s", response.ID)
	}
}

func TestErrorHandling_MockServer(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusInternalServerError,
		Payload:    []byte(`{"error":{"message":"Internal Server Error"}}`),
	})
	defer srv.Close()

	client := analytics.NewClient(mockConfig(srv.Server.URL))
	_, err := client.FetchGeneralAnalytics(context.Background(), &analytics.MessagingRequest{
		Start:       1543543200,
		End:         1544148000,
		Granularity: analytics.GranularityDay,
	})
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}
