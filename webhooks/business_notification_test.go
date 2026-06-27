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

func TestBusinessNotificationHandler_DedicatedCall(t *testing.T) {
	t.Parallel()

	notification := &webhooks.Notification{Object: "whatsapp_business_account"}
	entry := webhooks.Entry{ID: "entry-1", Time: 123456}

	t.Run("alerts handler is called", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{}
		var called bool
		bh.OnAlerts(
			func(_ context.Context, _ *webhooks.BusinessNotificationContext, d *webhooks.AlertNotification) error {
				called = true
				if d.AlertSeverity != "HIGH" {
					t.Errorf("AlertSeverity = %q, want %q", d.AlertSeverity, "HIGH")
				}
				return nil
			},
		)

		change := webhooks.Change{
			Field: webhooks.ChangeFieldAccountAlerts.String(),
			Value: &webhooks.Value{AlertSeverity: "HIGH"},
		}

		err := bh.Handle(context.Background(), notification, change, entry, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("alerts handler was not called")
		}
	})

	t.Run("nil handler is silently skipped", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{} // no handlers set

		change := webhooks.Change{
			Field: webhooks.ChangeFieldSecurity.String(),
			Value: &webhooks.Value{Event: "PIN_CHANGED"},
		}

		err := bh.Handle(context.Background(), notification, change, entry, nil)
		if err != nil {
			t.Errorf("expected nil for unhandled change, got: %v", err)
		}
	})

	t.Run("handler error propagates through onError", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnCalls(
			func(_ context.Context, _ *webhooks.BusinessNotificationContext, _ *webhooks.CallStatusUpdate) error {
				return errDummy
			},
		)

		var onErrorCalled bool
		onError := func(_ context.Context, err error) error {
			onErrorCalled = true
			if err == nil {
				t.Error("expected non-nil error from handler")
			}
			return nil
		}

		change := webhooks.Change{
			Field: webhooks.ChangeFieldCalls.String(),
			Value: &webhooks.Value{},
		}

		err := bh.Handle(context.Background(), notification, change, entry, onError)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !onErrorCalled {
			t.Error("onError was not called")
		}
	})

	t.Run("onError returns non-nil stops processing", func(t *testing.T) {
		t.Parallel()

		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnCapability(
			func(_ context.Context, _ *webhooks.BusinessNotificationContext, _ *webhooks.CapabilityUpdate) error {
				return errDummy
			},
		)

		onError := func(_ context.Context, _ error) error {
			return errDummy
		}

		change := webhooks.Change{
			Field: webhooks.ChangeFieldBusinessCapabilityUpdate.String(),
			Value: &webhooks.Value{},
		}

		err := bh.Handle(context.Background(), notification, change, entry, onError)
		if err == nil {
			t.Error("expected error from onError, got nil")
		}
	})
}

func TestBusinessNotificationHandler_Setters(t *testing.T) {
	t.Parallel()

	t.Run("OnAlerts sets Alerts", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnAlerts(
			func(_ context.Context, _ *webhooks.BusinessNotificationContext, _ *webhooks.AlertNotification) error {
				return nil
			},
		)
		if bh.Alerts == nil {
			t.Error("Alerts should not be nil after OnAlerts")
		}
	})

	t.Run("OnAccount sets Account", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnAccount(func(_ context.Context, _ *webhooks.BusinessNotificationContext, _ *webhooks.AccountUpdate) error {
			return nil
		})
		if bh.Account == nil {
			t.Error("Account should not be nil after OnAccount")
		}
	})

	t.Run("OnSecurity sets Security", func(t *testing.T) {
		t.Parallel()
		bh := &webhooks.BusinessNotificationHandler{}
		bh.OnSecurity(
			func(_ context.Context, _ *webhooks.BusinessNotificationContext, _ *webhooks.SecurityNotification) error {
				return nil
			},
		)
		if bh.Security == nil {
			t.Error("Security should not be nil after OnSecurity")
		}
	})
}
