package config

import (
	"testing"
	"time"
)

func baseConfig() *Config {
	return &Config{
		Server: ServerConfig{Port: 8443, TLS: TLSServerConfig{Enabled: false}},
		Databases: DatabasesConfig{
			System: DatabaseTargetConfig{
				Engine: "mysql",
				MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
			},
		},
		Security: SecurityConfig{
			CSRF:    CSRFConfig{Secret: "test-csrf-secret"},
			Session: SessionConfig{Secret: "test-session-secret"},
		},
	}
}

func TestValidateTLSModes(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid with tls disabled and empty cert paths",
			cfg: func() *Config {
				cfg := baseConfig()
				cfg.Server.Timeouts = TimeoutConfig{Read: 60 * time.Second, Write: 60 * time.Second, Idle: 120 * time.Second, ReadHeader: 5 * time.Second, ShutdownGrace: 10 * time.Second}
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "valid with tls enabled and cert paths",
			cfg: func() *Config {
				cfg := baseConfig()
				cfg.Server.TLS = TLSServerConfig{Enabled: true, CertPath: "pki/server/web-ui.crt", KeyPath: "pki/server/web-ui.key"}
				cfg.Server.Timeouts = TimeoutConfig{Read: 60 * time.Second, Write: 60 * time.Second, Idle: 120 * time.Second, ReadHeader: 5 * time.Second, ShutdownGrace: 10 * time.Second}
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "invalid tls enabled without cert path",
			cfg: func() *Config {
				cfg := baseConfig()
				cfg.Server.TLS = TLSServerConfig{Enabled: true, KeyPath: "pki/server/web-ui.key"}
				cfg.Server.Timeouts = TimeoutConfig{Read: 60 * time.Second, Write: 60 * time.Second, Idle: 120 * time.Second, ReadHeader: 5 * time.Second, ShutdownGrace: 10 * time.Second}
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid tls enabled without key path",
			cfg: func() *Config {
				cfg := baseConfig()
				cfg.Server.TLS = TLSServerConfig{Enabled: true, CertPath: "pki/server/web-ui.crt"}
				cfg.Server.Timeouts = TimeoutConfig{Read: 60 * time.Second, Write: 60 * time.Second, Idle: 120 * time.Second, ReadHeader: 5 * time.Second, ShutdownGrace: 10 * time.Second}
				return cfg
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimeoutDefaults(t *testing.T) {
	cfg := baseConfig()
	cfg.RateLimit = RateLimitConfig{Login: LoginRateLimitConfig{RequestsPerMinute: 5, Burst: 5}, API: APIRateLimitConfig{RequestsPerSecond: 100, Burst: 100}}

	if err := validate(cfg); err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.Server.Timeouts.Read == 0 || cfg.Server.Timeouts.Write == 0 || cfg.Server.Timeouts.Idle == 0 || cfg.Server.Timeouts.ReadHeader == 0 || cfg.Server.Timeouts.ShutdownGrace == 0 {
		t.Fatalf("expected default server.timeouts to be populated, got %+v", cfg.Server.Timeouts)
	}
	if cfg.Server.Limits.MaxHeaderBytes == 0 {
		t.Fatalf("expected default server.limits.max_header_bytes to be populated")
	}
}

func TestValidateTimeoutBounds(t *testing.T) {
	cfg := baseConfig()
	cfg.Server.Timeouts = TimeoutConfig{Read: -1 * time.Second}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for negative server.timeouts.read")
	}
}

func TestValidateMaxHeaderBytesBounds(t *testing.T) {
	cfg := baseConfig()
	cfg.Server.Limits = LimitsConfig{MaxHeaderBytes: 1024}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for too small server.limits.max_header_bytes")
	}
}

func TestValidateHTTP2Invalid(t *testing.T) {
	cfg := baseConfig()
	cfg.Server.HTTP2 = &HTTP2Config{MaxFrameSize: "invalid"}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for invalid server.http2")
	}
}

func TestValidateRateLimitFallbackFromSecurity(t *testing.T) {
	cfg := baseConfig()
	cfg.Security = SecurityConfig{
		CSRF:           CSRFConfig{Secret: "test-csrf-secret"},
		Session:        SessionConfig{Secret: "test-session-secret"},
		RateLimitLogin: 7,
		RateLimitAPI:   55,
	}

	if err := validate(cfg); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
	if cfg.RateLimit.Login.RequestsPerMinute != 7 {
		t.Fatalf("expected login fallback=7, got %d", cfg.RateLimit.Login.RequestsPerMinute)
	}
	if cfg.RateLimit.API.RequestsPerSecond != 55 {
		t.Fatalf("expected api fallback=55, got %d", cfg.RateLimit.API.RequestsPerSecond)
	}
}

func TestValidateRateLimitConflictBetweenLegacyAndNew(t *testing.T) {
	cfg := baseConfig()
	cfg.Security.RateLimitLogin = 7
	cfg.RateLimit.Login.RequestsPerMinute = 5

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail when security.rate_limit_login conflicts with rate_limit.login.requests_per_minute")
	}
}

func TestValidateSecuritySameSiteNoneRequiresSecure(t *testing.T) {
	cfg := baseConfig()
	cfg.Security.Session.CookieSameSite = "None"
	cfg.Security.Session.CookieSecure = false
	cfg.Server.TLS.Enabled = false

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for SameSite=None without secure cookies or TLS")
	}
}

