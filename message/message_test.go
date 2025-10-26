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

package message_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	mockhttp "github.com/piusalfred/whatsapp/mocks/http"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestBaseSender_Send(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	success := &message.Response{
		Product: whatsapp.MessageProduct,
		Contacts: []*message.ResponseContact{
			{Input: "1234567", WhatsappID: "16505551234"},
		},
		Messages: []*message.ID{
			{ID: "wamid.HBgLMTY0NjcwNDM1OTUVAgARGBI1RjQyNUE3NEYxMzAzMzQ5MkEA"},
		},
		Success: true,
	}

	conf := &config.Config{
		BaseURL:           whatsapp.BaseURL,
		APIVersion:        whatsapp.LowestSupportedAPIVersion,
		AccessToken:       "FAtes263t7sdvjhssw73w8y7w",
		PhoneNumberID:     "111111111",
		BusinessAccountID: "222222222",
	}

	tests := []struct {
		name           string
		recipient      string
		options        []message.Option
		requestOptions []message.BaseRequestOption
		mock           whttp.SenderFunc[message.Message]
		want           *message.Response
		wantErr        bool
	}{
		{
			name:      "send text message success",
			recipient: "1234567",
			options: []message.Option{
				message.WithTextMessage(&message.Text{
					PreviewURL: false,
					Body:       "Hello there",
				}),
			},
			requestOptions: []message.BaseRequestOption{
				message.WithBaseRequestEndpoints(message.Endpoint),
				message.WithBaseRequestMethod(http.MethodPost),
				message.WithBaseRequestType(whttp.RequestTypeSendMessage),
				message.WithBaseRequestDecodeOptions(whttp.DecodeOptions{
					DisallowUnknownFields: true,
					DisallowEmptyResponse: true,
					InspectResponseError:  true,
				}),
			},
			mock: whttp.SenderFunc[message.Message](func(ctx context.Context,
				_ *whttp.Request[message.Message],
				decoder whttp.ResponseDecoder,
			) error {
				response := &http.Response{
					Status:     "OK",
					StatusCode: http.StatusOK,
					Body: io.NopCloser(
						bytes.NewBufferString(`
{"messaging_product": "whatsapp", "contacts": [{"input": "1234567", "wa_id": "16505551234"}], "messages": [{"id": "wamid.HBgLMTY0NjcwNDM1OTUVAgARGBI1RjQyNUE3NEYxMzAzMzQ5MkEA"}], "success": true}`)),
				}

				return decoder.Decode(ctx, response)
			}),
			want:    success,
			wantErr: false,
		},
		{
			name:      "send text message failure with error",
			recipient: "1234567",
			options: []message.Option{
				message.WithTextMessage(&message.Text{
					PreviewURL: false,
					Body:       "Hello there",
				}),
			},
			requestOptions: []message.BaseRequestOption{
				message.WithBaseRequestEndpoints(message.Endpoint),
				message.WithBaseRequestMethod(http.MethodPost),
				message.WithBaseRequestType(whttp.RequestTypeSendMessage),
				message.WithBaseRequestDecodeOptions(whttp.DecodeOptions{
					DisallowUnknownFields: true,
					DisallowEmptyResponse: true,
					InspectResponseError:  true,
				}),
			},
			mock: whttp.SenderFunc[message.Message](func(ctx context.Context,
				_ *whttp.Request[message.Message],
				decoder whttp.ResponseDecoder,
			) error {
				response := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body: io.NopCloser(bytes.NewBufferString(`
{"error": {"message": "internal server error", "type": "OAuthException", "code": 500}}`)),
				}

				return decoder.Decode(ctx, response)
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := message.New(tt.recipient, tt.options...)
			if err != nil {
				t.Fatalf("create message: %v\n", err)
			}

			var (
				request    = message.NewBaseRequest(m, tt.requestOptions...)
				mockSender = mockhttp.NewMockSender[message.Message](ctrl)
			)

			mockSender.EXPECT().Send(
				gomock.Any(), gomock.Any(), gomock.Any(),
			).DoAndReturn(tt.mock).Times(1)

			sender := &message.BaseSender{Sender: mockSender}
			response, err := sender.SendRequest(t.Context(), conf, request)
			if err != nil {
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("BaseSender.SendRequest() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !gcmp.Equal(response, tt.want) {
				t.Errorf("BaseSender.SendRequest() response = %v, want %v", response, tt.want)

				return
			}
		})
	}
}

func TestNew(t *testing.T) {
	type testCase struct {
		name         string
		recipient    string
		options      []message.Option
		expectedJSON string
		wantErr      bool
	}
	tests := []testCase{
		{
			name:      "create text message for individual",
			recipient: "1234567",
			options: []message.Option{
				message.WithTextMessage(&message.Text{
					PreviewURL: false,
					Body:       "Hello there",
				}),
			},
			expectedJSON: `{"messaging_product": "whatsapp","recipient_type": "individual","to": "1234567","type": "text","text": {"preview_url": false,"body": "Hello there"}}`,
		},
		{
			name:      "create text message for group",
			recipient: "1234567",
			options: []message.Option{
				message.WithTextMessage(&message.Text{
					PreviewURL: false,
					Body:       "Hello there",
				}),
				message.WithRecipientType(message.RecipientTypeGroup),
			},
			expectedJSON: `{"messaging_product": "whatsapp","recipient_type": "group","to": "1234567","type": "text","text": {"preview_url": false,"body": "Hello there"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := message.New(tt.recipient, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				expected := &message.Message{}
				unmarshalErr := json.Unmarshal([]byte(tt.expectedJSON), expected)
				if unmarshalErr != nil {
					return
				}

				gcmp.Equal(got, expected)
			}
		})
	}
}
