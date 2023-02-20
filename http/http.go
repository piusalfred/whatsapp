// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this
// software and associated documentation files (the “Software”), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons
// to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or
// substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/piusalfred/whatsapp/errors"
)

const BaseURL = "https://graph.facebook.com"

type (
	// Response is the response from the WhatsApp server
	// Example:
	//		{
	//	  		"messaging_product": "whatsapp",
	//	  		"contacts": [{
	//	      		"input": "PHONE_NUMBER",
	//	      		"wa_id": "WHATSAPP_ID",
	//	    	}]
	//	  		"messages": [{
	//	      		"id": "wamid.ID",
	//	    	}]
	//		}
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
		Endpoints  []string
		Method     string
	}

	ErrorResponse struct {
		Code int           `json:"code,omitempty"`
		Err  *errors.Error `json:"error,omitempty"`
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

// CreateRequestURL creates a new request url by joining the base url, api version
// sender id and endpoints.
// It is called by the NewRequestWithContext function where these details are
// passed from the RequestParams.
func CreateRequestURL(baseURL, apiVersion, senderID string, endpoints ...string) (string, error) {
	elems := append([]string{apiVersion, senderID}, endpoints...)
	return url.JoinPath(baseURL, elems...)
}

// NewRequestWithContext creates a new *http.Request with context by using the
// RequestParams.
func NewRequestWithContext(ctx context.Context, params *RequestParams, payload []byte) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	requestURL, err := CreateRequestURL(params.BaseURL, params.ApiVersion, params.SenderID, params.Endpoints...)
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

func SendMessage(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error) {
	var (
		resp      *http.Response
		err       error
		bodybytes []byte
	)

	if resp, err = Send(ctx, client, params, payload); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, resp.Body)
	if err != nil {
		return nil, err
	}
	bodybytes = buff.Bytes()

	if resp.StatusCode != http.StatusOK {
		var errResponse ErrorResponse
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

// Send sends a http request and returns a *http.Response or an error.
func Send(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*http.Response, error) {
	req, err := NewRequestWithContext(ctx, params, payload)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// Error returns the error message for ErrorResponse.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("whatsapp error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

// Unwrap returns the underlying error.
func (e *ErrorResponse) Unwrap() []error {
	return []error{e.Err}
}
