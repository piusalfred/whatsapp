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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/piusalfred/whatsapp/qrcodes"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
)

var ErrNilRequest = errors.New("nil request")

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
	// 		whatsapp.WithBaseURL(whatsapp.baseURL),
	// 		whatsapp.WithVersion(whatsapp.LowestSupportedVersion),
	// 		whatsapp.WithAccessToken("access_token"),
	// 		whatsapp.WithPhoneNumberID("phone_number_id"),
	// 		whatsapp.WithBusinessAccountID("whatsapp_business_account_id"),
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
		rwm               *sync.RWMutex
		http              *http.Client
		baseURL           string
		version           string
		accessToken       string
		phoneNumberID     string
		businessAccountID string
	}

	ClientOption func(*Client)
)

func WithHTTPClient(http *http.Client) ClientOption {
	return func(client *Client) {
		client.http = http
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

func WithVersion(version string) ClientOption {
	return func(client *Client) {
		client.version = version
	}
}

func WithAccessToken(accessToken string) ClientOption {
	return func(client *Client) {
		client.accessToken = accessToken
	}
}

func WithPhoneNumberID(phoneNumberID string) ClientOption {
	return func(client *Client) {
		client.phoneNumberID = phoneNumberID
	}
}

func WithBusinessAccountID(whatsappBusinessAccountID string) ClientOption {
	return func(client *Client) {
		client.businessAccountID = whatsappBusinessAccountID
	}
}

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		rwm:               &sync.RWMutex{},
		http:              http.DefaultClient,
		baseURL:           BaseURL,
		version:           "v16.0",
		accessToken:       "",
		phoneNumberID:     "",
		businessAccountID: "",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (client *Client) AccessToken() string {
	client.rwm.RLock()
	defer client.rwm.RUnlock()

	return client.accessToken
}

func (client *Client) PhoneNumberID() string {
	client.rwm.RLock()
	defer client.rwm.RUnlock()

	return client.phoneNumberID
}

func (client *Client) BusinessAccountID() string {
	client.rwm.RLock()
	defer client.rwm.RUnlock()

	return client.businessAccountID
}

func (client *Client) SetAccessToken(accessToken string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.accessToken = accessToken
}

func (client *Client) SetPhoneNumberID(phoneNumberID string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.phoneNumberID = phoneNumberID
}

func (client *Client) SetWhatsappBusinessAccountID(whatsappBusinessAccountID string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.businessAccountID = whatsappBusinessAccountID
}

type TextMessage struct {
	Message    string
	PreviewURL bool
}

// SendTextMessage sends a text message to a WhatsApp Business Account.
func (client *Client) SendTextMessage(ctx context.Context, recipient string,
	message *TextMessage,
) (*whttp.Response, error) {
	httpC := client.http
	request := &SendTextRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.accessToken,
		PhoneNumberID: client.phoneNumberID,
		ApiVersion:    client.version,
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
func (client *Client) SendLocationMessage(ctx context.Context, recipient string,
	message *models.Location,
) (*whttp.Response, error) {
	request := &SendLocationRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.accessToken,
		PhoneNumberID: client.phoneNumberID,
		ApiVersion:    client.version,
		Recipient:     recipient,
		Name:          message.Name,
		Address:       message.Address,
		Latitude:      message.Latitude,
		Longitude:     message.Longitude,
	}

	resp, err := SendLocation(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send location message: %w", err)
	}

	return resp, nil
}

type ReactMessage struct {
	MessageID string
	Emoji     string
}

func (client *Client) React(ctx context.Context, recipient string, req *ReactMessage) (*whttp.Response, error) {
	request := &ReactRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.AccessToken(),
		PhoneNumberID: client.PhoneNumberID(),
		ApiVersion:    client.version,
		Recipient:     recipient,
		MessageID:     req.MessageID,
		Emoji:         req.Emoji,
	}

	resp, err := React(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("react: %w", err)
	}

	return resp, nil
}

type MediaMessage struct {
	Type      MediaType
	MediaID   string
	MediaLink string
	Caption   string
	Filename  string
	Provider  string
}

// SendMedia sends a media message to the recipient.
func (client *Client) SendMedia(ctx context.Context, recipient string, req *MediaMessage,
	cacheOptions *CacheOptions,
) (*whttp.Response, error) {
	request := &SendMediaRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.AccessToken(),
		PhoneNumberID: client.PhoneNumberID(),
		ApiVersion:    client.version,
		Recipient:     recipient,
		Type:          req.Type,
		MediaID:       req.MediaID,
		MediaLink:     req.MediaLink,
		Caption:       req.Caption,
		Filename:      req.Filename,
		Provider:      req.Provider,
		CacheOptions:  cacheOptions,
	}

	resp, err := SendMedia(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("client send media: %w", err)
	}

	return resp, nil
}

// ReplyMessage is a message that is sent as a reply to a previous message. The previous message's ID
// is needed and is set as Context in ReplyRequest.
// Content is the message content. It can be a Text, Location, Media, Template, or Contact.
type ReplyMessage struct {
	Context string
	Type    MessageType
	Content any
}

