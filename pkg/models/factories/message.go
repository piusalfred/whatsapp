package factories

import "github.com/piusalfred/whatsapp/pkg/models"

type (
	MessageOption func(*models.Message)
)

func NewMessage(recipient string, options ...MessageOption) *models.Message {
	message := &models.Message{
		Product:       "whatsapp",
		RecipientType: "individual",
		To:            recipient,
	}
	for _, option := range options {
		option(message)
	}

	return message
}

func WithMessageTemplate(template *models.Template) MessageOption {
	return func(m *models.Message) {
		m.Type = "template"
		m.Template = template
	}
}

func WithMessageText(text *models.Text) MessageOption {
	return func(m *models.Message) {
		m.Type = "text"
		m.Text = text
	}
}
