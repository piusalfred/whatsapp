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

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/pkg/crypto"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

// InstallAppParams contains the parameters required to install an app for a system user.
type InstallAppParams struct {
	AccessToken  string // The access token of an admin or system user who installs the app.
	AppID        string // The ID of the app being installed.
	SystemUserID string // The ID of the system user for whom the app is being installed.
}

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Success bool `json:"success,omitempty"`
}

// TwoStepVerificationRequest contains the parameters for setting up two-step verification
// on a WhatsApp Business API phone number.
type TwoStepVerificationRequest struct {
	SixDigitCode  string `json:"pin"`
	PhoneNumberID string `json:"-"`
	AccessToken   string `json:"-"`
}

// GenerateAccessTokenParams contains the parameters required to generate a persistent access token.
type GenerateAccessTokenParams struct {
	AccessToken         string   // The access token of the user generating the new access token.
	AppID               string   // The ID of the app for which the token is generated.
	SystemUserID        string   // The system user ID that is generating the token.
	AppSecret           string   // The app secret associated with the app.
	Scopes              []string // A list of permissions (scopes) to be granted to the new token.
	SetTokenExpiresIn60 bool     // If true, sets the token to expire in 60 days.
}

// GenerateAccessTokenResponse represents the response from generating an access token.
type GenerateAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// RevokeAccessTokenParams contains the parameters for revoking an access token.
type RevokeAccessTokenParams struct {
	ClientID     string
	ClientSecret string
	RevokeToken  string // Access token to revoke.
	AccessToken  string // Access token to identify the caller.
}

// RevokeAccessTokenResponse represents the response from revoking an access token.
type RevokeAccessTokenResponse struct {
	Success bool `json:"success"`
}

// RefreshAccessTokenParams contains the parameters for refreshing an access token.
type RefreshAccessTokenParams struct {
	ClientID            string
	ClientSecret        string
	FbExchangeToken     string // Current access token.
	SetTokenExpiresIn60 bool   // Set to true to refresh for another 60 days.
}

// RefreshAccessTokenResponse contains the response from a refresh token request.
type RefreshAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// CreateSystemUserRequest contains the parameters for creating a system user.
type CreateSystemUserRequest struct {
	Name string // Display name for the system user.
	Role string // Role for the system user: "ADMIN" or "EMPLOYEE".
}

// CreateSystemUserResponse represents the response from creating a system user.
type CreateSystemUserResponse struct {
	ID string `json:"id"`
}

// SystemUser represents a system user in the business manager.
type SystemUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ListSystemUsersResponse represents the response from listing system users.
type ListSystemUsersResponse struct {
	Data []SystemUser `json:"data"`
}

// UpdateSystemUserRequest contains the parameters for updating a system user's name.
type UpdateSystemUserRequest struct {
	SystemUserID string // The ID of the system user to update.
	Name         string // The new name for the system user.
}

// BaseClient is the low-level HTTP executor for auth operations. It builds requests
// using [whttp.NewRequestBuilder] and sends them via the embedded [whttp.BaseClient].
type BaseClient struct {
	whttp.BaseClient[any]
}

// InstallApp installs an app for a system user or an admin system user, allowing the app to make API calls
// on behalf of the user. Both the app and the system user should belong to the same Business Manager.
func (bc *BaseClient) InstallApp(ctx context.Context, conf *config.Config, params *InstallAppParams) error {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Type(whttp.RequestTypeInstallApp).
		Endpoints(conf.APIVersion, params.SystemUserID, "applications").
		QueryParams(map[string]string{
			"business_app": params.AppID,
			"access_token": params.AccessToken,
		}).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return fmt.Errorf("install app: %w", err)
	}

	return nil
}

