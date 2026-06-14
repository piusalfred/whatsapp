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
	"github.com/piusalfred/whatsapp/qrcode"
)

func (c *Client) CreateQR(ctx context.Context, req *qrcode.CreateRequest) (*qrcode.CreateResponse, error) {
	resp, err := c.sender.CreateQR(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create qr: %w", err)
	}
	return resp, nil
}

func (c *Client) GetQR(ctx context.Context, qrCodeID string) (*qrcode.Information, error) {
	resp, err := c.sender.GetQR(ctx, c.config, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}
	return resp, nil
}

func (c *Client) ListQR(
	ctx context.Context,
	opts *qrcode.ListOptions,
) (*qrcode.ListResponse, error) {
	resp, err := c.sender.ListQR(ctx, c.config, opts)
	if err != nil {
		return nil, fmt.Errorf("list qr: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteQR(ctx context.Context, qrCodeID string) (*qrcode.SuccessResponse, error) {
	resp, err := c.sender.DeleteQR(ctx, c.config, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("delete qr: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateQR(ctx context.Context, req *qrcode.UpdateRequest) (*qrcode.SuccessResponse, error) {
	resp, err := c.sender.UpdateQR(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateQR(
	ctx context.Context,
	conf *config.Config,
	req *qrcode.CreateRequest,
) (*qrcode.CreateResponse, error) {
	resp, err := bc.getQRCode().Create(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetQR(ctx context.Context, conf *config.Config, qrCodeID string) (*qrcode.Information, error) {
	resp, err := bc.getQRCode().Get(ctx, conf, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("get qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListQR(
	ctx context.Context,
	conf *config.Config,
	opts *qrcode.ListOptions,
) (*qrcode.ListResponse, error) {
	resp, err := bc.getQRCode().List(ctx, conf, opts)
	if err != nil {
		return nil, fmt.Errorf("list qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteQR(
	ctx context.Context,
	conf *config.Config,
	qrCodeID string,
) (*qrcode.SuccessResponse, error) {
	resp, err := bc.getQRCode().Delete(ctx, conf, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("delete qr: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateQR(
	ctx context.Context,
	conf *config.Config,
	req *qrcode.UpdateRequest,
) (*qrcode.SuccessResponse, error) {
	resp, err := bc.getQRCode().Update(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update qr: %w", err)
	}
	return resp, nil
}
