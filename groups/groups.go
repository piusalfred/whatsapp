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

package groups

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	werrors "github.com/piusalfred/whatsapp/pkg/errors"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const messagingProduct = "whatsapp"

const (
	JoinApprovalModeRequired JoinApprovalMode = "approval_required"
	JoinApprovalModeAuto     JoinApprovalMode = "auto_approve"
)

type (
	JoinApprovalMode string

	// CreateGroupRequest represents a request to create a new group.
	CreateGroupRequest struct {
		Subject          string
		Description      string
		JoinApprovalMode JoinApprovalMode
	}

	// DeleteGroupRequest represents a request to delete a group.
	DeleteGroupRequest struct {
		GroupID string
	}

	// GetGroupInviteLinkRequest represents a request to get a group's invite link.
	GetGroupInviteLinkRequest struct {
		GroupID string
	}

	// ResetGroupInviteLinkRequest represents a request to reset a group's invite link.
	ResetGroupInviteLinkRequest struct {
		GroupID string
	}

	// RemoveGroupParticipantsRequest represents a request to remove participants from a group.
	RemoveGroupParticipantsRequest struct {
		GroupID      string
		Participants []string
	}

	// GetGroupInfoRequest represents a request to retrieve metadata about a single group.
	GetGroupInfoRequest struct {
		GroupID string
		Fields  []string
	}

	// GetActiveGroupsRequest represents a request to retrieve a list of active groups.
	GetActiveGroupsRequest struct {
		Limit  int
		After  string
		Before string
	}

	// UpdateGroupSettingsRequest represents a request to update group settings.
	UpdateGroupSettingsRequest struct {
		GroupID            string
		Subject            string
		Description        string
		ProfilePictureFile string
	}

	// GetJoinRequestsRequest represents a request to get pending join requests for a group.
	GetJoinRequestsRequest struct {
		GroupID string
	}

	// ApproveJoinRequestsRequest represents a request to approve one or more join requests.
	ApproveJoinRequestsRequest struct {
		GroupID      string
		JoinRequests []string
	}

	// RejectJoinRequestsRequest represents a request to reject one or more join requests.
	RejectJoinRequestsRequest struct {
		GroupID      string
		JoinRequests []string
	}

	// Request is the combined struct that holds all fields needed for any group management operation.
	Request struct {
		MessagingProduct   string            `json:"messaging_product"`
		Subject            string            `json:"subject,omitempty"`
		GroupID            string            `json:"-"`
		Description        string            `json:"description,omitempty"`
		JoinApprovalMode   JoinApprovalMode  `json:"join_approval_mode,omitempty"`
		RequestType        whttp.RequestType `json:"-"`
		Participants       []*Participants   `json:"participants,omitempty"`
		JoinRequests       []string          `json:"join_requests,omitempty"`
		ProfilePictureFile string            `json:"-"`
		GroupInfoFields    []string          `json:"-"`
		Limit              int               `json:"-"`
		After              string            `json:"-"`
		Before             string            `json:"-"`
	}

	Participants struct {
		User string `json:"user"`
		WaID string `json:"wa_id"`
	}

	ResponseDataItem struct {
		JoinRequestID     string `json:"join_request_id"`
		WhatsappID        string `json:"wa_id"`
		CreationTimestamp string `json:"creation_timestamp"`
	}

	ActiveGroup struct {
		ID        string `json:"id"`
		Subject   string `json:"subject"`
		CreatedAt string `json:"created_at"`
	}

	FailedJoinRequest struct {
		JoinRequestID string          `json:"join_request_id"`
		Errors        []werrors.Error `json:"errors,omitempty"`
	}

	// Response is the general response struct that captures all possible fields
	// returned by the Groups API. Use the typed helper methods to extract
	// operation-specific responses.
	Response struct {
		MessagingProduct      string               `json:"messaging_product"`
		Subject               string               `json:"subject,omitempty"`
		CreationTimestamp     string               `json:"creation_timestamp,omitempty"`
		Suspended             bool                 `json:"suspended,omitempty"`
		Description           string               `json:"description,omitempty"`
		TotalParticipantCount int                  `json:"total_participant_count,omitempty"`
		Participants          []*Participants      `json:"participants,omitempty"`
		JoinApprovalMode      JoinApprovalMode     `json:"join_approval_mode,omitempty"`
		ID                    string               `json:"id,omitempty"`
		InviteLink            string               `json:"invite_link,omitempty"`
		Data                  json.RawMessage      `json:"data,omitempty"`
		ApprovedJoinRequests  []string             `json:"approved_join_requests,omitempty"`
		RejectedJoinRequests  []string             `json:"rejected_join_requests,omitempty"`
		FailedJoinRequests    []*FailedJoinRequest `json:"failed_join_requests,omitempty"`
		Errors                []werrors.Error      `json:"errors,omitempty"`
		Paging                whttp.Paging         `json:"paging"`
	}

	// GroupInfoResponse is returned by GetGroupInfo.
	GroupInfoResponse struct {
		ID                    string           `json:"id,omitempty"`
		Subject               string           `json:"subject,omitempty"`
		CreationTimestamp     string           `json:"creation_timestamp,omitempty"`
		Suspended             bool             `json:"suspended,omitempty"`
		Description           string           `json:"description,omitempty"`
		TotalParticipantCount int              `json:"total_participant_count,omitempty"`
		Participants          []*Participants  `json:"participants,omitempty"`
		JoinApprovalMode      JoinApprovalMode `json:"join_approval_mode,omitempty"`
	}

	// GroupInviteLinkResponse is returned by GetGroupInviteLink and ResetGroupInviteLink.
	GroupInviteLinkResponse struct {
		MessagingProduct string `json:"messaging_product"`
		InviteLink       string `json:"invite_link,omitempty"`
	}

	// ActiveGroupsResponse is returned by GetActiveGroups.
	ActiveGroupsResponse struct {
		Groups []*ActiveGroup `json:"groups,omitempty"`
		Paging whttp.Paging   `json:"paging"`
	}

	// JoinRequestsResponse is returned by GetJoinRequests.
	JoinRequestsResponse struct {
		JoinRequests []*ResponseDataItem `json:"data,omitempty"`
		Paging       whttp.Paging        `json:"paging"`
	}

	// ApproveJoinRequestsResponse is returned by ApproveJoinRequests.
	ApproveJoinRequestsResponse struct {
		MessagingProduct     string               `json:"messaging_product"`
		ApprovedJoinRequests []string             `json:"approved_join_requests,omitempty"`
		FailedJoinRequests   []*FailedJoinRequest `json:"failed_join_requests,omitempty"`
		Errors               []werrors.Error      `json:"errors,omitempty"`
	}

	// RejectJoinRequestsResponse is returned by RejectJoinRequests.
	RejectJoinRequestsResponse struct {
		MessagingProduct     string               `json:"messaging_product"`
		RejectedJoinRequests []string             `json:"rejected_join_requests,omitempty"`
		FailedJoinRequests   []*FailedJoinRequest `json:"failed_join_requests,omitempty"`
		Errors               []werrors.Error      `json:"errors,omitempty"`
	}
)