// TwoStepVerification sends a request to set up two-step verification for a WhatsApp Business API
// phone number.
func (bc *BaseClient) TwoStepVerification(
	ctx context.Context,
	conf *config.Config,
	request *TwoStepVerificationRequest,
) (*SuccessResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Bearer(request.AccessToken).
		Type(whttp.RequestTypeTwoStepVerification).
		Endpoints(conf.APIVersion, request.PhoneNumberID).
		Headers(map[string]string{
			"Content-Type": "application/json",
		}).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	msg := any(&TwoStepVerificationRequest{SixDigitCode: request.SixDigitCode})
	req := whttp.Build[any](b, &msg)

	res := &SuccessResponse{}
	decodeOptions := whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	}
	decoder := whttp.ResponseDecoderJSON(res, decodeOptions)

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send two step verification code: %w", err)
	}

	return res, nil
}

// GenerateAccessToken generates a persistent access token for a system user.
// The system user must have installed the app beforehand.
func (bc *BaseClient) GenerateAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *GenerateAccessTokenParams,
) (*GenerateAccessTokenResponse, error) {
	appSecretProof, err := crypto.GenerateAppSecretProof(params.AccessToken, params.AppSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app secret proof: %w", err)
	}

	formData := map[string]string{
		"business_app":    params.AppID,
		"appsecret_proof": appSecretProof,
		"access_token":    params.AccessToken,
		"scope":           strings.Join(params.Scopes, ","),
	}

	if params.SetTokenExpiresIn60 {
		formData["set_token_expires_in_60_days"] = "true"
	}

	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Type(whttp.RequestTypeGenerateToken).
		Endpoints(conf.APIVersion, params.SystemUserID, "access_tokens").
		QueryParams(formData).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &GenerateAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err = bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return res, nil
}

// RevokeAccessToken revokes an access token.
func (bc *BaseClient) RevokeAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *RevokeAccessTokenParams,
) (*RevokeAccessTokenResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Type(whttp.RequestTypeRevokeToken).
		Endpoints(conf.APIVersion, "oauth/revoke").
		QueryParams(map[string]string{
			"client_id":     params.ClientID,
			"client_secret": params.ClientSecret,
			"revoke_token":  params.RevokeToken,
			"access_token":  params.AccessToken,
		}).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &RevokeAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("revoke access token: %w", err)
	}

	return res, nil
}

// RefreshAccessToken sends a request to refresh an expiring system user access token.
func (bc *BaseClient) RefreshAccessToken(
	ctx context.Context,
	conf *config.Config,
	params *RefreshAccessTokenParams,
) (*RefreshAccessTokenResponse, error) {
	queryParams := map[string]string{
		"grant_type":        "fb_exchange_token",
		"client_id":         params.ClientID,
		"client_secret":     params.ClientSecret,
		"fb_exchange_token": params.FbExchangeToken,
	}

	if params.SetTokenExpiresIn60 {
		queryParams["set_token_expires_in_60_days"] = "true"
	}

	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Type(whttp.RequestTypeRefreshToken).
		Endpoints(conf.APIVersion, "oauth/access_token").
		QueryParams(queryParams).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &RefreshAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}

	return res, nil
}

// CreateSystemUser creates a system user in a business manager.
//
//nolint:dupl // similar structure to UpdateSystemUser but different request type, params, and response
func (bc *BaseClient) CreateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *CreateSystemUserRequest,
) (*CreateSystemUserResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		Type(whttp.RequestTypeCreateSystemUser).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users").
		QueryParams(map[string]string{
			"name": req.Name,
			"role": req.Role,
		}).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	httpReq := whttp.BuildAnyRequest(b)

	res := &CreateSystemUserResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, httpReq, decoder); err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}

	return res, nil
}

// ListSystemUsers retrieves all system users in a business manager.
func (bc *BaseClient) ListSystemUsers(
	ctx context.Context,
	conf *config.Config,
) (*ListSystemUsersResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		Type(whttp.RequestTypeListSystemUsers).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users").
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &ListSystemUsersResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}

	return res, nil
}

// UpdateSystemUser updates the name of an existing system user.
//
//nolint:dupl // similar structure to CreateSystemUser but different request type, params, and response
func (bc *BaseClient) UpdateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *UpdateSystemUserRequest,
) (*SuccessResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		Type(whttp.RequestTypeUpdateSystemUser).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users").
		QueryParams(map[string]string{
			"system_user_id": req.SystemUserID,
			"name":           req.Name,
		}).
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	httpReq := whttp.BuildAnyRequest(b)

	res := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, httpReq, decoder); err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}

	return res, nil
}

