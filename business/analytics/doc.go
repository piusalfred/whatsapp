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

// Package analytics provides access to the WhatsApp Business Account Analytics API.
//
// The Analytics API offers seven categories of analytics:
//
//   - Messaging analytics: number and type of messages sent and delivered
//     by the phone numbers associated with a WABA.
//   - Conversation analytics: cost and conversation information for a WABA.
//   - Pricing analytics: pricing breakdowns for messages delivered within a date range.
//   - Template analytics: sent, delivered, read, and click metrics for message templates.
//   - Call analytics: number and type of calls made and received.
//   - Group analytics: messages sent, delivered, read, and participant activity in groups.
//   - Template group analytics: metrics for templates within a template group.
//
// # Messaging, conversation, pricing, and call analytics
//
// These share a common request pattern using the "fields" query parameter.
// Create a Client with NewClient and call the corresponding Fetch method:
//
//	client := analytics.NewClient(config)
//
//	// Messaging analytics
//	resp, _ := client.FetchGeneralAnalytics(ctx, &analytics.MessagingRequest{
//	    Start: 1700000000, End: 1700604800,
//	    Granularity: analytics.GranularityDay,
//	    Options: []analytics.MessagingQueryParamsOption{
//	        analytics.WithMessagingPhoneNumbers("15551234567"),
//	    },
//	})
//
//	// Conversation analytics
//	convResp, _ := client.FetchConversationAnalytics(ctx, &analytics.ConversationalRequest{
//	    Start: 1700000000, End: 1700604800,
//	    Granularity: analytics.GranularityMonthly,
//	    Options: []analytics.ConversationalQueryParamsOption{
//	        analytics.WithConversationalDimensions(analytics.DimensionCountry),
//	    },
//	})
//
//	// Pricing analytics
//	pricingResp, _ := client.FetchPricingAnalytics(ctx, &analytics.PricingRequest{
//	    Start: 1700000000, End: 1700604800,
//	    Granularity: analytics.GranularityDaily,
//	})
//
//	// Call analytics
//	callResp, _ := client.FetchCallAnalytics(ctx, &analytics.CallAnalyticsRequest{
//	    Start: 1700000000, End: 1700604800,
//	    Granularity: analytics.GranularityDaily,
//	    Options: []analytics.CallAnalyticsQueryParamsOption{
//	        analytics.WithCallDirections(analytics.CallDirectionUserInitiated),
//	    },
//	})
//
// # Template analytics
//
// Template analytics must first be enabled via the Enable method (one-time confirmation).
// Once enabled, use Fetch to retrieve template metrics:
//
//	client.Enable(ctx)
//	templates, _ := client.Fetch(ctx, &analytics.TemplateAnalyticsRequest{
//	    Start: 1700000000, End: 1700604800,
//	    Templates: []string{"template_id_1"},
//	})
//
// Button click tracking can be disabled per-template via DisableButtonClickTracking.
//
// # Group and template group analytics
//
// These use dedicated endpoints (not the fields parameter). They follow a simpler
// request pattern:
//
//	// Group analytics
//	groups, _ := client.FetchGroupAnalytics(ctx, &analytics.GroupAnalyticsRequest{
//	    Start: 1700000000, End: 1700604800,
//	    GroupIDs: []string{"GROUP_ID"},
//	    MetricTypes: []analytics.GroupAnalyticsMetricType{
//	        analytics.GroupMetricSent,
//	        analytics.GroupMetricRead,
//	    },
//	})
//
//	// Template group analytics
//	tgResp, _ := client.FetchTemplateGroupAnalytics(ctx,
//	    &analytics.TemplateGroupAnalyticsRequest{
//	        Start: 1700000000, End: 1700604800,
//	        TemplateGroupIDs: []string{"template_group_id"},
//	    })
//
// # Multi-tenant usage
//
// All BaseClient methods accept a per-call config for dynamic credential rotation.
// The high-level Client methods delegate to these, passing their fixed config.
//
// # Limitations
//
//   - The maximum lookback window for messaging, conversation, and pricing analytics
//     is 1 year (as of December 1, 2025).
//   - Template and template group analytics have a 90-day lookback window.
//   - Group analytics has a 90-day lookback window.
//   - Template IDs: maximum 10 per request.
//   - Template group IDs: maximum 10 per request.
//   - Group IDs: maximum 1 per request.
//   - Button click analytics are only available for templates categorized as
//     MARKETING or UTILITY.
//   - COST metrics are not returned for WABAs that share a Solution Partner's
//     credit line.
package analytics
