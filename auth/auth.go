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
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package auth provides a client for the Meta system user and access token API.
//
// # Lifecycle
//
// System users follow a least-privilege pattern enforced by Meta:
//
//	REAL ADMIN (human)
//	  │
//	  ├─ 1. Create ADMIN SYSTEM USER ─── one per business, kept safe
//	  │     │
//	  │     ├─ 2. Install app on it ──── TOS acceptance (prerequisite for tokens)
//	  │     ├─ 3. Generate its token ─── now you can automate without the human token
//	  │     │
//	  │     └─ 4. Create REGULAR SYSTEM USER ─── one per access type, scoped
//	  │           │
//	  │           ├─ 5. Install app on it ──── same TOS step
//	  │           ├─ 6. Grant asset permissions ─── scoped to what it needs
//	  │           └─ 7. Generate its token ─── use this for daily API calls
//	  │
//	  └─ 8. Invalidate tokens ─── security escape hatch (can't delete users)
//
// Use the admin system user token only to manage other system users.
// Use regular system user tokens for all API calls. This way, if a token
// is compromised, the blast radius is limited to its scope.
//
// # Limits
//
// Standard access: 1 admin system user + 1 regular system user.
// Advanced access: 1 admin system user + 10 regular system users.
//
// # Token Lifecycle
//
// Tokens expire. Use [RefreshAccessToken] to extend without creating a new
// user, [RevokeAccessToken] to kill a single token, or
// [InvalidateSystemUserTokens] to kill all tokens for a user.
// [RotateAccessToken] provides atomic rotation: refresh → store new → revoke
// old, with no downtime.
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

// GenerateAccessTokenParams contains the parameters required to generate an
// access token for a system user.
//
// Non-expiring tokens are deprecated — Meta recommends expiring tokens.
// Set [GenerateAccessTokenParams.SetTokenExpiresIn60] to true; omitting it
// may result in an error for businesses required to use expiring tokens.
type GenerateAccessTokenParams struct {
	AccessToken         string   // Token of the admin/system user generating this token.
	AppID               string   // App to associate the token with.
	SystemUserID        string   // System user receiving the token.
	AppSecret           string   // App secret used to compute appsecret_proof.
	Scopes              []string // Permissions to grant (e.g., whatsapp_business_messaging).
	SetTokenExpiresIn60 bool     // Set to true for a 60-day expiring token (recommended).
}

// GenerateAccessTokenResponse represents the response from generating an access token.
type GenerateAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// RevokeAccessTokenParams contains the parameters for revoking an access token.
// Used for token rotation (refresh → deploy new → revoke old) or as a security
// measure when a token is compromised. Revocation is immediate and cannot be
// undone.
type RevokeAccessTokenParams struct {
	ClientID     string // App ID.
	ClientSecret string // App secret.
	RevokeToken  string // The access token to revoke.
	AccessToken  string // A valid token identifying the caller.
}

// RevokeAccessTokenResponse represents the response from revoking an access token.
type RevokeAccessTokenResponse struct {
	Success bool `json:"success"`
}

// RefreshAccessTokenParams contains the parameters for refreshing an access token.
// An expiring token is valid for 60 days from creation or last refresh. You must
// refresh within 60 days or the token is forfeited and a new one must be generated.
type RefreshAccessTokenParams struct {
	ClientID            string // App ID.
	ClientSecret        string // App secret.
	FbExchangeToken     string // The current, still-valid access token to refresh.
	SetTokenExpiresIn60 bool   // Set to true to extend for another 60 days.
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

// InstallApp accepts the TOS for an app on behalf of a system user. This is a
// prerequisite for [GenerateAccessToken] — tokens cannot be generated for a
// user until the app is installed. Both the app and the system user must
// belong to the same Business Manager.
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
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
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
		Flags: whttp.JSONDecodeDisallowEmptyResponse | whttp.JSONDecodeInspectResponseError,
	}
	decoder := whttp.ResponseDecoderJSON(res, decodeOptions)

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send two step verification code: %w", err)
	}

	return res, nil
}

