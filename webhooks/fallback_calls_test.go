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

package webhooks_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

// callConnectPayload returns a calls notification for a connect event
// (WebRTC call ready with SDP Answer).
func callConnectPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Contacts: []*webhooks.Contact{
								{
									WaID:    "15550001111",
									Profile: &webhooks.Profile{Name: "Test User"},
								},
							},
							Calls: []*webhooks.Call{
								{
									ID:        "wacid.test123",
									To:        "15550001111",
									From:      "15550783881",
									Event:     "connect",
									Timestamp: "1739321024",
									Direction: "USER_INITIATED",
									Session: &webhooks.CallSession{
										SDPType: "answer",
										SDP:     "v=0...",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// callTerminatePayload returns a calls notification for a terminate event
// — a different call event type for sub-fallback testing.
func callTerminatePayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Calls: []*webhooks.Call{
								{
									ID:        "wacid.test456",
									To:        "15550001111",
									From:      "15550783881",
									Event:     "terminate",
									Timestamp: "1739321024",
									Direction: "USER_INITIATED",
									Status:    "COMPLETED",
									StartTime: "1739321000",
									EndTime:   "1739321120",
									Duration:  120,
								},
							},
						},
					},
				},
			},
		},
	}
}

// callStatusPayload returns a calls notification with statuses
// (type "call" — ringing/accepted/rejected).
func callStatusPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "calls",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							Statuses: []*webhooks.Status{
								{
									ID:          "wacid.status123",
									Timestamp:   "1739321024",
									Type:        "call",
									StatusValue: "RINGING",
									RecipientID: "15550001111",
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_Calls_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	// No calls handlers → handler.calls is nil.

	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (calls silently skipped when sub-handler nil), got %d", resp.StatusCode)
	}
}

func TestFallback_Calls_NoSubHandler_GeneralFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	var gotField string
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			gotField = c.Field
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil calls sub-handler")
	}
	if gotField != "calls" {
		t.Errorf("expected field 'calls', got %q", gotField)
	}
}

func TestFallback_Calls_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			fired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_Calls_SubFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register connect handler to init handler.calls.
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			t.Error("connect handler should not fire for terminate event")
			return nil
		},
	))

	// Set sub-fallback. Terminate handler is nil → should hit this.
	var subFired bool
	h.Calls().Fallback = webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			subFired = true
			return nil
		},
	)

	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unhandled call event")
	}
}

func TestFallback_Calls_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return nil
		},
	))
	// CallsHandler.Fallback is nil, Terminate handler is nil → silently 200.

	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_Calls_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Register connect handler to init handler.calls.
	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return nil
		},
	))

	if h.Calls().Fallback != nil {
		t.Fatal("Calls.Fallback should be nil before OnFallback")
	}

	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			return nil
		},
	))

	// Propagation should have set CallsHandler.Fallback directly.
	if h.Calls().Fallback == nil {
		t.Fatal("OnFallback did not propagate to CallsHandler.Fallback")
	}

	// Send an unhandled event — CallsHandler.Fallback fires.
	resp := h.HandleNotification(context.Background(), callTerminatePayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("propagated fallback was not invoked for unhandled call event")
	}
}

// TestFallback_Calls_DedicatedStatusHandlerFires verifies that call status
// events (type "call" in statuses array) are dispatched to the Status handler.
func TestFallback_Calls_DedicatedStatusHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnCallStatus(webhooks.CallsEventHandlerFunc[webhooks.Status](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Status]) error {
			fired = true
			if req.Payload.Type != "call" {
				t.Errorf("expected status type 'call', got %q", req.Payload.Type)
			}
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), callStatusPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated call status handler was not invoked")
	}
}

// TestFallback_Calls_ErrorPropagationToErrorHandler verifies that errors from
// call handlers are routed through the CallsHandler's ErrorHandler.
func TestFallback_Calls_ErrorPropagationToErrorHandler(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnCallConnect(webhooks.CallsEventHandlerFunc[webhooks.Call](
		func(_ context.Context, req *webhooks.CallRequest[webhooks.Call]) error {
			return context.Canceled // triggers error handler
		},
	))

	var gotErr error
	h.Calls().ErrorHandler = webhooks.ErrorHandlerFunc(
		func(_ context.Context, err error) error {
			gotErr = err
			return nil // non-fatal
		},
	)

	resp := h.HandleNotification(context.Background(), callConnectPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (non-fatal error), got %d", resp.StatusCode)
	}
	if gotErr == nil {
		t.Fatal("error handler was not invoked for handler error")
	}
}
