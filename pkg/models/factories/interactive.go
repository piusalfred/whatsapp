package factories

import "github.com/piusalfred/whatsapp/pkg/models"

type CTAButtonURLParameters struct {
	DisplayText string
	URL         string
	Body        string
	Footer      string
	Header      string
}

func NewInteractiveCTAURLButton(parameters *CTAButtonURLParameters) *models.Interactive {
	return &models.Interactive{
		Type: models.InteractiveMessageCTAButton,
		Action: &models.InteractiveAction{
			Name: models.InteractiveMessageCTAButton,
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

func NewTextTemplate(name string, language *models.TemplateLanguage, parameters []*models.TemplateParameter) *models.Template {
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
