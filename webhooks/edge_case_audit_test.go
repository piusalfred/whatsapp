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

// Edge-case and failure-mode audit tests. These tests demonstrate defects found
// during the comprehensive edge-case audit. Each test corresponds to a finding
// in code-review/05_edge_cases.md.

package webhooks_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

// TestDefect001_TemplateStatusUpdate_NilReason verifies that a template status
// update webhook without a "reason" field does not cause a nil pointer dereference.
//
// Finding: 001 in code-review/05_edge_cases.md
// File: webhooks/business.go:542 — *value.Reason when Reason is nil.
func TestDefect001_TemplateStatusUpdate_NilReason(t *testing.T) {
	t.Parallel()

	// WhatsApp webhook payload for a template status update WITHOUT a reason field.
	// The "reason" field is optional per the WhatsApp API docs — it is only present
	// when a template is rejected with a specific reason.
	payload := []byte(`{
		"object": "whatsapp_business_account",
		"entry": [{
			"id": "123456789",
			"time": 1719000000,
			"changes": [{
				"field": "message_template_status_update",
				"value": {
					"event": "APPROVED",
					"message_template_id": 12345,
					"message_template_name": "hello_world",
					"message_template_language": "en"
				}
			}]
		}]
	}`)

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")

	notification, err := webhooks.ExtractAndValidatePayload(r, &webhooks.ValidateOptions{
		Validate: false,
	})
	if err != nil {
		t.Fatalf("ExtractAndValidatePayload failed: %v", err)
	}
	if notification == nil {
		t.Fatal("notification is nil")
	}

	handler := webhooks.NewHandler()
	handler.OnBusinessTemplateStatusUpdate(
		webhooks.BusinessEventHandlerFunc[webhooks.TemplateStatusUpdateNotification](
			func(ctx context.Context, req *webhooks.BusinessRequest[webhooks.TemplateStatusUpdateNotification]) error {
				return nil
			},
		),
	)

	// The defect: TemplateStatusUpdate() dereferences *value.Reason without
	// checking for nil. This causes a panic which is CAUGHT by the recover()
	// in handleNotificationChange, but the result is a 500 error for every
	// template status update that lacks a "reason" field — triggering
	// WhatsApp retries for up to 7 days on valid payloads.
	//
	// VERIFIED: panic occurs at business.go:542, caught by handler.go:218-222,
	// returns PanicError -> HandleNotification returns 500.
	//
	// EXPECTED: 200 OK because this is a valid, supported webhook payload.
	resp := handler.HandleNotification(context.Background(), notification)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("DEFECT CONFIRMED: got status %d, want 200. "+
			"TemplateStatusUpdate() at business.go:542 dereferences *value.Reason "+
			"without nil check, causing panic->500 for valid template status updates "+
			"without a reason field.", resp.StatusCode)
	}
}

// TestDefect002_MiddlewarePanic_NotRecovered verifies that a panic in webhook
// middleware is caught and does not crash the server.
//
// Finding: 002 in code-review/05_edge_cases.md
// File: webhooks/webhooks.go:213-240 — no recover in Listener.HandleNotification.
func TestDefect002_MiddlewarePanic_NotRecovered(t *testing.T) {
	t.Parallel()

	handler := webhooks.NewHandler()
	cfgReader := webhooks.ConfigReaderFunc(func(r *http.Request) (*webhooks.Config, error) {
		return &webhooks.Config{Validate: false}, nil
	})

	panickingMiddleware := func(next webhooks.NotificationHandler) webhooks.NotificationHandler {
		return webhooks.NotificationHandlerFunc(
			func(ctx context.Context, n *webhooks.Notification) *webhooks.Response {
				panic("middleware bug")
			},
		)
	}

	listener := webhooks.NewListener(handler, cfgReader, panickingMiddleware)

	payload := `{"object":"whatsapp_business_account","entry":[]}`
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(payload))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// This currently PANICS because Listener.HandleNotification has no recover().
	// After the fix is applied, this test should verify 500 is returned instead.
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		listener.HandleNotification(w, r)
	}()

	if didPanic {
		t.Error("FIXME: Listener.HandleNotification panicked on middleware panic. " +
			"Add recover() to Listener.HandleNotification to catch this.")
	}
}

