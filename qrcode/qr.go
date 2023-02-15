// Package qrcode provides APi to manage QR codes.
// It contains API oto create, update, get a list of, and delete QR Code Messages using
// the WhatsApp Business Management API.
//
// Customers can scan a QR code from their phone to quickly begin a conversation with your
// business. The WhatsApp Business Management API allows you to create and access these QR
// codes and associated short links.
//
// If you can use the Business Manager to manage your QR codes instead of the API, see Manage
// your WhatsApp QR Codes here https://web.facebook.com/business/help/890732351439459?_rdc=1&_rdr
//
// Before You Start

// You will need:

// - The ID for the current phone number for your business (https://developers.facebook.com/docs/whatsapp/business-management-api/manage-phone-numbers#get-all-phone-numbers)
// - A User access token requested by someone who is an admin for the Business Manager
// - The whatsapp_business_messages permission
//
// Create a QR Code Message

// To create a QR code for a business, send a POST request to the WhatsApp Business Phone Number >
// Message Qrdls endpoint with the prefilled_message parameter set to your message text and
// generate_qr_image parameter set to your preferred image format, either SVG or PNG.

// Example Request

// Formatted for readability
// curl -X POST "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls
//   ?prefilled_message={message-text}
//   &generate_qr_image={image-format}
//   &access_token={user-access-token}"
// On success, a JSON array is returned:

// {
//     "code": "{qr-code-id}",
//     "prefilled_message": "{business message text}",
//     "deep_link_url": "{short-link-to-qr-code}",
//     "qr_image_url": "{image-url}"
// }
// Retrieve a List of QR Code Messages

// To get a list of all the QR codes messages for a business, send a GET request to the WhatsApp Business Phone Number > Message Qrdls endpoint.

// Example Request

// Formatted for readability
// curl -X GET "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls
//    &access_token={user-access-token}"
// On success, a JSON array is returned:

// {
//     "data": [
//         {
//             "code": "ANOVZ6RINRD7G1",
//             "prefilled_message": "I need help with my account.",
//             "deep_link_url": "https://wa.me/message/ANOVZ6RINRD7G1"
//         },
//         {
//             "code": "TNGSHG326AIHH1",
//             "prefilled_message": "What are your store hours?",
//             "deep_link_url": "https://wa.me/message/TNGSHG326AIHH1"
//         },
//         {
//             "code": "R3LUI5KILJUYA1",
//             "prefilled_message": "When is the next sale?",
//             "deep_link_url": "https://wa.me/message/R3LUI5KILJUYA1"
//         },

//     ]
// }
// Retrieve Single QR Code Message

// To get information about a specific QR code message, send a GET request to the WhatsApp Business Phone Number > Message Qrdls endpoint and append the QR code ID as a path parameter.

// Example Request

// Formatted for readability
// curl -X GET "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}
//    &access_token={user-access-token}"
// On success, a JSON array is returned:

// {
//     "data": [
//         {
//             "code": "ANOVZ6RINRD7G1",
//             "prefilled_message": "I need help with my account.",
//             "deep_link_url": "https://wa.me/message/ANOVZ6RINRD7G1"
//         }
//     ]
// }
// Update QR Code Message

// To update a QR code for a business, send a POST request to the WhatsApp Business Phone Number > Message Qrdls endpoint and append the QR code ID as a path parameter.

// Example Request

// Formatted for readability
// curl -X POST "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}
//   ?prefilled_message={new-message-text}
//   &access_token={user-access-token}"
// On success, a JSON array is returned:

// {
//     "code": "{qr-code-id}",
//     "prefilled_message": "{business message text}",
//     "deep_link_url": "{short-link-to-qr-code}"
// }
// Delete QR Code Message

// QR codes do not expire. You must delete a QR code in order to retire it.

// To delete a QR code, send a DELETE request to the WhatsApp Business Phone Number > Message Qrdls endpoint and append the ID of the QR code you wish to retire as a path parameter.

// Example Request

// Formatted for readability
// curl -X DELETE "https://graph.facebook.com/v16.0/{phone-number-ID}/message_qrdls/{qr-code-id}
//   &access_token={user-access-token}"
// On success, a JSON array is returned:

// {
//     "success": true
// }

package qrcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	ImageFormatPNG ImageFormat = "PNG"
	ImageFormatSVG ImageFormat = "SVG"
)

type (
	ImageFormat string

	CreateRequest struct {
		PrefilledMessage string      `json:"prefilled_message"`
		ImageFormat      ImageFormat `json:"generate_qr_image"`
	}

	CreateResponse struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
		QRImageURL       string `json:"qr_image_url"`
	}

	Information struct {
		Code             string `json:"code"`
		PrefilledMessage string `json:"prefilled_message"`
		DeepLinkURL      string `json:"deep_link_url"`
	}

	ListResponse struct {
		Data []*Information `json:"data,omitempty"`
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}
)

func Create(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken string, req *CreateRequest) (*CreateResponse, error) {
	var (
		resp     CreateResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("prefilled_message", req.PrefilledMessage)
	q.Add("generate_qr_image", string(req.ImageFormat))
	q.Add("access_token", accessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func List(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken string) (*ListResponse, error) {
	var (
		resp     ListResponse
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func Get(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken, qrCodeID string) (*Information, error) {
	var (
		list     ListResponse
		resp     Information
		respBody []byte
	)
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Body != nil {
		defer response.Body.Close()
		respBody, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(respBody, &list)
	if err != nil {
		return nil, err
	}

	if len(list.Data) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	resp = *list.Data[0]

	return &resp, nil
}

func Delete(ctx context.Context, client *http.Client, baseURL, phoneID, accessToken, qrCodeID string) error {
	requestURL, err := url.JoinPath(baseURL, phoneID, "message_qrdls", qrCodeID)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return err
	}

	q := request.URL.Query()
	q.Add("access_token", accessToken)
	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
