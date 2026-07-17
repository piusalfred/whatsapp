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

package http_test

import (
	"context"
	"testing"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func TestBaseClient_SetMiddlewares_Ordering(t *testing.T) {
	t.Parallel()

	client := whttp.NewBaseClient[string]()
	order := make([]int, 0, 4)

	// Set the inner sender first — SetMiddlewares wraps whatever Sender is current.
	client.SetSender(
		whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				order = append(order, 4)
				return nil
			},
		),
	)

	mw1 := whttp.Middleware[string](func(next whttp.SenderFunc[string]) whttp.SenderFunc[string] {
		return whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				order = append(order, 1)
				return next(ctx, req, dec)
			},
		)
	})

	mw2 := whttp.Middleware[string](func(next whttp.SenderFunc[string]) whttp.SenderFunc[string] {
		return whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				order = append(order, 2)
				return next(ctx, req, dec)
			},
		)
	})

	mw3 := whttp.Middleware[string](func(next whttp.SenderFunc[string]) whttp.SenderFunc[string] {
		return whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				order = append(order, 3)
				return next(ctx, req, dec)
			},
		)
	})

	client.SetMiddlewares(mw1, mw2, mw3)

	_ = client.Send(context.Background(),
		whttp.MakeRequest[string]("GET", "http://example.com"),
		nil,
	)

	// middlewares[0] must run outermost: mw1 → mw2 → mw3 → inner
	want := []int{1, 2, 3, 4}
	if len(order) != len(want) {
		t.Fatalf("got %v, want %v", order, want)
	}
	for i := range order {
		if order[i] != want[i] {
			t.Fatalf("index %d: got %d, want %d (full: %v)", i, order[i], want[i], order)
		}
	}
}

func TestBaseClient_SetMiddlewares_SkipsNil(t *testing.T) {
	t.Parallel()

	client := whttp.NewBaseClient[string]()
	calls := 0

	client.SetSender(
		whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				calls++
				return nil
			},
		),
	)

	mw := whttp.Middleware[string](func(next whttp.SenderFunc[string]) whttp.SenderFunc[string] {
		return whttp.SenderFunc[string](
			func(ctx context.Context, req *whttp.Request[string], dec whttp.ResponseDecoder) error {
				calls++
				return next(ctx, req, dec)
			},
		)
	})

	client.SetMiddlewares(nil, mw, nil)

	_ = client.Send(context.Background(),
		whttp.MakeRequest[string]("GET", "http://example.com"),
		nil,
	)

	if calls != 2 {
		t.Fatalf("expected 2 calls (mw + inner), got %d", calls)
	}
}
