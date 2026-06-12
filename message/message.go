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

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

const (
	PinOperationPinMessage   PinOperation = "pin"
	PinOperationUnpinMessage PinOperation = "unpin"
)

type (
	SenderFunc func(ctx context.Context, message *Message) (*Response, error)
)

func (fn SenderFunc) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	return fn(ctx, message)
}

const (
	Endpoint         = "/messages"
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
		Product         string           `json:"messaging_product"`
		To              string           `json:"to,omitempty"`
		RecipientType   string           `json:"recipient_type,omitempty"`
		Type            string           `json:"type,omitempty"`
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
		Status          *string          `json:"status,omitempty"`     // used to update message Status
		MessageID       *string          `json:"message_id,omitempty"` // used to update message Status
		Template        *Template        `json:"template,omitempty"`
		TypingIndicator *TypingIndicator `json:"typing_indicator,omitempty"`
		Pin             *Pin             `json:"pin,omitempty"`
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
)

func New(recipient string, options ...Option) (*Message, error) {
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

	return msg, nil
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

type RequestLocationMessageParams struct {
	Prompt    string
	Recipient string
	ReplyTo   string
}

func NewRequestLocationMessage(params RequestLocationMessageParams) *Message {
	content := NewInteractiveMessageContent(
		TypeInteractiveLocationRequest,
		WithInteractiveBody(params.Prompt),
		WithInteractiveAction(&InteractiveAction{
			Name: InteractionActionSendLocation,
		}),
	)
	message := &Message{
		Product:       MessagingProduct,
		To:            params.Recipient,
		RecipientType: string(RecipientTypeIndividual),
		Type:          TypeInteractive,
		Context:       &Context{MessageID: params.ReplyTo},
		Interactive:   content,
	}

	return message
}
