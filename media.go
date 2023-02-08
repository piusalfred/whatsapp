package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
	"net/http"
	"strings"
	"time"
)

// Here’s a list of the currently supported media types. Check out Supported Media Types for more information.
// Audio (<16 MB) – ACC, MP4, MPEG, AMR, and OGG formats
// Documents (<100 MB) – text, PDF, Office, and Open Office formats
// Images (<5 MB) – JPEG and PNG formats
// Video (<16 MB) – MP4 and 3GP formats
// Stickers (<100 KB) – WebP format

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

//Be sure to keep the following in mind:

// Uploaded media only lasts thirty days
// Generated download URLs only last five minutes
// Always save the media ID when you upload a file
// Here’s a list of the currently supported media types. Check out Supported Media Types for more information.

type (
	CacheOptions struct {
		CacheControl string `json:"cache_control,omitempty"`
		LastModified string `json:"last_modified,omitempty"`
		ETag         string `json:"etag,omitempty"`
		Expires      int64  `json:"expires,omitempty"`
	}

	MediaType string

	// SendMediaOptions contains the options on how to send a media message.

	SendMediaOptions struct {
		Recipient    string
		Type         MediaType
		Media        *models.Media
		CacheOptions *CacheOptions
	}
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

	if options.CacheOptions != nil {
		if options.CacheOptions.CacheControl != "" {
			params.Headers["Cache-Control"] = options.CacheOptions.CacheControl
		} else if options.CacheOptions.Expires > 0 {
			params.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", options.CacheOptions.Expires)
		}
		if options.CacheOptions.LastModified != "" {
			params.Headers["Last-Modified"] = options.CacheOptions.LastModified
		}
		if options.CacheOptions.ETag != "" {
			params.Headers["ETag"] = options.CacheOptions.ETag
		}
	}

	return whttp.Send(ctx, client, params, payload)
}

// BuildPayloadForMediaMessage builds the payload for a media message. It accepts SendMediaOptions
// and returns a byte array and an error. This function is used internally by SendMedia.
// if neither ID nor Link is specified, it returns an error.
//
// For Link requests, the payload should be something like this:
// {"messaging_product": "whatsapp","recipient_type": "individual","to": "PHONE-NUMBER","type": "image","image": {"link" : "https://IMAGE_URL"}}
func BuildPayloadForMediaMessage(options *SendMediaOptions) ([]byte, error) {
	mediaJson, err := json.Marshal(options.Media)
	if err != nil {
		return nil, err
	}
	receipient := options.Recipient
	mediaType := string(options.Type)
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","recipient_type":"individual","to":"`)
	payloadBuilder.WriteString(receipient)
	payloadBuilder.WriteString(`","type": "`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`":`)
	payloadBuilder.Write(mediaJson)
	payloadBuilder.WriteString(`}`)

	return []byte(payloadBuilder.String()), nil
}
