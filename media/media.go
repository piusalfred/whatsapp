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

//go:generate mockgen -destination=../mocks/media/mock_media.go -package=media -source=media.go

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/piusalfred/libwhatsapp/config"
	whttp "github.com/piusalfred/libwhatsapp/pkg/http"
)

// https://developers.facebook.com/docs/whatsapp/cloud-api/reference/media/

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

type (
	Type    string
	Service interface {
		Upload(ctx context.Context, req *UploadRequest) (*UploadMediaResponse, error)
		GetInfo(ctx context.Context, request *BaseRequest) (*Information, error)
		Delete(ctx context.Context, request *BaseRequest) (*DeleteMediaResponse, error)
		Download(ctx context.Context, request *DownloadRequest, decoder whttp.ResponseDecoder) error
	}

	DownloadRequest struct {
		URL     string
		Retries int
	}

	UploadRequest struct {
		MediaType Type
		Filename  string
		Reader    io.Reader
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
		Sender     whttp.Sender[any]
	}

	BaseRequest struct {
		MediaID            string
		RestrictToOwnMedia bool
		PhoneNumberID      string
	}
)

func (s *BaseClient) Download(ctx context.Context, request *DownloadRequest, decoder whttp.ResponseDecoder) error {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return err
	}
	req := &whttp.Request[any]{
		Type:    whttp.RequestTypeDownloadMedia,
		Method:  http.MethodGet,
		BaseURL: request.URL,
		Bearer:  conf.AccessToken,
	}

	for i := 0; i <= request.Retries; i++ {
		if err := s.Sender.Send(ctx, req, decoder); err != nil {
			if i < request.Retries {
				continue
			}
			return fmt.Errorf("download media failed after %d attempts: %w", request.Retries+1, err)
		}

		return nil
	}

	return fmt.Errorf("download retries exceeded")
}

func (s *BaseClient) Delete(ctx context.Context, req *BaseRequest) (*DeleteMediaResponse, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	phoneNumberID := req.PhoneNumberID
	if phoneNumberID == "" && req.RestrictToOwnMedia {
		phoneNumberID = conf.PhoneNumberID
	}

	queryParams := map[string]string{}
	if phoneNumberID != "" {
		queryParams["phone_number_id"] = phoneNumberID
	}

	request := &whttp.Request[any]{
		Type:        whttp.RequestTypeDeleteMedia,
		Method:      http.MethodDelete,
		Bearer:      conf.AccessToken,
		BaseURL:     conf.BaseURL,
		QueryParams: queryParams,
		Endpoints:   []string{conf.APIVersion, req.MediaID},
	}

	var resp DeleteMediaResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
	})

	if err := s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("delete media failed: %w", err)
	}

	return &resp, nil
}

func (s *BaseClient) GetInfo(ctx context.Context, req *BaseRequest) (*Information, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	phoneNumberID := req.PhoneNumberID
	if phoneNumberID == "" && req.RestrictToOwnMedia {
		phoneNumberID = conf.PhoneNumberID
	}

	queryParams := map[string]string{}
	if phoneNumberID != "" {
		queryParams["phone_number_id"] = phoneNumberID
	}

	request := &whttp.Request[any]{
		Type:        whttp.RequestTypeGetMedia,
		Method:      http.MethodGet,
		Bearer:      conf.AccessToken,
		QueryParams: queryParams,
		BaseURL:     conf.BaseURL,
		Endpoints:   []string{conf.APIVersion, req.MediaID},
	}

	var info Information
	decoder := whttp.ResponseDecoderJSON(&info, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
	})

	if err := s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("get media info failed: %w", err)
	}

	return &info, nil
}

func (s *BaseClient) Upload(ctx context.Context, req *UploadRequest) (*UploadMediaResponse, error) {
	conf, err := s.ConfReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	request := &whttp.Request[any]{
		Type:        whttp.RequestTypeUploadMedia,
		Method:      http.MethodPost,
		Bearer:      conf.AccessToken,
		QueryParams: nil,
		BaseURL:     conf.BaseURL,
		Endpoints:   []string{conf.APIVersion, conf.PhoneNumberID, "media"},
		Metadata:    nil,
		Message:     nil,
		Form: &whttp.RequestForm{
			Fields: map[string]string{
				"type":              string(req.MediaType),
				"messaging_product": "whatsapp",
			},
			FormFile: &whttp.FormFile{
				Name: "file",
				Path: req.Filename,
			},
		},
	}

	var resp UploadMediaResponse
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
	})

	if err := s.Sender.Send(ctx, request, decoder); err != nil {
		return nil, fmt.Errorf("upload media failed: %w", err)
	}

	return &resp, nil
}
