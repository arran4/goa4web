package main

import (
	"crypto/rand"
)

// randomString returns a random alphabetic string of the given length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	for i, v := range b {
		b[i] = letters[int(v)%len(letters)]
	}
	return string(b)
}
