package analytics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	// ConversationAnalyticsEntry represents an entry in the conversation analytics data.
	ConversationAnalyticsEntry struct {
		DataPoints []*DataPoint `json:"data_points,omitempty"`
	}

	// ConversationData represents the data in the conversation analytics response.
	ConversationData struct {
		Data []*ConversationAnalyticsEntry `json:"data,omitempty"`
	}

	// ConversationAnalyticsResponse represents the response from the conversation analytics request.
	ConversationAnalyticsResponse struct {
		ConversationAnalytics *ConversationData `json:"conversation_analytics,omitempty"`
		ID                    string            `json:"id,omitempty"`
	}

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

	// PricingAnalyticsRequest represents the request to get pricing analytics.
	PricingAnalyticsRequest struct {
		Fields            string   `json:"fields"` // Always 'pricing_analytics'
		Start             int64    `json:"start"`
		End               int64    `json:"end"`
		Granularity       string   `json:"granularity"`
		PhoneNumbers      []string `json:"phone_numbers,omitempty"`
		CountryCodes      []string `json:"country_codes,omitempty"`
		MetricTypes       []string `json:"metric_types,omitempty"`
		PricingTypes      []string `json:"pricing_types,omitempty"`
		PricingCategories []string `json:"pricing_categories,omitempty"`
		Dimensions        []string `json:"dimensions,omitempty"`
	}

	// PricingDataPoint represents a data point in the pricing analytics response.
	PricingDataPoint struct {
		Start           int64   `json:"start,omitempty"`
		End             int64   `json:"end,omitempty"`
		Cost            float64 `json:"cost,omitempty"`
		Volume          int64   `json:"volume,omitempty"`
		PhoneNumber     string  `json:"phone_number,omitempty"`
		Country         string  `json:"country,omitempty"`
		PricingType     string  `json:"pricing_type,omitempty"`
		PricingCategory string  `json:"pricing_category,omitempty"`
	}

	// PricingAnalyticsEntry represents an entry in the pricing analytics data.
	PricingAnalyticsEntry struct {
		DataPoints []PricingDataPoint `json:"data_points,omitempty"`
	}

	// PricingAnalyticsData represents the data in the pricing analytics response.
	PricingAnalyticsData struct {
		Data []PricingAnalyticsEntry `json:"data,omitempty"`
	}

	// PricingAnalyticsResponse represents the response from the pricing analytics request.
	PricingAnalyticsResponse struct {
		PricingAnalytics PricingAnalyticsData `json:"pricing_analytics,omitempty"`
		ID               string               `json:"id,omitempty"`
	}
)

var _ Fetcher = (*BaseClient)(nil)

type Granularity string

const (
	GranularityDay      Granularity = "DAY"
	GranularityHour     Granularity = "HOUR"
	GranularityHalfHour Granularity = "HALF_HOUR"
	GranularityDaily    Granularity = "DAILY"
	GranularityMonthly  Granularity = "MONTHLY"
	GranularityLifetime Granularity = "LIFETIME"
	GranularityMonth    Granularity = "MONTH"
)

const (
	TypeGeneralAnalytics      Type = "analytics"
	TypeConversationAnalytics Type = "conversation_analytics"
	TypeTemplateAnalytics     Type = "template_analytics"
	TypePricingAnalytics      Type = "pricing_analytics"
)

type PricingCategory string

const (
	PricingCategoryAuthentication              PricingCategory = "AUTHENTICATION"
	PricingCategoryMarketing                   PricingCategory = "MARKETING"
	PricingCategoryService                     PricingCategory = "SERVICE"
	PricingCategoryUtility                     PricingCategory = "UTILITY"
	PricingCategoryAuthenticationInternational PricingCategory = "AUTHENTICATION_INTERNATIONAL"
)

type PricingType string

const (
	PricingTypeFreeCustomerService PricingType = "FREE_CUSTOMER_SERVICE"
	PricingTypeFreeEntryPoint      PricingType = "FREE_ENTRY_POINT"
	PricingTypeRegular             PricingType = "REGULAR"
)

type (
	Type       string
	BaseClient struct {
		Sender      whttp.Sender[Request]
		Config      config.Reader
		Middlewares []Middleware
	}

	Response struct {
		Analytics          *Analytics        `json:"analytics,omitempty"`
		ConversationalData *ConversationData `json:"conversation_analytics,omitempty"`
		ID                 string            `json:"id,omitempty"`
	}

	Request struct {
		Fields                 Type
		Start                  int64
		End                    int64
		Granularity            Granularity
		PhoneNumbers           []string
		ProductTypes           []int // 0 for notification messages, 2 for customer support messages
		CountryCodes           []string
		MetricTypes            []MetricType
		ConversationCategories []ConversationalCategory
		ConversationTypes      []ConversationalType
		ConversationDirections []ConversationalDirection
		Dimensions             []Dimension
		PricingCategories      []PricingCategory
		PricingTypes           []PricingType
	}

	Fetcher interface {
		FetchAnalytics(ctx context.Context, request *Request) (*Response, error)
	}

	FetcherFunc func(ctx context.Context, request *Request) (*Response, error)

	Middleware func(Fetcher) Fetcher
)

