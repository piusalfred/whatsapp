package whatsapp

import (
	"github.com/piusalfred/whatsapp/models"
	"reflect"
	"testing"
)

func TestBuildPayloadForMediaMessage(t *testing.T) {
	type args struct {
		options *SendMediaOptions
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "build audio message",
			args: args{
				options: &SendMediaOptions{
					Recipient: "2348123456789",
					Type:      "audio",
					Media: &models.Media{
						ID:       "1234567890",
						Link:     "https://example.com/audio.mp3",
						Caption:  "Audio caption",
						Filename: "audio.mp3",
						Provider: "whatsapp",
					},
					CacheOptions: nil,
				},
			},
			want: []byte(`{"product":"whatsapp","to":"2348123456789","recipient_type":"individual","type":"audio","audio":{"id":"1234567890","link":"https://example.com/audio.mp3","caption":"Audio caption","filename":"audio.mp3","provider":"whatsapp"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildPayloadForMediaMessage(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// print the payload
			t.Logf("\npayload: \n%s\n", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildPayloadForMediaMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
