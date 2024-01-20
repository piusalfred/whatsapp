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
	"context"
	"reflect"
	"testing"
)

func TestRetrieveRequestContext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input *RequestContext
	}{
		{
			name: "should return the request context",
			input: &RequestContext{
				ID:       "BF87030B-3A3B-449E-97F3-08BEA461E2BB",
				Action:   RequestActionSend,
				Category: RequestCategoryMessage,
				Name:     RequestNameTextMessage,
				Metadata: map[string]string{
					"platform": "whatsapp",
					"client":   "go",
					"version":  "v1.0.0",
					"plan":     "free",
				},
			},
		},
		{
			name:  "put nil in the context",
			input: nil,
		},
		{
			name:  "put empty in the context",
			input: &RequestContext{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := attachRequestContext(context.Background(), tt.input)
			if got := RetrieveRequestContext(ctx); !reflect.DeepEqual(got, tt.input) {
				t.Errorf("RetrieveRequestContext() = %v, want %v", got, tt.input)
			}
		})
	}
}
