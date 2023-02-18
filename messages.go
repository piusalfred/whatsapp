package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

type (
	StatusResponse struct {
		Success bool `json:"success,omitempty"`
	}

	MessageStatusUpdateRequest struct {
		MessagingProduct string `json:"messaging_product,omitempty"` // always whatsapp
		Status           string `json:"status,omitempty"`            // always read
		MessageID        string `json:"message_id,omitempty"`
	}
)

// MarkMessageRead sends a read receipt for a message.
// When you receive an incoming message from Webhooks, you can use the /messages endpoint
// to mark the message as read by changing its status to read. Messages marked as read
// display two blue check marks alongside their timestamp:
// We recommend marking incoming messages as read within 30 days of receipt. You cannot mark
// outgoing messages you sent as read. Marking a message as read will also mark earlier
// messages in the conversation as read.
func MarkMessageRead(ctx context.Context, client *http.Client, url, token string) (*StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	q := req.URL.Query()
	q.Add("access_token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type SendTextRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Message       string
	PreviewURL    bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, req *SendTextRequest) (*whttp.Response, error) {
	text := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "text",
		Text: &models.Text{
			PreviewUrl: req.PreviewURL,
			Body:       req.Message,
		},
	}

	params := &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer:  req.AccessToken,
		BaseURL: req.BaseURL,
		Method:  http.MethodPost,
		Endpoints: []string{
			"messages"},
	}

	payload, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}

	return whttp.Send(ctx, client, params, payload)
}

type SendLocationRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Location      *models.Location
}

func SendLocation(ctx context.Context, client *http.Client, req *SendLocationRequest) (*whttp.Response, error) {
	location := &models.Message{
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

	params := &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer:  req.AccessToken,
		BaseURL: req.BaseURL,
		Method:  http.MethodPost,
		Endpoints: []string{
			"messages"},
	}
	return whttp.Send(ctx, client, params, payload)
}

type ReactRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	MessageID     string
	Emoji         string
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
func React(ctx context.Context, client *http.Client, req *ReactRequest) (*whttp.Response, error) {
	reaction := &models.Message{
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

	params := &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer:  req.AccessToken,
		BaseURL: req.BaseURL,
		Method:  http.MethodPost,
		Endpoints: []string{
			"messages"},
	}

	return whttp.Send(ctx, client, params, payload)
}

type SendContactRequest struct {
	Recipient string
	Contacts  *models.Contacts
}

func SendContact(ctx context.Context, client *http.Client, params *whttp.RequestParams, req *SendContactRequest) (*whttp.Response, error) {
	contact := &models.Message{
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

	return whttp.Send(ctx, client, params, payload)
}

// ReplyParams contains options for replying to a message.
type ReplyParams struct {
	Recipient   string
	Context     string // this is ID of the message to reply to
	MessageType MessageType
	Content     any // this is a Text if MessageType is Text
}

// Reply is used to reply to a message. It accepts a ReplyParams and returns a Response and an error.
// You can send any message as a reply to a previous message in a conversation by including the previous
// message's ID set as Context in ReplyParams. The recipient will receive the new message along with a
// contextual bubble that displays the previous message's content.
//
// Recipients will not see a contextual bubble if:
//
// replying with a template message ("type":"template")
// replying with an image, video, PTT, or audio, and the recipient is on KaiOS
// These are known bugs which are being addressed.
// Example of Text reply:
// "messaging_product": "whatsapp",
//
//	  "context": {
//	    "message_id": "MESSAGE_ID"
//	  },
//	  "to": "<phone number> or <wa_id>",
//	  "type": "text",
//	  "text": {
//	    "preview_url": False,
//	    "body": "your-text-message-content"
//	  }
//	}'
func Reply(ctx context.Context, client *http.Client, params *whttp.RequestParams, options *ReplyParams) (*whttp.Response, error) {
	if options == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	payload, err := buildReplyPayload(options)
	if err != nil {
		return nil, err
	}

	return whttp.Send(ctx, client, params, payload)
}

// buildReplyPayload builds the payload for a reply. It accepts ReplyParams and returns a byte array
// and an error. This function is used internally by Reply.
func buildReplyPayload(options *ReplyParams) ([]byte, error) {
	contentByte, err := json.Marshal(options.Content)
	if err != nil {
		return nil, err
	}
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","context":{"message_id":"`)
	payloadBuilder.WriteString(options.Context)
	payloadBuilder.WriteString(`"},"to":"`)
	payloadBuilder.WriteString(options.Recipient)
	payloadBuilder.WriteString(`","type":"`)
	payloadBuilder.WriteString(string(options.MessageType))
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(string(options.MessageType))
	payloadBuilder.WriteString(`":`)
	payloadBuilder.Write(contentByte)
	payloadBuilder.WriteString(`}`)
	return []byte(payloadBuilder.String()), nil
}

type SendTemplateRequest struct {
	BaseURL                string
	AccessToken            string
	PhoneNumberID          string
	ApiVersion             string
	Recipient              string
	TemplateLanguageCode   string
	TemplateLanguagePolicy string
	TemplateName           string
	TemplateComponents     []*models.TemplateComponent
}

func SendTemplate(ctx context.Context, client *http.Client, req *SendTemplateRequest) (*whttp.Response, error) {
	template := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "template",
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   req.TemplateLanguageCode,
				Policy: req.TemplateLanguagePolicy,
			},
			Name:       req.TemplateName,
			Namespace:  "whatsapp",
			Components: req.TemplateComponents,
		},
	}
	params := &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer:  req.AccessToken,
		BaseURL: req.BaseURL,
		Method:  http.MethodPost,
		Endpoints: []string{
			"messages"},
	}
	payload, err := json.Marshal(template)
	if err != nil {
		return nil, err
	}

	return whttp.Send(ctx, client, params, payload)
}
