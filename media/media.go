/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package media

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	TypeAudioAAC        Type = "audio/aac"
	TypeAudioAMR        Type = "audio/amr"
	TypeAudioMP3        Type = "audio/mpeg"
	TypeAudioMP4        Type = "audio/mp4"
	TypeAudioOGG        Type = "audio/ogg" // OPUS codecs only
	TypeDocText         Type = "text/plain"
	TypeDocExcelXLS     Type = "application/vnd.ms-excel"
	TypeDocExcelXLSX    Type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	TypeDocWordDOC      Type = "application/msword"
	TypeDocWordDOCX     Type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	TypeDocPPT          Type = "application/vnd.ms-powerpoint"
	TypeDocPPTX         Type = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	TypeDocPDF          Type = "application/pdf"
	TypeImageJPEG       Type = "image/jpeg"
	TypeImagePNG        Type = "image/png"
	TypeStickerStatic   Type = "image/webp"
	TypeStickerAnimated Type = "image/webp"
	TypeVideo3GPP       Type = "video/3gp"
	TypeVideoMP4        Type = "video/mp4"
)

const (
	AudioMaxSize           = 16 * 1024 * 1024  // 16MB
	DocMaxSize             = 100 * 1024 * 1024 // 100MB
	ImageMaxSize           = 5 * 1024 * 1024
	StickerStaticMaxSize   = 100 * 1024
	StickerAnimatedMaxSize = 500 * 1024
	VideoMaxSize           = 16 * 1024 * 1024
)

const (
	ErrMediaUpload   = whatsapp.Error("media upload failed")
	ErrMediaDownload = whatsapp.Error("media download failed")
	ErrMediaDelete   = whatsapp.Error("media delete failed")
	ErrMediaGetInfo  = whatsapp.Error("media get info failed")
)

type Category string

const (
	CategoryAudio    Category = "audio"
	CategoryDocument Category = "document"
	CategoryImage    Category = "image"
	CategorySticker  Category = "sticker"
	CategoryVideo    Category = "video"
)

type Info struct {
	MediaType Type
	MIMEType  string
	MaxSize   int64
	Extension string
	Category  Category
}

var InfoMap = map[Type]Info{ //nolint:gochecknoglobals // ok
	TypeAudioAAC: {
		MediaType: TypeAudioAAC,
		MIMEType:  string(TypeAudioAAC),
		MaxSize:   AudioMaxSize,
		Extension: ".aac",
		Category:  CategoryAudio,
	},
	TypeAudioAMR: {
		MediaType: TypeAudioAMR,
		MIMEType:  string(TypeAudioAMR),
		MaxSize:   AudioMaxSize,
		Extension: ".amr",
		Category:  CategoryAudio,
	},
	TypeAudioMP3: {
		MediaType: TypeAudioMP3,
		MIMEType:  string(TypeAudioMP3),
		MaxSize:   AudioMaxSize,
		Extension: ".mp3",
		Category:  CategoryAudio,
	},
	TypeAudioMP4: {
		MediaType: TypeAudioMP4,
		MIMEType:  string(TypeAudioMP4),
		MaxSize:   AudioMaxSize,
		Extension: ".m4a",
		Category:  CategoryAudio,
	},
	TypeAudioOGG: {
		MediaType: TypeAudioOGG,
		MIMEType:  string(TypeAudioOGG),
		MaxSize:   AudioMaxSize,
		Extension: ".ogg",
		Category:  CategoryAudio,
	},
	TypeDocText: {
		MediaType: TypeDocText,
		MIMEType:  string(TypeDocText),
		MaxSize:   DocMaxSize,
		Extension: ".txt",
		Category:  CategoryDocument,
	},
	TypeDocExcelXLS: {
		MediaType: TypeDocExcelXLS,
		MIMEType:  string(TypeDocExcelXLS),
		MaxSize:   DocMaxSize,
		Extension: ".xls",
		Category:  CategoryDocument,
	},
	TypeDocExcelXLSX: {
		MediaType: TypeDocExcelXLSX,
		MIMEType:  string(TypeDocExcelXLSX),
		MaxSize:   DocMaxSize,
		Extension: ".xlsx",
		Category:  CategoryDocument,
	},
	TypeDocWordDOC: {
		MediaType: TypeDocWordDOC,
		MIMEType:  string(TypeDocWordDOC),
		MaxSize:   DocMaxSize,
		Extension: ".doc",
		Category:  CategoryDocument,
	},
	TypeDocWordDOCX: {
		MediaType: TypeDocWordDOCX,
		MIMEType:  string(TypeDocWordDOCX),
		MaxSize:   DocMaxSize,
		Extension: ".docx",
		Category:  CategoryDocument,
	},
	TypeDocPPT: {
		MediaType: TypeDocPPT,
		MIMEType:  string(TypeDocPPT),
		MaxSize:   DocMaxSize,
		Extension: ".ppt",
		Category:  CategoryDocument,
	},
	TypeDocPPTX: {
		MediaType: TypeDocPPTX,
		MIMEType:  string(TypeDocPPTX),
		MaxSize:   DocMaxSize,
		Extension: ".pptx",
		Category:  CategoryDocument,
	},
	TypeDocPDF: {
		MediaType: TypeDocPDF,
		MIMEType:  string(TypeDocPDF),
		MaxSize:   DocMaxSize,
		Extension: ".pdf",
		Category:  CategoryDocument,
	},
	TypeImageJPEG: {
		MediaType: TypeImageJPEG,
		MIMEType:  string(TypeImageJPEG),
		MaxSize:   ImageMaxSize,
		Extension: ".jpeg",
		Category:  CategoryImage,
	},
	TypeImagePNG: {
		MediaType: TypeImagePNG,
		MIMEType:  string(TypeImagePNG),
		MaxSize:   ImageMaxSize,
		Extension: ".png",
		Category:  CategoryImage,
	},
	TypeStickerStatic: {
		MediaType: TypeStickerStatic,
		MIMEType:  string(TypeStickerStatic),
		MaxSize:   StickerStaticMaxSize,
		Extension: ".webp",
		Category:  CategorySticker,
	},
	TypeVideo3GPP: {
		MediaType: TypeVideo3GPP,
		MIMEType:  string(TypeVideo3GPP),
		MaxSize:   VideoMaxSize,
		Extension: ".3gp",
		Category:  CategoryVideo,
	},
	TypeVideoMP4: {
		MediaType: TypeVideoMP4,
		MIMEType:  string(TypeVideoMP4),
		MaxSize:   VideoMaxSize,
		Extension: ".mp4",
		Category:  CategoryVideo,
	},
}

