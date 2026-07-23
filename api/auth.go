//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the “Software”), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package api

import (
	"context"
	"fmt"

	"github.com/piusalfred/whatsapp/auth"
	"github.com/piusalfred/whatsapp/config"
)

func (c *Client) InstallApp(ctx context.Context, params *auth.InstallAppParams) error {
	return c.sender.InstallApp(ctx, c.config, params)
}

func (c *Client) GenerateSystemUserToken(
	ctx context.Context,
	params *auth.GenerateAccessTokenParams,
) (*auth.GenerateAccessTokenResponse, error) {
	resp, err := c.sender.GenerateAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("generate system user token: %w", err)
	}
	return resp, nil
}

func (c *Client) RevokeSystemUserToken(
	ctx context.Context,
	params *auth.RevokeAccessTokenParams,
) (*auth.RevokeAccessTokenResponse, error) {
	resp, err := c.sender.RevokeAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("revoke system user token: %w", err)
	}
	return resp, nil
}

func (c *Client) RefreshSystemUserToken(
	ctx context.Context,
	params *auth.RefreshAccessTokenParams,
) (*auth.RefreshAccessTokenResponse, error) {
	resp, err := c.sender.RefreshAccessToken(ctx, c.config, params)
	if err != nil {
		return nil, fmt.Errorf("refresh system user token: %w", err)
	}
	return resp, nil
}

func (c *Client) TwoStepVerification(
	ctx context.Context,
	request *auth.TwoStepVerificationRequest,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.TwoStepVerification(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("two step verification: %w", err)
	}
	return resp, nil
}

func (c *Client) CreateSystemUser(
	ctx context.Context,
	req *auth.CreateSystemUserRequest,
) (*auth.CreateSystemUserResponse, error) {
	resp, err := c.sender.CreateSystemUser(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}
	return resp, nil
}

func (c *Client) ListSystemUsers(ctx context.Context) (*auth.ListSystemUsersResponse, error) {
	resp, err := c.sender.ListSystemUsers(ctx, c.config)
	if err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateSystemUser(
	ctx context.Context,
	req *auth.UpdateSystemUserRequest,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.UpdateSystemUser(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}
	return resp, nil
}

func (c *Client) InvalidateSystemUserTokens(
	ctx context.Context,
	systemUserID string,
) (*auth.SuccessResponse, error) {
	resp, err := c.sender.InvalidateSystemUserTokens(ctx, c.config, systemUserID)
	if err != nil {
		return nil, fmt.Errorf("invalidate system user tokens: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) InstallApp(
	ctx context.Context,
	conf *config.Config,
	params *auth.InstallAppParams,
) error {
	if err := bc.getAuth().InstallApp(ctx, conf, params); err != nil {
		return fmt.Errorf("install app: %w", err)
	}
	return nil
}

func (bc *BaseClient) GenerateAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.GenerateAccessTokenParams,
) (*auth.GenerateAccessTokenResponse, error) {
	resp, err := bc.getAuth().GenerateAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RevokeAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.RevokeAccessTokenParams,
) (*auth.RevokeAccessTokenResponse, error) {
	resp, err := bc.getAuth().RevokeAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("revoke access token: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RefreshAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *auth.RefreshAccessTokenParams,
) (*auth.RefreshAccessTokenResponse, error) {
	resp, err := bc.getAuth().RefreshAccessToken(ctx, conf, params)
	if err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) TwoStepVerification(
	ctx context.Context,
	conf *config.Config,
	request *auth.TwoStepVerificationRequest,
) (*auth.SuccessResponse, error) {
	resp, err := bc.getAuth().TwoStepVerification(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("two step verification: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *auth.CreateSystemUserRequest,
) (*auth.CreateSystemUserResponse, error) {
	resp, err := bc.getAuth().CreateSystemUser(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListSystemUsers(
	ctx context.Context,
	conf *config.Config,
) (*auth.ListSystemUsersResponse, error) {
	resp, err := bc.getAuth().ListSystemUsers(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *auth.UpdateSystemUserRequest,
) (*auth.SuccessResponse, error) {
	resp, err := bc.getAuth().UpdateSystemUser(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) InvalidateSystemUserTokens(
	ctx context.Context,
	conf *config.Config,
	systemUserID string,
) (*auth.SuccessResponse, error) {
	resp, err := bc.getAuth().InvalidateSystemUserTokens(ctx, conf, systemUserID)
	if err != nil {
		return nil, fmt.Errorf("invalidate system user tokens: %w", err)
	}
	return resp, nil
}
