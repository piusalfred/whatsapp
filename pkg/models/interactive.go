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

const (
	InteractiveMessageReplyButton = "button"
	InteractiveMessageList        = "list"
	InteractiveMessageProduct     = "product"
	InteractiveMessageProductList = "product_list"
	InteractiveMessageCTAButton   = "cta_url"
)

type (
	// InteractiveMessage is the type of interactive message you want to send.
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
		Type  string                  `json:"type,omitempty"`
		Title string                  `json:"title,omitempty"`
		ID    string                  `json:"id,omitempty"`
		Reply *InteractiveReplyButton `json:"reply,omitempty"`
	}

	// InteractiveReplyButton contains information about a reply button in an interactive message.
	// A reply button object can contain the following parameters:
	// ID: Unique identifier for your button. This ID is returned in the webhook when the button
	// Title: Button title. It cannot be an empty string and must be unique within the message.
	InteractiveReplyButton struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	}

	// InteractiveSectionRow contains information about a row in an interactive section.
	InteractiveSectionRow struct {
		ID          string `json:"id,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	// InteractiveSection contains information about a section in an interactive message.
	// A section object can contain the following parameters:
	//	- ProductItems product_items, array of objects, Required for Multi-Product Messages. Array of product objects.
	//      There is a minimum of 1 product per section and a maximum of 30 products across all sections.
	//      Each product object contains the following field:
	//           - product_retailer_idstring – Required for Multi-Product Messages. Unique identifier of the
	//             product in a catalog. To get this ID, go to the Meta Commerce Manager, select your account
	//             and the shop you want to use. Then, click Catalog > Items, and find the item you want to mention.
	//             The ID for that item is displayed under the item's name.
	//
	//    - Rows array of objects, Required for ListQR Messages. Contains a list of rows. You can have a total of
	//      10 rows across all sections. Each row must have a title (Maximum length: 24 characters) and an ID
	//      (Maximum length: 200 characters). You can add a description (Maximum length: 72 characters), but it
	//      is optional.
	//
	//		  - Example:
	//
	//				"rows": [
	//					{
	//						"id":"unique-row-identifier-here",
	//						"title": "row-title-content-here",
	//						"description": "row-description-content-here",
	//					},
	//				]
	//
	//    - Title string. Required if the message has more than one section. Title of the section.
	//      Maximum length: 24 characters.
	InteractiveSection struct {
		Title        string                   `json:"title,omitempty"`
		ProductItems []*Product               `json:"product_items,omitempty"`
		Rows         []*InteractiveSectionRow `json:"rows,omitempty"`
	}

	// InteractiveAction contains information about an interactive action.
	// An interactive action object can contain the following parameters:
	//
	//	- Button, button (string) Required for ListQR Messages. Button content. It cannot be an empty
	//	  string and must be unique within the message. Emojis are supported, markdown is not.
	//	  Maximum length: 20 characters.
	//
	//	- Buttons, buttons (array of objects) Required for Reply Buttons. A button object can contain
	//	  the following parameters:
	//		- Type: only supported type is reply (for Reply Button)
	//		- Title: Button title. It cannot be an empty string and must be unique within the message.
	//		  Emojis are supported,markdown is not. Maximum length: 20 characters.
	//		- ID: Unique identifier for your button. This ID is returned in the webhook when the button
	//		  is clicked by the user. Maximum length: 256 characters.
	//    You can have up to 3 buttons. You cannot have leading or trailing spaces when setting the ID.
	//
	//	- CatalogID, catalog_id (string) Required for Single Product Messages and Multi-Product Messages.
	//	  Unique identifier of the Facebook catalog linked to your WhatsApp Business Account. This ID can
	//	  be retrieved via the Meta Commerce Manager.
	//
	//	- ProductRetailerID, product_retailer_id (string) Required for Single Product Messages and Multi-Product
	//	  Messages.Unique identifier of the product in a catalog. To get this ID go to Meta Commerce Manager and
	//	  select your Meta Business account. You will see a list of shops connected to your account. Click the shop
	//	  you want to use. On the left-side panel,click Catalog > Items, and find the item you want to mention.
	//	  The ID for that item is displayed under the item's name.
	//
	//	- Sections, sections (array of objects) Required for ListQR Messages and Multi-Product Messages. Array of
	//	  section objects. Minimum of 1, maximum of 10. See InteractiveSection object.
	InteractiveAction struct {
		Name              string                       `json:"name,omitempty"`
		Parameters        *InteractiveActionParameters `json:"parameters,omitempty"`
		Button            string                       `json:"button,omitempty"`
		Buttons           []*InteractiveButton         `json:"buttons,omitempty"`
		CatalogID         string                       `json:"catalog_id,omitempty"`
		ProductRetailerID string                       `json:"product_retailer_id,omitempty"`
		Sections          []*InteractiveSection        `json:"sections,omitempty"`
	}

	// InteractiveHeader contains information about an interactive header.
	// An interactive header object can contain the following parameters:
	//	- Document, document (object) Required if type is set to document. Contains the media object for this document.
	//
	//	- Image, image (object) Required if type is set to image. Contains the media object for this image.
	//
	//	- Text, text (string) Required if type is set to text. Text for the header. Formatting allows emojis,
	//      but not markdown. Maximum length: 60 characters.
	//
	//	- Type, type (string) Required. The header type you would like to use. Supported values:
	//		- text: Used for ListQR Messages, Reply Buttons, and Multi-Product Messages.
	//		- video: Used for Reply Buttons.
	//		- image: Used for Reply Buttons.
	//		- document: Used for Reply Buttons.
	//
	//	- Video, video (object) Required if type is set to video. Contains the media object for this video.
	InteractiveHeader struct {
		Document *Media `json:"document,omitempty"`
		Image    *Media `json:"image,omitempty"`
		Video    *Media `json:"video,omitempty"`
		Text     string `json:"text,omitempty"`
		Type     string `json:"type,omitempty"`
	}

	// InteractiveFooter contains information about an interactive footer.
	InteractiveFooter struct {
		Text string `json:"text,omitempty"`
	}

	// InteractiveBody contains information about an interactive body.
	InteractiveBody struct {
		Text string `json:"text,omitempty"`
	}

	// Interactive contains information about an interactive message. An interactive message object can contain the
	// following parameters:
	//
	//	- Action, action (object) Required. Action you want the user to perform after reading the message.
	//
	//	- Body, body (object) Optional for type product. Required for other message types. An object with the
	//      body of the message. The body object contains the following field:
	//	  		- Text, text (string) Required if body is present. The content of the message. Emojis and markdown
	//              are supported. Maximum length: 1024 characters.
	//
	//	- Footer, footer (object) Optional. An object with the footer of the message. The footer object contains
	//      the following field:
	//	  		- Text, text (string) Required if footer is present. The footer content. Emojis, markdown, and links
	//              are supported. Maximum length: 60 characters.
	//
	//	- Header, header (object) Required for type product_list. Optional for other types. Header content displayed
	//      on top of a message. You cannot set a header if your interactive object is of product type. See header object
	//      for more information.
	//
	//	- Type, type (string) Required. The type of interactive message you want to send. Supported values:
	//		- button: Used for ListQR Messages and Reply Buttons.
	//		- product: Used for Single Product Messages.
	//		- product_list: Used for Multi-Product Messages.
	Interactive struct {
		Type   string             `json:"type,omitempty"`
		Action *InteractiveAction `json:"action,omitempty"`
		Body   *InteractiveBody   `json:"body,omitempty"`
		Footer *InteractiveFooter `json:"footer,omitempty"`
		Header *InteractiveHeader `json:"header,omitempty"`
	}

	InteractiveActionParameters struct {
		URL         string `json:"url,omitempty"`
		DisplayText string `json:"display_text,omitempty"`
	}
)
