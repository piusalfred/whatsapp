package whatsapp

import (
	"encoding/json"
	"testing"

	"github.com/piusalfred/whatsapp/models"
)

func TestBuildPayloadForMediaMessage(t *testing.T) {
	type args struct {
		options *SendMediaRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "build audio message",
			args: args{
				options: &SendMediaRequest{
					Recipient: "2348123456789",
					Type:      "audio",
					MediaID:   "1234567890",
					MediaLink: "https://example.com/audio.mp3",
					Caption:   "Audio caption",
					Filename:  "audio.mp3",
					Provider:  "whatsapp",

					CacheOptions: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildPayloadForMediaMessage(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var message models.Message
			err = json.Unmarshal(got, &message)
			if err != nil {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if message.Product != "whatsapp" {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if message.To != tt.args.options.Recipient {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if message.RecipientType != "individual" {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if message.Type != string(tt.args.options.Type) {
				t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// check the media values
			if tt.args.options.Type == "audio" {
				if message.Audio == nil {
					t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
					return
				} else {
					if message.Audio.ID != tt.args.options.MediaID {
						t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if message.Audio.Link != tt.args.options.MediaLink {
						t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if message.Audio.Caption != tt.args.options.Caption {
						t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if message.Audio.Filename != tt.args.options.Filename {
						t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if message.Audio.Provider != tt.args.options.Provider {
						t.Errorf("BuildPayloadForMediaMessage() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
				}
			}
		})
	}
}

func BenchmarkBuildPayloadForMediaMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildPayloadForMediaMessage(&SendMediaRequest{
			Recipient: "2348123456789",
			Type:      "audio",
			MediaID:   "1234567890",
			MediaLink: "https://example.com/audio.mp3",
			Caption:   "Audio caption",
			Filename:  "audio.mp3",
			Provider:  "whatsapp",

			CacheOptions: nil,
		})
	}
}
