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

	"github.com/piusalfred/whatsapp/pkg/types"
)

var (
	_ Service = (*Client)(nil)
	_ Service = (*BaseClient)(nil)
	_ Sender  = (*Client)(nil)
	_ Sender  = (*BaseClient)(nil)
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

	Sender interface {
		SendMessage(ctx context.Context, message *Message) (*Response, error)
	}

	SenderFunc func(ctx context.Context, message *Message) (*Response, error)
)

func (fn SenderFunc) SendMessage(ctx context.Context, message *Message) (*Response, error) {
	return fn(ctx, message)
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
		Context         *Context         `json:"config,omitempty"`
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
		RecipientType: RecipientTypeIndividual,
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
		RecipientType: RecipientTypeIndividual,
		Type:          TypeInteractive,
		Context:       &Context{MessageID: params.ReplyTo},
		Interactive:   content,
	}

	return message
}
