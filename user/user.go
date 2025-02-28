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

//go:generate mockgen -destination=../mocks/user/mock_user.go -package=user -source=user.go

const BlockEndpoint = "block_users"

type BlockAction string

const (
	BlockActionBlockUsers   BlockAction = "block"
	BlockActionUnblockUsers BlockAction = "unblock"
	BlockActionListBlocked  BlockAction = "list"
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

	BlockUsersResponse struct {
		MessagingProduct string         `json:"messaging_product"`
		BlockUsers       *BlockUsers    `json:"block_users"`
		Error            *werrors.Error `json:"error,omitempty"`
	}

	BlockUsersService interface {
		Block(ctx context.Context, numbers []string) (*BlockUsersResponse, error)
		Unblock(ctx context.Context, numbers []string) (*BlockUsersResponse, error)
		ListBlocked(ctx context.Context, request *ListBlockedUsersOptions) (*GetBlockedUsersResponse, error)
	}

	BlockUsersBaseClient struct {
		Config config.Reader
		Sender whttp.Sender[BlockUsersBaseRequest]
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

	GetBlockedUsersResponse struct {
		Data   []BlockedUsersData `json:"data"`
		Paging *whttp.Paging      `json:"paging,omitempty"`
	}

	BlockedUsersData struct {
		BlockUsers []BlockResult `json:"block_users"`
	}

	BlockUsersBaseRequest struct {
		MessagingProduct string                   `json:"messaging_product,omitempty"`
		BlockUsers       []Identifier             `json:"block_users,omitempty"`
		BlockAction      BlockAction              `json:"-"`
		ListOptions      *ListBlockedUsersOptions `json:"-"`
	}

	BlockUserBaseResponse struct {
		Data   []BlockedUsersData `json:"data"`
		Paging *whttp.Paging      `json:"paging,omitempty"`
	}
	BlockUsersBaseRequestOption func(*BlockUsersBaseRequest)
)

func (b *BlockUsersBaseClient) Unblock(ctx context.Context, numbers []string) (*BlockUsersResponse, error) {
	req := NewBlockUsersBaseRequest(BlockActionUnblockUsers, WithBlockUsersBaseRequestNumbers(numbers))

	resp, err := b.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to unblock users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (b *BlockUsersBaseClient) ListBlocked(ctx context.Context, request *ListBlockedUsersOptions) (
	*GetBlockedUsersResponse, error,
) {
	req := NewBlockUsersBaseRequest(BlockActionListBlocked, func(r *BlockUsersBaseRequest) {
		r.ListOptions = request
	})

	resp, err := b.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked users: %w", err)
	}

	return resp.BlockUsersListResponse(), nil
}

func (b *BlockUsersBaseClient) Block(ctx context.Context, numbers []string) (*BlockUsersResponse, error) {
	req := NewBlockUsersBaseRequest(BlockActionBlockUsers, WithBlockUsersBaseRequestNumbers(numbers))

	resp, err := b.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to block users: %w", err)
	}

	return resp.BlockUsersResponse(), nil
}

func (b *BlockUsersBaseClient) Do(ctx context.Context, request *BlockUsersBaseRequest) (*BlockUserBaseResponse, error) {
	var method string

	conf, err := b.Config.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	options := []whttp.RequestOption[BlockUsersBaseRequest]{
		whttp.WithRequestEndpoints[BlockUsersBaseRequest](conf.APIVersion, conf.PhoneNumberID, BlockEndpoint),
		whttp.WithRequestBearer[BlockUsersBaseRequest](conf.AccessToken),
		whttp.WithRequestAppSecret[BlockUsersBaseRequest](conf.AppSecret),
		whttp.WithRequestSecured[BlockUsersBaseRequest](conf.SecureRequests),
	}

	switch request.BlockAction {
	case BlockActionBlockUsers:
		method = http.MethodPost
		options = append(options, whttp.WithRequestMessage[BlockUsersBaseRequest](request))
		options = append(options, whttp.WithRequestType[BlockUsersBaseRequest](whttp.RequestTypeBlockUsers))
	case BlockActionUnblockUsers:
		method = http.MethodDelete
		options = append(options, whttp.WithRequestType[BlockUsersBaseRequest](whttp.RequestTypeUnblockUsers))
		options = append(options, whttp.WithRequestMessage[BlockUsersBaseRequest](request))
	case BlockActionListBlocked:
		method = http.MethodGet
		options = append(options, whttp.WithRequestType[BlockUsersBaseRequest](whttp.RequestTypeListBlockedUsers))
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
				options = append(options, whttp.WithRequestQueryParams[BlockUsersBaseRequest](params))
			}
		}
	default:
		method = http.MethodGet
	}

	var (
		req      = whttp.MakeRequest[BlockUsersBaseRequest](method, conf.BaseURL, options...)
		decoder  whttp.ResponseDecoderFunc
		response = &BlockUserBaseResponse{}
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

var _ BlockUsersService = (*BlockUsersBaseClient)(nil)

func (base *BlockUserBaseResponse) BlockUsersResponse() *BlockUsersResponse {
	return &BlockUsersResponse{}
}

func (base *BlockUserBaseResponse) BlockUsersListResponse() *GetBlockedUsersResponse {
	return &GetBlockedUsersResponse{
		Data:   nil,
		Paging: nil,
	}
}

func WithBlockUsersBaseRequestNumbers(numbers []string) BlockUsersBaseRequestOption {
	return func(r *BlockUsersBaseRequest) {
		for _, number := range numbers {
			r.BlockUsers = append(r.BlockUsers, Identifier{User: number})
		}
	}
}

func WithBlockUsersBaseRequestListLimit(limit int) BlockUsersBaseRequestOption {
	return func(r *BlockUsersBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.Limit = &limit
	}
}

func WithBlockUsersBaseRequestListAfter(after string) BlockUsersBaseRequestOption {
	return func(r *BlockUsersBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.After = &after
	}
}

func WithBlockUsersBaseRequestListBefore(before string) BlockUsersBaseRequestOption {
	return func(r *BlockUsersBaseRequest) {
		if r.ListOptions == nil {
			r.ListOptions = &ListBlockedUsersOptions{}
		}

		r.ListOptions.Before = &before
	}
}

func NewBlockUsersBaseRequest(action BlockAction, options ...BlockUsersBaseRequestOption) *BlockUsersBaseRequest {
	b := &BlockUsersBaseRequest{
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
