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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	whttp "github.com/piusalfred/whatsapp/http"
	"github.com/piusalfred/whatsapp/models"
	"github.com/piusalfred/whatsapp/qrcodes"
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

// MediaMaxAllowedSize returns the allowed maximum size for media. It returns
// -1 for unknown media type. Currently, it checks for MediaTypeAudio,MediaTypeVideo,
// MediaTypeImage, MediaTypeSticker,MediaTypeDocument.
func MediaMaxAllowedSize(mediaType MediaType) int {
	sizeMap := map[MediaType]int{
		MediaTypeAudio:    MaxAudioSize,
		MediaTypeDocument: MaxDocSize,
		MediaTypeSticker:  MaxStickerSize,
		MediaTypeImage:    MaxImageSize,
		MediaTypeVideo:    MaxVideoSize,
	}

	size, ok := sizeMap[mediaType]
	if ok {
		return size
	}

	return -1
}

type (
	ResponseMessage struct {
		Product  string             `json:"messaging_product,omitempty"`
		Contacts []*ResponseContact `json:"contacts,omitempty"`
		Messages []*MessageID       `json:"messages,omitempty"`
	}
	MessageID struct {
		ID string `json:"id,omitempty"`
	}

	ResponseContact struct {
		Input      string `json:"input"`
		WhatsappID string `json:"wa_id"`
	}

	// MessageType represents the type of message currently supported.
	// Which are Text messages,Reaction messages,Media messages,Location messages,Contact messages,
	// and Interactive messages.
	// You may also send any of these message types as a reply, except reaction messages.
	// For more go to https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages
	MessageType string

	// Client includes the http client, base url, apiVersion, access token, phone number id,
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
		apiVersion        string
		accessToken       string
		phoneNumberID     string
		businessAccountID string
		responseHooks     []whttp.ResponseHook
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
		client.apiVersion = version
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

func WithResponseHooks(hooks ...whttp.ResponseHook) ClientOption {
	return func(client *Client) {
		client.responseHooks = hooks
	}
}

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		rwm:               &sync.RWMutex{},
		http:              http.DefaultClient,
		baseURL:           BaseURL,
		apiVersion:        "v16.0",
		accessToken:       "",
		phoneNumberID:     "",
		businessAccountID: "",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type clientContext struct {
	baseURL           string
	apiVersion        string
	accessToken       string
	phoneNumberID     string
	businessAccountID string
}

func (client *Client) context() *clientContext {
	client.rwm.RLock()
	defer client.rwm.RUnlock()

	return &clientContext{
		baseURL:           client.baseURL,
		apiVersion:        client.apiVersion,
		accessToken:       client.accessToken,
		phoneNumberID:     client.phoneNumberID,
		businessAccountID: client.businessAccountID,
	}
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

func (client *Client) SetBusinessAccountID(businessAccountID string) {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	client.businessAccountID = businessAccountID
}

type TextMessage struct {
	Message    string
	PreviewURL bool
}

// SendTextMessage sends a text message to a WhatsApp Business Account.
func (client *Client) SendTextMessage(ctx context.Context, recipient string,
	message *TextMessage,
) (*ResponseMessage, error) {
	cctx := client.context()
	request := &SendTextRequest{
		BaseURL:       cctx.baseURL,
		AccessToken:   cctx.accessToken,
		PhoneNumberID: cctx.phoneNumberID,
		ApiVersion:    cctx.apiVersion,
		Recipient:     recipient,
		Message:       message.Message,
		PreviewURL:    message.PreviewURL,
	}
	resp, err := SendText(ctx, client.http, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	return resp, nil
}

// SendLocationMessage sends a location message to a WhatsApp Business Account.
func (client *Client) SendLocationMessage(ctx context.Context, recipient string,
	message *models.Location,
) (*ResponseMessage, error) {
	request := &SendLocationRequest{
		BaseURL:       client.baseURL,
		AccessToken:   client.accessToken,
		PhoneNumberID: client.phoneNumberID,
		ApiVersion:    client.apiVersion,
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

func (client *Client) React(ctx context.Context, recipient string, req *ReactMessage) (*ResponseMessage, error) {
	cctx := client.context()
	request := &ReactRequest{
		BaseURL:       cctx.baseURL,
		AccessToken:   cctx.accessToken,
		PhoneNumberID: cctx.phoneNumberID,
		ApiVersion:    cctx.apiVersion,
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
) (*ResponseMessage, error) {
	cctx := client.context()
	request := &SendMediaRequest{
		BaseURL:       cctx.baseURL,
		AccessToken:   cctx.accessToken,
		PhoneNumberID: cctx.phoneNumberID,
		ApiVersion:    cctx.apiVersion,
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

func (client *Client) Reply(ctx context.Context, recipient string, req *ReplyMessage) (*ResponseMessage, error) {
	cctx := client.context()
	request := &ReplyRequest{
		BaseURL:       cctx.baseURL,
		AccessToken:   cctx.accessToken,
		PhoneNumberID: cctx.phoneNumberID,
		ApiVersion:    cctx.apiVersion,
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

func (client *Client) SendContacts(ctx context.Context, recipient string, contacts *models.Contacts) (
	*ResponseMessage, error,
) {
	cctx := client.context()
	req := &SendContactRequest{
		BaseURL:       cctx.baseURL,
		AccessToken:   cctx.accessToken,
		PhoneNumberID: cctx.phoneNumberID,
		ApiVersion:    cctx.apiVersion,
		Recipient:     recipient,
		Contacts:      contacts,
	}

	resp, err := SendContact(ctx, client.http, req)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

// MarkMessageRead sends a read receipt for a message.
func (client *Client) MarkMessageRead(ctx context.Context, messageID string) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: "whatsapp",
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	cctx := client.context()

	reqCtx := &whttp.RequestContext{
		Name:       "mark read",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   cctx.phoneNumberID,
		Endpoints:  []string{"/messages"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/json"},
		Bearer:  cctx.accessToken,
		Payload: reqBody,
	}

	var success StatusResponse
	err := whttp.Send(ctx, client.http, params, &success)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return &success, nil
}

type Template struct {
	LanguageCode   string
	LanguagePolicy string
	Name           string
	Components     []*models.TemplateComponent
}

// SendTemplate sends a template message to the recipient.
func (client *Client) SendTemplate(ctx context.Context, recipient string, req *Template) (*ResponseMessage, error) {
	cctx := client.context()
	request := &SendTemplateRequest{
		BaseURL:                cctx.baseURL,
		AccessToken:            cctx.accessToken,
		PhoneNumberID:          cctx.phoneNumberID,
		ApiVersion:             cctx.apiVersion,
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

	cctx := client.context()

	rctx := &qrcodes.RequestContext{
		BaseURL:     cctx.baseURL,
		PhoneID:     cctx.phoneNumberID,
		ApiVersion:  cctx.apiVersion,
		AccessToken: client.accessToken,
	}
	resp, err := qrcodes.Create(ctx, client.http, rctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) ListQrCodes(ctx context.Context) (*qrcodes.ListResponse, error) {
	cctx := client.context()
	rctx := &qrcodes.RequestContext{
		BaseURL:     cctx.baseURL,
		PhoneID:     cctx.phoneNumberID,
		ApiVersion:  cctx.apiVersion,
		AccessToken: cctx.accessToken,
	}

	resp, err := qrcodes.List(ctx, client.http, rctx)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) GetQrCode(ctx context.Context, qrCodeID string) (*qrcodes.Information, error) {
	cctx := client.context()
	rctx := &qrcodes.RequestContext{
		BaseURL:     cctx.baseURL,
		PhoneID:     cctx.phoneNumberID,
		ApiVersion:  cctx.apiVersion,
		AccessToken: cctx.accessToken,
	}

	resp, err := qrcodes.Get(ctx, client.http, rctx, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) UpdateQrCode(ctx context.Context, qrCodeID string, request *qrcodes.CreateRequest,
) (*qrcodes.SuccessResponse, error) {
	cctx := client.context()
	rctx := &qrcodes.RequestContext{
		BaseURL:     cctx.baseURL,
		PhoneID:     cctx.phoneNumberID,
		ApiVersion:  cctx.apiVersion,
		AccessToken: cctx.accessToken,
	}

	resp, err := qrcodes.Update(ctx, client.http, rctx, qrCodeID, request)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

func (client *Client) DeleteQrCode(ctx context.Context, qrCodeID string) (*qrcodes.SuccessResponse, error) {
	cctx := client.context()
	rctx := &qrcodes.RequestContext{
		BaseURL:     cctx.baseURL,
		PhoneID:     cctx.phoneNumberID,
		ApiVersion:  cctx.apiVersion,
		AccessToken: cctx.accessToken,
	}

	resp, err := qrcodes.Delete(ctx, client.http, rctx, qrCodeID)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return resp, nil
}

////// PHONE NUMBERS

var (
	SMSVerificationMethod   VerificationMethod = "SMS"
	VoiceVerificationMethod VerificationMethod = "VOICE"
)

type (
	// VerificationMethod is the method to use to verify the phone number. It can be SMS or VOICE
	VerificationMethod string

	PhoneNumber struct {
		VerifiedName       string `json:"verified_name"`
		DisplayPhoneNumber string `json:"display_phone_number"`
		ID                 string `json:"id"`
		QualityRating      string `json:"quality_rating"`
	}

	PhoneNumbersList struct {
		Data    []*PhoneNumber `json:"data,omitempty"`
		Paging  *Paging        `json:"paging,omitempty"`
		Summary *Summary       `json:"summary,omitempty"`
	}

	Paging struct {
		Cursors *Cursors `json:"cursors,omitempty"`
	}

	Cursors struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	}

	Summary struct {
		TotalCount int `json:"total_count,omitempty"`
	}

	// PhoneNumberNameStatus value can be one of the following:
	// APPROVED: The name has been approved. You can download your certificate now.
	// AVAILABLE_WITHOUT_REVIEW: The certificate for the phone is available and display name is ready to use
	// without review.
	// DECLINED: The name has not been approved. You cannot download your certificate.
	// EXPIRED: Your certificate has expired and can no longer be downloaded.
	// PENDING_REVIEW: Your name request is under review. You cannot download your certificate.
	// NONE: No certificate is available.
	PhoneNumberNameStatus string

	FilterParams struct {
		Field    string `json:"field,omitempty"`
		Operator string `json:"operator,omitempty"`
		Value    string `json:"value,omitempty"`
	}
)

// RequestVerificationCode requests a verification code to be sent via SMS or VOICE.
// doc link: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers
//
// You need to verify the phone number you want to use to send messages to your customers. After the
// API call, you will receive your verification code via the method you selected. To finish the verification
// process, include your code in the VerifyCode method.
func (client *Client) RequestVerificationCode(ctx context.Context,
	codeMethod VerificationMethod, language string,
) error {
	cctx := client.context()
	reqCtx := &whttp.RequestContext{
		Name:       "request code",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   cctx.phoneNumberID,
		Endpoints:  []string{"request_code"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  cctx.accessToken,
		Form:    map[string]string{"code_method": string(codeMethod), "language": language},
		Payload: nil,
	}
	err := whttp.Send(ctx, client.http, params, nil, client.responseHooks...)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// VerifyCode should be run to verify the code retrieved by RequestVerificationCode.
func (client *Client) VerifyCode(ctx context.Context, code string) (*StatusResponse, error) {
	cctx := client.context()
	reqCtx := &whttp.RequestContext{
		Name:       "verify code",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   cctx.phoneNumberID,
		Endpoints:  []string{"verify_code"},
	}
	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Query:   nil,
		Bearer:  cctx.accessToken,
		Form:    map[string]string{"code": code},
	}

	var resp StatusResponse
	err := whttp.Send(ctx, client.http, params, &resp, client.responseHooks...)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &resp, nil
}

// ListPhoneNumbers returns a list of phone numbers that are associated with the business account.
// using the WhatsApp Business Management API.
//
// You will need to have
//   - The WhatsApp Business Account ID for the business' phone numbers you want to retrieve
//   - A System User access token linked to your WhatsApp Business Account
//   - The whatsapp_business_management permission
//
// Limitations
// This API can only retrieve phone numbers that have been registered. Adding, updating, or
// deleting phone numbers is not permitted using the API.
//
// The equivalent curl command to retrieve phone numbers is (formatted for readability):
//
//		curl -X GET "https://graph.facebook.com/v16.0/{whatsapp-business-account-id}/phone_numbers
//	      	?access_token={system-user-access-token}"
//
// On success, a JSON object is returned with a list of all the business names, phone numbers,
// phone number IDs, and quality ratings associated with a business.
//
//	{
//	  "data": [
//	    {
//	      "verified_name": "Jasper's Market",
//	      "display_phone_number": "+1 631-555-5555",
//	      "id": "1906385232743451",
//	      "quality_rating": "GREEN"
//
//		    },
//		    {
//		      "verified_name": "Jasper's Ice Cream",
//		      "display_phone_number": "+1 631-555-5556",
//		      "id": "1913623884432103",
//		      "quality_rating": "NA"
//		    }
//		  ],
//		}
//
// Filter Phone Numbers
// You can query phone numbers and filter them based on their account_mode. This filtering option
// is currently being tested in beta mode. Not all developers have access to it.
//
// Sample Request
//
//	curl -i -X GET "https://graph.facebook.com/v16.0/{whatsapp-business-account-ID}/phone_numbers?\
//		filtering=[{"field":"account_mode","operator":"EQUAL","value":"SANDBOX"}]&access_token=access-token"
//
// Sample Response
//
//	{
//	  "data": [
//	    {
//	      "id": "1972385232742141",
//	      "display_phone_number": "+1 631-555-1111",
//	      "verified_name": "John’s Cake Shop",
//	      "quality_rating": "UNKNOWN",
//	    }
//	  ],
//	  "paging": {
//		"cursors": {
//			"before": "abcdefghij",
//			"after": "klmnopqr"
//		}
//	   }
//	}
func (client *Client) ListPhoneNumbers(ctx context.Context, filters []*FilterParams) (*PhoneNumbersList, error) {
	cctx := client.context()
	reqCtx := &whttp.RequestContext{
		Name:       "list phone numbers",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   cctx.businessAccountID,
		Endpoints:  []string{"phone_numbers"},
	}

	params := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Query:   map[string]string{"access_token": cctx.accessToken},
	}
	if filters != nil {
		p := filters
		jsonParams, err := json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filter params: %w", err)
		}
		params.Query["filtering"] = string(jsonParams)
	}
	var phoneNumbersList PhoneNumbersList
	err := whttp.Send(ctx, client.http, params, &phoneNumbersList, client.responseHooks...)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return &phoneNumbersList, nil
}

// PhoneNumberByID returns the phone number associated with the given ID.
func (client *Client) PhoneNumberByID(ctx context.Context) (*PhoneNumber, error) {
	cctx := client.context()
	reqCtx := &whttp.RequestContext{
		Name:       "get phone number by id",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   cctx.phoneNumberID,
	}
	request := &whttp.Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Headers: map[string]string{
			"Authorization": "Bearer " + cctx.accessToken,
		},
	}
	var phoneNumber PhoneNumber
	if err := whttp.Send(ctx, client.http, request, &phoneNumber, client.responseHooks...); err != nil {
		return nil, fmt.Errorf("get phone muber by id: %w", err)
	}

	return &phoneNumber, nil
}

/// MEDIA

type MediaURLGetter func(ctx context.Context, mediaID string) (*MediaInformation, error)

// MediaURL returns the URL for the media file. All the media URL expires after 5 minutes. Another one can be
// retrieved by calling the MediaURL method again.
// MediaURL methods requires the media ID and the access token. The mediaId is the ID of the media file. There
// are two ways to get this ID:
//
//   - From the API call: Once you have successfully uploaded media files to the API, the media ID is included
//     in the response to your call.
//
//   - From Webhooks: When a business account receives a media message, it downloads the media and uploads it to
//     the Cloud API automatically. That event triggers the Webhooks and sends you a notification that includes
//     the media ID.
//
// More info https://developers.facebook.com/docs/whatsapp/cloud-api/reference/media
func (client *Client) MediaURL(ctx context.Context, mediaID string) (*MediaInformation, error) {
	cctx := client.context()
	rctx := &whttp.RequestContext{
		Name:       "retrieve-media-url",
		BaseURL:    cctx.baseURL,
		ApiVersion: cctx.apiVersion,
		SenderID:   "",
		Endpoints:  []string{mediaID},
	}

	request := &whttp.Request{
		Context: rctx,
		Method:  http.MethodGet,
		Headers: nil,
		Query:   nil,
		Bearer:  cctx.accessToken,
		Form:    nil,
		Payload: nil,
	}

	var information MediaInformation
	if err := whttp.Send(ctx, client.http, request, &information); err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	return &information, nil
}

type (
	DownloadResponse struct {
		Body    io.ReadCloser
		Headers http.Header
	}

	DownloadOptions struct {
		RetryOn404 bool
		MaxRetries int
	}
)

// DownloadMedia downloads the media file from the URL provided by the MediaURL method.To download media,
// make a GET call to your media’s URL. All media URLs expire after 5 minutes —you need to retrieve the
// media URL again if it expires. If you directly click on the URL you get from a /MEDIA_ID GET call, you
// get an access error.
//
// You must also provide a token for Endpoint Authentication. Developers can authenticate their API calls
// with the access token generated in the App Dashboard > WhatsApp > Getting Started (or Setup) panel.
//
// Business Solution Providers (BSPs) must authenticate themselves with an access token with the
// whatsapp_business_messaging permission.
//
// If successful, you will receive the binary data of media saved in media_file, response headers contain
// a content-type header to indicate the mime type of returned data. Check supported media types for more
// information.
//
// If media fails to download, you will receive a 404 Not Found response code. In that case, we recommend you
// try to retrieve a new media URL and download it again. If doing so doesn't resolve the issue, please try to
// renew the ACCESS_TOKEN then retry downloading the media.
func (client *Client) DownloadMedia(ctx context.Context, mediaID string, maxRetries int) (*DownloadResponse, error) {
	mediaURLGetter := client.MediaURL
	cctx := client.context()
	accessToken := cctx.accessToken

	for retry := 0; retry <= maxRetries; retry++ {
		mediaInfo, err := mediaURLGetter(ctx, mediaID)
		if err != nil {
			return nil, fmt.Errorf("client: download media: failed to get media URL: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaInfo.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("client: download media: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.http.Do(req)
		if err != nil {
			if retry < maxRetries {
				var netErr net.Error
				if ok := errors.Is(err, netErr); ok && netErr.Timeout() {
					continue // retry on timeouts
				}

				continue // retry on all errors
			}

			return nil, fmt.Errorf("client: download media: failed to download media: %w", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			if retry < maxRetries {
				continue // retry on 404 errors
			}

			return nil, fmt.Errorf("client: download media: %w: (404)", ErrMediaNotFound)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()

			return nil, fmt.Errorf("client: download media: %w: status code %d", ErrDownloadFailed, resp.StatusCode)
		}

		return &DownloadResponse{
			Body:    resp.Body,
			Headers: resp.Header,
		}, nil
	}

	return nil, fmt.Errorf("client: download media: retries (%d): %w", maxRetries, ErrMaxRetriesReached)
}

var (
	ErrDownloadFailed    = errors.New("media download failed")
	ErrMaxRetriesReached = errors.New("max retries reached")
	ErrMediaNotFound     = errors.New("media not found")
)

func (client *Client) UploadFile(ctx context.Context, mediaPath string, mediaType string) (string, error) {
	// Open media file
	file, err := os.Open(mediaPath)
	if err != nil {
		return "", fmt.Errorf("client: upload file: failed to open media file: %w", err)
	}
	defer file.Close()

	cctx := client.context()
	reqURL, err := whttp.CreateRequestURL(cctx.baseURL, cctx.apiVersion, cctx.phoneNumberID, "media")
	if err != nil {
		return "", err
	}

	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(mediaPath))
	if err != nil {
		return "", fmt.Errorf("client: upload file: failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("client: upload file: failed to copy media data: %w", err)
	}
	if err := writer.WriteField("type", mediaType); err != nil {
		return "", fmt.Errorf("client: upload file: failed to write field: %w", err)
	}
	if err := writer.WriteField("messaging_product", "whatsapp"); err != nil {
		return "", fmt.Errorf("client: upload file: failed to write field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("client: upload file: failed to close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cctx.accessToken)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.http.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	return "", nil
}
