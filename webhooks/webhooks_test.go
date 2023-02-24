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
	"errors"
	"testing"
)

func TestParseMessageType(t *testing.T) {
	t.Parallel()
	type args struct {
		messageType string
	}

	tests := []struct {
		name string
		args args
		want MessageType
	}{
		{
			name: "tExt",
			args: args{
				messageType: "text",
			},
			want: TextMessageType,
		},
		{
			name: "imageX",
			args: args{
				messageType: "imageX",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseMessageType(tt.args.messageType)
			if got != tt.want {
				t.Errorf("ParseMessageType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEncounteredError(t *testing.T) {
	t.Parallel()
	type args struct {
		nonFatalErrsMap map[string]error
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantErrString string
	}{
		{
			name: "empty",
			args: args{
				nonFatalErrsMap: map[string]error{},
			},
			wantErr:       false,
			wantErrString: "",
		},
		{
			name: "non single error",
			args: args{
				nonFatalErrsMap: map[string]error{
					"one": errors.New("single"),
				},
			},
			wantErr:       true,
			wantErrString: "one: single",
		},
		{
			name: "multiple errors",
			args: args{
				nonFatalErrsMap: map[string]error{
					"one":   errors.New("single"),
					"two":   errors.New("double"),
					"three": errors.New("triple"),
				},
			},
			wantErr:       true,
			wantErrString: "one: single, two: double, three: triple",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := getEncounteredError(tt.args.nonFatalErrsMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEncounteredError() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.wantErrString {
				t.Errorf("getEncounteredError() error = %v, wantErrString %v", err, tt.wantErrString)
			}
		})
	}
}
