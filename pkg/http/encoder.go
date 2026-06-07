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

var ErrMissingFormFile = errors.New("missing form file")

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

// EncodePayload takes different types of payloads (form data, readers, JSON) and returns an EncodeResponse.
func EncodePayload(payload any) (*EncodeResponse, error) {
	if payload == nil {
		return &EncodeResponse{
			Body:        nil,
			ContentType: "application/json",
		}, nil
	}

	if p, ok := payload.(PayloadEncoder); ok {
		return p.EncodePayload()
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

// encodeFormData encodes form fields and file data into multipart/form-data.
// It streams the data via io.Pipe to avoid buffering the entire file in memory.
func encodeFormData(form *RequestForm) (io.Reader, string, error) {
	if form.FormFile == nil {
		return nil, "", ErrMissingFormFile
	}

	// Eager validation: ensure file exists before streaming so common errors
	// (e.g. file not found) are surfaced immediately.
	if _, err := os.Stat(form.FormFile.Path); err != nil {
		return nil, "", fmt.Errorf("error opening file '%s': %w", form.FormFile.Path, err)
	}

	pr, pw := io.Pipe()
	multipartWriter := multipart.NewWriter(pw)

	go func() {
		var writeErr error

		defer func() {
			if r := recover(); r != nil {
				pw.CloseWithError(fmt.Errorf("panic encoding form data: %v", r))
			} else if writeErr != nil {
				pw.CloseWithError(writeErr)
			} else {
				pw.Close()
			}
		}()

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
			_ = file.Close()
			writeErr = fmt.Errorf("error creating form file part for field %q: %w", form.FormFile.Name, err)
			return
		}

		_, err = io.Copy(partWriter, file)
		if err != nil {
			_ = file.Close()
			writeErr = fmt.Errorf("error copying file content for field %q: %w", form.FormFile.Name, err)
			return
		}

		if err := file.Close(); err != nil {
			writeErr = fmt.Errorf("error closing file %q: %w", form.FormFile.Path, err)
			return
		}

		if err := multipartWriter.Close(); err != nil {
			writeErr = fmt.Errorf("error closing multipart writer: %w", err)
			return
		}
	}()

	return pr, multipartWriter.FormDataContentType(), nil
}
