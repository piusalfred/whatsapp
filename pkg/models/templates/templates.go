/*
Package templates provides structures and utilities for creating and manipulating WhatsApp message
templates.

WhatsApp message templates are specific message formats that businesses use to send out notifications
or customer care messages to people that have opted in to notifications. These notifications can include
a variety of messages such as appointment reminders, shipping information, issue resolution, or payment
updates.

This package supports the following template types:

- Text-based message templates
- Media-based message templates
- Interactive message templates
- Location-based message templates
- Authentication templates with one-time password buttons
- Multi-Product Message templates

All API calls made using this package must be authenticated with an access token.Developers can authenticate
their API calls with the access token generated in the App Dashboard > WhatsApp > API Setup panel.
Business Solution Providers (BSPs) need to authenticate themselves with an access token that has the
'whatsapp_business_messaging' permission.
*/
package templates

type Message struct {
	MessagingProduct string    `json:"messaging_product"`
	RecipientType    string    `json:"recipient_type"`
	To               string    `json:"to"`
	Type             string    `json:"type"`
	Template         *Template `json:"template"`
}

type Template struct {
	Name       string      `json:"name"`
	Language   Language    `json:"language"`
	Components []Component `json:"components"`
}

type Language struct {
	Code string `json:"code"`
}

type Component struct {
	Type       string      `json:"type"`
	SubType    string      `json:"sub_type,omitempty"`
	Index      string      `json:"index,omitempty"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
