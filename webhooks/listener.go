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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// EventListener wraps all the parts needed to listen and respond to incoming events
// to registered webhooks.
// It contains unexported *Hooks, *HandlerOptions, HooksErrorHandler, NotificationErrorHandler
// SubscriptionVerifier and GlobalNotificationHandler.
// All these can be set via exported ListenerOption functions like WithBeforeFunc and
// Setter methods like GlobalNotificationHandler which sets the GlobalNotificationHandler.
//
// Example:
//
//	  listener := NewEventListener(
//			WithNotificationErrorHandler(NoOpNotificationErrorHandler),
//			WithAfterFunc(func(ctx context.Context, notification *Notification, err error) {}),
//			WithBeforeFunc(func(ctx context.Context, notification *Notification) error {}),
//			WithGlobalNotificationHandler(nil),
//			WithHooks(&Hooks{
//				OnOrderMessageHook:        nil,
//				OnButtonMessageHook:       nil,
//				OnLocationMessageHook:     nil,
//				OnContactsMessageHook:     nil,
//				OnMessageReactionHook:     nil,
//				OnUnknownMessageHook:      nil,
//				OnProductEnquiryHook:      nil,
//				OnInteractiveMessageHook:  nil,
//				OnMessageErrorsHook:       nil,
//				OnTextMessageHook:         nil,
//				OnReferralMessageHook:     nil,
//			}),
//	  )
//
//	  example of setting a hook
//
//	  listener.OnOrderMessage(func(ctx context.Context, notification *Notification, order *Order) error {
//			// do something with the order
//			return nil
//	  })
//
//	  using a generic handler
//	   := listener.GlobalNotificationHandler()
type EventListener struct {
	h       *Hooks
	hef     HooksErrorHandler
	neh     NotificationErrorHandler
	v       SubscriptionVerifier
	options *HandlerOptions
	g       GlobalNotificationHandler
}

type ListenerOption func(*EventListener)

func NewEventListener(options ...ListenerOption) *EventListener {
	listener := &EventListener{
		h:   nil,
		hef: NoOpHooksErrorHandler,
		neh: NoOpNotificationErrorHandler,
		v:   nil,
		options: &HandlerOptions{
			BeforeFunc:        nil,
			AfterFunc:         nil,
			ValidateSignature: false,
			Secret:            "",
		},
		g: nil,
	}

	for _, option := range options {
		option(listener)
	}

	return listener
}

func (ls *EventListener) GenericNotificationHandler(handler GlobalNotificationHandler) {
	ls.g = handler
}

func (ls *EventListener) SubscriptionVerifier(verifier SubscriptionVerifier) {
	ls.v = verifier
}

func (ls *EventListener) NotificationErrorHandler(handler NotificationErrorHandler) {
	ls.neh = handler
}

func (ls *EventListener) HooksErrorHandler(handler HooksErrorHandler) {
	ls.hef = handler
}

func (ls *EventListener) OnOrderMessage(hook OnOrderMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnOrderMessageHook = hook
}

func (ls *EventListener) OnButtonMessage(hook OnButtonMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnButtonMessageHook = hook
}

func (ls *EventListener) OnLocationMessage(hook OnLocationMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnLocationMessageHook = hook
}

func (ls *EventListener) OnContactsMessage(hook OnContactsMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnContactsMessageHook = hook
}

func (ls *EventListener) OnMessageReaction(hook OnMessageReactionHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnMessageReactionHook = hook
}

func (ls *EventListener) OnUnknownMessage(hook OnUnknownMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnUnknownMessageHook = hook
}

func (ls *EventListener) OnProductEnquiry(hook OnProductEnquiryHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnProductEnquiryHook = hook
}

func (ls *EventListener) OnInteractiveMessage(hook OnInteractiveMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnInteractiveMessageHook = hook
}

func (ls *EventListener) OnMessageErrors(hook OnMessageErrorsHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnMessageErrorsHook = hook
}

func (ls *EventListener) OnTextMessage(hook OnTextMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnTextMessageHook = hook
}

