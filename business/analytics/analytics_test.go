package analytics

import "testing"

func TestQueryParamsString(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name     string
		generate func() *Request
		expected string
	}

	tests := []testCase{
		// General MessagingAnalytics Example
		{
			name: "General MessagingAnalytics Example",
			generate: func() *Request {
				return MakeMessagingAnalyticsQueryParams(
					1685602800,
					1688194800,
					GranularityMonthly,
					WithMessagingPhoneNumbers("16505550111", "16505550112"),
					WithMessagingProductTypes(0, 2),
					WithMessagingCountryCodes("US", "CA"),
				)
			},
			expected: "analytics.start(1685602800).end(1688194800).granularity(MONTHLY).phone_numbers([\"16505550111\",\"16505550112\"]).country_codes([\"US\",\"CA\"]).product_types([0,2])",
		},
		// Conversation MessagingAnalytics Example 1
		{
			name: "Conversation MessagingAnalytics Example 1",
			generate: func() *Request {
				return MakeConversationalAnalyticsQueryParams(
					1685602800,
					1688194800,
					GranularityMonthly,
					WithConversationalDimensions(
						DimensionConversationCategory,
						DimensionConversationType,
						DimensionCountry,
						DimensionPhone,
					),
				)
			},
			expected: "conversation_analytics.start(1685602800).end(1688194800).granularity(MONTHLY).phone_numbers([]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\",\"COUNTRY\",\"PHONE\"])",
		},
		// Conversation MessagingAnalytics Example 2
		{
			name: "Conversation MessagingAnalytics Example 2",
			generate: func() *Request {
				return MakeConversationalAnalyticsQueryParams(
					1685602800,
					1685689200,
					GranularityHalfHour,
					WithConversationalPhoneNumbers("19195552584"),
					WithConversationalDimensions(
						DimensionConversationCategory,
						DimensionConversationType,
						DimensionCountry,
						DimensionPhone,
					),
				)
			},
			expected: "conversation_analytics.start(1685602800).end(1685689200).granularity(HALF_HOUR).phone_numbers([\"19195552584\"]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\",\"COUNTRY\",\"PHONE\"])",
		},
		// Conversation MessagingAnalytics Example 3
		{
			name: "Conversation MessagingAnalytics Example 3",
			generate: func() *Request {
				return MakeConversationalAnalyticsQueryParams(
					1685527200,
					1685613600,
					GranularityHalfHour,
					WithConversationalCategories(
						ConversationalCategoryMarketing,
						ConversationalCategoryAuthentications,
					),
					WithConversationalDimensions(
						DimensionConversationCategory,
						DimensionConversationType,
					),
				)
			},
			expected: "conversation_analytics.start(1685527200).end(1685613600).granularity(HALF_HOUR).phone_numbers([]).dimensions([\"CONVERSATION_CATEGORY\",\"CONVERSATION_TYPE\"]).conversation_categories([\"MARKETING\",\"AUTHENTICATION\"])",
		},
		// Pricing MessagingAnalytics Example
		{
			name: "Pricing MessagingAnalytics Example",
			generate: func() *Request {
				return MakePricingAnalyticsQueryParams(
					1685602800,
					1688194800,
					GranularityMonthly,
					WithPricingPhoneNumbers("19195552584", "19195552585"),
					WithPricingMetricTypes(MetricTypeCost),
					WithPricingTypes(PricingTypeRegular),
					WithPricingCategories(PricingCategoryMarketing),
					WithPricingDimensions(
						DimensionCountry,
						DimensionPhone,
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