// GroupInfoResponse extracts a GroupInfoResponse from the general Response.
func (r *Response) GroupInfoResponse() *GroupInfoResponse {
	return &GroupInfoResponse{
		ID:                    r.ID,
		Subject:               r.Subject,
		CreationTimestamp:     r.CreationTimestamp,
		Suspended:             r.Suspended,
		Description:           r.Description,
		TotalParticipantCount: r.TotalParticipantCount,
		Participants:          r.Participants,
		JoinApprovalMode:      r.JoinApprovalMode,
	}
}

// GroupInviteLinkResponse extracts a GroupInviteLinkResponse from the general Response.
func (r *Response) GroupInviteLinkResponse() *GroupInviteLinkResponse {
	return &GroupInviteLinkResponse{
		MessagingProduct: r.MessagingProduct,
		InviteLink:       r.InviteLink,
	}
}

// ActiveGroupsResponse extracts an ActiveGroupsResponse from the general Response.
func (r *Response) ActiveGroupsResponse() *ActiveGroupsResponse {
	return &ActiveGroupsResponse{
		Groups: r.ActiveGroups(),
		Paging: r.Paging,
	}
}

// JoinRequestsResponse extracts a JoinRequestsResponse from the general Response.
func (r *Response) JoinRequestsResponse() *JoinRequestsResponse {
	return &JoinRequestsResponse{
		JoinRequests: r.JoinRequests(),
		Paging:       r.Paging,
	}
}