func (response *Response) GeneralAnalytics() *GeneralResponse {
	return &GeneralResponse{
		Analytics: response.Analytics,
		ID:        response.ID,
	}
}

func (response *Response) ConversationAnalytics() *ConversationAnalyticsResponse {
	return &ConversationAnalyticsResponse{
		ConversationAnalytics: response.ConversationalData,
		ID:                    response.ID,
	}
}

func (f FetcherFunc) FetchAnalytics(ctx context.Context, request *Request) (*Response, error) {
	return f(ctx, request)
}

//nolint:ireturn
func wrapMiddleware(fetcher Fetcher, middlewares ...Middleware) Fetcher {
	for i := len(middlewares) - 1; i >= 0; i-- {
		fetcher = middlewares[i](fetcher)
	}

	return fetcher
}

type GeneralAnalyticsRequest struct {
	Start       int64
	End         int64
	Granularity Granularity
	Options     []ParamsOption
}

type ConversationAnalyticsRequest struct {
	Start       int64
	End         int64
	Granularity Granularity
	Options     []ConversationalParamsOption
}

type GeneralResponse struct {
	Analytics *Analytics `json:"analytics,omitempty"`
	ID        string     `json:"id,omitempty"`
}

type Analytics struct {
	PhoneNumbers []string     `json:"phone_numbers,omitempty"`
	CountryCodes []string     `json:"country_codes,omitempty"`
	Granularity  string       `json:"granularity,omitempty"`
	DataPoints   []*DataPoint `json:"data_points,omitempty"`
}

type DataPoint struct {
	Start                 int64   `json:"start,omitempty"`
	End                   int64   `json:"end,omitempty"`
	Sent                  int64   `json:"sent,omitempty"`
	Delivered             int64   `json:"delivered,omitempty"`
	Conversation          int64   `json:"conversation,omitempty"`
	PhoneNumber           string  `json:"phone_number,omitempty"`
	Country               string  `json:"country,omitempty"`
	ConversationType      string  `json:"conversation_type,omitempty"`
	ConversationDirection string  `json:"conversation_direction,omitempty"`
	ConversationCategory  string  `json:"conversation_category,omitempty"`
	Cost                  float64 `json:"cost,omitempty"`
}

// FetchGeneralAnalytics fetches the general analytics data.
func (b *BaseClient) FetchGeneralAnalytics(ctx context.Context,
	request *GeneralAnalyticsRequest,
) (*GeneralResponse, error) {
	req := NewAnalyticsParameters(request.Start, request.End,
		request.Granularity, request.Options...)

	coreFetcher := FetcherFunc(b.FetchAnalytics)
	fetcher := wrapMiddleware(coreFetcher, b.Middlewares...)

	resp, err := fetcher.FetchAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch general analytics: %w", err)
	}

	return resp.GeneralAnalytics(), nil
}

// FetchConversationAnalytics fetches the conversation analytics data.
func (b *BaseClient) FetchConversationAnalytics(ctx context.Context,
	request *ConversationAnalyticsRequest,
) (*ConversationAnalyticsResponse, error) {
	req := NewConversationalParameters(request.Start, request.End, request.Granularity, request.Options...)

	coreFetcher := FetcherFunc(b.FetchAnalytics)
	fetcher := wrapMiddleware(coreFetcher, b.Middlewares...)

	resp, err := fetcher.FetchAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}

	return resp.ConversationAnalytics(), nil
}

func (b *BaseClient) FetchAnalytics(ctx context.Context, request *Request) (*Response, error) {
	conf, err := b.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	resp, err := b.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("fetch analytics: %w", err)
	}

	return resp, nil
}

