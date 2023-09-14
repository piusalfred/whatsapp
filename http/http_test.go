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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Context struct {
	Method     string
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

func testServer(t *testing.T, ctx *Context) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != ctx.Method {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		for key, value := range ctx.Headers {
			w.Header().Add(key, value)
		}

		w.WriteHeader(ctx.StatusCode)
		if ctx.Body != nil {
			body, err := json.Marshal(ctx.Body)
			if err != nil {
				t.Errorf("failed to marshal response body: %v", err)

				return
			}
			if _, err := w.Write(body); err != nil {
				t.Errorf("failed to write response body: %v", err)
			}
		}
	})

	return httptest.NewServer(handler)
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Male bool   `json:"male"`
}

func TestSend(t *testing.T) { //nolint:paralleltest
	ctx := &Context{
		Method:     http.MethodGet,
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body: &User{
			Name: "Pius Alfred",
			Age:  77,
			Male: true,
		},
	}

	server := testServer(t, ctx)
	defer server.Close()

	reqCtx := &RequestContext{
		Name:          "test",
		BaseURL:       server.URL,
		ApiVersion:    "",
		PhoneNumberID: "",
		Endpoints:     nil,
	}

	request := &Request{
		Context: reqCtx,
		Method:  http.MethodGet,
		Headers: map[string]string{"Content-Type": "application/json"},
		Query:   nil,
		Bearer:  "",
		Form:    nil,
		Payload: nil,
	}

	var user User
	base := &Client{http: http.DefaultClient}
	if err := base.Do(context.TODO(), request, &user); err != nil {
		t.Errorf("failed to send request: %v", err)
	}

	// Compare the response body with the expected response body
	usr, ok := ctx.Body.(*User)
	if !ok {
		t.Errorf("failed to cast body to user type")
	}

	if user != *usr {
		t.Errorf("response body mismatch: got %v, want %v", user, *usr)
	}

	t.Logf("user: %+v", user)
}

func TestRequestNameFromContext(t *testing.T) {
	t.Parallel()
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test request name from context",
			args: args{
				name: "test",
			},
			want: "test",
		},
		{
			name: "test request name from context",
			args: args{
				name: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		args := tt.args
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := withRequestName(context.TODO(), args.name)
			if got := RequestNameFromContext(ctx); got != tt.want {
				t.Errorf("RequestNameFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractRequestBody(t *testing.T) {
	t.Parallel()
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		Wise bool   `json:"wise"`
	}
	type args struct {
		payload interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test []byte payload",
			args: args{
				payload: []byte("test"),
			},
			want:    []byte("test"),
			wantErr: false,
		},
		{
			name: "test string payload",
			args: args{
				payload: "test",
			},
			want:    []byte("test"),
			wantErr: false,
		},
		{
			name: "test struct payload",
			args: args{
				payload: user{
					Name: "test",
					Age:  10,
					Wise: true,
				},
			},
			want:    []byte(`{"name":"test","age":10,"wise":true}`),
			wantErr: false,
		},
		{
			name: "test pointer to struct payload",
			args: args{
				payload: &user{
					Name: "test",
					Age:  10,
					Wise: true,
				},
			},
			want:    []byte(`{"name":"test","age":10,"wise":true}`),
			wantErr: false,
		},
		{
			name: "slice of struct payload",
			args: args{
				payload: []user{
					{
						Name: "test",
						Age:  10,
						Wise: true,
					},
					{
						Name: "test2",
						Age:  20,
						Wise: false,
					},
				},
			},
			want:    []byte(`[{"name":"test","age":10,"wise":true},{"name":"test2","age":20,"wise":false}]`),
			wantErr: false,
		},
		{
			name: "slice of pointer to struct payload",
			args: args{
				payload: []*user{
					{
						Name: "test",
						Age:  10,
						Wise: true,
					},
					{
						Name: "test2",
						Age:  20,
						Wise: false,
					},
				},
			},

			want:    []byte(`[{"name":"test","age":10,"wise":true},{"name":"test2","age":20,"wise":false}]`),
			wantErr: false,
		},
		{
			name: "test nil payload",
			args: args{
				payload: nil,
			},
			want:    []byte(""),
			wantErr: false,
		},
		{
			name: "test invalid payload",
			args: args{
				payload: make(chan int),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		args := tt.args
		name := tt.name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := extractRequestBody(args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: extractRequestBody() error = %v, wantErr %v", name, err, tt.wantErr)

				return
			}

			if (tt.wantErr && (err != nil)) || got == nil {
				t.Logf("%s: extractPaylodFromRequest has error as expected: %v or returned a nil io.Reader", name, err)

				return
			}

			// create a new buffer to read the payload
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(got)
			if err != nil {
				t.Errorf("%s extractRequestBody() error = %v, wantErr %v", name, err, tt.wantErr)

				return
			}

			// Json encoder by default adds a new line at the end of the payload,
			// So we need to remove it to compare the payload.
			bytesGot := bytes.TrimRight(buf.Bytes(), "\n")
			if !bytes.Equal(bytesGot, tt.want) {
				t.Errorf("%s: extractRequestBody() got = %s, want %s", name,
					string(bytesGot), string(tt.want))
			}
		})
	}
}
