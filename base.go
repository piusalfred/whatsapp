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
	"fmt"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
)

type (
	// BaseClient wraps the http client only and is used to make requests to the whatsapp api,
	// It does not have the context. This is ideally for making requests to the whatsapp api for
	// different users. The Client struct is used to make requests to the whatsapp api for a
	// single user.
	BaseClient struct {
		base *whttp.Client
		mw   []SendMiddleware
	}

	// BaseClientOption is a function that implements the BaseClientOption interface.
	BaseClientOption func(*BaseClient)

	InteractiveCTAButtonURLRequest struct {
		Recipient string
		Params    *models.CTAButtonURLParameters
	}
)

// WithBaseClientMiddleware adds a middleware to the base client.
func WithBaseClientMiddleware(mw ...SendMiddleware) BaseClientOption {
	return func(client *BaseClient) {
		client.mw = append(client.mw, mw...)
	}
}

// WithBaseHTTPClient sets the http client for the base client.
func WithBaseHTTPClient(httpClient *whttp.Client) BaseClientOption {
	return func(client *BaseClient) {
		client.base = httpClient
	}
}

// NewBaseClient creates a new base client.
func NewBaseClient(options ...BaseClientOption) *BaseClient {
	b := &BaseClient{base: whttp.NewClient()}

	for _, option := range options {
		option(b)
	}

	return b
}

func (c *BaseClient) SendTemplate(ctx context.Context, config *config.Values, req *SendTemplateRequest,
) (*ResponseMessage, error) {
	message := &models.Message{
		Product:       MessagingProduct,
		To:            req.Recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          models.MessageTypeTemplate,
		Template: &models.Template{
			Language: &models.TemplateLanguage{
				Code:   req.TemplateLanguageCode,
				Policy: req.TemplateLanguagePolicy,
			},
			Name:       req.TemplateName,
			Components: req.TemplateComponents,
		},
	}

	reqCtx := whttp.MakeRequestContext(config, whttp.RequestTypeTemplate, MessageEndpoint)

	response, err := c.Send(ctx, reqCtx, message)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *BaseClient) SendInteractiveCTAURLButton(ctx context.Context, config *config.Values,
	req *InteractiveCTAButtonURLRequest,
) (*ResponseMessage, error) {
	message := &models.Message{
		Product:       MessagingProduct,
		To:            req.Recipient,
		RecipientType: RecipientTypeIndividual,
		Type:          models.MessageTypeInteractive,
		Interactive:   models.NewInteractiveCTAURLButton(req.Params),
	}

	reqCtx := whttp.MakeRequestContext(config, whttp.RequestTypeInteractiveMessage, MessageEndpoint)

	response, err := c.Send(ctx, reqCtx, message)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *BaseClient) SendMedia(ctx context.Context, config *config.Values, req *SendMediaRequest,
) (*ResponseMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil: %w", ErrBadRequestFormat)
	}

	payload, err := formatMediaPayload(req)
	if err != nil {
		return nil, err
	}

	reqCtx := whttp.MakeRequestContext(config, whttp.RequestTypeMedia, MessageEndpoint)

	params := whttp.MakeRequest(whttp.WithRequestContext(reqCtx),
		whttp.WithBearer(config.AccessToken), whttp.WithPayload(payload))

	if req.CacheOptions != nil {
		if req.CacheOptions.CacheControl != "" {
			params.Headers["Cache-Control"] = req.CacheOptions.CacheControl
		} else if req.CacheOptions.Expires > 0 {
			params.Headers["Cache-Control"] = fmt.Sprintf("max-age=%d", req.CacheOptions.Expires)
		}
		if req.CacheOptions.LastModified != "" {
			params.Headers["Last-Modified"] = req.CacheOptions.LastModified
		}
		if req.CacheOptions.ETag != "" {
			params.Headers["ETag"] = req.CacheOptions.ETag
		}
	}

	var message ResponseMessage

	err = c.base.Do(ctx, params, &message)
	if err != nil {
		return nil, fmt.Errorf("send media: %w", err)
	}

	return &message, nil
}

// formatMediaPayload builds the payload for a media message. It accepts SendMediaOptions
// and returns a byte array and an error. This function is used internally by SendMedia.
// if neither ID nor Link is specified, it returns an error.
func formatMediaPayload(options *SendMediaRequest) ([]byte, error) {
	media := &models.Media{
		ID:       options.MediaID,
		Link:     options.MediaLink,
		Caption:  options.Caption,
		Filename: options.Filename,
		Provider: options.Provider,
	}
	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return nil, fmt.Errorf("format media payload: %w", err)
	}
	recipient := options.Recipient
	mediaType := string(options.Type)
	payloadBuilder := strings.Builder{}
	payloadBuilder.WriteString(`{"messaging_product":"whatsapp","recipient_type":"individual","to":"`)
	payloadBuilder.WriteString(recipient)
	payloadBuilder.WriteString(`","type": "`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`","`)
	payloadBuilder.WriteString(mediaType)
	payloadBuilder.WriteString(`":`)
	payloadBuilder.Write(mediaJSON)
	payloadBuilder.WriteString(`}`)

	return []byte(payloadBuilder.String()), nil
}

func (c *BaseClient) MarkMessageRead(ctx context.Context, req *whttp.RequestContext,
	messageID string,
) (*StatusResponse, error) {
	reqBody := &MessageStatusUpdateRequest{
		MessagingProduct: MessagingProduct,
		Status:           MessageStatusRead,
		MessageID:        messageID,
	}

	params := whttp.MakeRequest(
		whttp.WithRequestContext(req),
		whttp.WithBearer(req.Bearer),
		whttp.WithPayload(reqBody),
	)

	var success StatusResponse
	err := c.base.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("mark message read: %w", err)
	}

	return &success, nil
}

func (c *BaseClient) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message,
) (*ResponseMessage, error) {
	fs := WrapSender(SenderFunc(c.send), c.mw...)

	resp, err := fs.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("base client: %s: %w", req.RequestType, err)
	}

	return resp, nil
}

func (c *BaseClient) send(ctx context.Context, req *whttp.RequestContext,
	msg *models.Message,
) (*ResponseMessage, error) {
	request := whttp.MakeRequest(
		whttp.WithRequestContext(req),
		whttp.WithBearer(req.Bearer),
		whttp.WithPayload(msg))

	var resp ResponseMessage
	err := c.base.Do(ctx, request, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.RequestType, err)
	}

	return &resp, nil
}
