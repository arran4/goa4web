package middleware

import (
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/assert"
)

func TestRequestIPSpoofing_Untrusted(t *testing.T) {
	// Case 1: No trusted proxies configured (default).
	// requestIP should ignore X-Forwarded-For.

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234" // Real IP
	req.Header.Set("X-Forwarded-For", "6.6.6.6") // Spoofed IP

	cfg := &config.RuntimeConfig{} // Empty config
	ip := requestIP(req, cfg)

	assert.Equal(t, "1.2.3.4", ip, "Expected Real IP (ignoring spoofed header)")
}

func TestRequestIPSpoofing_Trusted(t *testing.T) {
	// Case 2: Trusted proxy.
	// We trust 10.0.0.1.

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234" // Trusted Proxy IP
	req.Header.Set("X-Forwarded-For", "6.6.6.6") // Client IP

	cfg := &config.RuntimeConfig{
		TrustedProxies: "10.0.0.1/32",
		TrustedProxiesParsed: []netip.Prefix{
			netip.MustParsePrefix("10.0.0.1/32"),
		},
	}
	ip := requestIP(req, cfg)

	assert.Equal(t, "6.6.6.6", ip, "Expected Client IP (via trusted proxy)")
}

func TestRequestIPSpoofing_TrustedChain(t *testing.T) {
	// Case 3: Trusted proxy chain.
	// We trust 10.0.0.1 and 10.0.0.2.
	// Request path: Client(1.2.3.4) -> Proxy1(10.0.0.2) -> Proxy2(10.0.0.1) -> Us.
	// Header at Us: "1.2.3.4, 10.0.0.2". RemoteAddr: 10.0.0.1.

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.2")

	cfg := &config.RuntimeConfig{
		TrustedProxiesParsed: []netip.Prefix{
			netip.MustParsePrefix("10.0.0.1/32"),
			netip.MustParsePrefix("10.0.0.2/32"),
		},
	}
	ip := requestIP(req, cfg)

	assert.Equal(t, "1.2.3.4", ip, "Expected Client IP (traversing trusted chain)")
}

func TestRequestIPSpoofing_UntrustedInChain(t *testing.T) {
	// Case 4: Untrusted proxy in chain.
	// Request path: Client(1.2.3.4) -> UntrustedProxy(6.6.6.6) -> TrustedProxy(10.0.0.1) -> Us.
	// Header: "1.2.3.4, 6.6.6.6". RemoteAddr: 10.0.0.1.

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 6.6.6.6")

	cfg := &config.RuntimeConfig{
		TrustedProxiesParsed: []netip.Prefix{
			netip.MustParsePrefix("10.0.0.1/32"),
		},
	}
	ip := requestIP(req, cfg)

	assert.Equal(t, "6.6.6.6", ip, "Expected Untrusted Proxy IP (client)")
}

func TestRequestIPSpoofing_GarbageHeader(t *testing.T) {
    // Case 5: Garbage in header
    req := httptest.NewRequest("GET", "/", nil)
    req.RemoteAddr = "10.0.0.1:1234"
    req.Header.Set("X-Forwarded-For", "garbage")

    cfg := &config.RuntimeConfig{
        TrustedProxiesParsed: []netip.Prefix{
            netip.MustParsePrefix("10.0.0.1/32"),
        },
    }
    ip := requestIP(req, cfg)
    // Should fallback to RemoteAddr because parsing failed?
    // Wait, my logic: netip.ParseAddr("garbage") fails. Returns normalizeIP(currentIP.String()).
    // currentIP is RemoteAddr.
    assert.Equal(t, "10.0.0.1", ip)
}
