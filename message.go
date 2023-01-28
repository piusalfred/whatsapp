package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
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
)

// MessageFn is a function that sends a WhatsApp message to a user
func MessageFunc(ctx context.Context, client *http.Client, url string, token string, message *Message) (*MessageResponse, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload)); err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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
	var messageResponse MessageResponse

	if resp.StatusCode != http.StatusOK {
		var errResponse ErrorResponse
		if err = json.Unmarshal(bodybytes, &errResponse); err != nil {
			return nil, err
		}
		errResponse.Code = resp.StatusCode
		return nil, &errResponse
	}

	if err = json.NewDecoder(bytes.NewBuffer(bodybytes)).Decode(&messageResponse); err != nil {
		return nil, err
	}

	return &messageResponse, nil
}
