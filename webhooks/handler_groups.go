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

package webhooks

import (
	"context"
)

func (handler *Handler) handleGroupWebhooks(ctx context.Context, change Change, entry Entry) error {
	switch change.Field {
	case ChangeFieldGroupLifecycleUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupLifecycleUpdate, change, entry, change.Value.Groups,
		)
	case ChangeFieldGroupParticipantsUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupParticipantsUpdate, change, entry, change.Value.Groups,
		)
	case ChangeFieldGroupSettingsUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupSettingsUpdate, change, entry, change.Value.Groups,
		)
	case ChangeFieldGroupStatusUpdate.String():
		return handleMessageChangeNotification(
			ctx, handler, handler.groupStatusUpdate, change, entry, change.Value.Groups,
		)
	}
	return nil
}
