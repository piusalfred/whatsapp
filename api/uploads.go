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
	"github.com/piusalfred/whatsapp/uploads"
)

func (c *Client) InitUploadSession(
	ctx context.Context,
	req *uploads.InitUploadSessionRequest,
) (*uploads.InitUploadSessionResponse, error) {
	resp, err := c.sender.InitUploadSession(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("init upload session: %w", err)
	}
	return resp, nil
}

func (c *Client) UploadChunk(
	ctx context.Context,
	req *uploads.UploadChunkRequest,
) (*uploads.UploadChunkResponse, error) {
	resp, err := c.sender.UploadChunk(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("upload chunk: %w", err)
	}
	return resp, nil
}

func (c *Client) GetUploadStatus(ctx context.Context, uploadSessionID string) (*uploads.UploadStatusResponse, error) {
	resp, err := c.sender.GetUploadStatus(ctx, c.config, uploadSessionID)
	if err != nil {
		return nil, fmt.Errorf("get upload status: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) InitUploadSession(
	ctx context.Context,
	conf *config.Config,
	req *uploads.InitUploadSessionRequest,
) (*uploads.InitUploadSessionResponse, error) {
	resp, err := bc.getUploads().InitUploadSession(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("init upload session: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UploadChunk(
	ctx context.Context,
	conf *config.Config,
	req *uploads.UploadChunkRequest,
) (*uploads.UploadChunkResponse, error) {
	resp, err := bc.getUploads().UploadChunk(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("upload chunk: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetUploadStatus(
	ctx context.Context,
	conf *config.Config,
	uploadSessionID string,
) (*uploads.UploadStatusResponse, error) {
	resp, err := bc.getUploads().GetUploadStatus(ctx, conf, uploadSessionID)
	if err != nil {
		return nil, fmt.Errorf("get upload status: %w", err)
	}
	return resp, nil
}
