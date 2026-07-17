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
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertNoError(t *testing.T, msg string, err error) {
	t.Helper()
	if err != nil {
		if msg != "" {
			t.Fatalf("%s: no error expected but found one, error=%v", msg, err)
		}

		t.Fatalf("no error expected but found one, error=%v", err)
	}
}

// AssertError checks that an error is not nil. If err is nil, the test fails.
func AssertError(t *testing.T, msg string, err error) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		}

		t.Fatalf("expected an error but got nil")
	}
}

// AssertErrorIs checks if err is or wraps target using errors.Is.
func AssertErrorIs(t *testing.T, msg string, err, target error) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		}

		t.Fatalf("expected an error but got nil")
		return
	}

	if !errors.Is(err, target) {
		if msg != "" {
			t.Fatalf("%s: error does not match target, got=%v, want=%v", msg, err, target)
		}

		t.Fatalf("error does not match target, got=%v, want=%v", err, target)
	}
}

// AssertErrorAs checks if err is or wraps an error of the same type as target using errors.As.
func AssertErrorAs(t *testing.T, msg string, err error, target any) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("%s: expected an error but got nil", msg)
		}

		t.Fatalf("expected an error but got nil")
		return
	}

	if !errors.As(err, target) {
		if msg != "" {
			t.Fatalf("%s: error type does not match target, got=%T", msg, err)
		}

		t.Fatalf("error type does not match target, got=%T", err)
	}
}

// AssertJSONMarshal marshals v to JSON and compares it against wantJSON using
// cmp.Diff, which handles map key ordering natively.
func AssertJSONMarshal[T any](t *testing.T, msg string, v T, wantJSON string) {
	t.Helper()

	gotBytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%s: failed to marshal value of type %T: %v", msg, v, err)
	}

	gotVal := parseJSONValue(t, gotBytes)
	wantVal := parseJSONValue(t, []byte(wantJSON))

	if diff := cmp.Diff(wantVal, gotVal); diff != "" {
		t.Errorf("%s: JSON mismatch (-want +got):\n%s", msg, diff)
	}
}

// AssertJSONUnmarshal unmarshals jsonStr into v.
func AssertJSONUnmarshal[T any](t *testing.T, msg, jsonStr string, v *T) {
	t.Helper()

	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		t.Fatalf("%s: failed to unmarshal into %T: %v\nJSON: %s", msg, v, err, jsonStr)
	}
}

// AssertJSONRoundTrip verifies that a struct marshals and unmarshals symmetrically.
// Accepts optional cmp.Option arguments for custom comparison (e.g., ignoring
// unexported fields).
func AssertJSONRoundTrip[T any](t *testing.T, msg string, v *T, opts ...cmp.Option) {
	t.Helper()

	gotBytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%s: failed to marshal %T: %v", msg, v, err)
	}

	var back T
	err = json.Unmarshal(gotBytes, &back)
	if err != nil {
		t.Fatalf("%s: failed to unmarshal %T: %v", msg, back, err)
	}

	backBytes, err := json.Marshal(&back)
	if err != nil {
		t.Fatalf("%s: failed to re-marshal %T: %v", msg, back, err)
	}

	gotVal := parseJSONValue(t, gotBytes)
	backVal := parseJSONValue(t, backBytes)

	if diff := cmp.Diff(gotVal, backVal); diff != "" {
		t.Errorf("%s: JSON round-trip mismatch (-original +roundtrip):\n%s", msg, diff)
	}

	if diff := cmp.Diff(*v, back, opts...); diff != "" {
		t.Errorf("%s: struct mismatch after round-trip (-original +roundtrip):\n%s", msg, diff)
	}
}

// parseJSONValue decodes raw JSON into an interface{} using UseNumber so that
// cmp.Diff can accurately compare floats and large integers without precision loss.
func parseJSONValue(t *testing.T, raw []byte) any {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("failed to decode JSON for comparison: %v\nRaw: %s", err, raw)
	}

	return v
}
