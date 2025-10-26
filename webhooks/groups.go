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

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type (
	Group struct {
		Timestamp           string             `json:"timestamp"`
		GroupID             string             `json:"group_id"`
		Type                string             `json:"type"`
		RequestID           string             `json:"request_id"`
		Subject             string             `json:"subject"`
		InviteLink          string             `json:"invite_link"`
		JoinApprovalMode    string             `json:"join_approval_mode"`
		Description         string             `json:"description"`
		Errors              []werrors.Error    `json:"errors,omitempty"`
		Reason              string             `json:"reason"`
		AddedParticipants   []GroupParticipant `json:"added_participants"`
		RemovedParticipants []GroupParticipant `json:"removed_participants"`
		JoinRequestID       string             `json:"join_request_id"`
		WaID                string             `json:"wa_id"`
		InitiatedBy         string             `json:"initiated_by"`
	}
	GroupParticipant struct {
		WaID  string `json:"wa_id"`
		Input string `json:"input"`
	}
)

func (handler *Handler) OnGroupLifecycleUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupLifecycleUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

func (handler *Handler) SetGroupLifecycleUpdateHandler(
	h GroupLifecycleUpdateHandler,
) {
	handler.groupLifecycleUpdate = h
}

func (handler *Handler) OnGroupParticipantsUpdate(
	fn func(ctx context.Context, notificationContext *MessageNotificationContext, groups []*Group) error,
) {
	handler.groupParticipantsUpdate = MessageChangeValueHandlerFunc[Group](fn)
}

func (handler *Handler) SetGroupParticipantsUpdateHandler(
	h GroupParticipantsUpdateHandler,
) {
	handler.groupParticipantsUpdate = h
}
