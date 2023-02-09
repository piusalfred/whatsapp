package whatsapp

import (
	"context"
	"fmt"
	whttp "github.com/piusalfred/whatsapp/http"
	"net/http"
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
	request, err := whttp.NewRequestWithContext(ctx, &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Query: nil,
		Form: map[string]string{
			"code_method": req.CodeMethod,
			"language":    req.Language,
		},
		Bearer:   req.Token,
		BaseURL:  req.BaseURL,
		Endpoint: "/request_code",
		Method:   http.MethodPost,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}
	// send the request
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	// check the response status code
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", response.Status)
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
	request, err := whttp.NewRequestWithContext(ctx, &whttp.RequestParams{
		SenderID:   req.PhoneNumberID,
		ApiVersion: req.ApiVersion,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Query:  nil,
		Bearer: req.Token,
		Form: map[string]string{
			"code": code,
		},
		BaseURL:  req.BaseURL,
		Endpoint: "/verify_code",
		Method:   http.MethodPost,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}
	// send the request
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	// check the response status code
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", response.Status)
	}
	return nil
}
