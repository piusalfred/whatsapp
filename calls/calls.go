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

// Package calls provides a client for the WhatsApp Business Calls API.
//
// The Calls API enables businesses to initiate and manage voice calls through WhatsApp.
// It supports checking user permissions, connecting calls via WebRTC signaling, and
// managing call lifecycle states (pre-accept, accept, reject, terminate, media update).
//
// # Getting Started
//
// Create a [Client] using [NewClient] with a [config.Config] and optional sender options:
//
//	conf := &config.Config{
//	    BaseURL:       "https://graph.facebook.com",
//	    APIVersion:    "v22.0",
//	    AccessToken:   "YOUR_ACCESS_TOKEN",
//	    PhoneNumberID: "YOUR_PHONE_NUMBER_ID",
//	}
//
//	client := calls.NewClient(conf,
//	    whttp.WithSenderTimeout(30*time.Second),
//	    whttp.WithSenderMaxBodyBytes(10<<20),
//	)
//
// # Checking Permissions
//
// Before initiating a call, check whether the business is permitted to call a specific user:
//
//	resp, err := client.CheckPermission(ctx, calls.NewCheckPermissionRequest("USER_WA_ID"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Permission.Status) // "granted", "pending", "denied", or "expired"
//
// # Connecting a Call
//
// To start a voice call, create a connect request with the recipient's WA ID and a WebRTC offer SDP:
//
//	req := calls.ConnectRequest("USER_WA_ID", sdpOffer)
//	resp, err := client.UpdateCallStatus(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Call ID:", resp.Calls[0].ID)
//
// # Managing Call Lifecycle
//
// After a call is connected, transition its state using the appropriate action:
//
//	// Accept an incoming call with an SDP answer
//	req := calls.AcceptRequest(callID, sdpAnswer)
//	resp, err := client.UpdateCallStatus(ctx, req)
//
//	// Reject or terminate a call
//	resp, err = client.UpdateCallStatus(ctx, calls.RejectRequest(callID))
//	resp, err = client.UpdateCallStatus(ctx, calls.TerminateRequest(callID))
//
// # Configuration Options
//
// [whttp.CoreSenderOption] functions customize the underlying HTTP behavior:
//
//	whttp.WithSenderHTTPClient(customHTTPClient)
//	whttp.WithSenderRequestInterceptor(myRequestHook)
//	whttp.WithSenderResponseInterceptor(myResponseHook)
//	whttp.WithSenderTimeout(30 * time.Second)
//	whttp.WithSenderMaxBodyBytes(10 << 20)
//	whttp.WithSenderMaxHeaderBytes(1 << 20)
//
// # Testing
//
// For unit tests, inject a mock sender via [Client.SetBaseClient]:
//
//	client := calls.NewClient(conf)
//	client.SetBaseClient(mockSender)
package calls

