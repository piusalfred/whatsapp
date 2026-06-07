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
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/internal/test"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestEncodePayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		payload         any
		wantContentType string
		wantErr         bool
		validateBody    func(t *testing.T, body io.Reader)
	}{
		{
			name:            "nil payload",
			payload:         nil,
			wantContentType: "application/json",
			wantErr:         false,
			validateBody: func(t *testing.T, body io.Reader) {
				t.Helper()
				if body != nil {
					t.Error("expected nil body for nil payload")
				}
			},
		},
		{
			name: "JSON struct payload",
			payload: TestMessage{
				Name:  "test",
				Value: 123,
			},
			wantContentType: "application/json",
			wantErr:         false,
			validateBody: func(t *testing.T, body io.Reader) {
				t.Helper()
				data, err := io.ReadAll(body)
				test.AssertNoError(t, "read encoded body", err)

				var msg TestMessage
				err = json.Unmarshal(data, &msg)
				test.AssertNoError(t, "unmarshal JSON", err)

				if msg.Name != "test" || msg.Value != 123 {
					t.Errorf("expected Name=test and Value=123, got Name=%s and Value=%d", msg.Name, msg.Value)
				}
			},
		},
		{
			name:            "io.Reader payload",
			payload:         strings.NewReader("test data"),
			wantContentType: "application/octet-stream",
			wantErr:         false,
			validateBody: func(t *testing.T, body io.Reader) {
				t.Helper()
				data, err := io.ReadAll(body)
				test.AssertNoError(t, "read body", err)
				if string(data) != "test data" {
					t.Errorf("expected 'test data', got %s", string(data))
				}
			},
		},
		{
			name:            "byte slice payload",
			payload:         []byte("test bytes"),
			wantContentType: "application/octet-stream",
			wantErr:         false,
			validateBody: func(t *testing.T, body io.Reader) {
				t.Helper()
				data, err := io.ReadAll(body)
				test.AssertNoError(t, "read body", err)
				if string(data) != "test bytes" {
					t.Errorf("expected 'test bytes', got %s", string(data))
				}
			},
		},
		{
			name:            "string payload",
			payload:         "test string",
			wantContentType: "text/plain",
			wantErr:         false,
			validateBody: func(t *testing.T, body io.Reader) {
				t.Helper()
				data, err := io.ReadAll(body)
				test.AssertNoError(t, "read body", err)
				if string(data) != "test string" {
					t.Errorf("expected 'test string', got %s", string(data))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := whttp.EncodePayload(tt.payload)

			if tt.wantErr {
				test.AssertError(t, "EncodePayload should return error", err)
				return
			}

			test.AssertNoError(t, "EncodePayload", err)

			if !strings.Contains(resp.ContentType, tt.wantContentType) {
				t.Errorf("expected ContentType to contain %s, got %s", tt.wantContentType, resp.ContentType)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, resp.Body)
			}
		})
	}
}

func TestEncodePayload_NoTrailingNewline(t *testing.T) {
	t.Parallel()

	resp, err := whttp.EncodePayload(TestMessage{Name: "test", Value: 123})
	test.AssertNoError(t, "EncodePayload", err)

	data, err := io.ReadAll(resp.Body)
	test.AssertNoError(t, "read body", err)

	if len(data) > 0 && data[len(data)-1] == '\n' {
		t.Errorf("expected no trailing newline, got %q", string(data))
	}
}

func TestEncodeFormData(t *testing.T) {
	t.Parallel()

	t.Run("form with file", func(t *testing.T) {
		t.Parallel()

		form := &whttp.RequestForm{
			Fields: map[string]string{
				"field1": "value1",
				"field2": "value2",
			},
			FormFile: &whttp.FormFile{
				Name: "testfile",
				Path: "testfile.txt",
				Type: "text/plain",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		if !strings.Contains(encoded.ContentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", encoded.ContentType)
		}

		if encoded.Body == nil {
			t.Error("expected non-nil body")
		}
	})
}

func TestEncodePayload_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid JSON encoding", func(t *testing.T) {
		t.Parallel()

		type invalidStruct struct {
			Ch chan int
		}

		payload := invalidStruct{Ch: make(chan int)}
		_, err := whttp.EncodePayload(payload)

		test.AssertError(t, "should error on invalid JSON", err)
	})
}

