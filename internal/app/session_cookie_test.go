package app

import (
	"net/http"
	"testing"
)

func TestSessionCookieOptions(t *testing.T) {
	t.Run("Happy Path - HTTPS", func(t *testing.T) {
		options := sessionCookieOptions("https://example.com", http.SameSiteStrictMode)
		if !options.Secure {
			t.Error("HTTPS session cookie is not secure")
		}
	})

	t.Run("HTTP development", func(t *testing.T) {
		options := sessionCookieOptions("http://localhost:8080", http.SameSiteStrictMode)
		if options.Secure {
			t.Error("HTTP session cookie is marked secure and will not be returned by the browser")
		}
	})
}
