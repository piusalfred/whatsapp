/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package message

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

type (
	Service interface {
		SendText(ctx context.Context, request *Request[Text]) (*Response, error)
		SendLocation(ctx context.Context, request *Request[Location]) (*Response, error)
		SendVideo(ctx context.Context, request *Request[Video]) (*Response, error)
		SendReaction(ctx context.Context, request *Request[Reaction]) (*Response, error)
		SendTemplate(ctx context.Context, request *Request[Template]) (*Response, error)
		SendImage(ctx context.Context, request *Request[Image]) (*Response, error)
		SendAudio(ctx context.Context, request *Request[Audio]) (*Response, error)
		SendDocument(ctx context.Context, request *Request[Document]) (*Response, error)
		SendSticker(ctx context.Context, request *Request[Sticker]) (*Response, error)
		SendContacts(ctx context.Context, request *Request[Contacts]) (*Response, error)
		RequestLocation(ctx context.Context, request *Request[string]) (*Response, error)
		SendInteractiveMessage(ctx context.Context, request *Request[Interactive]) (*Response, error)
	}

	Request[T any] struct {
		Recipient string
		ReplyTo   string
		Message   *T
	}

	BaseClient struct {
		sender Sender
		config config.Reader
	}

	BaseRequest struct {
		Method        string
		Endpoints     []string
		Type          whttp.RequestType
		Message       *Message
		DecodeOptions whttp.DecodeOptions
		Metadata      types.Metadata
	}

	BaseRequestOption func(request *BaseRequest)
)

func NewBaseRequest(message *Message, options ...BaseRequestOption) *BaseRequest {
	b := &BaseRequest{
		Method:    http.MethodPost,
		Endpoints: []string{Endpoint},
		Type:      whttp.RequestTypeSendMessage,
		Message:   message,
		DecodeOptions: whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: true,
		},
	}

	for _, option := range options {
		if option != nil {
			option(b)
		}
	}

	return b
}

func WithBaseRequestDecodeOptions(options whttp.DecodeOptions) BaseRequestOption {
	return func(request *BaseRequest) {
		request.DecodeOptions = options
	}
}

func WithBaseRequestMetadata(metadata map[string]any) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Metadata = metadata
	}
}

func WithBaseRequestEndpoints(endpoint ...string) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Endpoints = endpoint
	}
}

func WithBaseRequestMethod(method string) BaseRequestOption {
	return func(request *BaseRequest) {
		if method != "" {
			request.Method = method
		}
	}
}

func WithBaseRequestType(reqType whttp.RequestType) BaseRequestOption {
	return func(request *BaseRequest) {
		request.Type = reqType
	}
}

func NewBaseClient(sender whttp.Sender[Message], reader config.Reader,
	middlewares ...SenderMiddleware,
) (*BaseClient, error) {
	s := &BaseSender{sender}
	sf := s.Send
	if len(middlewares) > 0 {
		for i := len(middlewares) - 1; i >= 0; i-- {
			mw := middlewares[i]
			if mw != nil {
				sf = mw(sf)
			}
		}
	}
	c := &BaseClient{
		sender: SenderFunc(sf),
		config: reader,
	}

	return c, nil
}

func (c *BaseClient) SetConfigReader(fetcher config.Reader) {
	c.config = fetcher
}

func (c *BaseClient) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	conf, err := c.config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("base client: send message: read config: %w", err)
	}

	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeSendMessage),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: true,
			InspectResponseError:  true,
		}),
	)

	response, err := c.sender.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("base client: send message: %w", err)
	}

	return response, nil
}

func (c *BaseClient) UpdateStatus(ctx context.Context, request *StatusUpdateRequest) (*StatusUpdateResponse, error) {
	ms := string(request.Status)

	message := &Message{
		Product:   MessagingProduct,
		Status:    &ms,
		MessageID: &request.MessageID,
	}

	if request.WithTypingIndicator {
		message.TypingIndicator = &TypingIndicator{Type: "text"}
	}

	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeUpdateStatus),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: false,
		}),
	)

	conf, err := c.config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("base client: update message status: read config: %w", err)
	}

	response, err := c.sender.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("base client: update message status: %w", err)
	}

	return &StatusUpdateResponse{Success: response.Success}, nil
}

type (
	Client struct {
		mu     *sync.Mutex
		reader config.Reader
		config *config.Config
		sender Sender
	}
)

