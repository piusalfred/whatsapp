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