// GenerateAccessToken generates a persistent access token for a system user.
// The system user must have installed the app via [InstallApp] beforehand.
// Use admin system user tokens only to manage other users; use regular
// system user tokens for all API calls.
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
		Flags: whttp.JSONDecodePermissive,
	})

	if err = bc.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return res, nil
}

// RevokeAccessToken revokes a single access token. For bulk invalidation of
// all tokens belonging to a system user, use [InvalidateSystemUserTokens]
// instead.
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
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("revoke access token: %w", err)
	}

	return res, nil
}

// RefreshAccessToken sends a request to refresh an expiring system user access
// token. Tokens expire 60 days after creation or last refresh — failing to
// refresh within that window forfeits the token and a new one must be generated
// via [GenerateAccessToken].
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
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}

	return res, nil
}

// CreateSystemUser creates a system user in a business manager. Create one
// admin system user first (role="ADMIN"), then create regular system users
// (role="EMPLOYEE") for each access scope you need. This limits the blast
// radius if a token is compromised.
//

func (bc *BaseClient) CreateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *CreateSystemUserRequest,
) (*CreateSystemUserResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeCreateSystemUser).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users").
		QueryParams(map[string]string{
			"name": req.Name,
			"role": req.Role,
		})

	httpReq := whttp.BuildAnyRequest(b)

	res := &CreateSystemUserResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, httpReq, decoder); err != nil {
		return nil, fmt.Errorf("create system user: %w", err)
	}

	return res, nil
}

// ListSystemUsers retrieves all system users in a business manager. Useful
// for auditing who exists and finding app-scoped IDs.
func (bc *BaseClient) ListSystemUsers(
	ctx context.Context,
	conf *config.Config,
) (*ListSystemUsersResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodGet, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeListSystemUsers).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users")

	req := whttp.BuildAnyRequest(b)

	res := &ListSystemUsersResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("list system users: %w", err)
	}

	return res, nil
}

// UpdateSystemUser renames an existing system user. The role cannot be
// changed — create a new system user with the desired role instead.
//

func (bc *BaseClient) UpdateSystemUser(
	ctx context.Context,
	conf *config.Config,
	req *UpdateSystemUserRequest,
) (*SuccessResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodPost, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeUpdateSystemUser).
		Endpoints(conf.APIVersion, conf.BusinessAccountID, "system_users").
		QueryParams(map[string]string{
			"system_user_id": req.SystemUserID,
			"name":           req.Name,
		})

	httpReq := whttp.BuildAnyRequest(b)

	res := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, httpReq, decoder); err != nil {
		return nil, fmt.Errorf("update system user: %w", err)
	}

	return res, nil
}

// InvalidateSystemUserTokens invalidates all access tokens for a system user.
// This is the security escape hatch — system users cannot be deleted, but you
// can kill all their tokens. After invalidation, generate new tokens via
// [GenerateAccessToken].
func (bc *BaseClient) InvalidateSystemUserTokens(
	ctx context.Context,
	conf *config.Config,
	systemUserID string,
) (*SuccessResponse, error) {
	b := whttp.NewRequestBuilder(http.MethodDelete, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(whttp.RequestTypeInvalidateSystemUserTokens).
		Endpoints(conf.APIVersion, systemUserID, "access_tokens")

	req := whttp.BuildAnyRequest(b)

	res := &SuccessResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		Flags: whttp.JSONDecodePermissive,
	})

	if err := bc.BaseClient.Send(ctx, req, decoder); err != nil {
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

// RotateAccessToken performs atomic token rotation with no downtime:
//  1. Read the current token from the store
//  2. Refresh it to get a new token (old token still works)
//  3. Store the new token
//  4. Revoke the old token (immediate invalidation)
//
// This pattern limits the damage to leaked tokens, the old token is revoked as soon
// as the new one is deployed.
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
