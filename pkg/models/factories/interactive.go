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

import "github.com/piusalfred/whatsapp/pkg/models"

type CTAButtonURLParameters struct {
	DisplayText string
	URL         string
	Body        string
	Footer      string
	Header      string
}

func LocationRequestMessage(recipient, reply, message string) *models.Message {
	//{
	//  "messaging_product": "whatsapp",
	//  "recipient_type": "individual",
	//  "type": "interactive",
	//  "to": "+15551234567",
	//  "interactive": {
	//    "type": "location_request_message",
	//    "body": {
	//      "text": "Let us start with your pickup. You can either manually *enter an address* or *share your current location*."
	//    },
	//    "action": {
	//      "name": "send_location"
	//    }
	//  }
	//}'

	i := &models.Interactive{
		Type: InteractiveLocationRequest,
		Action: &models.InteractiveAction{
			Name: "send_location",
		},
		Body:   &models.InteractiveBody{Text: message},
		Footer: nil,
		Header: nil,
	}

	m := &models.Message{
		Product:       MessagingProductWhatsApp,
		RecipientType: RecipientTypeIndividual,
		Type:          "interactive",
		To:            recipient,
		Interactive:   i,
	}

	if reply != "" {
		m.Context = &models.Context{MessageID: reply}
	}

	return m
}

func NewInteractiveCTAURLButton(parameters *CTAButtonURLParameters) *models.Interactive {
	return &models.Interactive{
		Type: InteractiveMessageCTAButton,
		Action: &models.InteractiveAction{
			Name: InteractiveMessageCTAButton,
			Parameters: &models.InteractiveActionParameters{
				URL:         parameters.URL,
				DisplayText: parameters.DisplayText,
			},
		},
		Body:   &models.InteractiveBody{Text: parameters.Body},
		Footer: &models.InteractiveFooter{Text: parameters.Footer},
		Header: InteractiveHeaderText(parameters.Header),
	}
}

func InteractiveHeaderText(text string) *models.InteractiveHeader {
	return &models.InteractiveHeader{
		Type: "text",
		Text: text,
	}
}

func InteractiveHeaderImage(image *models.Media) *models.InteractiveHeader {
	return &models.InteractiveHeader{
		Type:  "image",
		Image: image,
	}
}

func InteractiveHeaderVideo(video *models.Media) *models.InteractiveHeader {
	return &models.InteractiveHeader{
		Type:  "video",
		Video: video,
	}
}

func InteractiveHeaderDocument(document *models.Media) *models.InteractiveHeader {
	return &models.InteractiveHeader{
		Type:     "document",
		Document: document,
	}
}

// CreateInteractiveRelyButtonList creates a list of InteractiveButton with type reply, A max of
// 3 buttons can be added to a message. So do not add more than 3 buttons.
func CreateInteractiveRelyButtonList(buttons ...*models.InteractiveReplyButton) []*models.InteractiveButton {
	var list []*models.InteractiveButton
	for _, button := range buttons {
		list = append(list, &models.InteractiveButton{
			Type:  "reply",
			Reply: button,
		})
	}

	return list
}

// NewInteractiveTemplate creates a new interactive template.
func NewInteractiveTemplate(name string, language *models.TemplateLanguage, headers []*models.TemplateParameter,
	bodies []*models.TemplateParameter, buttons []*models.InteractiveButtonTemplate,
) *models.Template {
	var components []*models.TemplateComponent
	headerTemplate := &models.TemplateComponent{
		Type:       "header",
		Parameters: headers,
	}
	components = append(components, headerTemplate)

	bodyTemplate := &models.TemplateComponent{
		Type:       "body",
		Parameters: bodies,
	}
	components = append(components, bodyTemplate)

	for _, button := range buttons {
		b := &models.TemplateComponent{
			Type:    "button",
			SubType: button.SubType,
			Index:   button.Index,
			Parameters: []*models.TemplateParameter{
				{
					Type:    button.Button.Type,
					Text:    button.Button.Text,
					Payload: button.Button.Payload,
				},
			},
		}

		components = append(components, b)
	}

	return &models.Template{
		Name:       name,
		Language:   language,
		Components: components,
	}
}

func NewTextTemplate(name string, language *models.TemplateLanguage,
	parameters []*models.TemplateParameter,
) *models.Template {
	component := &models.TemplateComponent{
		Type:       "body",
		Parameters: parameters,
	}

	return &models.Template{
		Name:       name,
		Language:   language,
		Components: []*models.TemplateComponent{component},
	}
}

// NewMediaTemplate create a media based template.
func NewMediaTemplate(name string, language *models.TemplateLanguage, header *models.TemplateParameter,
	bodies []*models.TemplateParameter,
) *models.Template {
	var components []*models.TemplateComponent
	headerTemplate := &models.TemplateComponent{
		Type:       "header",
		Parameters: []*models.TemplateParameter{header},
	}
	components = append(components, headerTemplate)

	bodyTemplate := &models.TemplateComponent{
		Type:       "body",
		Parameters: bodies,
	}
	components = append(components, bodyTemplate)

	return &models.Template{
		Name:       name,
		Language:   language,
		Components: components,
	}
}

type (
	InteractiveOption func(*models.Interactive)
)

func WithInteractiveFooter(footer string) InteractiveOption {
	return func(i *models.Interactive) {
		i.Footer = &models.InteractiveFooter{
			Text: footer,
		}
	}
}

func WithInteractiveBody(body string) InteractiveOption {
	return func(i *models.Interactive) {
		i.Body = &models.InteractiveBody{
			Text: body,
		}
	}
}

func WithInteractiveHeader(header *models.InteractiveHeader) InteractiveOption {
	return func(i *models.Interactive) {
		i.Header = header
	}
}

func WithInteractiveAction(action *models.InteractiveAction) InteractiveOption {
	return func(i *models.Interactive) {
		i.Action = action
	}
}

func NewInteractiveMessage(interactiveType string, options ...InteractiveOption) *models.Interactive {
	interactive := &models.Interactive{
		Type: interactiveType,
	}
	for _, option := range options {
		option(interactive)
	}

	return interactive
}

type InteractiveHeaderType string

const (
	// InteractiveHeaderTypeText is used for ListQR Messages, Reply Buttons, and Multi-Product Messages.
	InteractiveHeaderTypeText  InteractiveHeaderType = "text"
	InteractiveHeaderTypeVideo InteractiveHeaderType = "video"
	InteractiveHeaderTypeImage InteractiveHeaderType = "image"
	InteractiveHeaderTypeDoc   InteractiveHeaderType = "document"
)
