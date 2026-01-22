package common

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// passwordIterations sets the default PBKDF2 iteration count.
	passwordIterations = 10000
)

// HashPassword returns a PBKDF2-SHA256 hash of the password along with the
// algorithm descriptor storing iterations and salt.
func HashPassword(pw string) (string, string, error) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", "", err
	}
	hash := pbkdf2.Key([]byte(pw), salt[:], passwordIterations, 32, sha256.New)
	alg := fmt.Sprintf("pbkdf2-sha256:%d:%s", passwordIterations, hex.EncodeToString(salt[:]))
	return hex.EncodeToString(hash), alg, nil
}

// VerifyPassword checks the password against the stored hash and algorithm
// descriptor.
func VerifyPassword(pw, storedHash, alg string) bool {
	parts := strings.Split(alg, ":")
	switch parts[0] {
	case "pbkdf2-sha256":
		if len(parts) != 3 {
			return false
		}
		iter, err := strconv.Atoi(parts[1])
		if err != nil {
			return false
		}
		salt, err := hex.DecodeString(parts[2])
		if err != nil {
			return false
		}
		hash := pbkdf2.Key([]byte(pw), salt, iter, 32, sha256.New)
		return storedHash == hex.EncodeToString(hash)
	case "md5", "":
		sum := md5.Sum([]byte(pw))
		return storedHash == hex.EncodeToString(sum[:])
	default:
		return false
	}
}
