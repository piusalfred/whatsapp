package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/piusalfred/whatsapp/webhooks"
)

func main() {
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	))

	handler := webhooks.NewHandler()

	handler.OnTextMessage(func(ctx context.Context, nctx *webhooks.MessageNotificationContext,
		mctx *webhooks.MessageInfo, text *webhooks.Text) error {
		logger.LogAttrs(ctx, slog.LevelInfo, "new text just dropped",
			slog.String("sender", mctx.From),
			slog.String("message", text.Body),
			slog.String("platform", nctx.MessagingProduct),
		)
		return nil
	})

	notificationJSON := `{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "<WHATSAPP_BUSINESS_ACCOUNT_ID>",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": {
              "display_phone_number": "<BUSINESS_DISPLAY_PHONE_NUMBER>",
              "phone_number_id": "<BUSINESS_PHONE_NUMBER_ID>"
            },
            "contacts": [
              {
                "profile": {
                  "name": "<WHATSAPP_USER_NAME>"
                },
                "wa_id": "<WHATSAPP_USER_ID>"
              }
            ],
            "messages": [
              {
                "from": "<WHATSAPP_USER_PHONE_NUMBER>",
                "id": "<WHATSAPP_MESSAGE_ID>",
                "timestamp": "<WEBHOOK_SENT_TIMESTAMP>",
                "text": {
                  "body": "<MESSAGE_BODY_TEXT>"
                },
                "type": "text"
              }
            ]
          },
          "field": "messages"
        }
      ]
    }
  ]
}`

	var notification webhooks.Notification

	if err := json.Unmarshal([]byte(notificationJSON), &notification); err != nil {
		logger.Error("failed to unmarshal notification", "error", err)
		return
	}

	if err := handler.HandleNotification(context.Background(), &notification); err != nil {
		logger.Error("failed to handle notification", "error", err)
	} else {
		logger.Info("notification handled successfully")
	}
}
