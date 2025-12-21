package http

import (
	"context"

	"github.com/piusalfred/whatsapp/pkg/types"
)

const messageMetadataContextKey = "message-metadata-key"

type messageContextKey string

func InjectMessageMetadata(ctx context.Context, metadata types.Metadata) context.Context {
	return context.WithValue(ctx, messageContextKey(messageMetadataContextKey), metadata)
}

func RetrieveMessageMetadata(ctx context.Context) types.Metadata {
	metadata, ok := ctx.Value(messageContextKey(messageMetadataContextKey)).(types.Metadata)
	if !ok {
		return nil
	}

	return metadata
}
