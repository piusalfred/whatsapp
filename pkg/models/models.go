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

package models

type (
	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

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

	// MediaInfo provides information about a media be it an Audio, Video, etc.
	// Animated used with stickers only.
	MediaInfo struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
		Filename string `json:"filename,omitempty"`
		Animated bool   `json:"animated,omitempty"` // used with stickers true if animated
	}

	// Media represents a media object. This object is used to send media messages to WhatsApp users.
	// It contains the following fields:
	//
	//	- ID, id (string). Required when type is audio, document, image, sticker, or video, and you are
	//    not using a link. The media object ID. Do not use this field when message type is set to text.
	//
	//	- Link, link (string). Required when type is audio, document, image, sticker, or video, and you
	//	  are not using an uploaded media ID (i.e. you are hosting the media asset on your server). The
	//	  protocol and URL of the media to be sent. Use only with http/HTTPS URLs. Do not use this field
	//	  when message type is set to text.
	//
	//		- Cloud API users only:
	//
	//		- See Media http Caching if you would like us to cache the media asset for future messages.
	//
	//		- When we request the media asset from your server you must indicate the media's MIME type by including
	//        the Content-Type http header. For example: Content-Type: video/mp4. See Supported Media Types for a
	//        list of supported media and their MIME types.
	//
	//	- Caption, caption (string). For On-Premises API users on v2.41.2 or newer, this field is required when type
	//	  is audio,document, image, or video and is limited to 1024 characters. Optional. Describes the specified image,
	//	  document, or video media. Do not use with audio or sticker media.
	//
	//   - Filename, filename (string). Optional. Describes the filename for the specific document. Use only with document
	//	   media. The extension of the filename will specify what format the document is displayed as in WhatsApp.
	//
	//	- Provider, provider (string). Optional. Only used for On-Premises API. This path is optionally used with a
	//	  link when the http/HTTPS link is not directly accessible and requires additional configurations like a bearer
	//	  token. For information on configuring providers, see the Media Providers documentation.
	Media struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
		Provider string `json:"provider,omitempty"`
	}

	//

	// Product ...
	Product struct {
		RetailerID string `json:"product_retailer_id,omitempty"`
	}

	// Message is a WhatsApp message. It contains the following fields:
	//
	// Audio (object) Required when type=audio. A media object containing audio.
	//
	// Contacts (object) Required when type=contacts. A contacts object.
	//
	// Context (object) Required if replying to any message in the conversation. Only used for Cloud API.
	// An object containing the ID of a previous message you are replying to.
	// For example: {"message_id":"MESSAGE_ID"}
	//
	// Document (object). Required when type=document. A media object containing a document.
	//
	// Hsm (object). Only used for On-Premises API. Contains a hsm object. This option was deprecated with v2.39
	// of the On-Premises API. Use the template object instead. Cloud API users should not use this field.
	//
	// Image (object). Required when type=image. A media object containing an image.
	//
	// Interactive (object). Required when type=interactive. An interactive object. The components of each interactive
	// object generally follow a consistent pattern: header, body, footer, and action.
	//
	// Location (object). Required when type=location. A location objects.
	//
	// MessagingProduct messaging_product (string)	Required. Only used for Cloud API. Messaging service used
	// for the request. Use "whatsapp". On-Premises API users should not use this field.
	//
	// PreviewURL preview_url (boolean)	Required if type=text. Only used for On-Premises API. Allows for URL
	// previews in text messages — See the Sending URLs in Text Messages.
	// This field is optional if not including a URL in your message. Values: false (default), true.
	// Cloud API users can use the same functionality with the preview_url field inside the text object.
	//
	// RecipientType recipient_type (string) Optional. Currently, you can only send messages to individuals.
	// Set this as individual. Default: individual
	//
	// Status, status (string) A message's status. You can use this field to mark a message as read.
	// See the following guides for information:
	//	- Cloud API: Mark Messages as Read
	//	- On-Premises API: Mark Messages as Read
	//
	// Sticker, sticker (object). Required when type=sticker. A media object containing a sticker.
	//	- Cloud API: Static and animated third-party outbound stickers are supported in addition to all
	//      types of inbound stickers. A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
	//      An animated sticker must be 512x512 pixels and cannot exceed 500 KB.
	//	- On-Premises API: Only static third-party outbound stickers are supported in addition to all types
	//      of inbound stickers. A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
	//      Animated stickers are not supported.
	//      For Cloud API users, we support static third-party outbound stickers and all types of inbound stickers.
	//	  The sticker needs to be 512x512 pixels and the file size needs to be less than 100 KB.
	//
	// Template A template (object). Required when type=template. A template object.
	//
	// Text (object). Required for text messages. A text objects.
	//
	// To string. Required. WhatsApp ID or phone number for the person you want to send a message to.
	// See Phone Numbers, Formatting for more information. If needed, On-Premises API users can get this number by
	// calling the contacts' endpoint.
	//
	// Type (string). Optional. The type of message you want to send. Default: text.
	Message struct {
		Product       string       `json:"messaging_product"`
		To            string       `json:"to"`
		RecipientType string       `json:"recipient_type"`
		Type          MessageType  `json:"type"`
		PreviewURL    bool         `json:"preview_url,omitempty"`
		Context       *Context     `json:"context,omitempty"`
		Template      *Template    `json:"template,omitempty"`
		Text          *Text        `json:"text,omitempty"`
		Image         *Media       `json:"image,omitempty"`
		Audio         *Media       `json:"audio,omitempty"`
		Video         *Media       `json:"video,omitempty"`
		Document      *Media       `json:"document,omitempty"`
		Sticker       *Media       `json:"sticker,omitempty"`
		Reaction      *Reaction    `json:"reaction,omitempty"`
		Location      *Location    `json:"location,omitempty"`
		Contacts      Contacts     `json:"contacts,omitempty"`
		Interactive   *Interactive `json:"interactive,omitempty"`
	}

	MessageOption func(*Message)

	// InteractiveHeaderType represent required value of InteractiveHeader.Type
	// The header type you would like to use. Supported values:
	// text: Used for ListQR Messages, Reply Buttons, and Multi-Product Messages.
	// video: Used for Reply Buttons.
	// image: Used for Reply Buttons.
	// document: Used for Reply Buttons.
	InteractiveHeaderType string
)

