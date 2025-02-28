package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	gcmp "github.com/google/go-cmp/cmp"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type TestMessage struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestDecodeResponseJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		response *http.Response
		opts     whttp.DecodeOptions
		wantErr  bool
		want     *TestMessage
	}{
		{
			name: "successful decoding",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123}`)),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: false,
			want:    &TestMessage{Name: "test", Value: 123},
		},
		{
			name: "empty body with DisallowEmptyResponse",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			},
			opts:    whttp.DecodeOptions{DisallowEmptyResponse: true},
			wantErr: true,
			want:    nil,
		},
		{
			name: "empty body without DisallowEmptyResponse",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: false,
			want:    &TestMessage{Name: "", Value: 0},
		},
		{
			name: "non-2xx status code without InspectResponseError",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			},
			opts:    whttp.DecodeOptions{InspectResponseError: false},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields with DisallowUnknownFields",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{DisallowUnknownFields: true},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields without DisallowUnknownFields",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{DisallowUnknownFields: false},
			wantErr: false,
			want:    &TestMessage{Name: "test", Value: 123}, // "extra" field should be ignored
		},
		{
			name: "invalid JSON",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: true,
			want:    nil,
		},
		{
			name: "non-2xx status code with InspectResponseError",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"code": 500,
					"error": {
						"message": "(#131030) Recipient phone number not in allowed list",
						"type": "OAuthException",
						"code": 131030,
						"error_data": {
							"messaging_product": "whatsapp",
							"details": "Recipient phone number not in allowed list"
						},
						"error_subcode": 2655007,
						"error_user_title": "Recipient phone number not in allowed list",
						"error_user_msg": "Add recipient phone number to recipient list and try again.",
						"fbtrace_id": "AI5Ob2z72R0JAUB5zOF-nao"
					}
				}`)),
			},
			opts:    whttp.DecodeOptions{InspectResponseError: true},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got TestMessage
			err := whttp.DecodeResponseJSON[TestMessage](tt.response, &got, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeResponseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if diff := gcmp.Diff(tt.want, &got); diff != "" {
					t.Errorf("DecodeResponseJSON() mismatch (-want +got):\n%s", diff)
				}
			}

			if tt.wantErr && tt.opts.InspectResponseError {
				var respErr *whttp.ResponseError
				if !errors.As(err, &respErr) {
					t.Errorf("expected error of type ResponseError, got %T", err)
				}
			}
		})
	}
}

