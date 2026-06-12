/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package media provides a client for the WhatsApp Business Media API.
//
// The Media API lets businesses upload, retrieve, download, and delete media
// files such as images, audio, documents, stickers, and video.
//
// # Getting Started
//
// Create a [Client] using [NewClient] with a [config.Config] and optional sender options:
//
//	conf := &config.Config{
//	    BaseURL:       "https://graph.facebook.com",
//	    APIVersion:    "v22.0",
//	    AccessToken:   "YOUR_ACCESS_TOKEN",
//	    PhoneNumberID: "YOUR_PHONE_NUMBER_ID",
//	}
//
//	client := media.NewClient(conf,
//	    whttp.WithSenderHTTPClient(http.DefaultClient),
//	    whttp.WithSenderTimeout(30*time.Second),
//	)
//
// # Uploading Media
//
//	resp, err := client.Upload(ctx, &media.UploadRequest{
//	    MediaType: media.TypeImageJPEG,
//	    Filepath:  "/path/to/image.jpg",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Media ID:", resp.ID)
//
// # Retrieving Media Info
//
//	info, err := client.GetInfo(ctx, &media.BaseRequest{MediaID: resp.ID})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("URL:", info.URL) // URL is valid for 5 minutes
//
// # Downloading Media
//
//	err := client.DownloadByMediaID(ctx, &media.BaseRequest{MediaID: resp.ID},
//	    media.ResponseDecoderFunc(func(ctx context.Context, resp *http.Response) error {
//	        // handle response body
//	        return nil
//	    }),
//	    media.WithDownloadRetries(2),
//	)
//
// # Deleting Media
//
//	delResp, err := client.Delete(ctx, &media.BaseRequest{MediaID: resp.ID})
//
// # Configuration Options
//
// [whttp.CoreSenderOption] functions customize the underlying HTTP transport:
//
//	whttp.WithSenderHTTPClient(customHTTPClient)
//	whttp.WithSenderRequestInterceptor(myRequestHook)
//	whttp.WithSenderResponseInterceptor(myResponseHook)
//	whttp.WithSenderTimeout(30 * time.Second)
//	whttp.WithSenderMaxBodyBytes(10 << 20)
//	whttp.WithSenderMaxHeaderBytes(1 << 20)
//
// # Testing
//
// For unit tests, inject a mock sender via [Client.SetBaseClient]:
//
//	client := media.NewClient(conf)
//	client.SetBaseClient(mockSender)
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
	Type string

	// DownloadRequest carries the parameters for downloading media from a URL.
	DownloadRequest struct {
		URL     string
		Retries int
	}

	// UploadRequest carries the parameters for uploading media.
	UploadRequest struct {
		MediaType Type
		Filepath  string
	}

	// UploadMediaResponse is returned on a successful upload.
	UploadMediaResponse struct {
		ID string `json:"id"` // ID of the uploaded media
	}

	// Information describes a media object retrieved from the API.
	Information struct {
		MessagingProduct string `json:"messaging_product"`
		URL              string `json:"url"`
		MimeType         string `json:"mime_type"`
		SHA256           string `json:"sha256"`
		FileSize         int64  `json:"file_size"`
		ID               string `json:"id"`
	}

	// DeleteMediaResponse indicates whether a delete operation succeeded.
	DeleteMediaResponse struct {
		Success bool `json:"success"`
	}

	// DownloadMediaResponse holds the binary content and MIME type of downloaded media.
	DownloadMediaResponse struct {
		FileContent []byte
		ContentType string
	}

	// Client is a high-level client bound to a fixed [config.Config].
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	// BaseClient is the low-level HTTP executor for the Media API. It accepts a
	// concrete [*config.Config] per request, making it suitable for multi-tenant
	// SaaS scenarios. For a fixed-configuration client, use [Client].
	BaseClient struct {
		whttp.BaseClient[any]
	}

	// Request is an internal unified context data carrier mapping operation
	// metadata down to the HTTP executor.
	Request struct {
		Type          whttp.RequestType
		MediaID       string
		PhoneNumberID string
		Form          *whttp.RequestForm
	}

	// BaseResponse acts as a flexible intermediate data capture layer unmarshaling
	// varying response structures across disparate HTTP verbs.
	BaseResponse struct {
		ID               string `json:"id"`
		Success          bool   `json:"success"`
		URL              string `json:"url"`
		MimeType         string `json:"mime_type"`
		SHA256           string `json:"sha256"`
		FileSize         int64  `json:"file_size"`
		MessagingProduct string `json:"messaging_product"`
	}

	// BaseRequest carries the parameters for media operations that target a
	// specific media object.
	BaseRequest struct {
		MediaID            string
		RestrictToOwnMedia bool
		PhoneNumberID      string
	}
)

