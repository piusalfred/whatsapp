package whatsapp

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResponseMarshalJson(t *testing.T) {
	jsonstr := `{"messaging_product": "whatsapp","contacts": [{"input": "PHONE_NUMBER","wa_id": "WHATSAPP_ID"}],"messages": [{"id": "wamid.ID"}]}`
	contacts := []ResponseContact{
		{
			Input:      "123456789",
			WhatsappId: "123456789RTQRWDAR",
		},
	}

	messages := []MessageID{
		{
			ID: "wa:123456789RTQRWDAR:123456789",
		},
	}
	response := &Response{
		MessagingProduct: "whatsapp",
		Contacts:         contacts,
		Messages:         messages,
	}

	j, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(j))

	resulr := &Response{}
	err = json.Unmarshal([]byte(jsonstr), resulr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resulr)
}

func TestBuildPayloadForMediaMessage(t *testing.T) {
	type args struct {
		options *SendMediaOptions
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestBuildPayloadForMediaMessage",
			args: args{
				options: &SendMediaOptions{
					Recipient:  "123456789",
					Type:       MediaTypeImage,
					SendByLink: true,
					Link:       "https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png",
				},
			},
			want: `{"messaging_product": "whatsapp","recipient_type":"individual","to":"123456789","type":"image","image":{"link":"https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildPayloadForMediaMessage(tt.args.options)

			// print the json
			t.Log(string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf("BuildPayloadForMediaMessage() = \n\n%v, want \n%v", got, []byte(tt.want))
			}
		})
	}
}
