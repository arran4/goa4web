package admin

import (
	"net"
	"net/netip"
	"strings"
)

// normalizeIPNet trims and normalizes an IP or CIDR string.
func normalizeIPNet(ip string) string {
	ip = strings.TrimSpace(ip)
	if strings.Contains(ip, "/") {
		if p, err := netip.ParsePrefix(ip); err == nil {
			return p.String()
		}
		return ip
	}
	if parsed := net.ParseIP(ip); parsed != nil {
		if v4 := parsed.To4(); v4 != nil {
			return v4.String()
		}
		return parsed.String()
	}
	return ip
}
