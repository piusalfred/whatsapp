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
	"github.com/piusalfred/whatsapp/user"
)

func (c *Client) BlockUsers(ctx context.Context, req *user.BlockRequest) (*user.BlockResponse, error) {
	resp, err := c.sender.BlockUsers(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("block: %w", err)
	}
	return resp, nil
}

func (c *Client) UnblockUsers(ctx context.Context, req *user.UnblockRequest) (*user.BlockResponse, error) {
	resp, err := c.sender.UnblockUsers(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("unblock: %w", err)
	}
	return resp, nil
}

func (c *Client) ListBlockedUsers(
	ctx context.Context,
	opts *user.ListBlockedUsersOptions,
) (*user.ListBlockedUsersResponse, error) {
	resp, err := c.sender.ListBlockedUsers(ctx, c.config, opts)
	if err != nil {
		return nil, fmt.Errorf("list blocked: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) BlockUsers(
	ctx context.Context,
	conf *config.Config,
	req *user.BlockRequest,
) (*user.BlockResponse, error) {
	resp, err := bc.getUsers().Block(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("block: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UnblockUsers(
	ctx context.Context,
	conf *config.Config,
	req *user.UnblockRequest,
) (*user.BlockResponse, error) {
	resp, err := bc.getUsers().Unblock(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("unblock: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListBlockedUsers(
	ctx context.Context,
	conf *config.Config,
	opts *user.ListBlockedUsersOptions,
) (*user.ListBlockedUsersResponse, error) {
	resp, err := bc.getUsers().ListBlocked(ctx, conf, opts)
	if err != nil {
		return nil, fmt.Errorf("list blocked: %w", err)
	}
	return resp, nil
}