func (ls *EventListener) OnReferralMessage(hook OnReferralMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnReferralMessageHook = hook
}

func (ls *EventListener) OnCustomerIDChange(hook OnCustomerIDChangeMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnCustomerIDChangeHook = hook
}

func (ls *EventListener) OnSystemMessage(hook OnSystemMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnSystemMessageHook = hook
}

func (ls *EventListener) OnMediaMessage(hook OnMediaMessageHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnMediaMessageHook = hook
}

func (ls *EventListener) OnNotificationError(hook OnNotificationErrorHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnNotificationErrorHook = hook
}

func (ls *EventListener) OnMessageStatusChange(hook OnMessageStatusChangeHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnMessageStatusChangeHook = hook
}

func (ls *EventListener) OnMessageReceived(hook OnMessageReceivedHook) {
	if ls.h == nil {
		ls.h = &Hooks{}
	}
	ls.h.OnMessageReceivedHook = hook
}

func WithGlobalNotificationHandler(g GlobalNotificationHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.g = g
	}
}

func WithHooks(hooks *Hooks) ListenerOption {
	return func(ls *EventListener) {
		ls.h = hooks
	}
}

func WithHooksErrorHandler(hooksErrorHandler HooksErrorHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.hef = hooksErrorHandler
	}
}

func WithNotificationErrorHandler(notificationErrorHandler NotificationErrorHandler) ListenerOption {
	return func(ls *EventListener) {
		ls.neh = notificationErrorHandler
	}
}

func WithSubscriptionVerifier(verifier SubscriptionVerifier) ListenerOption {
	return func(ls *EventListener) {
		ls.v = verifier
	}
}

func WithHandlerOptions(options *HandlerOptions) ListenerOption {
	return func(ls *EventListener) {
		ls.options = options
	}
}

func WithBeforeFunc(beforeFunc BeforeFunc) ListenerOption {
	return func(ls *EventListener) {
		if ls.options == nil {
			ls.options = &HandlerOptions{}
		}
		ls.options.BeforeFunc = beforeFunc
	}
}

func WithAfterFunc(afterFunc AfterFunc) ListenerOption {
	return func(ls *EventListener) {
		if ls.options == nil {
			ls.options = &HandlerOptions{}
		}
		ls.options.AfterFunc = afterFunc
	}
}

// NotificationHandler returns a http.Handler that can be used to handle the notification.
func (ls *EventListener) NotificationHandler() http.Handler {
	return NotificationHandler(ls.h, ls.neh, ls.hef, ls.options)
}

// GlobalHandler returns a http.Handler that handles all type of notification in one function.
// It  calls GlobalNotificationHandler. So before using this function, you should set GlobalNotificationHandler
// with WithGlobalNotificationHandler.
//
//nolint:cyclop
func (ls *EventListener) GlobalHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var buff bytes.Buffer
		_, err := io.Copy(&buff, request.Body)
		defer func() {
			request.Body = io.NopCloser(&buff)
		}()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if ls.options != nil && ls.options.ValidateSignature {
			signature, _ := ExtractSignatureFromHeader(request.Header)
			if !ValidateSignature(buff.Bytes(), signature, ls.options.Secret) {
				if handleError(
					request.Context(), writer, request,
					ls.neh, ErrInvalidSignature) {
					return
				}
			}
		}

		// Construct the notification
		var notification Notification
		if err := json.NewDecoder(&buff).Decode(&notification); err != nil && !errors.Is(err, io.EOF) {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		// call the generic handler
		if err := ls.g(request.Context(), writer, &notification); err != nil {
			err = fmt.Errorf("%w: %w", ErrOnGenericHandlerFunc, err)
			if handleError(request.Context(), writer, request, ls.neh, err) {
				return
			}
		}

		writer.WriteHeader(http.StatusOK)
	})
}

// HandleNotification handles all the notification types. It is a Global/Generic handler.

// SubscriptionVerificationHandler returns a http.Handler that can be used to verify the subscription.
func (ls *EventListener) SubscriptionVerificationHandler() http.Handler {
	return VerifySubscriptionHandler(ls.v)
}
