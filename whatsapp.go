package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages

const (
	BaseURL                = "https://graph.facebook.com/"
	TextMessageType        = "text"
	ReactionMessageType    = "reaction"
	MediaMessageType       = "media"
	LocationMessageType    = "location"
	ContactMessageType     = "contact"
	InteractiveMessageType = "interactive"
)

type (

	// Reaction is a WhatsApp reaction
	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	// ReactionMessage is a WhatsApp reaction message
	// If the message you are reacting to is more than 30 days old, doesn't correspond to
	// any message in the conversation, has been deleted, or is itself a reaction message,
	// the reaction message will not be delivered and you will receive a webhook with the
	// code 131009.
	//'{
	//   "messaging_product": "whatsapp",
	//   "recipient_type": "individual",
	//   "to": "PHONE_NUMBER",
	//   "type": "reaction",
	//   "reaction": {
	//     "message_id": "wamid.HBgLM...",
	//     "emoji": "\uD83D\uDE00"
	//   }
	// }'
	ReactionMessage struct {
		Product       string    `json:"messaging_product"`
		RecipientType string    `json:"recipient_type"`
		To            string    `json:"to"`
		Type          string    `json:"type"`
		Reaction      *Reaction `json:"reaction"`
	}

	// Text ...
	Text struct {
		PreviewUrl bool   `json:"preview_url,omitempty"`
		Body       string `json:"body,omitempty"`
	}

	// TextMessage is a WhatsApp text message
	// //{
	// 	"messaging_product": "whatsapp",
	// 	"recipient_type": "individual",
	// 	"to": "PHONE_NUMBER",
	// 	"type": "text",
	// 	"text": { // the text object
	// 	  "preview_url": false,
	// 	  "body": "MESSAGE_CONTENT"
	// 	  }
	//   }'
	TextMessage struct {
		Product       string `json:"messaging_product"`
		RecipientType string `json:"recipient_type"`
		To            string `json:"to"`
		Type          string `json:"type"`
		Text          *Text  `json:"text"`
	}

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, with the exception of reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	ResponseHeaders map[string][]string

	// Message is a WhatsApp message
	Message struct {
		Product  string           `json:"messaging_product"`
		To       string           `json:"to"`
		Type     string           `json:"type"`
		Template *MessageTemplate `json:"template"`
	}

	TemplateLanguage struct {
		Code string `json:"code"`
	}

	MessageTemplate struct {
		Name     string           `json:"name"`
		Language TemplateLanguage `json:"language"`
	}

	// MessageResponse is a WhatsApp message response
	// {
	//		"messaging_product":"whatsapp",
	//		"contacts":[
	//				{
	//					"input":"255767001828",
	//					"wa_id":"255767001828",
	//				},
	//		],
	//		"messages":[
	//				{"id":"wamid.HBgMMjU1NzY3MDAxODI4FQIAERgSRjVDRDE5NjhBOEEwQzQ0NUE1AA=="},
	//		],
	//	}
	MessageResponse struct {
		MessagingProduct string `json:"messaging_product"`
		Contacts         []struct {
			Input      string `json:"input"`
			WhatsappID string `json:"wa_id"`
		} `json:"contacts"`
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}

	// Response is the response from the WhatsApp server
	//
	//	{
	//	  "messaging_product": "whatsapp",
	//	  "contacts": [{
	//	      "input": "PHONE_NUMBER",
	//	      "wa_id": "WHATSAPP_ID",
	//	    }]
	//	  "messages": [{
	//	      "id": "wamid.ID",
	//	    }]
	//	}
	MessageID struct {
		ID string `json:"id"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappId string `json:"wa_id"`
	}

	Response struct {
		MessagingProduct string            `json:"messaging_product"`
		Contacts         []ResponseContact `json:"contacts"`
		Messages         []MessageID       `json:"messages"`
	}

	// Image ...
	// The Cloud API supports media HTTP caching. If you are using a link (link) to a media
	// asset on your server (as opposed to the ID (id) of an asset you have uploaded to our servers),
	// you can instruct us to cache your asset for reuse with future messages by including
	// the headers below in your server response when we request the asset. If none of these
	// headers are included, we will not cache your asset.
	// Cache-Control: <CACHE_CONTROL>
	// Last-Modified: <LAST_MODIFIED>
	// ETag: <ETAG>
	// Cache-Control
	// The Cache-Control header tells us how to handle asset caching. We support the following directives:
	//
	// max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
	// messages until this time is exceeded, after which we will request the asset again, if needed.
	// Example: Cache-Control: max-age=604800.
	//
	// no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
	// is different from a previous response. Requires the Last-Modified header.
	// Example: Cache-Control: no-cache.
	//
	// no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.
	//
	// private: Indicates that the asset is personalized for the recipient and should not be cached.
	//
	// Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache.
	// If the Last-Modified value is different from a previous response and Cache-Control: no-cache is included
	// in the response, we will update our cached version of the asset with the asset in the response.
	// Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

	// ETag
	// The ETag header is a unique string that identifies a specific version of an asset.
	// Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified
	// headers are not included in the response. In this case, we will cache the asset according to our own,
	//internal logic (which we do not disclose).
	Image struct {
		Link string `json:"link,omitempty"`
		ID   string `json:"id,omitempty"`
	}

	// InteractiveMessage ...
	InteractiveMessage struct {
		Type   string `json:"type"`
		Header struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"header"`
		Body struct {
			Text string `json:"text"`
		} `json:"body"`
		Footer struct {
			Text string `json:"text"`
		} `json:"footer"`
		Action struct {
			Button   string `json:"button"`
			Sections []struct {
				Title string `json:"title"`
				Rows  []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"rows"`
			} `json:"sections"`
		} `json:"action"`
	}

	// RequestParams are parameters for a request containing headers, query params,
	// Bearer token, Method and the body.
	// These parameters are used to create a *http.Request
	RequestParams struct {
		SenderID   string
		ApiVersion string
		Headers    map[string]string
		Query      map[string]string
		Bearer     string
		BaseURL    string
		Method     string
	}

	Sender func(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error)

	SenderMiddleware func(next Sender) Sender
)

// NewRequestWithContext creates a new *http.Request with context by using the
// RequestParams.
func NewRequestWithContext(ctx context.Context, params *RequestParams, payload []byte) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	//https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages
	requestURL, err := url.JoinPath(params.BaseURL, params.ApiVersion, params.SenderID, "messages")
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

	var response Response
	if err = json.NewDecoder(bytes.NewBuffer(bodybytes)).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

type SendTextRequest struct {
	Recipient  string
	Message    string
	PreviewURL bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, params *RequestParams, req *SendTextRequest) (*Response, error) {
	text := &TextMessage{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "text",
		Text: &Text{
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

func SendLocation(ctx context.Context, client *http.Client, params *RequestParams, location *Location) (*Response, error) {
	payload, err := json.Marshal(location)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}
