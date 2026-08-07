//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
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

package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/piusalfred/whatsapp/auth"
	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/internal/test"
)

func mockConfig(baseURL string) *config.Config {
	return &config.Config{
		BaseURL:           baseURL,
		APIVersion:        "v25.0",
		PhoneNumberID:     "106540352242922",
		BusinessAccountID: "123456789",
		AccessToken:       "EAAJB...",
		AppSecret:         "appsecret",
		SecureRequests:    false,
	}
}

func TestCreateSystemUser(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    []byte(`{"id":"100000008899900"}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.CreateSystemUser(context.Background(), &auth.CreateSystemUserRequest{
		Name: "Ad Server",
		Role: "ADMIN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "100000008899900" {
		t.Errorf("expected id=100000008899900, got %s", resp.ID)
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("name") != "Ad Server" {
		t.Errorf("expected name=Ad Server, got %s", r.QueryParams.Get("name"))
	}
	if r.QueryParams.Get("role") != "ADMIN" {
		t.Errorf("expected role=ADMIN, got %s", r.QueryParams.Get("role"))
	}
}

func TestListSystemUsers(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"data": []map[string]string{
			{"id": "1000081799813", "name": "Reporting server", "role": "ADMIN"},
		},
	})
	srv := test.NewMockServer(test.MockBehavior{StatusCode: http.StatusOK, Payload: payload})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.ListSystemUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "1000081799813" {
		t.Fatalf("unexpected response: %+v", resp.Data)
	}

	r := srv.GetRequests()[0]
	if r.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", r.Method)
	}
}

func TestUpdateSystemUser(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.UpdateSystemUser(context.Background(), &auth.UpdateSystemUserRequest{
		SystemUserID: "1000081799813",
		Name:         "FBX Server",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("system_user_id") != "1000081799813" {
		t.Errorf("unexpected system_user_id: %s", r.QueryParams.Get("system_user_id"))
	}
}

func TestInvalidateSystemUserTokens(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK,
		Payload:    []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.InvalidateSystemUserTokens(context.Background(), "1000081799813")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.Method != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", r.Method)
	}
}

func TestGenerateAccessToken(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]string{"access_token": "CAAB3rQQ..."})
	srv := test.NewMockServer(test.MockBehavior{StatusCode: http.StatusOK, Payload: payload})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.GenerateAccessToken(context.Background(), &auth.GenerateAccessTokenParams{
		AccessToken:         "admin-token",
		AppID:               "1243595696",
		SystemUserID:        "1000081799813",
		AppSecret:           "appsecret",
		Scopes:              []string{"whatsapp_business_messaging", "whatsapp_business_management"},
		SetTokenExpiresIn60: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "CAAB3rQQ..." {
		t.Errorf("unexpected token: %s", resp.AccessToken)
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("scope") != "whatsapp_business_messaging,whatsapp_business_management" {
		t.Errorf("unexpected scope: %s", r.QueryParams.Get("scope"))
	}
	if r.QueryParams.Get("set_token_expires_in_60_days") != "true" {
		t.Errorf("expected set_token_expires_in_60_days=true")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"access_token": "new-token", "token_type": "bearer", "expires_in": 5183944,
	})
	srv := test.NewMockServer(test.MockBehavior{StatusCode: http.StatusOK, Payload: payload})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.RefreshAccessToken(context.Background(), &auth.RefreshAccessTokenParams{
		ClientID: "app-id", ClientSecret: "app-secret",
		FbExchangeToken: "current-token", SetTokenExpiresIn60: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "new-token" || resp.ExpiresIn != 5183944 {
		t.Errorf("unexpected response: token=%s expires=%d", resp.AccessToken, resp.ExpiresIn)
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("grant_type") != "fb_exchange_token" {
		t.Errorf("unexpected grant_type: %s", r.QueryParams.Get("grant_type"))
	}
}

func TestRevokeAccessToken(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK, Payload: []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.RevokeAccessToken(context.Background(), &auth.RevokeAccessTokenParams{
		ClientID: "app-id", ClientSecret: "app-secret",
		RevokeToken: "old-token", AccessToken: "caller-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("revoke_token") != "old-token" {
		t.Errorf("unexpected revoke_token: %s", r.QueryParams.Get("revoke_token"))
	}
}

func TestAssignAdAccountPermissions(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK, Payload: []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.AssignAdAccountPermissions(context.Background(), &auth.AssignAdAccountPermissionsRequest{
		SystemUserID: "1000081799813",
		Tasks:        []string{auth.AdAccountTaskManage, auth.AdAccountTaskAdvertise},
		AdAccountID:  "123456789",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("user") != "1000081799813" {
		t.Errorf("expected user=1000081799813, got %s", r.QueryParams.Get("user"))
	}
	if r.QueryParams.Get("tasks") != "MANAGE,ADVERTISE" {
		t.Errorf("expected tasks=MANAGE,ADVERTISE, got %s", r.QueryParams.Get("tasks"))
	}
	if r.QueryParams.Get("business") != "123456789" {
		t.Errorf("expected business=123456789, got %s", r.QueryParams.Get("business"))
	}
}

func TestAssignPagePermissions(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK, Payload: []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.AssignPagePermissions(context.Background(), &auth.AssignPagePermissionsRequest{
		SystemUserID: "1000081799813",
		Tasks:        []string{auth.PageTaskAdvertise, auth.PageTaskAnalyze},
		PageID:       "175026500996648",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("tasks") != "ADVERTISE,ANALYZE" {
		t.Errorf("expected tasks=ADVERTISE,ANALYZE, got %s", r.QueryParams.Get("tasks"))
	}
}

func TestClaimAppForBusiness(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusOK, Payload: []byte(`{"success":true}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	resp, err := client.ClaimAppForBusiness(context.Background(), &auth.ClaimAppForBusinessRequest{
		AppID: "123456789", AccessType: "OWNER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}

	r := srv.GetRequests()[0]
	if r.QueryParams.Get("app_id") != "123456789" {
		t.Errorf("expected app_id=123456789, got %s", r.QueryParams.Get("app_id"))
	}
	if r.QueryParams.Get("access_type") != "OWNER" {
		t.Errorf("expected access_type=OWNER, got %s", r.QueryParams.Get("access_type"))
	}
}

func TestHTTPError(t *testing.T) {
	t.Parallel()

	srv := test.NewMockServer(test.MockBehavior{
		StatusCode: http.StatusBadRequest,
		Payload:    []byte(`{"error":{"message":"Invalid parameter","code":100}}`),
	})
	defer srv.Close()

	client := auth.NewClient(mockConfig(srv.Server.URL))
	_, err := client.CreateSystemUser(context.Background(), &auth.CreateSystemUserRequest{
		Name: "Test", Role: "INVALID",
	})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

type (
	mockTokenStore     struct{ token string }
	mockTokenRefresher struct{ newToken string }
	mockTokenRevoker   struct{ revoked string }
)

func (s *mockTokenStore) Get(_ context.Context) (string, error) { return s.token, nil }
func (s *mockTokenStore) Add(_ context.Context, t string) error { s.token = t; return nil }
func (r *mockTokenRefresher) Refresh(_ context.Context, _ string) (string, error) {
	return r.newToken, nil
}
func (r *mockTokenRevoker) Revoke(_ context.Context, t string) error { r.revoked = t; return nil }

func TestRotateAccessToken(t *testing.T) {
	t.Parallel()

	store := &mockTokenStore{token: "old-token"}
	refresher := &mockTokenRefresher{newToken: "new-token"}
	revoker := &mockTokenRevoker{}

	err := auth.RotateAccessToken(context.Background(), refresher, revoker, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.token != "new-token" {
		t.Errorf("expected stored token=new-token, got %s", store.token)
	}
	if revoker.revoked != "old-token" {
		t.Errorf("expected revoked token=old-token, got %s", revoker.revoked)
	}
}
