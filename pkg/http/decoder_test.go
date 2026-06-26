//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
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
	"reflect"
	"strings"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

var errReaderFunction = errors.New("reader function error")

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
			opts:    whttp.DecodeOptions{Flags: whttp.JSONDecodeDisallowEmptyResponse},
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
			opts:    whttp.DecodeOptions{},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields with DisallowUnknownFields",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{Flags: whttp.JSONDecodeDisallowUnknownFields},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields without DisallowUnknownFields",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{},
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
			opts:    whttp.DecodeOptionsPermissive(),
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

			if tt.wantErr && tt.opts.Has(whttp.JSONDecodeInspectResponseError) {
				var respErr *whttp.ResponseError
				test.AssertErrorAs(t, "expected ResponseError", err, &respErr)
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
			opts:    whttp.DecodeOptions{Flags: whttp.JSONDecodeDisallowEmptyResponse},
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
			opts:    whttp.DecodeOptions{Flags: whttp.JSONDecodeDisallowUnknownFields},
			wantErr: true,
			want:    nil,
		},
		{
			name: "unknown fields without DisallowUnknownFields",
			request: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`{"name": "test", "value": 123, "extra": "unknown"}`)),
			},
			opts:    whttp.DecodeOptions{},
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
				var result map[string]any
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

		err := whttp.DecodeResponseJSON[TestMessage](resp, &msg, whttp.DecodeOptionsPermissive())

		test.AssertError(t, "should return error", err)

		var respErr *whttp.ResponseError
		test.AssertErrorAs(t, "expected ResponseError", err, &respErr)

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
			Flags: whttp.JSONDecodeDisallowEmptyResponse,
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
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{})

		test.AssertError(t, "should error on non-2xx status", err)
	})

	t.Run("error decoding error response", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`invalid json for error`)),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptionsPermissive())

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
			Flags: whttp.JSONDecodeDisallowEmptyResponse,
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

		expectedErr := errReaderFunction
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
	err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptionsPermissive())

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
	err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptionsPermissive())

	test.AssertError(t, "should error on invalid JSON in error response", err)
	test.AssertErrorIs(t, "should be ErrDecodeErrorResponse", err, whttp.ErrDecodeErrorResponse)
}

func TestDecodeRequestJSON_NilRequest(t *testing.T) {
	t.Parallel()

	var result TestMessage
	err := whttp.DecodeRequestJSON[TestMessage](nil, &result, whttp.DecodeOptions{})

	test.AssertError(t, "should error on nil request", err)
	test.AssertErrorIs(t, "should be ErrNilRequest", err, whttp.ErrNilRequest)
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
	err := whttp.DecodeRequestJSON[TestMessage](req, &result, whttp.DecodeOptions{})

	test.AssertNoError(t, "should allow empty body when not disallowed", err)
}

func TestDecodeRequestJSON_WithDisallowUnknownFields(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodPost, "http://example.com",
		strings.NewReader(`{"name":"test","value":123,"unknown":"field"}`))

	var result TestMessage
	err := whttp.DecodeRequestJSON[TestMessage](req, &result, whttp.DecodeOptions{
		Flags: whttp.JSONDecodeDisallowUnknownFields,
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

func TestDecodeResponseJSON_MaxBodyBytes(t *testing.T) {
	t.Parallel()

	t.Run("body within limit", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"name":"test","value":123}`)),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{
			MaxBodyBytes: 100,
		})

		test.AssertNoError(t, "should decode within limit", err)
	})

	t.Run("body exceeds limit", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"name":"test","value":123}`)),
		}

		var result TestMessage
		err := whttp.DecodeResponseJSON[TestMessage](resp, &result, whttp.DecodeOptions{
			MaxBodyBytes: 10,
		})

		test.AssertError(t, "should error when body exceeds limit", err)
		test.AssertErrorIs(t, "should be ErrBodyTooLarge", err, whttp.ErrBodyTooLarge)
	})
}

// mockDecoder is a simple spy/stub for testing the inner ResponseDecoder.
type mockDecoder struct {
	err          error
	receivedBody []byte
	called       bool
}

func (m *mockDecoder) Decode(ctx context.Context, response *http.Response) error {
	m.called = true
	if response.Body != nil {
		// Read the body to verify the capturer properly restored it
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("mock decoder read error: %w", err)
		}
		m.receivedBody = body
	}
	return m.err
}

var (
	errSimulatedRead      = errors.New("simulated read error")
	errInnerDecoderFailed = errors.New("inner decoder failed")
)

// errReader simulates a network failure during body reading.
type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errSimulatedRead
}

func (errReader) Close() error { return nil }

