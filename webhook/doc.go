/*
Package webhook provides a simple way to create a webhook server in Go.

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
*/
package webhook
