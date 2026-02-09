# Система логирования

**Дата:** 14 декабря 2024

## Обзор

Система логирования построена на основе **zerolog** - быстрой и структурированной библиотеки логирования для Go.

### Основные возможности

- ✅ Структурированное логирование (JSON или читаемый текст)
- ✅ Разные уровни логирования (Debug, Info, Warn, Error, Fatal, Panic)
- ✅ Вывод в консоль, файл или оба места одновременно
- ✅ Автоматическая ротация файлов логов
- ✅ Сжатие старых логов (gzip)
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
logger.Init()

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

Zerolog поддерживает разные типы полей:

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

## Контекстное логирование

Создание логгера с постоянными полями:

```go
// Создаём логгер с контекстом
userLogger := logger.WithContext().
    Str("user_id", "789").
    Str("session_id", "abc123").
    Logger()

// Все логи этого логгера будут содержать user_id и session_id
userLogger.Info().Msg("User action 1")
userLogger.Info().Msg("User action 2")
```

## Специальные функции

### Логирование HTTP запросов

```go
logger.LogRequest("POST", "/api/users", 200, time.Since(start), "192.168.1.1")
```

### Логирование SQL запросов

```go
start := time.Now()
result, err := db.Exec("SELECT * FROM users")
logger.LogDatabaseQuery("SELECT * FROM users", time.Since(start), err)
```

## Конфигурация

Настройки логирования в `config.yaml`:

```yaml
logging:
  level: "info"        # debug, info, warn, error
  format: "text"       # text (читаемый) или json (структурированный)
  output: "both"       # stdout, file или both
  file: "./logs/ct-system.log"
  max_size: 100        # МБ
  max_backups: 5       # Количество архивных файлов
  max_age: 30          # Дни
  compress: true       # Сжимать старые логи
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
  ct-system.log          # Текущий файл
  ct-system.log.20241214 # Архив за 14 декабря
  ct-system.log.20241213.gz # Сжатый архив за 13 декабря
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

### Создание пользователя

```go
func CreateUser(name, email string) error {
    logger.Info().
        Str("operation", "create_user").
        Str("email", email).
        Msg("Creating new user")
    
    userID, err := createUserInDB(name, email)
    if err != nil {
        logger.Error().
            Err(err).
            Str("email", email).
            Msg("Failed to create user")
        return err
    }
    
    logger.Info().
        Int("user_id", userID).
        Str("email", email).
        Msg("User created successfully")
    
    return nil
}
```

### HTTP запрос

```go
func HandleRequest(c *gin.Context) {
    start := time.Now()
    
    // ... обработка запроса ...
    
    logger.LogRequest(
        c.Request.Method,
        c.Request.URL.Path,
        c.Writer.Status(),
        time.Since(start),
        c.ClientIP(),
    )
}
```

## Дополнительные ресурсы

- [Zerolog Documentation](https://github.com/rs/zerolog)
- [Lumberjack (Rotation)](https://github.com/natefinch/lumberjack)

---

*Логирование - важный инструмент для отладки и мониторинга приложения!*

