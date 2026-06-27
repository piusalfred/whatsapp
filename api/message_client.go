//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package api

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/config"
	message "github.com/piusalfred/whatsapp/message"
)

// SendMessage sends a WhatsApp message using the fixed configuration.
// Build the message with [message.New] and the available Option helpers,
// or construct it directly. The SendInfo carries routing metadata.
func (c *Client) SendMessage(
	ctx context.Context,
	si *message.SendInfo,
	msg *message.Message,
) (*message.SendMessageResponse, error) {
	resp, err := c.sender.SendMessage(ctx, c.config, si, msg)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return resp, nil
}

// UpdateMessageStatus updates the status of a previously sent message using
// the fixed configuration.
func (c *Client) UpdateMessageStatus(
	ctx context.Context,
	req *message.StatusUpdateRequest,
) (*message.StatusUpdateResponse, error) {
	resp, err := c.sender.UpdateMessageStatus(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update message status: %w", err)
	}
	return resp, nil
}

// SendMessage sends a message to a WhatsApp user. Accepts a per-call configuration.
func (bc *BaseClient) SendMessage(
	ctx context.Context,
	conf *config.Config,
	si *message.SendInfo,
	msg *message.Message,
) (*message.SendMessageResponse, error) {
	resp, err := bc.getMessage().SendMessage(ctx, conf, si, msg)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return resp, nil
}

// UpdateMessageStatus updates the status of a previously sent message. Accepts a
// per-call configuration.
func (bc *BaseClient) UpdateMessageStatus(
	ctx context.Context,
	conf *config.Config,
	req *message.StatusUpdateRequest,
) (*message.StatusUpdateResponse, error) {
	resp, err := bc.getMessage().UpdateMessageStatus(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update message status: %w", err)
	}
	return resp, nil
}
