// Package logger предоставляет систему структурированного логирования на основе zerolog.
// Поддерживает разные уровни логирования, форматы вывода (text/json) и ротацию файлов.
package logger

import (
	"ctweb/internal/config"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack" // Ротация логов (автоматическое создание архивов)
	"github.com/rs/zerolog"            // Структурированное логирование
)

var (
	// globalLogger - глобальный экземпляр логгера
	// После вызова Init() доступен через Get()
	globalLogger zerolog.Logger
)

// Init инициализирует систему логирования на основе конфигурации.
//
// Что делает:
//   1. Определяет уровень логирования из конфигурации
//   2. Настраивает формат вывода (text или json)
//   3. Настраивает куда писать логи (stdout, file или both)
//   4. Настраивает ротацию файлов (если используется файл)
//   5. Создаёт глобальный логгер
//
// ВАЖНО: вызывать ОДИН раз при старте приложения, обычно в main.go после config.Load()
//
// Пример использования:
//
//	config.Load("")
//	logger.Init()
//	logger.Info().Msg("Application started")
func Init() {
	cfg := config.Get()
	logCfg := cfg.Logging

	// ============================================
	// ШАГ 1: Определяем уровень логирования
	// ============================================
	// Уровни логирования (от меньшего к большему):
	//   - Trace: очень детальная информация (обычно не используется)
	//   - Debug: отладочная информация (для разработки)
	//   - Info: общая информация о работе приложения
	//   - Warn: предупреждения (что-то не так, но не критично)
	//   - Error: ошибки (что-то пошло не так)
	//   - Fatal: критическая ошибка (приложение завершится)
	//   - Panic: паника (приложение завершится с паникой)
	var level zerolog.Level
	switch strings.ToLower(logCfg.Level) {
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	default:
		// Если уровень не указан или неизвестен, используем Info
		level = zerolog.InfoLevel
	}

	// ============================================
	// ШАГ 2: Настраиваем вывод логов
	// ============================================
	// Можем писать в несколько мест одновременно (multi-writer)
	var writers []io.Writer

	// Проверяем, нужно ли писать в консоль (stdout)
	if logCfg.Output == "stdout" || logCfg.Output == "both" {
		// Настраиваем формат вывода для консоли
		var consoleWriter io.Writer
		if logCfg.Format == "json" {
			// JSON формат - структурированный, удобен для парсинга
			consoleWriter = os.Stdout
		} else {
			// Text формат - читаемый человеком, с цветами
			consoleWriter = zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339, // Формат времени: 2006-01-02T15:04:05Z07:00
				NoColor:    false,        // Использовать цвета в консоли
			}
		}
		writers = append(writers, consoleWriter)
	}

	// Проверяем, нужно ли писать в файл
	if logCfg.Output == "file" || logCfg.Output == "both" {
		// Создаём директорию для логов, если её нет
		logDir := filepath.Dir(logCfg.File)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				// Если не удалось создать директорию, пишем только в stdout
				writers = append(writers, os.Stderr)
				// Логируем ошибку через временный логгер
				tempLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()
				tempLogger.Error().
					Err(err).
					Str("log_file", logCfg.File).
					Msg("Failed to create log directory, using stderr")
			} else {
				// Настраиваем ротацию файлов через lumberjack
				fileWriter := &lumberjack.Logger{
					Filename:   logCfg.File,      // Путь к файлу логов
					MaxSize:    logCfg.MaxSize,   // Максимальный размер файла в МБ
					MaxBackups: logCfg.MaxBackups, // Количество архивных файлов
					MaxAge:     logCfg.MaxAge,    // Хранить логи N дней
					Compress:   logCfg.Compress,  // Сжимать старые логи (gzip)
					LocalTime:  true,              // Использовать локальное время
				}

				// Если формат JSON, пишем напрямую в файл
				// Если формат text, используем ConsoleWriter для читаемого формата
				if logCfg.Format == "json" {
					writers = append(writers, fileWriter)
				} else {
					// Text формат в файле (без цветов)
					fileConsoleWriter := zerolog.ConsoleWriter{
						Out:        fileWriter,
						TimeFormat: time.RFC3339,
						NoColor:    true, // В файле цвета не нужны
					}
					writers = append(writers, fileConsoleWriter)
				}
			}
		}
	}

	// Если ни один writer не настроен, используем stderr по умолчанию
	if len(writers) == 0 {
		writers = append(writers, os.Stderr)
	}

	// Создаём multi-writer, который пишет во все настроенные места
	multiWriter := io.MultiWriter(writers...)

	// ============================================
	// ШАГ 3: Создаём глобальный логгер
	// ============================================
	// Настраиваем глобальный логгер с:
	//   - Уровнем логирования
	//   - Multi-writer для вывода
	//   - Временной зоной UTC
	//   - Временем в формате RFC3339
	globalLogger = zerolog.New(multiWriter).
		Level(level).    // Устанавливаем уровень
		With().          // Добавляем общие поля ко всем логам
		Timestamp().     // Добавляем время в каждый лог
		Caller().        // Добавляем информацию о месте вызова (файл:строка)
		Logger()         // Создаём финальный логгер

	// Устанавливаем глобальный уровень для zerolog
	zerolog.SetGlobalLevel(level)

	// Устанавливаем временную зону
	zerolog.TimeFieldFormat = time.RFC3339
}

