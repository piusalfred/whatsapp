/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package webhooks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ExampleNewEventListener() {
	_ = NewEventListener(
		WithNotificationErrorHandler(NoOpNotificationErrorHandler),
		WithAfterFunc(func(ctx context.Context, notification *Notification, err error) {}),
		WithBeforeFunc(func(ctx context.Context, notification *Notification) error {
			return nil
		}),
		WithGenericNotificationHandler(
			func(ctx context.Context, writer http.ResponseWriter, notification *Notification, handler NotificationErrorHandler) error {
				return nil
			}),
		WithHooks(&Hooks{
			N: nil,
			S: nil,
			M: nil,
			H: nil,
		}),
		WithSubscriptionVerifier(func(ctx context.Context, request *VerificationRequest) error {
			return fmt.Errorf("subscription verification failed")
		}),
		WithHandlerOptions(&HandlerOptions{
			BeforeFunc:        nil,
			AfterFunc:         nil,
			ValidateSignature: false,
			Secret:            "lilsecretofold",
		}),
		WithHooksErrorHandler(NoOpHooksErrorHandler),
	)
}

func TestParseMessageType(t *testing.T) {
	t.Parallel()
	type args struct {
		messageType string
	}

	tests := []struct {
		name string
		args args
		want MessageType
	}{
		{
			name: "tExt",
			args: args{
				messageType: "text",
			},
			want: TextMessageType,
		},
		{
			name: "imageX",
			args: args{
				messageType: "imageX",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseMessageType(tt.args.messageType)
			if got != tt.want {
				t.Errorf("ParseMessageType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEncounteredError(t *testing.T) {
	t.Parallel()
	type args struct {
		nonFatalErrs []error
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantErrString string
	}{
		{
			name: "empty",
			args: args{
				nonFatalErrs: []error{},
			},
			wantErr:       false,
			wantErrString: "",
		},
		{
			name: "non single error",
			args: args{
				nonFatalErrs: []error{
					errors.New("single"),
				},
			},
			wantErr:       true,
			wantErrString: "single",
		},
		{
			name: "multiple errors",
			args: args{
				nonFatalErrs: []error{
					errors.New("single"),
					errors.New("double"),
					errors.New("triple"),
				},
			},
			wantErr:       true,
			wantErrString: "single, double, triple",
		},
		{
			name: "multiple errors more than 6",
			args: args{
				nonFatalErrs: []error{
					errors.New("single"),
					errors.New("double"),
					errors.New("triple"),
					errors.New("quadruple"),
					errors.New("quintuple"),
					errors.New("sextuple"),
					errors.New("septuple"),
				},
			},
			wantErr:       true,
			wantErrString: "single, double, triple, quadruple, quintuple, sextuple, septuple",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := getEncounteredError(tt.args.nonFatalErrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEncounteredError() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.wantErrString {
				t.Errorf("getEncounteredError() error = %v, wantErrString %v", err, tt.wantErrString)
			}
		})
	}
}

// humanReadableNotification is a human-readable representation of a notification.
func humanReadableNotification(t *testing.T, notification *Notification) string {
	t.Helper()
	var buf bytes.Buffer
	if notification == nil {
		return "notification: <nil>"
	}

	buf.WriteString("notification: object: ")
	if notification.Object == "" {
		buf.WriteString("<nil>")
	} else {
		buf.WriteString(notification.Object)
	}
	buf.WriteString(", entries: ")
	for i, entry := range notification.Entry {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(humanReadableEntry(t, entry))
	}

	return buf.String()
}

func humanReadableEntry(t *testing.T, entry *Entry) string {
	t.Helper()
	var buf bytes.Buffer
	if entry == nil {
		return "entry: <nil>"
	}
	buf.WriteString("entry: id: ")
	if entry.ID == "" {
		buf.WriteString("<nil>")
	} else {
		buf.WriteString(entry.ID)
	}
	buf.WriteString(", changes: ")
	for i, change := range entry.Changes {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(humanReadableChange(t, change))
	}
	return buf.String()
}

func humanReadableChange(t *testing.T, change *Change) string {
	return fmt.Sprintf("change: field: %s, value: %+v", change.Field, change.Value)
}

func TestNotificationHandler_Options(t *testing.T) {
	t.Parallel()
	type fields struct {
		BeforeFunc        BeforeFunc
		AfterFunc         AfterFunc
		ValidateSignature bool
		Secret            string
		Hooks             *Hooks
		Body              []byte
	}

	testcases := []struct {
		name       string
		fields     fields
		wantStatus int
	}{
		{
			name: "no options",
			fields: fields{
				BeforeFunc:        nil,
				AfterFunc:         nil,
				ValidateSignature: false,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with options",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"WHATSAPP_BUSINESS_ACCOUNT_ID","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"PHONE_NUMBER","phone_number_id":"PHONE_NUMBER_ID"},"contacts":[{"profile":{"name":"NAME"},"wa_id":"WHATSAPP_ID"}],"messages":[{"from":"PHONE_NUMBER","id":"wamid.ID","timestamp":"TIMESTAMP","type":"image","image":{"caption":"CAPTION","mime_type":"image/jpeg","sha256":"IMAGE_HASH","id":"ID"}}]},"field":"messages"}]}]}`),
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hooks := tt.fields.Hooks
			options := &HandlerOptions{
				BeforeFunc:        tt.fields.BeforeFunc,
				AfterFunc:         tt.fields.AfterFunc,
				ValidateSignature: tt.fields.ValidateSignature,
				Secret:            tt.fields.Secret,
			}
			h := NotificationHandler(hooks, NoOpNotificationErrorHandler, NoOpHooksErrorHandler, options)

			req, err := http.NewRequest("POST", "/webhook", bytes.NewReader(tt.fields.Body))
			if err != nil {
				t.Logf("error creating request: %v", err)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}
