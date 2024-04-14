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
	"fmt"
	"testing"
)

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
			want: MessageTypeText,
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
	t.Helper()

	return fmt.Sprintf("change: field: %s, value: %+v", change.Field, change.Value)
}

func TestExtractSignatureFromHeader(t *testing.T) {
	t.Parallel()
	type args struct {
		header map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid signature",
			args: args{
				header: map[string][]string{
					SignatureHeaderKey: {"sha256=1234567890"},
				},
			},
			want:    "1234567890",
			wantErr: false,
		},
		{
			name: "invalid signature",
			args: args{
				header: map[string][]string{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		name := tt.name
		args := tt.args
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := ExtractSignatureFromHeader(args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSignatureFromHeader() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("ExtractSignatureFromHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}
