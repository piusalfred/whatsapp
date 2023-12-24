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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

var (
	ErrConfigNil        = errors.New("config is nil")
	ErrBadRequestFormat = errors.New("bad request")
)

const (
	MessageStatusRead         = "read"
	MessageEndpoint           = "messages"
	MessagingProduct          = "whatsapp"
	RecipientTypeIndividual   = "individual"
	BaseURL                   = "https://graph.facebook.com/"
	LowestSupportedVersion    = "v16.0"
	DateFormatContactBirthday = time.DateOnly // YYYY-MM-DD
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

type (
	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,MediaInformation messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, except reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	// Client is a struct that holds the configuration for the whatsapp client.
	// It is used to create a new whatsapp client for a single user. Uses the BaseClient
	// to make requests to the whatsapp api. If you want a client that's flexible and can
	// make requests to the whatsapp api for different users, use the TransparentClient.
	Client struct {
		bc     *BaseClient
		config *Config
	}

	ClientOption func(*Client)

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

	TextMessage struct {
		Message    string
		PreviewURL bool
	}

	ReactMessage struct {
		MessageID string
		Emoji     string
	}

	TextTemplateRequest struct {
		Name           string
		LanguageCode   string
		LanguagePolicy string
		Body           []*models.TemplateParameter
	}

	Template struct {
		LanguageCode   string
		LanguagePolicy string
		Name           string
		Components     []*models.TemplateComponent
	}

	InteractiveTemplateRequest struct {
		Name           string
		LanguageCode   string
		LanguagePolicy string
		Headers        []*models.TemplateParameter
		Body           []*models.TemplateParameter
		Buttons        []*models.InteractiveButtonTemplate
	}

	MediaMessage struct {
		Type      MediaType
		MediaID   string
		MediaLink string
		Caption   string
		Filename  string
		Provider  string
	}

	MediaTemplateRequest struct {
		Name           string
		LanguageCode   string
		LanguagePolicy string
		Header         *models.TemplateParameter
		Body           []*models.TemplateParameter
	}

	// ReplyRequest contains options for replying to a message.
	ReplyRequest struct {
		Recipient   string
		Context     string // this is ID of the message to reply to
		MessageType MessageType
		Content     any // this is a Text if MessageType is Text
	}

	StatusResponse struct {
		Success bool `json:"success,omitempty"`
	}

	MessageStatusUpdateRequest struct {
		MessagingProduct string `json:"messaging_product,omitempty"` // always whatsapp
		Status           string `json:"status,omitempty"`            // always read
		MessageID        string `json:"message_id,omitempty"`
	}

	SendTextRequest struct {
		BaseURL       string
		AccessToken   string
		PhoneNumberID string
		ApiVersion    string //nolint: revive,stylecheck
		Recipient     string
		Message       string
		PreviewURL    bool
	}

	SendLocationRequest struct {
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

	ReactRequest struct {
		BaseURL       string
		AccessToken   string
		PhoneNumberID string
		ApiVersion    string //nolint: revive,stylecheck
		Recipient     string
		MessageID     string
		Emoji         string
	}

	SendTemplateRequest struct {
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

	/*
	   CacheOptions contains the options on how to send a media message. You can specify either the
	   ID or the link of the media. Also, it allows you to specify caching options.

	   The Cloud API supports media http caching. If you are using a link (link) to a media asset on your
	   server (as opposed to the ID (id) of an asset you have uploaded to our servers),you can instruct us
	   to cache your asset for reuse with future messages by including the headers below
	   in your server Resp when we request the asset. If none of these headers are included, we will
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
	   	is different from a previous Resp.Requires the Last-Modified header.
	   	Example: Cache-Control: no-cache.

	   	no-store: Indicates that the asset should not be cached. Example: Cache-Control: no-store.

	   	private: Indicates that the asset is personalized for the recipient and should not be cached.

	   # LastModified

	   Last-Modified Indicates when the asset was last modified. Used with Cache-Control: no-cache. If the
	   goLast-Modified value
	   is different from a previous Resp and Cache-Control: no-cache is included in the Resp,
	   we will update our cached ApiVersion of the asset with the asset in the Resp.
	   Example: Date: Tue, 22 Feb 2022 22:22:22 GMT.

	   # ETag

	   The ETag header is a unique string that identifies a specific ApiVersion of an asset.
	   Example: ETag: "33a64df5". This header is ignored unless both Cache-Control and Last-Modified headers
	   are not included in the Resp. In this case, we will cache the asset according to our own, internal
	   logic (which we do not disclose).
	*/
	CacheOptions struct {
		CacheControl string `json:"cache_control,omitempty"`
		LastModified string `json:"last_modified,omitempty"`
		ETag         string `json:"etag,omitempty"`
		Expires      int64  `json:"expires,omitempty"`
	}

	SendMediaRequest struct {
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
)

func WithBaseClient(base *BaseClient) ClientOption {
	return func(client *Client) {
		client.bc = base
	}
}

func NewClient(reader ConfigReader, options ...ClientOption) (*Client, error) {
	config, err := reader.Read(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return NewClientWithConfig(config, options...)
}

func NewClientWithConfig(config *Config, options ...ClientOption) (*Client, error) {
	if config == nil {
		return nil, ErrConfigNil
	}
	client := &Client{
		bc:     NewBaseClient(),
		config: config,
	}

	if client.config.BaseURL == "" {
		client.config.BaseURL = BaseURL
	}

	if client.config.Version == "" {
		client.config.Version = LowestSupportedVersion
	}

	for _, option := range options {
		if option == nil {
			// skip nil options
			continue
		}
		option(client)
	}

	return client, nil
}

// Reply is used to reply to a message. It accepts a ReplyRequest and returns a Response and an error.
// You can send any message as a reply to a previous message in a conversation by including the previous
// message's ID set as Ctx in ReplyRequest. The recipient will receive the new message along with a
// contextual bubble that displays the previous message's content.
func (client *Client) Reply(ctx context.Context, request *ReplyRequest,
) (*ResponseMessage, error) {
	if request == nil {
		return nil, fmt.Errorf("reply request is nil: %w", ErrBadRequestFormat)
	}
	payload, err := formatReplyPayload(request)
	if err != nil {
		return nil, fmt.Errorf("reply: %w", err)
	}
	reqCtx := &whttp.RequestContext{
		Name:          "reply to message",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Endpoints:     []string{MessageEndpoint},
	}

	req := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  client.config.AccessToken,
		Form:    nil,
		Payload: payload,
	}

	var message ResponseMessage
	err = client.bc.base.Do(ctx, req, &message)
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

// SendText sends a text message to a WhatsApp Business Account.
func (client *Client) SendText(ctx context.Context, recipient string,
	message *TextMessage,
) (*ResponseMessage, error) {
	text := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          textMessageType,
		Text: &models.Text{
			PreviewURL: message.PreviewURL,
			Body:       message.Message,
		},
	}

	res, err := client.SendMessage(ctx, "send text", text)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	return res, nil
}

// React sends a reaction to a message.
// To send reaction messages, make a POST call to /PHONE_NUMBER_ID/messages and attach a message object
// with type=reaction. Then, add a reaction object.
//
// Sample request:
//
//	curl -X  POST \
//	 'https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages' \
//	 -H 'Authorization: Bearer ACCESS_TOKEN' \
//	 -H 'Content-Type: application/json' \
//	 -d '{
//	  "messaging_product": "whatsapp",
//	  "recipient_type": "individual",
//	  "to": "PHONE_NUMBER",
//	  "type": "reaction",
//	  "reaction": {
//	    "message_id": "wamid.HBgLM...",
//	    "emoji": "\uD83D\uDE00"
//	  }
//	}'
//
// If the message you are reacting to is more than 30 days old, doesn't correspond to any message
// in the conversation, has been deleted, or is itself a reaction message, the reaction message will
// not be delivered, and you will receive a webhooks with the code 131009.
//
// A successful Resp includes an object with an identifier prefixed with wamid. Use the ID listed
// after wamid to track your message status.
//
// Example Resp:
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
func (client *Client) React(ctx context.Context, recipient string, msg *ReactMessage) (*ResponseMessage, error) {
	reaction := &models.Message{
		Product: MessagingProduct,
		To:      recipient,
		Type:    reactionMessageType,
		Reaction: &models.Reaction{
			MessageID: msg.MessageID,
			Emoji:     msg.Emoji,
		},
	}

	res, err := client.SendMessage(ctx, "react", reaction)
	if err != nil {
		return nil, fmt.Errorf("failed to send reaction message: %w", err)
	}

	return res, nil
}

// SendContacts sends a contact message. Contacts can be easily built using the models.NewContact() function.
func (client *Client) SendContacts(ctx context.Context, recipient string, contacts []*models.Contact) (
	*ResponseMessage, error,
) {
	contact := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          contactsMessageType,
		Contacts:      contacts,
	}

	req := &whttp.RequestContext{
		Name:          "send contacts",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Bearer:        client.config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.bc.Send(ctx, req, contact)
}

// SendLocation sends a location message to a WhatsApp Business Account.
func (client *Client) SendLocation(ctx context.Context, recipient string,
	message *models.Location,
) (*ResponseMessage, error) {
	location := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          locationMessageType,
		Location: &models.Location{
			Name:      message.Name,
			Address:   message.Address,
			Latitude:  message.Latitude,
			Longitude: message.Longitude,
		},
	}

	req := &whttp.RequestContext{
		Name:          "send location",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Bearer:        client.config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.bc.Send(ctx, req, location)
}

// SendMessage sends a message.
func (client *Client) SendMessage(ctx context.Context, name string, message *models.Message) (
	*ResponseMessage, error,
) {
	req := &whttp.RequestContext{
		Name:          name,
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Bearer:        client.config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.bc.Send(ctx, req, message)
}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	req := &whttp.RequestContext{
		Name:          "mark message read",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.bc.MarkMessageRead(ctx, req, messageID)
}

// SendMediaTemplate sends a media template message to the recipient. This kind of template message has a media
// message as a header. This is its main distinguishing feature from the text based template message.
func (client *Client) SendMediaTemplate(ctx context.Context, recipient string, req *MediaTemplateRequest) (
	*ResponseMessage, error,
) {
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewMediaTemplate(req.Name, tmpLanguage, req.Header, req.Body)
	payload := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          templateMessageType,
		Template:      template,
	}

	reqCtx := &whttp.RequestContext{
		Name:          "send media template",
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
		return nil, fmt.Errorf("client: send media template: %w", err)
	}

	return &message, nil
}

// SendTextTemplate sends a text template message to the recipient. This kind of template message has a text
// message as a header. This is its main distinguishing feature from the media based template message.
func (client *Client) SendTextTemplate(ctx context.Context, recipient string, req *TextTemplateRequest) (
	*ResponseMessage, error,
) {
	tmpLanguage := &models.TemplateLanguage{
		Policy: req.LanguagePolicy,
		Code:   req.LanguageCode,
	}
	template := models.NewTextTemplate(req.Name, tmpLanguage, req.Body)
	payload := models.NewMessage(recipient, models.WithTemplate(template))
	reqCtx := &whttp.RequestContext{
		Name:          "send text template",
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
func (client *Client) SendTemplate(ctx context.Context, recipient string, template *Template) (
	*ResponseMessage, error,
) {
	message := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          templateMessageType,
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   template.LanguageCode,
				Policy: template.LanguagePolicy,
			},
			Name:       template.Name,
			Components: template.Components,
		},
	}

	req := &whttp.RequestContext{
		Name:          "send message",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Bearer:        client.config.AccessToken,
		Endpoints:     []string{"messages"},
	}

	return client.bc.Send(ctx, req, message)
}

// SendInteractiveMessage sends an interactive message to the recipient.
func (client *Client) SendInteractiveMessage(ctx context.Context, recipient string, req *models.Interactive) (
	*ResponseMessage, error,
) {
	template := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "interactive",
		Interactive:   req,
	}

	reqc := &whttp.RequestContext{
		Name:          "send interactive message",
		BaseURL:       client.config.BaseURL,
		ApiVersion:    client.config.Version,
		PhoneNumberID: client.config.PhoneNumberID,
		Bearer:        client.config.AccessToken,
		Endpoints:     []string{"messages"},
	}

	return client.bc.Send(ctx, reqc, template)
}

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

// Whatsapp is an interface that represents a whatsapp client.
type Whatsapp interface {
	SendText(ctx context.Context, recipient string, message *TextMessage) (*ResponseMessage, error)
	React(ctx context.Context, recipient string, msg *ReactMessage) (*ResponseMessage, error)
	SendContacts(ctx context.Context, recipient string, contacts []*models.Contact) (*ResponseMessage, error)
	SendLocation(ctx context.Context, recipient string, location *models.Location) (*ResponseMessage, error)
	SendInteractiveMessage(ctx context.Context, recipient string, req *models.Interactive) (*ResponseMessage, error)
	SendTemplate(ctx context.Context, recipient string, template *Template) (*ResponseMessage, error)
	SendMedia(ctx context.Context, recipient string, media *MediaMessage, options *CacheOptions) (*ResponseMessage, error)
}

var _ Whatsapp = (*Client)(nil)

type (
	// BaseClient wraps the http client only and is used to make requests to the whatsapp api,
	// It does not have the context. This is idealy for making requests to the whatsapp api for
	// different users. The Client struct is used to make requests to the whatsapp api for a
	// single user.
	BaseClient struct {
		base *whttp.Client
		mw   []SendMiddleware
	}

	// BaseClientOption is a function that implements the BaseClientOption interface.
	BaseClientOption func(*BaseClient)
)

// WithBaseClientMiddleware adds a middleware to the base client.
func WithBaseClientMiddleware(mw ...SendMiddleware) BaseClientOption {
	return func(client *BaseClient) {
		client.mw = append(client.mw, mw...)
	}
}

// WithBaseHTTPClient sets the http client for the base client.
func WithBaseHTTPClient(httpClient *whttp.Client) BaseClientOption {
	return func(client *BaseClient) {
		client.base = httpClient
	}
}

// NewBaseClient creates a new base client.
func NewBaseClient(options ...BaseClientOption) *BaseClient {
	b := &BaseClient{base: whttp.NewClient()}

	for _, option := range options {
		option(b)
	}

	return b
}

func (c *BaseClient) SendTemplate(ctx context.Context, req *SendTemplateRequest,
) (*ResponseMessage, error) {
	template := &models.Message{
		Product:       MessagingProduct,
		To:            req.Recipient,
		RecipientType: RecipientTypeIndividual,
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
	err := c.base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return &message, nil
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

A successful Resp includes an object with an identifier prefixed with wamid. If you are using a link to
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
func (c *BaseClient) SendMedia(ctx context.Context, req *SendMediaRequest,
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

	err = c.base.Do(ctx, params, &message)
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

func (c *BaseClient) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message,
) (*ResponseMessage, error) {
	fs := WrapSender(SenderFunc(c.send), c.mw...)

	resp, err := fs.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("base client: %s: %w", req.Name, err)
	}

	return resp, nil
}

func (c *BaseClient) send(ctx context.Context, req *whttp.RequestContext,
	msg *models.Message,
) (*ResponseMessage, error) {
	request := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: msg,
	}

	var resp ResponseMessage
	err := c.base.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.Name, err)
	}

	return &resp, nil
}

