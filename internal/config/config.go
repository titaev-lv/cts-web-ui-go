// Package config предоставляет функциональность для загрузки и управления
// конфигурацией приложения из YAML файлов и переменных окружения.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper" // Библиотека для работы с конфигурацией
)

// Config - главная структура конфигурации приложения.
// Содержит все настройки, разбитые по категориям (server, database, security и т.д.)
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`   // Настройки HTTP сервера
	Database DatabaseConfig `mapstructure:"database"` // Настройки базы данных
	Security SecurityConfig `mapstructure:"security"` // Настройки безопасности
	Logging  LoggingConfig  `mapstructure:"logging"`  // Настройки логирования
	Daemon   DaemonConfig   `mapstructure:"daemon"`   // Настройки демона
	App      AppConfig      `mapstructure:"app"`      // Общие настройки приложения
}

// ServerConfig - настройки HTTP сервера (Gin framework)
type ServerConfig struct {
	Port         int           `mapstructure:"port"`          // Порт, на котором будет работать сервер (например, 8443)
	Mode         string        `mapstructure:"mode"`          // Режим работы: "debug" (разработка) или "release" (продакшн)
	Host         string        `mapstructure:"host"`          // IP адрес для прослушивания ("0.0.0.0" = все интерфейсы)
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`  // Максимальное время чтения запроса (например, 60s)
	WriteTimeout time.Duration `mapstructure:"write_timeout"` // Максимальное время записи ответа (например, 60s)
}

// DatabaseConfig - настройки подключения к базе данных
type DatabaseConfig struct {
	Engine string       `mapstructure:"engine"` // Тип БД: "mysql" или "oracle" (пока только mysql)
	MySQL  MySQLConfig  `mapstructure:"mysql"`  // Настройки для MySQL
	Oracle OracleConfig `mapstructure:"oracle"` // Настройки для Oracle (на будущее)
}

