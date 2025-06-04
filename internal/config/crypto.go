package config

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

// EncryptBytes returns the encryption key configured.
func (c *Config) EncryptBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.EncryptKey)
}

// Hash returns the sha256 hash of the configuration in a standard base64 encoded string
func (c *Config) Hash() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}
