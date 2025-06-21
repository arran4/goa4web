package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// loadSessionSecret returns the session secret using the following priority:
//  1. cliSecret if non-empty
//  2. SESSION_SECRET environment variable
//  3. contents of the file at path. If path is empty it uses SESSION_SECRET_FILE
//     or a default file named ".session_secret" in the working directory.
//
// If the file does not exist, a new random secret is generated and saved.
func loadSessionSecret(cliSecret, path string) (string, error) {
	if cliSecret != "" {
		return cliSecret, nil
	}

	if env := os.Getenv("SESSION_SECRET"); env != "" {
		return env, nil
	}

	if path == "" {
		path = os.Getenv("SESSION_SECRET_FILE")
		if path == "" {
			path = ".session_secret"
		}
	}

	b, err := readFile(path)
	if err == nil {
		secret := strings.TrimSpace(string(b))
		if secret != "" {
			return secret, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	// Generate a new secret and store it
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(buf)
	if err := writeFile(path, []byte(secret), 0600); err != nil {
		return "", err
	}
	return secret, nil
}
