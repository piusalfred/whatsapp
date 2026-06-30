// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Fuzz tests for webhook payload parsing. These verify that the JSON
// unmarshaling paths never panic on arbitrary input and that the
// dispatch layer handles malformed payloads gracefully.

package webhooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

// FuzzNotificationUnmarshal verifies that Notification unmarshaling never
// panics on arbitrary input.
func FuzzNotificationUnmarshal(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"object": "not_whatsapp"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var n webhooks.Notification
		_ = json.Unmarshal(data, &n)
	})
}

// FuzzValueUnmarshal verifies that Value unmarshaling never panics
// on arbitrary input.
func FuzzValueUnmarshal(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"messaging_product":"whatsapp"}`))
	f.Add([]byte(`{"statuses":[{"status":"sent"}]}`))
	f.Add([]byte(`{"errors":[{"code":400}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var v webhooks.Value
		_ = json.Unmarshal(data, &v) // must never panic
	})
}

// FuzzMessageUnmarshal verifies that Message unmarshaling never panics
// on arbitrary input.
func FuzzMessageUnmarshal(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"text","text":{"body":"hello"}}`))
	f.Add([]byte(`{"type":"unknown_xyz","from":"123"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var m webhooks.Message
		_ = json.Unmarshal(data, &m) // must never panic
	})
}

// FuzzExtractAndValidatePayload verifies that the full payload extraction
// path never panics on arbitrary input.
func FuzzExtractAndValidatePayload(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`not json`))

	f.Fuzz(func(t *testing.T, body []byte) {
		r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		_, _ = webhooks.ExtractAndValidatePayload(r, &webhooks.ValidateOptions{
			Validate: false,
		})
	})
}

// FuzzHandleNotification verifies that the full dispatch path never panics
// on arbitrary notification shapes.
func FuzzHandleNotification(f *testing.F) {
	f.Add([]byte(`{"object":"whatsapp_business_account","entry":[]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var n webhooks.Notification
		if err := json.Unmarshal(data, &n); err != nil {
			return // skip malformed JSON
		}

		h := webhooks.NewHandler()
		resp := h.HandleNotification(context.Background(), &n)
		// Must never panic; status code is 200, 500, or 504.
		if resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusInternalServerError &&
			resp.StatusCode != http.StatusGatewayTimeout {
			t.Errorf("unexpected status %d", resp.StatusCode)
		}
	})
}