func (mt Type) String() string {
	return string(mt)
}

type (
	Type    string
	Service interface {
		Upload(ctx context.Context, req *UploadRequest) (*UploadMediaResponse, error)
		GetInfo(ctx context.Context, request *BaseRequest) (*Information, error)
		Delete(ctx context.Context, request *BaseRequest) (*DeleteMediaResponse, error)
		Download(ctx context.Context, request *DownloadRequest, decoder whttp.ResponseDecoder) error
		DownloadByMediaID(
			ctx context.Context,
			request *BaseRequest,
			decoder whttp.ResponseDecoder,
			options ...DownloadOptionFunc,
		) error
	}

	DownloadRequest struct {
		URL     string
		Retries int
	}

	UploadRequest struct {
		MediaType Type
		Filepath  string
	}

	UploadMediaResponse struct {
		ID string `json:"id"` // ID of the uploaded media
	}

	Information struct {
		MessagingProduct string `json:"messaging_product"`
		URL              string `json:"url"`
		MimeType         string `json:"mime_type"`
		SHA256           string `json:"sha256"`
		FileSize         int64  `json:"file_size"`
		ID               string `json:"id"`
	}

	DeleteMediaResponse struct {
		Success bool `json:"success"`
	}

	DownloadMediaResponse struct {
		FileContent []byte
		ContentType string
	}

	BaseClient struct {
		ConfReader config.Reader
		Sender     whttp.AnySender
	}

	BaseRequest struct {
		MediaID            string
		RestrictToOwnMedia bool
		PhoneNumberID      string
	}
)

func NewBaseClient(confReader config.Reader, sender whttp.AnySender) *BaseClient {
	return &BaseClient{
		ConfReader: confReader,
		Sender:     sender,
	}
}

func (s *BaseClient) Download(ctx context.Context, request *DownloadRequest, decoder whttp.ResponseDecoder) error {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return fmt.Errorf("%w: config read: %w", ErrMediaDownload, err)
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestSecured[any](false),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestType[any](whttp.RequestTypeDownloadMedia),
	}

	req := whttp.MakeDownloadRequest[any](request.URL, opts...)

	for i := 0; i <= request.Retries; i++ {
		if err = s.Sender.Send(ctx, req, decoder); err != nil {
			if i < request.Retries {
				continue
			}

			return fmt.Errorf("%w: %d attempts: %w", ErrMediaDownload, request.Retries+1, err)
		}

		return nil
	}

	return fmt.Errorf("%w: %d attempts", ErrMediaDownload, request.Retries+1)
}

