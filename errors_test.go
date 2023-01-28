package whatsapp

import (
	"fmt"
	"testing"
)

func TestIsError(t *testing.T) {
	t.Parallel()
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestIsError",
			args: args{
				err: &Error{},
			},
			want: true,
		},
		{
			name: "TestIsError",
			args: args{
				err: fmt.Errorf("TestError"),
			},
			want: false,
		},
		{
			name: "TestIsError",
			args: args{
				err: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsError(tt.args.err); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	t.Parallel()
	type fields struct {
		Message   string
		Type      string
		Code      int
		Data      *ErrorData
		Subcode   int
		UserTitle string
		UserMsg   string
		FBTraceID string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "TestError_Error",
			fields: fields{
				Message:   "TestMessage",
				Type:      "TestType",
				Code:      1,
				Data:      &ErrorData{MessagingProduct: "TestMessagingProduct", Details: "TestDetails"},
				Subcode:   2,
				UserTitle: "TestUserTitle",
				UserMsg:   "TestUserMsg",
				FBTraceID: "TestFBTraceID",
			},
			want: "whatsapp: Message: TestMessage, Type: TestType, Code: 1, Data: Messaging Product: TestMessagingProduct, Details: TestDetails, Subcode: 2, UserTitle: TestUserTitle, UserMsg: TestUserMsg, FBTraceID: TestFBTraceID",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := &Error{
				Message:   tt.fields.Message,
				Type:      tt.fields.Type,
				Code:      tt.fields.Code,
				Data:      tt.fields.Data,
				Subcode:   tt.fields.Subcode,
				UserTitle: tt.fields.UserTitle,
				UserMsg:   tt.fields.UserMsg,
				FBTraceID: tt.fields.FBTraceID,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
