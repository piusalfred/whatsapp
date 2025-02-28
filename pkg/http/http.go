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

package http

//go:generate mockgen -destination=../../mocks/http/mock_http.go -package=http -source=http.go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/crypto"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	"github.com/piusalfred/whatsapp/pkg/types"
)

type (
	CoreClient[T any] struct {
		http        *http.Client
		reqHook     RequestInterceptorFunc
		resHook     ResponseInterceptorFunc
		middlewares []Middleware[T]
		sender      Sender[T]
	}

	CoreClientOption[T any] func(client *CoreClient[T])
)

func (core *CoreClient[T]) SetHTTPClient(httpClient *http.Client) {
	if httpClient != nil {
		core.http = httpClient
	}
}

func (core *CoreClient[T]) SetRequestInterceptor(hook RequestInterceptorFunc) {
	core.reqHook = hook
}

func (core *CoreClient[T]) SetBaseSender(sender Sender[T]) {
	core.sender = sender
}

func (core *CoreClient[T]) SetResponseInterceptor(hook ResponseInterceptorFunc) {
	core.resHook = hook
}

func (core *CoreClient[T]) AppendMiddlewares(mws ...Middleware[T]) {
	core.middlewares = append(core.middlewares, mws...)
}

func (core *CoreClient[T]) PrependMiddlewares(mws ...Middleware[T]) {
	core.middlewares = append(mws, core.middlewares...)
}

func WithCoreClientHTTPClient[T any](httpClient *http.Client) CoreClientOption[T] {
	return func(client *CoreClient[T]) {
		client.http = httpClient
	}
}

func WithCoreClientRequestInterceptor[T any](hook RequestInterceptorFunc) CoreClientOption[T] {
	return func(client *CoreClient[T]) {
		client.reqHook = hook
	}
}

func WithCoreClientResponseInterceptor[T any](hook ResponseInterceptorFunc) CoreClientOption[T] {
	return func(client *CoreClient[T]) {
		client.resHook = hook
	}
}

func WithCoreClientMiddlewares[T any](mws ...Middleware[T]) CoreClientOption[T] {
	return func(client *CoreClient[T]) {
		client.middlewares = mws
	}
}

func NewSender[T any](options ...CoreClientOption[T]) *CoreClient[T] {
	core := &CoreClient[T]{
		http: http.DefaultClient,
	}

	core.sender = SenderFunc[T](core.send)

	for _, option := range options {
		if option != nil {
			option(core)
		}
	}

	return core
}

func NewAnySender(options ...CoreClientOption[any]) *CoreClient[any] {
	core := &CoreClient[any]{
		http: http.DefaultClient,
	}

	core.sender = SenderFunc[any](core.send)

	for _, option := range options {
		if option != nil {
			option(core)
		}
	}

	return core
}

func (core *CoreClient[T]) send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	if err := SendFuncWithInterceptors[T](core.http, core.reqHook, core.resHook)(ctx, request, decoder); err != nil {
		return err
	}

	return nil
}

func SendFuncWithInterceptors[T any](client *http.Client, reqHook RequestInterceptorFunc,
	resHook ResponseInterceptorFunc,
) SenderFunc[T] {
	fn := SenderFunc[T](func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
		req, err := RequestWithContext(ctx, request)
		if err != nil {
			return err
		}

		if reqHook != nil {
			if errHook := reqHook(ctx, req); errHook != nil {
				return errHook
			}
		}

		response, err := client.Do(req) //nolint:bodyclose
		if err != nil {
			return fmt.Errorf("send request: %w", err)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)

		if resHook != nil {
			bodyBytes, errRead := io.ReadAll(response.Body)
			if errRead != nil && !errors.Is(errRead, io.EOF) {
				return fmt.Errorf("read response body: %w", errRead)
			}
			response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if errHook := resHook.InterceptResponse(ctx, response); errHook != nil {
				return errHook
			}
			response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if err := decoder.Decode(ctx, response); err != nil {
			return fmt.Errorf("core send: decode: %w", err)
		}

		return nil
	})

	return fn
}

func (core *CoreClient[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	fn := wrapMiddlewares(core.sender.Send, core.middlewares)

	return fn(ctx, request, decoder)
}

type (
	Sender[T any] interface {
		Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error
	}

	SenderFunc[T any] func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error

	Middleware[T any] func(next SenderFunc[T]) SenderFunc[T]

	AnySender Sender[any]

	AnySenderFunc SenderFunc[any]
)

func (fn SenderFunc[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	return fn(ctx, request, decoder)
}

func (fn AnySenderFunc) Send(ctx context.Context, request *Request[any], decoder ResponseDecoder) error {
	return fn(ctx, request, decoder)
}

func wrapMiddlewares[T any](doFunc SenderFunc[T], middlewares []Middleware[T]) SenderFunc[T] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] != nil {
			doFunc = middlewares[i](doFunc)
		}
	}

	return doFunc
}