func (c *BaseClient) MarkMessageRead(ctx context.Context, req *whttp.RequestContext,
	messageID string,
) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: MessagingProduct,
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	params := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: reqBody,
	}

	var success StatusResponse
	err := c.base.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("mark message read: %w", err)
	}

	return &success, nil
}

var _ Sender = (*BaseClient)(nil)

// Sender is an interface that represents a sender of a message.
type Sender interface {
	Send(ctx context.Context, req *whttp.RequestContext, message *models.Message) (*ResponseMessage, error)
}

// SenderFunc is a function that implements the Sender interface.
type SenderFunc func(ctx context.Context, req *whttp.RequestContext,
	message *models.Message) (*ResponseMessage, error)

// Send calls the function that implements the Sender interface.
func (f SenderFunc) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message) (*ResponseMessage,
	error,
) {
	return f(ctx, req, message)
}

// SendMiddleware that takes a Sender and returns a new Sender that will wrap the original
// Sender and execute the middleware function before sending the message.
type SendMiddleware func(Sender) Sender

// WrapSender wraps a Sender with a SendMiddleware.
func WrapSender(sender Sender, middleware ...SendMiddleware) Sender {
	// iterate backwards so that the middleware is executed in the right order
	for i := len(middleware) - 1; i >= 0; i-- {
		sender = middleware[i](sender)
	}

	return sender
}

// TransparentClient is a client that can send messages to a recipient without knowing the configuration of the client.
// It uses Sender instead of already configured clients. It is ideal for having a client for different environments.
type TransparentClient struct {
	Middlewares []SendMiddleware
}

// Send sends a message to the recipient.
func (client *TransparentClient) Send(ctx context.Context, sender Sender,
	req *whttp.RequestContext, message *models.Message, mw ...SendMiddleware,
) (*ResponseMessage, error) {
	s := WrapSender(WrapSender(sender, client.Middlewares...), mw...)

	response, err := s.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("transparent client: %w", err)
	}

	return response, nil
}
