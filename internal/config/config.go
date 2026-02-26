// Package config предоставляет функциональность для загрузки и управления
// конфигурацией приложения из YAML файлов и переменных окружения.
package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper" // Библиотека для работы с конфигурацией
)

// Config - главная структура конфигурации приложения.
// Содержит все настройки, разбитые по категориям (server, database, security и т.д.)
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`     // Настройки HTTP сервера
	Proxy     ProxyConfig     `mapstructure:"proxy"`      // Настройки reverse-proxy режима
	Databases DatabasesConfig `mapstructure:"databases"`  // Унифицированные настройки подключений к БД
	Security  SecurityConfig  `mapstructure:"security"`   // Настройки безопасности
	RateLimit RateLimitConfig `mapstructure:"rate_limit"` // Глобальные настройки rate limiting
	Logging   LoggingConfig   `mapstructure:"logging"`    // Настройки логирования
}

// ProxyConfig - настройки работы web-ui за reverse proxy (nginx).
type ProxyConfig struct {
	Enabled             bool     `mapstructure:"enabled"`               // Включить proxy-aware режим
	TrustForwardHeaders bool     `mapstructure:"trust_forward_headers"` // Доверять X-Forwarded-* только trusted proxy
	TrustedHops         int      `mapstructure:"trusted_hops"`          // Количество доверенных proxy hops
	TrustedCIDRs        []string `mapstructure:"trusted_cidrs"`         // CIDR список trusted proxy источников
	StaticViaNginx      bool     `mapstructure:"static_via_nginx"`      // Отключить backend static routes (статику отдаёт nginx)
}

// ServerConfig - настройки HTTP сервера (Gin framework)
type ServerConfig struct {
	Port     int             `mapstructure:"port"`     // Порт, на котором будет работать сервер (например, 8443)
	TLS      TLSServerConfig `mapstructure:"tls"`      // Настройки TLS для входящих HTTP соединений
	Timeouts TimeoutConfig   `mapstructure:"timeouts"` // Таймауты HTTP сервера
	Limits   LimitsConfig    `mapstructure:"limits"`   // Лимиты HTTP сервера
	HTTP2    *HTTP2Config    `mapstructure:"http2"`    // Опциональные настройки HTTP/2
}

// TimeoutConfig - настройки таймаутов HTTP сервера.
type TimeoutConfig struct {
	Read          time.Duration `mapstructure:"read"`           // Максимальное время чтения запроса
	Write         time.Duration `mapstructure:"write"`          // Максимальное время записи ответа
	Idle          time.Duration `mapstructure:"idle"`           // Keep-alive timeout для idle соединений
	ReadHeader    time.Duration `mapstructure:"read_header"`    // Максимальное время чтения заголовков
	ShutdownGrace time.Duration `mapstructure:"shutdown_grace"` // Таймаут graceful shutdown
}

// LimitsConfig - настройки лимитов HTTP сервера.
type LimitsConfig struct {
	MaxHeaderBytes int `mapstructure:"max_header_bytes"` // Максимальный размер заголовков HTTP запроса в байтах
}

// TLSServerConfig - настройки TLS для HTTP сервера.
type TLSServerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`   // Включить HTTPS сервер
	CertPath string `mapstructure:"cert_path"` // Путь к серверному сертификату
	KeyPath  string `mapstructure:"key_path"`  // Путь к приватному ключу сервера
	CAPath   string `mapstructure:"ca_path"`   // Путь к CA сертификату (зарезервировано под mTLS)
}

// DatabasesConfig - унифицированные подключения по функциональным зонам.
type DatabasesConfig struct {
	System DatabaseTargetConfig `mapstructure:"system"`
	Audit  DatabaseTargetConfig `mapstructure:"audit"`
	Quotes DatabaseTargetConfig `mapstructure:"quotes"`
}

// DatabaseTargetConfig - универсальная цель подключения.
type DatabaseTargetConfig struct {
	Engine     string           `mapstructure:"engine"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	Oracle     OracleConfig     `mapstructure:"oracle"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
}

