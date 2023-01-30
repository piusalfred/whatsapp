package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	// Context used to store the context of the conversation.
	// You can send any message as a reply to a previous message in a conversation by including
	// the previous message's ID in the context object.
	// The recipient will receive the new message along with a contextual bubble that displays
	// the previous message's content.
	// Recipients will not see a contextual bubble if:
	//    - replying with a template message ("type":"template")
	//    - replying with an image, video, PTT, or audio, and the recipient is on KaiOS
	// These are known bugs which we are addressing.
	Context struct {
		MessageID string `json:"message_id"`
	}

	// Reaction is a WhatsApp reaction
	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	// ReactionMessage is a WhatsApp reaction message
	// If the message you are reacting to is more than 30 days old, doesn't correspond to
	// any message in the conversation, has been deleted, or is itself a reaction message,
	// the reaction message will not be delivered and you will receive a webhooks with the
	// code 131009.
	//'{
	//   "messaging_product": "whatsapp",
	//   "recipient_type": "individual",
	//   "to": "PHONE_NUMBER",
	//   "type": "reaction",
	//   "reaction": {
	//     "message_id": "wamid.HBgLM...",
	//     "emoji": "\uD83D\uDE00"
	//   }
	// }'
	ReactionMessage struct {
		Product       string    `json:"messaging_product"`
		RecipientType string    `json:"recipient_type"`
		To            string    `json:"to"`
		Type          string    `json:"type"`
		Reaction      *Reaction `json:"reaction"`
	}

	// Text ...
	Text struct {
		PreviewUrl bool   `json:"preview_url,omitempty"`
		Body       string `json:"body,omitempty"`
	}

	// TextMessage is a WhatsApp text message
	// //{
	// 	"messaging_product": "whatsapp",
	// 	"recipient_type": "individual",
	// 	"to": "PHONE_NUMBER",
	// 	"type": "text",
	// 	"text": { // the text object
	// 	  "preview_url": false,
	// 	  "body": "MESSAGE_CONTENT"
	// 	  }
	//   }'
	TextMessage struct {
		Product       string `json:"messaging_product"`
		RecipientType string `json:"recipient_type"`
		To            string `json:"to"`
		Type          string `json:"type"`
		Text          *Text  `json:"text"`
	}

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, with the exception of reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	ResponseHeaders map[string][]string

	// Message is a WhatsApp message
	Message struct {
		Product  string           `json:"messaging_product"`
		To       string           `json:"to"`
		Type     string           `json:"type"`
		Template *MessageTemplate `json:"template"`
	}

	TemplateLanguage struct {
		Code string `json:"code"`
	}

	MessageTemplate struct {
		Name     string           `json:"name"`
		Language TemplateLanguage `json:"language"`
	}

	// MessageResponse is a WhatsApp message response
	// {
	//		"messaging_product":"whatsapp",
	//		"contacts":[
	//				{
	//					"input":"255767001828",
	//					"wa_id":"255767001828",
	//				},
	//		],
	//		"messages":[
	//				{"id":"wamid.HBgMMjU1NzY3MDAxODI4FQIAERgSRjVDRDE5NjhBOEEwQzQ0NUE1AA=="},
	//		],
	//	}
	MessageResponse struct {
		MessagingProduct string `json:"messaging_product"`
		Contacts         []struct {
			Input      string `json:"input"`
			WhatsappID string `json:"wa_id"`
		} `json:"contacts"`
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}

	// Response is the response from the WhatsApp server
	//
	//	{
	//	  "messaging_product": "whatsapp",
	//	  "contacts": [{
	//	      "input": "PHONE_NUMBER",
	//	      "wa_id": "WHATSAPP_ID",
	//	    }]
	//	  "messages": [{
	//	      "id": "wamid.ID",
	//	    }]
	//	}
	MessageID struct {
		ID string `json:"id"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappId string `json:"wa_id"`
	}

	Response struct {
		MessagingProduct string            `json:"messaging_product"`
		Contacts         []ResponseContact `json:"contacts"`
		Messages         []MessageID       `json:"messages"`
	}

	// InteractiveMessage ...
	InteractiveMessage struct {
		Type   string `json:"type"`
		Header struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"header"`
		Body struct {
			Text string `json:"text"`
		} `json:"body"`
		Footer struct {
			Text string `json:"text"`
		} `json:"footer"`
		Action struct {
			Button   string `json:"button"`
			Sections []struct {
				Title string `json:"title"`
				Rows  []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"rows"`
			} `json:"sections"`
		} `json:"action"`
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
		Method     string
	}

	Sender func(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error)

	SenderMiddleware func(next Sender) Sender
)

// NewRequestWithContext creates a new *http.Request with context by using the
// RequestParams.
func NewRequestWithContext(ctx context.Context, params *RequestParams, payload []byte) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	//https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages
	requestURL, err := url.JoinPath(params.BaseURL, params.ApiVersion, params.SenderID, "messages")
	if err != nil {
		return nil, fmt.Errorf("failed to join url parts: %w", err)
	}

	if payload == nil {
		req, err = http.NewRequestWithContext(ctx, params.Method, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, params.Method, requestURL, bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}
	}

	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	if params.Bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", params.Bearer))
	}

	if len(params.Query) > 0 {
		query := req.URL.Query()
		for key, value := range params.Query {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

func Send(ctx context.Context, client *http.Client, params *RequestParams, payload []byte) (*Response, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	if req, err = NewRequestWithContext(ctx, params, payload); err != nil {
		return nil, err
	}

	if resp, err = client.Do(req); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	bodybytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResponse ErrorResponse
		if err = json.Unmarshal(bodybytes, &errResponse); err != nil {
			return nil, err
		}
		errResponse.Code = resp.StatusCode
		return nil, &errResponse
	}

	var response Response
	if err = json.NewDecoder(bytes.NewBuffer(bodybytes)).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

type SendTextRequest struct {
	Recipient  string
	Message    string
	PreviewURL bool
}

// SendText sends a text message to the recipient.
func SendText(ctx context.Context, client *http.Client, params *RequestParams, req *SendTextRequest) (*Response, error) {
	text := &TextMessage{
		Product:       "whatsapp",
		To:            req.Recipient,
		RecipientType: "individual",
		Type:          "text",
		Text: &Text{
			PreviewUrl: req.PreviewURL,
			Body:       req.Message,
		},
	}

	payload, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

// Location represents a location
//
//	"location": {
//		"longitude": LONG_NUMBER,
//		"latitude": LAT_NUMBER,
//		"name": LOCATION_NAME,
//		"address": LOCATION_ADDRESS
//	  }
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
}

func SendLocation(ctx context.Context, client *http.Client, params *RequestParams, location *Location) (*Response, error) {
	payload, err := json.Marshal(location)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
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
func React(ctx context.Context, client *http.Client, params *RequestParams, reaction *Reaction) (*Response, error) {
	payload, err := json.Marshal(reaction)
	if err != nil {
		return nil, err
	}

	return Send(ctx, client, params, payload)
}

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
SendMedia sends a media message to the recipient.
To send a media message, make a POST call to the /PHONE_NUMBER_ID/messages endpoint with
type parameter set to audio, document, image, sticker, or video, and the corresponding
information for the media type such as its ID or link (see Media HTTP Caching).

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
