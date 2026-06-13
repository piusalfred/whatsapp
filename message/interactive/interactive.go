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

// Package interactive provides constructors and helpers for composing
// WhatsApp interactive message payloads.
//
// Interactive messages let users respond by tapping buttons, selecting
// from lists, or navigating to URLs — without typing a reply. This
// package covers all interactive types: CTA URL buttons, reply buttons,
// list messages, flow messages, location requests, address messages,
// call permission requests, and media carousels.
//
// Each constructor returns a [*Message] suitable for use with
// [github.com/piusalfred/whatsapp/message.Client.SendInteractiveMessage].
//
//	inter := interactive.List(&interactive.ListRequest{
//	    Body:   "Which shipping option do you prefer?",
//	    Button: "Shipping Options",
//	    Sections: []*interactive.Section{...},
//	})
//	resp, err := client.SendInteractiveMessage(ctx, message.NewRequest("+16505551234", inter))
package interactive

import (
	"github.com/piusalfred/whatsapp/message/media"
)

const (
	MessageTypeLocationRequest       MessageType = "location_request_message"
	MessageTypeCTAURL                MessageType = "cta_url"
	MessageTypeButton                MessageType = "button"
	MessageTypeFlow                  MessageType = "flow"
	MessageTypeAddressMessage        MessageType = "address_message"
	MessageTypeCarousel              MessageType = "carousel"
	MessageTypeCallPermissionRequest MessageType = "call_permission_request"
	MessageTypeList                  MessageType = "list"
	ActionSendLocation               ActionName  = "send_location"
	ActionCTAURL                     ActionName  = "cta_url"
	ActionButtonReply                ActionName  = "reply"
	ActionFlow                       ActionName  = "flow"
)

type (
	MessageType string
	ActionName  string

	Button struct {
		Type  string       `json:"type,omitempty"`
		Title string       `json:"title,omitempty"`
		ID    string       `json:"id,omitempty"`
		Reply *ReplyButton `json:"reply,omitempty"`
	}

	ReplyButton struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	// SectionRow contains information about a row in an interactive section.
	SectionRow struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	Product struct {
		RetailerID string `json:"product_retailer_id,omitempty"`
	}

	Section struct {
		Title        string        `json:"title,omitempty"`
		ProductItems []*Product    `json:"product_items,omitempty"`
		Rows         []*SectionRow `json:"rows,omitempty"`
	}

	Action struct {
		Button            string            `json:"button,omitempty"`
		Buttons           []*Button         `json:"buttons,omitempty"`
		CatalogID         string            `json:"catalog_id,omitempty"`
		ProductRetailerID string            `json:"product_retailer_id,omitempty"`
		Sections          []*Section        `json:"sections,omitempty"`
		Name              string            `json:"name,omitempty"`
		Parameters        *ActionParameters `json:"parameters,omitempty"`
		Cards             []*Card           `json:"cards,omitempty"`
	}

	Card struct {
		CardIndex int         `json:"card_index,omitempty"`
		Type      string      `json:"type,omitempty"`
		Action    *CardAction `json:"action,omitempty"`
		Header    *CardHeader `json:"header,omitempty"`
		Body      *CardBody   `json:"body,omitempty"`
	}

	CardAction struct {
		Name              string         `json:"name,omitempty"`
		ProductRetailerID string         `json:"product_retailer_id,omitempty"`
		CatalogID         string         `json:"catalog_id,omitempty"`
		Parameters        map[string]any `json:"parameters,omitempty"`
		Buttons           []*CardButton  `json:"buttons,omitempty"`
	}

	CardButton struct {
		Type       string          `json:"type,omitempty"`
		QuickReply *CardQuickReply `json:"quick_reply,omitempty"`
	}

	CardQuickReply struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	CardHeader struct {
		Type  string           `json:"type,omitempty"`
		Image *CardHeaderImage `json:"image,omitempty"`
		Video *CardHeaderVideo `json:"video,omitempty"`
	}

	CardBody struct {
		Text string `json:"text,omitempty"`
	}

	CardHeaderImage struct {
		Link string `json:"link,omitempty"`
	}

	CardHeaderVideo struct {
		Link string `json:"link,omitempty"`
	}

	ActionParameters struct {
		DisplayText        string             `json:"display_text,omitempty"`
		URL                string             `json:"url,omitempty"`
		FlowMessageVersion string             `json:"flow_message_version"`
		FlowToken          string             `json:"flow_token"`
		FlowID             string             `json:"flow_id"`
		FlowCTA            string             `json:"flow_cta"`
		FlowAction         string             `json:"flow_action"`
		FlowActionPayload  *FlowActionPayload `json:"flow_action_payload"`
		Country            string             `json:"country,omitempty"`
		Values             map[string]any     `json:"values,omitempty"`
		ValidationErrors   map[string]any     `json:"validation_errors,omitempty"`
		SavedAddresses     []*SavedAddress    `json:"saved_addresses,omitempty"`
	}

	SavedAddress struct {
		ID    string         `json:"id,omitempty"`
		Value map[string]any `json:"value,omitempty"`
	}

	FlowActionPayload struct {
		Screen string         `json:"screen"`
		Data   map[string]any `json:"data"`
	}

	Header struct {
		Document *media.Document `json:"document,omitempty"`
		Image    *media.Image    `json:"image,omitempty"`
		Video    *media.Video    `json:"video,omitempty"`
		Text     string          `json:"text,omitempty"`
		Type     string          `json:"type,omitempty"`
	}

	// Footer contains information about an interactive footer.
	Footer struct {
		Text string `json:"text,omitempty"`
	}

	// Body contains information about an interactive body.
	Body struct {
		Text string `json:"text,omitempty"`
	}

	Message struct {
		Type   string  `json:"type,omitempty"`
		Action *Action `json:"action,omitempty"`
		Body   *Body   `json:"body,omitempty"`
		Footer *Footer `json:"footer,omitempty"`
		Header *Header `json:"header,omitempty"`
	}

	Option func(*Message)
)

