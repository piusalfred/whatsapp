package whatsapp

type (

	/*
		Template is a template for a message. It contains the parameters of the message as listed below.

		   - Name, name (string). Required. Name of the template.

		   - Language, language (object). Required. Contains a language object. Specifies the language the template may be rendered in.
		     The language object can contain the following fields:
			 	- Policy, policy (string). Required. The language policy the message should follow. The only supported option is deterministic.
				  See Language Policy Options here https://developers.facebook.com/docs/whatsapp/api/messages/message-templates#language-policy-options.
				- Code, code (string). Required. The code of the language or locale to use. Accepts both language and language_locale formats
				  (e.g., en and en_US). For all codes, see Supported Languages. https://developers.facebook.com/docs/whatsapp/api/messages/message-templates#supported-languages

		   - Components, components (array of objects). Optional. Array of components objects containing the parameters of the message.

		   - Namespace, namespace (string). Optional. Only used for On-Premises API. Namespace of the template.*/
	Template struct {
		Name      string            `json:"name,omitempty"`
		Namespace string            `json:"namespace,omitempty"`
		Language  *TemplateLanguage `json:"language,omitempty"`
	}

	TemplateLanguage struct {
		Policy string `json:"policy,omitempty"`
		Code   string `json:"code,omitempty"`
	}

	// TemplateButton contains information about a template button.
	// 		- Type, type (string) Required. Indicates the type of parameter for the button.
	//		- Payload, payload (string) Required for quick_reply buttons. Developer-defined payload that is returned
	// 		  when the button is clicked in addition to the display text on the button.
	//		- Text, text (string) Required for URL buttons. Developer-provided suffix that is appended to the predefined
	//		  prefix URL in the template.
	TemplateButton struct {
		Type    string `json:"type,omitempty"`
		Payload string `json:"payload,omitempty"`
		Text    string `json:"text,omitempty"`
	}

	// TemplateCurrency contains information about a currency parameter.
	TemplateCurrency struct{}

	// TemplateDateTime contains information about a date_time parameter.
	TemplateDateTime struct{}

	/*
		TemplateParameter contains information about a template parameter. A template parameter object can contain
		the following parameters:

		- Type, type (string) Required. Describes the parameter type. Supported values:
				- currency
				- date_time
				- document
				- image
				- text
				- video
			For text-based templates, the only supported parameter types are currency, date_time, and text.

		- Text, text (string) Required when type=text. The message’s text. Character limit varies based on the following
		  included component type. For the header component type, 60 characters. For the body component type, 1024 characters
		  if other component types are included, 32768 characters if body is the only component type included.

		- Currency, currency (object) Required when type=currency. A currency object.

		- DateTime, date_time (object) Required when type=date_time. A date_time object.

		- Image, image (object) Required when type=image. A media object of type image. Captions not supported when used in a media template.

		- Document, document (object) Required when type=document. A media object of type document. Only PDF documents are supported for media-based
		  message templates. Captions not supported when used in a media template.

		- Video, video (object) Required when type=video. A media object of type video. Captions not supported when used in a media template.
	*/
	TemplateParameter struct {
		Type     string            `json:"type,omitempty"`
		Text     string            `json:"text,omitempty"`
		Currency *TemplateCurrency `json:"currency,omitempty"`
		DateTime *TemplateDateTime `json:"date_time,omitempty"`
		Image    *Media            `json:"image,omitempty"`
		Document *Media            `json:"document,omitempty"`
		Video    *Media            `json:"video,omitempty"`
	}
)