const (
	RequestTypeSendMessage RequestType = iota
	RequestTypeUpdateStatus
	RequestTypeCreateQR
	RequestTypeListQR
	RequestTypeGetQR
	RequestTypeUpdateQR
	RequestTypeDeleteQR
	RequestTypeListPhoneNumbers
	RequestTypeGetPhoneNumber
	RequestTypeDownloadMedia
	RequestTypeUploadMedia
	RequestTypeDeleteMedia
	RequestTypeGetMedia
	RequestTypeUpdateBusinessProfile
	RequestTypeGetBusinessProfile
	RequestTypeRetrieveFlows
	RequestTypeRetrieveFlowDetails
	RequestTypeRetrieveAssets
	RequestTypePublishFlow
	RequestTypeDeprecateFlow
	RequestTypeDeleteFlow
	RequestTypeUpdateFlow
	RequestTypeCreateFlow
	RequestTypeRetrieveFlowPreview
	RequestTypeGetFlowMetrics
	RequestTypeInstallApp
	RequestTypeRefreshToken
	RequestTypeGenerateToken
	RequestTypeRevokeToken
	RequestTypeTwoStepVerification
	RequestTypeFetchMessagingAnalytics
	RequestTypeFetchTemplateAnalytics
	RequestTypeFetchPricingAnalytics
	RequestTypeFetchConversationAnalytics
	RequestTypeEnableTemplatesAnalytics
	RequestTypeDisableButtonClickTracking
	RequestTypeBlockUsers
	RequestTypeUnblockUsers
	RequestTypeListBlockedUsers
)

// String returns the string representation of the request type.
func (r RequestType) String() string {
	return [...]string{
		"send_message",
		"update_status",
		"create_qr",
		"list_qr",
		"get_qr",
		"update_qr",
		"delete_qr",
		"list_phone_numbers",
		"get_phone_number",
		"download_media",
		"upload_media",
		"delete_media",
		"get_media",
		"update_business_profile",
		"get_business_profile",
		"retrieve_flows",
		"retrieve_flow_details",
		"retrieve_assets",
		"publish_flow",
		"deprecate_flow",
		"delete_flow",
		"update_flow",
		"create_flow",
		"retrieve_flow_preview",
		"get_flow_metrics",
		"install_app",
		"refresh_token",
		"generate_token",
		"revoke_token",
		"two_step_verification",
		"fetch_messaging_analytics",
		"fetch_template_analytics",
		"fetch_pricing_analytics",
		"fetch_conversation_analytics",
		"enable_templates_analytics",
		"disable_button_click_tracking",
	}[r]
}

type (
	RequestType uint8

	Request[T any] struct {
		Type           RequestType
		Method         string
		Bearer         string
		Headers        map[string]string
		QueryParams    map[string]string
		BaseURL        string
		Endpoints      []string
		Metadata       types.Metadata
		Message        *T
		Form           *RequestForm
		AppSecret      string
		SecureRequests bool
	}

	RequestForm struct {
		Fields   map[string]string
		FormFile *FormFile
	}

	FormFile struct {
		Name string
		Path string
	}

	RequestOption[T any] func(request *Request[T])
)

// MakeRequest creates a new request with the provided options.
func MakeRequest[T any](method, baseURL string, options ...RequestOption[T]) *Request[T] {
	req := &Request[T]{
		Method:      method,
		BaseURL:     baseURL,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}

	for _, option := range options {
		if option != nil {
			option(req)
		}
	}

	return req
}

// NewRequestWithContext ...
func NewRequestWithContext[T any](ctx context.Context, method, baseURL string,
	options ...RequestOption[T],
) (*http.Request, error) {
	req := MakeRequest[T](method, baseURL, options...)

	return RequestWithContext(ctx, req)
}

// WithRequestType sets the request type for the request.
func WithRequestType[T any](requestType RequestType) RequestOption[T] {
	return func(request *Request[T]) {
		request.Type = requestType
	}
}

// WithRequestBearer sets the bearer token for the request.
func WithRequestBearer[T any](bearer string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Bearer = bearer
	}
}

// WithRequestEndpoints sets the endpoints for the request.
func WithRequestEndpoints[T any](endpoints ...string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Endpoints = endpoints
	}
}

