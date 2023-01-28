package whatsapp

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestMessageFn(t *testing.T) {
	type args struct {
		ctx     context.Context
		client  *http.Client
		url     string
		token   string
		message *Message
	}
	tests := []struct {
		name    string
		args    args
		want    *MessageResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MessageFn(tt.args.ctx, tt.args.client, tt.args.url, tt.args.token, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageFn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageFn() = %v, want %v", got, tt.want)
			}
		})
	}
}