const (
	// InteractiveHeaderTypeText is used for ListQR Messages, Reply Buttons, and Multi-Product Messages.
	InteractiveHeaderTypeText  InteractiveHeaderType = "text"
	InteractiveHeaderTypeVideo InteractiveHeaderType = "video"
	InteractiveHeaderTypeImage InteractiveHeaderType = "image"
	InteractiveHeaderTypeDoc   InteractiveHeaderType = "document"
)

const (
	BodyMaxLength   = 1024
	FooterMaxLength = 60
)

// NewMessage creates a new message.
func NewMessage(recipient string, options ...MessageOption) *Message {
	message := &Message{
		Product:       "whatsapp",
		RecipientType: "individual",
		To:            recipient,
	}
	for _, option := range options {
		option(message)
	}

	return message
}

func WithTemplate(template *Template) MessageOption {
	return func(m *Message) {
		m.Type = "template"
		m.Template = template
	}
}

// SetTemplate sets the template of the message.
func (m *Message) SetTemplate(template *Template) {
	m.Type = "template"
	m.Template = template
}

type MessageType string

const (
	MessageTypeTemplate    MessageType = "template"
	MessageTypeText        MessageType = "text"
	MessageTypeReaction    MessageType = "reaction"
	MessageTypeLocation    MessageType = "location"
	MessageTypeContacts    MessageType = "contacts"
	MessageTypeInteractive MessageType = "interactive"
)
