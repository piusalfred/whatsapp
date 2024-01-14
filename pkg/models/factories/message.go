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

package factories

import (
	"github.com/piusalfred/whatsapp/pkg/models"
)

const (
	BodyMaxLength   = 1024
	FooterMaxLength = 60
)

type MessageType string

const (
	MessageTypeTemplate    MessageType = "template"
	MessageTypeText        MessageType = "text"
	MessageTypeReaction    MessageType = "reaction"
	MessageTypeLocation    MessageType = "location"
	MessageTypeContacts    MessageType = "contacts"
	MessageTypeInteractive MessageType = "interactive"
)

const (
	TemplateComponentTypeHeader TemplateComponentType = "header"
	TemplateComponentTypeBody   TemplateComponentType = "body"
)

type TemplateComponentType string

const (
	InteractiveMessageReplyButton = "button"
	InteractiveMessageList        = "list"
	InteractiveMessageProduct     = "product"
	InteractiveMessageProductList = "product_list"
	InteractiveMessageCTAButton   = "cta_url"
	InteractiveLocationRequest    = "location_request_message"
)

const (
	MessagingProductWhatsApp = "whatsapp"
	RecipientTypeIndividual  = "individual"
)

type MessageOption func(*models.Message) error

// WithReplyToMessageID ...
func WithReplyToMessageID(id string) MessageOption {
	return func(m *models.Message) error {
		if id != "" {
			m.Context = &models.Context{
				MessageID: id,
			}
		}

		return nil
	}
}

// TextMessage ...
func TextMessage(recipient string, text *models.Text, options ...MessageOption) (*models.Message, error) {
	t := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeText),
		Text:          text,
	}

	for _, option := range options {
		if err := option(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// ReactionMessage ...
func ReactionMessage(recipient string, reaction *models.Reaction, options ...MessageOption) (*models.Message, error) {
	r := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeReaction),
		Reaction:      reaction,
	}

	for _, option := range options {
		if err := option(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// LocationMessage ...
func LocationMessage(recipient string, location *models.Location, options ...MessageOption) (*models.Message, error) {
	l := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeLocation),
		Location:      location,
	}

	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	return l, nil
}

// TemplateMessage ...
func TemplateMessage(recipient string, template *models.Template, options ...MessageOption) (*models.Message, error) {
	t := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeTemplate),
		Template:      template,
	}

	for _, option := range options {
		if err := option(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// InteractiveMessage ...
func InteractiveMessage(recipient string, interactive *models.Interactive,
	options ...MessageOption,
) (*models.Message, error) {
	i := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeInteractive),
		Interactive:   interactive,
	}

	for _, option := range options {
		if err := option(i); err != nil {
			return nil, err
		}
	}

	return i, nil
}

// ContactsMessage ...
func ContactsMessage(recipient string, contacts []*models.Contact, options ...MessageOption) (*models.Message, error) {
	c := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          string(MessageTypeContacts),
		Contacts:      contacts,
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// AudioMessage ...
func AudioMessage(recipient string, audio *models.Audio, options ...MessageOption) (*models.Message, error) {
	a := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "audio",
		Audio:         audio,
	}

	for _, option := range options {
		if err := option(a); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// ImageMessage ...
func ImageMessage(recipient string, image *models.Image, options ...MessageOption) (*models.Message, error) {
	i := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "image",
		Image:         image,
	}

	for _, option := range options {
		if err := option(i); err != nil {
			return nil, err
		}
	}

	return i, nil
}

// VideoMessage ...
func VideoMessage(recipient string, video *models.Video, options ...MessageOption) (*models.Message, error) {
	v := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "video",
		Video:         video,
	}

	for _, option := range options {
		if err := option(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// DocumentMessage ...
func DocumentMessage(recipient string, document *models.Document, options ...MessageOption) (*models.Message, error) {
	d := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "document",
		Document:      document,
	}

	for _, option := range options {
		if err := option(d); err != nil {
			return nil, err
		}
	}

	return d, nil
}

// StickerMessage ...
func StickerMessage(recipient string, sticker *models.Sticker, options ...MessageOption) (*models.Message, error) {
	s := &models.Message{
		Product:       MessagingProductWhatsApp,
		To:            recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          "sticker",
		Sticker:       sticker,
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}
