package whatsapp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type (

	// RequestParams are parameters for a request containing headers, query params,
	// Bearer token, Method and the body.
	// These parameters are used to create a *http.Request
	RequestParams struct {
		Headers  map[string]string
		Query    map[string]string
		Bearer   string
		UrlParts []string // url parts which will be joined
		Method   string
		Body     []byte
	}

	// MessageSendFunc a type for a function that sends a message
	MessageSendFunc func(ctx context.Context, client *http.Client, url, token string, message *Message) (*MessageResponse, error)
)

// NewRequestWithContext creates a new *http.Request with context by using the
// RequestParams.
func NewRequestWithContext(ctx context.Context, params *RequestParams) (*http.Request, error) {
	requestURL, err := url.JoinPath(params.UrlParts[0], params.UrlParts[1:]...)
	if err != nil {
		return nil, fmt.Errorf("failed to join url parts: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, params.Method, requestURL, bytes.NewBuffer(params.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
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
