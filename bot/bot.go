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
//	    whttp.WithSenderHTTPClient(http.DefaultClient),
//	    whttp.WithSenderTimeout(30*time.Second),
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
// [whttp.CoreSenderOption] functions customize the underlying HTTP transport:
//
//	whttp.WithSenderHTTPClient(customHTTPClient)
//	whttp.WithSenderRequestInterceptor(myRequestHook)
//	whttp.WithSenderResponseInterceptor(myResponseHook)
//	whttp.WithSenderTimeout(30 * time.Second)
//	whttp.WithSenderMaxBodyBytes(10 << 20)
//	whttp.WithSenderMaxHeaderBytes(1 << 20)
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
		*whttp.BaseClient[BaseRequest]
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
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.Sender = sender
}

// GetBotDetails retrieves comprehensive details about a WhatsApp Business Bot.
func (c *Client) GetBotDetails(ctx context.Context, request *Request) (*Bot, error) {
	return c.sender.GetBotDetails(ctx, c.config, request)
}

// Send translates a high-level [botRequest] into an HTTP transaction and decodes
// the response directly into a [Bot].
func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*Bot, error) {
	queryParams := map[string]string{}
	if request.Fields != "" {
		queryParams["fields"] = request.Fields
	}

	bld := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		Type(request.Type).
		Endpoints(conf.APIVersion, request.BotID)

	if len(queryParams) > 0 {
		bld = bld.QueryParams(queryParams)
	}

	req := whttp.BuildRequest(bld, request)

	var resp Bot
	decoder := whttp.ResponseDecoderJSON(&resp, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
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
