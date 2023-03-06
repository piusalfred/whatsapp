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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

const (
	// MessageStatusRead is the status for a read message
	MessageStatusRead = "read"
)

type (
	StatusResponse struct {
		Success bool `json:"success,omitempty"`
	}

	MessageStatusUpdateRequest struct {
		MessagingProduct string `json:"messaging_product,omitempty"` // always whatsapp
		Status           string `json:"status,omitempty"`            // always read
		MessageID        string `json:"message_id,omitempty"`
	}
)

type SendTextRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Message       string
	PreviewURL    bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, req *SendTextRequest) (*ResponseMessage, error) {
	text := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "text",
		Text: &models.Text{
			PreviewUrl: req.PreviewURL,
			Body:       req.Message,
		},
	}

	reqCtx := &whttp.RequestContext{
		Name:       "send text",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  req.AccessToken,
		Form:    nil,
		Payload: text,
	}

	var message ResponseMessage
	err := whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send text message: %w", err)
	}

	return &message, nil
}

type SendLocationRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Name          string
	Address       string
	Latitude      float64
	Longitude     float64
}

func SendLocation(ctx context.Context, client *http.Client, req *SendLocationRequest) (*ResponseMessage, error) {
	location := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "location",
		Location: &models.Location{
			Name:      req.Name,
			Address:   req.Address,
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		Contacts:    nil,
		Interactive: nil,
	}

	reqCtx := &whttp.RequestContext{
		Name:       "send location",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.AccessToken,
		Payload: location,
	}

	var message ResponseMessage
	err := whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send location: %w", err)
	}

	return &message, nil
}

type ReactRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	MessageID     string
	Emoji         string
}

/*
React sends a reaction to a message.
To send reaction messages, make a POST call to /PHONE_NUMBER_ID/messages and attach a message object
with type=reaction. Then, add a reaction object.

Sample request:

	curl -X  POST \
	 'https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages' \
	 -H 'Authorization: Bearer ACCESS_TOKEN' \
	 -H 'Content-Type: application/json' \
	 -d '{
	  "messaging_product": "whatsapp",
	  "recipient_type": "individual",
	  "to": "PHONE_NUMBER",
	  "type": "reaction",
	  "reaction": {
	    "message_id": "wamid.HBgLM...",
	    "emoji": "\uD83D\uDE00"
	  }
	}'

If the message you are reacting to is more than 30 days old, doesn't correspond to any message
in the conversation, has been deleted, or is itself a reaction message, the reaction message will
not be delivered and you will receive a webhooks with the code 131009.

A successful response includes an object with an identifier prefixed with wamid. Use the ID listed
after wamid to track your message status.

Example response:

	{
	  "messaging_product": "whatsapp",
	  "contacts": [{
	      "input": "PHONE_NUMBER",
	      "wa_id": "WHATSAPP_ID",
	    }]
	  "messages": [{
	      "id": "wamid.ID",
	    }]
	}
*/
func React(ctx context.Context, client *http.Client, req *ReactRequest) (*ResponseMessage, error) {
	reaction := &models.Message{
		Product: "whatsapp",
		To:      req.Recipient,
		Type:    "reaction",
		Reaction: &models.Reaction{
			MessageID: req.MessageID,
			Emoji:     req.Emoji,
		},
	}

	reqCtx := &whttp.RequestContext{
		Name:       "react",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.AccessToken,
		Payload: reaction,
	}

	var message ResponseMessage
	err := whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send reaction: %w", err)
	}

	return &message, nil
}

type SendContactRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Contacts      *models.Contacts
}

func SendContact(ctx context.Context, client *http.Client, req *SendContactRequest) (*ResponseMessage, error) {
	contact := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "contact",
		Contacts:      req.Contacts,
	}
	reqCtx := &whttp.RequestContext{
		Name:       "send contacts",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  req.AccessToken,
		Form:    nil,
		Payload: contact,
	}

	var message ResponseMessage

	err := whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send contact: %w", err)
	}

	return &message, nil
}

// ReplyRequest contains options for replying to a message.
type ReplyRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Context       string // this is ID of the message to reply to
	MessageType   MessageType
	Content       any // this is a Text if MessageType is Text
}

