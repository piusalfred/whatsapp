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

package whatsapp

import (
	"regexp"
	"strconv"
)

const (
	BaseURL                   = "https://graph.facebook.com"
	LowestSupportedAPIVersion = "v20.0" // This is the lowest version of the API that is supported
	MessageProduct            = "whatsapp"
	lowestMajorVersion        = 20
)

// IsCorrectAPIVersion checks if the provided API version string is valid and supported.
// The version string should be in the format "v<major_version>.<minor_version>".
// It returns true if the major version is 16 or higher, otherwise false.
func IsCorrectAPIVersion(apiVersion string) bool {
	reg := regexp.MustCompile(`^v(?P<major_version>\d+)\.(?P<minor_version>\d+)$`)
	matches := reg.FindStringSubmatch(apiVersion)
	if len(matches) != 3 { //nolint:mnd // ok
		return false
	}

	majorStr := matches[1]
	major, _ := strconv.Atoi(majorStr)
	if major < lowestMajorVersion {
		return false
	}

	minorStr := matches[2]
	minor, _ := strconv.Atoi(minorStr)

	return minor >= 0
}

type Error string

func (e Error) Error() string {
	return string(e)
}
