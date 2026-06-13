//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package message provides a message client for the WhatsApp Cloud API following the
// standard BaseClient/Send pattern used by every other domain package.
//
// It includes all message types (Text, Image, Video, Audio, Document, Sticker,
// Location, Reaction, Contacts, Pin) and delegates interactive and template
// payloads to the [interactive] and [template] sub-packages.
//
// Usage:
//
//	msg, _ := v2.New("+16505551234", v2.WithTextMessage(&v2.Text{Body: "Hello"}))
//	resp, _ := client.SendMessage(ctx, conf, msg)
package message

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message/interactive"
	"github.com/piusalfred/whatsapp/message/template"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

type Status string

const (
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusFailed    Status = "failed"
	StatusDeleted   Status = "deleted"
	StatusWarning   Status = "warning"
)

type StatusUpdateResponse struct {
	Success bool `json:"success"`
}

type StatusUpdateRequest struct {
	MessageID string
	Status    Status
}

type SendMessageResponse struct {
	Product         string              `json:"messaging_product,omitempty"`
	Contacts        []*ResponseContact  `json:"contacts,omitempty"`
	Messages        []*ID               `json:"messages,omitempty"`
	MessageMetadata types.Metadata      `json:"-"`
	Success         bool                `json:"success"`
	Debug           *whttp.DebugDetails `json:"__debug__,omitempty"`
}

const (
	PinOperationPinMessage   PinOperation = "pin"
	PinOperationUnpinMessage PinOperation = "unpin"
)

