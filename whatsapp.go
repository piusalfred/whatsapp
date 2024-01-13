/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package whatsapp

import (
	"context"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

// var (
//
//	ErrConfigNil        = errors.New("config is nil")
//	ErrBadRequestFormat = errors.New("bad request")
//
// )
//
// const (
//
//	MessageStatusRead         = "read"
//	MessageEndpoint           = "messages"
//	MessagingProduct          = "whatsapp"
//	RecipientTypeIndividual   = "individual"
//	BaseURL                   = "https://graph.facebook.com/"
//	LowestSupportedVersion    = "v16.0"
//	DateFormatContactBirthday = time.DateOnly // YYYY-MM-DD
//
// )
//
// const (
//
//	MaxAudioSize         = 16 * 1024 * 1024  // 16 MB
//	MaxDocSize           = 100 * 1024 * 1024 // 100 MB
//	MaxImageSize         = 5 * 1024 * 1024   // 5 MB
//	MaxVideoSize         = 16 * 1024 * 1024  // 16 MB
//	MaxStickerSize       = 100 * 1024        // 100 KB
//	UploadedMediaTTL     = 30 * 24 * time.Hour
//	MediaDownloadLinkTTL = 5 * time.Minute
//
// )
//
// const (
//
//	MediaTypeAudio    MediaType = "audio"
//	MediaTypeDocument MediaType = "document"
//	MediaTypeImage    MediaType = "image"
//	MediaTypeSticker  MediaType = "sticker"
//	MediaTypeVideo    MediaType = "video"
//
// )
//
// // MediaMaxAllowedSize returns the allowed maximum size for media. It returns
// // -1 for unknown media type. Currently, it checks for MediaTypeAudio,MediaTypeVideo,
// // MediaTypeImage, MediaTypeSticker,MediaTypeDocument.
//
//	func MediaMaxAllowedSize(mediaType MediaType) int {
//		sizeMap := map[MediaType]int{
//			MediaTypeAudio:    MaxAudioSize,
//			MediaTypeDocument: MaxDocSize,
//			MediaTypeSticker:  MaxStickerSize,
//			MediaTypeImage:    MaxImageSize,
//			MediaTypeVideo:    MaxVideoSize,
//		}
//
//		size, ok := sizeMap[mediaType]
//		if ok {
//			return size
//		}
//
//		return -1
//	}
type (
	statusUpdateRequest struct {
		MessagingProduct string `json:"messaging_product,omitempty"` // always whatsapp
		Status           string `json:"status,omitempty"`            // always read
		MessageID        string `json:"message_id,omitempty"`
	}

	RequestParams struct {
		ID        string
		Metadata  map[string]string
		Recipient string
		ReplyID   string
	}
)

// MessageSender ..
type MessageSender interface {
	Text(ctx context.Context, params *RequestParams, text *models.Text) (*whttp.ResponseMessage, error)
	React(ctx context.Context, params *RequestParams, reaction *models.Reaction) (*whttp.ResponseMessage, error)
	Contacts(ctx context.Context, params *RequestParams, contacts []*models.Contact) (*whttp.ResponseMessage, error)
	Location(ctx context.Context, params *RequestParams, request *models.Location) (*whttp.ResponseMessage, error)
	InteractiveMessage(ctx context.Context, params *RequestParams,
		interactive *models.Interactive) (*whttp.ResponseMessage, error)
	Template(ctx context.Context, params *RequestParams, template *models.Template) (*whttp.ResponseMessage, error)
}

// MediaSender ..
type MediaSender interface {
	Image(ctx context.Context, params *RequestParams, image *models.Image,
		options *whttp.CacheOptions) (*whttp.ResponseMessage, error)
	Audio(ctx context.Context, params *RequestParams, audio *models.Audio,
		options *whttp.CacheOptions) (*whttp.ResponseMessage, error)
	Video(ctx context.Context, params *RequestParams, video *models.Video,
		options *whttp.CacheOptions) (*whttp.ResponseMessage, error)
	Document(ctx context.Context, params *RequestParams, document *models.Document,
		options *whttp.CacheOptions) (*whttp.ResponseMessage, error)
	Sticker(ctx context.Context, params *RequestParams, sticker *models.Sticker,
		options *whttp.CacheOptions) (*whttp.ResponseMessage, error)
}

var (
	_ MessageSender = (*Client)(nil)
	_ MediaSender   = (*Client)(nil)
)