// MySQLConfig - детальные настройки для подключения к MySQL
type MySQLConfig struct {
	Host               string        `mapstructure:"host"`                     // Адрес сервера БД (например, "localhost")
	Port               int           `mapstructure:"port"`                     // Порт MySQL (обычно 3306)
	User               string        `mapstructure:"user"`                     // Имя пользователя БД
	Password           string        `mapstructure:"password"`                 // Пароль пользователя БД
	Database           string        `mapstructure:"database"`                 // Имя базы данных
	Charset            string        `mapstructure:"charset"`                  // Кодировка (обычно "utf8mb4")
	ParseTime          bool          `mapstructure:"parse_time"`               // Автоматически парсить время из БД в Go time.Time
	MaxConnections     int           `mapstructure:"max_connections"`          // Максимальное количество открытых соединений
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`     // Максимальное количество неактивных соединений в пуле
	ConnMaxLifetime    time.Duration `mapstructure:"connection_max_lifetime"`  // Максимальное время жизни соединения (300s = 5 минут)
	ConnMaxIdleTime    time.Duration `mapstructure:"connection_max_idle_time"` // Максимальное время простоя соединения (60s = 1 минута)
}

// OracleConfig - настройки для Oracle (заготовка на будущее)
type OracleConfig struct {
	Host     string `mapstructure:"host"`     // Адрес сервера Oracle
	Port     int    `mapstructure:"port"`     // Порт Oracle (обычно 1521)
	User     string `mapstructure:"user"`     // Имя пользователя
	Password string `mapstructure:"password"` // Пароль
	Database string `mapstructure:"database"` // Имя базы данных (SID)
}

// SecurityConfig - настройки безопасности приложения
type SecurityConfig struct {
	JWTSecret             string `mapstructure:"jwt_secret"`               // Секретный ключ для подписи JWT токенов (должен быть уникальным!)
	SessionSecret         string `mapstructure:"session_secret"`           // Секретный ключ для сессий (должен быть уникальным!)
	CSRFSecret            string `mapstructure:"csrf_secret"`              // Секретный ключ для CSRF защиты (должен быть уникальным!)
	BcryptCost            int    `mapstructure:"bcrypt_cost"`              // Сложность хеширования паролей (10 = хороший баланс скорости/безопасности)
	SessionCookieName     string `mapstructure:"session_cookie_name"`      // Имя cookie для сессии
	SessionCookieSecure   bool   `mapstructure:"session_cookie_secure"`    // Использовать только HTTPS для cookie (true в продакшн)
	SessionCookieHTTPOnly bool   `mapstructure:"session_cookie_http_only"` // Запретить доступ к cookie через JavaScript (защита от XSS)
	SessionCookieSameSite string `mapstructure:"session_cookie_same_site"` // Политика SameSite: "Strict", "Lax" или "None"
	SessionMaxAge         int    `mapstructure:"session_max_age"`          // Время жизни сессии в секундах (86400 = 24 часа)
	RememberMeDays        int    `mapstructure:"remember_me_days"`         // Количество дней для "Запомнить меня" (7 дней)
	RateLimitLogin        int    `mapstructure:"rate_limit_login"`         // Максимальное количество попыток входа в минуту (защита от брутфорса)
	RateLimitAPI          int    `mapstructure:"rate_limit_api"`           // Максимальное количество API запросов в секунду
}

// LoggingConfig - настройки системы логирования
type LoggingConfig struct {
	Level            string `mapstructure:"level"`              // Уровень логирования: "debug", "info", "warn", "error"
	Format           string `mapstructure:"format"`             // Формат логов: "json" (структурированный) или "text" (читаемый)
	Output           string `mapstructure:"output"`             // Куда писать логи: "stdout" (консоль), "file" (файл) или "both" (оба)
	File             string `mapstructure:"file"`               // Путь к файлу логов (если output = "file" или "both")
	MaxSize          int    `mapstructure:"max_size"`           // Максимальный размер файла лога в МБ (100 МБ)
	MaxBackups       int    `mapstructure:"max_backups"`        // Количество архивных файлов логов (5 файлов)
	MaxAge           int    `mapstructure:"max_age"`            // Хранить логи N дней (30 дней)
	Compress         bool   `mapstructure:"compress"`           // Сжимать старые логи (gzip)
	AccessFile       string `mapstructure:"access_file"`        // Путь к access log (если пусто, создается рядом с File)
	AccessMaxSize    int    `mapstructure:"access_max_size"`    // Максимальный размер access log в МБ
	AccessMaxBackups int    `mapstructure:"access_max_backups"` // Количество архивных access log файлов
	AccessMaxAge     int    `mapstructure:"access_max_age"`     // Хранить access log N дней
	AuditFile        string `mapstructure:"audit_file"`         // Путь к audit log (если пусто, создается рядом с File)
	AuditMaxSize     int    `mapstructure:"audit_max_size"`     // Максимальный размер audit log в МБ
	AuditMaxBackups  int    `mapstructure:"audit_max_backups"`  // Количество архивных audit log файлов
	AuditMaxAge      int    `mapstructure:"audit_max_age"`      // Хранить audit log N дней
}

// DaemonConfig - настройки для управления внешним демоном
type DaemonConfig struct {
	Path          string        `mapstructure:"path"`           // Путь к исполняемому файлу демона
	CheckInterval time.Duration `mapstructure:"check_interval"` // Как часто проверять статус демона (5s)
}

// AppConfig - общие настройки приложения
type AppConfig struct {
	Name            string `mapstructure:"name"`             // Название приложения
	Version         string `mapstructure:"version"`          // Версия приложения
	Timezone        string `mapstructure:"timezone"`         // Часовой пояс (например, "UTC", "Europe/Moscow")
	DefaultLanguage string `mapstructure:"default_language"` // Язык по умолчанию ("en", "ru")
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
	// CT_DATABASE_MYSQL_HOST -> database.mysql.host
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

	// Проверка, что указан тип базы данных
	if cfg.Database.Engine == "" {
		return fmt.Errorf("database engine is required")
	}

	// Если используется MySQL, проверяем обязательные поля
	if cfg.Database.Engine == "mysql" {
		if cfg.Database.MySQL.Host == "" {
			return fmt.Errorf("mysql host is required")
		}
		if cfg.Database.MySQL.Database == "" {
			return fmt.Errorf("mysql database name is required")
		}
		if cfg.Database.MySQL.User == "" {
			return fmt.Errorf("mysql user is required")
		}
	}

	// Проверка, что секретные ключи изменены с дефолтных значений
	// Это важно для безопасности - нельзя использовать дефолтные ключи в продакшн
	if cfg.Security.JWTSecret == "" || cfg.Security.JWTSecret == "change-this-secret-key-in-production" {
		return fmt.Errorf("jwt_secret must be set and changed from default")
	}

	if cfg.Security.SessionSecret == "" || cfg.Security.SessionSecret == "change-this-session-secret-in-production" {
		return fmt.Errorf("session_secret must be set and changed from default")
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
	cfg := c.Database.MySQL

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

	return dsn
}

// IsDebug проверяет, работает ли сервер в режиме отладки (debug).
// В режиме debug выводится больше информации об ошибках и запросах.
//
// Возвращает:
//   - bool: true если mode == "debug", иначе false
func (c *Config) IsDebug() bool {
	return c.Server.Mode == "debug"
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
