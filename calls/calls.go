// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the “Software”), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package calls

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	callStatusUpdateEndpoint     = "/calls"
	callPermissionsCheckEndpoint = "/call_permissions"
)

type CallAction string

const (
	PreAcceptCallAction   CallAction = "pre_accept"
	AcceptCallAction      CallAction = "accept"
	RejectCallAction      CallAction = "reject"
	TerminateCallAction   CallAction = "terminate"
	ConnectCallAction     CallAction = "connect"
	MediaUpdateCallAction CallAction = "media_update"
)

type (
	// Client orchestrates high-level calls API operations by resolving dynamic
	// configuration credentials per request via the ConfigReader.
	Client struct {
		Sender       *BaseClient
		ConfigReader config.Reader
	}

	CheckPermissionRequest struct {
		UserWaID              string `json:"user_wa_id"`
		BizOpaqueCallbackData string `json:"biz_opaque_callback_data,omitempty"` // Forwarded to subsequent webhooks for correlation
	}

	CallUpdateStatusRequest struct {
		CallID                string       `json:"call_id"`
		Action                CallAction   `json:"action"`
		Session               *SessionInfo `json:"session,omitempty"`                  // Required for connect (offer) and accept (answer)
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"` // Max 512 characters string for tracking
		To                    string       `json:"to,omitempty"`                       // Recipient's WA ID. Required only for 'connect'
	}

	CallPermissionCheckResponse struct {
		MessagingProduct string          `json:"messaging_product"`
		Permission       *Permission     `json:"permission"`
		Actions          []*ActionDetail `json:"actions,omitempty"`
	}

	Permission struct {
		Status         string `json:"status"` // "granted", "pending", "denied", or "expired"
		ExpirationTime int64  `json:"expiration_time,omitempty"`
	}

	ActionDetail struct {
		ActionName       string   `json:"action_name"` // "start_call" or "send_call_permission_request"
		CanPerformAction bool     `json:"can_perform_action"`
		Limits           []*Limit `json:"limits,omitempty"`
	}

	Limit struct {
		TimePeriod          string `json:"time_period"` // e.g., "24h"
		CurrentUsage        int    `json:"current_usage"`
		MaxAllowed          int    `json:"max_allowed"`
		LimitExpirationTime int64  `json:"limit_expiration_time,omitempty"`
	}

	// Request is an internal unified context data carrier mapping both
	// permission queries and state mutations down to the HTTP executor.
	Request struct {
		MessagingProduct      string            `json:"messaging_product,omitempty"`
		CallID                string            `json:"call_id,omitempty"`
		Action                CallAction        `json:"action,omitempty"`
		Session               *SessionInfo      `json:"session,omitempty"`
		BizOpaqueCallbackData string            `json:"biz_opaque_callback_data,omitempty"`
		RequestType           whttp.RequestType `json:"-"`
		UserWaID              string            `json:"-"`
		To                    string            `json:"to,omitempty"`
	}

	// BaseResponse acts as a flexible intermediate data capture layer unmarshaling
	// varying response structures across disparate HTTP verbs.
	BaseResponse struct {
		MessagingProduct string          `json:"messaging_product,omitempty"`
		Success          bool            `json:"success,omitempty"`
		Calls            []*Call         `json:"calls,omitempty"`
		Permission       *Permission     `json:"permission,omitempty"`
		Actions          []*ActionDetail `json:"actions,omitempty"`
		Error            *werrors.Error  `json:"error,omitempty"`
	}

	BaseRequest struct {
		MessagingProduct      string       `json:"messaging_product,omitempty"`
		To                    string       `json:"to,omitempty"`
		CallID                string       `json:"call_id,omitempty"`
		Action                CallAction   `json:"action,omitempty"`
		Session               *SessionInfo `json:"session,omitempty"`
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"`
	}

	CallUpdateStatusResponse struct {
		MessagingProduct string         `json:"messaging_product"`
		Success          bool           `json:"success"`
		Calls            []*Call        `json:"calls"`
		Error            *werrors.Error `json:"error,omitempty"`
	}

	Call struct {
		ID string `json:"id"` // Unique tracking reference string generated on successful 'connect'
	}

	// SessionInfo abstracts WebRTC signaling layers. The underlying SDP
	// content payload string must strict-comply with RFC 8866 specifications.
	SessionInfo struct {
		SDPType string `json:"sdp_type"` // "offer" or "answer"
		SDP     string `json:"sdp"`
	}

	BaseClient struct {
		Sender whttp.Sender[BaseRequest]
	}
)

var ErrUnknownRequestType = errors.New("unknown request type")

