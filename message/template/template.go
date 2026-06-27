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

// Package template provides types and constructors for WhatsApp message
// templates — the only message type that can be sent outside customer service
// windows.
//
// Templates must be approved before sending. Categories are authentication,
// marketing, or utility. Parameters use named ({{first_name}}) or positional
// ({{1}}) format. Each WABA can have up to 250 templates (6,000 if verified),
// with a creation limit of 100 per hour.
package template

import (
	"github.com/piusalfred/whatsapp/message/media"
)

const (
	TemplateComponentTypeCarousel         = "carousel"
	TemplateComponentTypeHeader           = "header"
	TemplateComponentTypeButton           = "button"
	TemplateComponentTypeButtons          = "buttons"
	TemplateComponentTypeBody             = "body"
	TemplateParameterTypeText             = "text"
	TemplateParameterTypeCurrency         = "currency"
	TemplateParameterTypeDateTime         = "date_time"
	TemplateParameterTypePayload          = "payload"
	TemplateParameterTypeImage            = "image"
	TemplateParameterTypeVideo            = "video"
	TemplateParameterTypeDocument         = "document"
	TemplateParameterTypeLocation         = "location"
	TemplateComponentTypeLimitedTimeOffer = "limited_time_offer"
	TemplateButtonSubTypeCopyCode         = "copy_code"
	TemplateButtonSubTypeURL              = "url"

	// HeaderFormatLocation declares the header component as a location map.
	// Used when creating location templates.
	HeaderFormatLocation = "LOCATION"
)

type (
	Button struct {
		Type           string `json:"type,omitempty"`
		Payload        string `json:"payload,omitempty"`
		Text           string `json:"text,omitempty"`
		FlowID         string `json:"flow_id"`
		NavigateScreen string `json:"navigate_screen"`
		FlowAction     string `json:"flow_action"`
	}

	Template struct {
		Name       string       `json:"name,omitempty"`
		Language   *Language    `json:"language,omitempty"`
		Category   string       `json:"category,omitempty"`
		Components []*Component `json:"components,omitempty"`
	}

	Language struct {
		Code   string `json:"code"`
		Policy string `json:"policy,omitempty"`
	}

	Component struct {
		Type       string       `json:"type"`
		Format     string       `json:"format,omitempty"`
		SubType    string       `json:"sub_type,omitempty"`
		Index      int          `json:"index,omitempty"`
		Parameters []*Parameter `json:"parameters,omitempty"`
		Buttons    []*Button    `json:"buttons,omitempty"`
		Text       string       `json:"text,omitempty"`
		Cards      []*MediaCard `json:"cards,omitempty"`
	}

	Parameter struct {
		Type             string            `json:"type"`
		Text             string            `json:"text,omitempty"`
		Name             string            `json:"parameter_name,omitempty"`
		Payload          string            `json:"payload,omitempty"`
		Currency         *Currency         `json:"currency,omitempty"`
		DateTime         *DateTime         `json:"date_time,omitempty"`
		LimitedTimeOffer *LimitedTimeOffer `json:"limited_time_offer,omitempty"`
		Image            *media.Image      `json:"image,omitempty"`
		Document         *media.Document   `json:"document,omitempty"`
		Video            *media.Video      `json:"video,omitempty"`
		Location         *media.Location   `json:"location,omitempty"`
		CouponCode       string            `json:"coupon_code,omitempty"`
	}

	LimitedTimeOffer struct {
		ExpirationTimeMs int64 `json:"expiration_time_ms"`
	}

	Currency struct {
		FallbackValue string  `json:"fallback_value"`
		Code          string  `json:"code"`
		Amount1000    float64 `json:"amount_1000"`
	}

	DateTime struct {
		FallbackValue string `json:"fallback_value"`
		DayOfWeek     int    `json:"day_of_week"`
		Year          int    `json:"year"`
		Month         int    `json:"month"`
		DayOfMonth    int    `json:"day_of_month"`
		Hour          int    `json:"hour"`
		Minute        int    `json:"minute"`
		Calendar      string `json:"calendar,omitempty"`
	}

	InteractiveButtonTemplate struct {
		SubType string
		Index   int
		Button  *Button
	}

	FlowButton struct {
		Type           string `json:"type"`
		Text           string `json:"text"`
		FlowID         string `json:"flow_id"`
		NavigateScreen string `json:"navigate_screen"`
		FlowAction     string `json:"flow_action"`
	}
)

