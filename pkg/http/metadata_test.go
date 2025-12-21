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

package http_test

import (
	"context"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/types"
)

func TestInjectMessageMetadata(t *testing.T) {
	t.Parallel()

	t.Run("inject and retrieve metadata", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		metadata := types.Metadata(
			map[string]any{
				"key1":    "value1",
				"product": "whatsapp",
				"version": "v20",
				"name":    "cloud-api",
			},
		)

		ctx = whttp.InjectMessageMetadata(ctx, metadata)

		if diff := gcmp.Diff(whttp.RetrieveMessageMetadata(ctx), metadata); diff != "" {
			t.Errorf("InjectMessageMetadata() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("retrieve from context without metadata", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		metadata := whttp.RetrieveMessageMetadata(ctx)
		if metadata != nil {
			t.Errorf("expected nil metadata, got %v", metadata)
		}
	})
}
