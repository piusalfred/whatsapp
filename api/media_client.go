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
	"github.com/piusalfred/whatsapp/media"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func (c *Client) UploadMedia(ctx context.Context, req *media.UploadRequest) (*media.UploadMediaResponse, error) {
	resp, err := c.sender.UploadMedia(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}
	return resp, nil
}

func (c *Client) GetMediaInfo(ctx context.Context, req *media.BaseRequest) (*media.Information, error) {
	resp, err := c.sender.GetMediaInfo(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get media info: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteMedia(ctx context.Context, req *media.BaseRequest) (*media.DeleteMediaResponse, error) {
	resp, err := c.sender.DeleteMedia(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}
	return resp, nil
}

func (c *Client) DownloadMedia(ctx context.Context, req *media.DownloadRequest, decoder whttp.ResponseDecoder) error {
	return c.sender.DownloadMedia(ctx, c.config, req, decoder)
}

func (c *Client) DownloadMediaByID(
	ctx context.Context,
	req *media.BaseRequest,
	decoder whttp.ResponseDecoder,
	opts ...media.DownloadOptionFunc,
) error {
	return c.sender.DownloadMediaByID(ctx, c.config, req, decoder, opts...)
}

func (bc *BaseClient) UploadMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.UploadRequest,
) (*media.UploadMediaResponse, error) {
	resp, err := bc.getMedia().Upload(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetMediaInfo(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
) (*media.Information, error) {
	resp, err := bc.getMedia().GetInfo(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get media info: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
) (*media.DeleteMediaResponse, error) {
	resp, err := bc.getMedia().Delete(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete media: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DownloadMedia(
	ctx context.Context,
	conf *config.Config,
	req *media.DownloadRequest,
	decoder whttp.ResponseDecoder,
) error {
	if err := bc.getMedia().Download(ctx, conf, req, decoder); err != nil {
		return fmt.Errorf("download media: %w", err)
	}
	return nil
}

func (bc *BaseClient) DownloadMediaByID(
	ctx context.Context,
	conf *config.Config,
	req *media.BaseRequest,
	decoder whttp.ResponseDecoder,
	opts ...media.DownloadOptionFunc,
) error {
	if err := bc.getMedia().DownloadByMediaID(ctx, conf, req, decoder, opts...); err != nil {
		return fmt.Errorf("download media by id: %w", err)
	}
	return nil
}
