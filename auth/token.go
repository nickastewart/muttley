package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func generateMagicToken() (raw string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(bytes)
	return raw, hashMagicToken(raw), nil
}

func hashMagicToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
