package auth

import (
	"fmt"
	"os"
	"strings"
)

func SigningKeyFromEnv() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		return nil, fmt.Errorf("auth: JWT_SECRET must be at least 32 bytes")
	}
	return []byte(secret), nil
}
