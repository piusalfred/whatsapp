package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/pkg/crypto"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type Client struct {
	baseURL    string
	apiVersion string
	sender     whttp.Sender[any]
}

func NewClient(baseURL, apiVersion string, sender whttp.Sender[any]) *Client {
	return &Client{
		baseURL:    baseURL,
		apiVersion: apiVersion,
		sender:     sender,
	}
}

// InstallAppParams contains the parameters required to install an app for a system user.
type InstallAppParams struct {
	AccessToken  string // The access token of an admin or system user who installs the app.
	AppID        string // The ID of the app being installed.
	SystemUserID string // The ID of the system user for whom the app is being installed.
}

// SuccessResponse represents the response for the InstallApp API call.
type SuccessResponse struct {
	Success bool `json:"success,omitempty"` // Indicates if the app installation was successful.
}

// InstallApp installs an app for a system user or an admin system user, allowing the app to make API calls
// on behalf of the user. Both the app and the system user should belong to the same Business Manager.
// Apps must have Ads Management API standard access or higher to be installed.
//
// Params:
// - ctx: The request context.
// - params: A struct containing:
//   - AccessToken: The access token of an admin user, admin system user, or another system user.
//   - AppID: The ID of the app being installed.
//   - BaseURL: The base URL of the API.
//   - APIVersion: The version of the API being called.
//   - SystemUserID: The ID of the system user on whose behalf the app is installed.
//
// Returns:
// - error: Any error encountered during the installation process.
func (c *Client) InstallApp(ctx context.Context, params InstallAppParams) error {
	req := &whttp.Request[any]{
		Type:      whttp.RequestTypeInstallApp,
		BaseURL:   c.baseURL,
		Method:    http.MethodPost,
		Endpoints: []string{c.apiVersion, params.SystemUserID, "applications"},
		QueryParams: map[string]string{
			"business_app": params.AppID,
			"access_token": params.AccessToken,
		},
	}

	res := &SuccessResponse{}

	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	err := c.sender.Send(ctx, req, decoder)
	if err != nil {
		return fmt.Errorf("install app: %w", err)
	}

	return nil
}

type TwoStepVerificationRequest struct {
	SixDigitCode  string `json:"pin"`
	PhoneNumberID string `json:"-"`
	AccessToken   string `json:"-"`
}

type TwoStepVerificationClient struct {
	baseURL    string
	apiVersion string
	sender     whttp.Sender[TwoStepVerificationRequest]
}

// TwoStepVerification sends a request to set up two-step verification for a WhatsApp Business API
// phone number.
func (c *TwoStepVerificationClient) TwoStepVerification(ctx context.Context,
	request *TwoStepVerificationRequest,
) (*SuccessResponse, error) {
	req := &whttp.Request[TwoStepVerificationRequest]{
		Type:      whttp.RequestTypeTwoStepVerification,
		BaseURL:   c.baseURL,
		Bearer:    request.AccessToken,
		Method:    http.MethodPost,
		Endpoints: []string{c.apiVersion, request.PhoneNumberID},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		// avoid leaking the phone number and access token.
		Message: &TwoStepVerificationRequest{SixDigitCode: request.SixDigitCode},
	}

	res := &SuccessResponse{}
	decodeOptions := whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	}

	decoder := whttp.ResponseDecoderJSON(res, decodeOptions)

	if err := c.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send two step verification code: %w", err)
	}

	return res, nil
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
	AccessToken string `json:"access_token"` // The newly generated access token.
}

// GenerateAccessToken generates a persistent access token for a system user.
// The system user must have installed the app beforehand. The generated access token is needed to make API
// calls on behalf of the system user.
//
// Params:
// - ctx: The request context.
// - params: A struct containing:
//   - AccessToken: The access token of the user generating the new token.
//   - AppID: The ID of the app being used for token generation.
//   - SystemUserID: The ID of the system user generating the token.
//   - BaseURL: The base URL of the API.
//   - APIVersion: The version of the API being called.
//   - AppSecret: The app secret for generating the app secret proof.
//   - Scopes: The scopes (permissions) required for the new access token.
//   - SetTokenExpiresIn60: Boolean flag to set the token expiration to 60 days.
//
// Returns:
// - *GenerateAccessTokenResponse: Contains the newly generated access token.
// - error: Any error encountered during the token generation process.
func (c *Client) GenerateAccessToken(ctx context.Context,
	params GenerateAccessTokenParams,
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

	req := &whttp.Request[any]{
		Type:        whttp.RequestTypeGenerateToken,
		BaseURL:     c.baseURL,
		Method:      http.MethodPost,
		Endpoints:   []string{c.apiVersion, params.SystemUserID, "access_tokens"},
		QueryParams: formData,
	}
	res := &GenerateAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := c.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return res, nil
}

// RevokeAccessTokenParams contains the parameters for revoking an access token.
type RevokeAccessTokenParams struct {
	ClientID        string
	ClientSecret    string
	RevokeToken     string // Access token to revoke
	AccessToken     string // Access token to identify the caller
	GraphAPIVersion string
}

