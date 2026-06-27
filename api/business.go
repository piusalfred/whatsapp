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

	"github.com/piusalfred/whatsapp/business"
	"github.com/piusalfred/whatsapp/config"
)

func (c *Client) GetBusinessProfile(ctx context.Context, fields []string) ([]*business.Profile, error) {
	resp, err := c.sender.GetBusinessProfile(ctx, c.config, fields)
	if err != nil {
		return nil, fmt.Errorf("get business profile: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateBusinessProfile(ctx context.Context, req *business.UpdateProfileRequest) (bool, error) {
	resp, err := c.sender.UpdateBusinessProfile(ctx, c.config, req)
	if err != nil {
		return false, fmt.Errorf("update business profile: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetBusinessProfile(
	ctx context.Context,
	conf *config.Config,
	fields []string,
) ([]*business.Profile, error) {
	resp, err := bc.getBiz().Get(ctx, conf, fields)
	if err != nil {
		return nil, fmt.Errorf("get business profile: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateBusinessProfile(
	ctx context.Context,
	conf *config.Config,
	req *business.UpdateProfileRequest,
) (bool, error) {
	resp, err := bc.getBiz().Update(ctx, conf, req)
	if err != nil {
		return false, fmt.Errorf("update business profile: %w", err)
	}
	return resp, nil
}
