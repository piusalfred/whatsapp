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

const (
	TypeInteractiveLocationRequest = "location_request_message"
	TypeInteractiveCTAURL          = "cta_url"
	TypeInteractiveButton          = "button"
	TypeInteractiveFlow            = "flow"
	TypeInteractiveAddressMessage  = "address_message"
	TypeInteractiveCarousel        = "carousel"
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
		Cards             []*InteractiveCard           `json:"cards,omitempty"`
	}

	InteractiveCard struct {
		CardIndex int                    `json:"card_index,omitempty"`
		Type      string                 `json:"type,omitempty"`
		Action    *InteractiveCardAction `json:"action,omitempty"`
		Header    *InteractiveCardHeader `json:"header,omitempty"`
		Body      *InteractiveCardBody   `json:"body,omitempty"`
	}

	InteractiveCardAction struct {
		Name              string         `json:"name,omitempty"`
		ProductRetailerID string         `json:"product_retailer_id,omitempty"`
		CatalogID         string         `json:"catalog_id,omitempty"`
		Parameters        map[string]any `json:"parameters,omitempty"`
	}

	InteractiveCardHeader struct {
		Type  string                      `json:"type,omitempty"`
		Image *InteractiveCardHeaderImage `json:"image,omitempty"`
		Video *InteractiveCardHeaderVideo `json:"video,omitempty"`
	}

	InteractiveCardBody struct {
		Text string `json:"text,omitempty"`
	}

	InteractiveCardHeaderImage struct {
		Link string `json:"link,omitempty"`
	}

	InteractiveCardHeaderVideo struct {
		Link string `json:"link,omitempty"`
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

type AddressMessageParams struct {
	Recipient        string         `json:"recipient"`
	ReplyTo          string         `json:"reply_to"`
	Country          string         `json:"country"`
	Values           map[string]any `json:"values"`
	ValidationErrors map[string]any `json:"validation_errors"`
}

type InteractiveAddressMessageRequest struct {
	Body             string             `json:"body"`
	Footer           string             `json:"footer"`
	Header           *InteractiveHeader `json:"header"`
	Country          string             `json:"country"`
	Values           map[string]any     `json:"values,omitempty"`
	ValidationErrors map[string]any     `json:"validation_errors,omitempty"`
	SavedAddresses   []*SavedAddress    `json:"saved_addresses,omitempty"`
}

func NewInteractiveAddressMessage(request *InteractiveAddressMessageRequest) *Interactive {
	return NewInteractiveMessageContent(
		TypeInteractiveAddressMessage,
		WithInteractiveBody(request.Body),
		WithInteractiveFooter(request.Footer),
		WithInteractiveHeader(request.Header),
		WithInteractiveAction(&InteractiveAction{
			Name: "address_message",
			Parameters: &InteractiveActionParameters{
				Country:          request.Country,
				Values:           request.Values,
				ValidationErrors: request.ValidationErrors,
				SavedAddresses:   request.SavedAddresses,
			},
		}),
	)
}

type InteractiveProductCarouselCard struct {
	Index             int    `json:"index"`
	ProductRetailerID string `json:"product_retailer_id"`
	CatalogID         string `json:"catalog_id"`
	HeaderMediaLink   string `json:"header_media_link"`
	BodyText          string `json:"body_text"`
	CardDisplayText   string `json:"card_display_text"`
	CardButtonURL     string `json:"card_button_url"`
}

func NewInteractiveProductCarousel(message string, cards []InteractiveProductCarouselCard) *Interactive {
	productCards := make([]*InteractiveCard, 0, len(cards))
	for _, card := range cards {
		c := &InteractiveCard{
			CardIndex: card.Index,
			Type:      "product",
			Action: &InteractiveCardAction{
				ProductRetailerID: card.ProductRetailerID,
				CatalogID:         card.CatalogID,
			},
		}

		productCards = append(productCards, c)
	}

	return NewInteractiveMessageContent(
		TypeInteractiveCarousel,
		WithInteractiveBody(message),
		WithInteractiveAction(&InteractiveAction{
			Cards: productCards,
		}),
	)
}

func NewInteractiveImageCarousel(message string, cards []InteractiveProductCarouselCard) *Interactive {
	productCards := make([]*InteractiveCard, 0, len(cards))
	for _, card := range cards {
		c := &InteractiveCard{
			CardIndex: card.Index,
			Type:      "cta_url",
			Header: &InteractiveCardHeader{
				Type: "image",
				Image: &InteractiveCardHeaderImage{
					Link: card.HeaderMediaLink,
				},
			},
			Body: &InteractiveCardBody{
				Text: card.BodyText,
			},
			Action: &InteractiveCardAction{
				Name: "cta_url",
				Parameters: map[string]any{
					"url":          card.CardButtonURL,
					"display_text": card.CardDisplayText,
				},
			},
		}

		productCards = append(productCards, c)
	}

	return NewInteractiveMessageContent(
		TypeInteractiveCarousel,
		WithInteractiveBody(message),
		WithInteractiveAction(&InteractiveAction{
			Cards: productCards,
		}),
	)
}

func NewInteractiveVideoCarousel(message string, cards []InteractiveProductCarouselCard) *Interactive {
	productCards := make([]*InteractiveCard, 0, len(cards))
	for _, card := range cards {
		c := &InteractiveCard{
			CardIndex: card.Index,
			Type:      "cta_url",
			Header: &InteractiveCardHeader{
				Type: "video",
				Video: &InteractiveCardHeaderVideo{
					Link: card.HeaderMediaLink,
				},
			},
			Body: &InteractiveCardBody{
				Text: card.BodyText,
			},
			Action: &InteractiveCardAction{
				Name: "cta_url",
				Parameters: map[string]any{
					"url":          card.CardButtonURL,
					"display_text": card.CardDisplayText,
				},
			},
		}

		productCards = append(productCards, c)
	}

	return NewInteractiveMessageContent(
		TypeInteractiveCarousel,
		WithInteractiveBody(message),
		WithInteractiveAction(&InteractiveAction{
			Cards: productCards,
		}),
	)
}

// AddressDetails represents a structured address
//
// Validation overview:
//   - Name: required, non-empty.
//   - PhoneNumber: required, Indian mobile format (optional +91/91 prefix; 10 digits starting 6–9).
//   - PinCode: required, exactly 6 numeric digits.
//   - City: required, non-empty.
//   - State: required, non-empty.
//   - All other fields are optional free text.
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
