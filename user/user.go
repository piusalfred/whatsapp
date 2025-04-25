/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
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
	BlockUserParams struct {
		Numbers []string
		Action  BlockAction
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
		MessagingProduct string             `json:"messaging_product"`
		BlockUsers       *BlockUsers        `json:"block_users"`
		Data             []BlockedUsersData `json:"data"`
		Paging           *whttp.Paging      `json:"paging"`
		Error            *werrors.Error     `json:"error,omitempty"`
	}

	BlockedUser struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	}

	Identifier struct {
		User string `json:"user"`
	}

	BlockUsers struct {
		AddedUsers  []BlockResult `json:"added_users"`
		FailedUsers []BlockResult `json:"failed_users,omitempty"`
	}

	BlockResult struct {
		Input  string          `json:"input"`
		WaID   string          `json:"wa_id"`
		Errors []werrors.Error `json:"errors,omitempty"`
	}

	BlockedUsersData struct {
		BlockUsers []BlockResult `json:"block_users"`
	}

	BlockBaseRequest struct {
		MessagingProduct string                   `json:"messaging_product,omitempty"`
		BlockUsers       []Identifier             `json:"block_users,omitempty"`
		BlockAction      BlockAction              `json:"-"`
		ListOptions      *ListBlockedUsersOptions `json:"-"`
	}

	BlockBaseResponse struct {
		MessagingProduct string             `json:"messaging_product"`
		Data             []BlockedUsersData `json:"data"`
		BlockUsers       *BlockUsers        `json:"block_users"`
		Paging           *whttp.Paging      `json:"paging"`
		Error            *werrors.Error     `json:"error,omitempty"`
	}
	BlockBaseRequestOption func(*BlockBaseRequest)

	BlockClient struct {
		Config config.Reader
		Base   *BlockBaseClient
	}

	BlockService interface {
		Block(ctx context.Context, numbers []string) (*BlockResponse, error)
		Unblock(ctx context.Context, numbers []string) (*BlockResponse, error)
		ListBlocked(ctx context.Context, request *ListBlockedUsersOptions) (*ListBlockedUsersResponse, error)
	}
)

func (b *BlockClient) Unblock(ctx context.Context, numbers []string) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionUnblock, WithBlockUsersBaseRequestNumbers(numbers))

	resp, err := b.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to unblock users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (b *BlockClient) ListBlocked(ctx context.Context, request *ListBlockedUsersOptions) (
	*ListBlockedUsersResponse, error,
) {
	req := NewBlockBaseRequest(BlockActionListBlocked, func(r *BlockBaseRequest) {
		r.ListOptions = request
	})

	resp, err := b.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked users: %w", err)
	}

	return resp.ListResponse(), nil
}

func (b *BlockClient) Block(ctx context.Context, numbers []string) (*BlockResponse, error) {
	req := NewBlockBaseRequest(BlockActionBlock, WithBlockUsersBaseRequestNumbers(numbers))

	resp, err := b.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to block users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (b *BlockClient) Send(ctx context.Context, request *BlockBaseRequest) (
	*BlockBaseResponse, error,
) {
	resp, err := b.Base.Send(ctx, b.Config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

var _ BlockService = (*BlockClient)(nil)

func (base *BlockBaseResponse) BlockUsersResponse() *BlockResponse {
	return &BlockResponse{
		MessagingProduct: base.MessagingProduct,
		BlockUsers:       base.BlockUsers,
		Error:            base.Error,
	}
}

func (base *BlockBaseResponse) ListResponse() *ListBlockedUsersResponse {
	return &ListBlockedUsersResponse{
		MessagingProduct: base.MessagingProduct,
		Data:             base.Data,
		Paging:           base.Paging,
		Error:            base.Error,
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

func NewBlockClient(reader config.Reader, sender whttp.Sender[BlockBaseRequest]) *BlockClient {
	return &BlockClient{Config: reader, Base: &BlockBaseClient{Sender: sender}}
}

// BlockBaseClient is a base client meaning it can be used with changing configurations to send block requests.
// compared to the BlockClient which is used to send block requests with a fixed configuration.
type BlockBaseClient struct {
	Sender whttp.Sender[BlockBaseRequest]
}

func (b *BlockBaseClient) Send(ctx context.Context, reader config.Reader, request *BlockBaseRequest) (
	*BlockBaseResponse, error,
) {
	var method string

	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	options := []whttp.RequestOption[BlockBaseRequest]{
		whttp.WithRequestEndpoints[BlockBaseRequest](conf.APIVersion, conf.PhoneNumberID, BlockEndpoint),
		whttp.WithRequestBearer[BlockBaseRequest](conf.AccessToken),
		whttp.WithRequestAppSecret[BlockBaseRequest](conf.AppSecret),
		whttp.WithRequestSecured[BlockBaseRequest](conf.SecureRequests),
	}

	switch request.BlockAction {
	case BlockActionBlock:
		method = http.MethodPost
		options = append(options, whttp.WithRequestMessage(request))
		options = append(options, whttp.WithRequestType[BlockBaseRequest](whttp.RequestTypeBlockUsers))
	case BlockActionUnblock:
		method = http.MethodDelete
		options = append(options, whttp.WithRequestType[BlockBaseRequest](whttp.RequestTypeUnblockUsers))
		options = append(options, whttp.WithRequestMessage(request))
	case BlockActionListBlocked:
		method = http.MethodGet
		options = append(options, whttp.WithRequestType[BlockBaseRequest](whttp.RequestTypeListBlockedUsers))
		params := map[string]string{}
		if request.ListOptions != nil {
			if request.ListOptions.Limit != nil {
				params["limit"] = strconv.Itoa(*request.ListOptions.Limit)
			}
			if request.ListOptions.After != nil {
				params["after"] = *request.ListOptions.After
			}
			if request.ListOptions.Before != nil {
				params["before"] = *request.ListOptions.Before
			}

			if len(params) > 0 {
				options = append(options, whttp.WithRequestQueryParams[BlockBaseRequest](params))
			}
		}
	default:
		method = http.MethodGet
	}

	var (
		req      = whttp.MakeRequest(method, conf.BaseURL, options...)
		decoder  whttp.ResponseDecoderFunc
		response = &BlockBaseResponse{}
	)

	decoder = whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: true,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err = b.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}
