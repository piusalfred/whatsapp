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

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

func (c *BaseClient) SendText(ctx context.Context, request *Request[Text]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithTextMessage)
}

func (c *BaseClient) SendLocation(ctx context.Context, request *Request[Location]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithLocationMessage)
}

func (c *BaseClient) SendVideo(ctx context.Context, request *Request[Video]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithVideo)
}

func (c *BaseClient) SendReaction(ctx context.Context, request *Request[Reaction]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithReaction)
}

func (c *BaseClient) SendTemplate(ctx context.Context, request *Request[Template]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithTemplateMessage)
}

func (c *BaseClient) SendImage(ctx context.Context, request *Request[Image]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithImage)
}

func (c *BaseClient) SendAudio(ctx context.Context, request *Request[Audio]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithAudio)
}

func (c *BaseClient) RequestLocation(ctx context.Context, request *Request[string]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithRequestLocationMessage)
}

func (c *BaseClient) SendDocument(ctx context.Context, request *Request[Document]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithDocument)
}

func (c *BaseClient) SendSticker(ctx context.Context, request *Request[Sticker]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithSticker)
}

func (c *BaseClient) SendContacts(ctx context.Context, request *Request[Contacts]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithContacts)
}

func (c *BaseClient) SendInteractiveMessage(ctx context.Context, request *Request[Interactive]) (*Response, error) {
	return sendMessage(ctx, c, request.Recipient, request.ReplyTo, request.Message, WithInteractiveMessage)
}

func NewBaseClient(sender whttp.Sender[Message], reader config.Reader,
	middlewares ...SenderMiddleware,
) (*BaseClient, error) {
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
	c := &BaseClient{
		sender: RequestSenderFunc(sf),
		config: reader,
	}

	return c, nil
}

func (c *BaseClient) SetConfigReader(reader config.Reader) {
	c.config = reader
}

func (c *BaseClient) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	conf, err := c.config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("base client: send message: read config: %w", err)
	}

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

	response, err := c.sender.SendRequest(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (c *BaseClient) UpdateStatus(ctx context.Context, request *StatusUpdateRequest) (*StatusUpdateResponse, error) {
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
		}),
	)

	conf, err := c.config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("base client: update message status: read config: %w", err)
	}

	response, err := c.sender.SendRequest(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("base client: update message status: %w", err)
	}

	return &StatusUpdateResponse{Success: response.Success}, nil
}

type (
	BaseSender struct {
		Sender whttp.Sender[Message]
	}

	RequestSenderFunc func(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error)

	RequestSender interface {
		SendRequest(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error)
	}

	SenderMiddleware func(senderFunc RequestSenderFunc) RequestSenderFunc
)

func (fn RequestSenderFunc) SendRequest(
	ctx context.Context,
	conf *config.Config,
	request *BaseRequest,
) (*Response, error) {
	return fn(ctx, conf, request)
}

func (c *BaseSender) SendRequest(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error) {
	options := []whttp.RequestOption[Message]{
		whttp.WithRequestEndpoints[Message](conf.APIVersion, conf.PhoneNumberID, Endpoint),
		whttp.WithRequestBearer[Message](conf.AccessToken),
		whttp.WithRequestType[Message](request.Type),
		whttp.WithRequestAppSecret[Message](conf.AppSecret),
		whttp.WithRequestSecured[Message](conf.SecureRequests),
		whttp.WithRequestMessage(request.Message),
		whttp.WithRequestMetadata[Message](request.Metadata),
	}

	req := whttp.MakeRequest(request.Method, conf.BaseURL, options...)

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, request.DecodeOptions)

	if err := c.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("base client: send request: %w", err)
	}

	return response, nil
}

func sendMessage[T any](
	ctx context.Context,
	sender Sender,
	recipient, reply string,
	message *T,
	fn func(*T) Option,
) (*Response, error) {
	options := buildOptions(message, reply, fn)

	m, err := New(recipient, options...)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	response, err := sender.SendMessage(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return response, nil
}

func buildOptions[T any](message *T, replyTo string, createMessageFunc func(*T) Option) []Option {
	options := make([]Option, 1, 2) //nolint: mnd // ok
	options[0] = createMessageFunc(message)
	if replyTo != "" {
		options = append(options, WithMessageAsReplyTo(replyTo))
	}

	return options
}

type (
	Request[T any] struct {
		Recipient string
		ReplyTo   string
		Message   *T
	}

	BaseClient struct {
		sender RequestSender
		config config.Reader
	}

	BaseRequest struct {
		Method        string
		Endpoints     []string
		Type          whttp.RequestType
		Message       *Message
		DecodeOptions whttp.DecodeOptions
		Metadata      types.Metadata
	}

	BaseRequestOption func(request *BaseRequest)
)

func NewBaseRequest(message *Message, options ...BaseRequestOption) *BaseRequest {
	b := &BaseRequest{
		Method:    http.MethodPost,
		Endpoints: []string{Endpoint},
		Type:      whttp.RequestTypeSendMessage,
		Message:   message,
		DecodeOptions: whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: true,
		},
	}

	for _, option := range options {
		if option != nil {
			option(b)
		}
	}

	return b
}

func WithBaseRequestDecodeOptions(options whttp.DecodeOptions) BaseRequestOption {
	return func(request *BaseRequest) {
		request.DecodeOptions = options
	}
}

func WithBaseRequestMetadata(metadata map[string]any) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Metadata = metadata
	}
}

func WithBaseRequestEndpoints(endpoint ...string) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Endpoints = endpoint
	}
}

func WithBaseRequestMethod(method string) BaseRequestOption {
	return func(request *BaseRequest) {
		if method != "" {
			request.Method = method
		}
	}
}

func WithBaseRequestType(reqType whttp.RequestType) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Type = reqType
	}
}

func NewRequest[T any](recipient string, message *T, replyTo string) *Request[T] {
	return &Request[T]{Recipient: recipient, Message: message, ReplyTo: replyTo}
}
