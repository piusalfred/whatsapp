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

package api

import (
	"sync"
	"testing"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

// ---------------------------------------------------------------------------
// get* method correctness
// ---------------------------------------------------------------------------

func TestGetCalls_NormalInit(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()
	client := bc.getCalls()
	if client == nil {
		t.Fatal("getCalls() returned nil for healthy BaseClient")
	}
}

func TestGetMessage_NormalInit(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()
	client := bc.getMessage()
	if client == nil {
		t.Fatal("getMessage() returned nil for healthy BaseClient")
	}
}

func TestGetGroups_NormalInit(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()
	client := bc.getGroups()
	if client == nil {
		t.Fatal("getGroups() returned nil for healthy BaseClient")
	}
}

// ---------------------------------------------------------------------------
// Panic recovery: a panicking CoreSenderOption must not permanently poison
// the sub-client. The mutex is released by defer and the field stays nil,
// so the next call retries successfully.
// ---------------------------------------------------------------------------

func TestGetCalls_PanickingOption(t *testing.T) {
	t.Parallel()

	panickingOpt := whttp.CoreSenderOptionFunc(func(cfg *whttp.CoreSenderConfig) {
		panic("misconfigured calls option")
	})

	bc := NewBaseClient(panickingOpt)

	// First call: option panics. Mutex is released by defer. bc.calls stays nil.
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic from getCalls() with panicking option")
			}
		}()
		bc.getCalls()
	}()

	// bc.calls is still nil — the panic was not swallowed.
	if bc.calls != nil {
		t.Error("bc.calls should be nil after panicking init")
	}
}

func TestGetCalls_RetryAfterPanic(t *testing.T) {
	t.Parallel()

	var callCount int
	var mu sync.Mutex

	flakyOpt := whttp.CoreSenderOptionFunc(func(cfg *whttp.CoreSenderConfig) {
		mu.Lock()
		count := callCount
		callCount++
		mu.Unlock()

		if count == 0 {
			panic("transient option failure")
		}
	})

	bc := NewBaseClient(flakyOpt)

	// First call: option panics → bc.calls stays nil.
	func() {
		defer func() { recover() }()
		bc.getCalls()
	}()
	if bc.calls != nil {
		t.Error("bc.calls should be nil after first failed attempt")
	}

	// Second call: the mutex was released, nil check fails, init retries.
	// This time the flaky option succeeds (count=1).
	var panicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		bc.getCalls()
	}()

	if panicked {
		t.Fatal("getCalls() should have succeeded on retry after mutex-based init")
	}
	if bc.calls == nil {
		t.Fatal("bc.calls should be non-nil after successful retry")
	}
}

// ---------------------------------------------------------------------------
// Concurrent get* safety
// ---------------------------------------------------------------------------

func TestGetCalls_ConcurrentInit(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			c := bc.getCalls()
			if c == nil {
				t.Error("getCalls() returned nil under concurrent access")
			}
		})
	}
	wg.Wait()
}

func TestAllSubClients_ConcurrentInit(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()

	var wg sync.WaitGroup
	getters := []func() bool{
		func() bool { return bc.getCalls() != nil },
		func() bool { return bc.getUsers() != nil },
		func() bool { return bc.getQRCode() != nil },
		func() bool { return bc.getAuto() != nil },
		func() bool { return bc.getFlows() != nil },
		func() bool { return bc.getMedia() != nil },
		func() bool { return bc.getSettings() != nil },
		func() bool { return bc.getPhone() != nil },
		func() bool { return bc.getGroups() != nil },
		func() bool { return bc.getBiz() != nil },
		func() bool { return bc.getAnalytics() != nil },
		func() bool { return bc.getUploads() != nil },
		func() bool { return bc.getAuth() != nil },
		func() bool { return bc.getCallbacks() != nil },
		func() bool { return bc.getTemplates() != nil },
		func() bool { return bc.getMessage() != nil },
	}

	for _, fn := range getters {
		wg.Add(1)
		go func(get func() bool) {
			defer wg.Done()
			if !get() {
				t.Error("sub-client getter returned nil under concurrent access")
			}
		}(fn)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Idempotency: repeated calls return the same instance
// ---------------------------------------------------------------------------

func TestGetCalls_Idempotent(t *testing.T) {
	t.Parallel()

	bc := NewBaseClient()

	first := bc.getCalls()
	second := bc.getCalls()

	if first != second {
		t.Fatal("getCalls() returned different instances — expected same pointer")
	}
}
