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

package settings

import (
	"context"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp"
	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	settingsEndpoint  = "/settings"
	ErrGetSettings    = whatsapp.Error("failed to get settings")
	ErrUpdateSettings = whatsapp.Error("failed to update settings")
)

type (
	GetSettingsRequest struct {
		Params map[string]string
	}

	Settings struct {
		Calling *Calling `json:"calling,omitempty"`
	}
	Client struct {
		configReader config.Reader
		base         *BaseClient
	}

	BaseClient struct {
		sender whttp.AnySender
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}

	Calling struct {
		Status                   string     `json:"status,omitempty"`
		CallIconVisibility       string     `json:"call_icon_visibility,omitempty"`
		CallbackPermissionStatus string     `json:"callback_permission_status,omitempty"`
		CallHours                *CallHours `json:"call_hours,omitempty"`
		SIP                      *SIP       `json:"sip,omitempty"`
	}

	CallHours struct {
		Status               string               `json:"status,omitempty"`
		TimezoneID           string               `json:"timezone_id,omitempty"` // e.g. "Europe/Berlin" or provider’s TZ id
		WeeklyOperatingHours []WeeklyOperatingDay `json:"weekly_operating_hours,omitempty"`
		HolidaySchedule      []Holiday            `json:"holiday_schedule,omitempty"`
	}

	WeeklyOperatingDay struct {
		DayOfWeek string `json:"day_of_week,omitempty"` // "MONDAY", ...
		OpenTime  string `json:"open_time,omitempty"`   // "HHMM" e.g., "0400"
		CloseTime string `json:"close_time,omitempty"`  // "HHMM" e.g., "1020"
	}

	Holiday struct {
		Date      string `json:"date,omitempty"`       // "YYYY-MM-DD"
		StartTime string `json:"start_time,omitempty"` // "HHMM"
		EndTime   string `json:"end_time,omitempty"`   // "HHMM"
	}

	SIP struct {
		Status  string      `json:"status,omitempty"`
		Servers []SIPServer `json:"servers,omitempty"`
	}

	SIPServer struct {
		Hostname        string `json:"hostname,omitempty"`
		SIPUserPassword string `json:"sip_user_password,omitempty"`
	}
)

func NewClient(configReader config.Reader, base *BaseClient) *Client {
	return &Client{
		configReader: configReader,
		base:         base,
	}
}

func (bc *Client) GetSettings(ctx context.Context, request *GetSettingsRequest) (*Settings, error) {
	conf, err := bc.configReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrUpdateSettings, err)
	}

	response, err := bc.base.GetSettings(ctx, conf, request)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrGetSettings, err)
	}

	return response, nil
}

func (bc *Client) UpdateSettings(ctx context.Context, settings *Settings) (*SuccessResponse, error) {
	conf, err := bc.configReader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: config read: %w", ErrUpdateSettings, err)
	}

	response, err := bc.base.UpdateSettings(ctx, conf, settings)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrUpdateSettings, err)
	}

	return response, nil
}

func NewBaseClient(sender whttp.AnySender) *BaseClient {
	return &BaseClient{sender: sender}
}

func (bc *BaseClient) GetSettings(
	ctx context.Context,
	conf *config.Config,
	request *GetSettingsRequest,
) (*Settings, error) {
	req := &baseRequest{
		Method: http.MethodGet,
		Type:   whttp.RequestTypeGetSettings,
		Params: request.Params,
	}

	response, err := bc.send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrGetSettings, err)
	}

	return response.settings(), nil
}

func (bc *BaseClient) UpdateSettings(
	ctx context.Context,
	conf *config.Config,
	settings *Settings,
) (*SuccessResponse, error) {
	req := &baseRequest{
		Method:  http.MethodPost,
		Type:    whttp.RequestTypeUpdateSettings,
		Payload: settings,
	}

	response, err := bc.send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrUpdateSettings, err)
	}

	return response.successResponse(), nil
}

func (bc *BaseClient) send(ctx context.Context, config *config.Config, request *baseRequest) (*baseResponse, error) {
	opts := []whttp.RequestOption[any]{
		whttp.WithRequestEndpoints[any](config.APIVersion, config.PhoneNumberID, settingsEndpoint),
		whttp.WithRequestQueryParams[any](request.Params),
		whttp.WithRequestBearer[any](config.AccessToken),
		whttp.WithRequestType[any](request.Type),
		whttp.WithRequestAppSecret[any](config.AppSecret),
		whttp.WithRequestSecured[any](config.SecureRequests),
		whttp.WithRequestDebugLogLevel[any](whttp.ParseDebugLogLevel(config.DebugLogLevel)),
	}

	if request.Payload != nil {
		opts = append(opts,
			whttp.WithRequestMessage(&request.Payload),
		)
	}

	req := whttp.MakeRequest(request.Method, config.BaseURL, opts...)

	response := &baseResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		DisallowUnknownFields: false,
		DisallowEmptyResponse: true,
		InspectResponseError:  true,
	})

	if err := bc.sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type (
	baseRequest struct {
		Method  string
		Type    whttp.RequestType
		Params  map[string]string
		Payload any
	}

	baseResponse struct {
		Success bool     `json:"success,omitempty"`
		Calling *Calling `json:"calling,omitempty"`
	}
)

func (r *baseResponse) settings() *Settings {
	return &Settings{Calling: r.Calling}
}

func (r *baseResponse) successResponse() *SuccessResponse {
	return &SuccessResponse{Success: r.Success}
}
