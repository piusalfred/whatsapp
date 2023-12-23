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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

var (
	ErrConfigNil        = errors.New("config is nil")
	ErrBadRequestFormat = errors.New("bad request")
)

const (
	MessageEndpoint = "messages"
)

const (
	templateMessageType = "template"
	textMessageType     = "text"
	reactionMessageType = "reaction"
	locationMessageType = "location"
	contactsMessageType = "contacts"
)

const (
	MaxAudioSize         = 16 * 1024 * 1024  // 16 MB
	MaxDocSize           = 100 * 1024 * 1024 // 100 MB
	MaxImageSize         = 5 * 1024 * 1024   // 5 MB
	MaxVideoSize         = 16 * 1024 * 1024  // 16 MB
	MaxStickerSize       = 100 * 1024        // 100 KB
	UploadedMediaTTL     = 30 * 24 * time.Hour
	MediaDownloadLinkTTL = 5 * time.Minute
)

const (
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeImage    MediaType = "image"
	MediaTypeSticker  MediaType = "sticker"
	MediaTypeVideo    MediaType = "video"
)

// MediaMaxAllowedSize returns the allowed maximum size for media. It returns
// -1 for unknown media type. Currently, it checks for MediaTypeAudio,MediaTypeVideo,
// MediaTypeImage, MediaTypeSticker,MediaTypeDocument.
func MediaMaxAllowedSize(mediaType MediaType) int {
	sizeMap := map[MediaType]int{
		MediaTypeAudio:    MaxAudioSize,
		MediaTypeDocument: MaxDocSize,
		MediaTypeSticker:  MaxStickerSize,
		MediaTypeImage:    MaxImageSize,
		MediaTypeVideo:    MaxVideoSize,
	}

	size, ok := sizeMap[mediaType]
	if ok {
		return size
	}

	return -1
}

type (

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,MediaInformation messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, except reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string
)

func (r *ResponseMessage) LogValue() slog.Value {
	if r == nil {
		return slog.StringValue("nil")
	}

	attr := []slog.Attr{
		slog.String("product", r.Product),
	}

	for i, message := range r.Messages {
		attr = append(attr, slog.String("message", fmt.Sprintf("%d.%s", i+1, message.ID)))
	}

	for i, contact := range r.Contacts {
		input := slog.String(fmt.Sprintf("contact.input.%d", i+1), contact.Input)
		waID := slog.String(fmt.Sprintf("contact.wa_id.%d", i+1), contact.WhatsappID)
		attr = append(attr, input, waID)
	}

	return slog.GroupValue(attr...)
}

var _ slog.LogValuer = (*ResponseMessage)(nil)

// SendMedia sends a media message to the recipient. Media can be sent using ID or Link. If using id, you must
// first upload your media asset to our servers and capture the returned media ID. If using link, your asset must
// be on a publicly accessible server or the message will fail to send.
func (client *Client) SendMedia(ctx context.Context, recipient string, req *MediaMessage,
	cacheOptions *CacheOptions,
) (*ResponseMessage, error) {
	request := &SendMediaRequest{
		BaseURL:       client.config.BaseURL,
		AccessToken:   client.config.AccessToken,
		PhoneNumberID: client.config.PhoneNumberID,
		ApiVersion:    client.config.Version,
		Recipient:     recipient,
		Type:          req.Type,
		MediaID:       req.MediaID,
		MediaLink:     req.MediaLink,
		Caption:       req.Caption,
		Filename:      req.Filename,
		Provider:      req.Provider,
		CacheOptions:  cacheOptions,
	}

	payload, err := formatMediaPayload(request)
	if err != nil {
		return nil, err
	}

	reqCtx := &whttp.RequestContext{
		Name:          "send media",
		BaseURL:       request.BaseURL,
		ApiVersion:    request.ApiVersion,
		PhoneNumberID: request.PhoneNumberID,
		Endpoints:     []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Bearer:  request.AccessToken,
		Headers: map[string]string{"Content-Type": "application/json"},
		Payload: payload,
	}

	if request.CacheOptions != nil {
		if request.CacheOptions.CacheControl != "" {
			params.Headers["Cache-Control"] = request.CacheOptions.CacheControl
		} else if request.CacheOptions.Expires > 0 {
			params.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", request.CacheOptions.Expires)
		}
		if request.CacheOptions.LastModified != "" {
			params.Headers["Last-Modified"] = request.CacheOptions.LastModified
		}
		if request.CacheOptions.ETag != "" {
			params.Headers["ETag"] = request.CacheOptions.ETag
		}
	}

	var message ResponseMessage

	err = client.bc.base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send media: %w", err)
	}

	return &message, nil
}

// SendInteractiveTemplate send an interactive template message which contains some buttons for user intraction.
// Interactive message templates expand the content you can send recipients beyond the standard message template
// and media messages template types to include interactive buttons using the components object. There are two types
// of predefined buttons:
//
//   - Call-to-Action — Allows your customer to call a phone number and visit a website.
//   - Quick Reply — Allows your customer to return a simple text message.
//
// These buttons can be attached to text messages or media messages. Once your interactive message templates have been
// created and approved, you can use them in notification messages as well as customer service/care messages.
func (client *Client) SendInteractiveTemplate(ctx context.Context, recipient string, req *InteractiveTemplateRequest) (
	*ResponseMessage, error,
) {
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewInteractiveTemplate(req.Name, tmpLanguage, req.Headers, req.Body, req.Buttons)
	payload := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          templateMessageType,
		Template:      template,
	}
	reqCtx := &whttp.RequestContext{
		Name:          "send template",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Endpoints:     []string{"messages"},
	}
	params := &whttp.Request{
		Method:  http.MethodPost,
		Payload: payload,
		Context: reqCtx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer: client.config.AccessToken,
	}
	var message ResponseMessage
	err := client.bc.base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return &message, nil
}
