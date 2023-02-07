package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/piusalfred/whatsapp/pkg/http"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/models"
)

type SendTextRequest struct {
	Recipient  string
	Message    string
	PreviewURL bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, params *http2.RequestParams, req *SendTextRequest) (*http2.Response, error) {
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

	payload, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}

	return http2.Send(ctx, client, params, payload)
}

type SendLocationRequest struct {
	Recipient string
	Location  *models.Location
}

func SendLocation(ctx context.Context, client *http.Client, params *http2.RequestParams, req *SendLocationRequest) (*http2.Response, error) {
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

	return http2.Send(ctx, client, params, payload)
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
func React(ctx context.Context, client *http.Client, params *http2.RequestParams, req *ReactRequest) (*http2.Response, error) {
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

	return http2.Send(ctx, client, params, payload)
}

type SendContactRequest struct {
	Recipient string
	Contacts  *models.Contacts
}

func SendContact(ctx context.Context, client *http.Client, params *http2.RequestParams, req *SendContactRequest) (*http2.Response, error) {
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

	return http2.Send(ctx, client, params, payload)
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
// These are known bugs which we are being addressed.
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
func Reply(ctx context.Context, client *http.Client, params *http2.RequestParams, options *ReplyParams) (*http2.Response, error) {
	if options == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	payload, err := buildReplyPayload(options)
	if err != nil {
		return nil, err
	}

	return http2.Send(ctx, client, params, payload)
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