// ApproveJoinRequestsResponse extracts an ApproveJoinRequestsResponse from the general Response.
func (r *Response) ApproveJoinRequestsResponse() *ApproveJoinRequestsResponse {
	return &ApproveJoinRequestsResponse{
		MessagingProduct:     r.MessagingProduct,
		ApprovedJoinRequests: r.ApprovedJoinRequests,
		FailedJoinRequests:   r.FailedJoinRequests,
		Errors:               r.Errors,
	}
}

// RejectJoinRequestsResponse extracts a RejectJoinRequestsResponse from the general Response.
func (r *Response) RejectJoinRequestsResponse() *RejectJoinRequestsResponse {
	return &RejectJoinRequestsResponse{
		MessagingProduct:     r.MessagingProduct,
		RejectedJoinRequests: r.RejectedJoinRequests,
		FailedJoinRequests:   r.FailedJoinRequests,
		Errors:               r.Errors,
	}
}

// JoinRequests returns the join requests from the Data field.
func (r *Response) JoinRequests() []*ResponseDataItem {
	if len(r.Data) == 0 {
		return nil
	}
	var items []*ResponseDataItem
	_ = json.Unmarshal(r.Data, &items)
	return items
}

// ActiveGroups returns the active groups from the Data field.
func (r *Response) ActiveGroups() []*ActiveGroup {
	if len(r.Data) == 0 {
		return nil
	}
	var wrapper struct {
		Groups []*ActiveGroup `json:"groups"`
	}
	if err := json.Unmarshal(r.Data, &wrapper); err != nil {
		return nil
	}
	return wrapper.Groups
}

type Client struct {
	sender *BaseClient
	config *config.Config
}

func NewClient(conf *config.Config, sender whttp.Sender[Request]) *Client {
	return &Client{
		config: conf,
		sender: NewBaseSender(sender),
	}
}

func (client *Client) CreateGroup(
	ctx context.Context,
	req *CreateGroupRequest,
) (*Response, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeCreateGroup,
		MessagingProduct: messagingProduct,
		Subject:          req.Subject,
		Description:      req.Description,
		JoinApprovalMode: req.JoinApprovalMode,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: create group: %w", err)
	}

	return response, nil
}

func (client *Client) DeleteGroup(
	ctx context.Context,
	req *DeleteGroupRequest,
) (*Response, error) {
	request := &Request{
		RequestType: whttp.RequestTypeDeleteGroup,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: delete group: %w", err)
	}

	return response, nil
}

func (client *Client) GetGroupInviteLink(
	ctx context.Context,
	req *GetGroupInviteLinkRequest,
) (*GroupInviteLinkResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetGroupInviteLink,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: get group invite link: %w", err)
	}

	return response.GroupInviteLinkResponse(), nil
}

func (client *Client) ResetGroupInviteLink(
	ctx context.Context,
	req *ResetGroupInviteLinkRequest,
) (*GroupInviteLinkResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeResetGroupInviteLink,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: reset group invite link: %w", err)
	}

	return response.GroupInviteLinkResponse(), nil
}

func (client *Client) RemoveGroupParticipants(
	ctx context.Context,
	req *RemoveGroupParticipantsRequest,
) (*Response, error) {
	participants := make([]*Participants, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = &Participants{User: p}
	}

	request := &Request{
		RequestType:      whttp.RequestTypeRemoveGroupParticipants,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		Participants:     participants,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: remove group participants: %w", err)
	}

	return response, nil
}

func (client *Client) GetGroupInfo(
	ctx context.Context,
	req *GetGroupInfoRequest,
) (*GroupInfoResponse, error) {
	request := &Request{
		RequestType:     whttp.RequestTypeGetGroupInfo,
		GroupID:         req.GroupID,
		GroupInfoFields: req.Fields,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: get group info: %w", err)
	}

	return response.GroupInfoResponse(), nil
}

func (client *Client) GetActiveGroups(
	ctx context.Context,
	req *GetActiveGroupsRequest,
) (*ActiveGroupsResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetActiveGroups,
		Limit:       req.Limit,
		After:       req.After,
		Before:      req.Before,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: get active groups: %w", err)
	}

	return response.ActiveGroupsResponse(), nil
}