func TestEncodeFormData_Errors(t *testing.T) {
	t.Parallel()

	t.Run("file not found error", func(t *testing.T) {
		t.Parallel()

		form := &whttp.RequestForm{
			Fields: map[string]string{"field": "value"},
			FormFile: &whttp.FormFile{
				Name: "file",
				Path: "/nonexistent/file/path/that/does/not/exist.txt",
				Type: "text/plain",
			},
		}

		_, err := whttp.EncodePayload(form)
		test.AssertError(t, "should error on missing file", err)
	})

	t.Run("missing form file", func(t *testing.T) {
		t.Parallel()

		form := &whttp.RequestForm{
			Fields: map[string]string{"field": "value"},
		}

		_, err := whttp.EncodePayload(form)
		test.AssertError(t, "should error on missing form file", err)
		test.AssertErrorIs(t, "should be ErrMissingFormFile", err, whttp.ErrMissingFormFile)
	})
}

func TestEncodeFormData_WithTempFile(t *testing.T) {
	t.Parallel()

	t.Run("form with temp file default content type", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "upload.bin")

		err := os.WriteFile(tmpFile, []byte("binary content"), 0o644)
		test.AssertNoError(t, "write temp file", err)

		form := &whttp.RequestForm{
			Fields: map[string]string{
				"description": "test upload",
			},
			FormFile: &whttp.FormFile{
				Name: "file",
				Path: tmpFile,
				Type: "",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		if !strings.Contains(encoded.ContentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", encoded.ContentType)
		}

		body, err := io.ReadAll(encoded.Body)
		test.AssertNoError(t, "read body", err)

		if !strings.Contains(string(body), "application/octet-stream") {
			t.Error("expected default content type application/octet-stream in form body")
		}
	})

	t.Run("form with special characters in field name", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "testfile.txt")

		err := os.WriteFile(tmpFile, []byte("test content"), 0o644)
		test.AssertNoError(t, "write temp file", err)

		form := &whttp.RequestForm{
			Fields: map[string]string{
				"field_with_quotes": "value",
			},
			FormFile: &whttp.FormFile{
				Name: "file_field",
				Path: tmpFile,
				Type: "text/plain",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		if encoded.Body == nil {
			t.Error("expected non-nil body")
		}
	})

	t.Run("form with multiple fields", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "multifield.txt")

		err := os.WriteFile(tmpFile, []byte("file content for multi-field test"), 0o644)
		test.AssertNoError(t, "write temp file", err)

		form := &whttp.RequestForm{
			Fields: map[string]string{
				"field1": "value1",
				"field2": "value2",
				"field3": "value3",
			},
			FormFile: &whttp.FormFile{
				Name: "document",
				Path: tmpFile,
				Type: "text/plain",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		body, err := io.ReadAll(encoded.Body)
		test.AssertNoError(t, "read body", err)

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "value1") || !strings.Contains(bodyStr, "value2") ||
			!strings.Contains(bodyStr, "value3") {
			t.Error("expected all field values in the encoded body")
		}
	})

	t.Run("form with image content type", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "image.png")

		err := os.WriteFile(tmpFile, []byte("fake png content"), 0o644)
		test.AssertNoError(t, "write temp file", err)

		form := &whttp.RequestForm{
			Fields: map[string]string{},
			FormFile: &whttp.FormFile{
				Name: "image",
				Path: tmpFile,
				Type: "image/png",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		body, err := io.ReadAll(encoded.Body)
		test.AssertNoError(t, "read body", err)

		if !strings.Contains(string(body), "image/png") {
			t.Error("expected image/png content type in form body")
		}
	})
}

