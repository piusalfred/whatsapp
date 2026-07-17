/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package http

import (
	"context"

	"github.com/piusalfred/whatsapp/pkg/types"
)

const messageMetadataContextKey = "message-metadata-key"

// messageContextKey is a typed string to avoid collisions with other packages
// using [context.WithValue] on the same string value. Because the type is
// unexported, external packages cannot construct an equivalent key, which
// guarantees that only functions within this package can inject or retrieve
// this context entry.
type messageContextKey string

// InjectMessageMetadata stores [types.Metadata] in ctx so it travels with the
// request. Called internally by [RequestWithContext]; server-side handlers use
// [RetrieveMessageMetadata] to extract it. Typical use cases include correlation
// IDs and tenant identifiers.
func InjectMessageMetadata(ctx context.Context, metadata types.Metadata) context.Context {
	return context.WithValue(ctx, messageContextKey(messageMetadataContextKey), metadata)
}

// RetrieveMessageMetadata extracts [types.Metadata] that was previously stored
// via [InjectMessageMetadata]. Returns nil when no metadata was injected.
func RetrieveMessageMetadata(ctx context.Context) types.Metadata {
	metadata, ok := ctx.Value(messageContextKey(messageMetadataContextKey)).(types.Metadata)
	if !ok {
		return nil
	}

	return metadata
}
