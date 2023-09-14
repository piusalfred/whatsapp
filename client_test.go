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

package whatsapp

import (
	"context"
	"fmt"
	"net/http"

	whttp "github.com/piusalfred/whatsapp/http"
)

func ExampleNewClient() {
	reader := ConfigReaderFunc(func(ctx context.Context) (*Config, error) {
		return &Config{
			BaseURL:           BaseURL,
			Version:           LowestSupportedVersion,
			AccessToken:       "[redacted]",
			PhoneNumberID:     "[redacted]",
			BusinessAccountID: "[redacted]",
		}, nil
	})

	base := &BaseClient{whttp.NewClient(
		whttp.WithHTTPClient(http.DefaultClient),
		whttp.WithRequestHooks(),
		whttp.WithResponseHooks(),
	)}

	client, err := NewClient(reader, WithBaseClient(base))
	if err != nil {
		panic(err)
	}

	fmt.Printf("BaseURL: %s\nVersion: %s", client.Config.BaseURL, client.Config.Version)

	// Output:
	// BaseURL: https://graph.facebook.com/
	// Version: v16.0
}
