package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

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

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"") //nolint:gochecknoglobals // this is a global variable

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// encodeFormData encodes form fields and file data into multipart/form-data.
func encodeFormData(form *RequestForm) (io.Reader, string, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	for key, value := range form.Fields {
		err := multipartWriter.WriteField(key, value)
		if err != nil {
			return nil, "", fmt.Errorf("error writing field '%s': %w", key, err)
		}
	}

	file, err := os.Open(form.FormFile.Path)
	if err != nil {
		return nil, "", fmt.Errorf("error opening file '%s': %w", form.FormFile.Path, err)
	}
	defer file.Close()

	h := make(textproto.MIMEHeader)
	actualFileName := filepath.Base(form.FormFile.Path)

	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(form.FormFile.Name), escapeQuotes(actualFileName)))

	if form.FormFile.Type != "" {
		h.Set("Content-Type", form.FormFile.Type)
	} else {
		h.Set("Content-Type", "application/octet-stream")
	}

	partWriter, err := multipartWriter.CreatePart(h)
	if err != nil {
		return nil, "", fmt.Errorf("error creating form file part for field '%s': %w", form.FormFile.Name, err)
	}

	_, err = io.Copy(partWriter, file)
	if err != nil {
		return nil, "", fmt.Errorf("error copying file content for field '%s': %w", form.FormFile.Name, err)
	}

	err = multipartWriter.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	return &requestBody, multipartWriter.FormDataContentType(), nil
}