// MySQLConfig - детальные настройки для подключения к MySQL
type MySQLConfig struct {
	Host      string             `mapstructure:"host"`       // Адрес сервера БД (например, "localhost")
	Port      int                `mapstructure:"port"`       // Порт MySQL (обычно 3306)
	User      string             `mapstructure:"user"`       // Имя пользователя БД
	Password  string             `mapstructure:"password"`   // Пароль пользователя БД
	Database  string             `mapstructure:"database"`   // Имя базы данных
	Charset   string             `mapstructure:"charset"`    // Кодировка (обычно "utf8mb4")
	ParseTime bool               `mapstructure:"parse_time"` // Автоматически парсить время из БД в Go time.Time
	Pool      DatabasePoolConfig `mapstructure:"pool"`       // Настройки пула соединений
	TLS       DatabaseTLSConfig  `mapstructure:"tls"`        // Настройки TLS для исходящих соединений к БД
	Retry     RetryConfig        `mapstructure:"retry"`      // Политика повторных попыток (зарезервировано)
}

type DatabasePoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type DatabaseTLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertPath string `mapstructure:"cert_path"`
	KeyPath  string `mapstructure:"key_path"`
	CAPath   string `mapstructure:"ca_path"`
}

type RetryConfig struct {
	MaxAttempts  int           `mapstructure:"max_attempts"`
	InitialDelay time.Duration `mapstructure:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay"`
	Multiplier   float64       `mapstructure:"multiplier"`
}

// OracleConfig - настройки для Oracle (заготовка на будущее)
type OracleConfig struct {
	Host     string `mapstructure:"host"`     // Адрес сервера Oracle
	Port     int    `mapstructure:"port"`     // Порт Oracle (обычно 1521)
	User     string `mapstructure:"user"`     // Имя пользователя
	Password string `mapstructure:"password"` // Пароль
	Database string `mapstructure:"database"` // Имя базы данных (SID)
}

// ClickHouseConfig - задел для унифицированной схемы подключений.
type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// SecurityConfig - настройки безопасности приложения
type SecurityConfig struct {
	CSRF           CSRFConfig    `mapstructure:"csrf"`             // Настройки CSRF middleware
	Session        SessionConfig `mapstructure:"session"`          // Настройки сессий и session cookie
	BcryptCost     int           `mapstructure:"bcrypt_cost"`      // Сложность хеширования паролей (10 = хороший баланс скорости/безопасности)
	RateLimitLogin int           `mapstructure:"rate_limit_login"` // DEPRECATED: используйте rate_limit.login.requests_per_minute
	RateLimitAPI   int           `mapstructure:"rate_limit_api"`   // DEPRECATED: используйте rate_limit.api.requests_per_second
}

// CSRFConfig - настройки включения и политики CSRF middleware.
type CSRFConfig struct {
	Secret         string   `mapstructure:"secret"`
	Enabled        bool     `mapstructure:"enabled"`
	CookieName     string   `mapstructure:"cookie_name"`
	HeaderName     string   `mapstructure:"header_name"`
	TrustedOrigins []string `mapstructure:"trusted_origins"`
}

// SessionConfig - настройки сессий и cookie.
type SessionConfig struct {
	Secret         string `mapstructure:"secret"`
	CookieName     string `mapstructure:"cookie_name"`
	CookieSecure   bool   `mapstructure:"cookie_secure"`
	CookieHTTPOnly bool   `mapstructure:"cookie_http_only"`
	CookieSameSite string `mapstructure:"cookie_same_site"`
	MaxAge         int    `mapstructure:"max_age"`
	RememberMeDays int    `mapstructure:"remember_me_days"`
}

// RateLimitConfig - настройки лимитов запросов приложения.
type RateLimitConfig struct {
	Login LoginRateLimitConfig `mapstructure:"login"`
	API   APIRateLimitConfig   `mapstructure:"api"`
}

// LoginRateLimitConfig - лимиты для endpoint входа.
type LoginRateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

// APIRateLimitConfig - лимиты для API endpoint'ов.
type APIRateLimitConfig struct {
	RequestsPerSecond int `mapstructure:"requests_per_second"`
	Burst             int `mapstructure:"burst"`
}