// NewLocationHeader returns a location header component for template creation.
// The actual coordinates (latitude, longitude, name, address) are provided at
// send time via [Parameter] with Type=[TemplateParameterTypeLocation] and
// a [media.Location] value.
func NewLocationHeader() *Component {
	return &Component{
		Type:   TemplateComponentTypeHeader,
		Format: HeaderFormatLocation,
	}
}

// NewLocationParameter creates a location parameter for send time. Use this
// with a location template built via [NewLocationHeader] to provide the
// actual coordinates at message send time.
func NewLocationParameter(loc *media.Location) *Parameter {
	return &Parameter{
		Type:     TemplateParameterTypeLocation,
		Location: loc,
	}
}

func NewInteractiveTemplate(name string, language *Language, headers []*Parameter,
	bodies []*Parameter, buttons []*InteractiveButtonTemplate,
) *Template {
	components := make([]*Component, 0, 2+len(buttons)) //nolint:mnd // 2 = fixed header + body components
	headerTemplate := &Component{
		Type:       TemplateComponentTypeHeader,
		Parameters: headers,
	}
	components = append(components, headerTemplate)

	bodyTemplate := &Component{
		Type:       TemplateComponentTypeBody,
		Parameters: bodies,
	}
	components = append(components, bodyTemplate)

	for _, button := range buttons {
		b := &Component{
			Type:    TemplateComponentTypeButton,
			SubType: button.SubType,
			Index:   button.Index,
			Parameters: []*Parameter{
				{
					Type:    button.Button.Type,
					Text:    button.Button.Text,
					Payload: button.Button.Payload,
				},
			},
		}

		components = append(components, b)
	}

	return &Template{
		Name:       name,
		Language:   language,
		Components: components,
	}
}

const (
	TemplateComponentButtonSubTypeURL = "url"
)

type AuthTemplateRequest struct {
	Name            string
	LanguageCode    string
	LanguagePolicy  string
	OneTimePassword string
}

func NewAuthTemplate(request *AuthTemplateRequest) *Template {
	parameter := &Parameter{
		Type: TemplateParameterTypeText,
		Text: request.OneTimePassword,
	}

	bodyComponent := &Component{
		Type:       TemplateComponentTypeBody,
		Parameters: []*Parameter{parameter},
	}

	buttonComponent := &Component{
		Type:       TemplateComponentTypeButton,
		SubType:    TemplateComponentButtonSubTypeURL,
		Index:      0,
		Parameters: []*Parameter{parameter},
	}

	tmpl := &Template{
		Name: request.Name,
		Language: &Language{
			Code:   request.LanguageCode,
			Policy: request.LanguagePolicy,
		},
		Components: []*Component{bodyComponent, buttonComponent},
	}

	return tmpl
}

type (
	MediaCardTemplateRequest struct {
		Name     string
		Language *Language
		BodyText string
		Cards    []*MediaCard
		Category string
	}

	MediaCard struct {
		Header  *MediaCardHeader `json:"header"`
		Body    *MediaCardBody   `json:"body"`
		Buttons []*Button        `json:"buttons,omitempty"`
	}

	MediaCardHeader struct {
		Format string `json:"format"`
		Handle string `json:"handle"`
	}

	MediaCardBody struct {
		Text string `json:"text"`
	}
)

