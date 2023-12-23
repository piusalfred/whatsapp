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
	"fmt"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

type (
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
