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

/*
Package webhooks provides a simple way to create a webhooks server in Go.

Webhook allows you to receive real-time http notifications of changes to specific objects. For example,
we could send you a notification when a user sends you a message or when a message template's status has
changed. This prevents you from having to query our APIs for changes to objects that may or may not have
happened, and helps you avoid reaching your rate limit.

Before using this API, You must set up/subscribe to a webhooks to receive notifications from the WhatsApp
Business Platform.
Follow Whatsapp Webhooks Getting Started guide (https://developers.facebook.com/docs/graph-api/webhooks/getting-started)
to create your endpoint and configure your Webhooks. When you configure your Webhooks, make sure to choose
WhatsApp Business Account and subscribe to one or more WhatsApp business account fields.

Webhooks are triggered when a customer performs an action or the status for a message a business sends
a customer changes.

You get a webhooks notification, When a customer performs one of the following an action

  - Sends a text message to the business
  - Sends an image, video, audio, document, or sticker to the business
  - Sends contact information to the business
  - Sends location information to the business
  - Clicks a reply button set up by the business
  - Clicks a call-to-actions button on an Ad that Clicks to WhatsApp
  - Clicks an item on a business list
  - Updates their profile information such as their phone number
  - Asks for information about a specific product
  - Orders products being sold by the business

Notification Payload Object
NotificationPayloadObject is a combination of nested objects of JSON arrays and objects that contain information
about a change.

Structure of the notification payload object

			{
		  		"object": "whatsapp_business_account",
		  		"entry": [{
		    		"id": "WHATSAPP-BUSINESS-ACCOUNT-ID",
		    		"changes": [{
		      		"value": {
		         		"messaging_product": "whatsapp",
		         		"metadata": {
		           		"display_phone_number": "PHONE-NUMBER",
		           		"phone_number_id": "PHONE-NUMBER-ID"
		         		},
		      		# Additional arrays and objects
		         		"contacts": [{...}]
		         		"errors": [{...}]
		         		"messages": [{...}]
		         		"statuses": [{...}]
		      		},
		      		"field": "messages"
		    }]
		  }]
		}

		Example of a notification payload object after receiving a text message

	{
	  "object": "whatsapp_business_account",
	  "entry": [{
	      "id": "WHATSAPP_BUSINESS_ACCOUNT_ID",
	      "changes": [{
	          "value": {
	              "messaging_product": "whatsapp",
	              "metadata": {
	                  "display_phone_number": PHONE_NUMBER,
	                  "phone_number_id": PHONE_NUMBER_ID
	              },
	              "contacts": [{
	                  "profile": {
	                    "name": "NAME"
	                  },
	                  "wa_id": PHONE_NUMBER
	                }],
	              "messages": [{
	                  "from": PHONE_NUMBER,
	                  "id": "wamid.ID",
	                  "timestamp": TIMESTAMP,
	                  "text": {
	                    "body": "MESSAGE_BODY"
	                  },
	                  "type": "text"
	                }]
	          },
	          "field": "messages"
	        }]
	  }]
	}

# Message Status Updates

The WhatsApp Business Platform sends notifications to inform you of the status of the messages between you and users.
When a message is sent successfully, you receive a notification when the message is sent, delivered, and read.
The order of these notifications in your app may not reflect the actual timing of the message status. View the
timestamp to determine the timing, if necessary.
*/
package webhooks
