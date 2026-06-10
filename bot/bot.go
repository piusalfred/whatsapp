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

// Package bot provides a client for retrieving WhatsApp Business Bot details
// via the WhatsApp Business Management API.
//
// The Bot Details API lets businesses retrieve comprehensive information about
// a WhatsApp Business Bot, including its prompts, commands, and welcome message
// configuration.
//
// # Getting Started
//
// Create a [Client] using [NewClient] with a [config.Config] and optional sender options:
//
//	conf := &config.Config{
//	    BaseURL:       "https://graph.facebook.com",
//	    APIVersion:    "v22.0",
//	    AccessToken:   "YOUR_ACCESS_TOKEN",
//	    PhoneNumberID: "YOUR_WABA_BOT_ID",
//	}
//
//	client := bot.NewClient(conf,
//	    bot.WithSenderHTTPClient(http.DefaultClient),
//	    bot.WithSenderTimeout(30*time.Second),
//	)
//
// # Retrieving Bot Details
//
//	resp, err := client.GetBotDetails(ctx, &bot.Request{
//	    BotID:  conf.PhoneNumberID,
//	    Fields: "id,prompts,commands,enable_welcome_message",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("ID:", resp.ID)
//	fmt.Println("Prompts:", resp.Prompts)
//	fmt.Println("Commands:", resp.Commands)
//	fmt.Println("Welcome Message Enabled:", resp.EnableWelcomeMessage)
//
// # Configuration Options
//
// [SenderOption] functions customize the underlying HTTP transport:
//
//	bot.WithSenderHTTPClient(customHTTPClient)
//	bot.WithSenderRequestInterceptor(myRequestHook)
//	bot.WithSenderResponseInterceptor(myResponseHook)
//	bot.WithSenderTimeout(30 * time.Second)
//	bot.WithSenderMaxBodyBytes(10 << 20)
//	bot.WithSenderMaxHeaderBytes(1 << 20)
//
// # Testing
//
// For unit tests, inject a mock sender via [Client.SetBaseClient]:
//
//	client := bot.NewClient(conf)
//	client.SetBaseClient(mockSender)
package bot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	// Request carries the parameters for retrieving bot details.
	Request struct {
		BotID  string
		Fields string // comma-separated list of fields
	}

	// Bot describes a WhatsApp Business Bot configuration and details.
	Bot struct {
		ID                   string    `json:"id"`
		Prompts              []string  `json:"prompts,omitempty"`
		Commands             []Command `json:"commands,omitempty"`
		EnableWelcomeMessage bool      `json:"enable_welcome_message,omitempty"`
	}

	// Command represents a bot command with its name and description.
	Command struct {
		CommandName        string `json:"command_name"`
		CommandDescription string `json:"command_description"`
	}

	// Client is a high-level client bound to a fixed [config.Config].
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	// BaseClient is the low-level HTTP executor for the Bot Details API. It
	// accepts a concrete [*config.Config] per request, making it suitable for
	// multi-tenant SaaS scenarios. For a fixed-configuration client, use [Client].
	BaseClient struct {
		sender whttp.Sender[BaseRequest]
	}

	// BaseRequest is an internal unified context data carrier mapping operation
	// metadata down to the HTTP executor.
	BaseRequest struct {
		Type   whttp.RequestType
		BotID  string
		Fields string
	}
)

// NewClient creates a high-level [Client] with a fixed configuration.
// Optional [SenderOption] functions tune the underlying HTTP transport.
func NewClient(conf *config.Config, options ...SenderOption) *Client {
	return &Client{
		sender: NewBaseClient(options...),
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.SetRequestSender(sender)
}

// GetBotDetails retrieves comprehensive details about a WhatsApp Business Bot.
func (c *Client) GetBotDetails(ctx context.Context, request *Request) (*Bot, error) {
	return c.sender.GetBotDetails(ctx, c.config, request)
}

// SenderOption configures the underlying [BaseClient] HTTP transport.
type SenderOption = whttp.CoreSenderOption

// WithSenderHTTPClient replaces the default [http.Client] used by the sender.
// A nil client is ignored.
func WithSenderHTTPClient(hc *http.Client) SenderOption {
	return whttp.WithSenderHTTPClient(hc)
}

// WithSenderRequestInterceptor registers a hook that inspects or mutates every
// outgoing [http.Request] before it is transmitted. A nil hook is ignored.
func WithSenderRequestInterceptor(hook whttp.RequestInterceptorFunc) SenderOption {
	return whttp.WithSenderRequestInterceptor(hook)
}

// WithSenderResponseInterceptor registers a hook that inspects or mutates every
// incoming [http.Response] before it is decoded. A nil hook is ignored.
func WithSenderResponseInterceptor(hook whttp.ResponseInterceptorFunc) SenderOption {
	return whttp.WithSenderResponseInterceptor(hook)
}

// WithSenderMaxBodyBytes sets the maximum allowable body size for request/response
// interceptors. Values less than or equal to zero are ignored.
func WithSenderMaxBodyBytes(n int64) SenderOption {
	return whttp.WithSenderMaxBodyBytes(n)
}

// WithSenderMaxHeaderBytes sets the maximum response header size. Values less than or
// equal to zero are ignored.
func WithSenderMaxHeaderBytes(n int64) SenderOption {
	return whttp.WithSenderMaxHeaderBytes(n)
}

// WithSenderTimeout sets the HTTP client timeout. Values less than or equal to zero
// are ignored.
func WithSenderTimeout(timeout time.Duration) SenderOption {
	return whttp.WithSenderTimeout(timeout)
}

// NewBaseClient creates a low-level [BaseClient] with optional [SenderOption]
// tuning. By default it builds a [whttp.CoreClient] with sensible defaults
// (30-second timeout, 10 MB body limit, 1 MB header limit).
func NewBaseClient(options ...SenderOption) *BaseClient {
	return &BaseClient{sender: whttp.NewCoreClient[BaseRequest](options...)}
}

// SetRequestSender replaces the internal sender, ignoring any HTTP
// configuration established by [NewBaseClient]. This is useful when you want
// to use a custom sender implementation or a mock during testing.
func (bc *BaseClient) SetRequestSender(sender whttp.Sender[BaseRequest]) {
	bc.sender = sender
}

// Send translates a high-level [botRequest] into an HTTP transaction and decodes
// the response directly into a [Bot].
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Bot, error) {
	queryParams := map[string]string{}
	if request.Fields != "" {
		queryParams["fields"] = request.Fields
	}

	bld := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		WithBearer(conf.AccessToken).
		WithAppSecret(conf.AppSecret, conf.SecureRequests).
		WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		WithRequestType(request.Type).
		WithEndpoints(conf.APIVersion, request.BotID)

	if len(queryParams) > 0 {
		bld = bld.WithQueryParams(queryParams)
	}

	req := whttp.BuildRequest(bld, request)

	var resp Bot
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &resp, nil
}

// GetBotDetails retrieves comprehensive details about a WhatsApp Business Bot.
func (bc *BaseClient) GetBotDetails(
	ctx context.Context,
	conf *config.Config,
	req *Request,
) (*Bot, error) {
	request := &BaseRequest{
		Type:   whttp.RequestTypeGetBotDetails,
		BotID:  req.BotID,
		Fields: req.Fields,
	}

	resp, err := bc.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("get bot details failed: %w", err)
	}

	return resp, nil
}
