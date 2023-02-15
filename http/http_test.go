package http

import "testing"

func TestCreateRequestURL(t *testing.T) {
	t.Parallel()
	type args struct {
		params *RequestParams
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test create phone number verification request url",
			args: args{
				params: &RequestParams{
					BaseURL:    BaseURL,
					SenderID:   "224225226",
					ApiVersion: "v16.0",
					Endpoints:  []string{"verify_code"},
				},
			},
			want:    "https://graph.facebook.com/v16.0/224225226/verify_code",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CreateRequestURL(tt.args.params.BaseURL, tt.args.params.ApiVersion, tt.args.params.SenderID, tt.args.params.Endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRequestURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