func (client *Client) UpdateGroupSettings(
	ctx context.Context,
	req *UpdateGroupSettingsRequest,
) (*Response, error) {
	request := &Request{
		RequestType:        whttp.RequestTypeUpdateGroupSettings,
		GroupID:            req.GroupID,
		MessagingProduct:   messagingProduct,
		Subject:            req.Subject,
		Description:        req.Description,
		ProfilePictureFile: req.ProfilePictureFile,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: update group settings: %w", err)
	}

	return response, nil
}

func (client *Client) GetJoinRequests(
	ctx context.Context,
	req *GetJoinRequestsRequest,
) (*JoinRequestsResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetJoinRequests,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: get join requests: %w", err)
	}

	return response.JoinRequestsResponse(), nil
}

func (client *Client) ApproveJoinRequests(
	ctx context.Context,
	req *ApproveJoinRequestsRequest,
) (*ApproveJoinRequestsResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeApproveJoinRequests,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		JoinRequests:     req.JoinRequests,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: approve join requests: %w", err)
	}

	return response.ApproveJoinRequestsResponse(), nil
}

func (client *Client) RejectJoinRequests(
	ctx context.Context,
	req *RejectJoinRequestsRequest,
) (*RejectJoinRequestsResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeRejectJoinRequests,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		JoinRequests:     req.JoinRequests,
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("client: reject join requests: %w", err)
	}

	return response.RejectJoinRequestsResponse(), nil
}

func (client *Client) Send(ctx context.Context, request *Request) (*Response, error) {
	response, err := client.sender.Send(ctx, client.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

type BaseClient struct {
	sender whttp.Sender[Request]
}

func NewBaseSender(sender whttp.Sender[Request]) *BaseClient {
	return &BaseClient{sender: sender}
}

func (client *BaseClient) Send(
	ctx context.Context,
	conf *config.Config,
	request *Request,
) (*Response, error) {
	req := buildHTTPRequest(conf, request)

	response := &Response{}
	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := client.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return response, nil
}

func buildHTTPRequest(conf *config.Config, request *Request) *whttp.Request[Request] {
	var (
		method      string
		endpoints   []string
		queryParams = map[string]string{}
	)

	switch request.RequestType {
	case whttp.RequestTypeCreateGroup:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, conf.PhoneNumberID, "groups"}

	case whttp.RequestTypeDeleteGroup:
		method = http.MethodDelete
		endpoints = []string{conf.APIVersion, request.GroupID}

	case whttp.RequestTypeGetGroupInviteLink:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, request.GroupID, "invite_link"}

	case whttp.RequestTypeResetGroupInviteLink:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, request.GroupID, "invite_link"}

	case whttp.RequestTypeRemoveGroupParticipants:
		method = http.MethodDelete
		endpoints = []string{conf.APIVersion, request.GroupID, "participants"}

	case whttp.RequestTypeGetGroupInfo:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, request.GroupID}
		if len(request.GroupInfoFields) > 0 {
			queryParams["fields"] = strings.Join(request.GroupInfoFields, ",")
		}

	case whttp.RequestTypeGetActiveGroups:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, conf.PhoneNumberID, "groups"}
		if request.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(request.Limit)
		}
		if request.After != "" {
			queryParams["after"] = request.After
		}
		if request.Before != "" {
			queryParams["before"] = request.Before
		}

	case whttp.RequestTypeUpdateGroupSettings:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, request.GroupID}

	case whttp.RequestTypeGetJoinRequests:
		method = http.MethodGet
		endpoints = []string{conf.APIVersion, request.GroupID, "join_requests"}

	case whttp.RequestTypeApproveJoinRequests:
		method = http.MethodPost
		endpoints = []string{conf.APIVersion, request.GroupID, "join_requests"}

	case whttp.RequestTypeRejectJoinRequests:
		method = http.MethodDelete
		endpoints = []string{conf.APIVersion, request.GroupID, "join_requests"}
	default:
		panic("unhandled default case")
	}

	return whttp.MakeRequest[Request](method, conf.BaseURL,
		whttp.WithRequestType[Request](request.RequestType),
		whttp.WithRequestEndpoints[Request](endpoints...),
		whttp.WithRequestQueryParams[Request](queryParams),
		whttp.WithRequestMessage[Request](request),
		whttp.WithRequestBearer[Request](conf.AccessToken),
		whttp.WithRequestAppSecret[Request](conf.AppSecret),
		whttp.WithRequestSecured[Request](conf.SecureRequests),
		whttp.WithRequestDebugLogLevel[Request](whttp.ParseDebugLogLevel(conf.DebugLogLevel)),
	)
}

