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

// historyEntriesPayload returns a history sync notification with chat
// history entries.
func historyEntriesPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "history",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
							History: []webhooks.HistoryEntry{
								{
									Metadata: webhooks.HistoryMetadata{
										Phase:      1,
										ChunkOrder: 0,
										Progress:   50,
									},
									Threads: []webhooks.HistoryThread{
										{
											ID: "15550001111",
											Messages: []*webhooks.Message{
												{
													From:      "15550001111",
													ID:        "wamid.test123",
													Timestamp: "1739321024",
													Type:      "text",
													Text:      &webhooks.Text{Body: "Historical message"},
												},
											},
										},
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

// historyUnknownPayload returns a history notification with neither
// history entries nor messages — used to trigger the sub-fallback.
func historyUnknownPayload() *webhooks.Notification {
	return &webhooks.Notification{
		Object: "whatsapp_business_account",
		Entry: []webhooks.Entry{
			{
				ID:   "123456789",
				Time: 1739321024,
				Changes: []webhooks.Change{
					{
						Field: "history",
						Value: &webhooks.Value{
							MessagingProduct: "whatsapp",
							Metadata: &webhooks.Metadata{
								DisplayPhoneNumber: "15550783881",
								PhoneNumberID:      "106540352242922",
							},
						},
					},
				},
			},
		},
	}
}

func TestFallback_History_NoSubHandler_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()
	// No history handlers → handler.history is nil.

	resp := h.HandleNotification(context.Background(), historyEntriesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack for nil sub-handler), got %d", resp.StatusCode)
	}
}

func TestFallback_History_NoSubHandler_GeneralFallbackFires(t *testing.T) {
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

	resp := h.HandleNotification(context.Background(), historyEntriesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("general fallback was not invoked for nil history sub-handler")
	}
	if gotField != "history" {
		t.Errorf("expected field 'history', got %q", gotField)
	}
}

func TestFallback_History_DedicatedHandlerFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	var fired bool
	h.OnHistorySync(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
			fired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), historyEntriesPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("dedicated handler was not invoked")
	}
}

func TestFallback_History_SubFallbackFires(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	// Init history but don't set Messages or MediaMessages.
	// Send a payload with neither entries nor messages → sub-fallback fires.
	h.OnHistorySync(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
			t.Error("history sync handler should not fire for unknown payload")
			return nil
		},
	))
	// Clear Messages so it falls through to the sub-fallback.
	h.History().Messages = nil

	var subFired bool
	h.History().OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			subFired = true
			return nil
		},
	))

	resp := h.HandleNotification(context.Background(), historyUnknownPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !subFired {
		t.Fatal("sub-handler fallback was not invoked for unrecognized history payload")
	}
}

func TestFallback_History_NoSubFallback_Silent200(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnHistorySync(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
			return nil
		},
	))
	h.History().Messages = nil
	// History.Fallback is nil → silently 200.

	resp := h.HandleNotification(context.Background(), historyUnknownPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (silent ack), got %d", resp.StatusCode)
	}
}

func TestFallback_History_OnFallbackPropagates(t *testing.T) {
	t.Parallel()

	h := webhooks.NewHandler()

	h.OnHistorySync(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
		func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
			return nil
		},
	))
	h.History().Messages = nil

	if h.History().Fallback != nil {
		t.Fatal("History.Fallback should be nil before OnFallback")
	}

	var fired bool
	h.OnFallback(webhooks.FallbackHandlerFunc(
		func(_ context.Context, ne webhooks.NotificationEntry, c webhooks.Change) error {
			fired = true
			return nil
		},
	))

	if h.History().Fallback == nil {
		t.Fatal("OnFallback did not propagate to History.Fallback")
	}

	resp := h.HandleNotification(context.Background(), historyUnknownPayload())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !fired {
		t.Fatal("fallback was not invoked for unrecognized history payload")
	}
}