func TestDecodeRequestJSON(t *testing.T) {
	t.Parallel()
	type Data struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		request *http.Request
		opts    whttp.DecodeOptions
		wantErr bool
		want    *Data
	}{
		{
			name: "successful decoding",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123}`)),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: false,
			want:    &Data{Name: "test", Value: 123},
		},
		{
			name: "empty body with DisallowEmptyResponse",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString("")),
			},
			opts:    whttp.DecodeOptions{DisallowEmptyResponse: true},
			wantErr: true,
			want:    nil,
		},
		{
			name: "empty body without DisallowEmptyResponse",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString("")),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: false,
			want:    &Data{Name: "", Value: 0},
		},
		{
			name: "unknown fields with DisallowUnknownFields",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{DisallowUnknownFields: true},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields without DisallowUnknownFields",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{DisallowUnknownFields: false},
			wantErr: false,
			want:    &Data{Name: "test", Value: 123}, // "extra" field should be ignored
		},
		{
			name: "invalid JSON",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`invalid json`)),
			},
			opts:    whttp.DecodeOptions{},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Data
			err := whttp.DecodeRequestJSON[Data](tt.request, &got, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeRequestJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if diff := gcmp.Diff(tt.want, &got); diff != "" {
					t.Errorf("DecodeRequestJSON() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestRequestWithContext(t *testing.T) {
	t.Parallel()
	type testCase[T any] struct {
		name    string
		req     *whttp.Request[T]
		want    *http.Request
		wantErr bool
	}

	messageTests := []testCase[TestMessage]{
		{
			name: "successful POST request with JSON body",
			req: &whttp.Request[TestMessage]{
				Type:        whttp.RequestTypeSendMessage,
				Method:      http.MethodPost,
				Bearer:      "test-token",
				BaseURL:     "http://localhost/api/v1",
				Endpoints:   []string{"send-message"},
				Message:     &TestMessage{Name: "Test", Value: 123},
				Headers:     map[string]string{"X-Custom-Header": "custom-value"},
				QueryParams: map[string]string{"param1": "value1", "param2": "value2"},
			},
			want: &http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Authorization":   []string{"Bearer test-token"},
					"X-Custom-Header": []string{"custom-value"},
					"Content-Type":    []string{"application/json"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/send-message?param1=value1&param2=value2")

					return u
				}(),
				Host: "localhost",
				Body: io.NopCloser(bytes.NewReader([]byte(`{"name":"Test","value":123}`))),
			},
			wantErr: false,
		},
		{
			name: "successful GET request with no body",
			req: &whttp.Request[TestMessage]{
				Type:        whttp.RequestTypeGetPhoneNumber,
				Method:      http.MethodGet,
				Bearer:      "test-token",
				BaseURL:     "http://localhost/api/v1",
				Endpoints:   []string{"phone-number"},
				QueryParams: map[string]string{"param": "value"},
			},
			want: &http.Request{
				Method: http.MethodGet,
				Header: http.Header{
					"Authorization": []string{"Bearer test-token"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/phone-number?param=value")

					return u
				}(),
				Host: "localhost",
				Body: nil,
			},
			wantErr: false,
		},
		{
			name: "successful multipart form request with file",
			req: &whttp.Request[TestMessage]{
				Type:      whttp.RequestTypeUploadMedia,
				Method:    http.MethodPost,
				Bearer:    "test-token",
				BaseURL:   "http://localhost/api/v1",
				Endpoints: []string{"upload"},
				Form: &whttp.RequestForm{
					Fields: map[string]string{
						"title": "example",
					},
					FormFile: &whttp.FormFile{
						Name: "file",
						Path: "testfile.txt",
					},
				},
				Headers: map[string]string{"X-Custom-Header": "custom-value"},
			},
			want: &http.Request{
				Method: http.MethodPost,
				Header: http.Header{
					"Authorization":   []string{"Bearer test-token"},
					"X-Custom-Header": []string{"custom-value"},
				},
				URL: func() *url.URL {
					u, _ := url.Parse("http://localhost/api/v1/upload")

					return u
				}(),
				Host: "localhost",
			},
			wantErr: false,
		},
	}

	for _, tt := range messageTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := whttp.RequestWithContext(t.Context(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestWithContext() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if diff := gcmp.Diff(got.URL.String(), tt.want.URL.String()); diff != "" {
				t.Errorf("RequestWithContext() URL mismatch (-want +got):\n%s", diff)
			}
			if diff := gcmp.Diff(got.Method, tt.want.Method); diff != "" {
				t.Errorf("RequestWithContext() Method mismatch (-want +got):\n%s", diff)
			}
			for key := range tt.want.Header {
				if got.Header.Get(key) != tt.want.Header.Get(key) {
					t.Errorf("RequestWithContext() header %v mismatch, want %v,"+
						" got %v", key, tt.want.Header.Get(key), got.Header.Get(key))
				}
			}

			if tt.want.Body != nil {
				var gotBody TestMessage
				var wantBody TestMessage

				gotBodyBytes, _ := io.ReadAll(got.Body)
				json.Unmarshal(gotBodyBytes, &gotBody)

				wantBodyBytes, _ := json.Marshal(tt.req.Message)
				json.Unmarshal(wantBodyBytes, &wantBody)

				if diff := gcmp.Diff(gotBody, wantBody); diff != "" {
					t.Errorf("RequestWithContext() Body mismatch (-want +got):\n%s", diff)
				}
			}

			if strings.Contains(got.Header.Get("Content-Type"), "multipart/form-data") {
				err := got.ParseMultipartForm(10 << 20) // 10 MB
				if err != nil {
					t.Fatalf("failed to parse multipart form: %v", err)
				}

				for key, expectedVal := range tt.req.Form.Fields {
					if gotVal := got.FormValue(key); gotVal != expectedVal {
						t.Errorf("expected form field %s = %v, got %v", key, expectedVal, gotVal)
					}
				}

				file, _, err := got.FormFile(tt.req.Form.FormFile.Name)
				if err != nil {
					t.Fatalf("failed to get form file: %v", err)
				}
				defer file.Close()

				fileContent, _ := io.ReadAll(file)
				expectedFileContent, _ := os.ReadFile(tt.req.Form.FormFile.Path)
				if string(fileContent) != string(expectedFileContent) {
					t.Errorf("expected file content = %v, got %v", string(expectedFileContent), string(fileContent))
				}
			}
		})
	}
}

func TestBodyReaderResponseDecoder(t *testing.T) {
	t.Parallel()
	filePath := "testfile.txt"

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	tests := []struct {
		name         string
		response     *http.Response
		fn           whttp.ResponseBodyReaderFunc
		expectedBody string
		expectErr    bool
	}{
		{
			name: "successful JSON body read",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"name": "John", "age": 30}`)),
			},
			fn: func(_ context.Context, reader io.Reader) error {
				var result map[string]interface{}
				err := json.NewDecoder(reader).Decode(&result)
				if err != nil {
					return fmt.Errorf("failed to decode JSON: %w", err)
				}

				if result["name"] != "John" || result["age"] != float64(30) {
					t.Errorf("expected name=John and age=30, got name=%v and age=%v", result["name"], result["age"])
				}

				return nil
			},
			expectedBody: `{"name": "John", "age": 30}`,
			expectErr:    false,
		},
		{
			name: "successful plain text body read",
			response: &http.Response{
				Body: io.NopCloser(strings.NewReader("Hello, World!")),
			},
			fn: func(_ context.Context, reader io.Reader) error {
				bodyBytes, err := io.ReadAll(reader)
				if err != nil {
					return fmt.Errorf("failed to read body: %w", err)
				}
				if string(bodyBytes) != "Hello, World!" {
					t.Errorf("expected body 'Hello, World!', got %s", string(bodyBytes))
				}

				return nil
			},
			expectedBody: "Hello, World!",
			expectErr:    false,
		},
		{
			name: "successful file content read",
			response: &http.Response{
				Body: io.NopCloser(bytes.NewReader(fileContent)),
			},
			fn: func(_ context.Context, reader io.Reader) error {
				bodyBytes, err := io.ReadAll(reader)
				if err != nil {
					return fmt.Errorf("failed to read file content: %w", err)
				}

				if string(bodyBytes) != string(fileContent) {
					t.Errorf("expected file content %q, got %q", "Hello World", string(bodyBytes))
				}

				return nil
			},
			expectedBody: "Hello World",
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			decoder := whttp.BodyReaderResponseDecoder(tt.fn)
			err := decoder(t.Context(), tt.response)

			if (err != nil) != tt.expectErr {
				t.Errorf("BodyReaderResponseDecoder() error = %v, expectErr %v", err, tt.expectErr)

				return
			}

			if tt.response.Body != nil {
				reReadBody, err := io.ReadAll(tt.response.Body)
				if err != nil {
					t.Fatalf("failed to re-read response body: %v", err)
				}
				if string(reReadBody) != tt.expectedBody {
					t.Errorf("expected re-read body = %v, got %v", tt.expectedBody, string(reReadBody))
				}
			}
		})
	}
}