// InvalidateSystemUserTokens invalidates all access tokens for a system user.
func (bc *BaseClient) InvalidateSystemUserTokens(
	ctx context.Context,
	conf *config.Config,
	systemUserID string,
) (*SuccessResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodDelete, conf.BaseURL).
		Bearer(conf.AccessToken).
		AppSecret(conf.AppSecret).Secured(conf.SecureRequests).
		Type(whttp.RequestTypeInvalidateSystemUserTokens).
		Endpoints(conf.APIVersion, systemUserID, "access_tokens").
		DebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel))

	req := whttp.BuildAnyRequest(b)

	res := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("invalidate system user tokens: %w", err)
	}

	return res, nil
}

// Client wraps BaseClient with a fixed configuration.
type Client struct {
	sender *BaseClient
	config *config.Config
}

// NewClient creates a Client with a fixed configuration and sender options.
func NewClient(conf *config.Config, opts ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[any](opts...)},
		config: conf,
	}
}

// InstallApp installs an app for a system user, using the client's fixed configuration.
func (c *Client) InstallApp(ctx context.Context, params *InstallAppParams) error {
	return c.sender.InstallApp(ctx, c.config, params)
}

// TwoStepVerification sets up two-step verification for a WhatsApp Business API phone number.
func (c *Client) TwoStepVerification(
	ctx context.Context,
	request *TwoStepVerificationRequest,
) (*SuccessResponse, error) {
	return c.sender.TwoStepVerification(ctx, c.config, request)
}

// GenerateAccessToken generates a persistent access token for a system user.
func (c *Client) GenerateAccessToken(
	ctx context.Context,
	params *GenerateAccessTokenParams,
) (*GenerateAccessTokenResponse, error) {
	return c.sender.GenerateAccessToken(ctx, c.config, params)
}

// RevokeAccessToken revokes an access token.
func (c *Client) RevokeAccessToken(
	ctx context.Context,
	params *RevokeAccessTokenParams,
) (*RevokeAccessTokenResponse, error) {
	return c.sender.RevokeAccessToken(ctx, c.config, params)
}

// RefreshAccessToken sends a request to refresh an expiring system user access token.
func (c *Client) RefreshAccessToken(
	ctx context.Context,
	params *RefreshAccessTokenParams,
) (*RefreshAccessTokenResponse, error) {
	return c.sender.RefreshAccessToken(ctx, c.config, params)
}

// CreateSystemUser creates a system user in a business manager.
func (c *Client) CreateSystemUser(
	ctx context.Context,
	req *CreateSystemUserRequest,
) (*CreateSystemUserResponse, error) {
	return c.sender.CreateSystemUser(ctx, c.config, req)
}

// ListSystemUsers retrieves all system users in a business manager.
func (c *Client) ListSystemUsers(ctx context.Context) (*ListSystemUsersResponse, error) {
	return c.sender.ListSystemUsers(ctx, c.config)
}

// UpdateSystemUser updates the name of an existing system user.
func (c *Client) UpdateSystemUser(
	ctx context.Context,
	req *UpdateSystemUserRequest,
) (*SuccessResponse, error) {
	return c.sender.UpdateSystemUser(ctx, c.config, req)
}

// InvalidateSystemUserTokens invalidates all access tokens for a system user.
func (c *Client) InvalidateSystemUserTokens(
	ctx context.Context,
	systemUserID string,
) (*SuccessResponse, error) {
	return c.sender.InvalidateSystemUserTokens(ctx, c.config, systemUserID)
}

type (
	TokenRotatorFunc func(ctx context.Context, refresher TokenRefresher, revoker TokenRevoker, store TokenStore) error

	TokenRotator interface {
		RotateToken(ctx context.Context, refresher TokenRefresher, revoker TokenRevoker, store TokenStore) error
	}
)

