package whatsapp_test

import (
	"testing"

	whttp "github.com/piusalfred/whatsapp/http"
)

func TestEndpointURLFromRequestParams(t *testing.T) {
	type args struct {
		ActionName  string
		params      *whttp.RequestParams
		ExpectedURL string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test verification code url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpointURL := whttp.EndpointURLFromRequestParams(tt.args.ActionName, tt.args.params)
			if endpointURL != tt.args.ExpectedURL {
				t.Errorf("expected %s, got %s", tt.args.ExpectedURL, endpointURL)
			}
		})
	}
}
