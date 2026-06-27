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
	"testing"

	"github.com/piusalfred/whatsapp/webhooks"
)

func TestInteractiveHandler_DedicatedSubTypes(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.MessageNotificationContext{EntryID: "entry-1"}
	info := &webhooks.MessageInfo{Type: "interactive"}

	t.Run("list reply is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnListReply(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, d *webhooks.ListReply) error {
				called = true
				if d.ID != "option-1" {
					t.Errorf("ID = %q, want %q", d.ID, "option-1")
				}
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:      webhooks.InteractiveTypeListReply,
				ListReply: &webhooks.ListReply{ID: "option-1", Title: "Select Me", Description: "An option"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("ListReply handler was not called")
		}
	})

	t.Run("button reply is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnButtonReply(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, d *webhooks.ButtonReply) error {
				called = true
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:        webhooks.InteractiveTypeButtonReply,
				ButtonReply: &webhooks.ButtonReply{ID: "btn-1", Title: "Click"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("ButtonReply handler was not called")
		}
	})

	t.Run("flow completion is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnFlowCompletion(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, d *webhooks.NFMReply) error {
				called = true
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:     webhooks.InteractiveTypeNFMReply,
				NFMReply: &webhooks.NFMReply{Name: "flow-1", Body: "done", ResponseJSON: []byte(`{}`)},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("FlowCompletion handler was not called")
		}
	})

	t.Run("address submission is handled", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnAddressSubmission(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, d *webhooks.NFMReply) error {
				called = true
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:     webhooks.InteractiveAddressSubmission,
				NFMReply: &webhooks.NFMReply{Name: "addr-1", Body: "123 Main St", ResponseJSON: []byte(`{}`)},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("AddressSubmission handler was not called")
		}
	})
}

func TestInteractiveHandler_Fallback(t *testing.T) {
	t.Parallel()

	nctx := &webhooks.MessageNotificationContext{EntryID: "entry-1"}
	info := &webhooks.MessageInfo{Type: "interactive"}

	t.Run("unrecognized sub-type falls through to Interactive handler", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		var called bool
		ih.OnFallback(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, d *webhooks.Interactive) error {
				called = true
				if d.Type != "some_new_type" {
					t.Errorf("Type = %q, want %q", d.Type, "some_new_type")
				}
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type: "some_new_type",
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("Interactive fallback was not called")
		}
	})

	t.Run("nil Interactive payload is silently skipped", func(t *testing.T) {
		t.Parallel()

		ih := &webhooks.InteractiveHandler{}
		// No handlers set at all.

		msg := &webhooks.Message{Type: "interactive"}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Errorf("expected nil for nil Interactive, got: %v", err)
		}
	})

	t.Run("dedicated handler takes priority over fallback", func(t *testing.T) {
		t.Parallel()

		var dedicatedCalled, fallbackCalled bool

		ih := &webhooks.InteractiveHandler{}
		ih.OnButtonReply(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, _ *webhooks.ButtonReply) error {
				dedicatedCalled = true
				return nil
			},
		)
		ih.OnFallback(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, _ *webhooks.Interactive) error {
				fallbackCalled = true
				return nil
			},
		)

		msg := &webhooks.Message{
			Type: "interactive",
			Interactive: &webhooks.Interactive{
				Type:        webhooks.InteractiveTypeButtonReply,
				ButtonReply: &webhooks.ButtonReply{ID: "btn-1"},
			},
		}

		err := ih.Handle(context.Background(), nctx, info, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dedicatedCalled {
			t.Error("dedicated handler was not called")
		}
		if fallbackCalled {
			t.Error("fallback should not be called when dedicated handler is set")
		}
	})
}

func TestInteractiveHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnListReply sets ListReply", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnListReply(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, _ *webhooks.ListReply) error {
				return nil
			},
		)
		if ih.ListReply == nil {
			t.Error("ListReply should not be nil")
		}
	})

	t.Run("OnButtonReply sets ButtonReply", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnButtonReply(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, _ *webhooks.ButtonReply) error {
				return nil
			},
		)
		if ih.ButtonReply == nil {
			t.Error("ButtonReply should not be nil")
		}
	})

	t.Run("OnInteractive sets Interactive", func(t *testing.T) {
		t.Parallel()
		ih := &webhooks.InteractiveHandler{}
		ih.OnFallback(
			func(_ context.Context, _ *webhooks.MessageNotificationContext, _ *webhooks.MessageInfo, _ *webhooks.Interactive) error {
				return nil
			},
		)
		if ih.Fallback == nil {
			t.Error("Interactive should not be nil")
		}
	})
}
