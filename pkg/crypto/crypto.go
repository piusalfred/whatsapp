//  Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy of this software
//  and associated documentation files (the "Software"), to deal in the Software without restriction,
//  including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
//  and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
//  subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in all copies or substantial
//  portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
//  LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
//  IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
//  SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package crypto provides cryptographic utilities for WhatsApp Cloud API request
// authentication. It generates appsecret_proof values ([GenerateAppSecretProof])
// and validates X-Hub-Signature-256 headers ([VerifySignature]) for webhook
// payloads.
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var ErrCreateAppSecretProof = errors.New("failed to create appsecret_proof")

// Signer produces appsecret_proof values for secure API requests. The default
// implementation is [HMACSHA256Signer]; inject a mock in tests via the [Signer]
// field on types that need proof generation.
type Signer interface {
	Sign(accessToken, appSecret string) (string, error)
}

// HMACSHA256Signer implements [Signer] using HMAC-SHA-256.
type HMACSHA256Signer struct{}

// Sign computes the HMAC-SHA-256 of accessToken keyed with appSecret and
// returns the hex-encoded result.
func (HMACSHA256Signer) Sign(accessToken, appSecret string) (string, error) {
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

// GenerateAppSecretProof generates the app secret proof required for secure API calls.
// It creates an HMAC-SHA-256 hash using the access token and the app secret.
// This is a convenience wrapper around [HMACSHA256Signer.Sign].
func GenerateAppSecretProof(accessToken, appSecret string) (string, error) {
	return HMACSHA256Signer{}.Sign(accessToken, appSecret)
}
