/*
Package webhooks provides a simple way to create a webhooks server in Go.

	Before using this API, You must set up/subscribe to a webhooks to receive notifications from the WhatsApp Business Platform.
	Follow Whatsapp Webhooks Getting Started guide (https://developers.facebook.com/docs/graph-api/webhooks/getting-started)
	to create your endpoint and configure your Webhooks. When you configure your Webhooks, make sure to choose WhatsApp Business
	Account and subscribe to one or more WhatsApp business account fields.



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
	NotificationPayloadObject is a combination of nested objects of JSON arrays and objects that contain information about a change.

	Example Text Message

	Message Status Updates
	he WhatsApp Business Platform sends notifications to inform you of the status of the messages between you and users.<\br>
	When a message is sent successfully, you receive a notification when the
	message is sent, delivered, and read. The order of these notifications in your app may not reflect
	the actual timing of the message status. View the timestamp to determine the timing, if necessary.
*/
package webhooks
