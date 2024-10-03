package message_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"
	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/message"
	mockhttp "github.com/piusalfred/whatsapp/mocks/http"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"go.uber.org/mock/gomock"
)

func TestBaseSender_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
		APIVersion:        whatsapp.LowestSupportedApiVersion,
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
			mock: whttp.SenderFunc[message.Message](func(ctx context.Context, request *whttp.Request[message.Message], decoder whttp.ResponseDecoder) error {
				response := &http.Response{
					Status:     "OK",
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"messaging_product": "whatsapp", "contacts": [{"input": "1234567", "wa_id": "16505551234"}], "messages": [{"id": "wamid.HBgLMTY0NjcwNDM1OTUVAgARGBI1RjQyNUE3NEYxMzAzMzQ5MkEA"}], "success": true}`)),
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
			mock: whttp.SenderFunc[message.Message](func(ctx context.Context, request *whttp.Request[message.Message], decoder whttp.ResponseDecoder) error {
				response := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"message": "internal server error", "type": "OAuthException", "code": 500}}`)),
				}

				return decoder.Decode(ctx, response)
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			response, err := sender.Send(context.TODO(), conf, request)
			if err != nil {
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("BaseSender.Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !gcmp.Equal(response, tt.want) {
				t.Errorf("BaseSender.Send() response = %v, want %v", response, tt.want)
				return
			}
		})
	}
}
