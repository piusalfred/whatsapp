/*
|
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

const MessageStatusRead = "read"

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
	ApiVersion    string //nolint: revive,stylecheck
	Recipient     string
	Message       string
	PreviewURL    bool
}

type SendLocationRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string //nolint: revive,stylecheck
	Recipient     string
	Name          string
	Address       string
	Latitude      float64
	Longitude     float64
}

type ReactRequest struct {
	BaseURL       string
	AccessToken   string
	PhoneNumberID string
	ApiVersion    string //nolint: revive,stylecheck
	Recipient     string
	MessageID     string
	Emoji         string
}

type SendTemplateRequest struct {
	BaseURL                string
	AccessToken            string
	PhoneNumberID          string
	ApiVersion             string //nolint: revive,stylecheck
	Recipient              string
	TemplateLanguageCode   string
	TemplateLanguagePolicy string
	TemplateName           string
	TemplateComponents     []*models.TemplateComponent
}

func (base *BaseClient) SendTemplate(ctx context.Context, req *SendTemplateRequest,
) (*ResponseMessage, error) {
	template := &models.Message{
		Product:       messagingProduct,
		To:            req.Recipient,
		RecipientType: individualRecipientType,
		Type:          templateMessageType,
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
		Name:          "send template",
		BaseURL:       req.BaseURL,
		ApiVersion:    req.ApiVersion,
		PhoneNumberID: req.PhoneNumberID,
		Endpoints:     []string{"messages"},
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
	err := base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return &message, nil
}

/*
CacheOptions contains the options on how to send a media message. You can specify either the
ID or the link of the media. Also, it allows you to specify caching options.

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
we will update our cached ApiVersion of the asset with the asset in the response.
Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

# ETag

The ETag header is a unique string that identifies a specific ApiVersion of an asset.
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
	ApiVersion    string //nolint: revive,stylecheck
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
link (see MediaInformation http Caching).

Be sure to keep the following in mind:
  - Uploaded media only lasts thirty days
  - Generated download URLs only last five minutes
  - Always save the media ID when you upload a file

Here’s a list of the currently supported media types. Check out Supported MediaInformation Types for more information.
  - Audio (<16 MB) – ACC, MP4, MPEG, AMR, and OGG formats
  - Documents (<100 MB) – text, PDF, Office, and OpenOffice formats
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
func (base *BaseClient) SendMedia(ctx context.Context, req *SendMediaRequest,
) (*ResponseMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil: %w", ErrBadRequestFormat)
	}

	payload, err := formatMediaPayload(req)
	if err != nil {
		return nil, err
	}

	reqCtx := &whttp.RequestContext{
		Name:          "send media",
		BaseURL:       req.BaseURL,
		ApiVersion:    req.ApiVersion,
		PhoneNumberID: req.PhoneNumberID,
		Endpoints:     []string{"messages"},
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

	err = base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send media: %w", err)
	}

	return &message, nil
}

// formatMediaPayload builds the payload for a media message. It accepts SendMediaOptions
// and returns a byte array and an error. This function is used internally by SendMedia.
// if neither ID nor Link is specified, it returns an error.
func formatMediaPayload(options *SendMediaRequest) ([]byte, error) {
	media := &models.Media{
		ID:       options.MediaID,
		Link:     options.MediaLink,
		Caption:  options.Caption,
		Filename: options.Filename,
		Provider: options.Provider,
	}
	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, fmt.Errorf("format media payload: %w", err)
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
	payloadBuilder.Write(mediaJSON)
	payloadBuilder.WriteString(`}`)

	return []byte(payloadBuilder.String()), nil
}
