package webhooks

import (
	"context"
	"errors"
	"fmt"

	werrors "github.com/piusalfred/whatsapp/errors"
)

const (
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusSent      MessageStatus = "sent"
)

type (

	// NotificationContext is the context of a notification contains information about the
	// notification and the business that is subscribed to the Webhooks.
	// these are common fields to all notifications.
	// ID - The WhatsApp Business Account ID for the business that is subscribed to the webhook.
	// Contacts - Array of contact objects with information for the customer who sent a message
	// to the business
	// Metadata - A metadata object describing the business subscribed to the webhook.
	NotificationContext struct {
		ID       string
		Contacts []*Contact
		Metadata *Metadata
	}

	// Hooks is a generic interface for all hooks.
	Hooks interface {
		// OnMessageStatusChange is a hook that is called when a message status changes.
		// Status change is triggered when a message is sent or delivered to a customer or
		// the customer reads the delivered message sent by a business that is subscribed
		// to the Webhooks.
		OnMessageStatusChange(ctx context.Context, nctx *NotificationContext, status *Status) error

		// OnNotificationError is a hook that is called when a notification error occurs.
		// Sometimes a webhook notification being sent to a business contains errors.
		// This hook is called when a webhook notification contains errors.
		OnNotificationError(ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

		// OnMessageReceived is a hook that is called when a message is received.
		// This message can be a text message, image, video, audio, document, location,
		// vcard, template, sticker, or file. It can be a reply to a message sent by the
		// business or a new message.
		OnMessageReceived(ctx context.Context, nctx *NotificationContext, message *Message) error
	}

	// MessageStatus is the status of a message.
	// delivered – A webhook is triggered when a message sent by a business has been delivered
	// read – A webhook is triggered when a message sent by a business has been read
	// sent – A webhook is triggered when a business sends a message to a customer
	MessageStatus string

	// OnMessageStatusChange is a hook that is called when a message status changes.
	// Status change is triggered when a message is sent or delivered to a customer or
	// the customer reads the delivered message sent by a business that is subscribed
	// to the Webhooks.
	OnMessageStatusChange func(ctx context.Context, nctx *NotificationContext, status *Status) error

	// OnNotificationError is a hook that is called when a notification error occurs.
	// Sometimes a webhook notification being sent to a business contains errors.
	// This hook is called when a webhook notification contains errors.
	// waba is the Whatsapp Business Account ID
	OnNotificationError func(
		ctx context.Context, nctx *NotificationContext, errors *werrors.Error) error

	// OnMessageReceived is a hook that is called when a message is received.
	// This message can be a text message, image, video, audio, document, location,
	// vcard, template, sticker, or file. It can be a reply to a message sent by the
	// business or a new message.
	OnMessageReceived func(ctx context.Context, nctx *NotificationContext,
		messages *Message) error
)

type ErrorHandler func(err error) error

// ApplyHooks applies the hooks to notification received. Sometimes the hooks can return
// errors. The errors are collected and returned as a single error. So in your implementation
// of Hooks, you can return a FatalError if you want to stop the processing of the notification.
// immediately. If you want to continue processing the notification, you can return a non-fatal
// error. The errors are collected and returned as a single error.
// Also since all hooks errors are passed to the ApplyHooksErrorHandler, you can decide to either
// escalate the non-fatal errors to fatal errors or just ignore them also you can decide to
// ignore the fatal errors.
//
// Example:
//
//	func ShouldIgnoreFatalErrors(ignore bool) ErrorHandler{
//	    return func(err error) error {
//	        if IsFatalError(err) {
//	            if ignore {
//	                return fmt.Errorf("ignoring fatal error: %v", err)
//	            }
//	            return err
//	        }
//	        return err
//	    }
//	}
func ApplyHooks(ctx context.Context, notification *Notification, hooks Hooks,
	eh ErrorHandler) error {
	if notification == nil || hooks == nil {
		return nil
	}

	entries := notification.Entry
	for _, entry := range entries {
		entry := entry
		changes := entry.Changes
		for _, change := range changes {
			change := change
			value := change.Value
			if value == nil {
				continue
			}

			nctx := &NotificationContext{
				ID:       entry.ID,
				Contacts: value.Contacts,
				Metadata: value.Metadata,
			}

			// call the hooks
			return applyHooks(ctx, nctx, value, hooks, eh)
		}

	}

	return nil

}

type FatalError struct {
	Err  error
	Desc string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Desc, e.Err.Error())
}

func IsFatalError(err error) bool {
	_, ok := err.(*FatalError)
	return ok
}

func applyHooks(ctx context.Context, nctx *NotificationContext, value *Value, hooks Hooks, ef ErrorHandler) error {
	if hooks == nil {
		return nil
	}

	var allErrors []error

	// call the hooks
	if value.Errors != nil {
		for _, ev := range value.Errors {
			ev := ev
			if err := hooks.OnNotificationError(ctx, nctx, ev); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Statuses != nil {
		for _, sv := range value.Statuses {
			sv := sv
			if err := hooks.OnMessageStatusChange(ctx, nctx, sv); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if value.Messages != nil {
		for _, mv := range value.Messages {
			mv := mv
			if err := hooks.OnMessageReceived(ctx, nctx, mv); err != nil {
				if IsFatalError(err) {
					return err
				}
				allErrors = append(allErrors, err)
			}
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}
