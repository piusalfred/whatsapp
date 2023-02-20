// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this
// software and associated documentation files (the “Software”), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons
// to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or
// substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http

import "testing"

func TestCreateRequestURL(t *testing.T) {
	t.Parallel()
	type args struct {
		params *RequestParams
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test create phone number verification request url",
			args: args{
				params: &RequestParams{
					BaseURL:    BaseURL,
					SenderID:   "224225226",
					ApiVersion: "v16.0",
					Endpoints:  []string{"verify_code"},
				},
			},
			want:    "https://graph.facebook.com/v16.0/224225226/verify_code",
			wantErr: false,
		},
		{
			name: "test create media delete request url",
			args: args{
				params: &RequestParams{
					BaseURL:    BaseURL,
					SenderID:   "224225226", // this should be meda id
					ApiVersion: "v16.0",
					Endpoints:  nil,
				},
			},
			want:    "https://graph.facebook.com/v16.0/224225226",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CreateRequestURL(tt.args.params.BaseURL, tt.args.params.ApiVersion, tt.args.params.SenderID, tt.args.params.Endpoints...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRequestURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