// NewClient creates a high-level [Client] with a fixed configuration.
// Optional [SenderOption] functions tune the underlying HTTP transport.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[any](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[any]) {
	c.sender.Sender = sender
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.Sender = whttp.WrapMiddlewareSender(c.sender.Sender, mws...)
}

// Upload uploads a media file to WhatsApp.
func (c *Client) Upload(ctx context.Context, req *UploadRequest) (*UploadMediaResponse, error) {
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
	request := &Request{
		Type: whttp.RequestTypeUploadMedia,
		Form: form,
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaUpload, err)
	}
	return resp.ToUploadMediaResponse(), nil
}

// GetInfo retrieves metadata and a temporary URL for a media object.
func (c *Client) GetInfo(ctx context.Context, req *BaseRequest) (*Information, error) {
	request := &Request{
		Type:          whttp.RequestTypeGetMedia,
		MediaID:       req.MediaID,
		PhoneNumberID: resolvePhoneNumberID(req, c.config),
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("get media info failed: %w", err)
	}
	return resp.ToInformation(), nil
}

// Delete removes a media object by ID.
func (c *Client) Delete(ctx context.Context, req *BaseRequest) (*DeleteMediaResponse, error) {
	request := &Request{
		Type:          whttp.RequestTypeDeleteMedia,
		MediaID:       req.MediaID,
		PhoneNumberID: resolvePhoneNumberID(req, c.config),
	}
	resp, err := c.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaDelete, err)
	}
	return resp.ToDeleteMediaResponse(), nil
}

// Send dispatches a raw [Request] through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return response, nil
}

// Download downloads media from a pre-signed URL using a custom decoder.
func (c *Client) Download(ctx context.Context, request *DownloadRequest, decoder whttp.ResponseDecoder) error {
	return c.sender.Download(ctx, c.config, request, decoder)
}

// DownloadByMediaID retrieves the media info and then downloads the file.
func (c *Client) DownloadByMediaID(
	ctx context.Context,
	request *BaseRequest,
	decoder whttp.ResponseDecoder,
	options ...DownloadOptionFunc,
) error {
	return c.sender.DownloadByMediaID(ctx, c.config, request, decoder, options...)
}

// ToUploadMediaResponse attempts to coerce a BaseResponse into an UploadMediaResponse.
func (r *BaseResponse) ToUploadMediaResponse() *UploadMediaResponse {
	return &UploadMediaResponse{ID: r.ID}
}

// ToInformation attempts to coerce a BaseResponse into an Information.
func (r *BaseResponse) ToInformation() *Information {
	return &Information{
		MessagingProduct: r.MessagingProduct,
		URL:              r.URL,
		MimeType:         r.MimeType,
		SHA256:           r.SHA256,
		FileSize:         r.FileSize,
		ID:               r.ID,
	}
}

// ToDeleteMediaResponse attempts to coerce a BaseResponse into a DeleteMediaResponse.
func (r *BaseResponse) ToDeleteMediaResponse() *DeleteMediaResponse {
	return &DeleteMediaResponse{Success: r.Success}
}

// resolvePhoneNumberID returns the phone number ID to use for the request,
// falling back to the config value when RestrictToOwnMedia is set.
func resolvePhoneNumberID(req *BaseRequest, conf *config.Config) string {
	if req.PhoneNumberID != "" {
		return req.PhoneNumberID
	}
	if req.RestrictToOwnMedia {
		return conf.PhoneNumberID
	}
	return ""
}

