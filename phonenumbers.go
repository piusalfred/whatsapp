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
	"io"
	"net/http"
	"net/url"

	whttp "github.com/piusalfred/whatsapp/http"
)

type (
	// VerificationCodeRequest is the request body for requesting a verification code.
	// doc link: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers
	// BaseURL is the base url for the request
	// ApiVersion is the version of the api
	// PhoneNumberID is the phone number id
	// CodeMethod is the method to use to send the code. It can be SMS or VOICE
	// Language is the language to use for the code. eg. en
	VerificationCodeRequest struct {
		Token         string `json:"token"`
		BaseURL       string `json:"base_url"`
		ApiVersion    string `json:"api_version"`
		PhoneNumberID string `json:"phone_number_id"`
		CodeMethod    string `json:"code_method"`
		Language      string `json:"language"` // eg. en
	}

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
	// AVAILABLE_WITHOUT_REVIEW: The certificate for the phone is available and display name is ready to use without review.
	// DECLINED: The name has not been approved. You cannot download your certificate.
	// EXPIRED: Your certificate has expired and can no longer be downloaded.
	// PENDING_REVIEW: Your name request is under review. You cannot download your certificate.
	// NONE: No certificate is available.
	PhoneNumberNameStatus string
)

// RequestCode sends a verification code to a phone number that will later be used for verification.
//
// doc link: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers
//
// You need to verify the phone number you want to use to send messages to your customers. Phone numbers
// must be verified using a code sent via an SMS/voice call. The verification process can be done via Graph API
// calls specified below.
//
// To verify a phone number using Graph API, make a POST request to PHONE_NUMBER_ID/request_code. In your call,
// include your chosen verification method and language.
//
//	curl -X POST \
//	 'https://graph.facebook.com/v16.0/FROM_PHONE_NUMBER_ID/request_code' \
//	 -H 'Authorization: Bearer ACCESS_TOKEN' \
//	 -F 'code_method=SMS' \
//	 -F 'language=en'
func RequestCode(ctx context.Context, client *http.Client, req *VerificationCodeRequest) error {
	reqCtx := &whttp.RequestContext{
		Name:       "request code",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"request_code"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  req.Token,
		Form:    map[string]string{"code_method": req.CodeMethod, "language": req.Language},
		Payload: nil,
	}
	err := whttp.Send(ctx, client, params, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// VerifyCode verifies a phone number using a verification code sent via SMS/voice call.
// After the API call, you will receive your verification code via the method you selected.
// To finish the verification process, include your code in a POST request to PHONE_NUMBER_ID/verify_code.
//
//	curl -X POST \
//	 'https://graph.facebook.com/v16.0/FROM_PHONE_NUMBER_ID/verify_code' \
//	 -H 'Authorization: Bearer ACCESS_TOKEN' \
//	 -F 'code=000000'
//
// A successful response looks like this:
//
//	{
//	 "success": true
//	}
func VerifyCode(ctx context.Context, client *http.Client, req *VerificationCodeRequest, code string) error {
	reqCtx := &whttp.RequestContext{
		Name:       "verify code",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.PhoneNumberID,
		Endpoints:  []string{"verify_code"},
	}
	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  req.Token,
		Form:    map[string]string{"code": code},
	}

	err := whttp.Send(ctx, client, params, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

type PhoneNumberFilterParams struct {
	Field    string `json:"field,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}
type ListPhoneNumbersRequest struct {
	BaseURL      string
	ApiVersion   string
	Token        string
	BusinessID   string
	FilterParams []*PhoneNumberFilterParams
}

// ListPhoneNumbers retrieve Phone Numbers that a business has registered for their WhatsApp
// Business Account using the WhatsApp Business Management API.
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
// curl -i -X GET "https://graph.facebook.com/v16.0/{whatsapp-business-account-ID}/phone_numbers?filtering=[{"field":"account_mode","operator":"EQUAL","value":"SANDBOX"}]&access_token=access-token"
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
func ListPhoneNumbers(ctx context.Context, client *http.Client, token string, req *ListPhoneNumbersRequest) (
	*PhoneNumbersList, error,
) {
	reqCtx := &whttp.RequestContext{
		Name:       "list phone numbers",
		BaseURL:    req.BaseURL,
		ApiVersion: req.ApiVersion,
		SenderID:   req.BusinessID,
		Endpoints:  []string{"phone_numbers"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Query:   map[string]string{"access_token": req.Token},
	}
	if req.FilterParams != nil {
		p := req.FilterParams
		jsonParams, err := json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter params: %w", err)
		}
		params.Query["filtering"] = string(jsonParams)
	}
	var phoneNumbersList PhoneNumbersList
	err := whttp.Send(ctx, client, params, &phoneNumbersList)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &phoneNumbersList, nil
}

// GetPhoneNumberByID returns a phone number by id.
func GetPhoneNumberByID(ctx context.Context, client *http.Client, token, id string) (*PhoneNumber, error) {
	requestURL, err := url.JoinPath("https://graph.facebook.com/v0.16", id)
	if err != nil {
		return nil, fmt.Errorf("failed to create request url: %w", err)
	}

	// create the request
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	// add the access token to the request
	q := request.URL.Query()
	q.Add("access_token", token)
	request.URL.RawQuery = q.Encode()

	// send the request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// check the response status code
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", response.Status)
	}

	// read the response body
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// unmarshal the response body
	var phoneNumber PhoneNumber
	if err := json.Unmarshal(body, &phoneNumber); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &phoneNumber, nil
}

type DisplayNameStatus struct {
	ID         string `json:"id,omitempty"`
	NameStatus string `json:"name_status,omitempty"`
}

// GetDisplayNameStatus
// //Include fields=name_status as a query string parameter to get the status of a display name associated with a specific phone number. This field is currently in beta and not available to all developers.
//
// Sample Request
// curl \
// 'https://graph.facebook.com/v15.0/105954558954427?fields=name_status' \
// -H 'Authorization: Bearer EAAFl...'
// Sample Response
//
//	{
//	  "id" : "105954558954427",
//	  "name_status" : "AVAILABLE_WITHOUT_REVIEW"
//	}
func GetDisplayNameStatus(ctx context.Context, client *http.Client, token string, id string) (*DisplayNameStatus, error) {
	requestURL, err := url.JoinPath("https://graph.facebook.com/v0.15", id)
	if err != nil {
		return nil, fmt.Errorf("failed to create request url: %w", err)
	}

	// create the request
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	// add the access token to the request
	q := request.URL.Query()
	q.Add("access_token", token)
	q.Add("fields", "name_status")
	request.URL.RawQuery = q.Encode()

	// send the request
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// check the response status code
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", response.Status)
	}

	// read the response body
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// unmarshal the response body
	var displayNameStatus DisplayNameStatus
	if err := json.Unmarshal(body, &displayNameStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &displayNameStatus, nil
}
