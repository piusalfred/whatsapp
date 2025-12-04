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
)

type (
	TemplateButton struct {
		Type           string `json:"type,omitempty"`
		Payload        string `json:"payload,omitempty"`
		Text           string `json:"text,omitempty"`
		FlowID         string `json:"flow_id"`
		NavigateScreen string `json:"navigate_screen"`
		FlowAction     string `json:"flow_action"`
	}

	Template struct {
		Name       string               `json:"name,omitempty"`
		Language   *TemplateLanguage    `json:"language,omitempty"`
		Category   string               `json:"category,omitempty"`
		Components []*TemplateComponent `json:"components,omitempty"`
	}

	TemplateLanguage struct {
		Code   string `json:"code"`
		Policy string `json:"policy"`
	}

	TemplateComponent struct {
		Type       string               `json:"type"`
		SubType    string               `json:"sub_type,omitempty"`
		Index      int                  `json:"index"`
		Parameters []*TemplateParameter `json:"parameters"`
		Buttons    []*TemplateButton    `json:"buttons,omitempty"`
		Text       string               `json:"text,omitempty"`
		Cards      []*MediaCard         `json:"cards,omitempty"`
	}

	TemplateParameter struct {
		Type             string            `json:"type"`
		Text             string            `json:"text"`
		Name             string            `json:"parameter_name,omitempty"`
		Payload          string            `json:"payload,omitempty"`
		Currency         *TemplateCurrency `json:"currency"`
		DateTime         *TemplateDateTime `json:"date_time"`
		LimitedTimeOffer *LimitedTimeOffer `json:"limited_time_offer,omitempty"`
		Image            *Image            `json:"image"`
		Document         *Document         `json:"document"`
		Video            *Video            `json:"video"`
		Location         *Location         `json:"location"`
		CouponCode       string            `json:"coupon_code,omitempty"`
	}

	LimitedTimeOffer struct {
		ExpirationTimeMs int64 `json:"expiration_time_ms"`
	}

	TemplateCurrency struct {
		FallbackValue string  `json:"fallback_value"`
		Code          string  `json:"code"`
		Amount1000    float64 `json:"amount_1000"`
	}

	TemplateDateTime struct {
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
		Button  *TemplateButton
	}

	TemplateFlowButton struct {
		Type           string `json:"type"`
		Text           string `json:"text"`
		FlowID         string `json:"flow_id"`
		NavigateScreen string `json:"navigate_screen"`
		FlowAction     string `json:"flow_action"`
	}
)

func WithTemplateMessage(tmpl *Template) Option {
	return func(message *Message) {
		message.Type = TypeTemplate
		message.Template = tmpl
	}
}

func NewInteractiveTemplate(name string, language *TemplateLanguage, headers []*TemplateParameter,
	bodies []*TemplateParameter, buttons []*InteractiveButtonTemplate,
) *Template {
	var components []*TemplateComponent
	headerTemplate := &TemplateComponent{
		Type:       TemplateComponentTypeHeader,
		Parameters: headers,
	}
	components = append(components, headerTemplate)

	bodyTemplate := &TemplateComponent{
		Type:       TemplateComponentTypeBody,
		Parameters: bodies,
	}
	components = append(components, bodyTemplate)

	for _, button := range buttons {
		b := &TemplateComponent{
			Type:    TemplateComponentTypeButton,
			SubType: button.SubType,
			Index:   button.Index,
			Parameters: []*TemplateParameter{
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
	parameter := &TemplateParameter{
		Type: TemplateParameterTypeText,
		Text: request.OneTimePassword,
	}

	bodyComponent := &TemplateComponent{
		Type:       TemplateComponentTypeBody,
		Parameters: []*TemplateParameter{parameter},
	}

	buttonComponent := &TemplateComponent{
		Type:       TemplateComponentTypeButton,
		SubType:    TemplateComponentButtonSubTypeURL,
		Index:      0,
		Parameters: []*TemplateParameter{parameter},
	}

	tmpl := &Template{
		Name: request.Name,
		Language: &TemplateLanguage{
			Code:   request.LanguageCode,
			Policy: request.LanguagePolicy,
		},
		Components: []*TemplateComponent{bodyComponent, buttonComponent},
	}

	return tmpl
}

type (
	MediaCardTemplateRequest struct {
		Name     string
		Language *TemplateLanguage
		BodyText string
		Cards    []*MediaCard
		Category string
	}

	MediaCard struct {
		Header  *MediaCardHeader  `json:"header"`
		Body    *MediaCardBody    `json:"body"`
		Buttons []*TemplateButton `json:"buttons,omitempty"`
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
		Components: []*TemplateComponent{
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
	Language        *TemplateLanguage
	HeaderComponent *TemplateComponent
	Body            []*TemplateParameter
	ExpirationTime  int64
	CouponCode      *string
	URLVariable     string
}

func NewLimitedTimeOfferTemplateImageHeader(image *Image) *TemplateComponent {
	return &TemplateComponent{
		Type: TemplateComponentTypeHeader,
		Parameters: []*TemplateParameter{
			{
				Type:  "image",
				Image: image,
			},
		},
	}
}

func NewLimitedTimeOfferTemplateDocumentHeader(document *Document) *TemplateComponent {
	return &TemplateComponent{
		Type: TemplateComponentTypeHeader,
		Parameters: []*TemplateParameter{
			{
				Type:     "document",
				Document: document,
			},
		},
	}
}

func NewTemplateComponentLimitedTimeOffer(expiresAt int64) *TemplateComponent {
	cmp := &TemplateComponent{
		Type:    TemplateComponentTypeLimitedTimeOffer,
		SubType: "",
		Index:   0,
		Parameters: []*TemplateParameter{
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

func NewCopyCodeButton(params *ButtonParams) *TemplateComponent {
	return &TemplateComponent{
		Type:    TemplateComponentTypeButton,
		SubType: TemplateButtonSubTypeCopyCode,
		Index:   params.Index,
		Parameters: []*TemplateParameter{
			{
				Type: "coupon_code",
				Text: params.Text,
			},
		},
	}
}

func NewURLButton(params *ButtonParams) *TemplateComponent {
	return &TemplateComponent{
		Type:    TemplateComponentTypeButton,
		SubType: TemplateButtonSubTypeURL,
		Index:   params.Index,
		Parameters: []*TemplateParameter{
			{
				Type: "text",
				Text: params.Text,
			},
		},
	}
}

func NewLimitedTimeOfferTemplate(req *LimitedTimeOfferTemplateRequest) *Template {
	components := []*TemplateComponent{
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
	Name        string               // Template name
	Language    *TemplateLanguage    // Language code and policy
	Body        []*TemplateParameter // Parameters for the body text
	CouponCode  string               // Coupon code to be copied
	ButtonIndex int                  // Index of the button in the template
}

func NewCouponCodeTemplate(req *CouponCodeTemplateRequest) *Template {
	components := []*TemplateComponent{
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