type RevokeAccessTokenResponse struct {
	Success bool `json:"success"`
}

func (c *Client) RevokeAccessToken(ctx context.Context,
	params RevokeAccessTokenParams,
) (*RevokeAccessTokenResponse, error) {
	queryParams := map[string]string{
		"client_id":     params.ClientID,
		"client_secret": params.ClientSecret,
		"revoke_token":  params.RevokeToken,
		"access_token":  params.AccessToken,
	}

	req := &whttp.Request[any]{
		Type:        whttp.RequestTypeRevokeToken,
		BaseURL:     c.baseURL,
		Method:      http.MethodGet,
		Endpoints:   []string{c.apiVersion, "/oauth/revoke"},
		QueryParams: queryParams,
	}

	res := &RevokeAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := c.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("revoke access token: %w", err)
	}

	return res, nil
}

// RefreshAccessTokenParams contains the parameters for refreshing an access token.
type RefreshAccessTokenParams struct {
	ClientID            string
	ClientSecret        string
	FbExchangeToken     string // Current access token
	GraphAPIVersion     string
	SetTokenExpiresIn60 bool // Set to true to refresh for another 60 days
}

// RefreshAccessTokenResponse contains the response from the refresh token request.
type RefreshAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// RefreshAccessToken sends a request to refresh an expiring system user access token.
func (c *Client) RefreshAccessToken(ctx context.Context,
	params RefreshAccessTokenParams,
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

	opts := []whttp.RequestOption[any]{
		whttp.WithRequestType[any](whttp.RequestTypeRefreshToken),
		whttp.WithRequestEndpoints[any](c.apiVersion, "/oauth/access_token"),
		whttp.WithRequestQueryParams[any](queryParams),
	}

	req := whttp.MakeRequest[any](http.MethodGet, c.baseURL, opts...)

	res := &RefreshAccessTokenResponse{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: false,
		InspectResponseError:  true,
	})

	if err := c.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}

	return res, nil
}

type (
	TokenRotator interface {
		RotateToken(ctx context.Context, refresher TokenRefresher, revoker TokenRevoker, store TokenStore) error
	}

	TokenRotatorFunc func(ctx context.Context, refresher TokenRefresher, revoker TokenRevoker, store TokenStore) error
)

func (fn TokenRotatorFunc) RotateToken(ctx context.Context, refresher TokenRefresher,
	revoker TokenRevoker, store TokenStore,
) error {
	return fn(ctx, refresher, revoker, store)
}

// TokenRefresher defines an interface to refresh and fetch a new token.
type TokenRefresher interface {
	Refresh(ctx context.Context, currentToken string) (string, error) // Refreshes and returns the new token
}

// TokenRevoker defines an interface to revoke a token.
type TokenRevoker interface {
	Revoke(ctx context.Context, oldToken string) error // Revokes the old token
}

// TokenStore defines an interface to store the new token.
type TokenStore interface {
	Add(ctx context.Context, newToken string) error
	Get(ctx context.Context) (string, error)
}

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
	TokenScopeInstagramBasic                  = "instagram_basic" //nolint:gosec
	TokenScopeInstagramBrandedContentAdsBrand = "instagram_branded_content_ads_brand"
	TokenScopeInstagramBrandedContentBrand    = "instagram_branded_content_brand"
	TokenScopeInstagramContentPublish         = "instagram_content_publish" //nolint:gosec
	TokenScopeInstagramManageComments         = "instagram_manage_comments"
	TokenScopeInstagramManageInsights         = "instagram_manage_insights"
	TokenScopeInstagramManageMessages         = "instagram_manage_messages"
	TokenScopeInstagramShoppingTagProducts    = "instagram_shopping_tag_products" //nolint:gosec
	TokenScopeLeadsRetrieval                  = "leads_retrieval"
	TokenScopePageEvents                      = "page_events"
	TokenScopePagesManageAds                  = "pages_manage_ads"
	TokenScopePagesManageCta                  = "pages_manage_cta" //nolint:gosec
	TokenScopePagesManageEngagement           = "pages_manage_engagement"
	TokenScopePagesManageInstantArticles      = "pages_manage_instant_articles"
	TokenScopePagesManageMetadata             = "pages_manage_metadata"
	TokenScopePagesManagePosts                = "pages_manage_posts"
	TokenScopePagesMessaging                  = "pages_messaging"
	TokenScopePagesReadEngagement             = "pages_read_engagement"   //nolint:gosec
	TokenScopePagesReadUserContent            = "pages_read_user_content" //nolint:gosec
	TokenScopePagesShowList                   = "pages_show_list"
	TokenScopePrivateComputationAccess        = "private_computation_access"
	TokenScopePublishVideo                    = "publish_video"
	TokenScopeReadAudienceNetworkInsights     = "read_audience_network_insights"
	TokenScopeReadInsights                    = "read_insights"
	TokenScopeReadPageMailboxes               = "read_page_mailboxes" //nolint:gosec
	TokenScopeWhatsappBusinessManagement      = "whatsapp_business_management"
	TokenScopeWhatsappBusinessMessaging       = "whatsapp_business_messaging"
)
