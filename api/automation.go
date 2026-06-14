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
	"github.com/piusalfred/whatsapp/conversation/automation"
)

func (c *Client) AddConversationComponents(
	ctx context.Context,
	commands []*automation.Command,
	prompts []string,
) (*automation.SuccessResponse, error) {
	resp, err := c.sender.AddConversationComponents(ctx, c.config, commands, prompts)
	if err != nil {
		return nil, fmt.Errorf("add components: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateWelcomeMessageStatus(
	ctx context.Context,
	shouldEnable bool,
) (*automation.SuccessResponse, error) {
	resp, err := c.sender.UpdateWelcomeMessageStatus(ctx, c.config, shouldEnable)
	if err != nil {
		return nil, fmt.Errorf("update welcome message status: %w", err)
	}
	return resp, nil
}

func (c *Client) ListConversationComponents(ctx context.Context) (*automation.BaseResponse, error) {
	resp, err := c.sender.ListConversationComponents(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}
	return resp, nil
}

func (c *Client) GetBotDetails(ctx context.Context, request *automation.BotRequest) (*automation.Bot, error) {
	resp, err := c.sender.GetBotDetails(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("get bot details: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) AddConversationComponents(ctx context.Context, conf *config.Config,
	commands []*automation.Command, prompts []string,
) (*automation.SuccessResponse, error) {
	resp, err := bc.getAuto().AddConversationComponents(ctx, conf, commands, prompts)
	if err != nil {
		return nil, fmt.Errorf("add components: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateWelcomeMessageStatus(ctx context.Context, conf *config.Config,
	shouldEnable bool,
) (*automation.SuccessResponse, error) {
	resp, err := bc.getAuto().UpdateWelcomeMessageStatus(ctx, conf, shouldEnable)
	if err != nil {
		return nil, fmt.Errorf("update welcome message status: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListConversationComponents(
	ctx context.Context,
	conf *config.Config,
) (*automation.BaseResponse, error) {
	resp, err := bc.getAuto().ListConversationComponents(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetBotDetails(
	ctx context.Context,
	conf *config.Config,
	request *automation.BotRequest,
) (*automation.Bot, error) {
	resp, err := bc.getAuto().GetBotDetails(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get bot details: %w", err)
	}
	return resp, nil
}
