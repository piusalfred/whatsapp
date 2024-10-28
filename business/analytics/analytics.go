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

//go:generate mockgen -destination=../../mocks/business/analytics/mock_analytics.go -package=analytics -source=analytics.go

var _ Fetcher = (*BaseClient)(nil)

type (
	Type       string
	BaseClient struct {
		Sender      whttp.Sender[Request]
		Config      config.Reader
		Middlewares []Middleware
	}

	Fetcher interface {
		FetchAnalytics(ctx context.Context, request *Request) (*Response, error)
	}

	FetcherFunc func(ctx context.Context, request *Request) (*Response, error)

	Middleware func(Fetcher) Fetcher
)

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

func (b *BaseClient) FetchGeneralAnalytics(ctx context.Context,
	request *MessagingRequest,
) (*MessagingResponse, error) {
	req := MakeMessagingAnalyticsQueryParams(request.Start, request.End,
		request.Granularity, request.Options...)

	coreFetcher := FetcherFunc(b.FetchAnalytics)
	fetcher := wrapMiddleware(coreFetcher, b.Middlewares...)

	resp, err := fetcher.FetchAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch general analytics: %w", err)
	}

	return resp.MessagingAnalytics(), nil
}

func (b *BaseClient) FetchConversationAnalytics(ctx context.Context,
	request *ConversationalRequest,
) (*ConversationalResponse, error) {
	req := MakeConversationalAnalyticsQueryParams(request.Start, request.End, request.Granularity, request.Options...)

	coreFetcher := FetcherFunc(b.FetchAnalytics)
	fetcher := wrapMiddleware(coreFetcher, b.Middlewares...)

	resp, err := fetcher.FetchAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch conversation analytics: %w", err)
	}

	return resp.ConversationAnalytics(), nil
}

func (b *BaseClient) FetchPricingAnalytics(ctx context.Context, params *PricingRequest) (
	*PricingResponse, error,
) {
	coreFetcher := FetcherFunc(b.FetchAnalytics)
	fetcher := wrapMiddleware(coreFetcher, b.Middlewares...)

	request := MakePricingAnalyticsQueryParams(params.Start, params.End, params.Granularity, params.Options...)
	resp, err := fetcher.FetchAnalytics(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("fetch pricing analytics: %w", err)
	}

	return resp.PricingAnalytics(), nil
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

func (b *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*Response, error) {
	queryParams := map[string]string{
		"fields":       request.QueryParamsString(),
		"access_token": conf.AccessToken,
	}

	req := whttp.MakeRequest[Request](http.MethodGet, conf.BaseURL,
		whttp.WithRequestType[Request](request.requestType),
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

type Response struct {
	Messaging      *MessagingAnalytics    `json:"analytics,omitempty"`
	Conversational *ConversationAnalytics `json:"conversation_analytics,omitempty"`
	Pricing        *PricingAnalytics      `json:"pricing_analytics,omitempty"`
	ID             string                 `json:"id,omitempty"`
}

func (response *Response) MessagingAnalytics() *MessagingResponse {
	return &MessagingResponse{
		Analytics: response.Messaging,
		ID:        response.ID,
	}
}

func (response *Response) ConversationAnalytics() *ConversationalResponse {
	return &ConversationalResponse{
		ConversationAnalytics: response.Conversational,
		ID:                    response.ID,
	}
}

func (response *Response) PricingAnalytics() *PricingResponse {
	return &PricingResponse{
		PricingAnalytics: response.Pricing,
		ID:               response.ID,
	}
}

type (
	MessagingRequest struct {
		Start       int64
		End         int64
		Granularity Granularity
		Options     []MessagingQueryParamsOption
	}

	MessagingResponse struct {
		Analytics *MessagingAnalytics `json:"analytics,omitempty"`
		ID        string              `json:"id,omitempty"`
	}

	MessagingAnalytics struct {
		PhoneNumbers []string     `json:"phone_numbers,omitempty"`
		CountryCodes []string     `json:"country_codes,omitempty"`
		Granularity  string       `json:"granularity,omitempty"`
		DataPoints   []*DataPoint `json:"data_points,omitempty"`
	}

	MessagingQueryParams struct {
		PhoneNumbers []string
		ProductTypes []int64
		CountryCodes []string
	}

	MessagingQueryParamsOption func(*MessagingQueryParams)
)

func WithMessagingPhoneNumbers(phoneNumbers ...string) MessagingQueryParamsOption {
	return func(p *MessagingQueryParams) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithMessagingProductTypes(productTypes ...int64) MessagingQueryParamsOption {
	return func(p *MessagingQueryParams) {
		p.ProductTypes = productTypes
	}
}

func WithMessagingCountryCodes(countryCodes ...string) MessagingQueryParamsOption {
	return func(p *MessagingQueryParams) {
		p.CountryCodes = countryCodes
	}
}

func MakeMessagingAnalyticsQueryParams(start, end int64, granularity Granularity,
	options ...MessagingQueryParamsOption,
) *Request {
	params := &MessagingQueryParams{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		requestType:  whttp.RequestTypeFetchMessagingAnalytics,
		Fields:       TypeMessagingAnalytics,
		Start:        start,
		End:          end,
		Granularity:  granularity,
		PhoneNumbers: params.PhoneNumbers,
		ProductTypes: params.ProductTypes,
		CountryCodes: params.CountryCodes,
	}
}

type (
	ConversationAnalytics struct {
		Data []*Data `json:"data,omitempty"`
	}

	ConversationalResponse struct {
		ConversationAnalytics *ConversationAnalytics `json:"conversation_analytics,omitempty"`
		ID                    string                 `json:"id,omitempty"`
	}

	ConversationalRequest struct {
		Start       int64
		End         int64
		Granularity Granularity
		Options     []ConversationalQueryParamsOption
	}

	ConversationalQueryParams struct {
		PhoneNumbers           []string
		MetricTypes            []MetricType
		ConversationCategories []ConversationalCategory
		ConversationTypes      []ConversationalType
		ConversationDirections []ConversationalDirection
		Dimensions             []Dimension
	}

	ConversationalQueryParamsOption func(*ConversationalQueryParams)
)

func WithConversationalPhoneNumbers(phoneNumbers ...string) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithConversationalMetricTypes(metricTypes ...MetricType) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.MetricTypes = metricTypes
	}
}

func WithConversationalCategories(conversationCategories ...ConversationalCategory) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.ConversationCategories = conversationCategories
	}
}

