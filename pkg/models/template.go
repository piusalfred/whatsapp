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

package models

type (
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
		DayOfWeek     string `json:"day_of_week"`
		Year          int    `json:"year"`
		Month         int    `json:"month"`
		DayOfMonth    int    `json:"day_of_month"`
		Hour          int    `json:"hour"`
		Minute        int    `json:"minute"`
		Calendar      string `json:"calendar,omitempty"`
	}

	// TemplateParameter contains information about a template parameter. A template parameter object can contain
	// the following parameters:
	//
	//- Type, type (string) Required. Describes the parameter type. Supported values:
	//		- currency
	//		- date_time
	//		- document
	//		- image
	//		- text
	//		- video
	//	For text-based templates, the only supported parameter types are currency, date_time, and text.
	//
	//- Text, text (string) Required when type=text. The message’s text. Character limit varies based on the following
	//  included component type. For the header component type, 60 characters. For the body component type, 1024 characters
	//  if other component types are included, 32768 characters if body is the only component type included.
	//
	//- Currency, currency (object) Required when type=currency. A currency object.
	//
	//- DateTime, date_time (object) Required when type=date_time. A date_time object.
	//
	//- Image, image (object) Required when type=image. A media object of type image. Captions not supported when used in
	//  a media template.
	//
	//- Document, document (object) Required when type=document. A media object of type document. Only PDF documents are
	//  supported for media-based message templates. Captions not supported when used in a media template.
	//
	//- Video, video (object) Required when type=video. A media object of type video. Captions not supported when used in
	//  a media template.
	TemplateParameter struct {
		Type     string            `json:"type,omitempty"`
		Text     string            `json:"text,omitempty"`
		Payload  string            `json:"payload,omitempty"`
		Currency *TemplateCurrency `json:"currency,omitempty"`
		DateTime *TemplateDateTime `json:"date_time,omitempty"`
		Image    *Media            `json:"image,omitempty"`
		Document *Media            `json:"document,omitempty"`
		Video    *Media            `json:"video,omitempty"`
	}

	// TemplateComponent contains information about a template component.
	// Type, type (string).Required. Describes the component type.
	//
	// Example of a components object with an array of parameters objects nested inside:
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
	// Parameters,parameters array of objects. Required when type=button. Array of parameter objects with
	// the content of the message.
	// For components of type=button, see the button parameter object.
	// Index, index. Required when type=button. Not used for the other types. Only used for Cloud API.
	// Position index of the button. You can have up to 3 buttons using index values of 0 to 2.
	TemplateComponent struct {
		Type       string               `json:"type,omitempty"`
		SubType    string               `json:"sub_type,omitempty"`
		Parameters []*TemplateParameter `json:"parameters,omitempty"`
		Index      int                  `json:"index"`
	}

	// Template is a template for a message. It contains the parameters of the message as listed below.
	//
	//  - Name, name (string). Required. Name of the template.
	//
	//  - Language, language (object). Required. Contains a language object. Specifies the language the
	//    template may be rendered in. The language object can contain the following fields:
	//
	//	- Policy, policy (string). Required. The language policy the message should follow. The only supported
	//	  option is deterministic. See Language Policy Options here
	//	  https://developers.facebook.com/docs/whatsapp/api/messages/message-templates#language-policy-options.
	//
	//	 - StatusCode, code (string). Required. The code of the language or locale to use. Accepts both language
	//	   and language_locale formats (e.g., en and en_US). For all codes, see Supported Languages.
	//	   https://developers.facebook.com/docs/whatsapp/api/messages/message-templates#supported-languages
	//
	//   - Components, components (array of objects). Optional. Array of components objects containing the parameters
	//     of the message.
	//
	//   - Namespace, namespace (string). Optional. Only used for On-Premises API. Namespace of the template.
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
	// 	- Type, type (string) Required. Indicates the type of parameter for the button.
	//	- Payload, payload (string) Required for quick_reply buttons. Developer-defined payload that is returned
	// 	  when the button is clicked in addition to the display text on the button.
	//	- Text, text (string) Required for URL buttons. Developer-provided suffix that is appended to the predefined
	//	  prefix URL in the template.
	TemplateButton struct {
		Type    string `json:"type,omitempty"`
		Payload string `json:"payload,omitempty"`
		Text    string `json:"text,omitempty"`
	}

	// TemplateCurrency contains information about a currency parameter.
	// FallbackValue, fallback_value. Required. Default text if localization fails.
	// Code, code. Required. Currency code as defined in ISO 4217.
	// Amount1000,amount_1000. Required.Amount multiplied by 1000.
	TemplateCurrency struct {
		FallbackValue string `json:"fallback_value,omitempty"`
		Code          string `json:"code,omitempty"`
		Amount1000    int    `json:"amount_1000"`
	}

	// TemplateComponentType is a type of component of a template message.
	// It can be a header, body.

	InteractiveButtonTemplate struct {
		SubType string
		Index   int
		Button  *TemplateButton
	}
)
