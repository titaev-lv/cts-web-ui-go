# Система логирования

**Дата:** 14 декабря 2024

## Обзор

Система логирования построена на основе **log/slog** - стандартной библиотеки логирования для Go.

### Основные возможности

- ✅ Структурированное логирование (JSON или читаемый текст)
- ✅ Разные уровни логирования (Debug, Info, Warn, Error, Fatal, Panic)
- ✅ Вывод в консоль, файл или оба места одновременно
- ✅ Разделение на error.log и access.log
- ✅ Отдельный audit.log для security/admin событий
- ✅ Автоматическая ротация файлов логов
- ✅ Сжатие старых логов (gzip)
- ✅ Fail-fast проверка директории логов на запись
- ✅ Автоматический `module` тег по caller path
- ✅ Интеграция с конфигурацией

## Уровни логирования

Уровни логирования (от меньшего к большему):

| Уровень | Когда использовать | Пример |
|---------|-------------------|--------|
| **Trace** | Очень детальная информация (редко используется) | Пошаговое выполнение алгоритма |
| **Debug** | Отладочная информация (только для разработки) | SQL запросы, промежуточные значения |
| **Info** | Общая информация о работе приложения | Запуск сервера, создание пользователя |
| **Warn** | Предупреждения (не критично, но стоит обратить внимание) | Медленный запрос, устаревший API |
| **Error** | Ошибки (что-то пошло не так) | Ошибка подключения к БД, валидация не прошла |
| **Fatal** | Критическая ошибка (приложение завершится) | Не удалось подключиться к БД при старте |
| **Panic** | Паника (приложение вызовет panic) | Критическая системная ошибка |

## Инициализация

### В main.go

```go
// 1. Загружаем конфигурацию
cfg, err := config.Load("")
if err != nil {
    log.Fatal(err)
}

// 2. Инициализируем логгер
if err := logger.Init(); err != nil {
    log.Fatal(err)
}
defer logger.Close()

// 3. Теперь можно использовать
logger.Info().Msg("Application started")
```

## Базовое использование

### Простое сообщение

```go
logger.Info().Msg("User logged in")
```

### С дополнительными полями

```go
logger.Info().
    Str("user_id", "123").
    Str("action", "login").
    Msg("User logged in")
```

### Логирование ошибок

```go
err := errors.New("connection failed")
logger.Error().
    Err(err).
    Str("operation", "connect_db").
    Msg("Failed to connect to database")
```

## Типы полей

Обёртка `logger.Event` поддерживает разные типы полей:

```go
logger.Info().
    Str("string", "value").          // Строка
    Int("int", 123).                 // Целое число
    Int64("int64", 456).             // 64-битное целое
    Float64("float", 3.14).          // Число с плавающей точкой
    Bool("bool", true).              // Булево значение
    Dur("duration", 2*time.Second). // Длительность
    Time("time", time.Now()).        // Время
    Msg("Log entry")
```

## Уровни логирования

### Debug

```go
logger.Debug().
    Str("query", "SELECT * FROM users").
    Dur("duration", 150*time.Millisecond).
    Msg("Database query executed")
```

### Info

```go
logger.Info().
    Str("event", "user_created").
    Int("user_id", 456).
    Msg("New user created")
```

### Warn

```go
logger.Warn().
    Str("reason", "slow_query").
    Dur("duration", 5*time.Second).
    Msg("Database query took too long")
```

### Error

```go
logger.Error().
    Err(err).
    Str("operation", "create_user").
    Msg("Failed to create user")
```

### Fatal (завершит программу!)

```go
logger.Fatal().
    Err(err).
    Msg("Critical error, shutting down")
// Программа завершится после этого лога
```

## Модульность

По умолчанию `module` добавляется автоматически на основе пути вызывающего пакета (например: `controllers`, `services`, `repositories`, `middleware`).

Если нужно переопределить модуль вручную:

```go
logger.Info().
    Str("module", "auth").
    Str("event", "login").
    Msg("User authenticated")
```

## Access логирование

HTTP access-логи пишутся через middleware `internal/middleware/access_log.go` в `access.log` и дублируются в stdout при `logging.access_to_stdout=true`.

