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

package webhooks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_readNotificationBuffer(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name    string
		req     *http.Request
		want    *bytes.Buffer
		wantErr bool
	}

	tests := []testCase{
		{
			name: "normal request body",
			req:  httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{"name": "pius"}`)),
			want: bytes.NewBufferString(`{"name": "pius"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := readNotificationBuffer(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("readNotificationBuffer() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readNotificationBuffer() got = %v, want %v", got, tt.want)
			}

			// check if the request body is still intact
			buf := new(bytes.Buffer)
			if _, err = buf.ReadFrom(tt.req.Body); err != nil {
				t.Errorf("error reading request body: %v", err)

				return
			}

			if !reflect.DeepEqual(buf, tt.want) {
				t.Errorf("request body got = %v, want %v", buf, tt.want)
			}
		})
	}
}