// WithRequestMetadata sets the metadata for the request.
func WithRequestMetadata[T any](metadata types.Metadata) RequestOption[T] {
	return func(request *Request[T]) {
		request.Metadata = metadata
	}
}

// WithRequestHeaders sets the headers for the request.
func WithRequestHeaders[T any](headers map[string]string) RequestOption[T] {
	return func(request *Request[T]) {
		request.Headers = headers
	}
}

// WithRequestQueryParams sets the query parameters for the request.
func WithRequestQueryParams[T any](queryParams map[string]string) RequestOption[T] {
	return func(request *Request[T]) {
		request.QueryParams = queryParams
	}
}

// WithRequestMessage sets the message for the request.
func WithRequestMessage[T any](message *T) RequestOption[T] {
	return func(request *Request[T]) {
		request.Message = message
	}
}

func WithRequestForm[T any](form *RequestForm) RequestOption[T] {
	return func(request *Request[T]) {
		request.Form = form
	}
}

// WithRequestAppSecret sets the app secret for the request and turns on secure requests.
func WithRequestAppSecret[T any](appSecret string) RequestOption[T] {
	return func(request *Request[T]) {
		if request.AppSecret != "" {
			request.AppSecret = appSecret
		}
	}
}

// WithRequestSecured sets the request to be secure.
func WithRequestSecured[T any](secured bool) RequestOption[T] {
	return func(request *Request[T]) {
		request.SecureRequests = secured
	}
}

var errNilRequest = errors.New("nil request provided")

