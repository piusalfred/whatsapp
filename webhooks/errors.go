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

package webhooks

import (
	"errors"
	"fmt"
)

var ErrInvalidSignature = fmt.Errorf("signature validation failed")

type (
	// FatalError wraps an error that that is fatal during the processing of a webhook notification.
	// it includes a description of the error.
	FatalError struct {
		Err  error
		Desc string
	}

	fatal interface {
		Fatal() bool
	}
)

// NewFatalError returns a new FatalError.
func NewFatalError(err error, desc string) *FatalError {
	return &FatalError{
		Err:  err,
		Desc: desc,
	}
}

func (e *FatalError) Unwrap() error {
	return e.Err
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Desc, e.Err.Error())
}

// IsFatalError returns true if the error is a FatalError or the error has implemented
// the fatal interface and Fatal() returns true.
func IsFatalError(err error) bool {
	var fe *FatalError
	if err != nil && errors.As(err, &fe) {
		return true
	}
	if err != nil {
		var f fatal
		if errors.As(err, &f) {
			return f.Fatal()
		}
	}

	return false
}
