package flow

import (
	"context"
	"fmt"
	"net/http"
	"time"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type metricEndpoint string

const (
	MetricEndpointRequestCount              metricEndpoint = "ENDPOINT_REQUEST_COUNT"
	MetricEndpointRequestError              metricEndpoint = "ENDPOINT_REQUEST_ERROR"
	MetricEndpointRequestErrorRate          metricEndpoint = "ENDPOINT_REQUEST_ERROR_RATE"
	MetricEndpointRequestLatencySecondsCeil metricEndpoint = "ENDPOINT_REQUEST_LATENCY_SECONDS_CEIL"
	MetricEndpointAvailability              metricEndpoint = "ENDPOINT_AVAILABILITY"
)

type granularity string

const (
	GranularityDay      granularity = "DAY"
	GranularityHour     granularity = "HOUR"
	GranularityLifetime granularity = "LIFETIME"
)

type (
	MetricsAPIResponse struct {
		ID     string  `json:"id"`
		Metric *Metric `json:"metric"`
	}

	Metric struct {
		Name        string       `json:"name"`
		Granularity string       `json:"granularity"`
		DataPoints  []*DataPoint `json:"data_points"`
	}

	DataPoint struct {
		Timestamp string        `json:"timestamp"`
		Data      []*MetricData `json:"data"`
	}

	MetricData struct {
		Key   string `json:"key"`
		Value int64  `json:"value"`
	}

	MetricsRequest struct {
		FlowID      string         `json:"flow_id"`         // Flow ID to get metrics for
		MetricName  metricEndpoint `json:"name"`            // Metric name (e.g., ENDPOINT_REQUEST_ERROR)
		Granularity granularity    `json:"granularity"`     // Time granularity (DAY, HOUR, LIFETIME)
		Since       time.Time      `json:"since,omitempty"` // Start of the time period (optional, YYYY-MM-DD)
		Until       time.Time      `json:"until,omitempty"` // End of the time period (optional, YYYY-MM-DD)
	}
)

func (client *BaseClient) GetFlowMetrics(ctx context.Context, request *MetricsRequest) (*MetricsAPIResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("get flow metrics failed: %w", err)
	}

	queryParams := map[string]string{
		"fields": fmt.Sprintf(
			"metric.name(%s).granularity(%s).since(%s).until(%s)",
			request.MetricName,
			request.Granularity,
			request.Since.Format(time.DateOnly),
			request.Until.Format(time.DateOnly),
		),
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeGetFlowMetrics),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestQueryParams[any](queryParams),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestEndpoints[any](conf.APIVersion, request.FlowID),
	}

	req := whttp.MakeRequest(http.MethodGet, conf.BaseURL, opts...)

	var resp MetricsAPIResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
	})

	if err = client.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("get flow metrics failed: %w", err)
	}

	return &resp, nil
}