func (c *Client) ReloadConfig(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	c.config, err = c.reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("reload config: %w", err)
	}

	return nil
}

func NewClient(ctx context.Context, reader config.Reader, sender whttp.Sender[Message],
	middlewares ...SenderMiddleware,
) (*Client, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	s := &BaseSender{sender}
	sf := s.Send
	if len(middlewares) > 0 {
		for i := len(middlewares) - 1; i >= 0; i-- {
			mw := middlewares[i]
			if mw != nil {
				sf = mw(sf)
			}
		}
	}

	c := &Client{
		mu:     &sync.Mutex{},
		reader: reader,
		config: conf,
		sender: SenderFunc(sf),
	}

	return c, nil
}

func (c *Client) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeSendMessage),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: true,
			InspectResponseError:  true,
		}),
	)

	response, err := c.sender.Send(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	return response, nil
}

func (c *Client) UpdateStatus(ctx context.Context, request *StatusUpdateRequest) (*StatusUpdateResponse, error) {
	ms := string(request.Status)
	message := &Message{
		Product:   MessagingProduct,
		Status:    &ms,
		MessageID: &request.MessageID,
	}

	if request.WithTypingIndicator {
		message.TypingIndicator = &TypingIndicator{Type: "text"}
	}

	req := NewBaseRequest(
		message,
		WithBaseRequestMethod(http.MethodPost),
		WithBaseRequestEndpoints(Endpoint),
		WithBaseRequestType(whttp.RequestTypeUpdateStatus),
		WithBaseRequestDecodeOptions(whttp.DecodeOptions{
			DisallowUnknownFields: true,
			DisallowEmptyResponse: false,
			InspectResponseError:  true,
		}),
	)

	response, err := c.sender.Send(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update message status: %w", err)
	}

	return &StatusUpdateResponse{Success: response.Success}, nil
}

const (
	StatusSent      status = "sent"
	StatusDelivered status = "delivered"
	StatusRead      status = "read"
	StatusFailed    status = "failed"
	StatusDeleted   status = "deleted"
	StatusWarning   status = "warning"
)

type (
	status string

	StatusUpdateResponse struct {
		Success bool `json:"success"`
	}

	StatusUpdateRequest struct {
		MessageID           string
		Status              status
		WithTypingIndicator bool
	}

	TypingIndicator struct {
		Type string `json:"type"`
	}

	StatusUpdater interface {
		UpdateStatus(ctx context.Context,
			request *StatusUpdateRequest) (*StatusUpdateResponse, error)
	}

	UpdateStatusFunc func(ctx context.Context,
		request *StatusUpdateRequest) (*StatusUpdateResponse, error)
)

func (fn UpdateStatusFunc) UpdateStatus(ctx context.Context,
	request *StatusUpdateRequest,
) (*StatusUpdateResponse, error) {
	return fn(ctx, request)
}

var (
	_ StatusUpdater = (*BaseClient)(nil)
	_ StatusUpdater = (*Client)(nil)
)

type (
	BaseSender struct {
		Sender whttp.Sender[Message]
	}

	SenderFunc func(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error)

	Sender interface {
		Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error)
	}

	SenderMiddleware func(senderFunc SenderFunc) SenderFunc
)

func (fn SenderFunc) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error) {
	return fn(ctx, conf, request)
}

func (c *BaseSender) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Response, error) {
	options := []whttp.RequestOption[Message]{
		whttp.WithRequestEndpoints[Message](conf.APIVersion, conf.PhoneNumberID, Endpoint),
		whttp.WithRequestBearer[Message](conf.AccessToken),
		whttp.WithRequestType[Message](request.Type),
		whttp.WithRequestAppSecret[Message](conf.AppSecret),
		whttp.WithRequestSecured[Message](conf.SecureRequests),
		whttp.WithRequestMessage(request.Message),
		whttp.WithRequestMetadata[Message](request.Metadata),
	}

	req := whttp.MakeRequest(request.Method, conf.BaseURL, options...)

	response := &Response{}

	decoder := whttp.ResponseDecoderJSON(response, request.DecodeOptions)

	if err := c.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("base client: send request: %w", err)
	}

	return response, nil
}

func NewRequest[T any](recipient string, message *T, replyTo string) *Request[T] {
	return &Request[T]{Recipient: recipient, Message: message, ReplyTo: replyTo}
}