func WithConversationalTypes(conversationTypes ...ConversationalType) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.ConversationTypes = conversationTypes
	}
}

func WithConversationalDirections(conversationDirections ...ConversationalDirection) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.ConversationDirections = conversationDirections
	}
}

func WithConversationalDimensions(dimensions ...Dimension) ConversationalQueryParamsOption {
	return func(p *ConversationalQueryParams) {
		p.Dimensions = dimensions
	}
}

func MakeConversationalAnalyticsQueryParams(start, end int64, granularity Granularity,
	options ...ConversationalQueryParamsOption,
) *Request {
	params := &ConversationalQueryParams{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		requestType:            whttp.RequestTypeFetchConversationAnalytics,
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

type (
	PricingResponse struct {
		PricingAnalytics *PricingAnalytics `json:"pricing_analytics,omitempty"`
		ID               string            `json:"id,omitempty"`
	}

	PricingRequest struct {
		Start       int64
		End         int64
		Granularity Granularity
		Options     []PricingQueryParamsOption
	}

	PricingAnalytics struct {
		Data []*Data `json:"data,omitempty"`
	}

	PricingQueryParams struct {
		PhoneNumbers      []string
		CountryCodes      []string
		MetricTypes       []MetricType
		PricingTypes      []PricingType
		PricingCategories []PricingCategory
		Dimensions        []Dimension
	}

	PricingQueryParamsOption func(*PricingQueryParams)
)

func WithPricingPhoneNumbers(phoneNumbers ...string) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.PhoneNumbers = phoneNumbers
	}
}

func WithPricingCountryCodes(countryCodes ...string) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.CountryCodes = countryCodes
	}
}

func WithPricingMetricTypes(metricTypes ...MetricType) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.MetricTypes = metricTypes
	}
}

func WithPricingTypes(pricingTypes ...PricingType) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.PricingTypes = pricingTypes
	}
}

func WithPricingCategories(pricingCategories ...PricingCategory) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.PricingCategories = pricingCategories
	}
}

func WithPricingDimensions(dimensions ...Dimension) PricingQueryParamsOption {
	return func(p *PricingQueryParams) {
		p.Dimensions = dimensions
	}
}

