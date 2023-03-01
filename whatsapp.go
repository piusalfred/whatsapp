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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

const (
	BaseURL                   = "https://graph.facebook.com/"
	LowestSupportedVersion    = "v16.0"
	ContactBirthDayDateFormat = "2006-01-02" // YYYY-MM-DD
)

const (
	TextMessageType        = "text"
	ReactionMessageType    = "reaction"
	MediaMessageType       = "media"
	LocationMessageType    = "location"
	ContactMessageType     = "contact"
	InteractiveMessageType = "interactive"
)

const (
	MaxAudioSize         = 16 * 1024 * 1024  // 16 MB
	MaxDocSize           = 100 * 1024 * 1024 // 100 MB
	MaxImageSize         = 5 * 1024 * 1024   // 5 MB
	MaxVideoSize         = 16 * 1024 * 1024  // 16 MB
	MaxStickerSize       = 100 * 1024        // 100 KB
	UploadedMediaTTL     = 30 * 24 * time.Hour
	MediaDownloadLinkTTL = 5 * time.Minute
)

const (
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeImage    MediaType = "image"
	MediaTypeSticker  MediaType = "sticker"
	MediaTypeVideo    MediaType = "video"
)

type (

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, except reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	// Client includes the http client, base url, version, access token, phone number id, and whatsapp business account id.
	// which are used to make requests to the whatsapp api.
	// Example:
	// 	client := whatsapp.NewClient(
	// 		whatsapp.WithHTTPClient(http.DefaultClient),
	// 		whatsapp.WithBaseURL(whatsapp.BaseURL),
	// 		whatsapp.WithVersion(whatsapp.LowestSupportedVersion),
	// 		whatsapp.WithAccessToken("access_token"),
	// 		whatsapp.WithPhoneNumberID("phone_number_id"),
	// 		whatsapp.WithWhatsappBusinessAccountID("whatsapp_business_account_id"),
	// 	)
	//  // create a text message
	//  message := whatsapp.TextMessage{
	//  	Recipient: "<phone_number>",
	//  	Message:   "Hello World",
	//      PreviewURL: false,
	//  }
	// // send the text message
	//  _, err := client.SendTextMessage(context.Background(), message)
	//  if err != nil {
	//  	log.Fatal(err)
	//  }
	Client struct {
		rwm                       *sync.RWMutex
		HTTP                      *http.Client
		BaseURL                   string
		Version                   string
		AccessToken               string
		PhoneNumberID             string
		WhatsappBusinessAccountID string
	}

	ClientOption func(*Client)
)

func WithHTTPClient(http *http.Client) ClientOption {
	return func(client *Client) {
		client.HTTP = http
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		client.BaseURL = baseURL
	}
}

func WithVersion(version string) ClientOption {
	return func(client *Client) {
		client.Version = version
	}
}

func WithAccessToken(accessToken string) ClientOption {
	return func(client *Client) {
		client.AccessToken = accessToken
	}
}

func WithPhoneNumberID(phoneNumberID string) ClientOption {
	return func(client *Client) {
		client.PhoneNumberID = phoneNumberID
	}
}

func WithWhatsappBusinessAccountID(whatsappBusinessAccountID string) ClientOption {
	return func(client *Client) {
		client.WhatsappBusinessAccountID = whatsappBusinessAccountID
	}
}

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		rwm:                       &sync.RWMutex{},
		HTTP:                      http.DefaultClient,
		BaseURL:                   BaseURL,
		Version:                   "v16.0",
		AccessToken:               "",
		PhoneNumberID:             "",
		WhatsappBusinessAccountID: "",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) SetAccessToken(accessToken string) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.AccessToken = accessToken
}

func (c *Client) SetPhoneNumberID(phoneNumberID string) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.PhoneNumberID = phoneNumberID
}

func (c *Client) SetWhatsappBusinessAccountID(whatsappBusinessAccountID string) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.WhatsappBusinessAccountID = whatsappBusinessAccountID
}

type TextMessage struct {
	Message    string
	PreviewURL bool
}

// SendTextMessage sends a text message to a WhatsApp Business Account.
func (c *Client) SendTextMessage(ctx context.Context, recipient string, message *TextMessage) (*whttp.Response, error) {
	httpC := c.HTTP
	request := &SendTextRequest{
		BaseURL:       c.BaseURL,
		AccessToken:   c.AccessToken,
		PhoneNumberID: c.PhoneNumberID,
		ApiVersion:    c.Version,
		Recipient:     recipient,
		Message:       message.Message,
		PreviewURL:    message.PreviewURL,
	}
	resp, err := SendText(ctx, httpC, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	return resp, nil
}

// SendLocationMessage sends a location message to a WhatsApp Business Account.
func (c *Client) SendLocationMessage(ctx context.Context, recipient string, message *models.Location,
) (*whttp.Response, error) {
	httpC := c.HTTP
	request := &SendLocationRequest{
		BaseURL:       c.BaseURL,
		AccessToken:   c.AccessToken,
		PhoneNumberID: c.PhoneNumberID,
		ApiVersion:    c.Version,
		Recipient:     recipient,
		Name:          message.Name,
		Address:       message.Address,
		Latitude:      message.Latitude,
		Longitude:     message.Longitude,
	}
	resp, err := SendLocation(ctx, httpC, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send location message: %w", err)
	}

	return resp, nil
}
