/*
 *  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 *  and associated documentation files (the “Software”), to deal in the Software without restriction,
 *  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 *  subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all copies or substantial
 *  portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 *  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 *  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 *  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 *  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/piusalfred/whatsapp"
)

type (
	Config struct {
		BaseURL           string
		APIVersion        string
		AccessToken       string
		PhoneNumberID     string
		BusinessAccountID string
		AppSecret         string
		AppID             string
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

const ErrInvalidConfig = whatsapp.Error("invalid config")

func Validate(conf *Config) error {
	errs := make([]error, 0)
	if !whatsapp.IsCorrectAPIVersion(conf.APIVersion) {
		errVersion := fmt.Errorf("invalid API version: %s,lowest supported version is :%s",
			conf.APIVersion, whatsapp.LowestSupportedAPIVersion)

		errs = append(errs, errVersion)
	}

	if conf.SecureRequests && conf.AppSecret == "" {
		errs = append(errs, errors.New("app secret is required for secure requests"))
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