// LoggingConfig - настройки системы логирования
type LoggingConfig struct {
	Level              string `mapstructure:"level"`                 // Уровень логирования: "debug", "info", "warn", "error"
	Format             string `mapstructure:"format"`                // Формат логов: "json" (структурированный) или "text" (читаемый)
	Output             string `mapstructure:"output"`                // Куда писать логи: "stdout" (консоль), "file" (файл) или "both" (оба)
	ErrorPath          string `mapstructure:"error_path"`            // Путь к файлу логов (если output = "file" или "both")
	MaxSizeMB          int    `mapstructure:"max_size_mb"`           // Максимальный размер файла лога в МБ (100 МБ)
	MaxBackups         int    `mapstructure:"max_backups"`           // Количество архивных файлов логов (5 файлов)
	MaxAgeDays         int    `mapstructure:"max_age_days"`          // Хранить логи N дней (30 дней)
	Compress           bool   `mapstructure:"compress"`              // Сжимать старые логи (gzip)
	AccessPath         string `mapstructure:"access_path"`           // Путь к access log (если пусто, создается рядом с File)
	AccessToStdout     *bool  `mapstructure:"access_to_stdout"`      // Дублировать access лог в stdout
	OutRequestToStdout *bool  `mapstructure:"out_request_to_stdout"` // Зарезервировано для отдельного out_request потока
	AccessMaxSizeMB    int    `mapstructure:"access_max_size_mb"`    // Максимальный размер access log в МБ
	AccessMaxBackups   int    `mapstructure:"access_max_backups"`    // Количество архивных access log файлов
	AccessMaxAgeDays   int    `mapstructure:"access_max_age_days"`   // Хранить access log N дней
	AuditPath          string `mapstructure:"audit_path"`            // Путь к audit log (если пусто, создается рядом с File)
	AuditToStdout      *bool  `mapstructure:"audit_to_stdout"`       // Дублировать audit лог в stdout
	AuditMaxSizeMB     int    `mapstructure:"audit_max_size_mb"`     // Максимальный размер audit log в МБ
	AuditMaxBackups    int    `mapstructure:"audit_max_backups"`     // Количество архивных audit log файлов
	AuditMaxAgeDays    int    `mapstructure:"audit_max_age_days"`    // Хранить audit log N дней
}

var (
	// globalConfig - глобальная переменная для хранения загруженной конфигурации.
	// После вызова Load() конфигурация доступна через Get() из любого места программы.
	globalConfig *Config
)

// Load загружает конфигурацию из YAML файла и переменных окружения.
//
// Параметры:
//   - configPath: путь к директории с config.yaml. Если пустая строка "", то ищет в стандартных местах:
//   - ./config (текущая директория)
//   - ../config (родительская директория)
//   - ../../config (на уровень выше)
//
// Возвращает:
//   - *Config: указатель на загруженную конфигурацию
//   - error: ошибка, если не удалось загрузить или валидировать конфигурацию
//
// Пример использования:
//
//	cfg, err := config.Load("")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load(configPath string) (*Config, error) {
	// Указываем Viper, что конфигурация в формате YAML
	viper.SetConfigType("yaml")
	// Имя файла конфигурации (без расширения) - будет искать config.yaml
	viper.SetConfigName("config")

	// Устанавливаем путь поиска конфигурации
	if configPath != "" {
		// Если путь указан явно, используем его
		viper.AddConfigPath(configPath)
	} else {
		// Иначе ищем в стандартных местах (по порядку приоритета)
		viper.AddConfigPath("./config")     // Текущая директория
		viper.AddConfigPath("../config")    // Родительская директория
		viper.AddConfigPath("../../config") // На уровень выше
	}

	// Настройка переменных окружения:
	// Префикс "CT_" означает, что переменные должны начинаться с CT_
	// Например: CT_SERVER_PORT=8080 переопределит server.port из YAML
	viper.SetEnvPrefix("CT")
	// Автоматически читать переменные окружения и маппить их в конфигурацию
	// CT_SERVER_PORT -> server.port
	// CT_DATABASES_SYSTEM_MYSQL_HOST -> databases.system.mysql.host
	viper.AutomaticEnv()

	// Читаем конфигурационный файл (config.yaml)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Преобразуем данные из Viper в нашу структуру Config
	// Unmarshal автоматически заполнит все поля структур из YAML
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Проверяем, что все обязательные поля заполнены и значения корректны
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Сохраняем конфигурацию в глобальную переменную для доступа через Get()
	globalConfig = &cfg
	return &cfg, nil
}