func NewMediaCardTemplate(req *MediaCardTemplateRequest) *Template {
	return &Template{
		Name:     req.Name,
		Language: req.Language,
		Category: req.Category,
		Components: []*Component{
			{
				Type: TemplateComponentTypeBody,
				Text: req.BodyText,
			},
			{
				Type:  TemplateComponentTypeCarousel,
				Cards: req.Cards,
			},
		},
	}
}

type LimitedTimeOfferTemplateRequest struct {
	Name            string
	Language        *Language
	HeaderComponent *Component
	Body            []*Parameter
	ExpirationTime  int64
	CouponCode      *string
	URLVariable     string
}

func NewLimitedTimeOfferTemplateImageHeader(image *media.Image) *Component {
	return &Component{
		Type: TemplateComponentTypeHeader,
		Parameters: []*Parameter{
			{
				Type:  "image",
				Image: image,
			},
		},
	}
}

func NewLimitedTimeOfferTemplateDocumentHeader(document *media.Document) *Component {
	return &Component{
		Type: TemplateComponentTypeHeader,
		Parameters: []*Parameter{
			{
				Type:     "document",
				Document: document,
			},
		},
	}
}

func NewTemplateComponentLimitedTimeOffer(expiresAt int64) *Component {
	cmp := &Component{
		Type:    TemplateComponentTypeLimitedTimeOffer,
		SubType: "",
		Index:   0,
		Parameters: []*Parameter{
			{
				Type:             TemplateComponentTypeLimitedTimeOffer,
				LimitedTimeOffer: &LimitedTimeOffer{ExpirationTimeMs: expiresAt},
			},
		},
	}

	return cmp
}

type ButtonParams struct {
	Index int
	Text  string
}

func NewCopyCodeButton(params *ButtonParams) *Component {
	return &Component{
		Type:    TemplateComponentTypeButton,
		SubType: TemplateButtonSubTypeCopyCode,
		Index:   params.Index,
		Parameters: []*Parameter{
			{
				Type: "coupon_code",
				Text: params.Text,
			},
		},
	}
}

func NewURLButton(params *ButtonParams) *Component {
	return &Component{
		Type:    TemplateComponentTypeButton,
		SubType: TemplateButtonSubTypeURL,
		Index:   params.Index,
		Parameters: []*Parameter{
			{
				Type: "text",
				Text: params.Text,
			},
		},
	}
}

func NewLimitedTimeOfferTemplate(req *LimitedTimeOfferTemplateRequest) *Template {
	components := []*Component{
		req.HeaderComponent,
		{
			Type:       TemplateComponentTypeBody,
			Parameters: req.Body,
		},
		NewTemplateComponentLimitedTimeOffer(req.ExpirationTime),
	}

	urlButtonIndex := 0
	if req.CouponCode != nil {
		components = append(components, NewCopyCodeButton(&ButtonParams{
			Index: urlButtonIndex,
			Text:  *req.CouponCode,
		}))
		urlButtonIndex = 1
	}

	components = append(components, NewURLButton(&ButtonParams{
		Index: urlButtonIndex,
		Text:  req.URLVariable,
	}))

	return &Template{
		Name:       req.Name,
		Language:   req.Language,
		Components: components,
	}
}

type CouponCodeTemplateRequest struct {
	Name        string       // Template name
	Language    *Language    // Language code and policy
	Body        []*Parameter // Parameters for the body text
	CouponCode  string       // Coupon code to be copied
	ButtonIndex int          // Index of the button in the template
}

func NewCouponCodeTemplate(req *CouponCodeTemplateRequest) *Template {
	components := []*Component{
		{
			Type:       TemplateComponentTypeBody,
			Parameters: req.Body,
		},

		NewCopyCodeButton(&ButtonParams{
			Index: req.ButtonIndex,
			Text:  req.CouponCode,
		}),
	}

	return &Template{
		Name:       req.Name,
		Language:   req.Language,
		Components: components,
	}
}
