//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhooks

import (
	"context"
)

type (
	Call struct {
		ID                    string       `json:"id"` // The WhatsApp call ID
		To                    string       `json:"to"` // The WhatsApp user's phone number (callee)
		From                  string       `json:"from"`
		Event                 string       `json:"event"`
		Timestamp             string       `json:"timestamp"`
		Direction             string       `json:"direction"`
		DeepLinkPayload       string       `json:"deeplink_payload,omitempty"`
		CTAPayload            string       `json:"cta_payload,omitempty"`
		Status                string       `json:"status"`
		StartTime             string       `json:"start_time"`
		EndTime               string       `json:"end_time"`
		Duration              int          `json:"duration"`
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"`
		Session               *CallSession `json:"session,omitempty"`
	}

	CallSession struct {
		SDPType string `json:"sdp_type"`
		SDP     string `json:"sdp"`
	}

	CallStatusUpdate struct {
		MessagingProduct string
		Contacts         []*Contact
		Metadata         *Metadata `json:"metadata,omitempty"`
		Calls            []*Call
		Errors           []ErrorInfo
	}
)

func (value *Value) CallStatusUpdate() *CallStatusUpdate {
	return &CallStatusUpdate{
		MessagingProduct: value.MessagingProduct,
		Metadata:         value.Metadata,
		Contacts:         value.Contacts,
		Calls:            value.Calls,
		Errors:           value.Errors,
	}
}

func (handler *Handler) SetBusinessCallStatusUpdateHandler(
	fn EventHandler[BusinessNotificationContext, CallStatusUpdate],
) {
	handler.callStatusUpdate = fn
}

func (handler *Handler) OnCallStatusUpdate(
	fn func(ctx context.Context, notificationContext *BusinessNotificationContext, details *CallStatusUpdate) error,
) {
	handler.callStatusUpdate = EventHandlerFunc[BusinessNotificationContext, CallStatusUpdate](fn)
}