func (client *Client) Reply(ctx context.Context, recipient string, req *ReplyMessage) (*whttp.Response, error) {
	request := &ReplyRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.AccessToken(),
		PhoneNumberID: client.PhoneNumberID(),
		ApiVersion:    client.version,
		Recipient:     recipient,
		Context:       req.Context,
		MessageType:   req.Type,
		Content:       req.Content,
	}

	resp, err := Reply(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("client reply: %w", err)
	}

	return resp, nil
}

func (client *Client) SendContacts(ctx context.Context, recipient string,
	contacts *models.Contacts,
) (*whttp.Response, error) {
	contact := &models.Message{
		Product:       "whatsapp",
		To:            recipient,
		RecipientType: "individual",
		Type:          "contact",
		Contacts:      contacts,
	}
	payload, err := json.Marshal(contact)
	if err != nil {
		return nil, fmt.Errorf("client send contacts: marshal contact: %w", err)
	}

	baseURL := client.baseURL
	apiVersion := client.version
	phoneNumberID := client.PhoneNumberID()
	accessToken := client.AccessToken()

	params := &whttp.RequestParams{
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  accessToken,
	}

	reqURL, err := whttp.CreateRequestURL(baseURL, apiVersion, phoneNumberID, "messages")
	if err != nil {
		return nil, fmt.Errorf("client send contacts: create request url: %w", err)
	}

	resp, err := whttp.SendMessage(ctx, client.http, http.MethodPost, reqURL, params, payload)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	baseURL := client.baseURL
	phoneNumberID := client.PhoneNumberID()
	apiVersion := client.version
	accessToken := client.AccessToken()

	reqURL, err := whttp.CreateRequestURL(baseURL, apiVersion, phoneNumberID, "/messages")
	if err != nil {
		return nil, fmt.Errorf("client mark message read: %w", err)
	}
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: "whatsapp",
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("client mark message read: %w", err)
	}

	params := &whttp.RequestParams{
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  accessToken,
	}

	request, err := whttp.NewRequestWithContext(ctx, http.MethodPost, reqURL, params, payload)
	if err != nil {
		return nil, fmt.Errorf("client mark message read: %w", err)
	}

	resp, err := client.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("client mark message read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client mark message read: status code: %d", resp.StatusCode)
	}

	var result StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("client mark message read: %w", err)
	}

	return &result, nil
}

type Template struct {
	LanguageCode   string
	LanguagePolicy string
	Name           string
	Components     []*models.TemplateComponent
}

// SendTemplate sends a template message to the recipient.
func (client *Client) SendTemplate(ctx context.Context, recipient string, req *Template) (*whttp.Response, error) {
	request := &SendTemplateRequest{
		BaseURL:                client.baseURL,
		AccessToken:            client.AccessToken(),
		PhoneNumberID:          client.PhoneNumberID(),
		ApiVersion:             client.version,
		Recipient:              recipient,
		TemplateLanguageCode:   req.LanguageCode,
		TemplateLanguagePolicy: req.LanguagePolicy,
		TemplateName:           req.Name,
		TemplateComponents:     req.Components,
	}

	resp, err := SendTemplate(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

////////////// QrCode

func (client *Client) CreateQrCode(ctx context.Context, message *qrcodes.CreateRequest) (
	*qrcodes.CreateResponse, error,
) {
	request := &qrcodes.CreateRequest{
		PrefilledMessage: message.PrefilledMessage,
		ImageFormat:      message.ImageFormat,
	}

	rctx := &qrcodes.RequestContext{
		BaseURL:     client.baseURL,
		PhoneID:     client.PhoneNumberID(),
		ApiVersion:  client.version,
		AccessToken: client.AccessToken(),
	}
	resp, err := qrcodes.Create(ctx, client.http, rctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) ListQrCodes(ctx context.Context) (*qrcodes.ListResponse, error) {
	rctx := &qrcodes.RequestContext{
		BaseURL:     client.baseURL,
		PhoneID:     client.PhoneNumberID(),
		ApiVersion:  client.version,
		AccessToken: client.AccessToken(),
	}

	resp, err := qrcodes.List(ctx, client.http, rctx)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) GetQrCode(ctx context.Context, qrCodeID string) (*qrcodes.Information, error) {
	rctx := &qrcodes.RequestContext{
		BaseURL:     client.baseURL,
		PhoneID:     client.PhoneNumberID(),
		ApiVersion:  client.version,
		AccessToken: client.AccessToken(),
	}

	resp, err := qrcodes.Get(ctx, client.http, rctx, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) UpdateQrCode(ctx context.Context, qrCodeID string, request *qrcodes.CreateRequest,
) (*qrcodes.SuccessResponse, error) {
	rctx := &qrcodes.RequestContext{
		BaseURL:     client.baseURL,
		PhoneID:     client.PhoneNumberID(),
		ApiVersion:  client.version,
		AccessToken: client.AccessToken(),
	}

	resp, err := qrcodes.Update(ctx, client.http, rctx, qrCodeID, request)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) DeleteQrCode(ctx context.Context, qrCodeID string) (*qrcodes.SuccessResponse, error) {
	rctx := &qrcodes.RequestContext{
		BaseURL:     client.baseURL,
		PhoneID:     client.PhoneNumberID(),
		ApiVersion:  client.version,
		AccessToken: client.AccessToken(),
	}

	resp, err := qrcodes.Delete(ctx, client.http, rctx, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}
