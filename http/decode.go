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
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// DecodeFunc is a function type that decodes data from an io.Reader into a specific
// type of value. The function takes an io.Reader and a pointer to an empty interface{},
// and returns an error. The type of value to decode is determined by the type of the
// interface{} parameter. The function should modify the value pointed to by the
// interface{} parameter to represent the decoded value.
type DecodeFunc func(reader io.Reader, v interface{}) error

// ErrorDecoder is a DecodeFunc that decodes an error from an io.Reader. The function
// takes an io.Reader and a pointer to an empty interface{}, and returns an error. The
// type of value to decode is determined by the type of the interface{} parameter. Which
// in this case is errors.Error.
// Usage example:
//
//	var err werrors.Error
//	er := ErrorDecoder()(resp.Body, nil)
var ErrorDecoder = func(statusCode int) DecodeFunc { //nolint:gochecknoglobals
	decoder := func(reader io.Reader, _ interface{}) error {
		var val ResponseError
		err := json.NewDecoder(reader).Decode(&val)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("error decoding error: %w", err)
		}
		val.Code = statusCode

		return &val
	}

	return decoder
}