// Reply is used to reply to a message. It accepts a ReplyRequest and returns a Response and an error.
// You can send any message as a reply to a previous message in a conversation by including the previous
// message's ID set as Context in ReplyRequest. The recipient will receive the new message along with a
// contextual bubble that displays the previous message's content.
//
// Recipients will not see a contextual bubble if:
//
// replying with a template message ("type":"template")
// replying with an image, video, PTT, or audio, and the recipient is on KaiOS
// These are known bugs which are being addressed.
// Example of Text reply:
// "messaging_product": "whatsapp",
//
//	  "context": {
//	    "message_id": "MESSAGE_ID"
//	  },
//	  "to": "<phone number> or <wa_id>",
//	  "type": "text",
//	  "text": {
//	    "preview_url": False,
//	    "body": "your-text-message-content"
//	  }
//	}'
func Reply(ctx context.Context, client *http.Client, request *ReplyRequest) (*ResponseMessage, error) {
	if request == nil {
		return nil, fmt.Errorf("reply request is nil: %w", ErrNilRequest)
	}
	payload, err := formatReplyPayload(request)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}
	reqCtx := &whttp.RequestContext{
		Name:       "reply",
		BaseURL:    request.BaseURL,
		ApiVersion: request.ApiVersion,
		SenderID:   request.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  request.AccessToken,
		Form:    nil,
		Payload: payload,
	}

	var message ResponseMessage
	err = whttp.Send(ctx, client, req, &message)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}

	return &message, nil
}

// formatReplyPayload builds the payload for a reply. It accepts ReplyRequest and returns a byte array
// and an error. This function is used internally by Reply.
func formatReplyPayload(options *ReplyRequest) ([]byte, error) {
	contentByte, err := json.Marshal(options.Content)
	if err != nil {
		return nil, fmt.Errorf("format reply payload: %w", err)
	}
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","context":{"message_id":"`)
	payloadBuilder.WriteString(options.Context)
	payloadBuilder.WriteString(`"},"to":"`)
	payloadBuilder.WriteString(options.Recipient)
	payloadBuilder.WriteString(`","type":"`)
	payloadBuilder.WriteString(string(options.MessageType))
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(string(options.MessageType))
	payloadBuilder.WriteString(`":`)
	payloadBuilder.Write(contentByte)
	payloadBuilder.WriteString(`}`)

	return []byte(payloadBuilder.String()), nil
}

type SendTemplateRequest struct {
	BaseURL                string
	AccessToken            string
	PhoneNumberID          string
	ApiVersion             string
	Recipient              string
	TemplateLanguageCode   string
	TemplateLanguagePolicy string
	TemplateName           string
	TemplateComponents     []*models.TemplateComponent
}

func SendTemplate(ctx context.Context, client *http.Client, req *SendTemplateRequest) (*ResponseMessage, error) {
	template := &models.Message{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "template",
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   req.TemplateLanguageCode,
				Policy: req.TemplateLanguagePolicy,
			},
			Name:       req.TemplateName,
			Components: req.TemplateComponents,
		},
	}
	reqCtx := &whttp.RequestContext{
		Name:       "send template",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}
	params := &whttp.Request{
		Method:  http.MethodPost,
		Payload: template,
		Context: reqCtx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Bearer: req.AccessToken,
	}
	var message ResponseMessage
	err := whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return &message, nil
}

