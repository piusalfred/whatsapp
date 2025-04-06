package analytics_test

import (
	"testing"

	"github.com/piusalfred/whatsapp/business/analytics"
)

func TestQueryParamsString(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name     string
		generate func() *analytics.Request
		expected string
	}

	tests := []testCase{
		// General MessagingAnalytics Example
		{
			name: "General MessagingAnalytics Example",
			generate: func() *analytics.Request {
				return analytics.MakeMessagingAnalyticsQueryParams(
					1685602800,
					1688194800,
					analytics.GranularityMonthly,
					analytics.WithMessagingPhoneNumbers("16505550111", "16505550112"),
					analytics.WithMessagingProductTypes(0, 2),
					analytics.WithMessagingCountryCodes("US", "CA"),
				)
			},
			expected: "analytics.start(1685602800).end(1688194800).granularity(MONTHLY).phone_numbers([\"16505550111\",\"16505550112\"]).country_codes([\"US\",\"CA\"]).product_types([0,2])",
		},
		// Conversation MessagingAnalytics Example 1
		{
			name: "Conversation MessagingAnalytics Example 1",
			generate: func() *analytics.Request {
				return analytics.MakeConversationalAnalyticsQueryParams(
					1685602800,
					1688194800,
					analytics.GranularityMonthly,
					analytics.WithConversationalDimensions(
						analytics.DimensionConversationCategory,
						analytics.DimensionConversationType,
						analytics.DimensionCountry,
						analytics.DimensionPhone,
					),
				)
			},
			expected: "conversation_analytics.start(1685602800).end(1688194800).granularity(MONTHLY).phone_numbers([]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\",\"COUNTRY\",\"PHONE\"])",
		},
		// Conversation MessagingAnalytics Example 2
		{
			name: "Conversation MessagingAnalytics Example 2",
			generate: func() *analytics.Request {
				return analytics.MakeConversationalAnalyticsQueryParams(
					1685602800,
					1685689200,
					analytics.GranularityHalfHour,
					analytics.WithConversationalPhoneNumbers("19195552584"),
					analytics.WithConversationalDimensions(
						analytics.DimensionConversationCategory,
						analytics.DimensionConversationType,
						analytics.DimensionCountry,
						analytics.DimensionPhone,
					),
				)
			},
			expected: "conversation_analytics.start(1685602800).end(1685689200).granularity(HALF_HOUR).phone_numbers([\"19195552584\"]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\",\"COUNTRY\",\"PHONE\"])",
		},
		// Conversation MessagingAnalytics Example 3
		{
			name: "Conversation MessagingAnalytics Example 3",
			generate: func() *analytics.Request {
				return analytics.MakeConversationalAnalyticsQueryParams(
					1685527200,
					1685613600,
					analytics.GranularityHalfHour,
					analytics.WithConversationalCategories(
						analytics.ConversationalCategoryMarketing,
						analytics.ConversationalCategoryAuthentications,
					),
					analytics.WithConversationalDimensions(
						analytics.DimensionConversationCategory,
						analytics.DimensionConversationType,
					),
				)
			},
			expected: "conversation_analytics.start(1685527200).end(1685613600).granularity(HALF_HOUR).phone_numbers([]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\"]).conversation_categories([\"MARKETING\",\"AUTHENTICATION\"])",
		},
		// Pricing MessagingAnalytics Example
		{
			name: "Pricing MessagingAnalytics Example",
			generate: func() *analytics.Request {
				return analytics.MakePricingAnalyticsQueryParams(
					1685602800,
					1688194800,
					analytics.GranularityMonthly,
					analytics.WithPricingPhoneNumbers("19195552584", "19195552585"),
					analytics.WithPricingMetricTypes(analytics.MetricTypeCost),
					analytics.WithPricingTypes(analytics.PricingTypeRegular),
					analytics.WithPricingCategories(analytics.PricingCategoryMarketing),
					analytics.WithPricingDimensions(
						analytics.DimensionCountry,
						analytics.DimensionPhone,
					),
				)
			},
			expected: "pricing_analytics.start(1685602800).end(1688194800).granularity(MONTHLY).phone_numbers([\"19195552584\",\"19195552585\"]).dimensions([\"COUNTRY\",\"PHONE\"]).metric_types([\"COST\"]).pricing_types([\"REGULAR\"]).pricing_categories([\"MARKETING\"])",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			req := test.generate()
			result := req.QueryParamsString()
			if result != test.expected {
				t.Errorf("Failed.\nExpected:\n%s\nGot:\n%s", test.expected, result)
			}
		})
	}
}
