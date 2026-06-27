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

package api

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/webhooks/callbacks"
)

func (c *Client) SetAlternativeCallback(
	ctx context.Context,
	request *callbacks.SetAlternativeCallbackRequest,
) (*callbacks.SuccessResponse, error) {
	resp, err := c.sender.SetAlternativeCallback(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("set alternative callback: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteAlternativeCallback(
	ctx context.Context,
	overrideType callbacks.OverrideType,
) (*callbacks.SuccessResponse, error) {
	resp, err := c.sender.DeleteAlternativeCallback(ctx, c.config, overrideType)
	if err != nil {
		return nil, fmt.Errorf("delete alternative callback: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) SetAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	request *callbacks.SetAlternativeCallbackRequest,
) (*callbacks.SuccessResponse, error) {
	resp, err := bc.getCallbacks().SetAlternativeCallback(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("set alternative callback: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteAlternativeCallback(
	ctx context.Context,
	conf *config.Config,
	overrideType callbacks.OverrideType,
) (*callbacks.SuccessResponse, error) {
	resp, err := bc.getCallbacks().DeleteAlternativeCallback(ctx, conf, overrideType)
	if err != nil {
		return nil, fmt.Errorf("delete alternative callback: %w", err)
	}
	return resp, nil
}
