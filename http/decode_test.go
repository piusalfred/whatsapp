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

package http

import (
	"bytes"
	"testing"

	"github.com/piusalfred/whatsapp/errors"
)

func TestErrorDecodeFunc(t *testing.T) {
	t.Parallel()
	type args struct{ data []byte }
	tests := []struct {
		name      string
		args      args
		wantsErr  *errors.Error
		shouldErr bool
	}{
		{
			name: "Test error decoding",
			args: args{
				data: []byte(`{
					"error": {
						"message": "(#131030) Recipient phone number not in allowed list",	
							"type": "OAuthException",
							"code": 131030,
							"error_data": {
								"messaging_product": "whatsapp",
								"details": "Recipient phone number not in allowed list: Add recipient phone number to recipient list and try again."
							},
							"error_subcode": 2655007,
							"error_user_title": "Recipient phone number not in allowed list",
							"error_user_msg": "Add recipient phone number to recipient list and try again.",
							"fbtrace_id": "AI5Ob2z72R0JAUB5zOF-nao"}`),
			},
			wantsErr: &errors.Error{
				Message: "(#131030) Recipient phone number not in allowed list",
				Type:    "OAuthException",
				Code:    131030,
				Data: &errors.ErrorData{
					MessagingProduct: "whatsapp",
					Details:          "Recipient phone number not in allowed list: Add recipient phone number to recipient list and try again.", //nolint:lll
				},
				Subcode:   2655007,
				UserTitle: "Recipient phone number not in allowed list",
				UserMsg:   "Add recipient phone number to recipient list and try again.",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var err errors.Error
			decodeErr := ErrorDecoder(bytes.NewReader(tt.args.data), &err)
			if decodeErr != nil {
				t.Logf("Error decoding error: %v", decodeErr)
			}
			if tt.shouldErr {
				if err != *tt.wantsErr {
					t.Errorf("Error decoding error, got: %v, wants: %v", err, tt.wantsErr)
				}
			}

			t.Logf("Error decoding error, got: %v, wants: %v", err, tt.wantsErr)
		})
	}
}
