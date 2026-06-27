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
	"github.com/piusalfred/whatsapp/templates"
)

func (c *Client) CreateTemplate(ctx context.Context, req *templates.CreateRequest) (*templates.CreateResponse, error) {
	resp, err := c.sender.CreateTemplate(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return resp, nil
}

func (c *Client) ListTemplates(ctx context.Context, req *templates.ListRequest) (*templates.ListResponse, error) {
	resp, err := c.sender.ListTemplates(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return resp, nil
}

func (c *Client) GetTemplate(ctx context.Context, req *templates.GetRequest) (*templates.TemplateInfo, error) {
	resp, err := c.sender.GetTemplate(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return resp, nil
}

func (c *Client) EditTemplate(ctx context.Context, req *templates.EditRequest) (*templates.CreateResponse, error) {
	resp, err := c.sender.EditTemplate(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("edit template: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteTemplate(ctx context.Context, req *templates.DeleteRequest) (*templates.BaseResponse, error) {
	resp, err := c.sender.DeleteTemplate(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("delete template: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateTemplate(
	ctx context.Context,
	conf *config.Config,
	req *templates.CreateRequest,
) (*templates.CreateResponse, error) {
	resp, err := bc.getTemplates().Create(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListTemplates(
	ctx context.Context,
	conf *config.Config,
	req *templates.ListRequest,
) (*templates.ListResponse, error) {
	resp, err := bc.getTemplates().List(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetTemplate(
	ctx context.Context,
	conf *config.Config,
	req *templates.GetRequest,
) (*templates.TemplateInfo, error) {
	resp, err := bc.getTemplates().Get(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) EditTemplate(
	ctx context.Context,
	conf *config.Config,
	req *templates.EditRequest,
) (*templates.CreateResponse, error) {
	resp, err := bc.getTemplates().Edit(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("edit template: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteTemplate(
	ctx context.Context,
	conf *config.Config,
	req *templates.DeleteRequest,
) (*templates.BaseResponse, error) {
	resp, err := bc.getTemplates().Delete(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete template: %w", err)
	}
	return resp, nil
}
