/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package config

import "context"

type (
	// Values is a struct that holds the configuration for the whatsapp client.
	// It is used to create a new whatsapp client.
	Values struct {
		BaseURL           string
		Version           string
		AccessToken       string
		PhoneNumberID     string
		BusinessAccountID string
	}

	// Reader is an interface that can be used to read the configuration
	// from a file or any other source.
	Reader interface {
		Read(ctx context.Context) (*Values, error)
	}

	ReaderFunc func(ctx context.Context) (*Values, error)
)

// Read implements the Reader interface.
func (fn ReaderFunc) Read(ctx context.Context) (*Values, error) {
	return fn(ctx)
}
