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

// Package api provides a unified client for the WhatsApp Cloud API.
//
// Client holds a fixed *config.Config and lazily initializes 15 domain
// sub-clients (calls, users, qrcode, flows, media, settings, groups,
// business, analytics, uploads, auth, callbacks, messages) on first use.
// Each sub-client shares a single http.Client for connection pooling.
//
// BaseClient is the multi-tenant variant — it accepts *config.Config
// per call instead of holding a fixed one.
//
// Usage:
//
//	client := api.NewClient(conf, whttp.WithSenderTimeout(30*time.Second))
//	client.SetCallsMiddlewares(logger)
//	resp, err := client.CheckPermission(ctx, req)
//	resp, err := client.CreateQR(ctx, &qrcode.CreateRequest{PrefilledMessage: "Hi"})
package api

import (
	"sync"

	"github.com/piusalfred/whatsapp/auth"
	"github.com/piusalfred/whatsapp/business"
	"github.com/piusalfred/whatsapp/business/analytics"
	"github.com/piusalfred/whatsapp/calls"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/conversation/automation"
	"github.com/piusalfred/whatsapp/flow"
	"github.com/piusalfred/whatsapp/groups"
	"github.com/piusalfred/whatsapp/media"
	"github.com/piusalfred/whatsapp/message"
	"github.com/piusalfred/whatsapp/phonenumber"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/qrcode"
	"github.com/piusalfred/whatsapp/settings"
	"github.com/piusalfred/whatsapp/templates"
	"github.com/piusalfred/whatsapp/uploads"
	"github.com/piusalfred/whatsapp/user"
	"github.com/piusalfred/whatsapp/webhooks/callbacks"
)

// Client wraps BaseClient with a fixed configuration.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// BaseClient is the multi-tenant layer — pass a *config.Config per call. It
// lazily initialises 15 domain sub-clients on first access and protects each
// with a [sync.Mutex]. If a [whttp.CoreSenderOption] panics inside the
// constructor, the sub-client field remains nil, the mutex is released, and
// the next call retries initialisation — preventing permanent poisoning.
//
// All sub-clients share the same [whttp.CoreSenderOption] values, ensuring
// consistent HTTP behaviour (timeouts, interceptors, limits) across all domains.
//
// Important: [Client.SetCallsMiddlewares] and similar per-sub-client middleware
// setters must complete before any API calls are made. They trigger lazy
// initialisation (goroutine-safe) but then call [whttp.BaseClient.SetMiddlewares]
// which is not goroutine-safe with respect to in-flight requests.
type BaseClient struct {
	calls       *calls.BaseClient
	callsMu     sync.Mutex
	users       *user.BlockBaseClient
	usersMu     sync.Mutex
	qrCode      *qrcode.BaseClient
	qrCodeMu    sync.Mutex
	auto        *automation.BaseClient
	autoMu      sync.Mutex
	flows       *flow.BaseClient
	flowsMu     sync.Mutex
	media       *media.BaseClient
	mediaMu     sync.Mutex
	settings    *settings.BaseClient
	settingsMu  sync.Mutex
	phone       *phonenumber.BaseClient
	phoneMu     sync.Mutex
	groups      *groups.BaseClient
	groupsMu    sync.Mutex
	biz         *business.BaseClient
	bizMu       sync.Mutex
	analytics   *analytics.BaseClient
	analyticsMu sync.Mutex
	uploads     *uploads.BaseClient
	uploadsMu   sync.Mutex
	auth        *auth.BaseClient
	authMu      sync.Mutex
	callbacks   *callbacks.BaseClient
	callbacksMu sync.Mutex
	message     *message.BaseClient
	messageMu   sync.Mutex
	templates   *templates.BaseClient
	templatesMu sync.Mutex

	opts []whttp.CoreSenderOption
}

// NewClient creates a Client with the given fixed configuration.
func NewClient(conf *config.Config, opts ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: NewBaseClient(opts...),
		config: conf,
	}
}

// NewBaseClient creates a BaseClient with the given sender options.
// The underlying HTTP client defaults are provided by [whttp.NewBaseClient]
// (30 s timeout, 10 MB body limit). Override via [whttp.WithSenderHTTPClient],
// [whttp.WithSenderTimeout], etc.
func NewBaseClient(opts ...whttp.CoreSenderOption) *BaseClient {
	return &BaseClient{opts: opts}
}

// Each get* method lazy-initialises its sub-client under a [sync.Mutex].
// If [whttp.NewBaseClient] panics (e.g. a misconfigured option), the mutex
// is released via defer and the sub-client field stays nil, so the next
// call retries instead of permanently returning a poisoned nil value.

