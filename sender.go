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

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

// Sender implementors.
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
