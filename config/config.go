/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the "Software"), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

// Package config provides configuration loading and validation for the WhatsApp
// Cloud API. Use [Validate] or [ReadValidate] to reject invalid configurations
// early. The [Reader] interface supports dynamic configuration sources (env vars,
// files, remote services) via the [ReaderFunc] adapter.
package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/piusalfred/whatsapp"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type (
	// Config holds all parameters needed to authenticate and target WhatsApp Cloud
	// API requests. Required fields (AccessToken, PhoneNumberID) must be set;
	// optional fields may be left empty. Use [Validate] to reject invalid
	// configurations early.
	//
	// DebugLogLevel accepts "all", "info", or "warning" (case-insensitive).
	// Any other value disables debug output. See [whttp.DebugLogLevel] for the
	// authoritative list of constants.
	Config struct {
		BaseURL           string
		APIVersion        string
		AccessToken       string
		PhoneNumberID     string
		BusinessAccountID string
		AppSecret         string
		AppID             string
		SystemUserID      string
		SecureRequests    bool
		DebugLogLevel     string
	}

	Reader interface {
		Read(ctx context.Context) (*Config, error)
	}

	ReaderFunc func(ctx context.Context) (*Config, error)
)

func (fn ReaderFunc) Read(ctx context.Context) (*Config, error) {
	return fn(ctx)
}

var (
	ErrInvalidConfig       = errors.New("invalid config")
	ErrInvalidAPIVersion   = errors.New("invalid API version")
	ErrAppSecretRequired   = errors.New("app secret is required for secure requests")
	ErrAccessTokenRequired = errors.New("access token is required")
	ErrPhoneNumberRequired = errors.New("phone number ID is required")
)

// AuthConfig returns the authentication fields as a [whttp.AuthConfig] suitable
// for [whttp.RequestBuilder.Auth]. Domain packages use this to bridge their
// [Config] to the transport layer without importing the HTTP package.
func (c *Config) AuthConfig() whttp.AuthConfig {
	return whttp.AuthConfig{
		AccessToken: c.AccessToken,
		AppSecret:   c.AppSecret,
		Secure:      c.SecureRequests,
		DebugLevel:  whttp.ParseDebugLogLevel(c.DebugLogLevel),
	}
}

func Validate(conf *Config) error {
	errs := make([]error, 0)
	if !whatsapp.IsCorrectAPIVersion(conf.APIVersion) {
		errs = append(errs, fmt.Errorf("%w: got %s, minimum is %s",
			ErrInvalidAPIVersion, conf.APIVersion, whatsapp.LowestSupportedAPIVersion))
	}
	if conf.AccessToken == "" {
		errs = append(errs, ErrAccessTokenRequired)
	}
	if conf.PhoneNumberID == "" {
		errs = append(errs, ErrPhoneNumberRequired)
	}
	if conf.SecureRequests && conf.AppSecret == "" {
		errs = append(errs, ErrAppSecretRequired)
	}

	return errors.Join(errs...)
}

func ReadValidate(ctx context.Context, reader Reader) (*Config, error) {
	conf, err := reader.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: read config: %w", ErrInvalidConfig, err)
	}

	return conf, Validate(conf)
}
