package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const AgentSecret = "replace-with-a-long-random-secret-string"

func ValidateAgentRegistration(uuid, fingerprint, timestamp, signature string) bool {
	mac := hmac.New(sha256.New, []byte(AgentSecret))
	mac.Write([]byte(uuid + "|" + fingerprint + "|" + timestamp))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func NewSessionToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}