// Get возвращает глобальный экземпляр конфигурации.
//
// ВАЖНО: перед вызовом Get() необходимо вызвать Load(), иначе будет паника.
//
// Возвращает:
//   - *Config: указатель на загруженную конфигурацию
//
// Пример использования:
//
//	cfg := config.Get()
//	port := cfg.Server.Port
//	dsn := cfg.GetMySQLDSN()
func Get() *Config {
	// Проверяем, что конфигурация была загружена
	if globalConfig == nil {
		panic("config not loaded, call Load() first")
	}
	return globalConfig
}

// validate проверяет корректность загруженной конфигурации.
// Вызывается автоматически в Load() после чтения файла.
//
// Проверяет:
//   - Порт сервера в допустимом диапазоне (1-65535)
//   - Указан тип БД (engine)
//   - Для MySQL заполнены обязательные поля (host, database, user)
//   - Секретные ключи изменены с дефолтных значений
//
// Возвращает:
//   - error: ошибка валидации, если что-то не так
func validate(cfg *Config) error {
	// Проверка порта сервера
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Server.Timeouts.Read == 0 {
		cfg.Server.Timeouts.Read = 60 * time.Second
	}
	if cfg.Server.Timeouts.Write == 0 {
		cfg.Server.Timeouts.Write = 60 * time.Second
	}
	if cfg.Server.Timeouts.Idle == 0 {
		cfg.Server.Timeouts.Idle = 120 * time.Second
	}
	if cfg.Server.Timeouts.ReadHeader == 0 {
		cfg.Server.Timeouts.ReadHeader = 5 * time.Second
	}
	if cfg.Server.Timeouts.ShutdownGrace == 0 {
		cfg.Server.Timeouts.ShutdownGrace = 10 * time.Second
	}
	if cfg.Server.Limits.MaxHeaderBytes == 0 {
		cfg.Server.Limits.MaxHeaderBytes = 1 << 20 // 1 MiB
	}

	if cfg.Server.Timeouts.Read < 0 || cfg.Server.Timeouts.Read > 24*time.Hour {
		return fmt.Errorf("invalid server.timeouts.read: %s", cfg.Server.Timeouts.Read)
	}
	if cfg.Server.Timeouts.Write < 0 || cfg.Server.Timeouts.Write > 24*time.Hour {
		return fmt.Errorf("invalid server.timeouts.write: %s", cfg.Server.Timeouts.Write)
	}
	if cfg.Server.Timeouts.Idle < 0 || cfg.Server.Timeouts.Idle > 24*time.Hour {
		return fmt.Errorf("invalid server.timeouts.idle: %s", cfg.Server.Timeouts.Idle)
	}
	if cfg.Server.Timeouts.ReadHeader < 0 || cfg.Server.Timeouts.ReadHeader > 24*time.Hour {
		return fmt.Errorf("invalid server.timeouts.read_header: %s", cfg.Server.Timeouts.ReadHeader)
	}
	if cfg.Server.Timeouts.ShutdownGrace < 0 || cfg.Server.Timeouts.ShutdownGrace > 24*time.Hour {
		return fmt.Errorf("invalid server.timeouts.shutdown_grace: %s", cfg.Server.Timeouts.ShutdownGrace)
	}
	if cfg.Server.Limits.MaxHeaderBytes < 4096 || cfg.Server.Limits.MaxHeaderBytes > (16<<20) {
		return fmt.Errorf("invalid server.limits.max_header_bytes: %d", cfg.Server.Limits.MaxHeaderBytes)
	}

	if cfg.Server.HTTP2 != nil {
		if _, err := cfg.Server.HTTP2.Parse(); err != nil {
			return fmt.Errorf("invalid server.http2: %w", err)
		}
	}

	if cfg.Proxy.Enabled {
		if cfg.Proxy.TrustedHops == 0 {
			cfg.Proxy.TrustedHops = 1
		}
		if cfg.Proxy.TrustedHops < 1 || cfg.Proxy.TrustedHops > 5 {
			return fmt.Errorf("proxy.trusted_hops must be between 1 and 5")
		}
	}
	if cfg.Proxy.Enabled && cfg.Proxy.TrustForwardHeaders {
		if cfg.Proxy.TrustedCIDRs == nil {
			cfg.Proxy.TrustedCIDRs = []string{"127.0.0.1/32", "10.0.0.0/8"}
		}
		for _, cidr := range cfg.Proxy.TrustedCIDRs {
			if strings.TrimSpace(cidr) == "" {
				return fmt.Errorf("proxy.trusted_cidrs must not contain empty values")
			}
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("invalid proxy.trusted_cidrs entry %q: %w", cidr, err)
			}
		}
	}

	if cfg.Server.TLS.Enabled {
		if cfg.Server.TLS.CertPath == "" {
			return fmt.Errorf("server.tls.cert_path is required when server.tls.enabled=true")
		}
		if cfg.Server.TLS.KeyPath == "" {
			return fmt.Errorf("server.tls.key_path is required when server.tls.enabled=true")
		}
	}

	// Проверка, что указан тип system-базы
	if cfg.Databases.System.Engine == "" {
		return fmt.Errorf("databases.system.engine is required")
	}

	// Если используется MySQL для system, проверяем обязательные поля
	if cfg.Databases.System.Engine == "mysql" {
		if cfg.Databases.System.MySQL.Host == "" {
			return fmt.Errorf("mysql host is required")
		}
		if cfg.Databases.System.MySQL.Database == "" {
			return fmt.Errorf("mysql database name is required")
		}
		if cfg.Databases.System.MySQL.User == "" {
			return fmt.Errorf("mysql user is required")
		}
		if cfg.Databases.System.MySQL.Pool.MaxOpenConns == 0 {
			cfg.Databases.System.MySQL.Pool.MaxOpenConns = 25
		}
		if cfg.Databases.System.MySQL.Pool.MaxIdleConns == 0 {
			cfg.Databases.System.MySQL.Pool.MaxIdleConns = 5
		}
		if cfg.Databases.System.MySQL.Pool.ConnMaxLifetime == 0 {
			cfg.Databases.System.MySQL.Pool.ConnMaxLifetime = 300 * time.Second
		}
		if cfg.Databases.System.MySQL.Pool.ConnMaxIdleTime == 0 {
			cfg.Databases.System.MySQL.Pool.ConnMaxIdleTime = 60 * time.Second
		}

		if cfg.Databases.System.MySQL.Pool.MaxOpenConns < 1 {
			return fmt.Errorf("databases.system.mysql.pool.max_open_conns must be > 0")
		}
		if cfg.Databases.System.MySQL.Pool.MaxIdleConns < 0 {
			return fmt.Errorf("databases.system.mysql.pool.max_idle_conns must be >= 0")
		}
		if cfg.Databases.System.MySQL.Pool.MaxIdleConns > cfg.Databases.System.MySQL.Pool.MaxOpenConns {
			return fmt.Errorf("databases.system.mysql.pool.max_idle_conns cannot be greater than max_open_conns")
		}

		if cfg.Databases.System.MySQL.Retry.MaxAttempts == 0 {
			cfg.Databases.System.MySQL.Retry.MaxAttempts = 3
		}
		if cfg.Databases.System.MySQL.Retry.InitialDelay == 0 {
			cfg.Databases.System.MySQL.Retry.InitialDelay = 1 * time.Second
		}
		if cfg.Databases.System.MySQL.Retry.MaxDelay == 0 {
			cfg.Databases.System.MySQL.Retry.MaxDelay = 5 * time.Second
		}
		if cfg.Databases.System.MySQL.Retry.Multiplier == 0 {
			cfg.Databases.System.MySQL.Retry.Multiplier = 2.0
		}

		if cfg.Databases.System.MySQL.Retry.MaxAttempts < 1 {
			return fmt.Errorf("databases.system.mysql.retry.max_attempts must be > 0")
		}
		if cfg.Databases.System.MySQL.Retry.InitialDelay < 0 {
			return fmt.Errorf("databases.system.mysql.retry.initial_delay must be >= 0")
		}
		if cfg.Databases.System.MySQL.Retry.MaxDelay < 0 {
			return fmt.Errorf("databases.system.mysql.retry.max_delay must be >= 0")
		}
		if cfg.Databases.System.MySQL.Retry.Multiplier < 1.0 {
			return fmt.Errorf("databases.system.mysql.retry.multiplier must be >= 1.0")
		}

		if cfg.Databases.System.MySQL.TLS.Enabled {
			if cfg.Databases.System.MySQL.TLS.CAPath == "" {
				return fmt.Errorf("databases.system.mysql.tls.ca_path is required when tls.enabled=true")
			}
			if cfg.Databases.System.MySQL.TLS.CertPath == "" {
				return fmt.Errorf("databases.system.mysql.tls.cert_path is required when tls.enabled=true")
			}
			if cfg.Databases.System.MySQL.TLS.KeyPath == "" {
				return fmt.Errorf("databases.system.mysql.tls.key_path is required when tls.enabled=true")
			}
		}
	}

	// Проверка, что секретные ключи изменены с дефолтных значений
	// Это важно для безопасности - нельзя использовать дефолтные ключи в продакшн
	if cfg.Security.Session.Secret == "" || cfg.Security.Session.Secret == "change-this-session-secret-in-production" {
		return fmt.Errorf("security.session.secret must be set and changed from default")
	}
	if cfg.Security.CSRF.Secret == "" || cfg.Security.CSRF.Secret == "change-this-csrf-secret-in-production" {
		return fmt.Errorf("security.csrf.secret must be set and changed from default")
	}
	if cfg.Security.BcryptCost == 0 {
		cfg.Security.BcryptCost = 10
	}
	if cfg.Security.BcryptCost < 10 || cfg.Security.BcryptCost > 14 {
		return fmt.Errorf("security.bcrypt_cost must be between 10 and 14")
	}

	if cfg.Security.Session.CookieName == "" {
		cfg.Security.Session.CookieName = "ct_session"
	}
	if cfg.Security.Session.MaxAge == 0 {
		cfg.Security.Session.MaxAge = 86400
	}
	if cfg.Security.Session.RememberMeDays == 0 {
		cfg.Security.Session.RememberMeDays = 7
	}
	if cfg.Security.Session.CookieSameSite == "" {
		cfg.Security.Session.CookieSameSite = "Lax"
	}
	sameSite := strings.ToLower(strings.TrimSpace(cfg.Security.Session.CookieSameSite))
	switch sameSite {
	case "strict", "lax", "none":
	default:
		return fmt.Errorf("security.session.cookie_same_site must be one of: Strict, Lax, None")
	}
	if sameSite == "none" && !(cfg.Security.Session.CookieSecure || cfg.Server.TLS.Enabled) {
		return fmt.Errorf("security.session.cookie_same_site=None requires security.session.cookie_secure=true or server.tls.enabled=true")
	}

	if cfg.Security.CSRF.CookieName == "" {
		cfg.Security.CSRF.CookieName = "csrf_token"
	}
	if cfg.Security.CSRF.HeaderName == "" {
		cfg.Security.CSRF.HeaderName = "X-CSRF-Token"
	}
	for _, origin := range cfg.Security.CSRF.TrustedOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("security.csrf.trusted_origins must not contain empty values")
		}
	}

	if cfg.Databases.System.Engine != "mysql" {
		return fmt.Errorf("unsupported databases.system.engine: %s", cfg.Databases.System.Engine)
	}

	if cfg.RateLimit.Login.RequestsPerMinute > 0 && cfg.Security.RateLimitLogin > 0 && cfg.RateLimit.Login.RequestsPerMinute != cfg.Security.RateLimitLogin {
		return fmt.Errorf("both rate_limit.login.requests_per_minute and security.rate_limit_login are set with different values; use rate_limit.login.requests_per_minute")
	}
	if cfg.RateLimit.Login.RequestsPerMinute == 0 {
		if cfg.Security.RateLimitLogin > 0 {
			cfg.RateLimit.Login.RequestsPerMinute = cfg.Security.RateLimitLogin
		} else {
			cfg.RateLimit.Login.RequestsPerMinute = 5
		}
	}
	if cfg.RateLimit.Login.Burst == 0 {
		cfg.RateLimit.Login.Burst = cfg.RateLimit.Login.RequestsPerMinute
	}
	if cfg.RateLimit.Login.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate_limit.login.requests_per_minute must be > 0")
	}
	if cfg.RateLimit.Login.Burst < 0 {
		return fmt.Errorf("rate_limit.login.burst must be >= 0")
	}

	if cfg.RateLimit.API.RequestsPerSecond > 0 && cfg.Security.RateLimitAPI > 0 && cfg.RateLimit.API.RequestsPerSecond != cfg.Security.RateLimitAPI {
		return fmt.Errorf("both rate_limit.api.requests_per_second and security.rate_limit_api are set with different values; use rate_limit.api.requests_per_second")
	}
	if cfg.RateLimit.API.RequestsPerSecond == 0 {
		if cfg.Security.RateLimitAPI > 0 {
			cfg.RateLimit.API.RequestsPerSecond = cfg.Security.RateLimitAPI
		} else {
			cfg.RateLimit.API.RequestsPerSecond = 100
		}
	}
	if cfg.RateLimit.API.Burst == 0 {
		cfg.RateLimit.API.Burst = cfg.RateLimit.API.RequestsPerSecond
	}
	if cfg.RateLimit.API.RequestsPerSecond <= 0 {
		return fmt.Errorf("rate_limit.api.requests_per_second must be > 0")
	}
	if cfg.RateLimit.API.Burst < 0 {
		return fmt.Errorf("rate_limit.api.burst must be >= 0")
	}

	return nil
}