func (bc *BaseClient) getCalls() *calls.BaseClient {
	bc.callsMu.Lock()
	defer bc.callsMu.Unlock()
	if bc.calls != nil {
		return bc.calls
	}
	bc.calls = &calls.BaseClient{BaseClient: *whttp.NewBaseClient[calls.BaseRequest](bc.opts...)}
	return bc.calls
}

func (bc *BaseClient) getUsers() *user.BlockBaseClient {
	bc.usersMu.Lock()
	defer bc.usersMu.Unlock()
	if bc.users != nil {
		return bc.users
	}
	bc.users = &user.BlockBaseClient{BaseClient: *whttp.NewBaseClient[user.BlockBaseRequest](bc.opts...)}
	return bc.users
}

func (bc *BaseClient) getQRCode() *qrcode.BaseClient {
	bc.qrCodeMu.Lock()
	defer bc.qrCodeMu.Unlock()
	if bc.qrCode != nil {
		return bc.qrCode
	}
	bc.qrCode = &qrcode.BaseClient{BaseClient: *whttp.NewBaseClient[qrcode.BaseRequest](bc.opts...)}
	return bc.qrCode
}

func (bc *BaseClient) getAuto() *automation.BaseClient {
	bc.autoMu.Lock()
	defer bc.autoMu.Unlock()
	if bc.auto != nil {
		return bc.auto
	}
	bc.auto = &automation.BaseClient{BaseClient: *whttp.NewBaseClient[automation.BaseRequest](bc.opts...)}
	return bc.auto
}

func (bc *BaseClient) getFlows() *flow.BaseClient {
	bc.flowsMu.Lock()
	defer bc.flowsMu.Unlock()
	if bc.flows != nil {
		return bc.flows
	}
	bc.flows = &flow.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.flows
}

func (bc *BaseClient) getMedia() *media.BaseClient {
	bc.mediaMu.Lock()
	defer bc.mediaMu.Unlock()
	if bc.media != nil {
		return bc.media
	}
	bc.media = &media.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.media
}

func (bc *BaseClient) getSettings() *settings.BaseClient {
	bc.settingsMu.Lock()
	defer bc.settingsMu.Unlock()
	if bc.settings != nil {
		return bc.settings
	}
	bc.settings = &settings.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.settings
}

func (bc *BaseClient) getPhone() *phonenumber.BaseClient {
	bc.phoneMu.Lock()
	defer bc.phoneMu.Unlock()
	if bc.phone != nil {
		return bc.phone
	}
	bc.phone = &phonenumber.BaseClient{BaseClient: *whttp.NewBaseClient[phonenumber.BaseRequest](bc.opts...)}
	return bc.phone
}

func (bc *BaseClient) getGroups() *groups.BaseClient {
	bc.groupsMu.Lock()
	defer bc.groupsMu.Unlock()
	if bc.groups != nil {
		return bc.groups
	}
	bc.groups = &groups.BaseClient{BaseClient: *whttp.NewBaseClient[groups.BaseRequest](bc.opts...)}
	return bc.groups
}

func (bc *BaseClient) getBiz() *business.BaseClient {
	bc.bizMu.Lock()
	defer bc.bizMu.Unlock()
	if bc.biz != nil {
		return bc.biz
	}
	bc.biz = &business.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.biz
}

func (bc *BaseClient) getAnalytics() *analytics.BaseClient {
	bc.analyticsMu.Lock()
	defer bc.analyticsMu.Unlock()
	if bc.analytics != nil {
		return bc.analytics
	}
	bc.analytics = &analytics.BaseClient{BaseClient: *whttp.NewBaseClient[analytics.BaseRequest](bc.opts...)}
	return bc.analytics
}

func (bc *BaseClient) getUploads() *uploads.BaseClient {
	bc.uploadsMu.Lock()
	defer bc.uploadsMu.Unlock()
	if bc.uploads != nil {
		return bc.uploads
	}
	bc.uploads = &uploads.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.uploads
}

func (bc *BaseClient) getAuth() *auth.BaseClient {
	bc.authMu.Lock()
	defer bc.authMu.Unlock()
	if bc.auth != nil {
		return bc.auth
	}
	bc.auth = &auth.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	return bc.auth
}

func (bc *BaseClient) getCallbacks() *callbacks.BaseClient {
	bc.callbacksMu.Lock()
	defer bc.callbacksMu.Unlock()
	if bc.callbacks != nil {
		return bc.callbacks
	}
	bc.callbacks = &callbacks.BaseClient{BaseClient: *whttp.NewBaseClient[callbacks.BaseRequest](bc.opts...)}
	return bc.callbacks
}

func (bc *BaseClient) getTemplates() *templates.BaseClient {
	bc.templatesMu.Lock()
	defer bc.templatesMu.Unlock()
	if bc.templates != nil {
		return bc.templates
	}
	bc.templates = &templates.BaseClient{BaseClient: *whttp.NewBaseClient[templates.BaseRequest](bc.opts...)}
	return bc.templates
}