func (client *BaseClient) CreateGroup(
	ctx context.Context,
	conf *config.Config,
	req *CreateGroupRequest,
) (*Response, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeCreateGroup,
		MessagingProduct: messagingProduct,
		Subject:          req.Subject,
		Description:      req.Description,
		JoinApprovalMode: req.JoinApprovalMode,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: create group: %w", err)
	}

	return response, nil
}

func (client *BaseClient) DeleteGroup(
	ctx context.Context,
	conf *config.Config,
	req *DeleteGroupRequest,
) (*Response, error) {
	request := &Request{
		RequestType: whttp.RequestTypeDeleteGroup,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: delete group: %w", err)
	}

	return response, nil
}

func (client *BaseClient) GetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *GetGroupInviteLinkRequest,
) (*GroupInviteLinkResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetGroupInviteLink,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: get group invite link: %w", err)
	}

	return response.GroupInviteLinkResponse(), nil
}

func (client *BaseClient) ResetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *ResetGroupInviteLinkRequest,
) (*GroupInviteLinkResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeResetGroupInviteLink,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: reset group invite link: %w", err)
	}

	return response.GroupInviteLinkResponse(), nil
}

func (client *BaseClient) RemoveGroupParticipants(
	ctx context.Context,
	conf *config.Config,
	req *RemoveGroupParticipantsRequest,
) (*Response, error) {
	participants := make([]*Participants, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = &Participants{User: p}
	}

	request := &Request{
		RequestType:      whttp.RequestTypeRemoveGroupParticipants,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		Participants:     participants,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: remove group participants: %w", err)
	}

	return response, nil
}

func (client *BaseClient) GetGroupInfo(
	ctx context.Context,
	conf *config.Config,
	req *GetGroupInfoRequest,
) (*GroupInfoResponse, error) {
	request := &Request{
		RequestType:     whttp.RequestTypeGetGroupInfo,
		GroupID:         req.GroupID,
		GroupInfoFields: req.Fields,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: get group info: %w", err)
	}

	return response.GroupInfoResponse(), nil
}

func (client *BaseClient) GetActiveGroups(
	ctx context.Context,
	conf *config.Config,
	req *GetActiveGroupsRequest,
) (*ActiveGroupsResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetActiveGroups,
		Limit:       req.Limit,
		After:       req.After,
		Before:      req.Before,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: get active groups: %w", err)
	}

	return response.ActiveGroupsResponse(), nil
}

func (client *BaseClient) UpdateGroupSettings(
	ctx context.Context,
	conf *config.Config,
	req *UpdateGroupSettingsRequest,
) (*Response, error) {
	request := &Request{
		RequestType:        whttp.RequestTypeUpdateGroupSettings,
		GroupID:            req.GroupID,
		MessagingProduct:   messagingProduct,
		Subject:            req.Subject,
		Description:        req.Description,
		ProfilePictureFile: req.ProfilePictureFile,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: update group settings: %w", err)
	}

	return response, nil
}

func (client *BaseClient) GetJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *GetJoinRequestsRequest,
) (*JoinRequestsResponse, error) {
	request := &Request{
		RequestType: whttp.RequestTypeGetJoinRequests,
		GroupID:     req.GroupID,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: get join requests: %w", err)
	}

	return response.JoinRequestsResponse(), nil
}

func (client *BaseClient) ApproveJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *ApproveJoinRequestsRequest,
) (*ApproveJoinRequestsResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeApproveJoinRequests,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		JoinRequests:     req.JoinRequests,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: approve join requests: %w", err)
	}

	return response.ApproveJoinRequestsResponse(), nil
}

func (client *BaseClient) RejectJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *RejectJoinRequestsRequest,
) (*RejectJoinRequestsResponse, error) {
	request := &Request{
		RequestType:      whttp.RequestTypeRejectJoinRequests,
		GroupID:          req.GroupID,
		MessagingProduct: messagingProduct,
		JoinRequests:     req.JoinRequests,
	}

	response, err := client.Send(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("base client: reject join requests: %w", err)
	}

	return response.RejectJoinRequestsResponse(), nil
}
