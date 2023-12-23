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
	"net/http"
	"strings"

	whttp "github.com/piusalfred/whatsapp/http"
)

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
) (*ResponseMessage, error) {
	if request == nil {
		return nil, fmt.Errorf("reply request is nil: %w", ErrBadRequestFormat)
	}
	payload, err := formatReplyPayload(request)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}
	reqCtx := &whttp.RequestContext{
		Name:          "reply to message",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Endpoints:     []string{MessageEndpoint},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  client.config.AccessToken,
		Form:    nil,
		Payload: payload,
	}

	var message ResponseMessage
	err = client.bc.base.Do(ctx, req, &message)
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