// GetMySQLDSN формирует строку подключения (Data Source Name) для MySQL.
// DSN используется в sql.Open() для подключения к базе данных.
//
// Формат DSN: user:password@tcp(host:port)/database?параметры
//
// Пример результата:
//
//	"root:password@tcp(localhost:3306)/ct_system?charset=utf8mb4&parseTime=true"
//
// Возвращает:
//   - string: строка подключения к MySQL
func (c *Config) GetMySQLDSN() string {
	cfg := c.Databases.System.MySQL

	// Базовая часть DSN: user:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// Добавляем параметры подключения
	if cfg.Charset != "" {
		// Если указана кодировка, добавляем её как первый параметр
		dsn += "?charset=" + cfg.Charset
	}
	if cfg.ParseTime {
		// Если нужно парсить время, добавляем parseTime=true
		if cfg.Charset != "" {
			// Если уже есть параметры (charset), добавляем через &
			dsn += "&parseTime=true"
		} else {
			// Если параметров ещё нет, добавляем через ?
			dsn += "?parseTime=true"
		}
	}
	if cfg.TLS.Enabled {
		if strings.Contains(dsn, "?") {
			dsn += "&tls=true"
		} else {
			dsn += "?tls=true"
		}
	}

	return dsn
}

// IsDebug проверяет, включен ли режим отладки по уровню логирования.
// При logging.level=debug включается debug-поведение приложения.
//
// Возвращает:
//   - bool: true если logging.level == "debug", иначе false
func (c *Config) IsDebug() bool {
	return strings.EqualFold(c.Logging.Level, "debug")
}

// ConfigPath ищет директорию с конфигурацией в стандартных местах.
// Используется для определения, где находится config.yaml.
//
// Ищет в следующем порядке:
//  1. ./config (текущая директория)
//  2. ../config (родительская директория)
//  3. ../../config (на уровень выше)
//
// Возвращает:
//   - string: абсолютный путь к директории с конфигурацией
func ConfigPath() string {
	// Список путей для поиска (в порядке приоритета)
	paths := []string{
		"./config",
		"../config",
		"../../config",
	}

	// Проверяем каждый путь, существует ли директория
	for _, path := range paths {
		// os.Stat проверяет существование файла/директории
		if _, err := os.Stat(path); err == nil {
			// Если директория найдена, возвращаем абсолютный путь
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Если ничего не найдено, возвращаем путь по умолчанию
	return "./config"
}