func (fn TokenRotatorFunc) RotateToken(ctx context.Context, refresher TokenRefresher,
	revoker TokenRevoker, store TokenStore,
) error {
	return fn(ctx, refresher, revoker, store)
}

// TokenRefresher defines an interface to refresh and fetch a new token.
type TokenRefresher interface {
	Refresh(ctx context.Context, currentToken string) (string, error)
}

// TokenRevoker defines an interface to revoke a token.
type TokenRevoker interface {
	Revoke(ctx context.Context, oldToken string) error
}

// TokenStore defines an interface to store the new token.
type TokenStore interface {
	Add(ctx context.Context, newToken string) error
	Get(ctx context.Context) (string, error)
}

// RotateAccessToken refreshes, stores the new token, and revokes the old token.
func RotateAccessToken(
	ctx context.Context,
	refresher TokenRefresher,
	revoker TokenRevoker,
	store TokenStore,
) error {
	oldToken, err := store.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	newToken, err := refresher.Refresh(ctx, oldToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	err = store.Add(ctx, newToken)
	if err != nil {
		return fmt.Errorf("failed to store new token: %w", err)
	}

	err = revoker.Revoke(ctx, oldToken)
	if err != nil {
		return fmt.Errorf("failed to revoke old token: %w", err)
	}

	return nil
}

const (
	TokenScopeAdsManagement                   = "ads_management"
	TokenScopeAdsRead                         = "ads_read"
	TokenScopeAttributionRead                 = "attribution_read"
	TokenScopeBusinessManagement              = "business_management"
	TokenScopeCatalogManagement               = "catalog_management"
	TokenScopeCommerceAccountManageOrders     = "commerce_account_manage_orders"
	TokenScopeCommerceAccountReadOrders       = "commerce_account_read_orders"
	TokenScopeCommerceAccountReadSettings     = "commerce_account_read_settings"
	TokenScopeInstagramBasic                  = "instagram_basic" //nolint:gosec // ok
	TokenScopeInstagramBrandedContentAdsBrand = "instagram_branded_content_ads_brand"
	TokenScopeInstagramBrandedContentBrand    = "instagram_branded_content_brand"
	TokenScopeInstagramContentPublish         = "instagram_content_publish" //nolint:gosec // ok
	TokenScopeInstagramManageComments         = "instagram_manage_comments"
	TokenScopeInstagramManageInsights         = "instagram_manage_insights"
	TokenScopeInstagramManageMessages         = "instagram_manage_messages"
	TokenScopeInstagramShoppingTagProducts    = "instagram_shopping_tag_products" //nolint:gosec // ok
	TokenScopeLeadsRetrieval                  = "leads_retrieval"
	TokenScopePageEvents                      = "page_events"
	TokenScopePagesManageAds                  = "pages_manage_ads"
	TokenScopePagesManageCta                  = "pages_manage_cta" //nolint:gosec // ok
	TokenScopePagesManageEngagement           = "pages_manage_engagement"
	TokenScopePagesManageInstantArticles      = "pages_manage_instant_articles"
	TokenScopePagesManageMetadata             = "pages_manage_metadata"
	TokenScopePagesManagePosts                = "pages_manage_posts"
	TokenScopePagesMessaging                  = "pages_messaging"
	TokenScopePagesReadEngagement             = "pages_read_engagement"   //nolint:gosec // ok
	TokenScopePagesReadUserContent            = "pages_read_user_content" //nolint:gosec // ok
	TokenScopePagesShowList                   = "pages_show_list"
	TokenScopePrivateComputationAccess        = "private_computation_access"
	TokenScopePublishVideo                    = "publish_video"
	TokenScopeReadAudienceNetworkInsights     = "read_audience_network_insights"
	TokenScopeReadInsights                    = "read_insights"
	TokenScopeReadPageMailboxes               = "read_page_mailboxes" //nolint:gosec // ok
	TokenScopeWhatsappBusinessManagement      = "whatsapp_business_management"
	TokenScopeWhatsappBusinessMessaging       = "whatsapp_business_messaging"
)