// Send translates a high-level [Request] into an HTTP transaction and returns
// the decoded [BaseResponse].
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	var method string

	switch request.Type {
	case whttp.RequestTypeUploadMedia:
		method = http.MethodPost
	case whttp.RequestTypeGetMedia:
		method = http.MethodGet
	case whttp.RequestTypeDeleteMedia:
		method = http.MethodDelete
	}

	queryParams := map[string]string{}
	if request.PhoneNumberID != "" {
		queryParams["phone_number_id"] = request.PhoneNumberID
	}

	bld := whttp.NewRequestBuilder(method, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.Type)

	if request.MediaID != "" {
		bld = bld.Endpoints(conf.APIVersion, request.MediaID)
	} else if request.Type == whttp.RequestTypeUploadMedia {
		bld = bld.Endpoints(conf.APIVersion, conf.PhoneNumberID, "media")
	}

	if len(queryParams) > 0 {
		bld = bld.QueryParams(queryParams)
	}

	if request.Form != nil {
		bld = bld.Form(request.Form)
	}

	req := whttp.Build[any](bld, nil)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// Upload uploads a media file to WhatsApp.
func (bc *BaseClient) Upload(
	ctx context.Context,
	conf *config.Config,
	req *UploadRequest,
) (*UploadMediaResponse, error) {
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

	request := &Request{
		Type: whttp.RequestTypeUploadMedia,
		Form: form,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaUpload, err)
	}

	return resp.ToUploadMediaResponse(), nil
}

// GetInfo retrieves metadata and a temporary URL for a media object.
func (bc *BaseClient) GetInfo(ctx context.Context, conf *config.Config, req *BaseRequest) (*Information, error) {
	request := &Request{
		Type:          whttp.RequestTypeGetMedia,
		MediaID:       req.MediaID,
		PhoneNumberID: resolvePhoneNumberID(req, conf),
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get media info failed: %w", err)
	}

	return resp.ToInformation(), nil
}

// Delete removes a media object by ID.
func (bc *BaseClient) Delete(ctx context.Context, conf *config.Config, req *BaseRequest) (*DeleteMediaResponse, error) {
	request := &Request{
		Type:          whttp.RequestTypeDeleteMedia,
		MediaID:       req.MediaID,
		PhoneNumberID: resolvePhoneNumberID(req, conf),
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMediaDelete, err)
	}

	return resp.ToDeleteMediaResponse(), nil
}

// Download downloads media from a pre-signed URL using a custom decoder.
func (bc *BaseClient) Download(
	ctx context.Context,
	conf *config.Config,
	req *DownloadRequest,
	decoder whttp.ResponseDecoder,
) error {
	bld := whttp.NewRequestBuilder(http.MethodGet, "").
		Bearer(conf.AccessToken).
		Type(whttp.RequestTypeDownloadMedia).
		DownloadURL(req.URL).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	request := whttp.Build[any](bld, nil)

	for i := 0; i <= req.Retries; i++ {
		if err := bc.Sender.Send(ctx, request, decoder); err != nil {
			if i < req.Retries {
				continue
			}

			return fmt.Errorf("%w: %d attempts: %w", ErrMediaDownload, req.Retries+1, err)
		}

		return nil
	}

	return fmt.Errorf("%w: %d attempts", ErrMediaDownload, req.Retries+1)
}

type downloadOption struct {
	Retries int
}

// DownloadOptionFunc configures retry behaviour for [BaseClient.DownloadByMediaID].
type DownloadOptionFunc func(*downloadOption)

// WithDownloadRetries sets the number of retries for a download operation.
func WithDownloadRetries(retries int) DownloadOptionFunc {
	return func(opt *downloadOption) {
		opt.Retries = retries
	}
}

// DefaultRetries is the default number of download retries.
const DefaultRetries = 2

// DownloadByMediaID retrieves the media info and then downloads the file.
func (bc *BaseClient) DownloadByMediaID(
	ctx context.Context,
	conf *config.Config,
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

	info, err := bc.GetInfo(ctx, conf, request)
	if err != nil {
		return fmt.Errorf("retrieve info for media(id: %s) %w: %w", request.MediaID, ErrMediaDownload, err)
	}

	downloadRequest := &DownloadRequest{
		URL:     info.URL,
		Retries: dOptions.Retries,
	}

	if err = bc.Download(ctx, conf, downloadRequest, decoder); err != nil {
		return fmt.Errorf("download media(id: %s) %w: %w", request.MediaID, ErrMediaDownload, err)
	}

	return nil
}

func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[any]) {
	bc.Sender = whttp.WrapMiddlewareSender(bc.Sender, mws...)
}
