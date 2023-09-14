/**
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

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

// Config is a struct that holds the configuration for the whatsapp client.
// It is used to create a new whatsapp client.
type Config struct {
	BaseURL           string
	Version           string
	AccessToken       string
	PhoneNumberID     string
	BusinessAccountID string
}

// ConfigReader is an interface that can be used to read the configuration
// from a file or any other source.
type ConfigReader interface {
	Read(ctx context.Context) (*Config, error)
}

// ConfigReaderFunc is a function that implements the ConfigReader interface.
type ConfigReaderFunc func(ctx context.Context) (*Config, error)

// Read implements the ConfigReader interface.
func (fn ConfigReaderFunc) Read(ctx context.Context) (*Config, error) {
	return fn(ctx)
}

type Client struct {
	Base   *BaseClient
	Config *Config
}

type ClientOption func(*Client)

// NewClient creates a new whatsapp client with the given options.
func NewClient(reader ConfigReader, options ...ClientOption) (*Client, error) {
	client := &Client{
		Base: &BaseClient{whttp.NewClient()},
	}

	config, err := reader.Read(context.Background())
	if err != nil || config == nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	client.Config = config

	if client.Config.BaseURL == "" {
		client.Config.BaseURL = BaseURL
	}

	if client.Config.Version == "" {
		client.Config.Version = LowestSupportedVersion
	}

	for i, option := range options {
		if option == nil {
			return nil, fmt.Errorf("option at index %d is nil", i)
		}
		option(client)
	}

	return client, nil
}

func WithBaseClient(base *BaseClient) ClientOption {
	return func(client *Client) {
		client.Base = base
	}
}

const MessageEndpoint = "messages"

type TextMessage struct {
	Message    string
	PreviewURL bool
}

// SendTextMessage sends a text message to a WhatsApp Business Account.
func (client *Client) SendTextMessage(ctx context.Context, recipient string,
	message *TextMessage,
) (*ResponseMessage, error) {
	text := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          textMessageType,
		Text: &models.Text{
			PreviewURL: message.PreviewURL,
			Body:       message.Message,
		},
	}

	req := &whttp.RequestContext{
		Name:          "send text message",
		BaseURL:       client.Config.BaseURL,
		ApiVersion:    client.Config.Version,
		PhoneNumberID: client.Config.PhoneNumberID,
		Bearer:        client.Config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.Base.SendMessage(ctx, req, text)
}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	req := &whttp.RequestContext{
		Name:          "mark message read",
		BaseURL:       client.Config.BaseURL,
		ApiVersion:    client.Config.Version,
		PhoneNumberID: client.Config.PhoneNumberID,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.Base.MarkMessageRead(ctx, req, messageID)
}

type ReactMessage struct {
	MessageID string
	Emoji     string
}

// React sends a reaction to a message.
// To send reaction messages, make a POST call to /PHONE_NUMBER_ID/messages and attach a message object
// with type=reaction. Then, add a reaction object.
//
// Sample request:
//
//	curl -X  POST \
//	 'https://graph.facebook.com/v15.0/FROM_PHONE_NUMBER_ID/messages' \
//	 -H 'Authorization: Bearer ACCESS_TOKEN' \
//	 -H 'Content-Type: application/json' \
//	 -d '{
//	  "messaging_product": "whatsapp",
//	  "recipient_type": "individual",
//	  "to": "PHONE_NUMBER",
//	  "type": "reaction",
//	  "reaction": {
//	    "message_id": "wamid.HBgLM...",
//	    "emoji": "\uD83D\uDE00"
//	  }
//	}'
//
// If the message you are reacting to is more than 30 days old, doesn't correspond to any message
// in the conversation, has been deleted, or is itself a reaction message, the reaction message will
// not be delivered, and you will receive a webhooks with the code 131009.
//
// A successful response includes an object with an identifier prefixed with wamid. Use the ID listed
// after wamid to track your message status.
//
// Example response:
//
//	{
//	  "messaging_product": "whatsapp",
//	  "contacts": [{
//	      "input": "PHONE_NUMBER",
//	      "wa_id": "WHATSAPP_ID",
//	    }]
//	  "messages": [{
//	      "id": "wamid.ID",
//	    }]
//	}
func (client *Client) React(ctx context.Context, recipient string, msg *ReactMessage) (*ResponseMessage, error) {
	reaction := &models.Message{
		Product: messagingProduct,
		To:      recipient,
		Type:    reactionMessageType,
		Reaction: &models.Reaction{
			MessageID: msg.MessageID,
			Emoji:     msg.Emoji,
		},
	}

	req := &whttp.RequestContext{
		Name:          "react to message",
		BaseURL:       client.Config.BaseURL,
		ApiVersion:    client.Config.Version,
		PhoneNumberID: client.Config.PhoneNumberID,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.Base.SendMessage(ctx, req, reaction)
}

// SendContacts sends a contact message. Contacts can be easily built using the models.NewContact() function.
func (client *Client) SendContacts(ctx context.Context, recipient string, contacts []*models.Contact) (
	*ResponseMessage, error,
) {
	contact := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          contactsMessageType,
		Contacts:      contacts,
	}

	req := &whttp.RequestContext{
		Name:          "send contacts",
		BaseURL:       client.Config.BaseURL,
		ApiVersion:    client.Config.Version,
		PhoneNumberID: client.Config.PhoneNumberID,
		Bearer:        client.Config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.Base.SendMessage(ctx, req, contact)
}

// SendLocationMessage sends a location message to a WhatsApp Business Account.
func (client *Client) SendLocationMessage(ctx context.Context, recipient string,
	message *models.Location,
) (*ResponseMessage, error) {
	location := &models.Message{
		Product:       messagingProduct,
		To:            recipient,
		RecipientType: individualRecipientType,
		Type:          locationMessageType,
		Location: &models.Location{
			Name:      message.Name,
			Address:   message.Address,
			Latitude:  message.Latitude,
			Longitude: message.Longitude,
		},
	}

	req := &whttp.RequestContext{
		Name:          "send location",
		BaseURL:       client.Config.BaseURL,
		ApiVersion:    client.Config.Version,
		PhoneNumberID: client.Config.PhoneNumberID,
		Bearer:        client.Config.AccessToken,
		Endpoints:     []string{MessageEndpoint},
	}

	return client.Base.SendMessage(ctx, req, location)
}

// BaseClient wraps the http client only and is used to make requests to the whatsapp api,
// It does not have the context. This is idealy for making requests to the whatsapp api for
// different users. The Client struct is used to make requests to the whatsapp api for a
// single user.
type BaseClient struct {
	*whttp.Client
}

func (base *BaseClient) SendMessage(ctx context.Context, req *whttp.RequestContext, msg *models.Message) (*ResponseMessage, error) {
	request := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: msg,
	}
	var resp ResponseMessage
	err := base.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.Name, err)
	}
	return &resp, nil
}

func (base *BaseClient) MarkMessageRead(ctx context.Context, req *whttp.RequestContext, messageID string) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: messagingProduct,
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	params := &whttp.Request{
		Context: req,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  req.Bearer,
		Payload: reqBody,
	}

	var success StatusResponse
	err := base.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("mark message read: %w", err)
	}

	return &success, nil
}