func (s *BaseClient) Delete(ctx context.Context, req *BaseRequest) (*DeleteMediaResponse, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrMediaDelete, err)
	}

	phoneNumberID := req.PhoneNumberID
	if phoneNumberID == "" && req.RestrictToOwnMedia {
		phoneNumberID = conf.PhoneNumberID
	}

	queryParams := map[string]string{}
	if phoneNumberID != "" {
		queryParams["phone_number_id"] = phoneNumberID
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestType[any](whttp.RequestTypeDeleteMedia),
		whttp.WithRequestQueryParams[any](queryParams),
		whttp.WithRequestEndpoints[any](conf.APIVersion, req.MediaID),
	}

	request := whttp.MakeRequest[any](http.MethodDelete, conf.BaseURL, opts...)

	var resp DeleteMediaResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaDelete, err)
	}

	return &resp, nil
}

func (s *BaseClient) GetInfo(ctx context.Context, req *BaseRequest) (*Information, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrMediaGetInfo, err)
	}

	phoneNumberID := req.PhoneNumberID
	if phoneNumberID == "" && req.RestrictToOwnMedia {
		phoneNumberID = conf.PhoneNumberID
	}

	queryParams := map[string]string{}
	if phoneNumberID != "" {
		queryParams["phone_number_id"] = phoneNumberID
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestType[any](whttp.RequestTypeGetMedia),
		whttp.WithRequestQueryParams[any](queryParams),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestEndpoints[any](conf.APIVersion, req.MediaID),
	}

	request := whttp.MakeRequest[any](http.MethodGet, conf.BaseURL, opts...)

	var info Information
	decoder := whttp.ResponseDecoderJSON(&info, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("get media info failed: %w", err)
	}

	return &info, nil
}

func (s *BaseClient) Upload(ctx context.Context, req *UploadRequest) (*UploadMediaResponse, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrMediaUpload, err)
	}

	_, ok := InfoMap[req.MediaType]
	isAnimated := req.MediaType == TypeStickerAnimated
	if !ok && !isAnimated {
		return nil, fmt.Errorf("%w: media type not supported", ErrMediaUpload)
	}

	fp, err := filepath.Abs(req.Filepath)
	if err != nil {
		return nil, fmt.Errorf("determine absolute filepath for %s: %w: %w", req.Filepath, ErrMediaUpload, err)
	}

	form := &whttp.RequestForm{
		Fields: map[string]string{
			"type":              string(req.MediaType),
			"messaging_product": "whatsapp",
		},
		FormFile: &whttp.FormFile{
			Name: "file",
			Path: fp,
			Type: req.MediaType.String(),
		},
	}

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestAppSecret[any](conf.AppSecret),
		whttp.WithRequestSecured[any](conf.SecureRequests),
		whttp.WithRequestBearer[any](conf.AccessToken),
		whttp.WithRequestType[any](whttp.RequestTypeUploadMedia),
		whttp.WithRequestEndpoints[any](conf.APIVersion, conf.PhoneNumberID, "media"),
		whttp.WithRequestForm[any](form),
	}

	request := whttp.MakeRequest[any](http.MethodPost, conf.BaseURL, opts...)

	var resp UploadMediaResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaUpload, err)
	}

	return &resp, nil
}

type downloadOption struct {
	Retries int
}

type DownloadOptionFunc func(*downloadOption)

func WithDownloadRetries(retries int) DownloadOptionFunc {
	return func(opt *downloadOption) {
		opt.Retries = retries
	}
}

const DefaultRetries = 2

// DownloadByMediaID downloads media by its ID.
func (s *BaseClient) DownloadByMediaID(
	ctx context.Context,
	request *BaseRequest,
	decoder whttp.ResponseDecoder,
	options ...DownloadOptionFunc,
) error {
	dOptions := &downloadOption{
		Retries: DefaultRetries,
	}
	for _, opt := range options {
		opt(dOptions)
	}

	info, err := s.GetInfo(ctx, request)
	if err != nil {
		return fmt.Errorf("retrieve info for media(id: %s) %w: %w", request.MediaID, ErrMediaDownload, err)
	}

	downloadRequest := &DownloadRequest{
		URL:     info.URL,
		Retries: dOptions.Retries,
	}

	if err = s.Download(ctx, downloadRequest, decoder); err != nil {
		return fmt.Errorf("download media(id: %s) %w: %w", request.MediaID, ErrMediaDownload, err)
	}

	return nil
}