func ExampleNewSender() {
	customHTTPClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Define middleware that logs request information
	loggingMiddleware := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Println("request logger init called before core middlewares")

			return next(ctx, req, decoder)
		}
	}

	methodPrinter := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Printf("from core middleware request method is: %s\n", req.Method)

			return next(ctx, req, decoder)
		}
	}

	loggingMiddleware2 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			fmt.Println("called after core middleware request logger final")

			return next(ctx, req, decoder)
		}
	}

	loggingMiddleware3 := func(next whttp.SenderFunc[TestMessage]) whttp.SenderFunc[TestMessage] {
		return func(ctx context.Context, req *whttp.Request[TestMessage], decoder whttp.ResponseDecoder) error {
			err := next(ctx, req, decoder)
			fmt.Println("called after request send execution and the err is:", err)

			return err
		}
	}

	reqInterceptor := func(_ context.Context, req *http.Request) error {
		fmt.Println("Just intercepted the request and the method is:", req.Method)

		return nil
	}

	resInterceptor := func(_ context.Context, res *http.Response) error {
		fmt.Println("Just intercepted the response and status code:", res.StatusCode)

		return nil
	}

	resBodyInterceptor := func(_ context.Context, res *http.Response) error {
		body, _ := io.ReadAll(res.Body)
		defer func() {
			res.Body = io.NopCloser(bytes.NewReader(body))
		}()
		fmt.Println("from response body interceptor:", string(body))

		return nil
	}

	options := []whttp.CoreClientOption[TestMessage]{
		whttp.WithCoreClientHTTPClient[TestMessage](customHTTPClient),
		whttp.WithCoreClientMiddlewares[TestMessage](methodPrinter),
		whttp.WithCoreClientRequestInterceptor[TestMessage](nil),
		whttp.WithCoreClientResponseInterceptor[TestMessage](resBodyInterceptor),
	}

	sender := whttp.NewSender(
		options...,
	)

	longLiveClient := &http.Client{Timeout: time.Hour}

	sender.PrependMiddlewares(loggingMiddleware) // they will be executed first before previous set middlewares if any
	sender.AppendMiddlewares(loggingMiddleware2, loggingMiddleware3)
	sender.SetHTTPClient(longLiveClient)
	sender.SetRequestInterceptor(reqInterceptor)  // overrides the default interceptors
	sender.SetResponseInterceptor(resInterceptor) // overrides the default interceptors

	echoHandler := func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read body", http.StatusInternalServerError)

			return
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}

	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer server.Close()

	req := &whttp.Request[TestMessage]{
		Method:    http.MethodPost,
		BaseURL:   server.URL,
		Endpoints: []string{"send"},
		Message:   &TestMessage{Name: "Hello", Value: 123},
		Bearer:    "example-token",
	}

	// Response decoder to handle the response
	decoder := whttp.ResponseDecoderJSON(&TestMessage{}, whttp.DecodeOptions{})

	// Send the request using the sender
	err := sender.Send(context.Background(), req, decoder)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}

	// Output:
	// request logger init called before core middlewares
	// from core middleware request method is: POST
	// called after core middleware request logger final
	// Just intercepted the request and the method is: POST
	// Just intercepted the response and status code: 200
	// called after request send execution and the err is: <nil>
}
