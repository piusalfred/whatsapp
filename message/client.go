//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package message

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type Client struct {
	mu     *sync.Mutex
	reader config.Reader
	config *config.Config
	sender RequestSender
}

func (c *Client) SendText(ctx context.Context, request *Request[Text]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithTextMessage)
}

func (c *Client) SendLocation(ctx context.Context, request *Request[Location]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithLocationMessage)
}

func (c *Client) SendVideo(ctx context.Context, request *Request[Video]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithVideo)
}

func (c *Client) SendReaction(ctx context.Context, request *Request[Reaction]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithReaction)
}

func (c *Client) SendTemplate(ctx context.Context, request *Request[Template]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithTemplateMessage)
}

func (c *Client) SendImage(ctx context.Context, request *Request[Image]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithImage)
}

func (c *Client) SendAudio(ctx context.Context, request *Request[Audio]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithAudio)
}

func (c *Client) RequestLocation(ctx context.Context, request *Request[string]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithRequestLocationMessage)
}

func (c *Client) SendDocument(ctx context.Context, request *Request[Document]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithDocument)
}

func (c *Client) SendSticker(ctx context.Context, request *Request[Sticker]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithSticker)
}

func (c *Client) SendContacts(ctx context.Context, request *Request[Contacts]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithContacts)
}

func (c *Client) SendInteractiveMessage(ctx context.Context, request *Request[Interactive]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.RecipientType, request.Message, WithInteractiveMessage)
}

func (c *Client) ReloadConfig(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	c.config, err = c.reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	return nil
}

func NewClient(ctx context.Context, reader config.Reader, sender whttp.Sender[Message],
	middlewares ...SenderMiddleware,
) (*Client, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	s := &BaseSender{sender}
	sf := s.SendRequest
	if len(middlewares) > 0 {
		for i := len(middlewares) - 1; i >= 0; i-- {
			mw := middlewares[i]
			if mw != nil {
				sf = mw(sf)
			}
		}
	}

	c := &Client{
		mu:     &sync.Mutex{},
		reader: reader,
		config: conf,
		sender: RequestSenderFunc(sf),
	}

	return c, nil
}

func (c *Client) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeSendMessage),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: true,
			InspectResponseError:  true,
		}),
	)

	response, err := c.sender.SendRequest(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (c *Client) UpdateStatus(ctx context.Context, request *StatusUpdateRequest) (*StatusUpdateResponse, error) {
	ms := string(request.Status)
	message := &Message{
		Product:   MessagingProduct,
		Status:    &ms,
		MessageID: &request.MessageID,
	}

	if request.WithTypingIndicator {
		message.TypingIndicator = &TypingIndicator{Type: "text"}
	}

	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeUpdateStatus),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: false,
			InspectResponseError:  true,
		}),
	)

	response, err := c.sender.SendRequest(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update message Status: %w", err)
	}

	return &StatusUpdateResponse{Success: response.Success}, nil
}
