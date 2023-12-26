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

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

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
