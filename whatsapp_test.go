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
	"fmt"
	"net/http"
	"testing"
)

func ExampleNewClient() {
	client := NewClient(
		WithHTTPClient(http.DefaultClient),
		WithBaseURL(BaseURL),
		WithVersion(LowestSupportedVersion),
		WithAccessToken("access_token"),
		WithPhoneNumberID("phone_number_id"),
		WithBusinessAccountID("whatsapp_business_account_id"),
		WithHooks(nil),
	)

	client.SetAccessToken("myexampletoken")
	client.SetPhoneNumberID("myexamplephoneid")
	client.SetBusinessAccountID("businessaccountID")

	cctx := client.Context()

	fmt.Printf("base url: %s\napi version: %s\ntoken: %s\nphone id: %s\nbusiness id: %s\n",
		cctx.BaseURL, cctx.ApiVersion, cctx.AccessToken, cctx.PhoneNumberID, cctx.BusinessAccountID)

	// Output:
	// base url: https://graph.facebook.com/
	// api version: v16.0
	// token: myexampletoken
	// phone id: myexamplephoneid
	// business id: businessaccountID
}

func ExampleClient_RequestVerificationCode() {
	client := NewClient(
		WithHTTPClient(http.DefaultClient),
		WithBaseURL(BaseURL),
		WithVersion(LowestSupportedVersion),
		WithAccessToken("access_token"),
		WithPhoneNumberID("phone_number_id"),
		WithBusinessAccountID("whatsapp_business_account_id"),
	)

	err := client.RequestVerificationCode(context.TODO(), SMSVerificationMethod, "en_US")
	if err != nil {
		fmt.Println(err)
	}
	err = client.RequestVerificationCode(context.TODO(), VoiceVerificationMethod, "en_US")
	if err != nil {
		fmt.Println(err)
	}

	// Output:
}

func TestMediaMaxAllowedSize(t *testing.T) {
	t.Parallel()
	type args struct {
		mediaType MediaType
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "video",
			args: args{
				MediaTypeVideo,
			},
			want: MaxVideoSize,
		},

		{
			name: "unknown",
			args: args{
				MediaType("unknown"),
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			args := tt.args
			if got := MediaMaxAllowedSize(args.mediaType); got != tt.want {
				t.Errorf("MediaMaxAllowedSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