func buildOptions[T any](message *T, replyTo string, createMessageFunc func(*T) Option) []Option {
	options := make([]Option, 1, 2) //nolint: mnd // ok
	options[0] = createMessageFunc(message)
	if replyTo != "" {
		options = append(options, WithMessageAsReplyTo(replyTo))
	}

	return options
}

func (c *BaseClient) SendText(ctx context.Context, request *Request[Text]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithTextMessage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendLocation(ctx context.Context, request *Request[Location]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithLocationMessage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendVideo(ctx context.Context, request *Request[Video]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithVideo)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendReaction(ctx context.Context, request *Request[Reaction]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithReaction)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendTemplate(ctx context.Context, request *Request[Template]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithTemplateMessage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendImage(ctx context.Context, request *Request[Image]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithImage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendAudio(ctx context.Context, request *Request[Audio]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithAudio)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) RequestLocation(ctx context.Context, request *Request[string]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithRequestLocationMessage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendDocument(ctx context.Context, request *Request[Document]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithDocument)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendSticker(ctx context.Context, request *Request[Sticker]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithSticker)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendContacts(ctx context.Context, request *Request[Contacts]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithContacts)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

func (c *BaseClient) SendInteractiveMessage(ctx context.Context, request *Request[Interactive]) (*Response, error) {
	options := buildOptions(request.Message, request.ReplyTo, WithInteractiveMessage)

	message, err := New(request.Recipient, options...)
	if err != nil {
		return nil, err
	}

	return c.SendMessage(ctx, message)
}

const (
	Endpoint                = "/messages"
	MessagingProduct        = "whatsapp"
	RecipientTypeIndividual = "individual"
	TypeText                = "text"
	TypeVideo               = "video"
	TypeAudio               = "audio"
	TypeSticker             = "sticker"
	TypeDocument            = "document"
	TypeImage               = "image"
	TypeLocation            = "location"
	TypeReaction            = "reaction"
	TypeContacts            = "contacts"
	TypeInteractive         = "interactive"
	TypeTemplate            = "template"
)

type (
	Text struct {
		PreviewURL bool   `json:"preview_url,omitempty"`
		Body       string `json:"body,omitempty"`
	}

	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
	}

	Context struct {
		MessageID string `json:"message_id"`
	}

	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	Message struct {
		Product         string           `json:"messaging_product"`
		To              string           `json:"to"`
		RecipientType   string           `json:"recipient_type"`
		Type            string           `json:"type"`
		PreviewURL      bool             `json:"preview_url,omitempty"`
		Context         *Context         `json:"context,omitempty"`
		Text            *Text            `json:"text,omitempty"`
		Location        *Location        `json:"location,omitempty"`
		Reaction        *Reaction        `json:"reaction,omitempty"`
		Contacts        Contacts         `json:"contacts,omitempty"`
		Interactive     *Interactive     `json:"interactive,omitempty"`
		Document        *Document        `json:"document,omitempty"`
		Sticker         *Sticker         `json:"sticker,omitempty"`
		Video           *Video           `json:"video,omitempty"`
		Image           *Image           `json:"image,omitempty"`
		Audio           *Audio           `json:"audio,omitempty"`
		Status          *string          `json:"status,omitempty"`     // used to update message status
		MessageID       *string          `json:"message_id,omitempty"` // used to update message status
		Template        *Template        `json:"template,omitempty"`
		TypingIndicator *TypingIndicator `json:"typing_indicator,omitempty"`
	}

	Option func(message *Message)

	Response struct {
		Product         string             `json:"messaging_product,omitempty"`
		Contacts        []*ResponseContact `json:"contacts,omitempty"`
		Messages        []*ID              `json:"messages,omitempty"`
		MessageMetadata types.Metadata     `json:"-"`
		Success         bool               `json:"success"`
	}

	ID struct {
		ID            string `json:"id,omitempty"`
		MessageStatus string `json:"message_status,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappID string `json:"wa_id"`
	}
)

func New(recipient string, options ...Option) (*Message, error) {
	msg := &Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "",
		PreviewURL:    false,
		Context:       nil,
		Text:          nil,
	}

	for _, option := range options {
		if option != nil {
			option(msg)
		}
	}

	return msg, nil
}

func WithImage(image *Image) Option {
	return func(message *Message) {
		message.Type = TypeImage
		message.Image = image
	}
}

func WithAudio(image *Audio) Option {
	return func(message *Message) {
		message.Type = TypeAudio
		message.Audio = image
	}
}

func WithSticker(image *Sticker) Option {
	return func(message *Message) {
		message.Type = TypeSticker
		message.Sticker = image
	}
}

func WithVideo(image *Video) Option {
	return func(message *Message) {
		message.Type = TypeVideo
		message.Video = image
	}
}

func WithDocument(doc *Document) Option {
	return func(message *Message) {
		message.Document = doc
		message.Type = TypeDocument
	}
}

func WithContacts(contacts *Contacts) Option {
	return func(message *Message) {
		message.Type = TypeContacts
		message.Contacts = *contacts
	}
}

func WithReaction(reaction *Reaction) Option {
	return func(message *Message) {
		message.Type = TypeReaction
		message.Reaction = reaction
	}
}

func WithMessageAsReplyTo(messageID string) Option {
	return func(message *Message) {
		message.Context = &Context{MessageID: messageID}
	}
}

func WithTextMessage(text *Text) Option {
	return func(message *Message) {
		message.Type = TypeText
		message.Text = text
	}
}

func WithLocationMessage(location *Location) Option {
	return func(message *Message) {
		message.Type = TypeLocation
		message.Location = location
	}
}

type (
	MediaInfo struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
		Filename string `json:"filename,omitempty"`
		Animated bool   `json:"animated,omitempty"` // used with stickers true if animated
	}

	Media struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
		Provider string `json:"provider,omitempty"`
	}

	Document struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
	}

	Video struct {
		ID      string `json:"id,omitempty"`
		Link    string `json:"link,omitempty"`
		Caption string `json:"caption,omitempty"`
	}

	Image struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
	}

	Sticker struct {
		ID string `json:"id,omitempty"`
	}

	Audio struct {
		ID string `json:"id,omitempty"`
	}
)

type (
	Address struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Type        string `json:"type"`
	}

	Addresses []*Address

	Email struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	Emails []*Email

	Name struct {
		FormattedName string `json:"formatted_name"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
		Prefix        string `json:"prefix"`
	}

	Org struct {
		Company    string `json:"company"`
		Department string `json:"department"`
		Title      string `json:"title"`
	}

	Phone struct {
		Phone string `json:"phone"`
		Type  string `json:"type"`
		WaID  string `json:"wa_id,omitempty"`
	}

	Phones []*Phone

	URL struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls []*URL

	Contact struct {
		Addresses Addresses `json:"addresses,omitempty"`
		Birthday  string    `json:"birthday"`
		Emails    Emails    `json:"emails,omitempty"`
		Name      *Name     `json:"name"`
		Org       *Org      `json:"org"`
		Phones    Phones    `json:"phones,omitempty"`
		Urls      Urls      `json:"urls,omitempty"`
	}

	Contacts []*Contact

	ContactOption func(*Contact)
)

func NewContact(options ...ContactOption) *Contact {
	contact := &Contact{}
	for _, option := range options {
		option(contact)
	}

	return contact
}

func WithContactName(name *Name) ContactOption {
	return func(c *Contact) {
		c.Name = name
	}
}

func WithContactAddresses(addresses ...*Address) ContactOption {
	return func(c *Contact) {
		c.Addresses = addresses
	}
}

func WithContactOrganization(organization *Org) ContactOption {
	return func(c *Contact) {
		c.Org = organization
	}
}

func WithContactURLs(urls ...*URL) ContactOption {
	return func(c *Contact) {
		c.Urls = urls
	}
}

func WithContactPhones(phones ...*Phone) ContactOption {
	return func(c *Contact) {
		c.Phones = phones
	}
}

func WithContactBirthdays(birthday time.Time) ContactOption {
	return func(c *Contact) {
		// should be formatted as YYYY-MM-DD
		bd := birthday.Format(time.DateOnly)
		c.Birthday = bd
	}
}

func WithContactEmails(emails ...*Email) ContactOption {
	return func(c *Contact) {
		c.Emails = emails
	}
}

const (
	TypeInteractiveLocationRequest = "location_request_message"
	TypeInteractiveCTAURL          = "cta_url"
	TypeInteractiveButton          = "button"
	TypeInteractiveFlow            = "flow"
	InteractionActionSendLocation  = "send_location"
	InteractiveActionCTAURL        = "cta_url"
	InteractiveActionButtonReply   = "reply"
	InteractiveActionFlow          = "flow"
)

type (
	InteractiveMessage string

	InteractiveButton struct {
		Type  string                  `json:"type,omitempty"`
		Title string                  `json:"title,omitempty"`
		ID    string                  `json:"id,omitempty"`
		Reply *InteractiveReplyButton `json:"reply,omitempty"`
	}

	InteractiveReplyButton struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	// InteractiveSectionRow contains information about a row in an interactive section.
	InteractiveSectionRow struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	Product struct {
		RetailerID string `json:"product_retailer_id,omitempty"`
	}

	InteractiveSection struct {
		Title        string                   `json:"title,omitempty"`
		ProductItems []*Product               `json:"product_items,omitempty"`
		Rows         []*InteractiveSectionRow `json:"rows,omitempty"`
	}

	InteractiveAction struct {
		Button            string                       `json:"button,omitempty"`
		Buttons           []*InteractiveButton         `json:"buttons,omitempty"`
		CatalogID         string                       `json:"catalog_id,omitempty"`
		ProductRetailerID string                       `json:"product_retailer_id,omitempty"`
		Sections          []*InteractiveSection        `json:"sections,omitempty"`
		Name              string                       `json:"name,omitempty"`
		Parameters        *InteractiveActionParameters `json:"parameters,omitempty"`
	}

	InteractiveActionParameters struct {
		DisplayText        string             `json:"display_text,omitempty"`
		URL                string             `json:"url,omitempty"`
		FlowMessageVersion string             `json:"flow_message_version"`
		FlowToken          string             `json:"flow_token"`
		FlowID             string             `json:"flow_id"`
		FlowCTA            string             `json:"flow_cta"`
		FlowAction         string             `json:"flow_action"`
		FlowActionPayload  *FlowActionPayload `json:"flow_action_payload"`
	}

	FlowActionPayload struct {
		Screen string                 `json:"screen"`
		Data   map[string]interface{} `json:"data"`
	}

	InteractiveHeader struct {
		Document *Document `json:"document,omitempty"`
		Image    *Image    `json:"image,omitempty"`
		Video    *Video    `json:"video,omitempty"`
		Text     string    `json:"text,omitempty"`
		Type     string    `json:"type,omitempty"`
	}

	// InteractiveFooter contains information about an interactive footer.
	InteractiveFooter struct {
		Text string `json:"text,omitempty"`
	}

	// InteractiveBody contains information about an interactive body.
	InteractiveBody struct {
		Text string `json:"text,omitempty"`
	}

	Interactive struct {
		Type   string             `json:"type,omitempty"`
		Action *InteractiveAction `json:"action,omitempty"`
		Body   *InteractiveBody   `json:"body,omitempty"`
		Footer *InteractiveFooter `json:"footer,omitempty"`
		Header *InteractiveHeader `json:"header,omitempty"`
	}

	InteractiveOption func(*Interactive)
)

type InteractiveFlowRequest struct {
	Body               string             `json:"body"`
	Header             *InteractiveHeader `json:"header"`
	Footer             string             `json:"footer"`
	FlowMessageVersion string             `json:"flow_message_version"`
	FlowToken          string             `json:"flow_token"`
	FlowID             string             `json:"flow_id"`
	FlowCTA            string             `json:"flow_cta"`
	FlowAction         string             `json:"flow_action"`
	FlowScreen         string             `json:"flow_screen"`
	FlowData           map[string]any     `json:"flow_data"`
}

func WithInteractiveFlow(req *InteractiveFlowRequest) Option {
	return func(message *Message) {
		content := NewInteractiveMessageContent(
			TypeInteractiveFlow,
			WithInteractiveFooter(req.Footer),
			WithInteractiveHeader(req.Header),
			WithInteractiveBody(req.Body),
			WithInteractiveAction(&InteractiveAction{
				Name: InteractiveActionFlow,
				Parameters: &InteractiveActionParameters{
					FlowMessageVersion: req.FlowMessageVersion,
					FlowToken:          req.FlowToken,
					FlowID:             req.FlowID,
					FlowCTA:            req.FlowCTA,
					FlowAction:         req.FlowAction,
					FlowActionPayload: &FlowActionPayload{
						Screen: req.FlowScreen,
						Data:   req.FlowData,
					},
				},
			}),
		)

		message.Type = TypeInteractive
		message.Interactive = content
	}
}

type InteractiveReplyButtonsRequest struct {
	Buttons []*InteractiveReplyButton
	Body    string
	Header  *InteractiveHeader
	Footer  string
}

func WithInteractiveReplyButtons(params *InteractiveReplyButtonsRequest) Option {
	return func(message *Message) {
		buttons := make([]*InteractiveButton, 0, len(params.Buttons))
		for _, button := range params.Buttons {
			buttons = append(buttons, &InteractiveButton{
				Type:  InteractiveActionButtonReply,
				Reply: button,
			})
		}
		content := NewInteractiveMessageContent(
			TypeInteractiveButton,
			WithInteractiveAction(&InteractiveAction{
				Buttons: buttons,
			}),
			WithInteractiveFooter(params.Footer),
			WithInteractiveBody(params.Body),
			WithInteractiveHeader(params.Header),
		)

		message.Interactive = content
		message.Type = TypeInteractive
	}
}

type InteractiveCTARequest struct {
	DisplayText string
	URL         string
	Body        string
	Header      *InteractiveHeader
	Footer      string
}

func NewInteractiveCTAURLButton(request *InteractiveCTARequest) *Interactive {
	return NewInteractiveMessageContent(
		TypeInteractiveCTAURL,
		WithInteractiveHeader(request.Header),
		WithInteractiveBody(request.Body),
		WithInteractiveFooter(request.Footer),
		WithInteractiveAction(&InteractiveAction{
			Name: InteractiveActionCTAURL,
			Parameters: &InteractiveActionParameters{
				DisplayText: request.DisplayText,
				URL:         request.URL,
			},
		}),
	)
}

func WithInteractiveCTAURLButton(request *InteractiveCTARequest) Option {
	return func(message *Message) {
		content := NewInteractiveMessageContent(
			TypeInteractiveCTAURL,
			WithInteractiveHeader(request.Header),
			WithInteractiveBody(request.Body),
			WithInteractiveFooter(request.Footer),
			WithInteractiveAction(&InteractiveAction{
				Name: InteractiveActionCTAURL,
				Parameters: &InteractiveActionParameters{
					DisplayText: request.DisplayText,
					URL:         request.URL,
				},
			}),
		)

		message.Type = TypeInteractive
		message.Interactive = content
	}
}

func WithInteractiveMessage(request *Interactive) Option {
	return func(message *Message) {
		message.Type = TypeInteractive
		message.Interactive = request
	}
}

func WithRequestLocationMessage(text *string) Option {
	return func(message *Message) {
		content := NewInteractiveMessageContent(
			TypeInteractiveLocationRequest,
			WithInteractiveBody(*text),
			WithInteractiveAction(&InteractiveAction{
				Name: InteractionActionSendLocation,
			}),
		)
		message.Type = TypeInteractive
		message.Interactive = content
	}
}

func WithInteractiveFooter(footer string) InteractiveOption {
	return func(i *Interactive) {
		i.Footer = &InteractiveFooter{
			Text: footer,
		}
	}
}

func WithInteractiveBody(body string) InteractiveOption {
	return func(i *Interactive) {
		i.Body = &InteractiveBody{
			Text: body,
		}
	}
}

func WithInteractiveHeader(header *InteractiveHeader) InteractiveOption {
	return func(i *Interactive) {
		i.Header = header
	}
}

func WithInteractiveAction(action *InteractiveAction) InteractiveOption {
	return func(i *Interactive) {
		i.Action = action
	}
}

func NewInteractiveMessageContent(interactiveType string, options ...InteractiveOption) *Interactive {
	interactive := &Interactive{
		Type: interactiveType,
	}
	for _, option := range options {
		option(interactive)
	}

	return interactive
}

func InteractiveHeaderVideo(video *Video) *InteractiveHeader {
	return &InteractiveHeader{
		Video: video,
		Type:  "video",
	}
}

func InteractiveHeaderImage(image *Image) *InteractiveHeader {
	return &InteractiveHeader{
		Image: image,
		Type:  "image",
	}
}

func InteractiveHeaderText(text string) *InteractiveHeader {
	return &InteractiveHeader{
		Text: text,
		Type: "text",
	}
}
