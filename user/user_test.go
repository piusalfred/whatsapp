// //  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
// //
// //  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// //  and associated documentation files (the "Software"), to deal in the Software without restriction,
// //  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// //  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// //  subject to the following conditions:
// //
// //  The above copyright notice and this permission notice shall be included in all copies or substantial
// //  portions of the Software.
// //
// //  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// //  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// //  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// //  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// //  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package user_test

//
// import (
//	"context"
//	"encoding/json"
//	"net/http"
//	"strings"
//	"testing"
//
//	"github.com/piusalfred/whatsapp/config"
//	"github.com/piusalfred/whatsapp/internal/test"
//	whttp "github.com/piusalfred/whatsapp/pkg/http"
//	"github.com/piusalfred/whatsapp/user"
//)
//
//func mockConfigReader(baseURL string) config.Reader {
//	return config.ReaderFunc(func(ctx context.Context) (*config.Config, error) {
//		return &config.Config{
//			BaseURL:        baseURL,
//			APIVersion:     "v25.0",
//			PhoneNumberID:  "106540352242922",
//			AccessToken:    "EAAJB...",
//			AppSecret:      "",
//			SecureRequests: false,
//		}, nil
//	})
//}
//
//func TestBlockUsers_RequestPayload(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"messaging_product": "whatsapp",
//			"block_users": map[string]any{
//				"added_users": []map[string]string{
//					{"input": "+16505551234", "wa_id": "16505551234"},
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	_, err := client.Block(context.Background(), &user.BlockRequest{Numbers: []string{"+16505551234"}})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if srv.LastRequest.Method != http.MethodPost {
//		t.Errorf("expected POST, got %s", srv.LastRequest.Method)
//	}
//	wantPath := "/v25.0/106540352242922/block_users"
//	if srv.LastRequest.URL.Path != wantPath {
//		t.Errorf("expected path %q, got %q", wantPath, srv.LastRequest.URL.Path)
//	}
//
//	var body map[string]any
//	if err := json.Unmarshal(srv.LastBody, &body); err != nil {
//		t.Fatalf("failed to unmarshal body: %v", err)
//	}
//	if body["messaging_product"] != "whatsapp" {
//		t.Errorf("expected messaging_product=whatsapp, got %v", body["messaging_product"])
//	}
//	users, ok := body["block_users"].([]any)
//	if !ok || len(users) != 1 {
//		t.Fatalf("expected 1 block user, got %v", body["block_users"])
//	}
//	u := users[0].(map[string]any)
//	if u["user"] != "+16505551234" {
//		t.Errorf("expected user=+16505551234, got %v", u["user"])
//	}
//}
//
//func TestUnblockUsers_RequestPayload(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"messaging_product": "whatsapp",
//			"block_users": map[string]any{
//				"removed_users": []map[string]string{
//					{"input": "+16505551234", "wa_id": "16505551234"},
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	_, err := client.Unblock(context.Background(), &user.UnblockRequest{Numbers: []string{"+16505551234"}})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if srv.LastRequest.Method != http.MethodDelete {
//		t.Errorf("expected DELETE, got %s", srv.LastRequest.Method)
//	}
//
//	var body map[string]any
//	if err := json.Unmarshal(srv.LastBody, &body); err != nil {
//		t.Fatalf("failed to unmarshal body: %v", err)
//	}
//	users, ok := body["block_users"].([]any)
//	if !ok || len(users) != 1 {
//		t.Fatalf("expected 1 block user in body, got %v", body["block_users"])
//	}
//}
//
//func TestListBlockedUsers_QueryParams(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"data": []map[string]string{
//				{"messaging_product": "whatsapp", "wa_id": "16505551234"},
//			},
//			"paging": map[string]any{
//				"cursors": map[string]string{
//					"after":  "after_cursor",
//					"before": "before_cursor",
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	limit := 10
//	after := "after123"
//	before := "before123"
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	_, err := client.ListBlocked(context.Background(), &user.ListBlockedUsersOptions{
//		Limit:  &limit,
//		After:  &after,
//		Before: &before,
//	})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if srv.LastRequest.Method != http.MethodGet {
//		t.Errorf("expected GET, got %s", srv.LastRequest.Method)
//	}
//
//	q := srv.LastRequest.URL.Query()
//	if q.Get("limit") != "10" {
//		t.Errorf("expected limit=10, got %s", q.Get("limit"))
//	}
//	if q.Get("after") != "after123" {
//		t.Errorf("expected after=after123, got %s", q.Get("after"))
//	}
//	if q.Get("before") != "before123" {
//		t.Errorf("expected before=before123, got %s", q.Get("before"))
//	}
//}
//
//func TestBlockUsers_SuccessResponse(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"messaging_product": "whatsapp",
//			"block_users": map[string]any{
//				"added_users": []map[string]string{
//					{"input": "+16505551234", "wa_id": "16505551234"},
//					{"input": "+14155559876", "wa_id": "14155559876"},
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	resp, err := client.Block(context.Background(), &user.BlockRequest{Numbers: []string{"+16505551234", "+14155559876"}})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if resp.MessagingProduct != "whatsapp" {
//		t.Errorf("expected MessagingProduct=whatsapp, got %s", resp.MessagingProduct)
//	}
//	if resp.BlockUsers == nil || len(resp.BlockUsers.AddedUsers) != 2 {
//		t.Fatalf("expected 2 added users, got %+v", resp.BlockUsers)
//	}
//	if resp.BlockUsers.AddedUsers[0].Input != "+16505551234" {
//		t.Errorf("unexpected input: %s", resp.BlockUsers.AddedUsers[0].Input)
//	}
//}
//
//func TestBlockUsers_MixedResponse(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		_, _ = w.Write([]byte(`{
//			"messaging_product": "whatsapp",
//			"block_users": {
//				"added_users": [
//					{"input": "+16505551234", "wa_id": "16505551234"}
//				],
//				"failed_users": [
//					{
//						"input": "+14155559876",
//						"wa_id": "14155559876",
//						"errors": [
//							{
//								"message": "Re-engagement required",
//								"code": 131047,
//								"error_data": {"details": "User has not messaged in the last 24 hours"}
//							}
//						]
//					}
//				]
//			},
//			"error": {
//				"message": "(#139100) Failed to block/unblock users",
//				"type": "OAuthException",
//				"code": 139100,
//				"error_data": {"details": "Failed to block some users"}
//			}
//		}`))
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	resp, err := client.Block(context.Background(), &user.BlockRequest{Numbers: []string{"+16505551234", "+14155559876"}})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if resp.BlockUsers == nil {
//		t.Fatal("expected BlockUsers")
//	}
//	if len(resp.BlockUsers.AddedUsers) != 1 {
//		t.Errorf("expected 1 added user, got %d", len(resp.BlockUsers.AddedUsers))
//	}
//	if len(resp.BlockUsers.FailedUsers) != 1 {
//		t.Errorf("expected 1 failed user, got %d", len(resp.BlockUsers.FailedUsers))
//	}
//	if len(resp.BlockUsers.FailedUsers[0].Errors) != 1 {
//		t.Fatalf("expected 1 error in failed user, got %d", len(resp.BlockUsers.FailedUsers[0].Errors))
//	}
//	if resp.BlockUsers.FailedUsers[0].Errors[0].Code != 131047 {
//		t.Errorf("expected error code 131047, got %d", resp.BlockUsers.FailedUsers[0].Errors[0].Code)
//	}
//	if resp.Error == nil {
//		t.Fatal("expected top-level error")
//	}
//	if !strings.Contains(resp.Error.Message, "139100") {
//		t.Errorf("expected top-level error to contain 139100, got %s", resp.Error.Message)
//	}
//}
//
//func TestUnblockUsers_RemovedUsersResponse(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"messaging_product": "whatsapp",
//			"block_users": map[string]any{
//				"removed_users": []map[string]string{
//					{"input": "+16505551234", "wa_id": "16505551234"},
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	resp, err := client.Unblock(context.Background(), &user.UnblockRequest{Numbers: []string{"+16505551234"}})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if resp.BlockUsers == nil || len(resp.BlockUsers.RemovedUsers) != 1 {
//		t.Fatalf("expected 1 removed user, got %+v", resp.BlockUsers)
//	}
//	if resp.BlockUsers.RemovedUsers[0].Input != "+16505551234" {
//		t.Errorf("unexpected input: %s", resp.BlockUsers.RemovedUsers[0].Input)
//	}
//}
//
//func TestListBlockedUsers_SuccessResponse(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(map[string]any{
//			"data": []map[string]string{
//				{"messaging_product": "whatsapp", "wa_id": "16505551234"},
//				{"messaging_product": "whatsapp", "wa_id": "14155559876"},
//			},
//			"paging": map[string]any{
//				"cursors": map[string]string{
//					"after":  "after_cursor",
//					"before": "before_cursor",
//				},
//			},
//		})
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	resp, err := client.ListBlocked(context.Background(), &user.ListBlockedUsersOptions{})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if len(resp.Data) != 2 {
//		t.Fatalf("expected 2 data entries, got %d", len(resp.Data))
//	}
//	if resp.Data[0].WaID != "16505551234" {
//		t.Errorf("unexpected wa_id: %s", resp.Data[0].WaID)
//	}
//	if resp.Paging == nil || resp.Paging.Cursors == nil {
//		t.Fatal("expected paging cursors")
//	}
//	if resp.Paging.Cursors.After != "after_cursor" {
//		t.Errorf("unexpected after cursor: %s", resp.Paging.Cursors.After)
//	}
//}
//
//func TestBlockUsers_HTTPError(t *testing.T) {
//	t.Parallel()
//
//	srv := test.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusTooManyRequests)
//		_, _ = w.Write([]byte(`{"error":{"message":"Rate limit","code":130429}}`))
//	})
//	defer srv.Close()
//
//	client := user.NewBlockClient(mockConfigReader(srv.Server.URL), whttp.NewSender[user.BlockBaseRequest]())
//	_, err := client.Block(context.Background(), &user.BlockRequest{Numbers: []string{"+16505551234"}})
//	if err == nil {
//		t.Fatal("expected error for 429 response")
//	}
//}
