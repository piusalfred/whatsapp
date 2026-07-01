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

package user

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/piusalfred/whatsapp/config"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const BlockEndpoint = "block_users"

type BlockAction string

const (
	BlockActionBlock       BlockAction = "block"
	BlockActionUnblock     BlockAction = "unblock"
	BlockActionListBlocked BlockAction = "list"
)

type (
	BlockRequest struct {
		Numbers []string
	}

	UnblockRequest struct {
		Numbers []string
	}

	ListBlockedUsersOptions struct {
		Limit  *int
		After  *string
		Before *string
	}

	BlockResponse struct {
		MessagingProduct string         `json:"messaging_product"`
		BlockUsers       *BlockUsers    `json:"block_users"`
		Error            *werrors.Error `json:"error,omitempty"`
	}

	ListBlockedUsersResponse struct {
		MessagingProduct string                 `json:"messaging_product"`
		Data             []ListBlockedUserEntry `json:"data"`
		Paging           *whttp.Paging          `json:"paging,omitempty"`
		Error            *werrors.Error         `json:"error,omitempty"`
	}

	BlockedUser struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	}

	Identifier struct {
		User string `json:"user"`
	}

	BlockUsers struct {
		AddedUsers   []BlockResult `json:"added_users,omitempty"`
		RemovedUsers []BlockResult `json:"removed_users,omitempty"`
		FailedUsers  []BlockResult `json:"failed_users,omitempty"`
	}

	BlockResult struct {
		Input  string          `json:"input"`
		WaID   string          `json:"wa_id"`
		Errors []werrors.Error `json:"errors,omitempty"`
	}

	ListBlockedUserEntry struct {
		MessagingProduct string `json:"messaging_product"`
		WaID             string `json:"wa_id"`
	}

	BlockBaseRequest struct {
		MessagingProduct string                   `json:"messaging_product,omitempty"`
		BlockUsers       []Identifier             `json:"block_users,omitempty"`
		BlockAction      BlockAction              `json:"-"`
		ListOptions      *ListBlockedUsersOptions `json:"-"`
	}

	BlockBaseResponse struct {
		MessagingProduct string                 `json:"messaging_product"`
		Data             []ListBlockedUserEntry `json:"data,omitempty"`
		BlockUsers       *BlockUsers            `json:"block_users,omitempty"`
		Paging           *whttp.Paging          `json:"paging,omitempty"`
		Error            *werrors.Error         `json:"error,omitempty"`
	}

	BlockBaseRequestOption func(*BlockBaseRequest)

	// BlockClient is a high-level client for the WhatsApp Block Users API.
	// It binds a fixed [config.Config] and a [BlockBaseClient] sender.
	BlockClient struct {
		config *config.Config
		sender *BlockBaseClient
	}
)

func (client *BlockClient) Unblock(ctx context.Context, request *UnblockRequest) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionUnblock, WithBlockUsersBaseRequestNumbers(request.Numbers))

	resp, err := client.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to unblock users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (client *BlockClient) ListBlocked(ctx context.Context, request *ListBlockedUsersOptions) (
	*ListBlockedUsersResponse, error,
) {
	req := NewBlockBaseRequest(BlockActionListBlocked, func(r *BlockBaseRequest) {
		r.ListOptions = request
	})

	resp, err := client.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked users: %w", err)
	}

	return resp.ListResponse(), nil
}

