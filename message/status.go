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

package message

import "context"

const (
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusFailed    Status = "failed"
	StatusDeleted   Status = "deleted"
	StatusWarning   Status = "warning"
)

type (
	Status string

	StatusUpdateResponse struct {
		Success bool `json:"success"`
	}

	StatusUpdateRequest struct {
		MessageID           string
		Status              Status
		WithTypingIndicator bool
	}

	StatusUpdater interface {
		UpdateStatus(ctx context.Context,
			request *StatusUpdateRequest) (*StatusUpdateResponse, error)
	}

	UpdateStatusFunc func(ctx context.Context,
		request *StatusUpdateRequest) (*StatusUpdateResponse, error)
)

func (fn UpdateStatusFunc) UpdateStatus(ctx context.Context,
	request *StatusUpdateRequest,
) (*StatusUpdateResponse, error) {
	return fn(ctx, request)
}

var (
	_ StatusUpdater = (*BaseClient)(nil)
	_ StatusUpdater = (*Client)(nil)
)
