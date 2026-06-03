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

package flow_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piusalfred/whatsapp/flow"
)

// TestDataExchangeRoundTrip exercises the full encrypted data-exchange path the
// way Meta's client does it: AES-128-GCM with a 16-byte IV, AES key wrapped with
// RSA-OAEP/SHA-256, and the response decrypted with the same key and the
// bitwise-flipped IV. It is a regression guard for two bugs:
//   - aesGCM{Decrypt,Encrypt} must accept Meta's 16-byte IV (not Go's 12-byte
//     GCM default), otherwise Open/Seal panic on every real request;
//   - DecryptRequest must return the RSA-decrypted AES key, otherwise
//     EncryptResponse feeds the RSA ciphertext to aes.NewCipher.
func TestDataExchangeRoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	srv := newDataExchangeServer(t, key)
	defer srv.Close()

	aesKey, iv := newAESKeyAndIV(t)
	reqBody := encryptPingRequest(t, &key.PublicKey, aesKey, iv)
	respBody := postFlowRequest(t, srv.URL, reqBody)

	if status := decryptStatus(t, aesKey, iv, respBody); status != "active" {
		t.Fatalf("status: got %q, want active", status)
	}
}

func newDataExchangeServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	loader := func(context.Context) (*rsa.PrivateKey, error) { return key, nil }
	handler := flow.DataExchangeHandlerFunc(
		func(_ context.Context, req *flow.DataExchangeRequest) (*flow.Response, error) {
			if req.Action == "ping" {
				return flow.CreateHealthCheckResponse("active"), nil
			}
			return flow.CreateErrorAcknowledgmentResponse(true), nil
		},
	)
	impl := flow.NewDataExchangeHandler(loader, handler)
	return httptest.NewServer(http.HandlerFunc(impl.Handle))
}

func newAESKeyAndIV(t *testing.T) ([]byte, []byte) {
	t.Helper()
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		t.Fatalf("aes key: %v", err)
	}
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("iv: %v", err)
	}
	return aesKey, iv
}

func encryptPingRequest(t *testing.T, pub *rsa.PublicKey, aesKey, iv []byte) []byte {
	t.Helper()
	plaintext, err := json.Marshal(map[string]any{"version": "3.0", "action": "ping"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}
	sealed := gcm.Seal(nil, iv, plaintext, nil)

	encAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		t.Fatalf("rsa encrypt: %v", err)
	}

	reqBody, err := json.Marshal(flow.Request{
		EncryptedFlowData: base64.StdEncoding.EncodeToString(sealed),
		EncryptedAesKey:   base64.StdEncoding.EncodeToString(encAESKey),
		InitialVector:     base64.StdEncoding.EncodeToString(iv),
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return reqBody
}

func postFlowRequest(t *testing.T, url string, reqBody []byte) []byte {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}

func decryptStatus(t *testing.T, aesKey, iv, body []byte) string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}

	flipped := make([]byte, len(iv))
	for i, b := range iv {
		flipped[i] = b ^ 0xFF
	}
	out, err := gcm.Open(nil, flipped, raw, nil)
	if err != nil {
		t.Fatalf("decrypt response: %v", err)
	}

	var decoded struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return decoded.Data.Status
}
