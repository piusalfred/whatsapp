package webhooks

import "testing"

func TestParseMessageType(t *testing.T) {
	t.Parallel()
	type args struct {
		messageType string
	}

	tests := []struct {
		name    string
		args    args
		want    MessageType
		wantErr bool
	}{
		{
			name: "tExt",
			args: args{
				messageType: "text",
			},
			want:    TextMessageType,
			wantErr: false,
		},
		{
			name: "imageX",
			args: args{
				messageType: "imageX",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseMessageType(tt.args.messageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessageType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseMessageType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
