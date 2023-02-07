package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pesakit/whatsapp/pkg/models"
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

type SendTextRequest struct {
	Recipient  string
	Message    string
	PreviewURL bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, params *RequestParams, req *SendTextRequest) (*Response, error) {
	text := &Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "text",
		Text: &models.Text{
			PreviewUrl: req.PreviewURL,
			Body:       req.Message,
		},
	}

	payload, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

type SendLocationRequest struct {
	Recipient string
	Location  *models.Location
}

func SendLocation(ctx context.Context, client *http.Client, params *RequestParams, req *SendLocationRequest) (*Response, error) {
	location := &Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "location",
		Location:      req.Location,
	}
	payload, err := json.Marshal(location)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

type ReactRequest struct {
	Recipient string
	MessageID string
	Emoji     string
}

/*
React sends a reaction to a message.
To send reaction messages, make a POST call to /PHONE_NUMBER_ID/messages and attach a message object
with type=reaction. Then, add a reaction object.

Sample request:

	curl -X  POST \
	 'https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages' \
	 -H 'Authorization: Bearer ACCESS_TOKEN' \
	 -H 'Content-Type: application/json' \
	 -d '{
	  "messaging_product": "whatsapp",
	  "recipient_type": "individual",
	  "to": "PHONE_NUMBER",
	  "type": "reaction",
	  "reaction": {
	    "message_id": "wamid.HBgLM...",
	    "emoji": "\uD83D\uDE00"
	  }
	}'

If the message you are reacting to is more than 30 days old, doesn't correspond to any message
in the conversation, has been deleted, or is itself a reaction message, the reaction message will
not be delivered and you will receive a webhooks with the code 131009.

A successful response includes an object with an identifier prefixed with wamid. Use the ID listed
after wamid to track your message status.

Example response:

	{
	  "messaging_product": "whatsapp",
	  "contacts": [{
	      "input": "PHONE_NUMBER",
	      "wa_id": "WHATSAPP_ID",
	    }]
	  "messages": [{
	      "id": "wamid.ID",
	    }]
	}
*/
func React(ctx context.Context, client *http.Client, params *RequestParams, req *ReactRequest) (*Response, error) {
	reaction := &Message{
		Product: "whatsapp",
		To:      req.Recipient,
		Type:    "reaction",
		Reaction: &models.Reaction{
			MessageID: req.MessageID,
			Emoji:     req.Emoji,
		},
	}

	payload, err := json.Marshal(reaction)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

type SendContactRequest struct {
	Recipient string
	Contacts  *models.Contacts
}

func SendContact(ctx context.Context, client *http.Client, params *RequestParams, req *SendContactRequest) (*Response, error) {
	contact := &Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "contact",
		Contacts:      req.Contacts,
	}
	payload, err := json.Marshal(contact)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}
