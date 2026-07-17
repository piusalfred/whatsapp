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
	"fmt"
)

type (
	// ErrorHandlerFunc adapts a bare function to the ErrorHandler interface.
	ErrorHandlerFunc func(ctx context.Context, err error) error

	// ErrorHandler is called when an error occurs during webhook processing.
	// A single webhook POST may carry multiple entries and changes.
	//
	// If Handle returns nil, the error is considered non-fatal and processing
	// continues to the next change/entry. If Handle returns a non-nil error,
	// processing stops immediately and an HTTP 500 is returned to WhatsApp
	// (which may trigger a retry of the entire payload).
	//
	// Design your ErrorHandler to decide per-error whether to continue or
	// abort — returning nil for transient errors keeps the pipeline moving.
	ErrorHandler interface {
		Handle(ctx context.Context, err error) error
	}
)

// handleSubHandlerError delegates to h if non-nil, otherwise returns err unchanged.
// Used by all webhook sub-handlers to avoid duplicating the error delegation logic.
func handleSubHandlerError(ctx context.Context, h ErrorHandler, err error) error {
	if h == nil {
		return err
	}
	if handlerErr := h.Handle(ctx, err); handlerErr != nil {
		return fmt.Errorf("error handler: %w", handlerErr)
	}
	return nil
}

func (fn ErrorHandlerFunc) Handle(ctx context.Context, err error) error {
	return fn(ctx, err)
}

type (
	FallbackHandlerFunc func(ctx context.Context, event NotificationEvent) error

	FallbackHandler interface {
		Handle(ctx context.Context, event NotificationEvent) error
	}
)

func (fn FallbackHandlerFunc) Handle(ctx context.Context, event NotificationEvent) error {
	return fn(ctx, event)
}
