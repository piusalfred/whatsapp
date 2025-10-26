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

package calls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	callsEndpoint                  = "/calls"
	PreAcceptCallAction CallAction = "pre_accept"
	AcceptCallAction    CallAction = "accept"
	RejectCallAction    CallAction = "reject"
	TerminateCallAction CallAction = "terminate"
)

type (
	Request struct {
		MessagingProduct string       `json:"messaging_product,omitempty"`
		CallID           string       `json:"call_id,omitempty"`
		Action           CallAction   `json:"action,omitempty"`
		Session          *SessionInfo `json:"session,omitempty"`
		// An arbitrary string you can pass in that is useful for tracking and logging purposes
		BizOpaqueCallbackData string `json:"biz_opaque_callback_data,omitempty"`
	}

	Response struct {
		MessagingProduct string `json:"messaging_product,omitempty"`
		Success          bool   `json:"success,omitempty"`
	}

	CallAction string

	SessionInfo struct {
		SDPType string
		SDP     string
	}

	BaseClient struct {
		Sender whttp.Sender[Request]
	}
)

func (sender *BaseClient) Send(ctx context.Context, conf *config.Config, request *Request) (*Response, error) {
	// TODO: determine if this uses access token or bearer token
	req := &whttp.Request[Request]{
		Type:      whttp.RequestTypeUpdateCallStatus,
		BaseURL:   conf.BaseURL,
		Method:    http.MethodPost,
		Endpoints: []string{conf.APIVersion, conf.PhoneNumberID, callsEndpoint},
		Message:   request,
		QueryParams: map[string]string{
			"access_token": conf.AccessToken,
		},
		Bearer: conf.AccessToken,
	}

	res := &Response{}
	decoder := whttp.ResponseDecoderJSON(res, whttp.DecodeOptions{
		InspectResponseError: true,
	})

	err := sender.Sender.Send(ctx, req, decoder)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res, nil
}
