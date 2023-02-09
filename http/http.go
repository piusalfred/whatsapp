package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/piusalfred/whatsapp/errors"
	"io"
	"net/http"
	"net/url"
)

type (
	Response struct {
		StatusCode int
		Headers    map[string][]string
		Message    *ResponseMessage
	}

	ResponseMessage struct {
		Product  string             `json:"messaging_product,omitempty"`
		Contacts []*ResponseContact `json:"contacts,omitempty"`
		Messages []*MessageID       `json:"messages,omitempty"`
	}
	RequestParams struct {
		SenderID   string
		ApiVersion string
		Headers    map[string]string
		Query      map[string]string
		Bearer     string
		Form       map[string]string
		BaseURL    string
		Endpoint   string
		Method     string
	}

	Sender func(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error)

	SenderMiddleware func(next Sender) Sender

	MessageID struct {
		ID string `json:"id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappId string `json:"wa_id"`
	}
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

	if params.Form != nil {
		form := url.Values{}
		for key, value := range params.Form {
			form.Add(key, value)
		}
		req, err = http.NewRequestWithContext(ctx, params.Method, requestURL, bytes.NewBufferString(form.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}
	} else {

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
