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

package uploads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

var (
	ErrUploadStart    = errors.New("resumable upload start failed")
	ErrUploadContinue = errors.New("resumable upload continue failed")
	ErrUploadStatus   = errors.New("resumable upload status check failed")
)

// Client orchestrates high-level Resumable Upload API operations.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// NewClient creates a high-level Client for the Resumable Upload API.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[any](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[any]) {
	c.sender.SetSender(sender)
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.SetMiddlewares(mws...)
}

// Send dispatches a Request through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}

func (c *Client) InitUploadSession(
	ctx context.Context,
	req *InitUploadSessionRequest,
) (*InitUploadSessionResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeInitResumableUploadSession,
		FileName:    req.FileName,
		FileLength:  req.FileLength,
		FileType:    req.FileType,
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStart, err)
	}

	return &InitUploadSessionResponse{ID: resp.ID}, nil
}

func (c *Client) UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error) {
	request := &Request{
		RequestType:     whttp.RequestTypePerformResumableUpload,
		UploadSessionID: req.UploadSessionID,
		FileOffset:      req.FileOffset,
		FileReader:      req.FileReader,
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadContinue, err)
	}

	return &UploadChunkResponse{FileHandle: resp.FileHandle}, nil
}

func (c *Client) GetUploadStatus(ctx context.Context, uploadSessionID string) (*UploadStatusResponse, error) {
	request := &Request{
		RequestType:     whttp.RequestTypeGetResumableUploadSessionStatus,
		UploadSessionID: uploadSessionID,
	}

	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStatus, err)
	}

	return &UploadStatusResponse{ID: resp.ID, FileOffset: resp.FileOffset}, nil
}

// BaseClient is the low-level HTTP executor for the Resumable Upload API. It
// accepts a concrete [*config.Config] per request, making it suitable for
// multi-tenant SaaS scenarios. For a fixed-configuration client, use [Client].
type BaseClient struct {
	whttp.BaseClient[any]
}

type (
	InitUploadSessionRequest struct {
		FileName   string `json:"file_name"`
		FileLength int64  `json:"file_length"`
		FileType   string `json:"file_type"`
	}

	InitUploadSessionResponse struct {
		ID string `json:"id"` // format: upload:<UPLOAD_SESSION_ID>
	}

	UploadChunkRequest struct {
		UploadSessionID string
		FileOffset      int64
		FileReader      io.ReadSeeker
	}

	UploadChunkResponse struct {
		FileHandle string `json:"h"` // "h": "uploaded file handle"
	}

	UploadStatusResponse struct {
		ID         string `json:"id"`          // upload:<UPLOAD_SESSION_ID>
		FileOffset int64  `json:"file_offset"` // bytes uploaded so far
	}

	// Request is an internal unified context data carrier that maps upload
	// operations down to the HTTP executor. Fields tagged `json:"-"` are
	// routing metadata and are not serialized.
	Request struct {
		RequestType whttp.RequestType `json:"-"`

		// InitUploadSession fields
		FileName   string
		FileLength int64
		FileType   string

		// UploadChunk and GetUploadStatus fields
		UploadSessionID string
		FileOffset      int64
		FileReader      io.ReadSeeker
	}

	// BaseResponse acts as a flexible intermediate data capture layer for
	// unmarshaling varying upload response structures.
	BaseResponse struct {
		ID         string `json:"id,omitempty"`
		FileHandle string `json:"h,omitempty"`
		FileOffset int64  `json:"file_offset,omitempty"`
	}
)

// Send translates a Request into an HTTP transaction and returns the decoded BaseResponse.
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.RequestType)

	switch request.RequestType {
	case whttp.RequestTypeInitResumableUploadSession:
		queryParams := map[string]string{
			"file_name":    request.FileName,
			"file_length":  strconv.FormatInt(request.FileLength, 10),
			"file_type":    request.FileType,
			"access_token": conf.AccessToken,
		}
		b = b.QueryParams(queryParams).Endpoints(conf.APIVersion, conf.AppID, "uploads")

	case whttp.RequestTypePerformResumableUpload:
		if _, err := request.FileReader.Seek(request.FileOffset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("file seek: %w", err)
		}
		b = b.Bearer(conf.AccessToken).
			BodyReader(request.FileReader).
			Headers(map[string]string{
				"file_offset": strconv.FormatInt(request.FileOffset, 10),
			}).
			Endpoints(conf.APIVersion, request.UploadSessionID)

	case whttp.RequestTypeGetResumableUploadSessionStatus:
		b = whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
			Bearer(conf.AccessToken).
			DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
			Type(request.RequestType).
			Endpoints(conf.APIVersion, request.UploadSessionID)

	default:
		return nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}

	req := whttp.Build[any](b, nil)

	response := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		Flags: whttp.JSONDecodeDisallowEmptyResponse | whttp.JSONDecodeInspectResponseError,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	return response, nil
}

func (bc *BaseClient) InitUploadSession(
	ctx context.Context,
	conf *config.Config,
	req *InitUploadSessionRequest,
) (*InitUploadSessionResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeInitResumableUploadSession,
		FileName:    req.FileName,
		FileLength:  req.FileLength,
		FileType:    req.FileType,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStart, err)
	}

	return &InitUploadSessionResponse{ID: resp.ID}, nil
}

func (bc *BaseClient) UploadChunk(ctx context.Context, conf *config.Config, req *UploadChunkRequest) (
	*UploadChunkResponse, error,
) {
	request := &Request{
		RequestType:     whttp.RequestTypePerformResumableUpload,
		UploadSessionID: req.UploadSessionID,
		FileOffset:      req.FileOffset,
		FileReader:      req.FileReader,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadContinue, err)
	}

	return &UploadChunkResponse{FileHandle: resp.FileHandle}, nil
}

func (bc *BaseClient) GetUploadStatus(ctx context.Context, conf *config.Config, uploadSessionID string) (
	*UploadStatusResponse, error,
) {
	request := &Request{
		RequestType:     whttp.RequestTypeGetResumableUploadSessionStatus,
		UploadSessionID: uploadSessionID,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStatus, err)
	}

	return &UploadStatusResponse{ID: resp.ID, FileOffset: resp.FileOffset}, nil
}
