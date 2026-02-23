package middleware

import (
	"ctweb/internal/config"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

type RequestMeta struct {
	RemoteAddr      string
	RealIP          string
	EffectiveScheme string
	EffectiveHost   string
	TrustedProxy    bool
}

func ExtractRequestMeta(c *gin.Context) RequestMeta {
	meta := RequestMeta{
		RemoteAddr:      remoteAddrHost(c),
		EffectiveScheme: requestScheme(c),
		EffectiveHost:   requestHost(c),
	}
	meta.RealIP = meta.RemoteAddr

	cfg := config.Get()
	if shouldTrustForwardHeaders(c, cfg) {
		meta.TrustedProxy = true
		if clientIP := strings.TrimSpace(c.ClientIP()); clientIP != "" {
			meta.RealIP = clientIP
		}
		if proto := firstCSVToken(c.GetHeader("X-Forwarded-Proto")); proto != "" {
			meta.EffectiveScheme = strings.ToLower(proto)
		}
		if host := firstCSVToken(c.GetHeader("X-Forwarded-Host")); host != "" {
			meta.EffectiveHost = host
		}
	}

	if meta.RealIP == "" {
		meta.RealIP = meta.RemoteAddr
	}
	return meta
}

func shouldTrustForwardHeaders(c *gin.Context, cfg *config.Config) bool {
	if c == nil || cfg == nil {
		return false
	}
	if !cfg.Proxy.Enabled || !cfg.Proxy.TrustForwardHeaders {
		return false
	}
	return isTrustedCIDR(remoteAddrHost(c), cfg.Proxy.TrustedCIDRs)
}

func requestScheme(c *gin.Context) string {
	if c != nil && c.Request != nil && c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return strings.TrimSpace(c.Request.Host)
}

func remoteAddrHost(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	remoteAddr := strings.TrimSpace(c.Request.RemoteAddr)
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func isTrustedCIDR(ipStr string, cidrs []string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func firstCSVToken(value string) string {
	parts := strings.Split(value, ",")
	for _, p := range parts {
		token := strings.TrimSpace(p)
		if token != "" {
			return token
		}
	}
	return ""
}