func TestEncodeFormData_MultipartStructure(t *testing.T) {
	t.Parallel()

	t.Run("parseable multipart body", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "data.txt")

		err := os.WriteFile(tmpFile, []byte("hello world"), 0o644)
		test.AssertNoError(t, "write temp file", err)

		form := &whttp.RequestForm{
			Fields: map[string]string{
				"caption": "my upload",
				"meta":    "123",
			},
			FormFile: &whttp.FormFile{
				Name: "media",
				Path: tmpFile,
				Type: "text/plain",
			},
		}

		encoded, err := whttp.EncodePayload(form)
		test.AssertNoError(t, "EncodePayload", err)

		body, err := io.ReadAll(encoded.Body)
		test.AssertNoError(t, "read body", err)

		// Extract boundary from Content-Type header.
		contentType := encoded.ContentType
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Fatalf("expected multipart/form-data, got %s", contentType)
		}

		_, params, err := mime.ParseMediaType(contentType)
		test.AssertNoError(t, "parse media type", err)

		boundary, ok := params["boundary"]
		if !ok || boundary == "" {
			t.Fatal("expected boundary in content type")
		}

		reader := multipart.NewReader(bytes.NewReader(body), boundary)

		foundFields := make(map[string]string)
		var fileData []byte

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			test.AssertNoError(t, "next part", err)

			data, err := io.ReadAll(part)
			test.AssertNoError(t, "read part", err)

			if part.FileName() != "" {
				fileData = data
				continue
			}

			name := part.FormName()
			foundFields[name] = string(data)
		}

		if len(foundFields) != 2 {
			t.Errorf("expected 2 form fields, got %d", len(foundFields))
		}
		if foundFields["caption"] != "my upload" {
			t.Errorf("expected caption='my upload', got %q", foundFields["caption"])
		}
		if foundFields["meta"] != "123" {
			t.Errorf("expected meta='123', got %q", foundFields["meta"])
		}

		if fileData == nil {
			t.Fatal("expected file part in multipart body")
		}
		if string(fileData) != "hello world" {
			t.Errorf("expected file content 'hello world', got %q", string(fileData))
		}
	})
}

func TestEncodeFormData_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.bin")

	err := os.WriteFile(tmpFile, []byte{}, 0o644)
	test.AssertNoError(t, "write empty file", err)

	form := &whttp.RequestForm{
		Fields: map[string]string{"note": "empty"},
		FormFile: &whttp.FormFile{
			Name: "file",
			Path: tmpFile,
			Type: "application/octet-stream",
		},
	}

	encoded, err := whttp.EncodePayload(form)
	test.AssertNoError(t, "EncodePayload", err)

	body, err := io.ReadAll(encoded.Body)
	test.AssertNoError(t, "read body", err)

	if !strings.Contains(string(body), "note") || !strings.Contains(string(body), "empty") {
		t.Error("expected form field in body")
	}
}

func TestEncodeFormData_SpecialFilename(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file with spaces & symbols.txt")

	err := os.WriteFile(tmpFile, []byte("special"), 0o644)
	test.AssertNoError(t, "write temp file", err)

	form := &whttp.RequestForm{
		Fields: map[string]string{},
		FormFile: &whttp.FormFile{
			Name: "upload",
			Path: tmpFile,
			Type: "text/plain",
		},
	}

	encoded, err := whttp.EncodePayload(form)
	test.AssertNoError(t, "EncodePayload", err)

	body, err := io.ReadAll(encoded.Body)
	test.AssertNoError(t, "read body", err)

	// The filename should appear in the Content-Disposition header.
	if !strings.Contains(string(body), `filename="file with spaces & symbols.txt"`) {
		t.Error("expected quoted filename in multipart body")
	}
}

func TestEncodePayload_PayloadEncoder(t *testing.T) {
	t.Parallel()

	custom := &customEncoder{
		data:        []byte("custom-encoded"),
		contentType: "application/x-custom",
	}

	resp, err := whttp.EncodePayload(custom)
	test.AssertNoError(t, "EncodePayload", err)

	if resp.ContentType != "application/x-custom" {
		t.Errorf("expected content type application/x-custom, got %s", resp.ContentType)
	}

	data, err := io.ReadAll(resp.Body)
	test.AssertNoError(t, "read body", err)

	if string(data) != "custom-encoded" {
		t.Errorf("expected body 'custom-encoded', got %q", string(data))
	}
}

func TestEncodePayload_PayloadEncoderError(t *testing.T) {
	t.Parallel()

	custom := &customEncoder{err: os.ErrNotExist}
	_, err := whttp.EncodePayload(custom)
	test.AssertError(t, "EncodePayload should return error", err)
}

type customEncoder struct {
	data        []byte
	contentType string
	err         error
}

func (c *customEncoder) EncodePayload() (*whttp.EncodeResponse, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &whttp.EncodeResponse{
		Body:        bytes.NewReader(c.data),
		ContentType: c.contentType,
	}, nil
}