func TestResponseCapturer_Decode(t *testing.T) {
	tests := []struct {
		name         string
		response     *http.Response
		innerDecoder *mockDecoder
		wantErr      bool
		errContains  string
		wantBody     []byte
		wantStatus   int
		wantHeader   http.Header
	}{
		{
			name: "successful capture with inner decoder",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
			innerDecoder: &mockDecoder{},
			wantErr:      false,
			wantBody:     []byte(`{"key":"value"}`),
			wantStatus:   http.StatusOK,
			wantHeader:   http.Header{"Content-Type": []string{"application/json"}},
		},
		{
			name: "nil response body",
			response: &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     http.Header{"X-Custom": []string{"test"}},
				Body:       nil,
			},
			innerDecoder: &mockDecoder{},
			wantErr:      false,
			wantBody:     nil,
			wantStatus:   http.StatusNoContent,
			wantHeader:   http.Header{"X-Custom": []string{"test"}},
		},
		{
			name: "nil inner decoder",
			response: &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{},
				Body:       io.NopCloser(bytes.NewReader([]byte("data"))),
			},
			innerDecoder: nil, // Tests c.inner != nil check
			wantErr:      false,
			wantBody:     []byte("data"),
			wantStatus:   http.StatusAccepted,
			wantHeader:   http.Header{},
		},
		{
			name: "inner decoder returns error",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{},
				Body:       io.NopCloser(bytes.NewReader([]byte("error details"))),
			},
			innerDecoder: &mockDecoder{err: errInnerDecoderFailed},
			wantErr:      true,
			errContains:  "inner decoder failed",
			wantBody:     []byte("error details"),
			wantStatus:   http.StatusBadRequest,
			wantHeader:   http.Header{},
		},
		{
			name: "body read failure",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{},
				Body:       errReader{}, // Will fail on io.ReadAll
			},
			innerDecoder: &mockDecoder{},
			wantErr:      true,
			errContains:  "capture response body",
			wantBody:     nil,
			wantStatus:   http.StatusInternalServerError,
			wantHeader:   http.Header{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inner whttp.ResponseDecoder
			if tt.innerDecoder != nil {
				inner = tt.innerDecoder
			}
			capturer := whttp.NewResponseCapturer(inner)

			err := capturer.Decode(context.Background(), tt.response)

			if tt.wantErr {
				test.AssertError(t, "Decode", err)
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Decode() error = %v, expected to contain %q", err, tt.errContains)
				}
			} else {
				test.AssertNoError(t, "Decode", err)
			}

			if tt.name == "body read failure" {
				return
			}

			if capturer.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %v, want %v", capturer.StatusCode, tt.wantStatus)
			}
			if !reflect.DeepEqual(capturer.Header, tt.wantHeader) {
				t.Errorf("Header = %v, want %v", capturer.Header, tt.wantHeader)
			}
			if !bytes.Equal(capturer.Body, tt.wantBody) {
				t.Errorf("Body = %q, want %q", capturer.Body, tt.wantBody)
			}

			if tt.innerDecoder != nil {
				if !tt.innerDecoder.called {
					t.Error("expected inner decoder to be called, but it was not")
				}
				if tt.response.Body != nil && !bytes.Equal(tt.innerDecoder.receivedBody, tt.wantBody) {
					t.Errorf("Inner decoder received body = %q, want %q (body not restored properly)",
						tt.innerDecoder.receivedBody, tt.wantBody)
				}
			}
		})
	}
}

func TestResponseCapturer_Reset(t *testing.T) {
	t.Parallel()

	capturer := whttp.NewResponseCapturer(nil)

	resp := &http.Response{
		StatusCode: http.StatusTeapot,
		Header:     http.Header{"Dirty-Header": []string{"true"}},
		Body:       io.NopCloser(strings.NewReader("dirty body")),
	}
	test.AssertNoError(t, "dirtying capturer", capturer.Decode(context.Background(), resp))

	if capturer.Body == nil || capturer.StatusCode == 0 || capturer.Header == nil {
		t.Fatal("expected dirty state before Reset")
	}

	capturer.Reset()

	if capturer.Body != nil {
		t.Errorf("Expected Body to be nil after Reset, got %q", capturer.Body)
	}
	if capturer.StatusCode != 0 {
		t.Errorf("Expected StatusCode to be 0 after Reset, got %d", capturer.StatusCode)
	}
	if capturer.Header != nil {
		t.Errorf("Expected Header to be nil after Reset, got %v", capturer.Header)
	}

	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"After-Reset": []string{"yes"}},
		Body:       io.NopCloser(strings.NewReader("fresh data")),
	}
	test.AssertNoError(t, "capturer should still work after Reset", capturer.Decode(context.Background(), resp2))
	if capturer.StatusCode != http.StatusOK {
		t.Errorf("StatusCode after reset = %d, want %d", capturer.StatusCode, http.StatusOK)
	}
	if !bytes.Equal(capturer.Body, []byte("fresh data")) {
		t.Errorf("Body after reset = %q, want %q", capturer.Body, "fresh data")
	}
}

func TestResponseCapturer_HeaderCloning(t *testing.T) {
	t.Parallel()

	originalHeader := http.Header{"X-Test": []string{"original"}}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     originalHeader,
		Body:       nil,
	}
	capturer := whttp.NewResponseCapturer(nil)

	test.AssertNoError(t, "HeaderCloning", capturer.Decode(context.Background(), resp))

	resp.Header.Set("X-Test", "mutated")

	if capturedVal := capturer.Header.Get("X-Test"); capturedVal != "original" {
		t.Errorf("capturer Header was affected by mutation. Got %q, want %q", capturedVal, "original")
	}
}
