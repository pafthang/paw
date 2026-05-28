package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/pafthang/paw/internal/config"
)

func EnsureToken() (string, error) {
	path, err := config.AccessTokenPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token, nil
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return RotateToken()
}

func ReadToken() (string, error) {
	path, err := config.AccessTokenPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func RotateToken() (string, error) {
	if err := config.EnsureDir(); err != nil {
		return "", err
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	path, err := config.AccessTokenPath()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", err
	}
	return token, nil
}

func Check(provided string) bool {
	expected, err := ReadToken()
	if err != nil || expected == "" {
		return false
	}
	return strings.TrimSpace(provided) == expected
}