/*
CacheOptions contains the options on how to send a media message. You can specify either the
ID or the link of the media. Also it allows you to specify caching options.

The Cloud API supports media http caching. If you are using a link (link) to a media asset on your
server (as opposed to the ID (id) of an asset you have uploaded to our servers),you can instruct us
to cache your asset for reuse with future messages by including the headers below
in your server response when we request the asset. If none of these headers are included, we will
not cache your asset.

	Cache-Control: <CACHE_CONTROL>
	Last-Modified: <LAST_MODIFIED>
	ETag: <ETAG>

# CacheControl

The Cache-Control header tells us how to handle asset caching. We support the following directives:

	max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
	messages until this time is exceeded, after which we will request the asset again, if needed.
	Example: Cache-Control: max-age=604800.

	no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
	is different from a previous response.Requires the Last-Modified header.
	Example: Cache-Control: no-cache.

	no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.

	private: Indicates that the asset is personalized for the recipient and should not be cached.

# LastModified

Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache. If the Last-Modified value
is different from a previous response and Cache-Control: no-cache is included in the response,
we will update our cached version of the asset with the asset in the response.
Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

# ETag

The ETag header is a unique string that identifies a specific version of an asset.
Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified headers
are not included in the response. In this case, we will cache the asset according to our own, internal
logic (which we do not disclose).
*/
type CacheOptions struct {
	CacheControl string `json:"cache_control,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	ETag         string `json:"etag,omitempty"`
	Expires      int64  `json:"expires,omitempty"`
}

type SendMediaRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string
	Recipient     string
	Type          MediaType
	MediaID       string
	MediaLink     string
	Caption       string
	Filename      string
	Provider      string
	CacheOptions  *CacheOptions
}

/*
SendMedia sends a media message to the recipient. To send a media message, make a POST call to the
/PHONE_NUMBER_ID/messages endpoint with type parameter set to audio, document, image, sticker, or
video, and the corresponding information for the media type such as its ID or
link (see Media http Caching).

Be sure to keep the following in mind:
  - Uploaded media only lasts thirty days
  - Generated download URLs only last five minutes
  - Always save the media ID when you upload a file

Here’s a list of the currently supported media types. Check out Supported Media Types for more information.
  - Audio (<16 MB) – ACC, MP4, MPEG, AMR, and OGG formats
  - Documents (<100 MB) – text, PDF, Office, and Open Office formats
  - Images (<5 MB) – JPEG and PNG formats
  - Video (<16 MB) – MP4 and 3GP formats
  - Stickers (<100 KB) – WebP format

Sample request using image with link:

	curl -X  POST \
	 'https://graph.facebook.com/v15.0/FROM-PHONE-NUMBER-ID/messages' \
	 -H 'Authorization: Bearer ACCESS_TOKEN' \
	 -H 'Content-Type: application/json' \
	 -d '{
	  "messaging_product": "whatsapp",
	  "recipient_type": "individual",
	  "to": "PHONE-NUMBER",
	  "type": "image",
	  "image": {
	    "link" : "https://IMAGE_URL"
	  }
	}'

Sample request using media ID:

	curl -X  POST \
	 'https://graph.facebook.com/v15.0/FROM-PHONE-NUMBER-ID/messages' \
	 -H 'Authorization: Bearer ACCESS_TOKEN' \
	 -H 'Content-Type: application/json' \
	 -d '{
	  "messaging_product": "whatsapp",
	  "recipient_type": "individual",
	  "to": "PHONE-NUMBER",
	  "type": "image",
	  "image": {
	    "id" : "MEDIA-OBJECT-ID"
	  }
	}'

A successful response includes an object with an identifier prefixed with wamid. If you are using a link to
send the media, please check the callback events delivered to your Webhook server whether the media has been
downloaded successfully.

	{
	  "messaging_product": "whatsapp",
	  "contacts": [{
	      "input": "PHONE_NUMBER",
	      "wa_id": "WHATSAPP_ID",
	    }]
	  "messages": [{
	      "id": "wamid.ID",
	    }]
	}
*/
func SendMedia(ctx context.Context, client *http.Client, req *SendMediaRequest) (*ResponseMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	payload, err := formatMediaPayload(req)
	if err != nil {
		return nil, err
	}

	reqCtx := &whttp.RequestContext{
		Name:       "send media",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Bearer:  req.AccessToken,
		Headers: map[string]string{"Content-Type": "application/json"},
		Payload: payload,
	}

	if req.CacheOptions != nil {
		if req.CacheOptions.CacheControl != "" {
			params.Headers["Cache-Control"] = req.CacheOptions.CacheControl
		} else if req.CacheOptions.Expires > 0 {
			params.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", req.CacheOptions.Expires)
		}
		if req.CacheOptions.LastModified != "" {
			params.Headers["Last-Modified"] = req.CacheOptions.LastModified
		}
		if req.CacheOptions.ETag != "" {
			params.Headers["ETag"] = req.CacheOptions.ETag
		}
	}

	var message ResponseMessage

	err = whttp.Send(ctx, client, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send media: %w", err)
	}

	return &message, nil
}

// formatMediaPayload builds the payload for a media message. It accepts SendMediaOptions
// and returns a byte array and an error. This function is used internally by SendMedia.
// if neither ID nor Link is specified, it returns an error.
//
// For Link requests, the payload should be something like this:
// {"messaging_product": "whatsapp","recipient_type": "individual","to": "PHONE-NUMBER","type": "image","image": {"link" : "https://IMAGE_URL"}}
func formatMediaPayload(options *SendMediaRequest) ([]byte, error) {
	media := &models.Media{
		ID:       options.MediaID,
		Link:     options.MediaLink,
		Caption:  options.Caption,
		Filename: options.Filename,
		Provider: options.Provider,
	}
	mediaJson, err := json.Marshal(media)
	if err != nil {
		return nil, err
	}
	recipient := options.Recipient
	mediaType := string(options.Type)
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","recipient_type":"individual","to":"`)
	payloadBuilder.WriteString(recipient)
	payloadBuilder.WriteString(`","type": "`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`":`)
	payloadBuilder.Write(mediaJson)
	payloadBuilder.WriteString(`}`)

	return []byte(payloadBuilder.String()), nil
}
