//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

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
func TestResponseDecoderJSON(t *testing.T) {
	t.Parallel()

	t.Run("successful JSON decoding", func(t *testing.T) {
		t.Parallel()

		msg := &TestMessage{}
		decoder := whttp.ResponseDecoderJSON(msg, whttp.DecodeOptions{})

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"name":"test","value":42}`)),
		}

		err := decoder.Decode(context.Background(), resp)
		test.AssertNoError(t, "Decode", err)

		if msg.Name != "test" || msg.Value != 42 {
			t.Errorf("expected Name=test and Value=42, got Name=%s and Value=%d", msg.Name, msg.Value)
		}
	})

	t.Run("error on invalid JSON", func(t *testing.T) {
		t.Parallel()

		msg := &TestMessage{}
		decoder := whttp.ResponseDecoderJSON(msg, whttp.DecodeOptions{})

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`invalid json`)),
		}

		err := decoder.Decode(context.Background(), resp)
		test.AssertError(t, "Decode should fail on invalid JSON", err)
	})
}
func TestDecodeResponseJSON_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil response", func(t *testing.T) {
		t.Parallel()

		var msg TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](nil, &msg, whttp.DecodeOptions{})

		test.AssertError(t, "should error on nil response", err)
		test.AssertErrorIs(t, "should be ErrNilResponse", err, whttp.ErrNilResponse)
	})

	t.Run("nil target", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"name":"test"}`)),
		}

		err := whttp.DecodeResponseJSON[TestMessage](resp, nil, whttp.DecodeOptions{})

		test.AssertError(t, "should error on nil target", err)
		test.AssertErrorIs(t, "should be ErrNilTarget", err, whttp.ErrNilTarget)
	})

	t.Run("error response with ResponseError type", func(t *testing.T) {
		t.Parallel()

		var msg TestMessage
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body: io.NopCloser(strings.NewReader(`{
				"code": 400,
				"error": {
					"message": "Invalid request",
					"type": "ValidationError",
					"code": 400
				}
			}`)),
		}

		err := whttp.DecodeResponseJSON[TestMessage](resp, &msg, whttp.DecodeOptions{InspectResponseError: true})

		test.AssertError(t, "should return error", err)

		var respErr *whttp.ResponseError
		if !errors.As(err, &respErr) {
			t.Fatalf("expected error to be ResponseError type, got %T", err)
		}

		if respErr.Code != http.StatusBadRequest {
			t.Errorf("expected Code=%d, got %d", http.StatusBadRequest, respErr.Code)
		}
	})
}
func TestDecodeResponseJSON_MoreEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("response with empty body and DisallowEmptyResponse", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{
			DisallowEmptyResponse: true,
		})

		test.AssertError(t, "should error on empty body with DisallowEmptyResponse", err)
	})

	t.Run("non-2xx status without InspectResponseError", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"error": "not found"}`)),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{InspectResponseError: false})

		test.AssertError(t, "should error on non-2xx status", err)
	})

	t.Run("error decoding error response", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`invalid json for error`)),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{InspectResponseError: true})

		test.AssertError(t, "should error when can't decode error response", err)
	})
}
func TestDecodeRequestJSON_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		var result TestMessage
		err := whttp.DecodeRequestJSON[TestMessage](nil, &result, whttp.DecodeOptions{})

		test.AssertError(t, "should error on nil request", err)
	})

	t.Run("request with empty body and DisallowEmptyResponse", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{
			Body: io.NopCloser(bytes.NewReader([]byte{})),
		}

		var result TestMessage
		err := whttp.DecodeRequestJSON[TestMessage](req, &result, whttp.DecodeOptions{
			DisallowEmptyResponse: true,
		})

		test.AssertError(t, "should error on empty body with DisallowEmptyResponse", err)
	})

	t.Run("nil target", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{
			Body: io.NopCloser(strings.NewReader(`{"name":"test"}`)),
		}

		err := whttp.DecodeRequestJSON[TestMessage](req, nil, whttp.DecodeOptions{})

		test.AssertError(t, "should error on nil target", err)
		test.AssertErrorIs(t, "should be ErrNilTarget", err, whttp.ErrNilTarget)
	})
}
func TestBodyReaderResponseDecoder_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("response with empty body reader", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}

		bodyCalled := false
		fn := func(ctx context.Context, reader io.Reader) error {
			bodyCalled = true
			data, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			if len(data) != 0 {
				t.Errorf("expected empty body, got %d bytes", len(data))
			}
			return nil
		}

		decoder := whttp.BodyReaderResponseDecoder(fn)
		err := decoder(context.Background(), resp)

		test.AssertNoError(t, "should handle empty body", err)
		if !bodyCalled {
			t.Error("expected body reader function to be called")
		}
	})

	t.Run("body reader function returns error", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("test content")),
		}

		expectedErr := errors.New("reader function error")
		fn := func(ctx context.Context, reader io.Reader) error {
			return expectedErr
		}

		decoder := whttp.BodyReaderResponseDecoder(fn)
		err := decoder(context.Background(), resp)

		test.AssertError(t, "should return reader function error", err)
		test.AssertErrorIs(t, "should be the expected error", err, expectedErr)
	})
}

func TestDecodeResponseJSON_NilResponse(t *testing.T) {
	t.Parallel()

	var result TestMessage
	err := whttp.DecodeResponseJSON[TestMessage](nil, &result, whttp.DecodeOptions{})

	test.AssertError(t, "should error on nil response", err)
	test.AssertErrorIs(t, "should be ErrNilResponse", err, whttp.ErrNilResponse)
}

func TestDecodeResponseJSON_NilTarget(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"name":"test","value":123}`)),
	}

	err := whttp.DecodeResponseJSON[TestMessage](resp, nil, whttp.DecodeOptions{})

	test.AssertError(t, "should error on nil target", err)
	test.AssertErrorIs(t, "should be ErrNilTarget", err, whttp.ErrNilTarget)
}