func (r *BaseResponse) ToCallPermissionCheckResponse() *CallPermissionCheckResponse {
	if r.Permission == nil && len(r.Actions) == 0 {
		return nil
	}
	return &CallPermissionCheckResponse{
		MessagingProduct: r.MessagingProduct,
		Permission:       r.Permission,
		Actions:          r.Actions,
	}
}

func (r *BaseResponse) ToCallUpdateResponse() *CallUpdateStatusResponse {
	if !r.Success && len(r.Calls) == 0 && r.Error == nil {
		return nil
	}
	return &CallUpdateStatusResponse{
		MessagingProduct: r.MessagingProduct,
		Success:          r.Success,
		Calls:            r.Calls,
		Error:            r.Error,
	}
}

// UpdateCallStatus acts on a call life cycle status (accept, reject, terminate, or connect).
// If a user lacks business call permissions, the API returns Error Code 138006.
func (c *Client) UpdateCallStatus(
	ctx context.Context,
	request *CallUpdateStatusRequest,
) (*CallUpdateStatusResponse, error) {
	req := &Request{
		MessagingProduct:      "whatsapp",
		CallID:                request.CallID,
		Action:                request.Action,
		Session:               request.Session,
		BizOpaqueCallbackData: request.BizOpaqueCallbackData,
		RequestType:           whttp.RequestTypeUpdateCallStatus,
		To:                    request.To,
	}

	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update call status failed: %w", err)
	}
	return resp.ToCallUpdateResponse(), nil
}

// CheckPermission checks if a business is permitted to dial a specific user
// along with evaluating current velocity limit quotas.
func (c *Client) CheckPermission(
	ctx context.Context,
	request *CheckPermissionRequest,
) (*CallPermissionCheckResponse, error) {
	req := &Request{
		MessagingProduct:      "whatsapp",
		RequestType:           whttp.RequestTypeCheckCallPermissions,
		UserWaID:              request.UserWaID,
		BizOpaqueCallbackData: request.BizOpaqueCallbackData,
	}

	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("check permission failed: %w", err)
	}
	return resp.ToCallPermissionCheckResponse(), nil
}

func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	conf, err := c.ConfigReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	response, err := c.Sender.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}

func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*BaseResponse, error) {
	var (
		method      string
		endpoint    string
		queryParams map[string]string
		message     *BaseRequest
	)

	switch request.RequestType {
	case whttp.RequestTypeCheckCallPermissions:
		endpoint = callPermissionsCheckEndpoint
		method = http.MethodGet
		queryParams = map[string]string{"user_wa_id": request.UserWaID}

	case whttp.RequestTypeUpdateCallStatus:
		message = &BaseRequest{
			MessagingProduct:      request.MessagingProduct,
			To:                    request.To,
			CallID:                request.CallID,
			Action:                request.Action,
			Session:               request.Session,
			BizOpaqueCallbackData: request.BizOpaqueCallbackData,
		}
		endpoint = callStatusUpdateEndpoint
		method = http.MethodPost

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownRequestType, request.RequestType)
	}

	b := whttp.NewRequestBuilder(method, conf.BaseURL).
		WithBearer(conf.AccessToken).
		WithAppSecret(conf.AppSecret, conf.SecureRequests).
		WithDebugLogLevel(whttp.ParseDebugLogLevel(conf.DebugLogLevel)).
		WithRequestType(request.RequestType).
		WithEndpoints(conf.APIVersion, conf.PhoneNumberID, endpoint)

	if len(queryParams) > 0 {
		b = b.WithQueryParams(queryParams)
	}

	req := whttp.BuildRequest(b, message)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

func (perm *CheckPermissionRequest) SetBizOpaqueCallbackData(bizOpaqueCallbackData string) {
	perm.BizOpaqueCallbackData = bizOpaqueCallbackData
}

func (s *CallUpdateStatusRequest) SetBizOpaqueCallbackData(bizOpaqueCallbackData string) {
	s.BizOpaqueCallbackData = bizOpaqueCallbackData
}

func NewCheckPermissionRequest(userWaID string) *CheckPermissionRequest {
	return &CheckPermissionRequest{
		UserWaID: userWaID,
	}
}

func ConnectRequest(to, sdp string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		To:      to,
		Action:  ConnectCallAction,
		Session: &SessionInfo{SDPType: "offer", SDP: sdp},
	}
}

func AcceptRequest(callID, sdp string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID:  callID,
		Action:  AcceptCallAction,
		Session: &SessionInfo{SDPType: "answer", SDP: sdp},
	}
}

func PreAcceptRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: PreAcceptCallAction,
	}
}

func RejectRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: RejectCallAction,
	}
}

func TerminateRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: TerminateCallAction,
	}
}

func MediaUpdateRequest(callID string, session *SessionInfo) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID:  callID,
		Action:  MediaUpdateCallAction,
		Session: session,
	}
}