// Send sends the analytics request to the WhatsApp Business API.
func (b *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*Response, error) {
	queryParams := map[string]string{
		"fields":       request.QueryParamsString(),
		"access_token": conf.AccessToken,
	}

	req := whttp.MakeRequest[Request](http.MethodGet, conf.BaseURL,
		whttp.WithRequestAppSecret[Request](conf.AppSecret),
		whttp.WithRequestSecured[Request](conf.SecureRequests),
		whttp.WithRequestQueryParams[Request](queryParams),
		whttp.WithRequestEndpoints[Request](conf.APIVersion, conf.BusinessAccountID),
	)

	response := &Response{}
	err := b.Sender.Send(ctx, req, whttp.ResponseDecoderJSON[Response](response, whttp.DecodeOptions{
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	}))
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

type Params struct {
	PhoneNumbers []string
	ProductTypes []int
	CountryCodes []string
}

type ParamsOption func(*Params)

func WithPhoneNumbers(phoneNumbers ...string) ParamsOption {
	return func(p *Params) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithProductTypes(productTypes ...int) ParamsOption {
	return func(p *Params) {
		p.ProductTypes = productTypes
	}
}

func WithCountryCodes(countryCodes ...string) ParamsOption {
	return func(p *Params) {
		p.CountryCodes = countryCodes
	}
}

func NewAnalyticsParameters(start, end int64, granularity Granularity, options ...ParamsOption) *Request {
	params := &Params{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		Fields:       TypeGeneralAnalytics,
		Start:        start,
		End:          end,
		Granularity:  granularity,
		PhoneNumbers: params.PhoneNumbers,
		ProductTypes: params.ProductTypes,
		CountryCodes: params.CountryCodes,
	}
}

type ConversationalCategory string

const (
	ConversationalCategoryAuthentications ConversationalCategory = "AUTHENTICATION"
	ConversationalCategoryMarketing       ConversationalCategory = "MARKETING"
	ConversationalCategoryService         ConversationalCategory = "SERVICE"
	ConversationalCategoryUtility         ConversationalCategory = "UTILITY"
)

type ConversationalType string

const (
	ConversationalTypeFreeEntry ConversationalType = "FREE_ENTRY"
	ConversationalTypeFreeTier  ConversationalType = "FREE_TIER"
	ConversationalTypeRegular   ConversationalType = "REGULAR"
)

type ConversationalDirection string

const (
	ConversationalDirectionBusinessInitiated ConversationalDirection = "BUSINESS_INITIATED"
	ConversationalDirectionUserInitiated     ConversationalDirection = "USER_INITIATED"
)

type Dimension string

const (
	DimensionConversationCategory  Dimension = "CONVERSATION_CATEGORY"
	DimensionConversationDirection Dimension = "CONVERSATION_DIRECTION"
	DimensionConversationType      Dimension = "CONVERSATION_TYPE"
	DimensionCountry               Dimension = "COUNTRY"
	DimensionPhone                 Dimension = "PHONE"
)

type MetricType string

const (
	MetricTypeCost         MetricType = "COST"
	MetricTypeConversation MetricType = "CONVERSATION"
)

type ConversationalParams struct {
	PhoneNumbers           []string
	MetricTypes            []MetricType
	ConversationCategories []ConversationalCategory
	ConversationTypes      []ConversationalType
	ConversationDirections []ConversationalDirection
	Dimensions             []Dimension
}

type ConversationalParamsOption func(*ConversationalParams)

func WithConversationalPhoneNumbers(phoneNumbers ...string) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithConversationalMetricTypes(metricTypes ...MetricType) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.MetricTypes = metricTypes
	}
}

func WithConversationalCategories(conversationCategories ...ConversationalCategory) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.ConversationCategories = conversationCategories
	}
}

func WithConversationalTypes(conversationTypes ...ConversationalType) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.ConversationTypes = conversationTypes
	}
}

func WithConversationalDirections(conversationDirections ...ConversationalDirection) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.ConversationDirections = conversationDirections
	}
}

func WithConversationalDimensions(dimensions ...Dimension) ConversationalParamsOption {
	return func(p *ConversationalParams) {
		p.Dimensions = dimensions
	}
}

func NewConversationalParameters(start, end int64, granularity Granularity,
	options ...ConversationalParamsOption,
) *Request {
	params := &ConversationalParams{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		Fields:                 TypeConversationAnalytics,
		Start:                  start,
		End:                    end,
		Granularity:            granularity,
		PhoneNumbers:           params.PhoneNumbers,
		MetricTypes:            params.MetricTypes,
		ConversationCategories: params.ConversationCategories,
		ConversationTypes:      params.ConversationTypes,
		ConversationDirections: params.ConversationDirections,
		Dimensions:             params.Dimensions,
	}
}

type PricingAnalyticsParams struct {
	PhoneNumbers      []string
	CountryCodes      []string
	MetricTypes       []MetricType
	PricingTypes      []PricingType
	PricingCategories []PricingCategory
	Dimensions        []Dimension
}

type PricingAnalyticsParamsOption func(*PricingAnalyticsParams)

func WithPricingPhoneNumbers(phoneNumbers ...string) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithPricingCountryCodes(countryCodes ...string) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.CountryCodes = countryCodes
	}
}

func WithPricingMetricTypes(metricTypes ...MetricType) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.MetricTypes = metricTypes
	}
}

func WithPricingTypes(pricingTypes ...PricingType) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.PricingTypes = pricingTypes
	}
}

func WithPricingCategories(pricingCategories ...PricingCategory) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.PricingCategories = pricingCategories
	}
}