func RequestWithContext[T any](ctx context.Context, req *Request[T]) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("request: %w", errNilRequest)
	}
	ctx = InjectMessageMetadata(ctx, req.Metadata)

	fmtURL, err := url.JoinPath(req.BaseURL, req.Endpoints...)
	if err != nil {
		return nil, fmt.Errorf("format url: %w", err)
	}

	parsedURL, err := url.Parse(fmtURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := parsedURL.Query()

	for key, value := range req.QueryParams {
		q.Set(key, value)
	}

	if req.SecureRequests {
		proof, err := crypto.GenerateAppSecretProof(req.Bearer, req.AppSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to generate app secret proof: %w", err)
		}
		q.Set("appsecret_proof", proof)
	}

	parsedURL.RawQuery = q.Encode()

	var body io.Reader
	contentType := "application/json"

	if req.Message != nil {
		encodeResp, err := EncodePayload(req.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request payload: %w", err)
		}
		body = encodeResp.Body
		contentType = encodeResp.ContentType
	}

	r, err := http.NewRequestWithContext(ctx, req.Method, parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	r.Header.Set("Content-Type", contentType)

	if req.Bearer != "" {
		r.Header.Set("Authorization", "Bearer "+req.Bearer)
	}

	for key, value := range req.Headers {
		r.Header.Set(key, value)
	}

	return r, nil
}

type EncodeResponse struct {
	Body        io.Reader
	ContentType string
}

// EncodePayload takes different types of payloads (form data, readers, JSON) and returns an EncodeResponse.
func EncodePayload(payload any) (*EncodeResponse, error) {
	switch p := payload.(type) {
	case nil:
		return &EncodeResponse{
			Body:        nil,
			ContentType: "application/json",
		}, nil
	case *RequestForm:
		body, contentType, err := encodeFormData(p)
		if err != nil {
			return nil, fmt.Errorf("failed to encode form data: %w", err)
		}

		return &EncodeResponse{
			Body:        body,
			ContentType: contentType,
		}, nil
	case io.Reader:
		return &EncodeResponse{
			Body:        p,
			ContentType: "application/octet-stream",
		}, nil
	case []byte:
		return &EncodeResponse{
			Body:        bytes.NewReader(p),
			ContentType: "application/octet-stream",
		}, nil
	case string:
		return &EncodeResponse{
			Body:        strings.NewReader(p),
			ContentType: "text/plain",
		}, nil
	default:
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(p); err != nil {
			return nil, fmt.Errorf("failed to encode payload as JSON: %w", err)
		}

		return &EncodeResponse{
			Body:        buf,
			ContentType: "application/json",
		}, nil
	}
}

// encodeFormData encodes form fields and file data into multipart/form-data.
func encodeFormData(formData *RequestForm) (io.Reader, string, error) {
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	for key, value := range formData.Fields {
		err := writer.WriteField(key, value)
		if err != nil {
			return nil, "", fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	if formData.FormFile != nil {
		file, err := os.Open(formData.FormFile.Path)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open file %s: %w", formData.FormFile.Path, err)
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		part, err := writer.CreateFormFile(formData.FormFile.Name, filepath.Base(formData.FormFile.Path))
		if err != nil {
			return nil, "", fmt.Errorf("failed to create form file part: %w", err)
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return &payload, writer.FormDataContentType(), nil
}

type DecodeOptions struct {
	DisallowUnknownFields bool
	DisallowEmptyResponse bool
	InspectResponseError  bool
}

func DecodeResponseJSON[T any](response *http.Response, v *T, opts DecodeOptions) error {
	if response == nil {
		return ErrNilResponse
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	if len(responseBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty response body", ErrEmptyResponseBody)
	}

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode < 300

	if !isResponseOk { //nolint:nestif
		if opts.InspectResponseError {
			if len(responseBody) == 0 {
				return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
			}

			var errorResponse ResponseError
			if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
				return fmt.Errorf("%w: %w, status code: %d", ErrDecodeErrorResponse, err, response.StatusCode)
			}

			return &errorResponse
		}

		return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
	}

	//	At this point, we know the response is 2xx
	//	If the body is empty here, that means empty bodies are allowed, so we return early.
	//	If the body is not empty but `v` is nil, we return an error as there is no target to decode into.
	//	Otherwise, we proceed with decoding the body into `v` if the body is not empty and `v` is provided.

	if len(responseBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, err)
	}

	return nil
}

func DecodeRequestJSON[T any](request *http.Request, v *T, opts DecodeOptions) error {
	if request == nil {
		return ErrNilResponse
	}

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}()

	if len(requestBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty request body", ErrEmptyResponseBody)
	}

	//	At this point, we know:
	//	- If the body is empty, it's allowed based on `DisallowEmptyResponse`.
	//	- If the body is not empty and `v == nil`, return an error as there’s no target to decode into.
	//	- Otherwise, proceed with decoding into `v` if the body is not empty and `v` is provided.

	if len(requestBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(requestBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if decodeErr := decoder.Decode(v); decodeErr != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, decodeErr)
	}

	return nil
}

type ResponseError struct {
	Code int            `json:"code,omitempty"`
	Err  *werrors.Error `json:"error,omitempty"`
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("whatsapp message error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

func (e *ResponseError) Unwrap() error {
	return e.Err
}

const (
	ErrNilResponse         = httpError("nil response provided")
	ErrEmptyResponseBody   = httpError("empty response body")
	ErrNilTarget           = httpError("nil value passed for decoding target")
	ErrRequestFailure      = httpError("request failed")
	ErrDecodeResponseBody  = httpError("failed to decode response body")
	ErrDecodeErrorResponse = httpError("failed to decode error response")
)

type httpError string

func (e httpError) Error() string {
	return string(e)
}

const MessageMetadataContextKey = "message-metadata-key"

type MessageContextKey string

func InjectMessageMetadata(ctx context.Context, metadata types.Metadata) context.Context {
	return context.WithValue(ctx, MessageContextKey(MessageMetadataContextKey), metadata)
}

type (
	RequestInterceptorFunc func(ctx context.Context, request *http.Request) error
	RequestInterceptor     interface {
		InterceptRequest(ctx context.Context, request *http.Request) error
	}

	ResponseInterceptorFunc func(ctx context.Context, response *http.Response) error
	ResponseInterceptor     interface {
		InterceptResponse(ctx context.Context, response *http.Response) error
	}
)

func (fn RequestInterceptorFunc) InterceptRequest(ctx context.Context, request *http.Request) error {
	return fn(ctx, request)
}

func (fn ResponseInterceptorFunc) InterceptResponse(ctx context.Context, response *http.Response) error {
	return fn(ctx, response)
}

type (
	ResponseDecoderFunc func(ctx context.Context, response *http.Response) error

	ResponseDecoder interface {
		Decode(ctx context.Context, response *http.Response) error
	}

	ResponseBodyReaderFunc func(ctx context.Context, reader io.Reader) error
)

func (decoder ResponseDecoderFunc) Decode(ctx context.Context, response *http.Response) error {
	return decoder(ctx, response)
}

func ResponseDecoderJSON[T any](v *T, options DecodeOptions) ResponseDecoderFunc {
	fn := ResponseDecoderFunc(func(_ context.Context, response *http.Response) error {
		if err := DecodeResponseJSON(response, v, options); err != nil {
			return fmt.Errorf("decode json: %w", err)
		}

		return nil
	})

	return fn
}

func BodyReaderResponseDecoder(fn ResponseBodyReaderFunc) ResponseDecoderFunc {
	return func(ctx context.Context, response *http.Response) error {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer func() {
			response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}()

		if err := fn(ctx, bytes.NewReader(responseBody)); err != nil {
			return err
		}

		return nil
	}
}
