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
	"github.com/piusalfred/whatsapp/flow"
)

func (c *Client) CreateFlow(ctx context.Context, req flow.CreateRequest) (*flow.CreateResponse, error) {
	resp, err := c.sender.CreateFlow(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create flow: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateFlow(ctx context.Context, id string, req flow.UpdateRequest) (*flow.UpdateResponse, error) {
	resp, err := c.sender.UpdateFlow(ctx, c.config, id, req)
	if err != nil {
		return nil, fmt.Errorf("update flow: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateFlowJSON(
	ctx context.Context,
	req *flow.UpdateFlowJSONRequest,
) (*flow.UpdateFlowJSONResponse, error) {
	resp, err := c.sender.UpdateFlowJSON(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update flow json: %w", err)
	}
	return resp, nil
}

func (c *Client) ListFlows(ctx context.Context) (*flow.ListResponse, error) {
	resp, err := c.sender.ListFlows(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	return resp, nil
}

func (c *Client) ListFlowAssets(ctx context.Context, id string) (*flow.RetrieveAssetsResponse, error) {
	resp, err := c.sender.ListFlowAssets(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("list flow assets: %w", err)
	}
	return resp, nil
}

func (c *Client) PublishFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.PublishFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("publish flow: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.DeleteFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("delete flow: %w", err)
	}
	return resp, nil
}

func (c *Client) DeprecateFlow(ctx context.Context, id string) (*flow.SuccessResponse, error) {
	resp, err := c.sender.DeprecateFlow(ctx, c.config, id)
	if err != nil {
		return nil, fmt.Errorf("deprecate flow: %w", err)
	}
	return resp, nil
}

func (c *Client) GetFlow(ctx context.Context, req *flow.GetRequest) (*flow.SingleFlowResponse, error) {
	resp, err := c.sender.GetFlow(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}
	return resp, nil
}

func (c *Client) GenerateFlowPreview(ctx context.Context, req *flow.PreviewRequest) (*flow.PreviewResponse, error) {
	resp, err := c.sender.GenerateFlowPreview(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("generate flow preview: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateFlow(
	ctx context.Context,
	conf *config.Config,
	req flow.CreateRequest,
) (*flow.CreateResponse, error) {
	resp, err := bc.getFlows().Create(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateFlow(
	ctx context.Context,
	conf *config.Config,
	id string,
	req flow.UpdateRequest,
) (*flow.UpdateResponse, error) {
	resp, err := bc.getFlows().Update(ctx, conf, id, req)
	if err != nil {
		return nil, fmt.Errorf("update flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateFlowJSON(
	ctx context.Context,
	conf *config.Config,
	req *flow.UpdateFlowJSONRequest,
) (*flow.UpdateFlowJSONResponse, error) {
	resp, err := bc.getFlows().UpdateFlowJSON(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update flow json: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListFlows(ctx context.Context, conf *config.Config) (*flow.ListResponse, error) {
	resp, err := bc.getFlows().ListAll(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListFlowAssets(
	ctx context.Context,
	conf *config.Config,
	id string,
) (*flow.RetrieveAssetsResponse, error) {
	resp, err := bc.getFlows().ListAssets(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("list flow assets: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) PublishFlow(ctx context.Context, conf *config.Config, id string) (*flow.SuccessResponse, error) {
	resp, err := bc.getFlows().Publish(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("publish flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteFlow(ctx context.Context, conf *config.Config, id string) (*flow.SuccessResponse, error) {
	resp, err := bc.getFlows().Delete(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("delete flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeprecateFlow(
	ctx context.Context,
	conf *config.Config,
	id string,
) (*flow.SuccessResponse, error) {
	resp, err := bc.getFlows().Deprecate(ctx, conf, id)
	if err != nil {
		return nil, fmt.Errorf("deprecate flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetFlow(
	ctx context.Context,
	conf *config.Config,
	req *flow.GetRequest,
) (*flow.SingleFlowResponse, error) {
	resp, err := bc.getFlows().Get(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get flow: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GenerateFlowPreview(
	ctx context.Context,
	conf *config.Config,
	req *flow.PreviewRequest,
) (*flow.PreviewResponse, error) {
	resp, err := bc.getFlows().GeneratePreview(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("generate flow preview: %w", err)
	}
	return resp, nil
}
