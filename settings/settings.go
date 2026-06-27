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
	"errors"
	"fmt"
	"net/http"

	"github.com/piusalfred/whatsapp/config"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const settingsEndpoint = "/settings"

var (
	ErrGetSettings    = errors.New("failed to get settings")
	ErrUpdateSettings = errors.New("failed to update settings")
)

type (
	GetSettingsRequest struct {
		Params map[string]string
	}

	Settings struct {
		Calling *Calling `json:"calling,omitempty"`
	}

	// Client orchestrates high-level Settings API operations.
	Client struct {
		sender *BaseClient
		config *config.Config
	}

	// BaseClient is the low-level HTTP executor for the Settings API. It accepts a
	// concrete [*config.Config] per request, making it suitable for multi-tenant
	// SaaS scenarios. For a fixed-configuration client, use [Client].
	BaseClient struct {
		whttp.BaseClient[any]
	}

	SuccessResponse struct {
		Success bool `json:"success"`
	}

	Calling struct {
		Status                   string        `json:"status,omitempty"`
		CallIconVisibility       string        `json:"call_icon_visibility,omitempty"`
		CallbackPermissionStatus string        `json:"callback_permission_status,omitempty"`
		CallHours                *CallHours    `json:"call_hours,omitempty"`
		SIP                      *SIP          `json:"sip,omitempty"`
		CallIcons                *CallIcons    `json:"call_icons,omitempty"`
		Audio                    *Audio        `json:"audio,omitempty"`
		Voicemail                *Voicemail    `json:"voicemail,omitempty"`
		Restrictions             *Restrictions `json:"restrictions,omitempty"`
	}

	// CallIcons configures country-based visibility of the call button.
	CallIcons struct {
		RestrictToUserCountries []string `json:"restrict_to_user_countries,omitempty"`
	}

	// Audio defines additional codecs supported for calls (Opus is always present).
	Audio struct {
		AdditionalCodecs []string `json:"additional_codecs,omitempty"` // "PCMA", "PCMU"
	}

	// Voicemail config (alpha feature).
	Voicemail struct {
		Status   string          `json:"status,omitempty"`   // "ENABLED", "DISABLED"
		Triggers []string        `json:"triggers,omitempty"` // "REJECT", "TIMEOUT"
		Audio    *VoicemailAudio `json:"audio,omitempty"`
	}

	VoicemailAudio struct {
		Default *VoicemailDefault `json:"default,omitempty"`
	}

	VoicemailDefault struct {
		AnnouncementMediaID int64 `json:"announcement_media_id,omitempty"`
		TimeoutSeconds      int   `json:"timeout_seconds,omitempty"` // 0-30, required for TIMEOUT trigger
	}

	// Restrictions appears only in GET responses when a restriction is active.
	Restrictions struct {
		RestrictionsList []*Restriction `json:"restrictions_list,omitempty"`
	}

	Restriction struct {
		Type       string `json:"type,omitempty"` // "RESTRICTED_BUSINESS_INITIATED_CALLING", "RESTRICTED_USER_INITIATED_CALLING"
		Reason     string `json:"reason,omitempty"`
		Expiration int64  `json:"expiration,omitempty"` // UNIX timestamp
	}

	// SIPServer – add missing RequestURIUserParams field.
	SIPServer struct {
		Hostname             string            `json:"hostname,omitempty"`
		SIPUserPassword      string            `json:"sip_user_password,omitempty"`
		RequestURIUserParams map[string]string `json:"request_uri_user_params,omitempty"` // NEW
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
)

// NewClient creates a high-level Client for the Settings API.
func NewClient(conf *config.Config, options ...whttp.CoreSenderOption) *Client {
	return &Client{
		sender: &BaseClient{BaseClient: *whttp.NewBaseClient[any](options...)},
		config: conf,
	}
}

// SetBaseClient replaces the underlying request sender.
func (c *Client) SetBaseClient(sender whttp.Sender[any]) {
	c.sender.SetSender(sender)
}

// SetMiddlewares wraps the underlying Sender with the provided middlewares.
// Middlewares are applied in order: middlewares[0] runs outermost.
func (c *Client) SetMiddlewares(mws ...whttp.Middleware[any]) {
	c.sender.SetMiddlewares(mws...)
}

// GetSettings retrieves calling settings.
func (c *Client) GetSettings(ctx context.Context, request *GetSettingsRequest) (*Settings, error) {
	req := &BaseRequest{
		Method: http.MethodGet,
		Type:   whttp.RequestTypeGetSettings,
		Params: request.Params,
	}
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrGetSettings, err)
	}
	return resp.settings(), nil
}

// UpdateSettings updates calling settings.
func (c *Client) UpdateSettings(ctx context.Context, settings *Settings) (*SuccessResponse, error) {
	req := &BaseRequest{
		Method:  http.MethodPost,
		Type:    whttp.RequestTypeUpdateSettings,
		Payload: settings,
	}
	resp, err := c.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrUpdateSettings, err)
	}
	return resp.successResponse(), nil
}

// Send dispatches a raw BaseRequest through the underlying BaseClient.
func (c *Client) Send(ctx context.Context, request *BaseRequest) (*BaseResponse, error) {
	response, err := c.sender.Send(ctx, c.config, request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return response, nil
}

func (bc *BaseClient) GetSettings(
	ctx context.Context,
	conf *config.Config,
	request *GetSettingsRequest,
) (*Settings, error) {
	req := &BaseRequest{
		Method: http.MethodGet,
		Type:   whttp.RequestTypeGetSettings,
		Params: request.Params,
	}

	response, err := bc.Send(ctx, conf, req)
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
	req := &BaseRequest{
		Method:  http.MethodPost,
		Type:    whttp.RequestTypeUpdateSettings,
		Payload: settings,
	}

	response, err := bc.Send(ctx, conf, req)
	if err != nil {
		return nil, fmt.Errorf("%w: send request: %w", ErrUpdateSettings, err)
	}

	return response.successResponse(), nil
}

func (bc *BaseClient) Send(ctx context.Context, conf *config.Config, request *BaseRequest) (*BaseResponse, error) {
	b := whttp.NewRequestBuilder(request.Method, conf.BaseURL).
		Auth(conf.AuthConfig()).
		Type(request.Type).
		Endpoints(conf.APIVersion, conf.PhoneNumberID, settingsEndpoint)

	if len(request.Params) > 0 {
		b = b.QueryParams(request.Params)
	}

	var msg *any
	if request.Payload != nil {
		msg = &request.Payload
	}

	req := whttp.Build[any](b, msg)

	response := &BaseResponse{}

	decoder := whttp.ResponseDecoderJSON(response, whttp.DecodeOptions{
		Flags: whttp.JSONDecodeDisallowEmptyResponse | whttp.JSONDecodeInspectResponseError,
	})

	if err := bc.Sender.Send(ctx, req, decoder); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}

type (
	BaseRequest struct {
		Method  string
		Type    whttp.RequestType
		Params  map[string]string
		Payload any
	}

	BaseResponse struct {
		Success bool     `json:"success,omitempty"`
		Calling *Calling `json:"calling,omitempty"`
	}
)

func (r *BaseResponse) settings() *Settings {
	return &Settings{Calling: r.Calling}
}

func (r *BaseResponse) successResponse() *SuccessResponse {
	return &SuccessResponse{Success: r.Success}
}
