/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
)

// ErrMissingFormFile is returned by EncodePayload when a *RequestForm is provided
// without a FormFile.
var ErrMissingFormFile = errors.New("missing form file")

// EncodeResponse holds the encoded body as an io.Reader together with its
// Content-Type. It is returned by EncodePayload.
type EncodeResponse struct {
	Body        io.Reader
	ContentType string
}

// PayloadEncoder allows custom types to provide their own encoding logic.
// When a payload implements this interface, EncodePayload delegates to it
// instead of using the built-in type switch.
type PayloadEncoder interface {
	EncodePayload() (*EncodeResponse, error)
}

// EncodePayload encodes payload into an EncodeResponse ready to be sent as an
// HTTP request body. Dispatch order:
//  1. PayloadEncoder interface — if payload implements PayloadEncoder, its
//     EncodePayload method is called.
//  2. Built-in types in this order: *RequestForm, io.Reader, []byte, string.
//  3. Default — any other value is JSON-marshalled.
//
// A nil payload returns an EncodeResponse with a nil Body and
// "application/json" Content-Type.
//
// Built-in type handling:
//   - *RequestForm — encoded as multipart/form-data (returns ErrMissingFormFile
//     when FormFile is nil).
//   - io.Reader — used directly as the body with "application/octet-stream".
//   - []byte — wrapped in a bytes.Reader with "application/octet-stream".
//   - string — wrapped in a bytes.Reader with "text/plain".
//   - any JSON-marshallable struct — marshalled with "application/json".
func EncodePayload(payload any) (*EncodeResponse, error) {
	if payload == nil {
		return &EncodeResponse{
			Body:        nil,
			ContentType: "application/json",
		}, nil
	}

	if p, ok := payload.(PayloadEncoder); ok {
		resp, err := p.EncodePayload()
		if err != nil {
			return nil, fmt.Errorf("encode payload: encoder: %T: error: %w", p, err)
		}

		return resp, nil
	}

	switch p := payload.(type) {
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
			Body:        bytes.NewReader([]byte(p)),
			ContentType: "text/plain",
		}, nil
	default:
		data, err := json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload as JSON: %w", err)
		}

		return &EncodeResponse{
			Body:        bytes.NewReader(data),
			ContentType: "application/json",
		}, nil
	}
}

var errPanicMessage = errors.New("panic")

// encodeFormData packages the provided form fields and file into a multipart/form-data stream.
// It uses an io.Pipe to stream data concurrently, avoiding loading the entire file into memory.
//
// Validation for file existence happens eagerly before returning. Any errors encountered
// mid-stream (e.g., disk I/O or network failures) are propagated asynchronously to the
// consumer via the returned io.Reader's Read method.
func encodeFormData( //nolint:gocognit // complexity is OK
	form *RequestForm,
) (io.Reader, string, error) {
	if form == nil || form.FormFile == nil {
		return nil, "", ErrMissingFormFile
	}

	// Eagerly validate that the target file exists before initializing the pipeline.
	if _, err := os.Stat(form.FormFile.Path); err != nil {
		return nil, "", fmt.Errorf("error opening file %q: %w", form.FormFile.Path, err)
	}

	pr, pw := io.Pipe()
	multipartWriter := multipart.NewWriter(pw)

	go func() {
		var writeErr error

		// Ensures the pipe is always closed and captures panics or streaming errors,
		// routing them directly to the reader side of the pipe.
		defer func() {
			if r := recover(); r != nil {
				pw.CloseWithError(fmt.Errorf("%w: encoding form data: %v", errPanicMessage, r))
			} else if writeErr != nil {
				pw.CloseWithError(writeErr)
			} else {
				_ = multipartWriter.Close()
				_ = pw.Close()
			}
		}()

		// Write standard key/value metadata fields.
		for key, value := range form.Fields {
			if err := multipartWriter.WriteField(key, value); err != nil {
				writeErr = fmt.Errorf("error writing field %q: %w", key, err)
				return
			}
		}

		file, err := os.Open(form.FormFile.Path)
		if err != nil {
			writeErr = fmt.Errorf("error opening file %q: %w", form.FormFile.Path, err)
			return
		}
		defer file.Close()

		actualFileName := filepath.Base(form.FormFile.Path)
		contentType := form.FormFile.Type
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		h := make(textproto.MIMEHeader)
		disp := mime.FormatMediaType("form-data", map[string]string{
			"name":     form.FormFile.Name,
			"filename": actualFileName,
		})
		h.Set("Content-Disposition", disp)
		h.Set("Content-Type", contentType)

		partWriter, err := multipartWriter.CreatePart(h)
		if err != nil {
			writeErr = fmt.Errorf("error creating form file part for field %q: %w", form.FormFile.Name, err)
			return
		}

		// Stream the file payload into the multipart part writer.
		if _, err = io.Copy(partWriter, file); err != nil {
			writeErr = fmt.Errorf("error copying file content for field %q: %w", form.FormFile.Name, err)
			return
		}
	}()

	return pr, multipartWriter.FormDataContentType(), nil
}
