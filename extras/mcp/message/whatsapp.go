//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package message

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/piusalfred/whatsapp/message"
)

type (
	Sender interface {
		SendMessage(ctx context.Context, message *message.Message) (*message.Response, error)
	}

	SendTextRequest struct {
		Text       string `json:"text"`
		PreviewURL bool   `json:"preview_url"`
		Recipient  string `json:"recipient"`
		ReplyTo    string `json:"reply_to"`
	}

	RequestLocationRequest struct {
		Recipient string `json:"recipient"`
		ReplyTo   string `json:"reply_to"`
		Message   string `json:"message"`
	}

	SendLocationRequest struct {
		Longitude    float64 `json:"longitude"`
		Latitude     float64 `json:"latitude"`
		LocationName string  `json:"name"`
		Address      string  `json:"address"`
		Recipient    string  `json:"recipient"`
		ReplyTo      string  `json:"reply_to"`
	}

	SendImageRequest struct {
		Recipient string `json:"recipient"`
		Link      string `json:"link"`
		Caption   string `json:"caption"`
		Filename  string `json:"filename"`
		ReplyTo   string `json:"reply_to"`
		ImageID   string `json:"image_id"`
	}

	SendDocumentRequest struct {
		Recipient  string `json:"recipient"`
		Link       string `json:"link"`
		Caption    string `json:"caption"`
		Filename   string `json:"filename"`
		ReplyTo    string `json:"reply_to"`
		DocumentID string `json:"document_id"`
	}

	SendAudioRequest struct {
		Recipient string `json:"recipient"`
		ReplyTo   string `json:"reply_to"`
		AudioID   string `json:"audio_id"`
	}

	SendVideoRequest struct {
		Recipient string `json:"recipient"`
		Link      string `json:"link"`
		Caption   string `json:"caption"`
		ReplyTo   string `json:"reply_to"`
		VideoID   string `json:"video_id"`
	}

	SendReactionRequest struct {
		Recipient string `json:"recipient"`
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	SendTemplateRequest struct {
		Recipient  string              `json:"recipient"`
		Name       string              `json:"name"`
		Language   string              `json:"language"`
		Components []TemplateComponent `json:"components"`
		ReplyTo    string              `json:"reply_to"`
	}

	TemplateComponent struct {
		Type       string              `json:"type"`
		Parameters []TemplateParameter `json:"parameters"`
	}

	TemplateParameter struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	SendStickerRequest struct {
		Recipient string `json:"recipient"`
		StickerID string `json:"sticker_id"`
		ReplyTo   string `json:"reply_to"`
	}

	SendContactsRequest struct {
		Recipient string    `json:"recipient"`
		Contacts  []Contact `json:"contacts"`
		ReplyTo   string    `json:"reply_to"`
	}

	Contact struct {
		Name      ContactName      `json:"name"`
		Phones    []ContactPhone   `json:"phones"`
		Emails    []ContactEmail   `json:"emails"`
		Addresses []ContactAddress `json:"addresses"`
		Urls      []ContactURL     `json:"urls"`
		Org       ContactOrg       `json:"org"`
		Birthday  string           `json:"birthday"`
	}

	ContactName struct {
		FormattedName string `json:"formatted_name"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
		Prefix        string `json:"prefix"`
	}

	ContactPhone struct {
		Phone string `json:"phone"`
		Type  string `json:"type"`
		WaID  string `json:"wa_id"`
	}

	ContactEmail struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	ContactAddress struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Type        string `json:"type"`
	}

	ContactURL struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	ContactOrg struct {
		Company    string `json:"company"`
		Department string `json:"department"`
		Title      string `json:"title"`
	}

	SendInteractiveRequest struct {
		Recipient string `json:"recipient"`
		Type      string `json:"type"`
		Body      string `json:"body"`
		Footer    string `json:"footer"`
		Header    string `json:"header"`
		Action    string `json:"action"`
		ReplyTo   string `json:"reply_to"`
	}

	Response struct {
		Product       string `json:"product"`
		Input         string `json:"input"`
		WhatsappID    string `json:"whatsapp_id"`
		MessageID     string `json:"message_id"`
		MessageStatus string `json:"message_status"`
	}
)

func (s *Server) SendText(ctx context.Context, request *SendTextRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithTextMessage(&message.Text{
			PreviewURL: request.PreviewURL,
			Body:       request.Text,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send text: %w", err)
	}

	return response, nil
}

func (s *Server) RequestLocation(ctx context.Context, request *RequestLocationRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithMessageAsReplyTo(request.ReplyTo),
		message.WithRequestLocationMessage(&request.Message),
	)
	if err != nil {
		return nil, fmt.Errorf("send location request: %w", err)
	}

	return response, nil
}

func (s *Server) SendLocation(ctx context.Context, request *SendLocationRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithMessageAsReplyTo(request.ReplyTo),
		message.WithLocationMessage(&message.Location{
			Longitude: request.Longitude,
			Latitude:  request.Latitude,
			Name:      request.LocationName,
			Address:   request.Address,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("send location: %w", err)
	}

	return response, nil
}

func (s *Server) SendImage(ctx context.Context, request *SendImageRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithImage(&message.Image{
			ID:       request.ImageID,
			Link:     request.Link,
			Caption:  request.Caption,
			Filename: request.Filename,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send image: %w", err)
	}

	return response, nil
}

func (s *Server) SendDocument(ctx context.Context, request *SendDocumentRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithDocument(&message.Document{
			ID:       request.DocumentID,
			Link:     request.Link,
			Caption:  request.Caption,
			Filename: request.Filename,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send document: %w", err)
	}

	return response, nil
}

func (s *Server) SendAudio(ctx context.Context, request *SendAudioRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithAudio(&message.Audio{ID: request.AudioID}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send audio: %w", err)
	}

	return response, nil
}

func (s *Server) SendVideo(ctx context.Context, request *SendVideoRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithVideo(&message.Video{
			ID:      request.VideoID,
			Link:    request.Link,
			Caption: request.Caption,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send video: %w", err)
	}

	return response, nil
}

func (s *Server) SendReaction(ctx context.Context, request *SendReactionRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithReaction(&message.Reaction{
			MessageID: request.MessageID,
			Emoji:     request.Emoji,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("send reaction: %w", err)
	}

	return response, nil
}

func (s *Server) SendTemplate(ctx context.Context, request *SendTemplateRequest) (*Response, error) {
	components := make([]*message.TemplateComponent, len(request.Components))
	for i, comp := range request.Components {
		parameters := make([]*message.TemplateParameter, len(comp.Parameters))
		for j, param := range comp.Parameters {
			parameters[j] = &message.TemplateParameter{
				Type: param.Type,
				Text: param.Text,
			}
		}
		components[i] = &message.TemplateComponent{
			Type:       comp.Type,
			Parameters: parameters,
		}
	}

	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithTemplateMessage(&message.Template{
			Name:       request.Name,
			Language:   &message.TemplateLanguage{Code: request.Language},
			Components: components,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send template: %w", err)
	}

	return response, nil
}

func (s *Server) SendSticker(ctx context.Context, request *SendStickerRequest) (*Response, error) {
	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithSticker(&message.Sticker{
			ID: request.StickerID,
		}),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send sticker: %w", err)
	}

	return response, nil
}

func (s *Server) SendContacts(ctx context.Context, request *SendContactsRequest) (*Response, error) {
	contacts := make([]*message.Contact, len(request.Contacts))
	for i, contact := range request.Contacts {
		phones := make([]*message.Phone, len(contact.Phones))
		for j, phone := range contact.Phones {
			phones[j] = &message.Phone{
				Phone: phone.Phone,
				Type:  phone.Type,
				WaID:  phone.WaID,
			}
		}

		emails := make([]*message.Email, len(contact.Emails))
		for j, email := range contact.Emails {
			emails[j] = &message.Email{
				Email: email.Email,
				Type:  email.Type,
			}
		}

		addresses := make([]*message.Address, len(contact.Addresses))
		for j, addr := range contact.Addresses {
			addresses[j] = &message.Address{
				Street:      addr.Street,
				City:        addr.City,
				State:       addr.State,
				Zip:         addr.Zip,
				Country:     addr.Country,
				CountryCode: addr.CountryCode,
				Type:        addr.Type,
			}
		}

		urls := make([]*message.URL, len(contact.Urls))
		for j, url := range contact.Urls {
			urls[j] = &message.URL{
				URL:  url.URL,
				Type: url.Type,
			}
		}

		contacts[i] = &message.Contact{
			Name: &message.Name{
				FormattedName: contact.Name.FormattedName,
				FirstName:     contact.Name.FirstName,
				LastName:      contact.Name.LastName,
				MiddleName:    contact.Name.MiddleName,
				Suffix:        contact.Name.Suffix,
				Prefix:        contact.Name.Prefix,
			},
			Phones:    phones,
			Emails:    emails,
			Addresses: addresses,
			Urls:      urls,
			Org: &message.Org{
				Company:    contact.Org.Company,
				Department: contact.Org.Department,
				Title:      contact.Org.Title,
			},
			Birthday: contact.Birthday,
		}
	}

	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithContacts((*message.Contacts)(&contacts)),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send contacts: %w", err)
	}

	return response, nil
}

func (s *Server) SendInteractiveMessage(ctx context.Context, request *SendInteractiveRequest) (*Response, error) {
	var interactive *message.Interactive

	switch request.Type {
	case "button":
		// For button type, we'll create a simple interactive message
		interactive = &message.Interactive{
			Type: request.Type,
			Body: &message.InteractiveBody{
				Text: request.Body,
			},
			Footer: &message.InteractiveFooter{
				Text: request.Footer,
			},
			Header: &message.InteractiveHeader{
				Text: request.Header,
				Type: "text",
			},
		}
	case "cta_url":
		interactive = &message.Interactive{
			Type: request.Type,
			Body: &message.InteractiveBody{
				Text: request.Body,
			},
			Footer: &message.InteractiveFooter{
				Text: request.Footer,
			},
			Header: &message.InteractiveHeader{
				Text: request.Header,
				Type: "text",
			},
			Action: &message.InteractiveAction{
				Name: request.Action,
			},
		}
	default:
		return nil, fmt.Errorf("unsupported interactive message type: %s", request.Type)
	}

	response, err := s.SendRequest(ctx, request.Recipient,
		message.WithInteractiveMessage(interactive),
		message.WithMessageAsReplyTo(request.ReplyTo),
	)
	if err != nil {
		return nil, fmt.Errorf("send interactive message: %w", err)
	}

	return response, nil
}

func (s *Server) SendRequest(ctx context.Context, recipient string, options ...message.Option) (*Response, error) {
	msg, err := message.New(recipient, options...)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	output, err := s.whatsapp.SendMessage(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	response := &Response{
		Product:       output.Product,
		Input:         output.Contacts[0].Input,
		WhatsappID:    output.Contacts[0].WhatsappID,
		MessageStatus: output.Messages[0].MessageStatus,
		MessageID:     output.Messages[0].ID,
	}

	return response, nil
}

type Schemas struct {
	response               *jsonschema.Schema
	sendTextRequest        *jsonschema.Schema
	requestLocationRequest *jsonschema.Schema
	sendLocationRequest    *jsonschema.Schema
	sendImageRequest       *jsonschema.Schema
	sendDocumentRequest    *jsonschema.Schema
	sendAudioRequest       *jsonschema.Schema
	sendVideoRequest       *jsonschema.Schema
	sendReactionRequest    *jsonschema.Schema
	sendTemplateRequest    *jsonschema.Schema
	sendStickerRequest     *jsonschema.Schema
	sendContactsRequest    *jsonschema.Schema
	sendInteractiveRequest *jsonschema.Schema
}

func (s *Server) initSchemas() { //nolint:funlen // it's OK
	s.schemas = &Schemas{
		response: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"product":        {Type: "string", Description: "API product type (e.g., 'whatsapp')."},
				"input":          {Type: "string", Description: "Normalized recipient identifier."},
				"whatsapp_id":    {Type: "string", Description: "WhatsApp ID of the recipient."},
				"message_status": {Type: "string", Description: "Delivery status of the message (e.g., 'accepted')."},
				"message_id":     {Type: "string", Description: "Unique ID of the message."},
			},
			Required:    []string{"product", "input", "whatsapp_id", "message_status", "message_id"},
			Description: "Standard response after sending a WhatsApp message.",
		},
		sendTextRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"text":        {Type: "string", Description: "Text message body to send."},
				"recipient":   {Type: "string", Description: "Recipient phone number in international format."},
				"preview_url": {Type: "boolean", Description: "Allow URL preview if a link is in the text."},
				"reply_to":    {Type: "string", Description: "Optional message ID this text is replying to."},
			},
			Required:    []string{"text", "recipient"},
			Description: "Input for sending a WhatsApp text message.",
		},
		requestLocationRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"message":   {Type: "string", Description: "The message prompt to send to the user."},
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"reply_to":  {Type: "string", Description: "Optional message ID this request is replying to."},
			},
			Required:    []string{"message", "recipient"},
			Description: "Input for asking a user to share their location.",
		},
		sendLocationRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"longitude": {Type: "number", Description: "Longitude coordinate of the location."},
				"latitude":  {Type: "number", Description: "Latitude coordinate of the location."},
				"name":      {Type: "string", Description: "Name of the location."},
				"address":   {Type: "string", Description: "Address of the location."},
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"reply_to":  {Type: "string", Description: "Optional message ID this location is replying to."},
			},
			Required:    []string{"longitude", "latitude", "recipient"},
			Description: "Input for sending a WhatsApp location message.",
		},
		sendImageRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"link":      {Type: "string", Description: "URL link to the image."},
				"caption":   {Type: "string", Description: "Caption text for the image."},
				"filename":  {Type: "string", Description: "Filename of the image."},
				"reply_to":  {Type: "string", Description: "Optional message ID this image is replying to."},
				"image_id":  {Type: "string", Description: "Media ID of the image (if uploaded previously)."},
			},
			Required:    []string{"recipient"},
			Description: "Input for sending a WhatsApp image message. Supported image formats: JPEG (.jpeg, image/jpeg), PNG (.png, image/png). Images must be 8-bit, RGB or RGBA. Max size: 5 MB.",
		},
		sendDocumentRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient":   {Type: "string", Description: "Recipient phone number in international format."},
				"link":        {Type: "string", Description: "URL link to the document."},
				"caption":     {Type: "string", Description: "Caption text for the document."},
				"filename":    {Type: "string", Description: "Filename of the document."},
				"reply_to":    {Type: "string", Description: "Optional message ID this document is replying to."},
				"document_id": {Type: "string", Description: "Media ID of the document (if uploaded previously)."},
			},
			Required:    []string{"recipient"},
			Description: "Input for sending a WhatsApp document message.",
		},
		sendAudioRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"reply_to":  {Type: "string", Description: "Optional message ID this audio is replying to."},
				"audio_id":  {Type: "string", Description: "Media ID of the audio (if uploaded previously)."},
			},
			Required:    []string{"recipient", "audio_id"},
			Description: "Input for sending a WhatsApp audio message. Supported audio formats: AAC (.aac, audio/aac), AMR (.amr, audio/amr), MP3 (.mp3, audio/mpeg), MP4 Audio (.m4a, audio/mp4), OGG Audio (.ogg, audio/ogg with OPUS codecs only; mono input only). Max size: 16 MB. Always ensure the audio file's actual MIME type matches one of the supported types. Files with unsupported MIME types (like audio/mp3) will be rejected by the API.",
		},
		sendVideoRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"link":      {Type: "string", Description: "URL link to the video."},
				"caption":   {Type: "string", Description: "Caption text for the video."},
				"reply_to":  {Type: "string", Description: "Optional message ID this video is replying to."},
				"video_id":  {Type: "string", Description: "Media ID of the video (if uploaded previously)."},
			},
			Required:    []string{"recipient"},
			Description: "Input for sending a WhatsApp video message. Supported video formats: 3GPP (.3gp, video/3gpp), MP4 Video (.mp4, video/mp4). Only H.264 video codec and AAC audio codec supported. Single audio stream or no audio stream only. Max size: 16 MB.",
		},
		sendReactionRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient":  {Type: "string", Description: "Recipient phone number in international format."},
				"message_id": {Type: "string", Description: "ID of the message to react to."},
				"emoji":      {Type: "string", Description: "Emoji to use for the reaction."},
			},
			Required:    []string{"recipient", "message_id", "emoji"},
			Description: "Input for sending a WhatsApp reaction message.",
		},
		sendTemplateRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient":  {Type: "string", Description: "Recipient phone number in international format."},
				"name":       {Type: "string", Description: "Name of the template."},
				"language":   {Type: "string", Description: "Language code for the template."},
				"components": {Type: "array", Description: "Template components with parameters."},
				"reply_to":   {Type: "string", Description: "Optional message ID this template is replying to."},
			},
			Required:    []string{"recipient", "name", "language"},
			Description: "Input for sending a WhatsApp template message.",
		},
		sendStickerRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient":  {Type: "string", Description: "Recipient phone number in international format."},
				"sticker_id": {Type: "string", Description: "Media ID of the sticker (if uploaded previously)."},
				"reply_to":   {Type: "string", Description: "Optional message ID this sticker is replying to."},
			},
			Required:    []string{"recipient", "sticker_id"},
			Description: "Input for sending a WhatsApp sticker message.",
		},
		sendContactsRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"contacts":  {Type: "array", Description: "Array of contact information to send."},
				"reply_to":  {Type: "string", Description: "Optional message ID this contacts message is replying to."},
			},
			Required:    []string{"recipient", "contacts"},
			Description: "Input for sending WhatsApp contact information.",
		},
		sendInteractiveRequest: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"recipient": {Type: "string", Description: "Recipient phone number in international format."},
				"type":      {Type: "string", Description: "Type of interactive message (button, cta_url, etc.)."},
				"body":      {Type: "string", Description: "Body text of the interactive message."},
				"footer":    {Type: "string", Description: "Footer text of the interactive message."},
				"header":    {Type: "string", Description: "Header text of the interactive message."},
				"action":    {Type: "string", Description: "Action for the interactive message."},
				"reply_to": {
					Type:        "string",
					Description: "Optional message ID this interactive message is replying to.",
				},
			},
			Required:    []string{"recipient", "type", "body"},
			Description: "Input for sending a WhatsApp interactive message.",
		},
	}
}