func WithPricingDimensions(dimensions ...Dimension) PricingAnalyticsParamsOption {
	return func(p *PricingAnalyticsParams) {
		p.Dimensions = dimensions
	}
}

func NewPricingParameters(start, end int64, granularity Granularity, options ...PricingAnalyticsParamsOption) *Request {
	params := &PricingAnalyticsParams{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		Fields:            TypePricingAnalytics,
		Start:             start,
		End:               end,
		Granularity:       granularity,
		PhoneNumbers:      params.PhoneNumbers,
		CountryCodes:      params.CountryCodes,
		MetricTypes:       params.MetricTypes,
		PricingTypes:      params.PricingTypes,
		PricingCategories: params.PricingCategories,
		Dimensions:        params.Dimensions,
	}
}

func (r *Request) QueryParamsString() string {
	var buffer strings.Builder
	buffer.WriteString(string(r.Fields))

	AppendParamValue(&buffer, "start", r.Start, formatInt64)
	AppendParamValue(&buffer, "end", r.End, formatInt64)
	AppendParamValue(&buffer, "granularity", r.Granularity, noQuotesStringer)

	AppendParamValue(&buffer, "phone_numbers", r.PhoneNumbers, func(phoneNumbers []string) string {
		return formatArray(phoneNumbers, quoteString)
	})

	if len(r.CountryCodes) > 0 {
		AppendParamValue(&buffer, "country_codes", r.CountryCodes, func(countryCodes []string) string {
			return formatArray(countryCodes, quoteString)
		})
	}

	// dimensions
	if len(r.Dimensions) > 0 {
		AppendParamValue(&buffer, "dimensions", r.Dimensions, func(dimensions []Dimension) string {
			return formatArray(dimensions, quoteString)
		})
	}

	// Fields specific to analytics type
	switch r.Fields { //nolint:exhaustive
	case TypeGeneralAnalytics:
		if len(r.ProductTypes) > 0 {
			AppendParamValue(&buffer, "product_types", r.ProductTypes, func(productTypes []int) string {
				return formatArray(productTypes, formatInt)
			})
		}
	case TypeConversationAnalytics:
		if len(r.MetricTypes) > 0 {
			AppendParamValue(&buffer, "metric_types", r.MetricTypes,
				func(metricTypes []MetricType) string {
					return formatArray(metricTypes, quoteString)
				})
		}
		if len(r.ConversationCategories) > 0 {
			AppendParamValue(&buffer, "conversation_categories", r.ConversationCategories,
				func(conversationCategories []ConversationalCategory) string {
					return formatArray(conversationCategories, quoteString)
				})
		}
		if len(r.ConversationTypes) > 0 {
			AppendParamValue(&buffer, "conversation_types", r.ConversationTypes,
				func(conversationTypes []ConversationalType) string {
					return formatArray(conversationTypes, quoteString)
				})
		}
		if len(r.ConversationDirections) > 0 {
			AppendParamValue(&buffer, "conversation_directions", r.ConversationDirections,
				func(conversationDirections []ConversationalDirection) string {
					return formatArray(conversationDirections, quoteString)
				})
		}

	case TypePricingAnalytics:
		if len(r.MetricTypes) > 0 {
			AppendParamValue(&buffer, "metric_types", r.MetricTypes,
				func(metricTypes []MetricType) string {
					return formatArray(metricTypes, quoteString)
				})
		}
		if len(r.PricingTypes) > 0 {
			AppendParamValue(&buffer, "pricing_types", r.PricingTypes,
				func(pricingTypes []PricingType) string {
					return formatArray(pricingTypes, quoteString)
				})
		}

		if len(r.PricingCategories) > 0 {
			AppendParamValue(&buffer, "pricing_categories", r.PricingCategories,
				func(pricingCategories []PricingCategory) string {
					return formatArray(pricingCategories, quoteString)
				})
		}
	}

	return buffer.String()
}

// formatArray formats an array of any type into a string representation suitable for query parameters.
// The formatter function determines how each element is formatted.
func formatArray[T any](arr []T, formatter func(T) string) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(formatter(v))
	}
	sb.WriteString("]")

	return sb.String()
}

// quoteString formats a string by adding double quotes around it.
func quoteString[T ~string](s T) string {
	return fmt.Sprintf("\"%s\"", s)
}

// formatInt formats an integer as a string.
func formatInt(n int) string {
	return strconv.Itoa(n)
}

// formatInt64 formats an int64 as a string.
func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

func noQuotesStringer[T ~string](s T) string {
	return string(s)
}

func AppendParamValue[T any](buffer *strings.Builder, key string, value T, formatter func(T) string) string {
	buffer.WriteString(".")
	buffer.WriteString(key)
	buffer.WriteString("(")
	buffer.WriteString(formatter(value))
	buffer.WriteString(")")

	return buffer.String()
}
