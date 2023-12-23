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
	"fmt"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

type (
	// BaseClient wraps the http client only and is used to make requests to the whatsapp api,
	// It does not have the context. This is idealy for making requests to the whatsapp api for
	// different users. The Client struct is used to make requests to the whatsapp api for a
	// single user.
	BaseClient struct {
		base *whttp.Client
		mw   []SendMiddleware
	}

	// BaseClientOption is a function that implements the BaseClientOption interface.
	BaseClientOption func(*BaseClient)
)

// WithBaseClientMiddleware adds a middleware to the base client.
func WithBaseClientMiddleware(mw ...SendMiddleware) BaseClientOption {
	return func(client *BaseClient) {
		client.mw = append(client.mw, mw...)
	}
}

// WithBaseHTTPClient sets the http client for the base client.
func WithBaseHTTPClient(httpClient *whttp.Client) BaseClientOption {
	return func(client *BaseClient) {
		client.base = httpClient
	}
}

// NewBaseClient creates a new base client.
func NewBaseClient(options ...BaseClientOption) *BaseClient {
	b := &BaseClient{base: whttp.NewClient()}

	for _, option := range options {
		option(b)
	}

	return b
}

func (c *BaseClient) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message,
) (*ResponseMessage, error) {
	fs := WrapSender(SenderFunc(c.send), c.mw...)

	resp, err := fs.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("base client: %s: %w", req.Name, err)
	}

	return resp, nil
}

func (c *BaseClient) send(ctx context.Context, req *whttp.RequestContext,
	msg *models.Message,
) (*ResponseMessage, error) {
	request := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: msg,
	}

	var resp ResponseMessage
	err := c.base.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.Name, err)
	}

	return &resp, nil
}

func (c *BaseClient) MarkMessageRead(ctx context.Context, req *whttp.RequestContext,
	messageID string,
) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: MessagingProduct,
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	params := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: reqBody,
	}

	var success StatusResponse
	err := c.base.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("mark message read: %w", err)
	}

	return &success, nil
}

var _ Sender = (*BaseClient)(nil)

// Sender is an interface that represents a sender of a message.
type Sender interface {
	Send(ctx context.Context, req *whttp.RequestContext, message *models.Message) (*ResponseMessage, error)
}

// SenderFunc is a function that implements the Sender interface.
type SenderFunc func(ctx context.Context, req *whttp.RequestContext,
	message *models.Message) (*ResponseMessage, error)

// Send calls the function that implements the Sender interface.
func (f SenderFunc) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message) (*ResponseMessage,
	error,
) {
	return f(ctx, req, message)
}

// SendMiddleware that takes a Sender and returns a new Sender that will wrap the original
// Sender and execute the middleware function before sending the message.
type SendMiddleware func(Sender) Sender

// WrapSender wraps a Sender with a SendMiddleware.
func WrapSender(sender Sender, middleware ...SendMiddleware) Sender {
	// iterate backwards so that the middleware is executed in the right order
	for i := len(middleware) - 1; i >= 0; i-- {
		sender = middleware[i](sender)
	}

	return sender
}

// TransparentClient is a client that can send messages to a recipient without knowing the configuration of the client.
// It uses Sender instead of already configured clients. It is ideal for having a client for different environments.
type TransparentClient struct {
	Middlewares []SendMiddleware
}

// Send sends a message to the recipient.
func (client *TransparentClient) Send(ctx context.Context, sender Sender,
	req *whttp.RequestContext, message *models.Message, mw ...SendMiddleware,
) (*ResponseMessage, error) {
	s := WrapSender(WrapSender(sender, client.Middlewares...), mw...)

	response, err := s.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("transparent client: %w", err)
	}

	return response, nil
}
