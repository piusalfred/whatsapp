package whatsapp

import (
	"context"
	"errors"
	"fmt"
	whttp "github.com/piusalfred/whatsapp/http"
	"net/http"
	"strings"
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

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, with the exception of reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

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

	// RequestParams are parameters for a request containing headers, query params,
	// Bearer token, Method and the body.
	// These parameters are used to create a *http.Request

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
func SendMedia(ctx context.Context, client *http.Client, params *whttp.RequestParams, options *SendMediaOptions) (*whttp.Response, error) {
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

	return whttp.Send(ctx, client, params, payload)
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
