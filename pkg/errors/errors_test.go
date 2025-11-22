//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package errors_test

import (
	"testing"

	"github.com/piusalfred/whatsapp/pkg/errors"
)

func TestError_Error(t *testing.T) {
	type fields struct {
		Message   string
		Type      string
		Details   string
		Code      int
		Data      *errors.ErrorData
		Subcode   int
		UserTitle string
		UserMsg   string
		FBTraceID string
		Href      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test with all fields",
			fields: fields{
				Message: "(#130429) Rate limit hit",
				Type:    "OAuthException",
				Code:    130429,
				Data: &errors.ErrorData{
					MessagingProduct: "whatsapp",
					Details:          "too many messages sent from this phone number in a short period of time",
				},
				Subcode:   2494055,
				FBTraceID: "Az8or2yhqkZfEZ-_4Qn_Bam",
				Href:      "https://developers.facebook.com/docs/whatsapp/api/errors/",
			},
			want: "whatsapp error: message: (#130429) rate limit hit, type: oauthexception, statuscode: 130429, data: messaging product: whatsapp, details: too many messages sent from this phone number in a short period of time, subcode: 2494055, fbtraceid: az8or2yhqkzfez-_4qn_bam, href: https://developers.facebook.com/docs/whatsapp/api/errors/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &errors.Error{
				Message:   tt.fields.Message,
				Type:      tt.fields.Type,
				Details:   tt.fields.Details,
				Code:      tt.fields.Code,
				Data:      tt.fields.Data,
				Subcode:   tt.fields.Subcode,
				UserTitle: tt.fields.UserTitle,
				UserMsg:   tt.fields.UserMsg,
				FBTraceID: tt.fields.FBTraceID,
				Href:      tt.fields.Href,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
