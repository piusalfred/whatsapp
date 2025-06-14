package uploads

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	ErrUploadStart    = whatsapp.Error("resumable upload start failed")
	ErrUploadContinue = whatsapp.Error("resumable upload continue failed")
	ErrUploadStatus   = whatsapp.Error("resumable upload status check failed")
)

type BaseClient struct {
	ConfReader config.Reader
	Sender     whttp.AnySender
}

func NewBaseClient(confReader config.Reader, sender whttp.AnySender) *BaseClient {
	return &BaseClient{
		ConfReader: confReader,
		Sender:     sender,
	}
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
)

func (c *BaseClient) InitUploadSession(
	ctx context.Context,
	req *InitUploadSessionRequest,
) (*InitUploadSessionResponse, error) {
	conf, err := c.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrUploadStart, err)
	}

	queryParams := map[string]string{
		"file_name":    req.FileName,
		"file_length":  strconv.FormatInt(req.FileLength, 10),
		"file_type":    req.FileType,
		"access_token": conf.AccessToken,
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeInitResumableUploadSession),
		whttp.WithRequestQueryParams[any](queryParams),
		whttp.WithRequestEndpoints[any](conf.APIVersion, conf.AppID, "uploads"),
	}

	request := whttp.MakeRequest(http.MethodPost, conf.BaseURL, opts...)

	var resp InitUploadSessionResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStart, err)
	}

	return &resp, nil
}

func (c *BaseClient) UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error) {
	conf, err := c.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrUploadContinue, err)
	}

	_, err = req.FileReader.Seek(req.FileOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("%w: file seek: %w", ErrUploadContinue, err)
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestBodyReader[any](req.FileReader),
		whttp.WithRequestHeaders[any](map[string]string{
			"file_offset": strconv.FormatInt(req.FileOffset, 10),
		}),
		whttp.WithRequestType[any](whttp.RequestTypePerformResumableUpload),
		whttp.WithRequestEndpoints[any](conf.APIVersion, req.UploadSessionID),
	}

	request := whttp.MakeRequest(http.MethodPost, conf.BaseURL, opts...)

	var resp UploadChunkResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadContinue, err)
	}

	return &resp, nil
}

func (c *BaseClient) GetUploadStatus(ctx context.Context, uploadSessionID string) (*UploadStatusResponse, error) {
	conf, err := c.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrUploadStatus, err)
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestType[any](whttp.RequestTypeGetResumableUploadSessionStatus),
		whttp.WithRequestEndpoints[any](conf.APIVersion, uploadSessionID),
	}

	request := whttp.MakeRequest(http.MethodGet, conf.BaseURL, opts...)

	var resp UploadStatusResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = c.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUploadStatus, err)
	}

	return &resp, nil
}
