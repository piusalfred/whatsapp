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
	"context"
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

func TestHistoryHandler_DedicatedHandlers(t *testing.T) {
	t.Parallel()

	ne := webhooks.NotificationEntry{Object: "whatsapp_business_account", ID: "entry-1", Time: 123456}

	t.Run("history entries calls Messages handler", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var called bool
		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, req *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				called = true
				if len(req.Payload) != 1 {
					t.Errorf("expected 1 entry, got %d", len(req.Payload))
				}
				if req.Payload[0].Metadata.Phase != 1 {
					t.Errorf("Phase = %d, want 1", req.Payload[0].Metadata.Phase)
				}
				if req.Notification.EntryID != "entry-1" {
					t.Errorf("EntryID = %q, want %q", req.Notification.EntryID, "entry-1")
				}
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 1, ChunkOrder: 1, Progress: 50}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("Messages handler was not called")
		}
	})

	t.Run("media messages calls MediaMessages handler", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var called bool
		hh.OnMediaMessages(webhooks.MessageHandlerFunc[webhooks.Message](
			func(_ context.Context, req *webhooks.MessageRequest[webhooks.Message]) error {
				called = true
				if req.Info.MessageID != "wamid.abc" {
					t.Errorf("MessageID = %q, want %q", req.Info.MessageID, "wamid.abc")
				}
				if req.Info.From != "15550783881" {
					t.Errorf("From = %q, want %q", req.Info.From, "15550783881")
				}
				if req.Payload.ID != "wamid.abc" {
					t.Errorf("Payload.ID = %q, want %q", req.Payload.ID, "wamid.abc")
				}
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				Messages: []*webhooks.Message{
					{
						From:      "15550783881",
						ID:        "wamid.abc",
						Timestamp: "1738796547",
						Type:      "image",
					},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("MediaMessages handler was not called")
		}
	})

	t.Run("history entries with nil handler falls back", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var fallbackCalled bool
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				fallbackCalled = true
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 0}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fallbackCalled {
			t.Error("Fallback was not called when Messages handler is nil")
		}
	})

	t.Run("media messages with nil handler falls back", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var fallbackCalled bool
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				fallbackCalled = true
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				Messages: []*webhooks.Message{
					{ID: "wamid.xyz", Type: "image"},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fallbackCalled {
			t.Error("Fallback was not called when MediaMessages handler is nil")
		}
	})

	t.Run("Messages handler error propagates", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, _ *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				return errDummy
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 0}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err == nil {
			t.Error("expected error from Messages handler, got nil")
		}
	})

	t.Run("MediaMessages handler error propagates", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		hh.OnMediaMessages(webhooks.MessageHandlerFunc[webhooks.Message](
			func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Message]) error {
				return errDummy
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				Messages: []*webhooks.Message{
					{ID: "mid-1", Type: "image"},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err == nil {
			t.Error("expected error from MediaMessages handler, got nil")
		}
	})

	t.Run("empty history and empty messages falls back", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var fallbackCalled bool
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				fallbackCalled = true
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !fallbackCalled {
			t.Error("Fallback was not called for empty history payload")
		}
	})

	t.Run("empty history with nil fallback returns nil", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{} // no handlers, no fallback

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Errorf("expected nil for empty payload with no fallback, got: %v", err)
		}
	})
}

func TestHistoryHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnMessages sets Messages", func(t *testing.T) {
		t.Parallel()
		hh := &webhooks.HistoryHandler{}
		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, _ *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				return nil
			},
		))
		if hh.Messages == nil {
			t.Error("Messages should not be nil after OnMessages")
		}
	})

	t.Run("OnMediaMessages sets MediaMessages", func(t *testing.T) {
		t.Parallel()
		hh := &webhooks.HistoryHandler{}
		hh.OnMediaMessages(webhooks.MessageHandlerFunc[webhooks.Message](
			func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Message]) error {
				return nil
			},
		))
		if hh.MediaMessages == nil {
			t.Error("MediaMessages should not be nil after OnMediaMessages")
		}
	})

	t.Run("OnFallback sets Fallback", func(t *testing.T) {
		t.Parallel()
		hh := &webhooks.HistoryHandler{}
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				return nil
			},
		))
		if hh.Fallback == nil {
			t.Error("Fallback should not be nil after OnFallback")
		}
	})
}

