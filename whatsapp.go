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
	"fmt"
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
	version, err := ParseAPIVersion(apiVersion)
	if err != nil {
		return false
	}

	return version.Major >= lowestMajorVersion && version.Minor >= 0
}

type Error string

func (e Error) Error() string {
	return string(e)
}

type APIVersion struct {
	Major int
	Minor int
}

func (v APIVersion) String() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

const ErrAPIVersionInvalid Error = "API version is invalid"

func ParseAPIVersion(versionStr string) (*APIVersion, error) {
	reg := regexp.MustCompile(`^v(?P<major_version>\d+)\.(?P<minor_version>\d+)$`)
	matches := reg.FindStringSubmatch(versionStr)
	if len(matches) != 3 { //nolint:mnd // ok
		return nil, fmt.Errorf("%w: %s", ErrAPIVersionInvalid, versionStr)
	}

	majorStr := matches[1]
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return nil, fmt.Errorf("%w: major version should be a number: %s", ErrAPIVersionInvalid, versionStr)
	}

	minorStr := matches[2]
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return nil, fmt.Errorf("%w: minor version should be a number: %s", ErrAPIVersionInvalid, versionStr)
	}

	version := &APIVersion{
		Major: major,
		Minor: minor,
	}

	return version, nil
}