// HeaderImage returns a header with an image asset.
func HeaderImage(img *media.Image) *Header {
	return &Header{Image: img, Type: "image"}
}

// HeaderVideo returns a header with a video asset.
func HeaderVideo(vid *media.Video) *Header {
	return &Header{Video: vid, Type: "video"}
}

// HeaderDocument returns a header with a document asset.
func HeaderDocument(doc *media.Document) *Header {
	return &Header{Document: doc, Type: "document"}
}

// HeaderText returns a header with plain text. Use when the interactive
// type only supports text headers (e.g., list messages).
func HeaderText(text string) *Header {
	return &Header{Text: text, Type: "text"}
}

// NewMessageContent creates a Message from an interactive type and options.
// Most callers should use the typed constructors (List, CTAURLButton, etc.).
func NewMessageContent(typ MessageType, opts ...Option) *Message {
	m := &Message{Type: string(typ)}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	return m
}

// WithBody sets the message body text. URLs in the text are
// automatically hyperlinked by WhatsApp.
func WithBody(text string) Option {
	return func(m *Message) { m.Body = &Body{Text: text} }
}

// WithFooter sets optional footer text. URLs are automatically hyperlinked.
func WithFooter(text string) Option {
	return func(m *Message) { m.Footer = &Footer{Text: text} }
}

// WithHeader sets the message header. Use [HeaderImage], [HeaderVideo],
// [HeaderDocument], or [HeaderText] to construct the header.
func WithHeader(header *Header) Option {
	return func(m *Message) { m.Header = header }
}

// WithAction sets the interactive action payload.
func WithAction(action *Action) Option {
	return func(m *Message) { m.Action = action }
}

// CTAURLRequest describes a Call-to-Action URL button message.
// The button maps a URL to a label so users don't see raw URLs.
type CTAURLRequest struct {
	DisplayText string  // Button label (max 20 chars)
	URL         string  // URL opened in the device browser when tapped
	Body        string  // Message body (max 1024 chars)
	Header      *Header // Optional; supports image, video, document, or text
	Footer      string  // Optional footer (max 60 chars)
}

// CTAURLButton creates a CTA URL button message. When tapped, the URL
// opens in the user's default browser — no webhook is sent for the tap.
func CTAURLButton(req *CTAURLRequest) *Message {
	return NewMessageContent(
		MessageTypeCTAURL,
		WithHeader(req.Header),
		WithBody(req.Body),
		WithFooter(req.Footer),
		WithAction(&Action{
			Name: string(ActionCTAURL),
			Parameters: &ActionParameters{
				DisplayText: req.DisplayText,
				URL:         req.URL,
			},
		}),
	)
}