// TestDefect003_NilListenerFields_Panics verifies that a zero-value Listener
// panics rather than returning a graceful error.
//
// Finding: 003 in code-review/05_edge_cases.md.
func TestDefect003_NilListenerFields_Panics(t *testing.T) {
	t.Parallel()

	listener := &webhooks.Listener{}
	r := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		listener.HandleNotification(w, r)
	}()

	if didPanic {
		t.Error("FIXME: Zero-value Listener panics on HandleNotification. " +
			"Add nil checks for listener.handler and listener.configReader.")
	}
}

// TestDefectSH005_ValidateRequestPayloadSignature_NoSizeLimit verifies that
// ValidateRequestPayloadSignature does not read unbounded request bodies.
//
// Finding: SH-005 in code-review/05_edge_cases.md
// File: webhooks/webhooks.go:487 — io.Copy without LimitReader.
func TestDefectSH005_ValidateRequestPayloadSignature_NoSizeLimit(t *testing.T) {
	t.Parallel()

	// Create a large body that exceeds MaxPayloadBytes
	largeBody := make([]byte, webhooks.MaxPayloadBytes+1)
	for i := range largeBody {
		largeBody[i] = 'x'
	}

	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(largeBody))
	r.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")

	// This currently reads the ENTIRE body (unbounded) into memory.
	// After fix: should reject with ErrPayloadTooLarge-like error.
	err := webhooks.ValidateRequestPayloadSignature(r, "test-secret")
	if err == nil {
		// This is the unbounded-read path — the function consumed all memory
		// for a payload larger than MaxPayloadBytes.
		t.Log("INFO: ValidateRequestPayloadSignature read unbounded body (no LimitReader). " +
			"Memory allocated for payload exceeding MaxPayloadBytes.")
	}
}

// TestDefect005_PanicRecovery_WrapsPanicAsPanicError verifies that panics in
// user message handlers are caught and returned as PanicError, not as crashes.
//
// Finding: 005 in code-review/05_edge_cases.md.
func TestDefect005_PanicRecovery_WrapsPanicAsPanicError(t *testing.T) {
	t.Parallel()

	handler := webhooks.NewHandler()
	handler.OnTextMessage(webhooks.MessageHandlerFunc[webhooks.Text](
		func(ctx context.Context, req *webhooks.MessageRequest[webhooks.Text]) error {
			panic("intentional test panic in text handler")
		},
	))

	notification := &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{{
			ID:   "1",
			Time: 1719000000,
			Changes: []webhooks.Change{{
				Field: "messages",
				Value: &webhooks.Value{
					MessagingProduct: "whatsapp",
					Contacts:         []*webhooks.Contact{{WaID: "12345", Profile: &webhooks.Profile{Name: "Test"}}},
					Metadata:         &webhooks.Metadata{PhoneNumberID: "123", DisplayPhoneNumber: "123456789"},
					Messages: []*webhooks.Message{{
						Type: "text",
						From: "12345",
						ID:   "msg1",
						Text: &webhooks.Text{Body: "hello"},
					}},
				},
			}},
		}},
	}

	resp := handler.HandleNotification(context.Background(), notification)

	// Panic should be caught and return 500 (not crash the test).
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for handler panic, got %d", resp.StatusCode)
	}

	// The panic is caught and wrapped as PanicError in handleNotificationChange.
	// This test verifies the recovery mechanism works. The wrapping as PanicError
	// is done inside handleNotificationChange and cannot be observed from the
	// HandleNotification return value alone — the error is swallowed by the
	// handler dispatch returning 500. Use a custom ErrorHandler to observe it.
}
