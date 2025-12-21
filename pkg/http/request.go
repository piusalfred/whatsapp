package http

import (
	"context"
	"fmt"
	"io"
	http2 "net/http"
	"net/url"

	"github.com/piusalfred/whatsapp/pkg/crypto"
	"github.com/piusalfred/whatsapp/pkg/types"
)

type (
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
		DownloadURL    string // this is used for downloading media (it is taken as is)
		BodyReader     io.Reader
		debugLogLevel  DebugLogLevel
	}

	RequestForm struct {
		Fields   map[string]string
		FormFile *FormFile
	}

	FormFile struct {
		Name string
		Path string
		Type string
	}

	RequestOption[T any] func(request *Request[T])
)

// SetDebugLogLevel sets the debug log level for the request.
func (request *Request[T]) SetDebugLogLevel(level DebugLogLevel) {
	request.debugLogLevel = level
}

// MakeRequest creates a new request with the provided options.
func MakeRequest[T any](method, baseURL string, options ...RequestOption[T]) *Request[T] {
	req := &Request[T]{
		Method:        method,
		BaseURL:       baseURL,
		Headers:       make(map[string]string),
		QueryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
	}

	for _, option := range options {
		if option != nil {
			option(req)
		}
	}

	return req
}

// MakeDownloadRequest creates a new request for downloading media.
func MakeDownloadRequest[T any](downloadURL string, options ...RequestOption[T]) *Request[T] {
	req := &Request[T]{
		Method:        http2.MethodGet,
		DownloadURL:   downloadURL,
		Headers:       make(map[string]string),
		QueryParams:   make(map[string]string),
		debugLogLevel: DebugLogLevelNone,
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
) (*http2.Request, error) {
	req := MakeRequest(method, baseURL, options...)

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

func (request *Request[T]) SetEndpoints(endpoints ...string) {
	request.Endpoints = endpoints
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

// SetRequestMessage sets the body of the request.
func (request *Request[T]) SetRequestMessage(message *T) {
	request.Message = message
}

func WithRequestForm[T any](form *RequestForm) RequestOption[T] {
	return func(request *Request[T]) {
		request.Form = form
	}
}

// WithRequestAppSecret sets the app secret for the request and turns on secure requests.
func WithRequestAppSecret[T any](appSecret string) RequestOption[T] {
	return func(request *Request[T]) {
		if appSecret != "" {
			request.SecureRequests = true
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

func WithRequestBodyReader[T any](bodyReader io.Reader) RequestOption[T] {
	return func(request *Request[T]) {
		request.BodyReader = bodyReader
	}
}

// URL returns the formatted URL for the request.
func (request *Request[T]) URL() (string, error) {
	if request.DownloadURL != "" {
		return request.DownloadURL, nil
	}

	fmtURL, err := url.JoinPath(request.BaseURL, request.Endpoints...)
	if err != nil {
		return "", fmt.Errorf("format url: %w", err)
	}

	parsedURL, err := url.Parse(fmtURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	q := parsedURL.Query()

	for key, value := range request.QueryParams {
		q.Set(key, value)
	}

	shouldEnableDebugLogging := request.debugLogLevel != DebugLogLevelNone && request.debugLogLevel != ""
	if shouldEnableDebugLogging {
		q.Set("debug", string(request.debugLogLevel))
	}

	if request.SecureRequests {
		proof, proofErr := crypto.GenerateAppSecretProof(request.Bearer, request.AppSecret)
		if proofErr != nil {
			return "", fmt.Errorf("failed to generate app secret proof: %w", proofErr)
		}
		q.Set("appsecret_proof", proof)
	}

	parsedURL.RawQuery = q.Encode()

	return parsedURL.String(), nil
}

func RequestWithContext[T any](ctx context.Context, req *Request[T]) (*http2.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("request: %w", ErrNilRequest)
	}
	ctx = InjectMessageMetadata(ctx, req.Metadata)

	parsedURL, err := req.URL()
	if err != nil {
		return nil, fmt.Errorf("format url: %w", err)
	}

	var body io.Reader
	contentType := "application/json"

	if req.Message != nil {
		encodeResp, encodeErr := EncodePayload(req.Message)
		if encodeErr != nil {
			return nil, fmt.Errorf("failed to encode request payload: %w", encodeErr)
		}
		body = encodeResp.Body
		contentType = encodeResp.ContentType
	}

	if req.Form != nil {
		encodeResp, encodeErr := EncodePayload(req.Form)
		if encodeErr != nil {
			return nil, fmt.Errorf("failed to encode request payload: %w", encodeErr)
		}
		body = encodeResp.Body
		contentType = encodeResp.ContentType
	}

	if req.BodyReader != nil {
		body = req.BodyReader
		contentType = "application/octet-stream"
	}

	r, err := http2.NewRequestWithContext(ctx, req.Method, parsedURL, body)
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
