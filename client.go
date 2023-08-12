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

package whatsapp

import (
	whttp "github.com/piusalfred/whatsapp/http"
	"net/http"
	"sync"
)

type (

	// Client includes the http client, base url, ApiVersion, access token, phone number id,
	// and whatsapp business account id.
	// which are used to make requests to the whatsapp api.
	// Example:
	// 	client := whatsapp.NewClient(
	// 		whatsapp.WithHTTPClient(http.DefaultClient),
	// 		whatsapp.WithBaseURL(whatsapp.BaseURL),
	// 		whatsapp.WithVersion(whatsapp.LowestSupportedVersion),
	// 		whatsapp.WithAccessToken("access_token"),
	// 		whatsapp.WithPhoneNumberID("phone_number_id"),
	// 		whatsapp.WithBusinessAccountID("whatsapp_business_account_id"),
	// 	)
	//  // create a text message.
	//  message := whatsapp.TextMessage{
	//  	Recipient: "<phone_number>",
	//  	Message:   "Hello World",
	//      PreviewURL: false,
	//  }
	// // send the text message
	//  _, err := client.SendTextMessage(Ctx.Background(), message)
	//  if err != nil {
	//  	log.Fatal(err)
	//  }
	Client struct {
		rwm   *sync.RWMutex
		http  *whttp.Client
		debug bool
		Ctx   *Context
	}

	Context struct {
		BaseURL           string
		ApiVersion        string
		AccessToken       string
		PhoneNumberID     string
		BusinessAccountID string
	}
)

type ClientOption func(*Client)

func WithHTTPClient(http *http.Client) ClientOption {
	return func(client *Client) {
		client.http = whttp.NewClient(whttp.WithHTTPClient(http))
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		client.Ctx.BaseURL = baseURL
	}
}

func WithVersion(version string) ClientOption {
	return func(client *Client) {
		client.Ctx.ApiVersion = version
	}
}

func WithAccessToken(accessToken string) ClientOption {
	return func(client *Client) {
		client.Ctx.AccessToken = accessToken
	}
}

func WithPhoneNumberID(phoneNumberID string) ClientOption {
	return func(client *Client) {
		client.Ctx.PhoneNumberID = phoneNumberID
	}
}

func WithBusinessAccountID(whatsappBusinessAccountID string) ClientOption {
	return func(client *Client) {
		client.Ctx.BusinessAccountID = whatsappBusinessAccountID
	}
}

func WithContext(ctx *Context) ClientOption {
	return func(client *Client) {
		client.Ctx = ctx
	}
}

func WithDebug(debug bool) ClientOption {
	return func(client *Client) {
		client.debug = debug
	}
}

func WithResponseHooks(hooks ...whttp.ResponseHook) ClientOption {
	return func(client *Client) {
		if client.http != nil {
			client.http.SetResponseHooks(hooks...)
		}
	}
}

func WithRequestHooks(hooks ...whttp.RequestHook) ClientOption {
	return func(client *Client) {
		if client.http != nil {
			client.http.SetRequestHooks(hooks...)
		}
	}
}

// NewClient returns a new whatsapp client that can be used to send requests to the whatsapp api server.
// the options are applied in the order they are passed.
// Remember if there is an option to set http client, it should be the first option, before setting the
// request and response hooks.
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		rwm:   &sync.RWMutex{},
		http:  whttp.NewClient(),
		debug: true,
		Ctx: &Context{
			BaseURL:           BaseURL,
			ApiVersion:        "v16.0",
			AccessToken:       "",
			PhoneNumberID:     "",
			BusinessAccountID: "",
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (client *Client) Context() *Context {
	client.rwm.RLock()
	defer client.rwm.RUnlock()

	return client.Ctx
}

func (client *Client) SetAccessToken(accessToken string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.Ctx.AccessToken = accessToken
}

func (client *Client) SetPhoneNumberID(phoneNumberID string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.Ctx.PhoneNumberID = phoneNumberID
}

func (client *Client) SetBusinessAccountID(businessAccountID string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.Ctx.BusinessAccountID = businessAccountID
}