func MakePricingAnalyticsQueryParams(start, end int64, granularity Granularity,
	options ...PricingQueryParamsOption,
) *Request {
	params := &PricingQueryParams{}

	for _, option := range options {
		option(params)
	}

	return &Request{
		requestType:       whttp.RequestTypeFetchPricingAnalytics,
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

type Request struct {
	requestType            whttp.RequestType
	Fields                 Type
	Start                  int64
	End                    int64
	Granularity            Granularity
	PhoneNumbers           []string
	ProductTypes           []int64
	CountryCodes           []string
	MetricTypes            []MetricType
	ConversationCategories []ConversationalCategory
	ConversationTypes      []ConversationalType
	ConversationDirections []ConversationalDirection
	Dimensions             []Dimension
	PricingCategories      []PricingCategory
	PricingTypes           []PricingType
}

func (r *Request) QueryParamsString() string {
	var buffer strings.Builder
	buffer.WriteString(string(r.Fields))

	appendParamValue(&buffer, "start", r.Start, formatInt)
	appendParamValue(&buffer, "end", r.End, formatInt)
	appendParamValue(&buffer, "granularity", r.Granularity, noQuotesString)

	appendParamValue(&buffer, "phone_numbers", r.PhoneNumbers, func(phoneNumbers []string) string {
		return formatArray(phoneNumbers, quoteString)
	})

	if len(r.CountryCodes) > 0 {
		appendParamValue(&buffer, "country_codes", r.CountryCodes, func(countryCodes []string) string {
			return formatArray(countryCodes, quoteString)
		})
	}

	if len(r.Dimensions) > 0 {
		appendParamValue(&buffer, "dimensions", r.Dimensions, func(dimensions []Dimension) string {
			return formatArray(dimensions, quoteString)
		})
	}

	switch r.Fields { //nolint:exhaustive
	case TypeMessagingAnalytics:
		if len(r.ProductTypes) > 0 {
			appendParamValue(&buffer, "product_types", r.ProductTypes, func(productTypes []int64) string {
				return formatArray(productTypes, formatInt)
			})
		}
	case TypeConversationAnalytics:
		if len(r.MetricTypes) > 0 {
			appendParamValue(&buffer, "metric_types", r.MetricTypes,
				func(metricTypes []MetricType) string {
					return formatArray(metricTypes, quoteString)
				})
		}
		if len(r.ConversationCategories) > 0 {
			appendParamValue(&buffer, "conversation_categories", r.ConversationCategories,
				func(conversationCategories []ConversationalCategory) string {
					return formatArray(conversationCategories, quoteString)
				})
		}
		if len(r.ConversationTypes) > 0 {
			appendParamValue(&buffer, "conversation_types", r.ConversationTypes,
				func(conversationTypes []ConversationalType) string {
					return formatArray(conversationTypes, quoteString)
				})
		}
		if len(r.ConversationDirections) > 0 {
			appendParamValue(&buffer, "conversation_directions", r.ConversationDirections,
				func(conversationDirections []ConversationalDirection) string {
					return formatArray(conversationDirections, quoteString)
				})
		}

	case TypePricingAnalytics:
		if len(r.MetricTypes) > 0 {
			appendParamValue(&buffer, "metric_types", r.MetricTypes,
				func(metricTypes []MetricType) string {
					return formatArray(metricTypes, quoteString)
				})
		}
		if len(r.PricingTypes) > 0 {
			appendParamValue(&buffer, "pricing_types", r.PricingTypes,
				func(pricingTypes []PricingType) string {
					return formatArray(pricingTypes, quoteString)
				})
		}

		if len(r.PricingCategories) > 0 {
			appendParamValue(&buffer, "pricing_categories", r.PricingCategories,
				func(pricingCategories []PricingCategory) string {
					return formatArray(pricingCategories, quoteString)
				})
		}
	}

	return buffer.String()
}

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

func quoteString[T ~string](s T) string {
	return fmt.Sprintf("\"%s\"", s)
}

func formatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

func noQuotesString[T ~string](s T) string {
	return string(s)
}

func appendParamValue[T any](buffer *strings.Builder, key string, value T, formatter func(T) string) string {
	buffer.WriteString(".")
	buffer.WriteString(key)
	buffer.WriteString("(")
	buffer.WriteString(formatter(value))
	buffer.WriteString(")")

	return buffer.String()
}

type (
	Data struct {
		DataPoints []*DataPoint `json:"data_points,omitempty"`
	}

	DataPoint struct {
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
		Volume                int64   `json:"volume,omitempty"`
		PricingType           string  `json:"pricing_type,omitempty"`
		PricingCategory       string  `json:"pricing_category,omitempty"`
	}
)

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

// String returns the string representation of the granularity.
func (g Granularity) String() string {
	return string(g)
}

const (
	TypeMessagingAnalytics    Type = "analytics"
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
	MetricTypeDelivered    MetricType = "DELIVERED"
	MetricTypeRead         MetricType = "READ"
	MetricTypeSent         MetricType = "SENT"
	MetricTypeClicked      MetricType = "CLICKED"
)

type ProductType int64

const (
	ProductTypeNotificationMessage ProductType = 0
	ProductTypeCustomerSupport     ProductType = 2
)