// ListRequest describes an interactive list message. Users tap a button
// to reveal a picker with selectable rows organized into sections.
//
// Supports up to 10 sections with 10 rows total. Only text headers are
// supported. When a user selects a row, a list_reply webhook is sent
// containing the row's ID, title, and description.
type ListRequest struct {
	Body     string     // Message body (max 4096 chars)
	Button   string     // Button label that reveals the picker (max 20 chars)
	Footer   string     // Optional footer (max 60 chars)
	Header   string     // Optional header text (max 60 chars)
	Sections []*Section // Sections with rows (required, min 1)
}

// List creates an interactive list message. When the user taps the
// button and selects a row, the row ID is returned via webhook.
func List(req *ListRequest) *Message {
	return NewMessageContent(
		MessageTypeList,
		WithBody(req.Body),
		WithHeader(HeaderText(req.Header)),
		WithFooter(req.Footer),
		WithAction(&Action{
			Button:   req.Button,
			Sections: req.Sections,
		}),
	)
}

// ReplyButtonsRequest describes an interactive reply buttons message with
// up to 3 predefined reply options. When a user taps a button, a
// button_reply webhook is sent with the button's ID and title.
type ReplyButtonsRequest struct {
	Buttons []*ReplyButton // Reply buttons (1–3, IDs max 256 chars, titles max 20 chars)
	Body    string         // Message body (max 1024 chars)
	Header  *Header        // Optional; supports image, video, document, or text
	Footer  string         // Optional footer (max 60 chars)
}

// ReplyButtons creates an interactive reply buttons message. Supports up
// to 3 buttons; all must have unique IDs. When tapped, the button's ID
// and title are sent via a button_reply webhook.
func ReplyButtons(req *ReplyButtonsRequest) *Message {
	buttons := make([]*Button, 0, len(req.Buttons))
	for _, b := range req.Buttons {
		buttons = append(buttons, &Button{
			Type:  string(ActionButtonReply),
			Reply: b,
		})
	}
	return NewMessageContent(
		MessageTypeButton,
		WithAction(&Action{Buttons: buttons}),
		WithFooter(req.Footer),
		WithBody(req.Body),
		WithHeader(req.Header),
	)
}

// FlowRequest describes an interactive flow message that embeds a
// WhatsApp Flow into the conversation.
type FlowRequest struct {
	Body               string         // Message body
	Header             *Header        // Optional header
	Footer             string         // Optional footer
	FlowMessageVersion string         // Flow API version
	FlowToken          string         // Flow token for the session
	FlowID             string         // Flow identifier
	FlowCTA            string         // CTA button label for the flow
	FlowAction         string         // Flow action name
	FlowScreen         string         // Initial screen to render
	FlowData           map[string]any // Data payload passed to the flow
}