func TestDecodeResponseJSON_InspectErrorWithEmptyBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	var result TestMessage
	err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	test.AssertError(t, "should error on non-2xx with empty body", err)
	test.AssertErrorIs(t, "should be ErrRequestFailure", err, whttp.ErrRequestFailure)
}

func TestDecodeResponseJSON_InspectErrorWithInvalidJSON(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("not valid json")),
	}

	var result TestMessage
	err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	test.AssertError(t, "should error on invalid JSON in error response", err)
	test.AssertErrorIs(t, "should be ErrDecodeErrorResponse", err, whttp.ErrDecodeErrorResponse)
}

func TestDecodeRequestJSON_NilRequest(t *testing.T) {
	t.Parallel()

	var result TestMessage
	err := whttp.DecodeRequestJSON[TestMessage](nil, &result, whttp.DecodeOptions{})

	test.AssertError(t, "should error on nil request", err)
	test.AssertErrorIs(t, "should be ErrNilResponse", err, whttp.ErrNilResponse)
}

func TestDecodeRequestJSON_NilTarget(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(`{"name":"test"}`))

	err := whttp.DecodeRequestJSON[TestMessage](req, nil, whttp.DecodeOptions{})

	test.AssertError(t, "should error on nil target", err)
	test.AssertErrorIs(t, "should be ErrNilTarget", err, whttp.ErrNilTarget)
}

func TestDecodeRequestJSON_EmptyBodyAllowed(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(""))

	var result TestMessage
	err := whttp.DecodeRequestJSON[TestMessage](req, &result, whttp.DecodeOptions{
		DisallowEmptyResponse: false,
	})

	test.AssertNoError(t, "should allow empty body when not disallowed", err)
}

func TestDecodeRequestJSON_WithDisallowUnknownFields(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodPost, "http://example.com",
		strings.NewReader(`{"name":"test","value":123,"unknown":"field"}`))

	var result TestMessage
	err := whttp.DecodeRequestJSON[TestMessage](req, &result, whttp.DecodeOptions{
		DisallowUnknownFields: true,
	})

	test.AssertError(t, "should error on unknown fields", err)
	test.AssertErrorIs(t, "should be ErrDecodeResponseBody", err, whttp.ErrDecodeResponseBody)
}

func TestResponseDecoderFunc_Decode(t *testing.T) {
	t.Parallel()

	decodeCalled := false
	decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
		decodeCalled = true
		return nil
	})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := decoder.Decode(context.Background(), resp)

	test.AssertNoError(t, "Decode", err)
	if !decodeCalled {
		t.Error("expected decoder function to be called")
	}
}
