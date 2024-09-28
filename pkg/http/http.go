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

	werrors "github.com/piusalfred/libwhatsapp/pkg/errors"
	"github.com/piusalfred/libwhatsapp/pkg/types"
)

type (
	CoreClient[T any] struct {
		http        *http.Client
		reqHook     RequestInterceptorFunc
		resHook     ResponseInterceptorFunc
		middlewares []Middleware[T]
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

func WithCoreClientMiddlewares[T any](mws []Middleware[T]) CoreClientOption[T] {
	return func(client *CoreClient[T]) {
		client.middlewares = mws
	}
}

func NewSender[T any](options ...CoreClientOption[T]) *CoreClient[T] {
	core := &CoreClient[T]{
		http: http.DefaultClient,
	}

	for _, option := range options {
		if option != nil {
			option(core)
		}
	}

	return core
}

func (core *CoreClient[T]) send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	req, err := RequestWithContext(ctx, request)
	if err != nil {
		return err
	}

	if errHook := core.reqHook(ctx, req); errHook != nil {
		return errHook
	}

	response, err := core.http.Do(req)
	if err != nil {
		return fmt.Errorf("send req: %w", err)
	}
	defer response.Body.Close()

	if core.resHook != nil {
		bodyBytes, errRead := io.ReadAll(response.Body)
		if errRead != nil && !errors.Is(errRead, io.EOF) {
			return fmt.Errorf("read response body: %w", errRead)
		}
		response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if errHook := core.resHook(ctx, response); errHook != nil {
			return errHook
		}
		response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if err := decoder.Decode(ctx, response); err != nil {
		return fmt.Errorf("core send: decode: %w", err)
	}

	return nil
}

func (core *CoreClient[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
	fn := wrapMiddlewares(core.send, core.middlewares)

	return fn(ctx, request, decoder)
}

type (
	Sender[T any] interface {
		Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error
	}

	SenderFunc[T any] func(ctx context.Context, request *Request[T], decoder ResponseDecoder) error

	Middleware[T any] func(next SenderFunc[T]) SenderFunc[T]
)

func (fn SenderFunc[T]) Send(ctx context.Context, request *Request[T], decoder ResponseDecoder) error {
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
)

type (
	RequestType uint8

	Request[T any] struct {
		Type        RequestType
		Method      string
		Bearer      string
		Headers     map[string]string
		QueryParams map[string]string
		BaseURL     string
		Endpoints   []string
		Metadata    types.Metadata
		Message     *T
		Form        *RequestForm
	}

	RequestForm struct {
		Fields   map[string]string
		FormFile *FormFile
	}

	FormFile struct {
		Name string
		Path string
	}
)

func RequestWithContext[T any](ctx context.Context, req *Request[T]) (*http.Request, error) {
	if req == nil {
		return nil, fmt.Errorf("nil context")
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
		return nil, err
	}

	r.Header.Set("Content-Type", contentType)

	if req.Bearer != "" {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", req.Bearer))
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

// encodeFormData encodes form fields and file data into multipart/form-data
func encodeFormData(formData *RequestForm) (io.Reader, string, error) {
	// Create a buffer and a multipart writer
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	// Add regular form fields
	for key, value := range formData.Fields {
		err := writer.WriteField(key, value)
		if err != nil {
			return nil, "", fmt.Errorf("failed to write form field %s: %w", key, err)
		}
	}

	// Add the form file if present
	if formData.FormFile != nil {
		// Open the file
		file, err := os.Open(formData.FormFile.Path)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open file %s: %w", formData.FormFile.Path, err)
		}
		defer file.Close()

		// Create the form file part
		part, err := writer.CreateFormFile(formData.FormFile.Name, filepath.Base(formData.FormFile.Path))
		if err != nil {
			return nil, "", fmt.Errorf("failed to create form file part: %w", err)
		}

		// Copy the file content into the form part
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	// Close the multipart writer to finalize the form
	err := writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Return the body and the content type, which includes the multipart boundary
	return &payload, writer.FormDataContentType(), nil
}

type DecodeOptions struct {
	DisallowUnknownFields bool
	DisallowEmptyResponse bool
	InspectResponseError  bool
}

func DecodeResponseJSON[T any](response *http.Response, v *T, opts DecodeOptions) error {
	if response == nil {
		return fmt.Errorf("nil response provided")
	}

	if response.Body == nil {
		if opts.DisallowEmptyResponse {
			return fmt.Errorf("unexpected empty response body")
		}
		return nil
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode < 300
	if !isResponseOk {
		if opts.InspectResponseError {
			if len(responseBody) == 0 {
				return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
			}

			var errorResponse ResponseError
			if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
				return fmt.Errorf("failed to decode error response: %w, status code: %d, response body: %s", err, response.StatusCode, string(responseBody))
			}

			return &errorResponse
		}
	}

	if len(responseBody) == 0 {
		if opts.DisallowEmptyResponse {
			return fmt.Errorf("expected non-empty response body, but got empty")
		}
		return nil
	}

	if v == nil {
		return fmt.Errorf("nil value passed for decoding target")
	}

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if decodeErr := decoder.Decode(v); decodeErr != nil {
		return fmt.Errorf("decode response body: %w", decodeErr)
	}

	return nil
}

func DecodeRequestJSON[T any](request *http.Request, v *T, opts DecodeOptions) error {
	if request == nil {
		return fmt.Errorf("nil request provided")
	}

	if request.Body == nil {
		if opts.DisallowEmptyResponse {
			return fmt.Errorf("unexpected empty request body")
		}
		return nil
	}

	responseBody, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	if len(responseBody) == 0 {
		if opts.DisallowEmptyResponse {
			return fmt.Errorf("expected non-empty request body, but got empty")
		}
		return nil
	}

	if v == nil {
		return fmt.Errorf("nil value passed for decoding target")
	}

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if decodeErr := decoder.Decode(v); decodeErr != nil {
		return fmt.Errorf("decode request body: %w", decodeErr)
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

const ErrRequestFailure customErr = "send request failed"

type customErr string

func (e customErr) Error() string {
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
	fn := ResponseDecoderFunc(func(ctx context.Context, response *http.Response) error {
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

		if err := fn(ctx, bytes.NewReader(responseBody)); err != nil {
			return err
		}

		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))

		return nil
	}
}