const (
	MessagingProduct = "whatsapp"
	TypeText         = "text"
	TypeVideo        = "video"
	TypeAudio        = "audio"
	TypeSticker      = "sticker"
	TypeDocument     = "document"
	TypeImage        = "image"
	TypeLocation     = "location"
	TypeReaction     = "reaction"
	TypeContacts     = "contacts"
	TypeInteractive  = "interactive"
	TypeTemplate     = "template"
	TypePinMessage   = "pin"
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
		Product         string               `json:"messaging_product"`
		To              string               `json:"to,omitempty"`
		RecipientType   string               `json:"recipient_type,omitempty"`
		Type            string               `json:"type,omitempty"`
		PreviewURL      bool                 `json:"preview_url,omitempty"`
		Context         *Context             `json:"context,omitempty"`
		Text            *Text                `json:"text,omitempty"`
		Location        *Location            `json:"location,omitempty"`
		Reaction        *Reaction            `json:"reaction,omitempty"`
		Contacts        Contacts             `json:"contacts,omitempty"`
		Interactive     *interactive.Message `json:"interactive,omitempty"`
		Document        *Document            `json:"document,omitempty"`
		Sticker         *Sticker             `json:"sticker,omitempty"`
		Video           *Video               `json:"video,omitempty"`
		Image           *Image               `json:"image,omitempty"`
		Audio           *Audio               `json:"audio,omitempty"`
		Status          *string              `json:"status,omitempty"`
		MessageID       *string              `json:"message_id,omitempty"`
		Template        *template.Template   `json:"template,omitempty"`
		TypingIndicator *TypingIndicator     `json:"typing_indicator,omitempty"`
		Pin             *Pin                 `json:"pin,omitempty"`
	}

	Pin struct {
		Type           PinOperation `json:"type,omitempty"`
		MessageID      string       `json:"message_id,omitempty"`
		ExpirationDays int          `json:"expiration_days,omitempty"`
	}

	PinOperation string

	Option func(message *Message)

	Response struct {
		Product         string              `json:"messaging_product,omitempty"`
		Contacts        []*ResponseContact  `json:"contacts,omitempty"`
		Messages        []*ID               `json:"messages,omitempty"`
		MessageMetadata types.Metadata      `json:"-"`
		Success         bool                `json:"success"`
		Debug           *whttp.DebugDetails `json:"__debug__,omitempty"`
	}

	ID struct {
		ID            string `json:"id,omitempty"`
		MessageStatus string `json:"message_status,omitempty"`
		GroupID       string `json:"group_id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappID string `json:"wa_id"`
	}

	TypingIndicator struct {
		Type string `json:"type"`
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
		ID    string `json:"id,omitempty"`
		Link  string `json:"link,omitempty"`
		Voice bool   `json:"voice,omitempty"`
	}
)

func New(recipient string, options ...Option) *Message {
	msg := &Message{
		Product:       MessagingProduct,
		To:            recipient,
		RecipientType: string(RecipientTypeIndividual),
	}

	for _, option := range options {
		if option != nil {
			option(msg)
		}
	}

	return msg
}

type recipientType string

const (
	RecipientTypeIndividual recipientType = "individual"
	RecipientTypeGroup      recipientType = "group"
)

func WithRecipientType(recipientType recipientType) Option {
	return func(message *Message) {
		message.RecipientType = string(recipientType)
	}
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

func WithPinGroupMessageInfo(pin *Pin) Option {
	return func(message *Message) {
		message.Type = TypePinMessage
		message.Pin = pin
	}
}

func WithTemplateMessage(tmpl *template.Template) Option {
	return func(message *Message) {
		message.Type = TypeTemplate
		message.Template = tmpl
	}
}

func WithInteractiveMessage(message *interactive.Message) Option {
	return func(m *Message) {
		m.Type = TypeInteractive
		m.Interactive = message
	}
}

const endpoint = "/messages"

type RequestType string

const (
	RequestTypeSendMessage  RequestType = "send_message"
	RequestTypeUpdateStatus RequestType = "update_status"
)

type Request struct {
	RequestType RequestType `json:"-"`
	Message     *Message    `json:"-"`
	Status      string      `json:"-"`
	MessageID   string      `json:"-"`
}

type BaseRequest struct {
	MessagingProduct string               `json:"messaging_product"`
	To               string               `json:"to,omitempty"`
	RecipientType    string               `json:"recipient_type,omitempty"`
	Type             string               `json:"type,omitempty"`
	PreviewURL       bool                 `json:"preview_url,omitempty"`
	Context          *Context             `json:"context,omitempty"`
	Text             *Text                `json:"text,omitempty"`
	Image            *Image               `json:"image,omitempty"`
	Video            *Video               `json:"video,omitempty"`
	Audio            *Audio               `json:"audio,omitempty"`
	Document         *Document            `json:"document,omitempty"`
	Sticker          *Sticker             `json:"sticker,omitempty"`
	Location         *Location            `json:"location,omitempty"`
	Reaction         *Reaction            `json:"reaction,omitempty"`
	Contacts         Contacts             `json:"contacts,omitempty"`
	Interactive      *interactive.Message `json:"interactive,omitempty"`
	Template         *template.Template   `json:"template,omitempty"`
	Pin              *Pin                 `json:"pin,omitempty"`
	Status           string               `json:"status,omitempty"`
	MessageID        string               `json:"message_id,omitempty"`
}

type BaseResponse struct {
	MessagingProduct string              `json:"messaging_product,omitempty"`
	Contacts         []*ResponseContact  `json:"contacts,omitempty"`
	Messages         []*ID               `json:"messages,omitempty"`
	Success          bool                `json:"success,omitempty"`
	MessageMetadata  types.Metadata      `json:"-"`
	Debug            *whttp.DebugDetails `json:"__debug__,omitempty"`
}

type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	switch request.RequestType {
	case RequestTypeSendMessage:
		return bc.sendMessage(ctx, conf, request)
	case RequestTypeUpdateStatus:
		return bc.updateStatus(ctx, conf, request)
	default:
		return nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}
}

func (bc *BaseClient) sendMessage(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	msg := request.Message
	body := &BaseRequest{
		MessagingProduct: "whatsapp",
		To:               msg.To,
		RecipientType:    msg.RecipientType,
		Type:             msg.Type,
		PreviewURL:       msg.PreviewURL,
		Context:          msg.Context,
		Text:             msg.Text,
		Image:            msg.Image,
		Video:            msg.Video,
		Audio:            msg.Audio,
		Document:         msg.Document,
		Sticker:          msg.Sticker,
		Location:         msg.Location,
		Reaction:         msg.Reaction,
		Contacts:         msg.Contacts,
		Interactive:      msg.Interactive,
		Template:         msg.Template,
		Pin:              msg.Pin,
	}

	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(whttp.RequestTypeSendMessage).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, endpoint)

	req := whttp.BuildRequest(b, body)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) updateStatus(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	body := &BaseRequest{
		MessagingProduct: "whatsapp",
		Status:           request.Status,
		MessageID:        request.MessageID,
	}

	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(whttp.RequestTypeUpdateStatus).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, endpoint)

	req := whttp.BuildRequest(b, body)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: false,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	bc.Sender = whttp.WrapMiddlewareSender(bc.Sender, mws...)
}

func (bc *BaseClient) SendMessage(
	ctx context.Context,
	conf *config.Config,
	msg *Message,
) (*SendMessageResponse, error) {
	req := &Request{
		RequestType: RequestTypeSendMessage,
		Message:     msg,
	}
	resp, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	return &SendMessageResponse{
		Product:         resp.MessagingProduct,
		Contacts:        resp.Contacts,
		Messages:        resp.Messages,
		Success:         resp.Success,
		MessageMetadata: resp.MessageMetadata,
		Debug:           resp.Debug,
	}, nil
}

func (bc *BaseClient) UpdateMessageStatus(
	ctx context.Context,
	conf *config.Config,
	req *StatusUpdateRequest,
) (*StatusUpdateResponse, error) {
	r := &Request{
		RequestType: RequestTypeUpdateStatus,
		Status:      string(req.Status),
		MessageID:   req.MessageID,
	}
	resp, err := bc.Send(ctx, conf, r)
	if err != nil {
		return nil, fmt.Errorf("update message status: %w", err)
	}
	return &StatusUpdateResponse{Success: resp.Success}, nil
}

type Client struct {
	sender *BaseClient
	config *config.Config
}

func NewClient(conf *config.Config, opts ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](opts...)},
		config: conf,
	}
}

func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.Sender = sender
}

func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.Sender = whttp.WrapMiddlewareSender(c.sender.Sender, mws...)
}

func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	return c.sender.Send(ctx, c.config, request)
}

func (c *Client) SendMessage(ctx context.Context, msg *Message) (*SendMessageResponse, error) {
	return c.sender.SendMessage(ctx, c.config, msg)
}

func (c *Client) UpdateMessageStatus(ctx context.Context, req *StatusUpdateRequest) (*StatusUpdateResponse, error) {
	return c.sender.UpdateMessageStatus(ctx, c.config, req)
}

func (c *Client) SendTextMessage(ctx context.Context, to string, text *Text) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithTextMessage(text)))
}

func (c *Client) SendImageMessage(ctx context.Context, to string, image *Image) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithImage(image)))
}

func (c *Client) SendAudioMessage(ctx context.Context, to string, audio *Audio) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithAudio(audio)))
}

func (c *Client) SendVideoMessage(ctx context.Context, to string, video *Video) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithVideo(video)))
}

func (c *Client) SendDocumentMessage(ctx context.Context, to string, doc *Document) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithDocument(doc)))
}

func (c *Client) SendStickerMessage(ctx context.Context, to string, sticker *Sticker) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithSticker(sticker)))
}

func (c *Client) SendLocationMessage(ctx context.Context, to string, loc *Location) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithLocationMessage(loc)))
}

func (c *Client) SendReactionMessage(ctx context.Context, to string, reaction *Reaction) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithReaction(reaction)))
}

func (c *Client) SendContactsMessage(ctx context.Context, to string, contacts *Contacts) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithContacts(contacts)))
}

func (c *Client) SendInteractiveMessage(
	ctx context.Context,
	to string,
	inter *interactive.Message,
) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, func(m *Message) { m.Type = TypeInteractive; m.Interactive = inter }))
}

func (c *Client) SendTemplateMessage(
	ctx context.Context,
	to string,
	tmpl *template.Template,
) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithTemplateMessage(tmpl)))
}

func (c *Client) SendPinMessage(ctx context.Context, to string, pin *Pin) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, New(to, WithPinGroupMessageInfo(pin)))
}
