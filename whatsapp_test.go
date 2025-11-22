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

package whatsapp_test

import (
	"reflect"
	"testing"

	"github.com/piusalfred/whatsapp"
)

func TestIsCorrectAPIVersion(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name    string
		version string
		want    bool
	}
	tests := []testCase{
		{name: "valid version", version: "v20.1", want: true},
		{name: "valid version", version: "v21.0", want: true},
		{name: "invalid version below 20", version: "v15.9", want: false},
		{name: "valid three digit version", version: "v100.0", want: true},
		{name: "invalid zero version", version: "v0.0", want: false},
		{name: "invalid minor version", version: "v0.hello", want: false},
		{name: "invalid major version", version: "vhi.1", want: false},
		{name: "valid version", version: "v20.0", want: true},
		{name: "missing v prefix", version: "20.1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := whatsapp.IsCorrectAPIVersion(tt.version); got != tt.want {
				t.Errorf("IsCorrectAPIVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAPIVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		versionStr string
		wantMajor  int
		wantMinor  int
		wantErr    bool
	}{
		{name: "valid version v20.0", versionStr: "v20.0", wantMajor: 20, wantMinor: 0, wantErr: false},
		{name: "valid version v21.5", versionStr: "v21.5", wantMajor: 21, wantMinor: 5, wantErr: false},
		{name: "valid three digit major version", versionStr: "v100.10", wantMajor: 100, wantMinor: 10, wantErr: false},
		{name: "valid version v1.0", versionStr: "v1.0", wantMajor: 1, wantMinor: 0, wantErr: false},
		{name: "missing v prefix", versionStr: "20.0", wantMajor: 0, wantMinor: 0, wantErr: true},
		{name: "invalid format - no dot", versionStr: "v20", wantMajor: 0, wantMinor: 0, wantErr: true},
		{
			name:       "invalid major version - not a number",
			versionStr: "vhello.0",
			wantMajor:  0,
			wantMinor:  0,
			wantErr:    true,
		},
		{
			name:       "invalid minor version - not a number",
			versionStr: "v20.world",
			wantMajor:  0,
			wantMinor:  0,
			wantErr:    true,
		},
		{name: "empty string", versionStr: "", wantMajor: 0, wantMinor: 0, wantErr: true},
		{name: "only v prefix", versionStr: "v", wantMajor: 0, wantMinor: 0, wantErr: true},
		{
			name:       "negative major version",
			versionStr: "v-1.0",
			wantMajor:  0,
			wantMinor:  0,
			wantErr:    true,
		},
		{name: "negative minor version", versionStr: "v20.-1", wantMajor: 0, wantMinor: 0, wantErr: true},
		{name: "extra segments", versionStr: "v20.0.1", wantMajor: 0, wantMinor: 0, wantErr: true},
		{name: "spaces in version", versionStr: "v 20.0", wantMajor: 0, wantMinor: 0, wantErr: true},
		{name: "uppercase V", versionStr: "V20.0", wantMajor: 0, wantMinor: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := whatsapp.ParseAPIVersion(tt.versionStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAPIVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				want := &whatsapp.APIVersion{
					Major: tt.wantMajor,
					Minor: tt.wantMinor,
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("ParseAPIVersion() got = %v, want %v", got, want)
				}
			}
		})
	}
}
