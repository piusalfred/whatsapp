/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
	"net/http"
	"strings"
)

const (
	MessageEndpoint = "messages"
)

type (
	TextMessage struct {
		Message    string
		PreviewURL bool
	}
)

// SendTextMessage sends a text message to a WhatsApp Business Account.
func (client *Client) SendTextMessage(ctx context.Context, recipient string,
	message *TextMessage) (*ResponseMessage, error) {
	text := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          textMessageType,
		Text: &models.Text{
			PreviewURL: message.PreviewURL,
			Body:       message.Message,
		},
	}

	return client.SendMessage(ctx, "text message", text)
}

// SendLocationMessage sends a location message to a WhatsApp Business Account.
func (client *Client) SendLocationMessage(ctx context.Context, recipient string,
	message *models.Location) (*ResponseMessage, error) {

	location := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          locationMessageType,
		Location: &models.Location{
			Name:      message.Name,
			Address:   message.Address,
			Latitude:  message.Latitude,
			Longitude: message.Longitude,
		},
	}

	return client.SendMessage(ctx, "location message", location)
}

type ReactMessage struct {
	MessageID string
	Emoji     string
}

// React sends a reaction to a message.
// To send reaction messages, make a POST call to /PHONE_NUMBER_ID/messages and attach a message object
// with type=reaction. Then, add a reaction object.
//
// Sample request:
//
//	curl -X  POST \
//	 'https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages' \
//	 -H 'Authorization: Bearer ACCESS_TOKEN' \
//	 -H 'Content-Type: application/json' \
//	 -d '{
//	  "messaging_product": "whatsapp",
//	  "recipient_type": "individual",
//	  "to": "PHONE_NUMBER",
//	  "type": "reaction",
//	  "reaction": {
//	    "message_id": "wamid.HBgLM...",
//	    "emoji": "\uD83D\uDE00"
//	  }
//	}'
//
// If the message you are reacting to is more than 30 days old, doesn't correspond to any message
// in the conversation, has been deleted, or is itself a reaction message, the reaction message will
// not be delivered, and you will receive a webhooks with the code 131009.
//
// A successful response includes an object with an identifier prefixed with wamid. Use the ID listed
// after wamid to track your message status.
//
// Example response:
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
func (client *Client) React(ctx context.Context, recipient string, req *ReactMessage) (*ResponseMessage, error) {
	reaction := &models.Message{
		Product: messagingProduct,
		To:      recipient,
		Type:    reactionMessageType,
		Reaction: &models.Reaction{
			MessageID: req.MessageID,
			Emoji:     req.Emoji,
		},
	}

	return client.SendMessage(ctx, "react to message", reaction)
}

func (client *Client) SendMessage(ctx context.Context, name string, msg *models.Message) (*ResponseMessage, error) {
	cctx := client.Context()

	reqCtx := &whttp.RequestContext{
		Name:       name,
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{MessageEndpoint},
	}
	request := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  cctx.AccessToken,
		Payload: msg,
	}
	var resp ResponseMessage
	err := client.http.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return &resp, nil
}

// SendContacts sends a contact message. Contacts can be easily built using the models.NewContact() function.
func (client *Client) SendContacts(ctx context.Context, recipient string, contacts []*models.Contact) (
	*ResponseMessage, error) {
	contact := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          contactsMessageType,
		Contacts:      contacts,
	}

	return client.SendMessage(ctx, "send contacts", contact)

}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: messagingProduct,
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	cctx := client.Context()

	reqCtx := &whttp.RequestContext{
		Name:       "mark read",
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{MessageEndpoint},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  cctx.AccessToken,
		Payload: reqBody,
	}

	var success StatusResponse
	err := client.http.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return &success, nil
}

// ReplyRequest contains options for replying to a message.
type ReplyRequest struct {
	Recipient   string
	Context     string // this is ID of the message to reply to
	MessageType MessageType
	Content     any // this is a Text if MessageType is Text
}

// Reply is used to reply to a message. It accepts a ReplyRequest and returns a Response and an error.
// You can send any message as a reply to a previous message in a conversation by including the previous
// message's ID set as Ctx in ReplyRequest. The recipient will receive the new message along with a
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
//	  "Ctx": {
//	    "message_id": "MESSAGE_ID"
//	  },
//	  "to": "<phone number> or <wa_id>",
//	  "type": "text",
//	  "text": {
//	    "preview_url": False,
//	    "body": "your-text-message-content"
//	  }
//	}'
func (client *Client) Reply(ctx context.Context, request *ReplyRequest,
	hooks ...whttp.Hook,
) (*ResponseMessage, error) {
	if request == nil {
		return nil, fmt.Errorf("reply request is nil: %w", ErrBadRequestFormat)
	}
	payload, err := formatReplyPayload(request)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}
	cctx := client.Context()

	reqCtx := &whttp.RequestContext{
		Name:       "reply to message",
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  cctx.AccessToken,
		Form:    nil,
		Payload: payload,
	}

	var message ResponseMessage
	err = client.http.Do(ctx, req, &message)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}

	return &message, nil
}

// formatReplyPayload builds the payload for a reply. It accepts ReplyRequest and returns a byte array
// and an error. This function is used internally by Reply.
func formatReplyPayload(options *ReplyRequest) ([]byte, error) {
	contentByte, err := json.Marshal(options.Content)
	if err != nil {
		return nil, fmt.Errorf("format reply payload: %w", err)
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
