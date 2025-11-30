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
	"net/http"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	JoinApprovalModeRequired JoinApprovalMode = "approval_required"
	JoinApprovalModeAuto     JoinApprovalMode = "auto_approve"
)

type (
	JoinApprovalMode  string
	CreateGroupParams struct {
		MessagingProduct string           `json:"messaging_product"`
		Subject          string           `json:"subject,omitempty"`
		Description      string           `json:"description,omitempty"`
		JoinApprovalMode JoinApprovalMode `json:"join_approval_mode,omitempty"`
	}

	Request struct {
		MessagingProduct string            `json:"messaging_product"`
		Subject          string            `json:"subject,omitempty"`
		GroupID          string            `json:"group_id,omitempty"`
		Description      string            `json:"description,omitempty"`
		JoinApprovalMode JoinApprovalMode  `json:"join_approval_mode,omitempty"`
		RequestType      whttp.RequestType `json:"-"`
		Participants     []*Participants   `json:"participants,omitempty"`
		GroupInfoFields  []string          `json:"group_info_fields,omitempty"`
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

	Response struct {
		MessagingProduct      string              `json:"messaging_product"`
		Subject               string              `json:"subject,omitempty"`
		CreationTimestamp     string              `json:"creation_timestamp,omitempty"`
		Suspended             bool                `json:"suspended,omitempty"`
		Description           string              `json:"description,omitempty"`
		TotalParticipantCount int                 `json:"total_participant_count,omitempty"`
		Participants          []*Participants     `json:"participants,omitempty"`
		JoinApprovalMode      JoinApprovalMode    `json:"join_approval_mode,omitempty"`
		Data                  []*ResponseDataItem `json:"data,omitempty"`
		Paging                whttp.Paging        `json:"paging,omitempty"`
	}

	BaseClient struct {
		sender whttp.Sender[Request]
	}
)

func createRequest(conf *config.Config, request *Request) *whttp.Request[Request] {
	var (
		method    string
		endpoints []string
		params    = map[string]string{}
	)
	switch request.RequestType {
	case whttp.RequestTypeRemoveGroupParticipants:
		method = http.MethodDelete
		endpoints = []string{
			conf.APIVersion,
			request.GroupID,
			"participants",
		}

	case whttp.RequestTypeGetGroupInfo:
		method = http.MethodGet
		endpoints = []string{
			conf.APIVersion,
			request.GroupID,
		}
		params["fields"] = strings.Join(request.GroupInfoFields, ",")
	default:
	}
	return whttp.MakeRequest[Request](method, conf.BaseURL,
		whttp.WithRequestType[Request](request.RequestType),
		whttp.WithRequestEndpoints[Request](endpoints...),
	)
}

func (client *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*Response, error) {
	return nil, nil
}
