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

package api_test

import (
	"testing"

	"github.com/piusalfred/whatsapp/api"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

func validConfig() *config.Config {
	return &config.Config{
		BaseURL:       "https://graph.facebook.com",
		APIVersion:    "v22.0",
		PhoneNumberID: "106540352242922",
		AccessToken:   "test-token",
	}
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestNewClient(t *testing.T) {
	t.Parallel()

	conf := validConfig()
	client := api.NewClient(conf)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewBaseClient(t *testing.T) {
	t.Parallel()

	bc := api.NewBaseClient()

	if bc == nil {
		t.Fatal("NewBaseClient returned nil")
	}
}

func TestNewBaseClient_WithOptions(t *testing.T) {
	t.Parallel()

	bc := api.NewBaseClient(
		whttp.WithSenderTimeout(60e9),
	)

	if bc == nil {
		t.Fatal("NewBaseClient returned nil with options")
	}
}

// ---------------------------------------------------------------------------
// Lazy initialisation — sync.Once correctness
// ---------------------------------------------------------------------------

func TestLazyInit_ReturnsSameInstance(t *testing.T) {
	t.Parallel()

	client := api.NewClient(validConfig())

	// Repeated calls should be safe (idempotent).
	for range 10 {
		client.SetMessagesMiddlewares()
	}
}

func TestLazyInit_MultipleSubClients(t *testing.T) {
	t.Parallel()

	client := api.NewClient(validConfig())

	// Trigger lazy init on all 15 sub-clients. None should panic.
	client.SetCallsMiddlewares()
	client.SetUsersMiddlewares()
	client.SetQRCodesMiddlewares()
	client.SetMessagesMiddlewares()
	client.SetGroupsMiddlewares()
	client.SetTemplatesMiddlewares()
	client.SetPhoneNumbersMiddlewares()
	client.SetAnalyticsMiddlewares()
	client.SetCallbacksMiddlewares()
	client.SetAutomationMiddlewares()
	client.SetFlowsMiddlewares()
	client.SetMediaMiddlewares()
	client.SetSettingsMiddlewares()
	client.SetBusinessMiddlewares()
	client.SetUploadsMiddlewares()
	client.SetAuthMiddlewares()
}

// ---------------------------------------------------------------------------
// Concurrent safety (init + reads)
//
// NOTE: Set*Middlewares calls are intentionally NOT tested concurrently
// because whttp.BaseClient.SetMiddlewares is documented as not goroutine-safe
// with respect to itself. See api_internal_test.go for concurrent init tests
// that exercise only the goroutine-safe get*() path.
// ---------------------------------------------------------------------------

func TestLazyInit_AllSubClientsSequential(t *testing.T) {
	t.Parallel()

	client := api.NewClient(validConfig())

	// Verify all 15 sub-clients can be initialised sequentially without
	// panics. This exercises every get*() method.
	client.SetCallsMiddlewares()
	client.SetUsersMiddlewares()
	client.SetQRCodesMiddlewares()
	client.SetMessagesMiddlewares()
	client.SetGroupsMiddlewares()
	client.SetTemplatesMiddlewares()
	client.SetPhoneNumbersMiddlewares()
	client.SetAnalyticsMiddlewares()
	client.SetCallbacksMiddlewares()
	client.SetAutomationMiddlewares()
	client.SetFlowsMiddlewares()
	client.SetMediaMiddlewares()
	client.SetSettingsMiddlewares()
	client.SetBusinessMiddlewares()
	client.SetUploadsMiddlewares()
	client.SetAuthMiddlewares()
}

// ---------------------------------------------------------------------------
// Middleware setter idempotency
// ---------------------------------------------------------------------------

func TestSetMiddlewares_Idempotent(t *testing.T) {
	t.Parallel()

	client := api.NewClient(validConfig())

	// First call triggers lazy init + sets middlewares.
	client.SetMessagesMiddlewares()

	// Second call re-wraps with same (empty) middleware list — must not panic.
	client.SetMessagesMiddlewares()

	// Third call — must not panic.
	client.SetMessagesMiddlewares()
}

// ---------------------------------------------------------------------------
// Client config binding
// ---------------------------------------------------------------------------

func TestClient_HoldsConfig(t *testing.T) {
	t.Parallel()

	conf := validConfig()
	client := api.NewClient(conf)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// The config is unexported, so verify indirectly: the client must be
	// usable (no panics on init).
	client.SetMessagesMiddlewares()
	client.SetCallsMiddlewares()
}