func TestHistoryHandler_DedicatedOverridesFallback(t *testing.T) {
	t.Parallel()

	ne := webhooks.NotificationEntry{Object: "whatsapp_business_account", ID: "entry-1", Time: 123456}

	t.Run("history entries prefer Messages over fallback", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var dedicatedCalled, fallbackCalled bool

		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, _ *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				dedicatedCalled = true
				return nil
			},
		))
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				fallbackCalled = true
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 1}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dedicatedCalled {
			t.Error("Messages handler was not called")
		}
		if fallbackCalled {
			t.Error("Fallback should not be called when Messages is set")
		}
	})

	t.Run("media messages prefer MediaMessages over fallback", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		var dedicatedCalled, fallbackCalled bool

		hh.OnMediaMessages(webhooks.MessageHandlerFunc[webhooks.Message](
			func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Message]) error {
				dedicatedCalled = true
				return nil
			},
		))
		hh.OnFallback(webhooks.FallbackHandlerFunc(
			func(_ context.Context, _ webhooks.NotificationEntry, _ webhooks.Change) error {
				fallbackCalled = true
				return nil
			},
		))

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				Messages: []*webhooks.Message{
					{ID: "msg-1", Type: "video"},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dedicatedCalled {
			t.Error("MediaMessages handler was not called")
		}
		if fallbackCalled {
			t.Error("Fallback should not be called when MediaMessages is set")
		}
	})
}

func TestHistoryHandler_ErrorHandler(t *testing.T) {
	t.Parallel()

	ne := webhooks.NotificationEntry{Object: "whatsapp_business_account", ID: "entry-1", Time: 123456}

	t.Run("ErrorHandler swallows non-fatal errors from Messages", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, _ *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				return errDummy
			},
		))
		hh.ErrorHandler = webhooks.ErrorHandlerFunc(func(_ context.Context, err error) error {
			return nil // non-fatal, continue
		})

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 0}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Errorf("expected nil after non-fatal error handler, got: %v", err)
		}
	})

	t.Run("ErrorHandler escalates fatal errors from Messages", func(t *testing.T) {
		t.Parallel()

		hh := &webhooks.HistoryHandler{}
		hh.OnMessages(webhooks.ChangeValueHandlerFunc[webhooks.HistoryEntry](
			func(_ context.Context, _ *webhooks.ChangeValueRequest[webhooks.HistoryEntry]) error {
				return errDummy
			},
		))
		hh.ErrorHandler = webhooks.ErrorHandlerFunc(func(_ context.Context, err error) error {
			return err // fatal, escalate
		})

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				History: []webhooks.HistoryEntry{
					{Metadata: webhooks.HistoryMetadata{Phase: 0}},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err == nil {
			t.Error("expected error after fatal error handler, got nil")
		}
	})

	t.Run("ErrorHandler routes MediaMessages errors", func(t *testing.T) {
		t.Parallel()

		var onErrorCalled bool
		hh := &webhooks.HistoryHandler{}
		hh.OnMediaMessages(webhooks.MessageHandlerFunc[webhooks.Message](
			func(_ context.Context, _ *webhooks.MessageRequest[webhooks.Message]) error {
				return errDummy
			},
		))
		hh.ErrorHandler = webhooks.ErrorHandlerFunc(func(_ context.Context, _ error) error {
			onErrorCalled = true
			return nil
		})

		change := webhooks.Change{
			Field: webhooks.ChangeFieldHistory.String(),
			Value: &webhooks.Value{
				Messages: []*webhooks.Message{
					{ID: "msg-1", Type: "image"},
				},
			},
		}

		err := hh.Handle(context.Background(), ne, change)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !onErrorCalled {
			t.Error("ErrorHandler was not called for MediaMessages error")
		}
	})
}