// Flow creates an interactive flow message.
func Flow(req *FlowRequest) *Message {
	return NewMessageContent(
		MessageTypeFlow,
		WithFooter(req.Footer),
		WithHeader(req.Header),
		WithBody(req.Body),
		WithAction(&Action{
			Name: string(ActionFlow),
			Parameters: &ActionParameters{
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
}

// AddressMessageRequest describes an interactive address message that
// collects a physical address from the user.
type AddressMessageRequest struct {
	Body             string          // Message body
	Footer           string          // Optional footer
	Header           *Header         // Optional header
	Country          string          // Country code for address format
	Values           map[string]any  // Pre-filled address values
	ValidationErrors map[string]any  // Validation errors from a previous submission
	SavedAddresses   []*SavedAddress // Previously saved addresses
}

// AddressMessage creates an interactive address message.
func AddressMessage(req *AddressMessageRequest) *Message {
	return NewMessageContent(
		MessageTypeAddressMessage,
		WithBody(req.Body),
		WithFooter(req.Footer),
		WithHeader(req.Header),
		WithAction(&Action{
			Name: "address_message",
			Parameters: &ActionParameters{
				Country:          req.Country,
				Values:           req.Values,
				ValidationErrors: req.ValidationErrors,
				SavedAddresses:   req.SavedAddresses,
			},
		}),
	)
}

// MediaCarouselCard describes a single card in a media carousel.
//
// Each card must have either an image or video header (set HeaderType
// to "image" or "video" and HeaderLink to the media URL). Body text is
// optional.
//
// For a URL button, set URLButtonLabel and URLButtonURL. For quick-reply
// buttons, populate QuickReplyButtons instead — this takes priority.
// All cards in the same carousel must use the same button type.
type MediaCarouselCard struct {
	HeaderType        string                // "image" or "video"
	HeaderLink        string                // Publicly accessible media URL
	BodyText          string                // Optional card body (max 160 chars)
	URLButtonLabel    string                // URL button label (max 20 chars)
	URLButtonURL      string                // URL to open when tapped
	QuickReplyButtons []MediaCarouselButton // Quick-reply buttons (takes priority over URL)
}

// MediaCarouselButton describes a quick-reply button on a carousel card.
type MediaCarouselButton struct {
	ID    string // Unique button ID (max 256 chars)
	Title string // Button label (max 20 chars)
}

// MediaCarousel creates an interactive media carousel with 2–10
// horizontally scrollable cards. Each card can display an image or video
// header and either a URL button or quick-reply buttons.
//
// Cards may mix image and video headers within the same carousel.
func MediaCarousel(body string, cards []*MediaCarouselCard) *Message {
	icards := make([]*Card, 0, len(cards))
	for i, card := range cards {
		c := &Card{
			CardIndex: i,
			Type:      string(MessageTypeCTAURL),
			Body:      &CardBody{Text: card.BodyText},
		}

		switch card.HeaderType {
		case "image":
			c.Header = &CardHeader{
				Type:  "image",
				Image: &CardHeaderImage{Link: card.HeaderLink},
			}
		case "video":
			c.Header = &CardHeader{
				Type:  "video",
				Video: &CardHeaderVideo{Link: card.HeaderLink},
			}
		}

		if len(card.QuickReplyButtons) > 0 {
			btns := make([]*CardButton, 0, len(card.QuickReplyButtons))
			for _, b := range card.QuickReplyButtons {
				btns = append(btns, &CardButton{
					Type: "quick_reply",
					QuickReply: &CardQuickReply{
						ID:    b.ID,
						Title: b.Title,
					},
				})
			}
			c.Action = &CardAction{Buttons: btns}
		} else {
			c.Action = &CardAction{
				Name: string(ActionCTAURL),
				Parameters: map[string]any{
					"url":          card.URLButtonURL,
					"display_text": card.URLButtonLabel,
				},
			}
		}

		icards = append(icards, c)
	}

	return NewMessageContent(
		MessageTypeCarousel,
		WithBody(body),
		WithAction(&Action{Cards: icards}),
	)
}

// LocationRequest creates an interactive location request message.
// WhatsApp displays a "Send Location" button that prompts the user
// to share their current location.
func LocationRequest(body string) *Message {
	return NewMessageContent(
		MessageTypeLocationRequest,
		WithBody(body),
		WithAction(&Action{Name: string(ActionSendLocation)}),
	)
}

// CallPermissionRequest creates an interactive call permission request.
// WhatsApp displays a button that initiates a voice call from the user.
func CallPermissionRequest(body string) *Message {
	return &Message{
		Type:   string(MessageTypeCallPermissionRequest),
		Action: &Action{Name: "call_permission_request"},
		Body:   &Body{Text: body},
	}
}

// AddressDetails represents a structured physical address returned in an
// address_message webhook. All fields map to WhatsApp's address form.
//
// Required fields: Name, PhoneNumber, PinCode, City, State.
// PhoneNumber expects Indian mobile format (optional +91/91 prefix,
// 10 digits starting 6–9). PinCode must be exactly 6 numeric digits.
type AddressDetails struct {
	Name         string `json:"name"`
	PhoneNumber  string `json:"phone_number"`
	PinCode      string `json:"in_pin_code"`
	HouseNumber  string `json:"house_number,omitempty"`
	FloorNumber  string `json:"floor_number,omitempty"`
	TowerNumber  string `json:"tower_number,omitempty"`
	BuildingName string `json:"building_name,omitempty"`
	Address      string `json:"address,omitempty"`
	LandmarkArea string `json:"landmark_area,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
}
