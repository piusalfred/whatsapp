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
//
// Params:
// - accessToken: The access token used in the call.
// - appSecret: The app secret associated with the app.
//
// Returns:
// - string: The generated app secret proof.
// - error: Any error encountered during hash generation.
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
