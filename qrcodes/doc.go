/*
Package qrcodes provides API implementations to manage QR codes for WhatsApp Business. It contains
functions to create, retrieve, update and delete QR codes using the WhatsApp Business Management API.

Customers can scan a QR code from their phone to quickly begin a conversation with your
business. The WhatsApp Business Management API allows you to create and access these QR
codes and associated short links.

If you can use the Business Manager to manage your QR codes instead of the API, see Manage
your WhatsApp QR Codes here https://web.facebook.com/business/help/890732351439459?_rdc=1&_rdr

Before using this API in your application. You must have the following:
  - The ID for the current phone number for your business (https://developers.facebook.com/docs/whatsapp/business-management-api/manage-phone-numbers#get-all-phone-numbers)
  - A User access token requested by someone who is an admin for the Business Manager
  - The whatsapp_business_messages permission

# Create a QR Code Message

Example:

	resp, err := qrcodes.Create(context.Background(), http.DefaultClient,&qrcodes.CreateRequest{
		PrefilledMessage: "Hello World",
		ImageFormat:      qrcodes.ImageFormatPNG,
	})
	// handle error
	......
	// handle response
	.......

This is equivalent to the following API call using curl which has been reformatted for readability:

	curl -X POST "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls
	?prefilled_message="Hello World"
	&generate_qr_image=PNG
	&access_token={user-access-token}"

On success, a JSON array is returned:

	{
		"code": "{qr-code-id}",
		"prefilled_message": "{business message text}",
		"deep_link_url": "{short-link-to-qr-code}",
		"qr_image_url": "{image-url}"
	}

# Retrieve All QR Code Messages

Example:

	resp, err := qrcodes.List(context.Background(), http.DefaultClient, "phoneNumberID","token")
	// handle error
	......
	// handle response
	......

This is an equivalent to the following curl command:

		curl -X GET "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls
		    &access_token={user-access-token}"
	 On success, a JSON array is returned:

		{
			"data": [
				{
					"code": "ANOVZ6RINRD7G1",
					"prefilled_message": "I need help with my account.",
					"deep_link_url": "https://wa.me/message/ANOVZ6RINRD7G1"
				},
				{
					"code": "TNGSHG326AIHH1",
	             "prefilled_message": "What are your store hours?",
	             "deep_link_url": "https://wa.me/message/TNGSHG326AIHH1"
	         },
	         {
	             "code": "R3LUI5KILJUYA1",
	             "prefilled_message": "When is the next sale?",
	             "deep_link_url": "https://wa.me/message/R3LUI5KILJUYA1"
	         },

	     ]
		}

# Retrive a single QR Code

# Example Request

Formatted for readability
curl -X GET "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}

	    &access_token={user-access-token}"
	 On success, a JSON array is returned:

	 {
	     "data": [
	         {
	             "code": "ANOVZ6RINRD7G1",
	             "prefilled_message": "I need help with my account.",
	             "deep_link_url": "https://wa.me/message/ANOVZ6RINRD7G1"
	         }
	     ]
	}

Example:

	info, err := qrcodes.Get(context.TODO(),http.DefaultClient, "baseURL", "phoneID", "accessToken", "qrCodeID")

# Update Details of a QR Code

# Example Request

Formatted for readability

	 curl -X POST "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}
	   ?prefilled_message={new-message-text}
	   &access_token={user-access-token}"
	 On success, a JSON array is returned:

	 {
	     "code": "{qr-code-id}",
	     "prefilled_message": "{business message text}",
	     "deep_link_url": "{short-link-to-qr-code}"
	}

Delete QR Code
QR codes do not expire. You must delete a QR code in order to retire it.

Formatted (for readability) curl command to delete a QR Code.
curl -X DELETE "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}

	&access_token={user-access-token}"

On success, a JSON array is returned:

	{
		"success": true
	}
*/
package qrcodes
