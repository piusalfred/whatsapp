//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks_test

import (
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

func TestParseMessageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected webhooks.MessageType
	}{
		{"text", webhooks.MessageTypeText},
		{"image", webhooks.MessageTypeImage},
		{"audio", webhooks.MessageTypeAudio},
		{"video", webhooks.MessageTypeVideo},
		{"document", webhooks.MessageTypeDocument},
		{"sticker", webhooks.MessageTypeSticker},
		{"location", webhooks.MessageTypeLocation},
		{"contacts", webhooks.MessageTypeContacts},
		{"reaction", webhooks.MessageTypeReaction},
		{"interactive", webhooks.MessageTypeInteractive},
		{"button", webhooks.MessageTypeButton},
		{"order", webhooks.MessageTypeOrder},
		{"system", webhooks.MessageTypeSystem},
		{"request_welcome", webhooks.MessageTypeRequestWelcome},
		{"unsupported", webhooks.MessageTypeUnsupported},
		{"unknown", webhooks.MessageTypeUnknown},
		{"revoke", webhooks.MessageTypeRevoke},
		{"edit", webhooks.MessageTypeEdit},
		// Case-insensitive
		{"TEXT", webhooks.MessageTypeText},
		{"Image", webhooks.MessageTypeImage},
		// Whitespace trimming
		{"  text  ", webhooks.MessageTypeText},
		// Unknown type
		{"nonexistent", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := webhooks.ParseMessageType(tt.input)
			if got != tt.expected {
				t.Errorf("ParseMessageType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMessage_IsAReply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  *webhooks.Message
		want bool
	}{
		{
			name: "nil context",
			msg:  &webhooks.Message{Context: nil},
			want: false,
		},
		{
			name: "forwarded message",
			msg: &webhooks.Message{
				Context: &webhooks.Context{Forwarded: true},
			},
			want: false,
		},
		{
			name: "product inquiry",
			msg: &webhooks.Message{
				Context: &webhooks.Context{
					ReferredProduct: &webhooks.ReferredProduct{},
				},
			},
			want: false,
		},
		{
			name: "forwarded and product inquiry",
			msg: &webhooks.Message{
				Context: &webhooks.Context{
					Forwarded:       true,
					ReferredProduct: &webhooks.ReferredProduct{},
				},
			},
			want: false,
		},
		{
			name: "plain reply",
			msg: &webhooks.Message{
				Context: &webhooks.Context{ID: "wamid_xxx"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.msg.IsAReply(); got != tt.want {
				t.Errorf("IsAReply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageNotificationContext_SenderInfo(t *testing.T) {
	t.Parallel()

	t.Run("empty contacts returns nil", func(t *testing.T) {
		t.Parallel()
		ctx := &webhooks.MessageNotificationContext{
			Contacts: nil,
		}
		if got := ctx.SenderInfo(); got != nil {
			t.Errorf("SenderInfo() = %v, want nil for empty contacts", got)
		}
	})

	t.Run("returns first contact", func(t *testing.T) {
		t.Parallel()
		ctx := &webhooks.MessageNotificationContext{
			Contacts: []*webhooks.Contact{
				{Profile: &webhooks.Profile{Name: "Alice"}, WaID: "111"},
				{Profile: &webhooks.Profile{Name: "Bob"}, WaID: "222"},
			},
		}
		got := ctx.SenderInfo()
		if got == nil {
			t.Fatal("SenderInfo() returned nil")
		}
		if got.Name != "Alice" || got.WaID != "111" {
			t.Errorf("SenderInfo() = {Name: %q, WaID: %q}, want {Alice, 111}", got.Name, got.WaID)
		}
	})
}