import (
	"context"
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

// CallAction represents the lifecycle action applied to a call.
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
	// Client orchestrates high-level Calls API operations. It holds a reference to
	// a [BaseClient] for HTTP transport and a [config.Config] for endpoint and
	// credential resolution.
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	// CheckPermissionRequest queries whether a business is allowed to call a
	// specific WhatsApp user and retrieves current rate-limit quotas.
	CheckPermissionRequest struct {
		UserWaID              string `json:"user_wa_id"`
		BizOpaqueCallbackData string `json:"biz_opaque_callback_data,omitempty"` // Forwarded to subsequent webhooks for correlation
	}

	// CallUpdateStatusRequest carries the payload for transitioning a call's
	// lifecycle state. The Action field determines which fields are required.
	CallUpdateStatusRequest struct {
		CallID                string       `json:"call_id"`
		Action                CallAction   `json:"action"`
		Session               *SessionInfo `json:"session,omitempty"`                  // Required for connect (offer) and accept (answer)
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"` // Max 512 characters string for tracking
		To                    string       `json:"to,omitempty"`                       // Recipient's WA ID. Required only for 'connect'
	}

	// CallPermissionCheckResponse is the decoded response from a permission check.
	CallPermissionCheckResponse struct {
		MessagingProduct string          `json:"messaging_product"`
		Permission       *Permission     `json:"permission"`
		Actions          []*ActionDetail `json:"actions,omitempty"`
	}

	// Permission describes the calling permission status for a user.
	Permission struct {
		Status         string `json:"status"` // "granted", "pending", "denied", or "expired"
		ExpirationTime int64  `json:"expiration_time,omitempty"`
	}

	// ActionDetail provides granular information about a specific callable action
	// and its current usage limits.
	ActionDetail struct {
		ActionName       string   `json:"action_name"` // "start_call" or "send_call_permission_request"
		CanPerformAction bool     `json:"can_perform_action"`
		Limits           []*Limit `json:"limits,omitempty"`
	}

	// Limit describes a rate-limit bucket for a specific action.
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

	// BaseRequest is the wire-format payload sent for call status updates.
	BaseRequest struct {
		MessagingProduct      string       `json:"messaging_product,omitempty"`
		To                    string       `json:"to,omitempty"`
		CallID                string       `json:"call_id,omitempty"`
		Action                CallAction   `json:"action,omitempty"`
		Session               *SessionInfo `json:"session,omitempty"`
		BizOpaqueCallbackData string       `json:"biz_opaque_callback_data,omitempty"`
	}

	// CallUpdateStatusResponse is the API response for a successful status update.
	CallUpdateStatusResponse struct {
		MessagingProduct string         `json:"messaging_product"`
		Success          bool           `json:"success"`
		Calls            []*Call        `json:"calls"`
		Error            *werrors.Error `json:"error,omitempty"`
	}

	// Call identifies a specific call instance. The ID is generated by WhatsApp
	// upon a successful connect action.
	Call struct {
		ID string `json:"id"`
	}

	// SessionInfo abstracts WebRTC signaling layers. The underlying SDP content
	// payload must strictly comply with RFC 8866 specifications.
	SessionInfo struct {
		SDPType string `json:"sdp_type"` // "offer" or "answer"
		SDP     string `json:"sdp"`
	}
)

// ToCallPermissionCheckResponse attempts to coerce a BaseResponse into a
// CallPermissionCheckResponse. It returns nil when neither Permission nor
// Actions are present.
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

// ToCallUpdateResponse attempts to coerce a BaseResponse into a
// CallUpdateStatusResponse. It returns nil when the response does not indicate
// success, carries no calls, and contains no error.
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

// UpdateCallStatus acts on a call lifecycle status (accept, reject, terminate,
// connect, pre_accept, or media_update).
//
// If the user lacks business call permissions, the API returns Error Code 138006.
//
// Example:
//
//	req := calls.ConnectRequest("USER_WA_ID", sdpOffer)
//	resp, err := client.UpdateCallStatus(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Call ID:", resp.Calls[0].ID)
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
//
// Example:
//
//	resp, err := client.CheckPermission(ctx, calls.NewCheckPermissionRequest("USER_WA_ID"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Permission status:", resp.Permission.Status)
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

// Send dispatches a raw [Request] through the underlying [BaseClient].
func (c *Client) Send(ctx context.Context, request *Request) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return response, nil
}

