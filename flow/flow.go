/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package flow

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	CategorySignUp             = "SIGN_UP"
	CategorySignIn             = "SIGN_IN"
	CategoryAppointmentBooking = "APPOINTMENT_BOOKING"
	CategoryLeadGeneration     = "LEAD_GENERATION"
	CategoryContactUs          = "CONTACT_US"
	CategoryCustomerSupport    = "CUSTOMER_SUPPORT"
	CategorySurvey             = "SURVEY"
	CategoryOther              = "OTHER"
	StatusDraft                = "DRAFT"
	StatusPublished            = "PUBLISHED"
	StatusDeprecated           = "DEPRECATED"
	StatusBlocked              = "BLOCKED"
	StatusThrottled            = "THROTTLED"
	HealthStatusAvailable      = "AVAILABLE"
	HealthStatusLimited        = "LIMITED"
	HealthStatusBlocked        = "BLOCKED"
)

type (
	CreateRequest struct {
		Name        string   `json:"name,omitempty"`
		Categories  []string `json:"categories,omitempty"`
		CloneFlowID string   `json:"clone_flow_id,omitempty"`
		EndpointURI string   `json:"endpoint_uri,omitempty"`
	}

	CreateResponse struct {
		ID string `json:"id"`
	}

	UpdateRequest struct {
		Name          string   `json:"name,omitempty"`
		Categories    []string `json:"categories,omitempty"`
		EndpointURI   string   `json:"endpoint_uri,omitempty"`
		ApplicationID string   `json:"application_id,omitempty"`
	}

	UpdateResponse struct {
		Success bool `json:"success"`
	}

	UpdateFlowJSONRequest struct {
		Name string
		File string
	}

	UpdateFlowJSONResponse struct {
		Success          bool              `json:"success"`
		ValidationErrors []ValidationError `json:"validation_errors"`
	}

	ValidationError struct {
		Error       string `json:"error"`
		ErrorType   string `json:"error_type"`
		Message     string `json:"message"`
		LineStart   int    `json:"line_start"`
		LineEnd     int    `json:"line_end"`
		ColumnStart int    `json:"column_start"`
		ColumnEnd   int    `json:"column_end"`
	}

	Flow struct {
		ID               string            `json:"id"`
		Name             string            `json:"name"`
		Status           string            `json:"status"`
		Categories       []string          `json:"categories"`
		ValidationErrors []ValidationError `json:"validation_errors"`
	}

	ListResponse struct {
		Data   []*Flow       `json:"data"`
		Paging *whttp.Paging `json:"paging"`
	}

	Preview struct {
		PreviewURL string `json:"preview_url"`
		ExpiresAt  string `json:"expires_at"`
	}

	EntityError struct {
		ErrorCode        int    `json:"error_code"`
		ErrorDescription string `json:"error_description"`
		PossibleSolution string `json:"possible_solution"`
	}

	HealthStatusEntity struct {
		EntityType     string        `json:"entity_type"`
		ID             string        `json:"id"`
		CanSendMessage string        `json:"can_send_message"`
		Errors         []EntityError `json:"errors,omitempty"`
		AdditionalInfo []string      `json:"additional_info,omitempty"`
	}

	HealthStatus struct {
		CanSendMessage string               `json:"can_send_message"`
		Entities       []HealthStatusEntity `json:"entities"`
	}

	SingleFlowResponse struct {
		ID                      string                 `json:"id"`
		Name                    string                 `json:"name"`
		Status                  string                 `json:"status"`
		Categories              []string               `json:"categories"`
		ValidationErrors        []ValidationError      `json:"validation_errors"`
		JSONVersion             string                 `json:"json_version"`
		DataAPIVersion          string                 `json:"data_api_version"`
		EndpointURI             string                 `json:"endpoint_uri"`
		Preview                 Preview                `json:"preview"`
		WhatsAppBusinessAccount map[string]interface{} `json:"whatsapp_business_account,omitempty"` // Optional
		Application             map[string]interface{} `json:"application,omitempty"`               // Optional
		HealthStatus            HealthStatus           `json:"health_status"`
	}

	Asset struct {
		Name        string `json:"name"`
		AssetType   string `json:"asset_type"`
		DownloadURL string `json:"download_url"`
	}

	RetrieveAssetsResponse struct {
		Data   []*Asset      `json:"data"`
		Paging *whttp.Paging `json:"paging"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}
)

type (
	PreviewRequest struct {
		FlowID     string `json:"flow_id"`    // The ID of the flow being previewed.
		Invalidate bool   `json:"invalidate"` // If true, it invalidates the previous preview and generates a new one.
	}

	// PreviewResponse represents the response from the API when generating a preview URL for a Flow.
	PreviewResponse struct {
		PreviewURL string `json:"preview_url"` // The URL to visualize the Flow.
		ExpiresAt  string `json:"expires_at"`  // The expiration time of the preview URL.
		ID         string `json:"id"`          // The ID of the flow.
	}
	BaseRequest struct {
		Type        whttp.RequestType
		Method      string
		Endpoints   []string // follows after baseURL/version
		QueryParams map[string]string
		Payload     any
	}

	BaseResponse struct {
		ID                      string                 `json:"id,omitempty"`
		Success                 bool                   `json:"success,omitempty"`
		Data                    []*BaseResponseData    `json:"data,omitempty"`
		Paging                  *whttp.Paging          `json:"paging,omitempty"`
		ValidationErrors        []ValidationError      `json:"validation_errors,omitempty"`
		PreviewURL              string                 `json:"preview_url,omitempty"`
		ExpiresAt               string                 `json:"expires_at,omitempty"`
		Name                    string                 `json:"name,omitempty"`
		Status                  string                 `json:"status,omitempty"`
		Categories              []string               `json:"categories,omitempty"`
		JSONVersion             string                 `json:"json_version,omitempty"`
		DataAPIVersion          string                 `json:"data_api_version,omitempty"`
		EndpointURI             string                 `json:"endpoint_uri,omitempty"`
		Preview                 Preview                `json:"preview,omitempty"`
		WhatsAppBusinessAccount map[string]interface{} `json:"whatsapp_business_account,omitempty"` // Optional
		Application             map[string]interface{} `json:"application,omitempty"`               // Optional
		HealthStatus            HealthStatus           `json:"health_status,omitempty"`
	}

	BaseResponseData struct {
		Name             string            `json:"name,omitempty"`
		AssetType        string            `json:"asset_type,omitempty"`
		DownloadURL      string            `json:"download_url,omitempty"`
		ID               string            `json:"id,omitempty"`
		Status           string            `json:"status,omitempty"`
		Categories       []string          `json:"categories,omitempty"`
		ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
	}

	Sender interface {
		Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*BaseResponse, error)
	}

	SenderFunc func(ctx context.Context, conf *config.Config, req *BaseRequest) (*BaseResponse, error)

	SenderMiddleware func(sender SenderFunc) SenderFunc
)

func (fn SenderFunc) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*BaseResponse, error) {
	return fn(ctx, conf, req)
}

// FlowDetails converts BaseResponse into a SingleFlowResponse.
func (baseResp *BaseResponse) FlowDetails() *SingleFlowResponse {
	return &SingleFlowResponse{
		ID:                      baseResp.ID,
		Name:                    baseResp.Name,
		Status:                  baseResp.Status,
		Categories:              baseResp.Categories,
		ValidationErrors:        baseResp.ValidationErrors,
		JSONVersion:             baseResp.JSONVersion,
		DataAPIVersion:          baseResp.DataAPIVersion,
		EndpointURI:             baseResp.EndpointURI,
		Preview:                 baseResp.Preview,
		WhatsAppBusinessAccount: baseResp.WhatsAppBusinessAccount,
		Application:             baseResp.Application,
		HealthStatus:            baseResp.HealthStatus,
	}
}

func (baseResp *BaseResponse) ListAssetsResponse() *RetrieveAssetsResponse {
	assets := make([]*Asset, len(baseResp.Data))

	for i, baseData := range baseResp.Data {
		assets[i] = &Asset{
			Name:        baseData.Name,
			AssetType:   baseData.AssetType,
			DownloadURL: baseData.DownloadURL,
		}
	}

	return &RetrieveAssetsResponse{
		Data:   assets,
		Paging: baseResp.Paging,
	}
}

func (baseResp *BaseResponse) ListResponse() *ListResponse {
	flows := make([]*Flow, len(baseResp.Data))

	for i, baseData := range baseResp.Data {
		flows[i] = &Flow{
			ID:               baseData.ID,
			Name:             baseData.Name,
			Status:           baseData.Status,
			Categories:       baseData.Categories,
			ValidationErrors: baseData.ValidationErrors,
		}
	}

	return &ListResponse{
		Data:   flows,
		Paging: baseResp.Paging,
	}
}

func (client *BaseClient) Send(ctx context.Context, conf *config.Config, req *BaseRequest) (*BaseResponse, error) {
	if req.QueryParams == nil {
		req.QueryParams = map[string]string{}
	}

	endpoints := []string{conf.APIVersion}
	endpoints = append(endpoints, req.Endpoints...)
	request := &whttp.Request[any]{
		Type:        req.Type,
		Method:      req.Method,
		Bearer:      conf.AccessToken,
		BaseURL:     conf.BaseURL,
		Endpoints:   endpoints,
		QueryParams: req.QueryParams,
		Message:     &req.Payload,
	}

	response := &BaseResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  false,
	})
	if err := client.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type (
	BaseClient struct {
		Reader config.Reader
		Sender whttp.AnySender
	}

	GetRequest struct {
		FlowID string
		Fields []string
	}
)

func NewBaseClient(reader config.Reader, sender whttp.AnySender) *BaseClient {
	return &BaseClient{
		Reader: reader,
		Sender: sender,
	}
}

func (client *BaseClient) Get(ctx context.Context, request *GetRequest) (*SingleFlowResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeRetrieveFlowDetails,
		Method:    http.MethodGet,
		Endpoints: []string{request.FlowID},
		Payload:   request,
		QueryParams: map[string]string{
			"fields": strings.Join(request.Fields, ","),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("send get request: %w", err)
	}

	return response.FlowDetails(), nil
}

func (client *BaseClient) GeneratePreview(ctx context.Context, request *PreviewRequest) (*PreviewResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeRetrieveFlowPreview,
		Method:    http.MethodGet,
		Endpoints: []string{request.FlowID},
		Payload:   request,
		QueryParams: map[string]string{
			"fields": fmt.Sprintf("preview.invalidate(%t)", request.Invalidate),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("send preview request: %w", err)
	}

	return &PreviewResponse{
		PreviewURL: response.PreviewURL,
		ExpiresAt:  response.ExpiresAt,
		ID:         response.ID,
	}, nil
}

func (client *BaseClient) UpdateFlowJSON(ctx context.Context,
	request *UpdateFlowJSONRequest,
) (*UpdateFlowJSONResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	form := &whttp.RequestForm{
		Fields: map[string]string{
			"name":       request.Name,
			"asset_type": "FLOW_JSON",
		},
		FormFile: &whttp.FormFile{
			Name: "file",
			Path: request.File,
		},
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeUploadMedia),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestForm[any](form),
		whttp.WithRequestEndpoints[any](conf.APIVersion, conf.PhoneNumberID, "media"),
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestSecured[any](conf.SecureRequests),
	}

	req := whttp.MakeRequest[any](http.MethodPost, conf.BaseURL, opts...)

	var resp UpdateFlowJSONResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
	})

	if err = client.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("upload media failed: %w", err)
	}

	return &resp, nil
}

func (client *BaseClient) Update(ctx context.Context, id string, request UpdateRequest) (*UpdateResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeUpdateFlow,
		Method:    http.MethodPost,
		Endpoints: []string{id},
		Payload:   request,
	})
	if err != nil {
		return nil, fmt.Errorf("send update request: %w", err)
	}

	return &UpdateResponse{Success: response.Success}, nil
}

func (client *BaseClient) Create(ctx context.Context, request CreateRequest) (*CreateResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeCreateFlow,
		Method:    http.MethodPost,
		Endpoints: []string{conf.BusinessAccountID, "flows"},
		Payload:   request,
	})
	if err != nil {
		return nil, fmt.Errorf("send create request: %w", err)
	}

	return &CreateResponse{ID: response.ID}, nil
}

func (client *BaseClient) ListAll(ctx context.Context) (*ListResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeRetrieveFlows,
		Method:    http.MethodGet,
		Endpoints: []string{conf.BusinessAccountID, "flows"},
	})
	if err != nil {
		return nil, fmt.Errorf("send list request: %w", err)
	}

	return response.ListResponse(), nil
}

func (client *BaseClient) Delete(ctx context.Context, id string) (*SuccessResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeDeleteFlow,
		Method:    http.MethodDelete,
		Endpoints: []string{id},
	})
	if err != nil {
		return nil, fmt.Errorf("send delete request: %w", err)
	}

	return &SuccessResponse{Success: response.Success}, nil
}

func (client *BaseClient) ListAssets(ctx context.Context, id string) (*RetrieveAssetsResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeRetrieveAssets,
		Method:    http.MethodGet,
		Endpoints: []string{id, "assets"},
	})
	if err != nil {
		return nil, fmt.Errorf("send list request: %w", err)
	}

	return response.ListAssetsResponse(), nil
}

func (client *BaseClient) Publish(ctx context.Context, id string) (*SuccessResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypePublishFlow,
		Method:    http.MethodPost,
		Endpoints: []string{id, "publish"},
	})
	if err != nil {
		return nil, fmt.Errorf("send publish request: %w", err)
	}

	return &SuccessResponse{Success: response.Success}, nil
}

func (client *BaseClient) Deprecate(ctx context.Context, id string) (*SuccessResponse, error) {
	conf, err := client.Reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	response, err := client.Send(ctx, conf, &BaseRequest{
		Type:      whttp.RequestTypeDeprecateFlow,
		Method:    http.MethodPost,
		Endpoints: []string{id, "deprecate"},
	})
	if err != nil {
		return nil, fmt.Errorf("send deprecate request: %w", err)
	}

	return &SuccessResponse{Success: response.Success}, nil
}

var _ Service = (*BaseClient)(nil)

type (
	Service interface {
		Create(ctx context.Context, request CreateRequest) (*CreateResponse, error)
		Update(ctx context.Context, id string, request UpdateRequest) (*UpdateResponse, error)
		UpdateFlowJSON(ctx context.Context, request *UpdateFlowJSONRequest) (*UpdateFlowJSONResponse, error)
		ListAll(ctx context.Context) (*ListResponse, error)
		ListAssets(ctx context.Context, id string) (*RetrieveAssetsResponse, error)
		Publish(ctx context.Context, id string) (*SuccessResponse, error)
		Delete(ctx context.Context, id string) (*SuccessResponse, error)
		Deprecate(ctx context.Context, id string) (*SuccessResponse, error)
		GeneratePreview(ctx context.Context, request *PreviewRequest) (*PreviewResponse, error)
		Get(ctx context.Context, request *GetRequest) (*SingleFlowResponse, error)
		GetFlowMetrics(ctx context.Context, request *MetricsRequest) (*MetricsAPIResponse, error)
	}
)

type PreviewURLConfigOptions struct {
	Interactive       bool   // If true, the preview will run in interactive mode.
	FlowToken         string // Token sent with requests for Flows with an endpoint.
	FlowAction        string // The initial action: "navigate" or "data_exchange".
	FlowActionPayload string // JSON-encoded initial screen data.
	PhoneNumber       string // Phone number used for sending the Flow (used for encryption).
	Debug             bool   // Show actions in a separate panel while interacting with the preview.
}

func ConfigurePreviewURL(response *PreviewResponse, options *PreviewURLConfigOptions) (string, error) {
	u, err := url.Parse(response.PreviewURL)
	if err != nil {
		return "", fmt.Errorf("error parsing preview URL: %w", err)
	}

	params := url.Values{}
	if options.Interactive {
		params.Add("interactive", "true")
	}

	if options.FlowToken != "" {
		params.Add("flow_token", options.FlowToken)
	}

	if options.FlowAction != "" {
		params.Add("flow_action", options.FlowAction)
	}

	if options.FlowActionPayload != "" {
		params.Add("flow_action_payload", options.FlowActionPayload)
	}

	if options.PhoneNumber != "" {
		params.Add("phone_number", options.PhoneNumber)
	}

	if options.Debug {
		params.Add("debug", "true")
	}

	u.RawQuery = params.Encode()

	return u.String(), nil
}

type Event int

const (
	EventUserOpenFlow Event = iota
	EventUserSubmitScreen
	EventUserPressBackButton
	EventUserChangeComponent
	EventInvalidContentReply
	EventHealthCheck
)

type (
	DataExchangeHandlerImpl struct {
		Handler          DataExchangeHandler
		PrivateKeyLoader func(ctx context.Context) (*rsa.PrivateKey, error)
	}

	DataExchangeHandler interface {
		Handle(ctx context.Context, request *DataExchangeRequest) (*Response, error)
	}

	DataExchangeHandlerFunc func(ctx context.Context, request *DataExchangeRequest) (*Response, error)

	DataExchangeRequestHandlerMiddleware func(handlerFunc DataExchangeHandlerFunc) DataExchangeHandlerFunc

	Request struct {
		EncryptedFlowData string `json:"encrypted_flow_data"`
		EncryptedAesKey   string `json:"encrypted_aes_key"`
		InitialVector     string `json:"initial_vector"`
	}

	DecryptedRequest struct {
		FlowData      []byte
		AesKey        []byte
		InitialVector []byte
	}

	DataExchangeRequest struct {
		Version   string                 `json:"version"`
		Action    string                 `json:"action"`
		Screen    string                 `json:"screen"`
		Data      map[string]interface{} `json:"data"`
		FlowToken string                 `json:"flow_token"`
	}

	Response struct {
		Screen string                 `json:"screen,omitempty"`
		Data   map[string]interface{} `json:"data"`
	}

	NextScreenResponseData struct {
		Screen       string
		Data         map[string]interface{}
		ErrorMessage string
	}

	FinalScreenResponseData struct {
		FlowToken      string
		OptionalParams map[string]interface{}
	}
)

func (f DataExchangeHandlerFunc) Handle(ctx context.Context, request *DataExchangeRequest) (*Response, error) {
	return f(ctx, request)
}

func NewDataExchangeHandler(loader func(ctx context.Context) (*rsa.PrivateKey, error),
	baseHandler DataExchangeHandler, mws ...DataExchangeRequestHandlerMiddleware,
) *DataExchangeHandlerImpl {
	handlerFunc := DataExchangeHandlerFunc(baseHandler.Handle)

	for i := len(mws) - 1; i >= 0; i-- {
		handlerFunc = mws[i](handlerFunc)
	}

	return &DataExchangeHandlerImpl{
		Handler:          handlerFunc,
		PrivateKeyLoader: loader,
	}
}

func CreateNextScreenResponse(nextScreenData NextScreenResponseData) *Response {
	if nextScreenData.ErrorMessage != "" {
		nextScreenData.Data["error_message"] = nextScreenData.ErrorMessage
	}

	return &Response{
		Screen: nextScreenData.Screen,
		Data:   nextScreenData.Data,
	}
}

func CreateFinalScreenResponse(finalScreenData FinalScreenResponseData) *Response {
	params := map[string]interface{}{
		"flow_token": finalScreenData.FlowToken,
	}
	for k, v := range finalScreenData.OptionalParams {
		params[k] = v
	}

	return &Response{
		Screen: "SUCCESS",
		Data: map[string]interface{}{
			"extension_message_response": map[string]interface{}{
				"params": params,
			},
		},
	}
}

func CreateHealthCheckResponse(status string) *Response {
	return &Response{
		Data: map[string]interface{}{
			"status": status,
		},
	}
}

func CreateErrorAcknowledgmentResponse(acknowledged bool) *Response {
	return &Response{
		Data: map[string]interface{}{
			"acknowledged": acknowledged,
		},
	}
}

func (h *DataExchangeHandlerImpl) Handle(w http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	var req *Request
	if err := whttp.DecodeRequestJSON(request, req, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  false,
	}); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	decryptedRequest, err := h.DecryptRequest(ctx, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	var message *DataExchangeRequest
	if err = json.Unmarshal(decryptedRequest.FlowData, &message); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	response, err := h.Handler.Handle(ctx, message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp, err := h.EncryptResponse(response, decryptedRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resp))
}

func (h *DataExchangeHandlerImpl) EncryptResponse(response *Response, request *DecryptedRequest) (string, error) {
	// flip the IV
	flippedIV := make([]byte, len(request.InitialVector))
	for i, b := range request.InitialVector {
		flippedIV[i] = b ^ 0xFF //nolint:mnd // ok
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response data: %w", err)
	}

	encryptedResponse, tag, err := aesGCMEncrypt(responseData, request.AesKey, flippedIV)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt response: %w", err)
	}

	encryptedResponse = append(encryptedResponse, tag...)

	return base64.StdEncoding.EncodeToString(encryptedResponse), nil
}

func (h *DataExchangeHandlerImpl) DecryptRequest(ctx context.Context, request *Request) (*DecryptedRequest, error) {
	privateKey, err := h.PrivateKeyLoader(ctx)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	aesKey, err := base64.StdEncoding.DecodeString(request.EncryptedAesKey)
	if err != nil {
		return nil, fmt.Errorf("decode aes key: %w", err)
	}

	decryptedAesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt aes key: %w", err)
	}

	encryptedFlowData, err := base64.StdEncoding.DecodeString(request.EncryptedFlowData)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted flow data: %w", err)
	}

	initialVector, err := base64.StdEncoding.DecodeString(request.InitialVector)
	if err != nil {
		return nil, fmt.Errorf("decode initial vector: %w", err)
	}

	tagLength := 16
	encryptedFlowDataBody := encryptedFlowData[:len(encryptedFlowData)-tagLength]
	encryptedFlowDataTag := encryptedFlowData[len(encryptedFlowData)-tagLength:]

	decryptedFlowData, err := aesGCMDecrypt(encryptedFlowDataBody, decryptedAesKey, initialVector, encryptedFlowDataTag)
	if err != nil {
		return nil, err
	}

	return &DecryptedRequest{
		FlowData:      decryptedFlowData,
		AesKey:        aesKey,
		InitialVector: initialVector,
	}, nil
}

func aesGCMDecrypt(ciphertext, key, iv, tag []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new GCM: %w", err)
	}

	bb, err := aesgcm.Open(nil, iv, append(ciphertext, tag...), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return bb, nil
}

func aesGCMEncrypt(plaintext, key, iv []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new GCM: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, iv, plaintext, nil)
	tag := ciphertext[len(ciphertext)-aesgcm.Overhead():]

	return ciphertext[:len(ciphertext)-aesgcm.Overhead()], tag, nil
}
