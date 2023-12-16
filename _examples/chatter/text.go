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

package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/models"
	hooks "github.com/piusalfred/whatsapp/webhooks"
)

// OnTextMessageHook is a function that is called when a text message is received.
func (svc *Service) OnTextMessageHook(
	ctx context.Context,
	nctx *hooks.NotificationContext,
	mctx *hooks.MessageContext,
	text *hooks.Text,
) error {
	svc.Logger.LogAttrs(ctx, slog.LevelInfo, "received text message",
		slog.String("message_id", mctx.ID),
		slog.String("sender", mctx.From),
		slog.String("timestamp", mctx.Timestamp),
		slog.String("type", mctx.Type),
	)

	name := nctx.Contacts[0].Profile.Name
	message := &models.Text{
		PreviewURL: true,
		Body:       fmt.Sprintf("Hello %s, I am a bot, did you say %q?", name, text.Body),
	}
	reply, err := svc.whatsapp.Reply(ctx, &whatsapp.ReplyRequest{
		Recipient:   mctx.From,
		Context:     mctx.ID,
		MessageType: "text",
		Content:     message,
	})
	if err != nil {
		svc.Logger.LogAttrs(ctx, slog.LevelError, "failed to send reply",
			slog.String("message_id", mctx.ID),
			slog.String("sender", mctx.From),
			slog.String("timestamp", mctx.Timestamp),
			slog.String("type", mctx.Type),
		)

		return err
	}

	svc.Logger.LogAttrs(ctx, slog.LevelInfo, "sent reply",
		slog.Group("reply", reply))

	return nil
}
