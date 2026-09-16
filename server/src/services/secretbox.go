package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// secretBox encrypts provider credentials at rest. Provider tokens are
// long-lived, high-value secrets: they are never stored in clear text, never
// logged and never returned by the API.
type secretBox struct {
	aead cipher.AEAD
}

const secretBoxPrefix = "enc:v1:"

// newSecretBox derives a 32-byte AES key from the configured secret. The key is
// domain-separated with SHA-256 so that a JWT secret can be reused without
// weakening either use.
func newSecretBox(secret string) (*secretBox, error) {
	key := strings.TrimSpace(secret)
	if key == "" {
		return nil, errors.New("secret box: no encryption key configured")
	}
	derived := sha256.Sum256([]byte("code-platform/provider-credentials\x00" + key))
	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return nil, fmt.Errorf("secret box: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secret box: %w", err)
	}
	return &secretBox{aead: aead}, nil
}

// Seal encrypts a secret. An empty input stays empty so that optional tokens
// do not create noise in the database.
func (s *secretBox) Seal(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	if s == nil {
		return "", errors.New("secret box: unavailable")
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("secret box: %w", err)
	}
	sealed := s.aead.Seal(nil, nonce, []byte(value), nil)
	return secretBoxPrefix + base64.RawURLEncoding.EncodeToString(append(nonce, sealed...)), nil
}

// Open decrypts a previously sealed secret.
func (s *secretBox) Open(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if s == nil {
		return "", errors.New("secret box: unavailable")
	}
	if !strings.HasPrefix(value, secretBoxPrefix) {
		return "", errors.New("secret box: unsupported payload format")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, secretBoxPrefix))
	if err != nil {
		return "", fmt.Errorf("secret box: %w", err)
	}
	nonceSize := s.aead.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("secret box: truncated payload")
	}
	plaintext, err := s.aead.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("secret box: %w", err)
	}
	return string(plaintext), nil
}