func TestValidateSecurityBcryptRange(t *testing.T) {
	cfg := baseConfig()
	cfg.Security.BcryptCost = 15

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for security.bcrypt_cost out of range")
	}
}

func TestValidateSecurityDefaults(t *testing.T) {
	cfg := baseConfig()
	cfg.Security.BcryptCost = 0
	cfg.Security.Session.CookieSameSite = ""

	if err := validate(cfg); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
	if cfg.Security.BcryptCost != 10 {
		t.Fatalf("expected security.bcrypt_cost default=10, got %d", cfg.Security.BcryptCost)
	}
	if cfg.Security.Session.CookieSameSite != "Lax" {
		t.Fatalf("expected security.session.cookie_same_site default=Lax, got %s", cfg.Security.Session.CookieSameSite)
	}
	if cfg.Security.CSRF.CookieName != "csrf_token" {
		t.Fatalf("expected security.csrf.cookie_name default=csrf_token, got %s", cfg.Security.CSRF.CookieName)
	}
	if cfg.Security.CSRF.HeaderName != "X-CSRF-Token" {
		t.Fatalf("expected security.csrf.header_name default=X-CSRF-Token, got %s", cfg.Security.CSRF.HeaderName)
	}
}

func TestValidateProxyDefaults(t *testing.T) {
	cfg := baseConfig()
	cfg.Proxy = ProxyConfig{Enabled: true, TrustForwardHeaders: true}

	if err := validate(cfg); err != nil {
		t.Fatalf("validate() error = %v", err)
	}

	if cfg.Proxy.TrustedHops != 1 {
		t.Fatalf("expected proxy.trusted_hops default=1, got %d", cfg.Proxy.TrustedHops)
	}
	if len(cfg.Proxy.TrustedCIDRs) != 2 {
		t.Fatalf("expected default trusted CIDRs, got %+v", cfg.Proxy.TrustedCIDRs)
	}
}

func TestValidateProxyTrustedHopsBounds(t *testing.T) {
	tests := []int{0, -1, 6}
	for _, hops := range tests {
		cfg := baseConfig()
		cfg.Proxy = ProxyConfig{Enabled: true, TrustedHops: hops, TrustedCIDRs: []string{"127.0.0.1/32"}}

		err := validate(cfg)
		if hops == 0 {
			if err != nil {
				t.Fatalf("trusted_hops=0 should default to 1, got error: %v", err)
			}
			continue
		}

		if err == nil {
			t.Fatalf("expected error for trusted_hops=%d", hops)
		}
	}
}

func TestValidateProxyCIDR(t *testing.T) {
	cfg := baseConfig()
	cfg.Proxy = ProxyConfig{Enabled: true, TrustForwardHeaders: true, TrustedHops: 1, TrustedCIDRs: []string{"127.0.0.1/32", "not-a-cidr"}}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for invalid proxy.trusted_cidrs entry")
	}
}

func TestValidateProxyCIDRSkippedWhenProxyDisabled(t *testing.T) {
	cfg := baseConfig()
	cfg.Proxy = ProxyConfig{Enabled: false, TrustForwardHeaders: true, TrustedHops: 1, TrustedCIDRs: []string{"not-a-cidr"}}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected validate() to skip proxy.trusted_cidrs when proxy.enabled=false, got error: %v", err)
	}
}

func TestValidateProxyTrustedHopsSkippedWhenProxyDisabled(t *testing.T) {
	cfg := baseConfig()
	cfg.Proxy = ProxyConfig{Enabled: false, TrustedHops: 99, TrustedCIDRs: []string{"127.0.0.1/32"}}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected validate() to skip proxy.trusted_hops when proxy.enabled=false, got error: %v", err)
	}
}
