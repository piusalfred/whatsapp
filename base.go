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

	"github.com/piusalfred/whatsapp/pkg/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/pkg/models"
	"github.com/piusalfred/whatsapp/pkg/models/factories"
)

type (
	Client struct {
		base *whttp.BaseClient
		mw   []SendMiddleware
	}

	ClientOption func(*Client)

	InteractiveCTAButtonURLRequest struct {
		Recipient string
		Params    *factories.CTAButtonURLParameters
	}
)

func (c *Client) Image(ctx context.Context, params *RequestParams, image *models.Image,
	options *whttp.CacheOptions,
) (*whttp.ResponseMessage, error) {
	message, err := factories.ImageMessage(params.Recipient, image,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("image message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, options), message)
}

func (c *Client) Audio(ctx context.Context, params *RequestParams, audio *models.Audio,
	options *whttp.CacheOptions,
) (*whttp.ResponseMessage, error) {
	message, err := factories.AudioMessage(params.Recipient, audio,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("audio message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, options), message)
}

func (c *Client) Video(ctx context.Context, params *RequestParams, video *models.Video,
	options *whttp.CacheOptions,
) (*whttp.ResponseMessage, error) {
	message, err := factories.VideoMessage(params.Recipient, video,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("video message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, options), message)
}

func (c *Client) Document(ctx context.Context, params *RequestParams, document *models.Document,
	options *whttp.CacheOptions,
) (*whttp.ResponseMessage, error) {
	message, err := factories.DocumentMessage(params.Recipient, document,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("document message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, options), message)
}

func (c *Client) Sticker(ctx context.Context, params *RequestParams, sticker *models.Sticker,
	options *whttp.CacheOptions,
) (*whttp.ResponseMessage, error) {
	message, err := factories.StickerMessage(params.Recipient, sticker,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("sticker message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, options), message)
}

func (c *Client) InteractiveMessage(ctx context.Context, params *RequestParams,
	interactive *models.Interactive,
) (*whttp.ResponseMessage, error) {
	message, err := factories.InteractiveMessage(params.Recipient, interactive,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("interactive message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

func (c *Client) Template(ctx context.Context, params *RequestParams,
	template *models.Template,
) (*whttp.ResponseMessage, error) {
	message, err := factories.TemplateMessage(params.Recipient, template,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("template message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

func (c *Client) Contacts(ctx context.Context, params *RequestParams, contacts []*models.Contact) (
	*whttp.ResponseMessage, error,
) {
	message, err := factories.ContactsMessage(params.Recipient, contacts,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("contacts message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

func (c *Client) Location(ctx context.Context, params *RequestParams,
	request *models.Location,
) (*whttp.ResponseMessage, error) {
	message, err := factories.LocationMessage(params.Recipient, request,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("location message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

func (c *Client) React(ctx context.Context, params *RequestParams,
	msg *models.Reaction,
) (*whttp.ResponseMessage, error) {
	message, err := factories.ReactionMessage(params.Recipient, msg)
	if err != nil {
		return nil, fmt.Errorf("reaction message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

func (c *Client) Text(ctx context.Context, params *RequestParams, text *models.Text) (*whttp.ResponseMessage, error) {
	message, err := factories.TextMessage(params.Recipient, text,
		factories.WithReplyToMessageID(params.ReplyID))
	if err != nil {
		return nil, fmt.Errorf("text message: %w", err)
	}

	return c.Send(ctx, fmtParamsToContext(params, nil), message)
}

// WithBaseClientMiddleware adds a middleware to the base client.
func WithBaseClientMiddleware(mw ...SendMiddleware) ClientOption {
	return func(client *Client) {
		client.mw = append(client.mw, mw...)
	}
}

// WithBaseHTTPClient sets the http client for the base client.
func WithBaseHTTPClient(httpClient *whttp.BaseClient) ClientOption {
	return func(client *Client) {
		client.base = httpClient
	}
}

// WithBaseClientOptions sets the options for the base client.
func WithBaseClientOptions(options []whttp.BaseClientOption) ClientOption {
	return func(client *Client) {
		client.base.ApplyOptions(options...)
	}
}

// NewBaseClient creates a new base client.
func NewBaseClient(ctx context.Context, configure config.Reader, options ...ClientOption) (*Client, error) {
	inner, err := whttp.InitBaseClient(ctx, configure)
	if err != nil {
		return nil, fmt.Errorf("init base client: %w", err)
	}

	b := &Client{base: inner}

	for _, option := range options {
		option(b)
	}

	return b, nil
}

func (c *Client) MarkMessageRead(ctx context.Context, req *whttp.RequestContext,
	messageID string,
) (*whttp.ResponseStatus, error) {
	reqBody := &statusUpdateRequest{
		MessagingProduct: factories.MessagingProductWhatsApp,
		Status:           "read",
		MessageID:        messageID,
	}

	params := whttp.MakeRequest(
		whttp.WithRequestContext(req),
		whttp.WithRequestPayload(reqBody),
	)

	var success whttp.ResponseStatus
	err := c.base.Do(ctx, params, &success)
	if err != nil {
		return nil, fmt.Errorf("mark message read: %w", err)
	}

	return &success, nil
}

func (c *Client) Send(ctx context.Context, req *whttp.RequestContext,
	message *models.Message,
) (*whttp.ResponseMessage, error) {
	fs := WrapSender(SenderFunc(c.send), c.mw...)

	resp, err := fs.Send(ctx, req, message)
	if err != nil {
		return nil, fmt.Errorf("base client: %w", err)
	}

	return resp, nil
}

func (c *Client) send(ctx context.Context, req *whttp.RequestContext,
	msg *models.Message,
) (*whttp.ResponseMessage, error) {
	request := whttp.MakeRequest(
		whttp.WithRequestContext(req),
		whttp.WithRequestPayload(msg),
		whttp.WithRequestMethod(http.MethodPost),
		whttp.WithRequestHeaders(map[string]string{
			"Content-Type": "application/json",
		}),
	)

	resp, err := c.base.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.String(), err)
	}

	return resp, nil
}

func fmtParamsToContext(params *RequestParams, cache *whttp.CacheOptions) *whttp.RequestContext {
	return whttp.MakeRequestContext(
		whttp.WithRequestContextID(params.ID),
		whttp.WithRequestContextMetadata(params.Metadata),
		whttp.WithRequestContextAction(whttp.RequestActionSend),
		whttp.WithRequestContextCategory(whttp.RequestCategoryMessage),
		whttp.WithRequestContextName(whttp.RequestNameLocation),
		whttp.WithRequestContextCacheOptions(cache),
	)
}