func (bc *BaseClient) getMessage() *message.BaseClient {
	bc.messageMu.Lock()
	defer bc.messageMu.Unlock()
	if bc.message != nil {
		return bc.message
	}
	bc.message = &message.BaseClient{BaseClient: *whttp.NewBaseClient[message.BaseRequest](bc.opts...)}
	return bc.message
}

func (c *Client) SetCallsMiddlewares(mws ...whttp.Middleware[calls.BaseRequest]) {
	c.sender.getCalls().SetMiddlewares(mws...)
}

func (c *Client) SetUsersMiddlewares(mws ...whttp.Middleware[user.BlockBaseRequest]) {
	c.sender.getUsers().SetMiddlewares(mws...)
}

func (c *Client) SetQRCodesMiddlewares(mws ...whttp.Middleware[qrcode.BaseRequest]) {
	c.sender.getQRCode().SetMiddlewares(mws...)
}

func (c *Client) SetAutomationMiddlewares(mws ...whttp.Middleware[automation.BaseRequest]) {
	c.sender.getAuto().SetMiddlewares(mws...)
}

func (c *Client) SetFlowsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getFlows().SetMiddlewares(mws...)
}

func (c *Client) SetMediaMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getMedia().SetMiddlewares(mws...)
}

func (c *Client) SetSettingsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getSettings().SetMiddlewares(mws...)
}

func (c *Client) SetPhoneNumbersMiddlewares(mws ...whttp.Middleware[phonenumber.BaseRequest]) {
	c.sender.getPhone().SetMiddlewares(mws...)
}

func (c *Client) SetGroupsMiddlewares(mws ...whttp.Middleware[groups.BaseRequest]) {
	c.sender.getGroups().SetMiddlewares(mws...)
}

func (c *Client) SetBusinessMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getBiz().SetMiddlewares(mws...)
}

func (c *Client) SetAnalyticsMiddlewares(mws ...whttp.Middleware[analytics.BaseRequest]) {
	c.sender.getAnalytics().SetMiddlewares(mws...)
}

func (c *Client) SetUploadsMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getUploads().SetMiddlewares(mws...)
}

func (c *Client) SetAuthMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.getAuth().SetMiddlewares(mws...)
}

func (c *Client) SetCallbacksMiddlewares(mws ...whttp.Middleware[callbacks.BaseRequest]) {
	c.sender.getCallbacks().SetMiddlewares(mws...)
}

func (c *Client) SetMessagesMiddlewares(mws ...whttp.Middleware[message.BaseRequest]) {
	c.sender.getMessage().SetMiddlewares(mws...)
}

// SetTemplatesMiddlewares configures middlewares for the templates sub-client.
func (c *Client) SetTemplatesMiddlewares(mws ...whttp.Middleware[templates.BaseRequest]) {
	c.sender.getTemplates().SetMiddlewares(mws...)
}

// CloseIdleConnections drains idle HTTP connections for all initialized
// sub-clients. Call during graceful shutdown to prevent socket leaks.
// Uninitialized sub-clients are safely skipped.
func (bc *BaseClient) CloseIdleConnections() {
	// Check each sub-client pointer individually — a nil pointer converted to
	// an interface still produces a non-nil interface, so a loop would call
	// CloseIdleConnections on nil receivers and panic.
	if bc.calls != nil {
		bc.calls.CloseIdleConnections()
	}
	if bc.users != nil {
		bc.users.CloseIdleConnections()
	}
	if bc.qrCode != nil {
		bc.qrCode.CloseIdleConnections()
	}
	if bc.auto != nil {
		bc.auto.CloseIdleConnections()
	}
	if bc.flows != nil {
		bc.flows.CloseIdleConnections()
	}
	if bc.media != nil {
		bc.media.CloseIdleConnections()
	}
	if bc.settings != nil {
		bc.settings.CloseIdleConnections()
	}
	if bc.phone != nil {
		bc.phone.CloseIdleConnections()
	}
	if bc.groups != nil {
		bc.groups.CloseIdleConnections()
	}
	if bc.biz != nil {
		bc.biz.CloseIdleConnections()
	}
	if bc.analytics != nil {
		bc.analytics.CloseIdleConnections()
	}
	if bc.uploads != nil {
		bc.uploads.CloseIdleConnections()
	}
	if bc.auth != nil {
		bc.auth.CloseIdleConnections()
	}
	if bc.callbacks != nil {
		bc.callbacks.CloseIdleConnections()
	}
	if bc.message != nil {
		bc.message.CloseIdleConnections()
	}
	if bc.templates != nil {
		bc.templates.CloseIdleConnections()
	}
}

// CloseIdleConnections drains idle HTTP connections across all sub-clients.
func (c *Client) CloseIdleConnections() {
	c.sender.CloseIdleConnections()
}
