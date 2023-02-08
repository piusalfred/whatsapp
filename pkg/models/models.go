package models

type (
	Reaction struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}

	Text struct {
		PreviewUrl bool   `json:"preview_url,omitempty"`
		Body       string `json:"body,omitempty"`
	}

	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Name      string  `json:"name"`
		Address   string  `json:"address"`
	}

	Address struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Type        string `json:"type"`
	}

	Addresses struct {
		Addresses []Address `json:"addresses"`
	}

	Email struct {
		Email string `json:"email"`
		Type  string `json:"type"`
	}

	Emails struct {
		Emails []Email `json:"emails"`
	}

	Name struct {
		FormattedName string `json:"formatted_name"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		MiddleName    string `json:"middle_name"`
		Suffix        string `json:"suffix"`
		Prefix        string `json:"prefix"`
	}

	Org struct {
		Company    string `json:"company"`
		Department string `json:"department"`
		Title      string `json:"title"`
	}

	Phone struct {
		Phone string `json:"phone"`
		Type  string `json:"type"`
		WaID  string `json:"wa_id,omitempty"`
	}

	Phones struct {
		Phones []Phone `json:"phones"`
	}

	Url struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}

	Urls struct {
		Urls []Url `json:"urls"`
	}

	Contact struct {
		Addresses Addresses `json:"addresses,omitempty"`
		Birthday  string    `json:"birthday"`
		Emails    Emails    `json:"emails,omitempty"`
		Name      Name      `json:"name"`
		Org       Org       `json:"org"`
		Phones    Phones    `json:"phones,omitempty"`
		Urls      Urls      `json:"urls,omitempty"`
	}

	Contacts struct {
		Contacts []Contact `json:"contacts"`
	}

	// Context used to store the context of the conversation.
	// You can send any message as a reply to a previous message in a conversation by including
	// the previous message's ID in the context object.
	// The recipient will receive the new message along with a contextual bubble that displays
	// the previous message's content.
	// Recipients will not see a contextual bubble if:
	//    - replying with a template message ("type":"template")
	//    - replying with an image, video, PTT, or audio, and the recipient is on KaiOS
	// These are known bugs which we are addressing.
	Context struct {
		MessageID string `json:"message_id"`
	}
	MediaInfo struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
	}

	/*
	   Media represents a media object. This object is used to send media messages to WhatsApp users. It contains the following fields:

	     - ID, id (string). Required when type is audio, document, image, sticker, or video and you are not using a link.
	       The media object ID. Do not use this field when message type is set to text.

	     - Link, link (string). Required when type is audio, document, image, sticker, or video and you are not using an uploaded
	       media ID (i.e. you are hosting the media asset on your server). The protocol and URL of the media to be sent. Use only
	       with HTTP/HTTPS URLs. Do not use this field when message type is set to text.

	     - Cloud API users only:

	     - See Media HTTP Caching if you would like us to cache the media asset for future messages.

	     - When we request the media asset from your server you must indicate the media's MIME type by including the
	       Content-Type HTTP header. For example: Content-Type: video/mp4. See Supported Media Types for a list of supported
	       media and their MIME types.

	     - Caption, caption (string). For On-Premises API users on v2.41.2 or newer, this field is required when type is audio,
	       document, image, or video and is limited to 1024 characters. Optional. Describes the specified image, document, or
	       video media. Do not use with audio or sticker media.

	     - Filename, filename (string). Optional. Describes the filename for the specific document. Use only with document media.
	       The extension of the filename will specify what format the document is displayed as in WhatsApp.

	     - Provider, provider (string). Optional. Only used for On-Premises API. This path is optionally used with a link when the
	       HTTP/HTTPS link is not directly accessible and requires additional configurations like a bearer token. For information
	       on configuring providers, see the Media Providers documentation.
	*/
	Media struct {
		ID       string `json:"id,omitempty"`
		Link     string `json:"link,omitempty"`
		Caption  string `json:"caption,omitempty"`
		Filename string `json:"filename,omitempty"`
		Provider string `json:"provider,omitempty"`
	}

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
		Name       string               `json:"name,omitempty"`
		Namespace  string               `json:"namespace,omitempty"`
		Language   *TemplateLanguage    `json:"language,omitempty"`
		Components []*TemplateComponent `json:"components,omitempty"`
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
	// FallbackValue, fallback_value. Required. Default text if localization fails.
	// Code, code. Required. Currency code as defined in ISO 4217.
	//Amount1000,amount_1000. Required.Amount multiplied by 1000.
	TemplateCurrency struct {
		FallbackValue string `json:"fallback_value,omitempty"`
		Code          string `json:"code,omitempty"`
		Amount1000    int    `json:"amount_1000,omitempty"`
	}

	// TemplateDateTime contains information about a date_time parameter.
	// FallbackValue, fallback_value. Required. Default text if localization fails.
	// DayOfWeek, day_of_week. Required. Day of the week, where 0 is Sunday and 6 is Saturday.
	// Year, year. Required. Year.
	// Month, month. Required. Month, where 1 is January and 12 is December.
	// DayOfMonth, day_of_month. Required. Day of the month.
	// Hour, hour. Required. Hour of the day, where 0 is midnight and 23 is 11pm.
	// Minute, minute. Required. Minute of the hour.
	// Calendar, calendar. Required. Calendar type. Supported values:
	//	- GREGORIAN
	// Example:
	//       "date_time": {
	//               "fallback_value": "February 25, 1977",
	//                "day_of_week": 5,
	//                "year": 1977,
	//                "month": 2,
	//                "day_of_month": 25,
	//                "hour": 15,
	//                "minute": 33,
	//                "calendar": "GREGORIAN"
	//        }
	TemplateDateTime struct {
		FallbackValue string `json:"fallback_value,omitempty"`
		DayOfWeek     int    `json:"day_of_week,omitempty"`
		Year          int    `json:"year,omitempty"`
		Month         int    `json:"month,omitempty"`
		DayOfMonth    int    `json:"day_of_month,omitempty"`
		Hour          int    `json:"hour,omitempty"`
		Minute        int    `json:"minute,omitempty"`
		Calendar      string `json:"calendar,omitempty"`
	}

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

	// TemplateComponent contains information about a template component.
	// Type, type (string).Required. Describes the component type.
	//
	//Example of a components object with an array of parameters object nested inside:
	// "components": [{
	//   "type": "body",
	//   "parameters": [
	//      {
	//         "type": "text",
	//          "text": "name"
	//       },
	//       {
	//          "type": "text",
	//          "text": "Hi there"
	//        },
	//        {
	//           "type": "date_time",
	//           "date_time": {
	//                  "fallback_value": "February 25, 1977",
	//                  "day_of_week": 5,
	//                  "year": 1977,
	//                  "month": 2,
	//                  "day_of_month": 25,
	//                  "hour": 15,
	//                  "minute": 33,
	//                  "calendar": "GREGORIAN"
	//                 }
	//          }
	//     ]
	//   }]
	// SubType,sub_type (string). Required when type=button. Not used for the other types. Type of button to create.
	// Parameters,parameters array of objects. Required when type=button. Array of parameter objects with the content of the message.
	// For components of type=button, see the button parameter object.
	// Index, index. Required when type=button. Not used for the other types. Only used for Cloud API.
	// Position index of the button. You can have up to 3 buttons using index values of 0 to 2.
	TemplateComponent struct {
		Type       string               `json:"type,omitempty"`
		SubType    string               `json:"sub_type,omitempty"`
		Parameters []*TemplateParameter `json:"parameters,omitempty"`
		Index      int                  `json:"index,omitempty"`
	}
)
