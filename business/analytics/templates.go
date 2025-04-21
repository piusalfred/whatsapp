package analytics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	TemplateCostMetric struct {
		Type  string  `json:"type,omitempty"`
		Value float64 `json:"value,omitempty"`
	}

	TemplateClicked struct {
		Type          string `json:"type,omitempty"`
		ButtonContent string `json:"button_content,omitempty"`
		Count         int64  `json:"count,omitempty"`
	}

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

	TemplateAnalyticsData struct {
		Granularity string                   `json:"granularity,omitempty"`
		ProductType string                   `json:"product_type,omitempty"`
		DataPoints  []TemplateAnalyticsPoint `json:"data_points,omitempty"`
	}

	TemplateAnalyticsResponse struct {
		Data   []TemplateAnalyticsData `json:"data,omitempty"`
		Paging *whttp.Paging           `json:"paging,omitempty"`
	}

	TemplatesClient struct {
		reader config.Reader
		sender whttp.AnySender
	}
)

// NewTemplateAnalyticsClient returns a new instance of the TemplatesClient.
func NewTemplateAnalyticsClient(reader config.Reader, sender whttp.AnySender) *TemplatesClient {
	return &TemplatesClient{reader: reader, sender: sender}
}

type DisableButtonClickTrackingRequest struct {
	TemplateID string `json:"template_id"`
	OptOut     bool   `json:"cta_url_link_tracking_opted_out"`
	Category   string `json:"category"`
}

type DisableButtonClickTrackingResponse struct {
	Success bool `json:"success"`
}

func (c *TemplatesClient) DisableButtonClickTracking(ctx context.Context,
	req *DisableButtonClickTrackingRequest,
) (*DisableButtonClickTrackingResponse, error) {
	conf, err := c.reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	queryParams := map[string]string{
		"cta_url_link_tracking_opted_out": strconv.FormatBool(req.OptOut),
		"category":                        req.Category,
	}

	options := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeDisableButtonClickTracking),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestEndpoints[any](conf.APIVersion, req.TemplateID),
		whttp.WithRequestQueryParams[any](queryParams),
	}

	request := whttp.MakeRequest[any](http.MethodPost, conf.BaseURL, options...)

	response := &DisableButtonClickTrackingResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

type EnableTemplateAnalyticsResponse struct {
	ID string `json:"id"`
}

// Enable confirms template analytics on your WhatsApp Business Account. Once confirmed,
// template analytics cannot be disabled.
func (c *TemplatesClient) Enable(ctx context.Context) (string, error) {
	conf, err := c.reader.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}

	queryParams := map[string]string{
		"is_enabled_for_insights": "true",
	}

	options := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeEnableTemplatesAnalytics),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestEndpoints[any](conf.APIVersion, conf.BusinessAccountID),
		whttp.WithRequestQueryParams[any](queryParams),
	}

	req := whttp.MakeRequest[any](http.MethodPost, conf.BaseURL, options...)

	response := &EnableTemplateAnalyticsResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.sender.Send(ctx, req, decoder); err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}

	return response.ID, nil
}

type TemplateAnalyticsRequest struct {
	Start       int64    `json:"start"`
	End         int64    `json:"end"`
	Templates   []string `json:"template_ids"`
	MetricTypes []string `json:"metric_types,omitempty"`
}

const ErrInvalidTemplatesCount = whatsapp.Error("invalid number of templates")

// Fetch fetches template analytics for the specified templates within the specified date range.
func (c *TemplatesClient) Fetch(ctx context.Context, params *TemplateAnalyticsRequest) (
	*TemplateAnalyticsResponse, error,
) {
	queryParams := map[string]string{}
	queryParams["start"] = strconv.FormatInt(params.Start, 10)
	queryParams["end"] = strconv.FormatInt(params.End, 10)
	queryParams["granularity"] = GranularityDaily.String()
	if len(params.Templates) > 0 && len(params.Templates) <= 10 {
		queryParams["template_ids"] = "[" + strings.Join(params.Templates, ",") + "]"
	} else {
		return nil, fmt.Errorf("%w: count: %d: shold be >0 and < 11", ErrInvalidTemplatesCount, len(params.Templates))
	}

	if len(params.MetricTypes) > 0 {
		queryParams["metric_types"] = strings.Join(params.MetricTypes, ",")
	}

	conf, err := c.reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	options := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeFetchTemplateAnalytics),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestEndpoints[any](conf.APIVersion, conf.BusinessAccountID, string(TypeTemplateAnalytics)),
		whttp.WithRequestQueryParams[any](queryParams),
	}

	req := whttp.MakeRequest[any](http.MethodGet, conf.BaseURL,
		options...,
	)

	response := &TemplateAnalyticsResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

type TemplatesAnalytics interface {
	DisableButtonClickTracking(ctx context.Context,
		req *DisableButtonClickTrackingRequest,
	) (*DisableButtonClickTrackingResponse, error)
	Enable(ctx context.Context) (string, error)
	Fetch(ctx context.Context, params *TemplateAnalyticsRequest) (
		*TemplateAnalyticsResponse, error,
	)
}
