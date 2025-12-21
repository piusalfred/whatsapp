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

package test

import (
	"errors"
	"testing"
)

func AssertNoError(t *testing.T, msg string, err error) {
	t.Helper()
	if err != nil {
		if msg != "" {
			t.Fatalf("%s: no error expected but found one, error=%v", msg, err)
		} else {
			t.Fatalf("no error expected but found one, error=%v", err)
		}
	}
}

// AssertError checks that an error is not nil. If err is nil, the test fails.
func AssertError(t *testing.T, msg string, err error) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		} else {
			t.Fatalf("expected an error but got nil")
		}
	}
}

// AssertErrorIs checks if err is or wraps target using errors.Is.
// If the errors don't match, the test fails.
func AssertErrorIs(t *testing.T, msg string, err, target error) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		} else {
			t.Fatalf("expected an error but got nil")
		}
		return
	}

	if !errors.Is(err, target) {
		if msg != "" {
			t.Fatalf("%s: error does not match target, got=%v, want=%v", msg, err, target)
		} else {
			t.Fatalf("error does not match target, got=%v, want=%v", err, target)
		}
	}
}

// AssertErrorAs checks if err is or wraps an error of the same type as target using errors.As.
// target must be a pointer to an error type. If the errors don't match, the test fails.
func AssertErrorAs(t *testing.T, msg string, err error, target any) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		} else {
			t.Fatalf("expected an error but got nil")
		}
		return
	}

	if !errors.As(err, &target) {
		if msg != "" {
			t.Fatalf("%s: error type does not match target, got=%T, want=%T", msg, err, target)
		} else {
			t.Fatalf("error type does not match target, got=%T, want=%T", err, target)
		}
	}
}
