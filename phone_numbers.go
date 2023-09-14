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

package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/http"
)

////// PHONE NUMBERS

const (
	SMSVerificationMethod   VerificationMethod = "SMS"
	VoiceVerificationMethod VerificationMethod = "VOICE"
)

type (
	// VerificationMethod is the method to use to verify the phone number. It can be SMS or VOICE.
	VerificationMethod string

	PhoneNumber struct {
		VerifiedName       string `json:"verified_name"`
		DisplayPhoneNumber string `json:"display_phone_number"`
		ID                 string `json:"id"`
		QualityRating      string `json:"quality_rating"`
	}

	PhoneNumbersList struct {
		Data    []*PhoneNumber `json:"data,omitempty"`
		Paging  *Paging        `json:"paging,omitempty"`
		Summary *Summary       `json:"summary,omitempty"`
	}

	Paging struct {
		Cursors *Cursors `json:"cursors,omitempty"`
	}

	Cursors struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	}

	Summary struct {
		TotalCount int `json:"total_count,omitempty"`
	}

	// PhoneNumberNameStatus value can be one of the following:
	// APPROVED: The name has been approved. You can download your certificate now.
	// AVAILABLE_WITHOUT_REVIEW: The certificate for the phone is available and display name is ready to use
	// without review.
	// DECLINED: The name has not been approved. You cannot download your certificate.
	// EXPIRED: Your certificate has expired and can no longer be downloaded.
	// PENDING_REVIEW: Your name request is under review. You cannot download your certificate.
	// NONE: No certificate is available.
	PhoneNumberNameStatus string

	FilterParams struct {
		Field    string `json:"field,omitempty"`
		Operator string `json:"operator,omitempty"`
		Value    string `json:"value,omitempty"`
	}
)

// RequestVerificationCode requests a verification code to be sent via SMS or VOICE.
// doc link: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers
//
// You need to verify the phone number you want to use to send messages to your customers. After the
// API call, you will receive your verification code via the method you selected. To finish the verification
// process, include your code in the VerifyCode method.
func (client *Client) RequestVerificationCode(ctx context.Context,
	codeMethod VerificationMethod, language string,
) error {
	cctx := client.Context()
	reqCtx := &whttp.RequestContext{
		Name:          "request code",
		BaseURL:       cctx.BaseURL,
		ApiVersion:    cctx.ApiVersion,
		PhoneNumberID: cctx.PhoneNumberID,
		Endpoints:     []string{"request_code"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  cctx.AccessToken,
		Form:    map[string]string{"code_method": string(codeMethod), "language": language},
		Payload: nil,
	}
	err := client.http.Do(ctx, params, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// VerifyCode should be run to verify the code retrieved by RequestVerificationCode.
func (client *Client) VerifyCode(ctx context.Context, code string) (*StatusResponse, error) {
	cctx := client.Context()
	reqCtx := &whttp.RequestContext{
		Name:          "verify code",
		BaseURL:       cctx.BaseURL,
		ApiVersion:    cctx.ApiVersion,
		PhoneNumberID: cctx.PhoneNumberID,
		Endpoints:     []string{"verify_code"},
	}
	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  cctx.AccessToken,
		Form:    map[string]string{"code": code},
	}

	var resp StatusResponse
	err := client.http.Do(ctx, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &resp, nil
}

// ListPhoneNumbers returns a list of phone numbers that are associated with the business account.
// using the WhatsApp Business Management API.
//
// You will need to have
//   - The WhatsApp Business Account ID for the business' phone numbers you want to retrieve
//   - A System User access token linked to your WhatsApp Business Account
//   - The whatsapp_business_management permission
//
// Limitations
// This API can only retrieve phone numbers that have been registered. Adding, updating, or
// deleting phone numbers is not permitted using the API.
//
// The equivalent curl command to retrieve phone numbers is (formatted for readability):
//
//		curl -X GET "https://graph.facebook.com/v16.0/{whatsapp-business-account-id}/phone_numbers
//	      	?access_token={system-user-access-token}"
//
// On success, a JSON object is returned with a list of all the business names, phone numbers,
// phone number IDs, and quality ratings associated with a business.
//
//	{
//	  "data": [
//	    {
//	      "verified_name": "Jasper's Market",
//	      "display_phone_number": "+1 631-555-5555",
//	      "id": "1906385232743451",
//	      "quality_rating": "GREEN"
//
//		    },
//		    {
//		      "verified_name": "Jasper's Ice Cream",
//		      "display_phone_number": "+1 631-555-5556",
//		      "id": "1913623884432103",
//		      "quality_rating": "NA"
//		    }
//		  ],
//		}
//
// Filter Phone Numbers
// You can query phone numbers and filter them based on their account_mode. This filtering option
// is currently being tested in beta mode. Not all developers have access to it.
//
// Sample Request
//
//	curl -i -X GET "https://graph.facebook.com/v16.0/{whatsapp-business-account-ID}/phone_numbers?\
//		filtering=[{"field":"account_mode","operator":"EQUAL","value":"SANDBOX"}]&access_token=access-token"
//
// Sample Response
//
//	{
//	  "data": [
//	    {
//	      "id": "1972385232742141",
//	      "display_phone_number": "+1 631-555-1111",
//	      "verified_name": "John’s Cake Shop",
//	      "quality_rating": "UNKNOWN",
//	    }
//	  ],
//	  "paging": {
//		"cursors": {
//			"before": "abcdefghij",
//			"after": "klmnopqr"
//		}
//	   }
//	}
func (client *Client) ListPhoneNumbers(ctx context.Context, filters []*FilterParams) (*PhoneNumbersList, error) {
	cctx := client.Context()
	reqCtx := &whttp.RequestContext{
		Name:          "list phone numbers",
		BaseURL:       cctx.BaseURL,
		ApiVersion:    cctx.ApiVersion,
		PhoneNumberID: cctx.BusinessAccountID,
		Endpoints:     []string{"phone_numbers"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Query:   map[string]string{"access_token": cctx.AccessToken},
	}
	if filters != nil {
		p := filters
		jsonParams, err := json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter params: %w", err)
		}
		params.Query["filtering"] = string(jsonParams)
	}
	var phoneNumbersList PhoneNumbersList
	err := client.http.Do(ctx, params, &phoneNumbersList)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &phoneNumbersList, nil
}

// PhoneNumberByID returns the phone number associated with the given ID.
func (client *Client) PhoneNumberByID(ctx context.Context) (*PhoneNumber, error) {
	cctx := client.Context()
	reqCtx := &whttp.RequestContext{
		Name:          "get phone number by id",
		BaseURL:       cctx.BaseURL,
		ApiVersion:    cctx.ApiVersion,
		PhoneNumberID: cctx.PhoneNumberID,
	}
	request := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Headers: map[string]string{
			"Authorization": "Bearer " + cctx.AccessToken,
		},
	}
	var phoneNumber PhoneNumber
	if err := client.http.Do(ctx, request, &phoneNumber); err != nil {
		return nil, fmt.Errorf("get phone muber by id: %w", err)
	}

	return &phoneNumber, nil
}
