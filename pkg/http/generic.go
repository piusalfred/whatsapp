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

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/models"
)

var ErrInvalidPayload = fmt.Errorf("invalid payload")

type (
	MessageContent interface {
		models.Location |
			models.Text |
			models.Image |
			models.Audio |
			models.Video |
			models.Sticker |
			models.Document |
			models.Contacts |
			models.Template |
			models.Interactive
	}

	Message[M MessageContent] struct {
		Product       string          `json:"messaging_product,omitempty"`
		To            string          `json:"to,omitempty"`
		RecipientType string          `json:"recipient_type,omitempty"`
		Type          string          `json:"type,omitempty"`
		Context       *models.Context `json:"context,omitempty"`
		Content       *M              `json:"content,omitempty"`
	}

	GenericRequest[P MessageContent] struct {
		Recipient string
		Reply     string
		Context   *RequestContext
		Content   *P
	}
)

func contentType[P MessageContent](content *P) string {
	t := reflect.TypeOf(content)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return strings.ToLower(t.Name())
}

// EncodeMessageJSON encodes the given interface into the request body bytes.
func EncodeMessageJSON[P MessageContent](recipient string, reply string, content *P) ([]byte, error) {
	if recipient == "" {
		return nil, fmt.Errorf("%w: recipient is empty", ErrInvalidPayload)
	}

	if content == nil {
		return nil, fmt.Errorf("%w: content is nil", ErrInvalidPayload)
	}

	c := &Message[P]{
		Product:       "whatsapp",
		To:            recipient,
		RecipientType: "individual",
		Type:          contentType(content),
		Content:       content,
	}

	if reply != "" {
		c.Context = &models.Context{
			MessageID: reply,
		}
	}

	b, err := json.Marshal(c.Content)
	if err != nil {
		return nil, fmt.Errorf("content json: %w", err)
	}

	var recipientTypeContext string

	if c.Context != nil && c.Context.MessageID != "" {
		recipientTypeContext = fmt.Sprintf(`"context":{"message_id":"%s"},`, c.Context.MessageID)
	} else {
		// recipient_type is not required for reply messages.
		recipientTypeContext = `"recipient_type":"individual",`
	}

	jsonString := fmt.Sprintf(
		`{"messaging_product":"whatsapp",%s"to":"%s","type":"%s","%s":%s}`,
		recipientTypeContext,
		recipient,
		c.Type,
		c.Type,
		b,
	)

	return []byte(jsonString), nil
}

// Send ...
func Send[P MessageContent](ctx context.Context, sender Sender, req *GenericRequest[P]) (*ResponseMessage, error) {
	payload, err := EncodeMessageJSON(req.Recipient, req.Reply, req.Content)
	if err != nil {
		return nil, fmt.Errorf("encode message json: %w", err)
	}

	request := MakeRequest(
		WithRequestContext(req.Context),
		WithRequestPayload(payload),
		WithRequestMethod(http.MethodPost),
		WithRequestHeaders(map[string]string{
			"Content-Type": "application/json",
		}),
	)

	resp, err := sender.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	return resp, nil
}