func (client *BlockClient) Block(ctx context.Context, request *BlockRequest) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionBlock, WithBlockUsersBaseRequestNumbers(request.Numbers))

	resp, err := client.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to block users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (client *BlockClient) Send(ctx context.Context, request *BlockBaseRequest) (
	*BlockBaseResponse, error,
) {
	resp, err := client.sender.Send(ctx, client.config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func (response *BlockBaseResponse) BlockUsersResponse() *BlockResponse {
	return &BlockResponse{
		MessagingProduct: response.MessagingProduct,
		BlockUsers:       response.BlockUsers,
		Error:            response.Error,
	}
}

func (response *BlockBaseResponse) ListResponse() *ListBlockedUsersResponse {
	return &ListBlockedUsersResponse{
		MessagingProduct: response.MessagingProduct,
		Data:             response.Data,
		Paging:           response.Paging,
		Error:            response.Error,
	}
}

// RemovedUsersResponse returns a response tailored for unblock operations.
func (response *BlockBaseResponse) RemovedUsersResponse() *BlockResponse {
	return &BlockResponse{
		MessagingProduct: response.MessagingProduct,
		BlockUsers:       response.BlockUsers,
		Error:            response.Error,
	}
}

func WithBlockUsersBaseRequestNumbers(numbers []string) BlockBaseRequestOption {
	return func(r *BlockBaseRequest) {
		for _, number := range numbers {
			r.BlockUsers = append(r.BlockUsers, Identifier{User: number})
		}
	}
}

func WithBlockUsersBaseRequestListLimit(limit int) BlockBaseRequestOption {
	return func(r *BlockBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.Limit = &limit
	}
}

func WithBlockUsersBaseRequestListAfter(after string) BlockBaseRequestOption {
	return func(r *BlockBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.After = &after
	}
}

func WithBlockUsersBaseRequestListBefore(before string) BlockBaseRequestOption {
	return func(r *BlockBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.Before = &before
	}
}

func NewBlockBaseRequest(action BlockAction, options ...BlockBaseRequestOption) *BlockBaseRequest {
	b := &BlockBaseRequest{
		MessagingProduct: "whatsapp",
		BlockAction:      action,
	}

	for _, option := range options {
		if option != nil {
			option(b)
		}
	}

	return b
}

// NewBlockClient creates a high-level [BlockClient] with a fixed configuration.
// Optional [SenderOption] functions tune the underlying HTTP transport.
func NewBlockClient(config *config.Config, options ...whttp.CoreSenderOption) *BlockClient {
	return &BlockClient{
		config: config,
		sender: &BlockBaseClient{BaseClient: *whttp.NewBaseClient[BlockBaseRequest](options...)},
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (client *BlockClient) SetBaseClient(sender whttp.Sender[BlockBaseRequest]) {
	client.sender.SetSender(sender)
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (client *BlockClient) SetMiddlewares(mws ...whttp.Middleware[BlockBaseRequest]) {
	client.sender.SetMiddlewares(mws...)
}

// BlockBaseClient is a base client that accepts a concrete *config.Config per request.
// This makes it suitable for multi-tenant SaaS scenarios where each call may target a
// different tenant. For a fixed-configuration client, use BlockClient.
type BlockBaseClient struct {
	whttp.BaseClient[BlockBaseRequest]
}

func (client *BlockBaseClient) Block(
	ctx context.Context,
	conf *config.Config,
	request *BlockRequest,
) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionBlock, WithBlockUsersBaseRequestNumbers(request.Numbers))

	resp, err := client.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("failed to block users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (client *BlockBaseClient) Unblock(
	ctx context.Context,
	conf *config.Config,
	request *UnblockRequest,
) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionUnblock, WithBlockUsersBaseRequestNumbers(request.Numbers))

	resp, err := client.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("failed to unblock users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (client *BlockBaseClient) ListBlocked(
	ctx context.Context,
	conf *config.Config,
	request *ListBlockedUsersOptions,
) (*ListBlockedUsersResponse, error) {
	req := NewBlockBaseRequest(BlockActionListBlocked, func(r *BlockBaseRequest) {
		r.ListOptions = request
	})

	resp, err := client.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked users: %w", err)
	}

	return resp.ListResponse(), nil
}

func (client *BlockBaseClient) Send(ctx context.Context, conf *config.Config, request *BlockBaseRequest) (
	*BlockBaseResponse, error,
) {
	var (
		method  string
		params  map[string]string
		message *BlockBaseRequest
		reqType whttp.RequestType
	)

	switch request.BlockAction {
	case BlockActionBlock:
		method = http.MethodPost
		message = request
		reqType = whttp.RequestTypeBlockUsers
	case BlockActionUnblock:
		method = http.MethodDelete
		message = request
		reqType = whttp.RequestTypeUnblockUsers
	case BlockActionListBlocked:
		method = http.MethodGet
		reqType = whttp.RequestTypeListBlockedUsers
		if request.ListOptions != nil {
			params = map[string]string{}
			if request.ListOptions.Limit != nil {
				params["limit"] = strconv.Itoa(*request.ListOptions.Limit)
			}
			if request.ListOptions.After != nil {
				params["after"] = *request.ListOptions.After
			}
			if request.ListOptions.Before != nil {
				params["before"] = *request.ListOptions.Before
			}
		}
	default:
		method = http.MethodGet
	}

	bld := whttp.NewRequestBuilder(method, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(reqType).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, BlockEndpoint)

	if len(params) > 0 {
		bld = bld.QueryParams(params)
	}

	req := whttp.Build(bld, message)

	response := &BlockBaseResponse{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptionsPermissive())

	if err := client.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}
