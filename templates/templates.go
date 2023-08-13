/*
Package templates provides structures and utilities for creating and manipulating WhatsApp message
templates.

WhatsApp message templates are specific message formats that businesses use to send out notifications
or customer care messages to people that have opted in to notifications. These notifications can include
a variety of messages such as appointment reminders, shipping information, issue resolution, or payment
updates.

Before using this package to send a template message, you must first create a template. For more information
on creating templates, refer to the guide titled "Create Message Templates for Your WhatsApp Business Account"
(https://developers.facebook.com/micro_site/url/?click_from_context_menu=true&country=TZ&destination=https%3A%2F%2Fwww.facebook.com%2Fbusiness%2Fhelp%2F2055875911147364&event_type=click&last_nav_impression_id=0jjgfFiSZMkMJ8TP8&max_percent_page_viewed=10&max_viewport_height_px=999&max_viewport_width_px=1414&orig_http_referrer=https%3A%2F%2Fdevelopers.facebook.com%2Fdocs%2Fwhatsapp%2Fcloud-api%2Fguides%2Fsend-message-templates&orig_request_uri=https%3A%2F%2Fdevelopers.facebook.com%2Fajax%2Fdocs%2Fnav%2F%3Fpath1%3Dwhatsapp%26path2%3Dcloud-api%26path3%3Dguides%26path4%3Dsend-message-templates&region=emea&scrolled=false&session_id=1Qi3IoIHE5nPFSkug&site=developers)
If your account is not yet verified, you can use one of the pre-approved templates provided.

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
