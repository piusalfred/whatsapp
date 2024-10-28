package analytics

type (
	// TemplateAnalyticsRequest represents the request to get template analytics.
	TemplateAnalyticsRequest struct {
		Fields      string   `json:"fields"` // Always 'template_analytics'
		Start       int64    `json:"start"`
		End         int64    `json:"end"`
		Granularity string   `json:"granularity"`  // Must be 'DAILY'
		TemplateIDs []string `json:"template_ids"` // Max 10
		MetricTypes []string `json:"metric_types,omitempty"`
	}

	// TemplateCostMetric represents a cost metric in the template analytics response.
	TemplateCostMetric struct {
		Type  string  `json:"type,omitempty"`
		Value float64 `json:"value,omitempty"`
	}

	// TemplateClicked represents a clicked metric in the template analytics response.
	TemplateClicked struct {
		Type          string `json:"type,omitempty"`
		ButtonContent string `json:"button_content,omitempty"`
		Count         int64  `json:"count,omitempty"`
	}

	// TemplateAnalyticsPoint represents a data point in the template analytics response.
	TemplateAnalyticsPoint struct {
		TemplateID string               `json:"template_id,omitempty"`
		Start      int64                `json:"start,omitempty"`
		End        int64                `json:"end,omitempty"`
		Sent       int64                `json:"sent,omitempty"`
		Delivered  int64                `json:"delivered,omitempty"`
		Read       int64                `json:"read,omitempty"`
		Clicked    []TemplateClicked    `json:"clicked,omitempty"`
		Cost       []TemplateCostMetric `json:"cost,omitempty"`
	}

	// TemplateAnalyticsData represents the data in the template analytics response.
	TemplateAnalyticsData struct {
		Granularity string                   `json:"granularity,omitempty"`
		ProductType string                   `json:"product_type,omitempty"`
		DataPoints  []TemplateAnalyticsPoint `json:"data_points,omitempty"`
	}

	// Cursor represents pagination cursors.
	Cursor struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	}

	// Paging represents pagination information.
	Paging struct {
		Cursors Cursor `json:"cursors,omitempty"`
	}

	// TemplateAnalyticsResponse represents the response from the template analytics request.
	TemplateAnalyticsResponse struct {
		Data   []TemplateAnalyticsData `json:"data,omitempty"`
		Paging Paging                  `json:"paging,omitempty"`
	}
)