// NewClient creates a high-level [Client] for the Calls API. The conf argument
// provides endpoint and credential configuration. Optional [SenderOption]
// functions tune the underlying HTTP transport.
//
// Example:
//
//	client := calls.NewClient(conf,
//	    whttp.WithSenderTimeout(30*time.Second),
//	    whttp.WithSenderMaxBodyBytes(10<<20),
//	)
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[BaseRequest](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender. This is useful during
// testing when you want to inject a mock [whttp.Sender] and bypass the default
// HTTP stack entirely.
func (c *Client) SetBaseClient(sender whttp.Sender[BaseRequest]) {
	c.sender.SetSender(sender)
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[BaseRequest]) {
	c.sender.SetMiddlewares(mws...)
}

// BaseClient is the low-level HTTP executor for the Calls API. It converts
// domain [Request] values into HTTP traffic and decodes JSON responses.
type BaseClient struct {
	whttp.BaseClient[BaseRequest]
}

// Send translates a high-level [Request] into an HTTP transaction and returns
// the decoded [BaseResponse].
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
		return nil, fmt.Errorf("%w: %s", whttp.ErrUnknownRequestType, request.RequestType)
	}

	b := whttp.NewRequestBuilder(method, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(request.RequestType).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, endpoint)

	if len(queryParams) > 0 {
		b = b.QueryParams(queryParams)
	}

	req := whttp.BuildRequest(b, message)

	resp := &BaseResponse{}
	decoder := whttp.ResponseDecoderJSON(resp, whttp.DecodeOptionsPermissive())

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

// CheckPermission checks if a business is permitted to call a user.
func (bc *BaseClient) CheckPermission(
	ctx context.Context,
	conf *config.Config,
	request *CheckPermissionRequest,
) (*CallPermissionCheckResponse, error) {
	req := &Request{
		MessagingProduct:      "whatsapp",
		RequestType:           whttp.RequestTypeCheckCallPermissions,
		UserWaID:              request.UserWaID,
		BizOpaqueCallbackData: request.BizOpaqueCallbackData,
	}
	resp, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("check permission failed: %w", err)
	}
	return resp.ToCallPermissionCheckResponse(), nil
}

// UpdateCallStatus transitions a call's lifecycle state.
func (bc *BaseClient) UpdateCallStatus(
	ctx context.Context,
	conf *config.Config,
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
	resp, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update call status failed: %w", err)
	}
	return resp.ToCallUpdateResponse(), nil
}

// SetBizOpaqueCallbackData attaches an opaque callback data string to the
// permission request for webhook correlation.
func (perm *CheckPermissionRequest) SetBizOpaqueCallbackData(bizOpaqueCallbackData string) {
	perm.BizOpaqueCallbackData = bizOpaqueCallbackData
}

// SetBizOpaqueCallbackData attaches an opaque callback data string to the
// status update request for webhook correlation.
func (s *CallUpdateStatusRequest) SetBizOpaqueCallbackData(bizOpaqueCallbackData string) {
	s.BizOpaqueCallbackData = bizOpaqueCallbackData
}

// NewCheckPermissionRequest creates a permission check request for the given
// WhatsApp user ID.
func NewCheckPermissionRequest(userWaID string) *CheckPermissionRequest {
	return &CheckPermissionRequest{
		UserWaID: userWaID,
	}
}

// ConnectRequest creates a call initiation request. The to argument is the
// recipient's WA ID and sdp is the WebRTC offer SDP.
func ConnectRequest(to, sdp string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		To:      to,
		Action:  ConnectCallAction,
		Session: &SessionInfo{SDPType: "offer", SDP: sdp},
	}
}

// AcceptRequest creates a call acceptance request. The sdp argument is the
// WebRTC answer SDP.
func AcceptRequest(callID, sdp string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID:  callID,
		Action:  AcceptCallAction,
		Session: &SessionInfo{SDPType: "answer", SDP: sdp},
	}
}

// PreAcceptRequest creates a pre-accept request to warm up callee resources
// before the final accept.
func PreAcceptRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: PreAcceptCallAction,
	}
}

// RejectRequest creates a call rejection request.
func RejectRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: RejectCallAction,
	}
}

// TerminateRequest creates a call termination request.
func TerminateRequest(callID string) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID: callID,
		Action: TerminateCallAction,
	}
}

// MediaUpdateRequest creates a request to update media parameters for an
// ongoing call.
func MediaUpdateRequest(callID string, session *SessionInfo) *CallUpdateStatusRequest {
	return &CallUpdateStatusRequest{
		CallID:  callID,
		Action:  MediaUpdateCallAction,
		Session: session,
	}
}
