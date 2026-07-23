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

	"github.com/piusalfred/whatsapp/config"
	"github.com/piusalfred/whatsapp/groups"
)

func (c *Client) CreateGroup(ctx context.Context, req *groups.CreateGroupRequest) (*groups.BaseResponse, error) {
	resp, err := c.sender.CreateGroup(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteGroup(ctx context.Context, req *groups.DeleteGroupRequest) (*groups.BaseResponse, error) {
	resp, err := c.sender.DeleteGroup(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("delete group: %w", err)
	}
	return resp, nil
}

func (c *Client) GetGroupInfo(ctx context.Context, req *groups.GetGroupInfoRequest) (*groups.GroupInfoResponse, error) {
	resp, err := c.sender.GetGroupInfo(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get group info: %w", err)
	}
	return resp, nil
}

func (c *Client) GetGroupInviteLink(
	ctx context.Context,
	req *groups.GetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := c.sender.GetGroupInviteLink(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("get group invite link: %w", err)
	}
	return resp, nil
}

func (c *Client) ResetGroupInviteLink(
	ctx context.Context,
	req *groups.ResetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := c.sender.ResetGroupInviteLink(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("reset group invite link: %w", err)
	}
	return resp, nil
}

func (c *Client) RemoveGroupParticipants(
	ctx context.Context,
	req *groups.RemoveGroupParticipantsRequest,
) (*groups.BaseResponse, error) {
	resp, err := c.sender.RemoveGroupParticipants(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("remove group participants: %w", err)
	}
	return resp, nil
}

func (c *Client) ListActiveGroups(
	ctx context.Context,
	req *groups.GetActiveGroupsRequest,
) (*groups.ActiveGroupsResponse, error) {
	resp, err := c.sender.ListActiveGroups(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	return resp, nil
}

func (c *Client) UpdateGroupSettings(
	ctx context.Context,
	req *groups.UpdateGroupSettingsRequest,
) (*groups.BaseResponse, error) {
	resp, err := c.sender.UpdateGroupSettings(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("update group settings: %w", err)
	}
	return resp, nil
}

func (c *Client) ListJoinRequests(
	ctx context.Context,
	req *groups.GetJoinRequestsRequest,
) (*groups.JoinRequestsResponse, error) {
	resp, err := c.sender.ListJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return resp, nil
}

func (c *Client) ApproveJoinRequests(
	ctx context.Context,
	req *groups.ApproveJoinRequestsRequest,
) (*groups.ApproveJoinRequestsResponse, error) {
	resp, err := c.sender.ApproveJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("approve join requests: %w", err)
	}
	return resp, nil
}

func (c *Client) RejectJoinRequests(
	ctx context.Context,
	req *groups.RejectJoinRequestsRequest,
) (*groups.RejectJoinRequestsResponse, error) {
	resp, err := c.sender.RejectJoinRequests(ctx, c.config, req)
	if err != nil {
		return nil, fmt.Errorf("reject join requests: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) CreateGroup(
	ctx context.Context,
	conf *config.Config,
	req *groups.CreateGroupRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.getGroups().CreateGroup(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) DeleteGroup(
	ctx context.Context,
	conf *config.Config,
	req *groups.DeleteGroupRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.getGroups().DeleteGroup(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("delete group: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetGroupInfo(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetGroupInfoRequest,
) (*groups.GroupInfoResponse, error) {
	resp, err := bc.getGroups().GetGroupInfo(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get group info: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) GetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := bc.getGroups().GetGroupInviteLink(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("get group invite link: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ResetGroupInviteLink(
	ctx context.Context,
	conf *config.Config,
	req *groups.ResetGroupInviteLinkRequest,
) (*groups.GroupInviteLinkResponse, error) {
	resp, err := bc.getGroups().ResetGroupInviteLink(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("reset group invite link: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RemoveGroupParticipants(
	ctx context.Context,
	conf *config.Config,
	req *groups.RemoveGroupParticipantsRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.getGroups().RemoveGroupParticipants(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("remove group participants: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListActiveGroups(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetActiveGroupsRequest,
) (*groups.ActiveGroupsResponse, error) {
	resp, err := bc.getGroups().GetActiveGroups(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) UpdateGroupSettings(
	ctx context.Context,
	conf *config.Config,
	req *groups.UpdateGroupSettingsRequest,
) (*groups.BaseResponse, error) {
	resp, err := bc.getGroups().UpdateGroupSettings(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("update group settings: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ListJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.GetJoinRequestsRequest,
) (*groups.JoinRequestsResponse, error) {
	resp, err := bc.getGroups().GetJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) ApproveJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.ApproveJoinRequestsRequest,
) (*groups.ApproveJoinRequestsResponse, error) {
	resp, err := bc.getGroups().ApproveJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("approve join requests: %w", err)
	}
	return resp, nil
}

func (bc *BaseClient) RejectJoinRequests(
	ctx context.Context,
	conf *config.Config,
	req *groups.RejectJoinRequestsRequest,
) (*groups.RejectJoinRequestsResponse, error) {
	resp, err := bc.getGroups().RejectJoinRequests(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("reject join requests: %w", err)
	}
	return resp, nil
}
