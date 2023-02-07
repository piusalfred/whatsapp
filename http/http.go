package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/errors"
	"io"
	"net/http"
	"net/url"
)

// NewRequestWithContext creates a new *http.Request with context by using the
// RequestParams.
func NewRequestWithContext(ctx context.Context, params *RequestParams, payload []byte) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	//https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages
	requestURL, err := url.JoinPath(params.BaseURL, params.ApiVersion, params.SenderID, params.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to join url parts: %w", err)
	}

	if payload == nil {
		req, err = http.NewRequestWithContext(ctx, params.Method, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, params.Method, requestURL, bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}
	}

	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	if params.Bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", params.Bearer))
	}

	if len(params.Query) > 0 {
		query := req.URL.Query()
		for key, value := range params.Query {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

func Send(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	if req, err = NewRequestWithContext(ctx, params, payload); err != nil {
		return nil, err
	}

	if resp, err = client.Do(req); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	bodybytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResponse errors.ErrorResponse
		if err = json.Unmarshal(bodybytes, &errResponse); err != nil {
			return nil, err
		}
		errResponse.Code = resp.StatusCode
		return nil, &errResponse
	}

	var (
		response Response
		message  ResponseMessage
	)

	if err = json.NewDecoder(bytes.NewBuffer(bodybytes)).Decode(&message); err != nil {
		return nil, err
	}
	response.StatusCode = resp.StatusCode
	response.Headers = resp.Header
	response.Message = &message

	return &response, nil
}

type Response struct {
	StatusCode int
	Headers    map[string][]string
	Message    *ResponseMessage
}

type ResponseMessage struct {
	Product  string                      `json:"messaging_product,omitempty"`
	Contacts []*whatsapp.ResponseContact `json:"contacts,omitempty"`
	Messages []*whatsapp.MessageID       `json:"messages,omitempty"`
}

type RequestParams struct {
	SenderID   string
	ApiVersion string
	Headers    map[string]string
	Query      map[string]string
	Bearer     string
	BaseURL    string
	Endpoint   string
	Method     string
}

type Sender func(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error)

type SenderMiddleware func(next Sender) Sender