Стандартные поля access-лога:
- `module`
- `method`
- `path`
- `status`
- `latency_ms`
- `ip`
- `user_agent`
- `user_id` (если пользователь авторизован)

## Audit логирование

Audit-логи пишутся через middleware `internal/middleware/audit_log.go` в `audit.log` и дублируются в stdout при `logging.audit_to_stdout=true`.

Audit покрывает Web-UI ориентированные события:
- auth события (`/auth/login`, `/auth/logout`)
- mutating операции (`POST/PUT/PATCH/DELETE`), кроме `ajax_get*`

Стандартные поля audit-лога:
- `event_type` (`audit`)
- `action`
- `resource_type`
- `method`, `path`, `status`, `result`
- `request_id`
- `user_id`, `user_login` (если пользователь известен)
- `ip`, `user_agent`, `latency_ms`

## Конфигурация

Настройки логирования в `config.yaml`:

```yaml
logging:
    level: "info"            # debug, info, warn, error
    format: "text"           # text (читаемый) или json (структурированный)
    output: "both"           # legacy fallback для stdout/file
    error_path: "/var/log/web-ui/error.log"
    max_size_mb: 100          # МБ
    max_backups: 5            # Количество архивных файлов
    max_age_days: 30          # Дни
    compress: true            # Сжимать старые логи
    access_path: "/var/log/web-ui/access.log"
    access_to_stdout: true
    out_request_to_stdout: true  # зарезервировано (поток out_request пока не выделен)
    audit_path: "/var/log/web-ui/audit.log"
    audit_to_stdout: true
```

## Форматы вывода

### Text формат (читаемый)

```
2024-12-14T18:30:45+03:00 INF Application started app=CT-System version=1.0.0
2024-12-14T18:30:46+03:00 INF User logged in user_id=123 action=login
```

### JSON формат (структурированный)

```json
{"level":"info","time":"2024-12-14T18:30:45+03:00","message":"Application started","app":"CT-System","version":"1.0.0"}
{"level":"info","time":"2024-12-14T18:30:46+03:00","message":"User logged in","user_id":"123","action":"login"}
```

## Ротация файлов

Логи автоматически ротируются при достижении `max_size`:

```
logs/
    error.log
    error.log.1.gz
    access.log
    access.log.1.gz
```

## Сравнение с PHP кодом

| PHP | Go |
|-----|-----|
| `Log::writeLog('info', 'message')` | `logger.Info().Msg("message")` |
| `Log::writeLog('error', 'message')` | `logger.Error().Msg("message")` |
| `Log::writeLog('debug', 'message')` | `logger.Debug().Msg("message")` |

## Лучшие практики

### ✅ ДЕЛАТЬ:

1. **Использовать структурированные поля**
   ```go
   logger.Info().
       Str("user_id", userID).
       Str("action", "login").
       Msg("User logged in")
   ```

2. **Добавлять контекст к ошибкам**
   ```go
   logger.Error().
       Err(err).
       Str("operation", "create_user").
       Int("user_id", userID).
       Msg("Failed to create user")
   ```

3. **Использовать правильный уровень**
   - Debug - только для разработки
   - Info - важные события
   - Warn - предупреждения
   - Error - ошибки

### ❌ НЕ ДЕЛАТЬ:

1. **Не логировать чувствительные данные**
   ```go
   // ПЛОХО
   logger.Info().Str("password", password).Msg("User login")
   
   // ХОРОШО
   logger.Info().Str("user_id", userID).Msg("User login")
   ```

2. **Не использовать Fatal без необходимости**
   - Fatal завершит программу
   - Используйте только для критических ошибок при старте

3. **Не логировать слишком часто**
   - Debug логи в цикле могут замедлить приложение
   - Используйте Info для важных событий

## Примеры из реального кода

### Лог события в сервисе

```go
logger.Info().
    Str("event", "user_created").
    Int("user_id", userID).
    Msg("User created successfully")
```

### Лог ошибки

```go
logger.Error().
    Err(err).
    Str("operation", "create_user").
    Msg("Failed to create user")
```

## Дополнительные ресурсы

- Go `log/slog` package documentation
- [Lumberjack (Rotation)](https://github.com/natefinch/lumberjack)

---

*Логирование - важный инструмент для отладки и мониторинга приложения!*