// Get возвращает глобальный экземпляр логгера.
//
// ВАЖНО: перед вызовом Get() необходимо вызвать Init(),
// иначе будет использован логгер по умолчанию (stderr, Info level).
//
// Возвращает:
//   - zerolog.Logger: глобальный логгер
//
// Пример использования:
//
//	logger := logger.Get()
//	logger.Info().Msg("Application started")
//	logger.Error().Err(err).Msg("Failed to connect")
func Get() zerolog.Logger {
	// Если Init() не был вызван, возвращаем логгер по умолчанию
	// Проверяем, был ли Init() вызван (если уровень Disabled, значит нет)
	if globalLogger.GetLevel() == zerolog.Disabled {
		// Создаём простой логгер в консоль для случаев, когда Init() не вызван
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().
			Timestamp().
			Logger()
	}
	return globalLogger
}

// ============================================
// УДОБНЫЕ ФУНКЦИИ ДЛЯ БЫСТРОГО ЛОГИРОВАНИЯ
// ============================================
// Эти функции позволяют быстро логировать без получения логгера через Get()

// Debug логирует сообщение на уровне Debug.
// Используется для отладочной информации, которая нужна только при разработке.
//
// Пример:
//
//	logger.Debug().Str("user_id", "123").Msg("User logged in")
func Debug() *zerolog.Event {
	log := Get()
	return log.Debug()
}

// Info логирует сообщение на уровне Info.
// Используется для общей информации о работе приложения.
//
// Пример:
//
//	logger.Info().Str("action", "user_created").Msg("New user created")
func Info() *zerolog.Event {
	log := Get()
	return log.Info()
}

// Warn логирует сообщение на уровне Warn.
// Используется для предупреждений - что-то не так, но не критично.
//
// Пример:
//
//	logger.Warn().Str("reason", "slow_query").Msg("Database query took too long")
func Warn() *zerolog.Event {
	log := Get()
	return log.Warn()
}

// Error логирует сообщение на уровне Error.
// Используется для ошибок - что-то пошло не так.
//
// Пример:
//
//	logger.Error().Err(err).Str("operation", "create_user").Msg("Failed to create user")
func Error() *zerolog.Event {
	log := Get()
	return log.Error()
}

// Fatal логирует сообщение на уровне Fatal и завершает программу (os.Exit(1)).
// Используется только для критических ошибок, после которых приложение не может работать.
//
// ВАЖНО: после вызова Fatal() программа завершится!
//
// Пример:
//
//	logger.Fatal().Err(err).Msg("Failed to connect to database")
func Fatal() *zerolog.Event {
	log := Get()
	return log.Fatal()
}

// Panic логирует сообщение на уровне Panic и вызывает панику.
// Используется только в крайних случаях, когда нужно остановить выполнение.
//
// ВАЖНО: после вызова Panic() программа вызовет panic()!
//
// Пример:
//
//	logger.Panic().Err(err).Msg("Critical system error")
func Panic() *zerolog.Event {
	log := Get()
	return log.Panic()
}

// ============================================
// СПЕЦИАЛЬНЫЕ ФУНКЦИИ ДЛЯ РАЗНЫХ СЛУЧАЕВ
// ============================================

// WithContext создаёт новый логгер с дополнительными полями.
// Эти поля будут добавлены ко всем последующим логам.
//
// Полезно для добавления контекста к логам (например, user_id, request_id).
//
// Пример:
//
//	userLogger := logger.WithContext().Str("user_id", "123").Logger()
//	userLogger.Info().Msg("User action") // В логе будет поле user_id=123
func WithContext() zerolog.Context {
	log := Get()
	return log.With()
}

// LogRequest логирует HTTP запрос.
// Удобная функция для логирования входящих HTTP запросов.
//
// Параметры:
//   - method: HTTP метод (GET, POST, etc.)
//   - path: путь запроса
//   - statusCode: HTTP статус код ответа
//   - duration: время выполнения запроса
//   - clientIP: IP адрес клиента
//
// Пример:
//
//	logger.LogRequest("POST", "/api/users", 200, time.Since(start), "192.168.1.1")
func LogRequest(method, path string, statusCode int, duration time.Duration, clientIP string) {
	log := Get()
	log.Info().
		Str("method", method).
		Str("path", path).
		Int("status", statusCode).
		Dur("duration", duration).
		Str("client_ip", clientIP).
		Msg("HTTP request")
}

// LogDatabaseQuery логирует SQL запрос к базе данных.
// Полезно для отладки и мониторинга медленных запросов.
//
// Параметры:
//   - query: SQL запрос
//   - duration: время выполнения запроса
//   - err: ошибка (если была)
//
// Пример:
//
//	start := time.Now()
//	result, err := db.Exec("SELECT * FROM users")
//	logger.LogDatabaseQuery("SELECT * FROM users", time.Since(start), err)
func LogDatabaseQuery(query string, duration time.Duration, err error) {
	log := Get()
	event := log.Debug().
		Str("query", query).
		Dur("duration", duration)

	if err != nil {
		event.Err(err).Msg("Database query failed")
	} else {
		event.Msg("Database query")
	}
}

