package user

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const passwordIterations = 10000

// hashPassword returns a PBKDF2-SHA256 hash and algorithm descriptor.
func hashPassword(pw string) (string, string, error) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", "", err
	}
	hash := pbkdf2.Key([]byte(pw), salt[:], passwordIterations, 32, sha256.New)
	alg := fmt.Sprintf("pbkdf2-sha256:%d:%s", passwordIterations, hex.EncodeToString(salt[:]))
	return hex.EncodeToString(hash), alg, nil
}
