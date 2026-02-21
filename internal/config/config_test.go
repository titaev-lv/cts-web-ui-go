package config

import (
	"testing"
	"time"
)

func TestValidateTLSModes(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid with tls disabled and empty cert paths",
			cfg: &Config{
				Server: ServerConfig{
					Port: 8443,
					TLS:  TLSServerConfig{Enabled: false},
					Timeouts: TimeoutConfig{
						Read:          60 * time.Second,
						Write:         60 * time.Second,
						Idle:          120 * time.Second,
						ReadHeader:    5 * time.Second,
						ShutdownGrace: 10 * time.Second,
					},
				},
				Database: DatabaseConfig{
					Engine: "mysql",
					MySQL: MySQLConfig{
						Host:     "mysql",
						Database: "ct_system",
						User:     "root",
					},
				},
				Security: SecurityConfig{
					JWTSecret:     "test-jwt-secret",
					SessionSecret: "test-session-secret",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with tls enabled and cert paths",
			cfg: &Config{
				Server: ServerConfig{
					Port: 8443,
					TLS: TLSServerConfig{
						Enabled:  true,
						CertPath: "pki/server/web-ui.crt",
						KeyPath:  "pki/server/web-ui.key",
					},
					Timeouts: TimeoutConfig{
						Read:          60 * time.Second,
						Write:         60 * time.Second,
						Idle:          120 * time.Second,
						ReadHeader:    5 * time.Second,
						ShutdownGrace: 10 * time.Second,
					},
				},
				Database: DatabaseConfig{
					Engine: "mysql",
					MySQL: MySQLConfig{
						Host:     "mysql",
						Database: "ct_system",
						User:     "root",
					},
				},
				Security: SecurityConfig{
					JWTSecret:     "test-jwt-secret",
					SessionSecret: "test-session-secret",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid tls enabled without cert path",
			cfg: &Config{
				Server: ServerConfig{
					Port: 8443,
					TLS: TLSServerConfig{
						Enabled: true,
						KeyPath: "pki/server/web-ui.key",
					},
					Timeouts: TimeoutConfig{
						Read:          60 * time.Second,
						Write:         60 * time.Second,
						Idle:          120 * time.Second,
						ReadHeader:    5 * time.Second,
						ShutdownGrace: 10 * time.Second,
					},
				},
				Database: DatabaseConfig{
					Engine: "mysql",
					MySQL: MySQLConfig{
						Host:     "mysql",
						Database: "ct_system",
						User:     "root",
					},
				},
				Security: SecurityConfig{
					JWTSecret:     "test-jwt-secret",
					SessionSecret: "test-session-secret",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tls enabled without key path",
			cfg: &Config{
				Server: ServerConfig{
					Port: 8443,
					TLS: TLSServerConfig{
						Enabled:  true,
						CertPath: "pki/server/web-ui.crt",
					},
					Timeouts: TimeoutConfig{
						Read:          60 * time.Second,
						Write:         60 * time.Second,
						Idle:          120 * time.Second,
						ReadHeader:    5 * time.Second,
						ShutdownGrace: 10 * time.Second,
					},
				},
				Database: DatabaseConfig{
					Engine: "mysql",
					MySQL: MySQLConfig{
						Host:     "mysql",
						Database: "ct_system",
						User:     "root",
					},
				},
				Security: SecurityConfig{
					JWTSecret:     "test-jwt-secret",
					SessionSecret: "test-session-secret",
				},
			},
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
	cfg := &Config{
		Server: ServerConfig{
			Port: 8443,
			TLS:  TLSServerConfig{Enabled: false},
		},
		Database: DatabaseConfig{
			Engine: "mysql",
			MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
		},
		Security: SecurityConfig{JWTSecret: "test-jwt-secret", SessionSecret: "test-session-secret"},
		RateLimit: RateLimitConfig{
			Login: LoginRateLimitConfig{RequestsPerMinute: 5, Burst: 5},
			API:   APIRateLimitConfig{RequestsPerSecond: 100, Burst: 100},
		},
	}

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
	cfg := &Config{
		Server: ServerConfig{
			Port: 8443,
			TLS:  TLSServerConfig{Enabled: false},
			Timeouts: TimeoutConfig{
				Read: -1 * time.Second,
			},
		},
		Database: DatabaseConfig{
			Engine: "mysql",
			MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
		},
		Security: SecurityConfig{JWTSecret: "test-jwt-secret", SessionSecret: "test-session-secret"},
	}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for negative server.timeouts.read")
	}
}

func TestValidateMaxHeaderBytesBounds(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8443,
			TLS:  TLSServerConfig{Enabled: false},
			Limits: LimitsConfig{
				MaxHeaderBytes: 1024,
			},
		},
		Database: DatabaseConfig{
			Engine: "mysql",
			MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
		},
		Security: SecurityConfig{JWTSecret: "test-jwt-secret", SessionSecret: "test-session-secret"},
	}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for too small server.limits.max_header_bytes")
	}
}

func TestValidateHTTP2Invalid(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8443,
			TLS:  TLSServerConfig{Enabled: false},
			HTTP2: &HTTP2Config{
				MaxFrameSize: "invalid",
			},
		},
		Database: DatabaseConfig{
			Engine: "mysql",
			MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
		},
		Security: SecurityConfig{JWTSecret: "test-jwt-secret", SessionSecret: "test-session-secret"},
	}

	if err := validate(cfg); err == nil {
		t.Fatal("expected validate() to fail for invalid server.http2")
	}
}

func TestValidateRateLimitFallbackFromSecurity(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8443,
			TLS:  TLSServerConfig{Enabled: false},
		},
		Database: DatabaseConfig{
			Engine: "mysql",
			MySQL:  MySQLConfig{Host: "mysql", Database: "ct_system", User: "root"},
		},
		Security: SecurityConfig{
			JWTSecret:      "test-jwt-secret",
			SessionSecret:  "test-session-secret",
			RateLimitLogin: 7,
			RateLimitAPI:   55,
		},
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
