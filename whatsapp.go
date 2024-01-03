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
	"strings"
	"time"

	"github.com/piusalfred/whatsapp/pkg/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
	"github.com/piusalfred/whatsapp/pkg/models/factories"
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
	// Client is a struct that holds the configuration for the whatsapp client.
	// It is used to create a new whatsapp client for a single user. Uses the BaseClient
	// to make requests to the whatsapp api. If you want a client that's flexible and can
	// make requests to the whatsapp api for different users, use the TransparentClient.
	Client struct {
		bc     *BaseClient
		config *config.Values
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
		MessageType models.MessageType
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
		Recipient    string
		Type         MediaType
		MediaID      string
		MediaLink    string
		Caption      string
		Filename     string
		Provider     string
		CacheOptions *CacheOptions
	}
)

func WithBaseClient(base *BaseClient) ClientOption {
	return func(client *Client) {
		client.bc = base
	}
}

func NewClient(reader config.Reader, options ...ClientOption) (*Client, error) {
	values, err := reader.Read(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read values: %w", err)
	}

	return NewClientWithConfig(values, options...)
}

func NewClientWithConfig(config *config.Values, options ...ClientOption) (*Client, error) {
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

	reqCtx := whttp.MakeRequestContext(client.config, whttp.RequestTypeReply, MessageEndpoint)

	req := whttp.MakeRequest(whttp.WithRequestContext(reqCtx),
		whttp.WithBearer(client.config.AccessToken),
		whttp.WithPayload(payload),
	)

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
		Type:          models.MessageTypeText,
		Text: &models.Text{
			PreviewURL: message.PreviewURL,
			Body:       message.Message,
		},
	}

	res, err := client.SendMessage(ctx, whttp.RequestTypeTextMessage, text)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	return res, nil
}

func (client *Client) React(ctx context.Context, recipient string, msg *ReactMessage) (*ResponseMessage, error) {
	reaction := &models.Message{
		Product: MessagingProduct,
		To:      recipient,
		Type:    models.MessageTypeReaction,
		Reaction: &models.Reaction{
			MessageID: msg.MessageID,
			Emoji:     msg.Emoji,
		},
	}

	res, err := client.SendMessage(ctx, whttp.RequestTypeReact, reaction)
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
		Type:          models.MessageTypeContacts,
		Contacts:      contacts,
	}

	return client.SendMessage(ctx, whttp.RequestTypeContacts, contact)
}

// SendLocation sends a location message to a WhatsApp Business Account.
func (client *Client) SendLocation(ctx context.Context, recipient string,
	message *models.Location,
) (*ResponseMessage, error) {
	location := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          models.MessageTypeLocation,
		Location: &models.Location{
			Name:      message.Name,
			Address:   message.Address,
			Latitude:  message.Latitude,
			Longitude: message.Longitude,
		},
	}

	return client.SendMessage(ctx, whttp.RequestTypeLocation, location)
}

// SendMessage sends a message.
func (client *Client) SendMessage(ctx context.Context, name whttp.RequestType, message *models.Message) (
	*ResponseMessage, error,
) {
	req := whttp.MakeRequestContext(client.config, name, MessageEndpoint)

	return client.bc.Send(ctx, req, message)
}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	req := whttp.MakeRequestContext(client.config, whttp.RequestTypeMarkMessageRead, MessageEndpoint)

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
	template := factories.NewMediaTemplate(req.Name, tmpLanguage, req.Header, req.Body)
	payload := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          models.MessageTypeTemplate,
		Template:      template,
	}

	reqCtx := whttp.MakeRequestContext(client.config, whttp.RequestTypeMediaTemplate, MessageEndpoint)

	response, err := client.bc.Send(ctx, reqCtx, payload)
	if err != nil {
		return nil, err
	}

	return response, nil
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
	template := factories.NewTextTemplate(req.Name, tmpLanguage, req.Body)
	payload := factories.NewMessage(recipient, factories.WithMessageTemplate(template))
	reqCtx := whttp.MakeRequestContext(client.config, whttp.RequestTypeTextTemplate, MessageEndpoint)
	params := whttp.MakeRequest(whttp.WithRequestContext(reqCtx),
		whttp.WithPayload(payload),
		whttp.WithBearer(client.config.AccessToken))

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
		Type:          models.MessageTypeTemplate,
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   template.LanguageCode,
				Policy: template.LanguagePolicy,
			},
			Name:       template.Name,
			Components: template.Components,
		},
	}

	req := whttp.MakeRequestContext(client.config, whttp.RequestTypeTemplate, MessageEndpoint)

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
		Type:          models.MessageTypeInteractive,
		Interactive:   req,
	}

	reqc := whttp.MakeRequestContext(client.config, whttp.RequestTypeInteractiveMessage, MessageEndpoint)

	return client.bc.Send(ctx, reqc, template)
}

// SendMedia sends a media message to the recipient. Media can be sent using ID or Link. If using id, you must
// first upload your media asset to our servers and capture the returned media ID. If using link, your asset must
// be on a publicly accessible server or the message will fail to send.
func (client *Client) SendMedia(ctx context.Context, recipient string, req *MediaMessage,
	cacheOptions *CacheOptions,
) (*ResponseMessage, error) {
	request := &SendMediaRequest{
		Recipient:    recipient,
		Type:         req.Type,
		MediaID:      req.MediaID,
		MediaLink:    req.MediaLink,
		Caption:      req.Caption,
		Filename:     req.Filename,
		Provider:     req.Provider,
		CacheOptions: cacheOptions,
	}

	response, err := client.bc.SendMedia(ctx, client.config, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SendInteractiveTemplate send an interactive template message which contains some buttons for user interaction.
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
	template := factories.NewInteractiveTemplate(req.Name, tmpLanguage, req.Headers, req.Body, req.Buttons)
	message := &models.Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          models.MessageTypeTemplate,
		Template:      template,
	}
	reqCtx := whttp.MakeRequestContext(client.config, whttp.RequestTypeInteractiveTemplate, MessageEndpoint)

	response, err := client.bc.Send(ctx, reqCtx, message)
	if err != nil {
		return nil, err
	}

	return response, nil
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
