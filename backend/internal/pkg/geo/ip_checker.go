package geo

import (
	"fmt"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// IPChecker checks IP geolocation using MaxMind GeoIP2
type IPChecker struct {
	db          *geoip2.Reader
	blockChina  bool
	mu          sync.RWMutex
}

// NewIPChecker creates a new IP checker
func NewIPChecker(dbPath string, blockChina bool) (*IPChecker, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoIP2 database: %w", err)
	}

	return &IPChecker{
		db:         db,
		blockChina: blockChina,
	}, nil
}

// Close closes the GeoIP2 database
func (c *IPChecker) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// CheckIP checks if an IP address should be blocked
func (c *IPChecker) CheckIP(ipStr string) (bool, string, error) {
	if !c.blockChina {
		return false, "", nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Allow private/local IPs
	if isPrivateIP(ip) {
		return false, "", nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	record, err := c.db.Country(ip)
	if err != nil {
		// If we can't determine location, allow access
		return false, "", nil
	}

	// Block China mainland (CN), but allow Hong Kong (HK), Macau (MO), Taiwan (TW)
	if record.Country.IsoCode == "CN" {
		return true, record.Country.Names["zh-CN"], nil
	}

	return false, "", nil
}

// GetCountry returns the country for an IP address
func (c *IPChecker) GetCountry(ipStr string) (string, string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	if isPrivateIP(ip) {
		return "LOCAL", "Local Network", nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	record, err := c.db.Country(ip)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup IP: %w", err)
	}

	return record.Country.IsoCode, record.Country.Names["en"], nil
}

// isPrivateIP checks if an IP is private/local
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}

	// Check for local network ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}

	return false
}

// GetClientIP extracts the real client IP from request headers
func GetClientIP(remoteAddr string, xForwardedFor, xRealIP string) string {
	// Try X-Real-IP first
	if xRealIP != "" {
		ip := net.ParseIP(xRealIP)
		if ip != nil {
			return ip.String()
		}
	}

	// Try X-Forwarded-For
	if xForwardedFor != "" {
		ips := parseXForwardedFor(xForwardedFor)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return ip
}

// parseXForwardedFor parses X-Forwarded-For header
func parseXForwardedFor(header string) []string {
	var ips []string
	for _, ip := range splitAndTrim(header, ",") {
		if parsed := net.ParseIP(ip); parsed != nil {
			ips = append(ips, parsed.String())
		}
	}
	return ips
}

// splitAndTrim splits a string and trims whitespace
func splitAndTrim(s, sep string) []string {
	var result []string
	for i := 0; i < len(s); {
		end := i
		for end < len(s) && s[end] != sep[0] {
			end++
		}
		part := s[i:end]
		// Trim spaces
		start := 0
		for start < len(part) && part[start] == ' ' {
			start++
		}
		end = len(part)
		for end > start && part[end-1] == ' ' {
			end--
		}
		if end > start {
			result = append(result, part[start:end])
		}
		i += len(part)
		if i < len(s) {
			i++ // Skip separator
		}
	}
	return result
}
