package models

const (
	InteractiveMessageButton      = "button"
	InteractiveMessageList        = "list"
	InteractiveMessageProduct     = "product"
	InteractiveMessageProductList = "product_list"
)

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
		Contacts []*Contact `json:"contacts"`
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

	// MediaInfo provides information about a media be it an Audio, Video, etc
	// Animated used with stickers only
	MediaInfo struct {
		ID       string `json:"id,omitempty"`
		Caption  string `json:"caption,omitempty"`
		MimeType string `json:"mime_type,omitempty"`
		Sha256   string `json:"sha256,omitempty"`
		Filename string `json:"filename,omitempty"`
		Animated bool   `json:"animated,omitempty"` // used with stickers true if animated
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

	// Product ...
	Product struct {
		RetailerID string `json:"product_retailer_id,omitempty"`
	}

	// InteractiveMessage is the type of interactive message you want to send. Supported values are:
	// 	- button: Use it for Reply Buttons.
	// 	- list: Use it for List Messages.
	// 	- product: Use for Single Product Messages.
	// 	- product_list: Use for Multi-Product Messages.
	InteractiveMessage string

	// InteractiveButton contains information about a button in an interactive message.
	// A button object can contain the following parameters:
	// 	- Type: only supported type is reply (for Reply Button)
	// 	- Title: Button title. It cannot be an empty string and must be unique within the message.
	// 	  Emojis are supported, markdown is not. Maximum length: 20 characters.
	// 	- ID: Unique identifier for your button. This ID is returned in the webhook when the button
	//	  is clicked by the user. Maximum length: 256 characters.
	//
	// You can have up to 3 buttons. You cannot have leading or trailing spaces when setting the ID.
	InteractiveButton struct {
		Type  string `json:"type,omitempty"`
		Title string `json:"title,omitempty"`
		ID    string `json:"id,omitempty"`
	}

	// InteractiveSectionRow contains information about a row in an interactive section.
	InteractiveSectionRow struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	/*
		InteractiveSection contains information about a section in an interactive message.
		A section object can contain the following parameters:
			- ProductItems product_items, array of objects, Required for Multi-Product Messages. Array of product objects.
		      There is a minimum of 1 product per section and a maximum of 30 products across all sections.
		      Each product object contains the following field:
		           - product_retailer_idstring – Required for Multi-Product Messages. Unique identifier of the product in a catalog.
		             To get this ID, go to the Meta Commerce Manager, select your account and the shop you want to use. Then, click
					 Catalog > Items, and find the item you want to mention. The ID for that item is displayed under the item's name.

		    - Rows array of objects, Required for List Messages. Contains a list of rows. You can have a total of 10 rows across
				  all sections. Each row must have a title (Maximum length: 24 characters) and an ID (Maximum length: 200 characters).
				  You can add a description (Maximum length: 72 characters), but it is optional.

				  - Example:

						"rows": [
							{
								"id":"unique-row-identifier-here",
								"title": "row-title-content-here",
								"description": "row-description-content-here",
							},
						]

		    - Title string. Required if the message has more than one section. Title of the section. Maximum length: 24 characters.
	*/
	InteractiveSection struct {
		Title        string                   `json:"title,omitempty"`
		ProductItems []*Product               `json:"product_items,omitempty"`
		Rows         []*InteractiveSectionRow `json:"rows,omitempty"`
	}

	/*
		InteractiveAction contains information about an interactive action.
		An interactive action object can contain the following parameters:

			- Button, button (string) Required for List Messages. Button content. It cannot be an empty string and must
			  be unique within the message. Emojis are supported, markdown is not.
			  Maximum length: 20 characters.

			- Buttons, buttons (array of objects) Required for Reply Buttons. A button object can contain the following parameters:
				- Type: only supported type is reply (for Reply Button)
				- Title: Button title. It cannot be an empty string and must be unique within the message. Emojis are supported,
				  markdown is not. Maximum length: 20 characters.
				- ID: Unique identifier for your button. This ID is returned in the webhook when the button is clicked by the user.
				  Maximum length: 256 characters.

				You can have up to 3 buttons. You cannot have leading or trailing spaces when setting the ID.
			- CatalogID, catalog_id (string) Required for Single Product Messages and Multi-Product Messages. Unique identifier of
			  the Facebook catalog linked to your WhatsApp Business Account. This ID can be retrieved via the Meta Commerce Manager.

			- ProductRetailerID, product_retailer_id (string) Required for Single Product Messages and Multi-Product Messages.
			  Unique identifier of the product in a catalog. To get this ID go to Meta Commerce Manager and select your Meta Business
			  account. You will see a list of shops connected to your account. Click the shop you want to use. On the left-side panel,
			  click Catalog > Items, and find the item you want to mention. The ID for that item is displayed under the item's name.

			- Sections, sections (array of objects) Required for List Messages and Multi-Product Messages. Array of section objects.
			  Minimum of 1, maximum of 10. See InteractiveSection object.
	*/
	InteractiveAction struct {
		Button            string                `json:"button,omitempty"`
		Buttons           []*InteractiveButton  `json:"buttons,omitempty"`
		CatalogID         string                `json:"catalog_id,omitempty"`
		ProductRetailerID string                `json:"product_retailer_id,omitempty"`
		Sections          []*InteractiveSection `json:"sections,omitempty"`
	}

	/*
		InteractiveHeader contains information about an interactive header.
		An interactive header object can contain the following parameters:
			- Document, document (object) Required if type is set to document. Contains the media object for this document.

			- Image, image (object) Required if type is set to image. Contains the media object for this image.

			- Text, text (string) Required if type is set to text. Text for the header. Formatting allows emojis, but not markdown.
			  Maximum length: 60 characters.

			- Type, type (string) Required. The header type you would like to use. Supported values:
				- text: Used for List Messages, Reply Buttons, and Multi-Product Messages.
				- video: Used for Reply Buttons.
				- image: Used for Reply Buttons.
				- document: Used for Reply Buttons.

			- Video, video (object) Required if type is set to video. Contains the media object for this video.
	*/
	InteractiveHeader struct {
		Text string `json:"text,omitempty"`
		Type string `json:"type,omitempty"`
	}

	// InteractiveFooter contains information about an interactive footer.
	InteractiveFooter struct {
		Text string `json:"text,omitempty"`
	}

	// InteractiveBody contains information about an interactive body.
	InteractiveBody struct {
		Text string `json:"text,omitempty"`
	}

	/*
		Interactive contains information about an interactive message. An interactive message object can contain the
		following parameters:

			- Action, action (object) Required. Action you want the user to perform after reading the message.

			- Body, body (object) Optional for type product. Required for other message types. An object with the body of the message.
			  The body object contains the following field:
			  		- Text, text (string) Required if body is present. The content of the message. Emojis and markdown are supported.
			  		  Maximum length: 1024 characters.

			- Footer, footer (object) Optional. An object with the footer of the message. The footer object contains the following field:
			  		- Text, text (string) Required if footer is present. The footer content. Emojis, markdown, and links are supported.
			  		  Maximum length: 60 characters.

			- Header, header (object) Required for type product_list. Optional for other types. Header content displayed on top of a message.
			  You cannot set a header if your interactive object is of product type. See header object for more information.

			- Type, type (string) Required. The type of interactive message you want to send. Supported values:
				- button: Used for List Messages and Reply Buttons.
				- product: Used for Single Product Messages.
				- product_list: Used for Multi-Product Messages.
	*/
	Interactive struct {
		Type   string             `json:"type,omitempty"`
		Action *InteractiveAction `json:"action,omitempty"`
		Body   *InteractiveBody   `json:"body,omitempty"`
		Footer *InteractiveFooter `json:"footer,omitempty"`
		Header *InteractiveHeader `json:"header,omitempty"`
	}

	/*
		Message is a WhatsApp message. It contins the following fields:

		Audio (object) Required when type=audio. A media object containing audio.

		Contacts (object) Required when type=contacts. A contacts object.

		Context (object) Required if replying to any message in the conversation. Only used for Cloud API.
		An object containing the ID of a previous message you are replying to.
		For example: {"message_id":"MESSAGE_ID"}

		Document (object). Required when type=document. A media object containing a document.

		Hsm (object). Only used for On-Premises API. Contains an hsm object. This option was deprecated with v2.39
		of the On-Premises API. Use the template object instead. Cloud API users should not use this field.

		Image (object). Required when type=image. A media object containing an image.

		Interactive (object). Required when type=interactive. An interactive object. The components of each interactive
		object generally follow a consistent pattern: header, body, footer, and action.

		Location (object). Required when type=location. A location object.

		MessagingProduct messaging_product (string)	Required. Only used for Cloud API. Messaging service used
		for the request. Use "whatsapp". On-Premises API users should not use this field.

		PreviewURL preview_url (boolean)	Required if type=text. Only used for On-Premises API. Allows for URL
		previews in text messages — See the Sending URLs in Text Messages.
		This field is optional if not including a URL in your message. Values: false (default), true.
		Cloud API users can use the same functionality with the preview_url field inside the text object.

		RecipientType recipient_type (string) Optional. Currently, you can only send messages to individuals.
		Set this as individual. Default: individual

		Status, status (string) A message's status. You can use this field to mark a message as read.
		See the following guides for information:
			- Cloud API: Mark Messages as Read
			- On-Premises API: Mark Messages as Read

		Sticker, sticker (object). Required when type=sticker. A media object containing a sticker.
			- Cloud API: Static and animated third-party outbound stickers are supported in addition to all types of inbound stickers.
			  A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
		      An animated sticker must be 512x512 pixels and cannot exceed 500 KB.
			- On-Premises API: Only static third-party outbound stickers are supported in addition to all types of inbound stickers.
			  A static sticker needs to be 512x512 pixels and cannot exceed 100 KB.
		      Animated stickers are not supported.
		      For Cloud API users, we support static third-party outbound stickers and all types of inbound stickers. The sticker needs
		      to be 512x512 pixels and the file size needs to be less than 100 KB.

		Template A template (object). Required when type=template. A template object.

		Text text (object). Required for text messages. A text object.

		To string. Required. WhatsApp ID or phone number for the person you want to send a message to.
		See Phone Numbers, Formatting for more information. If needed, On-Premises API users can get this number by
		calling the contacts' endpoint.

		Type type (string). Optional. The type of message you want to send. Default: text
	*/
	Message struct {
		Product       string       `json:"messaging_product"`
		To            string       `json:"to"`
		RecipientType string       `json:"recipient_type"`
		Type          string       `json:"type"`
		PreviewURL    bool         `json:"preview_url,omitempty"`
		Context       *Context     `json:"context,omitempty"`
		Template      *Template    `json:"template,omitempty"`
		Text          *Text        `json:"text,omitempty"`
		Image         *Media       `json:"image,omitempty"`
		Audio         *Media       `json:"audio,omitempty"`
		Video         *Media       `json:"video,omitempty"`
		Document      *Media       `json:"document,omitempty"`
		Sticker       *Media       `json:"sticker,omitempty"`
		Reaction      *Reaction    `json:"reaction,omitempty"`
		Location      *Location    `json:"location,omitempty"`
		Contacts      *Contacts    `json:"contacts,omitempty"`
		Interactive   *Interactive `json:"interactive,omitempty"`
	}
)
