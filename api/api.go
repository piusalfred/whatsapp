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
	"net/http"
	"sync"
	"time"

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

// BaseClient is the multi-tenant layer. Pass a *config.Config per call.
type BaseClient struct {
	calls         *calls.BaseClient
	callsOnce     sync.Once
	users         *user.BlockBaseClient
	usersOnce     sync.Once
	qrCode        *qrcode.BaseClient
	qrCodeOnce    sync.Once
	auto          *automation.BaseClient
	autoOnce      sync.Once
	flows         *flow.BaseClient
	flowsOnce     sync.Once
	media         *media.BaseClient
	mediaOnce     sync.Once
	settings      *settings.BaseClient
	settingsOnce  sync.Once
	phone         *phonenumber.BaseClient
	phoneOnce     sync.Once
	groups        *groups.BaseClient
	groupsOnce    sync.Once
	biz           *business.BaseClient
	bizOnce       sync.Once
	analytics     *analytics.BaseClient
	analyticsOnce sync.Once
	uploads       *uploads.BaseClient
	uploadsOnce   sync.Once
	auth          *auth.BaseClient
	authOnce      sync.Once
	callbacks     *callbacks.BaseClient
	callbacksOnce sync.Once
	message       *message.BaseClient
	messageOnce   sync.Once
	templates     *templates.BaseClient
	templatesOnce sync.Once

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
func NewBaseClient(opts ...whttp.CoreSenderOption) *BaseClient {
	sharedOpts := append([]whttp.CoreSenderOption{
		whttp.WithSenderHTTPClient(&http.Client{
			Timeout: 30 * time.Second, //nolint:mnd // default HTTP client timeout
		}),
	}, opts...)
	return &BaseClient{opts: sharedOpts}
}

func (bc *BaseClient) getCalls() *calls.BaseClient {
	bc.callsOnce.Do(func() {
		bc.calls = &calls.BaseClient{BaseClient: *whttp.NewBaseClient[calls.BaseRequest](bc.opts...)}
	})
	return bc.calls
}

func (bc *BaseClient) getUsers() *user.BlockBaseClient {
	bc.usersOnce.Do(func() {
		bc.users = &user.BlockBaseClient{BaseClient: *whttp.NewBaseClient[user.BlockBaseRequest](bc.opts...)}
	})
	return bc.users
}

func (bc *BaseClient) getQRCode() *qrcode.BaseClient {
	bc.qrCodeOnce.Do(func() {
		bc.qrCode = &qrcode.BaseClient{BaseClient: *whttp.NewBaseClient[qrcode.BaseRequest](bc.opts...)}
	})
	return bc.qrCode
}

func (bc *BaseClient) getAuto() *automation.BaseClient {
	bc.autoOnce.Do(func() {
		bc.auto = &automation.BaseClient{BaseClient: *whttp.NewBaseClient[automation.BaseRequest](bc.opts...)}
	})
	return bc.auto
}

func (bc *BaseClient) getFlows() *flow.BaseClient {
	bc.flowsOnce.Do(func() {
		bc.flows = &flow.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.flows
}

func (bc *BaseClient) getMedia() *media.BaseClient {
	bc.mediaOnce.Do(func() {
		bc.media = &media.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.media
}

func (bc *BaseClient) getSettings() *settings.BaseClient {
	bc.settingsOnce.Do(func() {
		bc.settings = &settings.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.settings
}

func (bc *BaseClient) getPhone() *phonenumber.BaseClient {
	bc.phoneOnce.Do(func() {
		bc.phone = &phonenumber.BaseClient{BaseClient: *whttp.NewBaseClient[phonenumber.BaseRequest](bc.opts...)}
	})
	return bc.phone
}

func (bc *BaseClient) getGroups() *groups.BaseClient {
	bc.groupsOnce.Do(func() {
		bc.groups = &groups.BaseClient{BaseClient: *whttp.NewBaseClient[groups.BaseRequest](bc.opts...)}
	})
	return bc.groups
}

func (bc *BaseClient) getBiz() *business.BaseClient {
	bc.bizOnce.Do(func() {
		bc.biz = &business.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.biz
}

func (bc *BaseClient) getAnalytics() *analytics.BaseClient {
	bc.analyticsOnce.Do(func() {
		bc.analytics = &analytics.BaseClient{BaseClient: *whttp.NewBaseClient[analytics.BaseRequest](bc.opts...)}
	})
	return bc.analytics
}

func (bc *BaseClient) getUploads() *uploads.BaseClient {
	bc.uploadsOnce.Do(func() {
		bc.uploads = &uploads.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.uploads
}

func (bc *BaseClient) getAuth() *auth.BaseClient {
	bc.authOnce.Do(func() {
		bc.auth = &auth.BaseClient{BaseClient: *whttp.NewBaseClient[any](bc.opts...)}
	})
	return bc.auth
}

func (bc *BaseClient) getCallbacks() *callbacks.BaseClient {
	bc.callbacksOnce.Do(func() {
		bc.callbacks = &callbacks.BaseClient{BaseClient: *whttp.NewBaseClient[callbacks.BaseRequest](bc.opts...)}
	})
	return bc.callbacks
}

func (bc *BaseClient) getTemplates() *templates.BaseClient {
	bc.templatesOnce.Do(func() {
		bc.templates = &templates.BaseClient{BaseClient: *whttp.NewBaseClient[templates.BaseRequest](bc.opts...)}
	})
	return bc.templates
}

func (bc *BaseClient) getMessage() *message.BaseClient {
	bc.messageOnce.Do(func() {
		bc.message = &message.BaseClient{BaseClient: *whttp.NewBaseClient[message.BaseRequest](bc.opts...)}
	})
	return bc.message
}

func (c *Client) SetCallsMiddlewares(mws ...whttp.Middleware[calls.BaseRequest]) {
	c.sender.getCalls().SetMiddlewares(mws...)
}

func (c *Client) SetUsersBlockMiddlewares(mws ...whttp.Middleware[user.BlockBaseRequest]) {
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
