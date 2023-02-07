package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/pesakit/whatsapp/pkg/models"
)

// https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages

const (
	BaseURL                = "https://graph.facebook.com/"
	TextMessageType        = "text"
	ReactionMessageType    = "reaction"
	MediaMessageType       = "media"
	LocationMessageType    = "location"
	ContactMessageType     = "contact"
	InteractiveMessageType = "interactive"
)

type (

	/*
		Message is a WhatsApp message. It contins the following fields:

				Audio (object) Required when type=audio. A media object containing audio.

				Contacts (object) Required when type=contacts. A contacts object.

				Context (object) Required if replying to any message in the conversation. Only used for Cloud API.
				An object containing the ID of a previous message you are replying to.
				For example: {"message_id":"MESSAGE_ID"}

				Document (object). Required when type=document. A media object containing a document.

				Hsm (object). Only used for On-Premises API. Contains an hsm object. This option was deprecated with v2.39
				of the On-Premises API. Use the template object instead. Cloud API users should not use this field.

				Image (object). Required when type=image. A media object containing an image.

				Interactive (object). Required when type=interactive. An interactive object. The components of each interactive
				object generally follow a consistent pattern: header, body, footer, and action.

				Location (object). Required when type=location. A location object.

				MessagingProduct messaging_product (string)	Required. Only used for Cloud API. Messaging service used
				for the request. Use "whatsapp". On-Premises API users should not use this field.

				PreviewURL preview_url (boolean)	Required if type=text. Only used for On-Premises API. Allows for URL
				previews in text messages — See the Sending URLs in Text Messages.
				This field is optional if not including a URL in your message. Values: false (default), true.
				Cloud API users can use the same functionality with the preview_url field inside the text object.

				RecipientType recipient_type (string) Optional. Currently, you can only send messages to individuals.
			 	Set this as individual. Default: individual

				Status status (string) A message's status. You can use this field to mark a message as read.
				See the following guides for information:
				- Cloud API: Mark Messages as Read
				- On-Premises API: Mark Messages as Read

				Sticker sticker (object). Required when type=sticker. A media object containing a sticker.
				- Cloud API: Static and animated third-party outbound stickers are supported in addition to all types of inbound stickers.
				A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
				An animated sticker must be 512x512 pixels and cannot exceed 500 KB.
				- On-Premises API: Only static third-party outbound stickers are supported in addition to all types of inbound stickers.
				A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
				Animated stickers are not supported.
				For Cloud API users, we support static third-party outbound stickers and all types of inbound stickers. The sticker needs
				to be 512x512 pixels and the file size needs to be less than 100 KB.

			    Template template (object). Required when type=template. A template object.

				Text text (object). Required for text messages. A text object.

				To string. Required. WhatsApp ID or phone number for the person you want to send a message to.
				See Phone Numbers, Formatting for more information. If needed, On-Premises API users can get this number by
				calling the contacts endpoint.

				Type type (string). Optional. The type of message you want to send. Default: text
	*/
	Message struct {
		Product       string           `json:"messaging_product"`
		To            string           `json:"to"`
		RecipientType string           `json:"recipient_type"`
		Type          string           `json:"type"`
		PreviewURL    bool             `json:"preview_url,omitempty"`
		Context       *models.Context  `json:"context,omitempty"`
		Template      *MessageTemplate `json:"template,omitempty"`
		Text          *models.Text     `json:"text,omitempty"`
		Image         *Media           `json:"image,omitempty"`
		Audio         *Media           `json:"audio,omitempty"`
		Video         *Media           `json:"video,omitempty"`
		Document      *Media           `json:"document,omitempty"`
		Sticker       *Media           `json:"sticker,omitempty"`
		Reaction      *models.Reaction `json:"reaction,omitempty"`
		Location      *models.Location `json:"location,omitempty"`
		Contacts      *models.Contacts `json:"contacts,omitempty"`
		Interactive   *Interactive     `json:"interactive,omitempty"`
	}

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, with the exception of reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	MessageTemplate struct {
		Name     string           `json:"name"`
		Language TemplateLanguage `json:"language"`
	}

	MessageID struct {
		ID string `json:"id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappId string `json:"wa_id"`
	}

	// Response is the response from the WhatsApp server
	// Example:
	//		{
	//	  		"messaging_product": "whatsapp",
	//	  		"contacts": [{
	//	      		"input": "PHONE_NUMBER",
	//	      		"wa_id": "WHATSAPP_ID",
	//	    	}]
	//	  		"messages": [{
	//	      		"id": "wamid.ID",
	//	    	}]
	//		}
	ResponseMessage struct {
		Product  string             `json:"messaging_product,omitempty"`
		Contacts []*ResponseContact `json:"contacts,omitempty"`
		Messages []*MessageID       `json:"messages,omitempty"`
	}

	Response struct {
		StatusCode int
		Headers    map[string][]string
		Message    *ResponseMessage
	}

	// RequestParams are parameters for a request containing headers, query params,
	// Bearer token, Method and the body.
	// These parameters are used to create a *http.Request
	RequestParams struct {
		SenderID   string
		ApiVersion string
		Headers    map[string]string
		Query      map[string]string
		Bearer     string
		BaseURL    string
		Endpoint   string
		Method     string
	}

	Sender func(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error)

	SenderMiddleware func(next Sender) Sender
)

// MediaType is the type of media to send it can be audio, document, image, sticker, or video.
type MediaType string

/*
SendMediaOptions contains the options on how to send a media message. You can specify either the
ID or the link of the media. Also it allows you to specify caching options.

The Cloud API supports media HTTP caching. If you are using a link (link) to a media asset on your
server (as opposed to the ID (id) of an asset you have uploaded to our servers),you can instruct us
to cache your asset for reuse with future messages by including the headers below
in your server response when we request the asset. If none of these headers are included, we will
not cache your asset.

	Cache-Control: <CACHE_CONTROL>
	Last-Modified: <LAST_MODIFIED>
	ETag: <ETAG>

# Cache-Control

The Cache-Control header tells us how to handle asset caching. We support the following directives:

	max-age=n: Indicates how many seconds (n) to cache the asset. We will reuse the cached asset in subsequent
	messages until this time is exceeded, after which we will request the asset again, if needed.
	Example: Cache-Control: max-age=604800.

	no-cache: Indicates the asset can be cached but should be updated if the Last-Modified header value
	is different from a previous response.Requires the Last-Modified header.
	Example: Cache-Control: no-cache.

	no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.

	private: Indicates that the asset is personalized for the recipient and should not be cached.

# Last-Modified

Indicates when the asset was last modified. Used with Cache-Control: no-cache. If the Last-Modified value
is different from a previous response and Cache-Control: no-cache is included in the response,
we will update our cached version of the asset with the asset in the response.
Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

# ETag

The ETag header is a unique string that identifies a specific version of an asset.
Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified headers
are not included in the response. In this case, we will cache the asset according to our own, internal
logic (which we do not disclose).
*/
type SendMediaOptions struct {
	SendByLink   bool
	SendByID     bool
	Cache        bool
	Type         MediaType
	Recipient    string
	ID           string
	Link         string
	CacheControl string
	Expires      int
	LastModified string
	ETag         string
}

const (
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeImage    MediaType = "image"
	MediaTypeSticker  MediaType = "sticker"
	MediaTypeVideo    MediaType = "video"
)

/*
SendMedia sends a media message to the recipient. To send a media message, make a POST call to the
/PHONE_NUMBER_ID/messages endpoint with type parameter set to audio, document, image, sticker, or
video, and the corresponding information for the media type such as its ID or
link (see Media HTTP Caching).

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
func SendMedia(ctx context.Context, client *http.Client, params *RequestParams, options *SendMediaOptions) (*Response, error) {
	if options == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	payload, err := BuildPayloadForMediaMessage(options)
	if err != nil {
		return nil, err
	}

	if options.Cache {
		if options.CacheControl != "" {
			params.Headers["Cache-Control"] = options.CacheControl
		} else if options.Expires > 0 {
			params.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", options.Expires)
		}
		if options.LastModified != "" {
			params.Headers["Last-Modified"] = options.LastModified
		}
		if options.ETag != "" {
			params.Headers["ETag"] = options.ETag
		}
	}

	return Send(ctx, client, params, payload)
}

//var InternalSendMediaError = errors.New("internal error while sending media")

// BuildPayloadForMediaMessage builds the payload for a media message. It accepts SendMediaOptions
// and returns a byte array and an error. This function is used internally by SendMedia.
// if neither ID nor Link is specified, it returns an error.
//
// For Link requests, the payload should be something like this:
// {"messaging_product": "whatsapp","recipient_type": "individual","to": "PHONE-NUMBER","type": "image","image": {"link" : "https://IMAGE_URL"}}
func BuildPayloadForMediaMessage(options *SendMediaOptions) ([]byte, error) {
	receipient := options.Recipient
	mediaType := string(options.Type)
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","recipient_type":"individual","to":"`)
	payloadBuilder.WriteString(receipient)
	payloadBuilder.WriteString(`","type": "`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(mediaType)

	// check if we are sending by link or by id
	if options.SendByLink {
		payloadBuilder.WriteString(`":{"link":"`)
		payloadBuilder.WriteString(options.Link)
		payloadBuilder.WriteString(`"}}`)
	} else if options.SendByID {
		payloadBuilder.WriteString(`":{"id":"`)
		payloadBuilder.WriteString(options.ID)
		payloadBuilder.WriteString(`"}}`)
	} else {
		return nil, errors.New("must specify either ID or Link")
	}

	return []byte(payloadBuilder.String()), nil
}

// ReplyOptions contains options for replying to a message.
type ReplyOptions struct {
	Recipient   string
	Context     string // this is ID of the message to reply to
	MessageType MessageType
	Content     any // this is a Text if MessageType is Text
}

// Reply is used to reply to a message. It accepts a ReplyOptions and returns a Response and an error.
// You can send any message as a reply to a previous message in a conversation by including the previous
// message's ID set as Context in ReplyOptions. The recipient will receive the new message along with a
// contextual bubble that displays the previous message's content.
//
// Recipients will not see a contextual bubble if:
//
// replying with a template message ("type":"template")
// replying with an image, video, PTT, or audio, and the recipient is on KaiOS
// These are known bugs which we are addressing.
func Reply(ctx context.Context, client *http.Client, params *RequestParams, options *ReplyOptions) (*Response, error) {
	if options == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	payload, err := BuildPayloadForReply(options)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

// BuildPayloadForReply builds the payload for a reply. It accepts ReplyOptions and returns a byte array
// and an error. This function is used internally by Reply.
func BuildPayloadForReply(options *ReplyOptions) ([]byte, error) {
	contentByte, err := json.Marshal(options.Content)
	if err != nil {
		return nil, err
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
