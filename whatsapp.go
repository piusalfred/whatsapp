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
	"net/http"
	"time"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

var ErrBadRequestFormat = errors.New("bad request")

const (
	messagingProduct          = "whatsapp"
	individualRecipientType   = "individual"
	BaseURL                   = "https://graph.facebook.com/"
	LowestSupportedVersion    = "v16.0"
	ContactBirthDayDateFormat = "2006-01-02" // YYYY-MM-DD
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
	ResponseMessage struct {
		Product  string             `json:"messaging_product,omitempty"`
		Contacts []*ResponseContact `json:"contacts,omitempty"`
		Messages []*MessageID       `json:"messages,omitempty"`
	}
	MessageID struct {
		ID string `json:"id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappID string `json:"wa_id"`
	}

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,MediaInformation messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, except reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string
)

type MediaMessage struct {
	Type      MediaType
	MediaID   string
	MediaLink string
	Caption   string
	Filename  string
	Provider  string
}

// SendMedia sends a media message to the recipient. Media can be sent using ID or Link. If using id, you must
// first upload your media asset to our servers and capture the returned media ID. If using link, your asset must
// be on a publicly accessible server or the message will fail to send.
func (client *Client) SendMedia(ctx context.Context, recipient string, req *MediaMessage,
	cacheOptions *CacheOptions,
) (*ResponseMessage, error) {
	cctx := client.Context()
	request := &SendMediaRequest{
		BaseURL:       cctx.BaseURL,
		AccessToken:   cctx.AccessToken,
		PhoneNumberID: cctx.PhoneNumberID,
		ApiVersion:    cctx.ApiVersion,
		Recipient:     recipient,
		Type:          req.Type,
		MediaID:       req.MediaID,
		MediaLink:     req.MediaLink,
		Caption:       req.Caption,
		Filename:      req.Filename,
		Provider:      req.Provider,
		CacheOptions:  cacheOptions,
	}

	resp, err := SendMedia(ctx, client.http, request, client.hooks...)
	if err != nil {
		return nil, fmt.Errorf("client send media: %w", err)
	}

	return resp, nil
}

type Template struct {
	LanguageCode   string
	LanguagePolicy string
	Name           string
	Components     []*models.TemplateComponent
}

type InteractiveTemplateRequest struct {
	Name           string
	LanguageCode   string
	LanguagePolicy string
	Headers        []*models.TemplateParameter
	Body           []*models.TemplateParameter
	Buttons        []*models.InteractiveButtonTemplate
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
	cctx := client.Context()
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewInteractiveTemplate(req.Name, tmpLanguage, req.Headers, req.Body, req.Buttons)
	payload := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          templateMessageType,
		Template:      template,
	}
	reqCtx := &whttp.RequestContext{
		Name:       "send template",
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}
	params := &whttp.Request{
		Method:  http.MethodPost,
		Payload: payload,
		Context: reqCtx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer: cctx.AccessToken,
	}
	var message ResponseMessage
	err := whttp.Do(ctx, client.http, params, &message, client.hooks...)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return &message, nil
}

type MediaTemplateRequest struct {
	Name           string
	LanguageCode   string
	LanguagePolicy string
	Header         *models.TemplateParameter
	Body           []*models.TemplateParameter
}

// SendMediaTemplate sends a media template message to the recipient. This kind of template message has a media
// message as a header. This is its main distinguishing feature from the text based template message.
func (client *Client) SendMediaTemplate(ctx context.Context, recipient string, req *MediaTemplateRequest) (
	*ResponseMessage, error,
) {
	cctx := client.Context()
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewMediaTemplate(req.Name, tmpLanguage, req.Header, req.Body)
	payload := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          templateMessageType,
		Template:      template,
	}

	reqCtx := &whttp.RequestContext{
		Name:       "send media template",
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Method:  http.MethodPost,
		Payload: payload,
		Context: reqCtx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer: cctx.AccessToken,
	}

	var message ResponseMessage
	err := whttp.Do(ctx, client.http, params, &message, client.hooks...)
	if err != nil {
		return nil, fmt.Errorf("client: send media template: %w", err)
	}

	return &message, nil
}

type TextTemplateRequest struct {
	Name           string
	LanguageCode   string
	LanguagePolicy string
	Body           []*models.TemplateParameter
}

// SendTextTemplate sends a text template message to the recipient. This kind of template message has a text
// message as a header. This is its main distinguishing feature from the media based template message.
func (client *Client) SendTextTemplate(ctx context.Context, recipient string, req *TextTemplateRequest) (
	*ResponseMessage, error,
) {
	cctx := client.Context()
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewTextTemplate(req.Name, tmpLanguage, req.Body)
	payload := models.NewMessage(recipient, models.WithTemplate(template))
	reqCtx := &whttp.RequestContext{
		Name:       "send text template",
		BaseURL:    cctx.BaseURL,
		ApiVersion: cctx.ApiVersion,
		SenderID:   cctx.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Method:  http.MethodPost,
		Payload: payload,
		Context: reqCtx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer: cctx.AccessToken,
	}

	var message ResponseMessage
	err := whttp.Do(ctx, client.http, params, &message, client.hooks...)
	if err != nil {
		return nil, fmt.Errorf("client: send text template: %w", err)
	}

	return &message, nil
}

// SendTemplate sends a template message to the recipient. There are at the moment three types of templates messages
// you can send to the user, Text Based Templates, Media Based Templates and Interactive Templates. Text Based templates
// have a text message for a Header and Media Based templates have a Media message for a Header. Interactive Templates
// can have any of the above as a Header and also have a list of buttons that the user can interact with.
// You can use models.NewTextTemplate, models.NewMediaTemplate and models.NewInteractiveTemplate to create a Template.
// These are helper functions that will make your life easier.
func (client *Client) SendTemplate(ctx context.Context, recipient string, req *Template) (*ResponseMessage, error) {
	template := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          templateMessageType,
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   req.LanguageCode,
				Policy: req.LanguagePolicy,
			},
			Name:       req.Name,
			Components: req.Components,
		},
	}

	return client.SendMessage(ctx, "send template message", template)
}

// SendInteractiveMessage sends an interactive message to the recipient.
func (client *Client) SendInteractiveMessage(ctx context.Context, recipient string, req *models.Interactive) (
	*ResponseMessage, error) {
	template := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          "interactive",
		Interactive:   req,
	}

	return client.SendMessage(ctx, "send interactive message", template)
}
