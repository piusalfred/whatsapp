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

package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/piusalfred/whatsapp"
)

const ErrCreateAppSecretProof = whatsapp.Error("failed to create appsecret_proof")

// GenerateAppSecretProof generates the app secret proof required for secure API calls.
// It creates an HMAC-SHA-256 hash using the access token and the app secret.
func GenerateAppSecretProof(accessToken, appSecret string) (string, error) {
	if accessToken == "" || appSecret == "" {
		return "", fmt.Errorf("%w: access token and app secret are required", ErrCreateAppSecretProof)
	}

	h := hmac.New(sha256.New, []byte(appSecret))

	_, err := h.Write([]byte(accessToken))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateAppSecretProof, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
