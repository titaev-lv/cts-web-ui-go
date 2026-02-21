# Configuration

Конфигурационные файлы приложения.

## Файлы

- **config.yaml** - Основной конфигурационный файл (не коммитится в git, создаётся из примера)
- **config.example.yaml** - Пример конфигурации (коммитится в git)

## Быстрый старт

1. Скопируйте пример конфигурации:
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

2. Отредактируйте `config/config.yaml` и укажите свои значения:
   - Параметры подключения к БД
   - Секретные ключи (JWT, Session, CSRF)
   - Настройки сервера

3. Запустите приложение - конфигурация загрузится автоматически.

## Переменные окружения

Конфигурация также может быть переопределена через переменные окружения с префиксом `CT_`:

```bash
# Примеры
export CT_SERVER_PORT=8080
export CT_DATABASE_MYSQL_HOST=mysql.example.com
export CT_SECURITY_JWT_SECRET=my-secret-key
```

## Структура конфигурации

См. `config.example.yaml` для полной структуры. Основные секции:

- **server** - Настройки HTTP сервера
   - `server.tls.enabled` включает HTTPS
   - `server.tls.cert_path` и `server.tls.key_path` задают сертификат и ключ
   - `server.tls.ca_path` сейчас резервный параметр (под будущую проверку цепочки/mTLS), на старте HTTP сервера пока не используется
   - `server.timeouts.read`, `write`, `idle`, `read_header`, `shutdown_grace` управляют таймаутами сервера
   - `server.limits.max_header_bytes` ограничивает общий размер HTTP заголовков
   - `server.http2.*` (опционально) позволяет тюнить HTTP/2: `max_concurrent_streams`, `initial_window_size`, `max_frame_size`, `max_header_list_size`, `idle_timeout_seconds`, `max_upload_buffer_per_conn`, `max_upload_buffer_per_stream`
- **database** - Настройки подключения к БД
- **security** - Секретные ключи и настройки безопасности
- `security.session_cookie_secure` автоматически считается включенным при `server.tls.enabled=true`
- **rate_limit** - Лимиты запросов приложения
   - `rate_limit.login.requests_per_minute`, `rate_limit.login.burst`
   - `rate_limit.api.requests_per_second`, `rate_limit.api.burst`
   - `security.rate_limit_login` и `security.rate_limit_api` считаются устаревшими (fallback для совместимости)
- **logging** - Настройки логирования
- **daemon** - Настройки демона
- **app** - Общие настройки приложения

## Использование в коде

```go
import "ctweb/internal/config"

// Загрузка конфигурации (обычно в main.go)
cfg, err := config.Load("")
if err != nil {
    log.Fatal(err)
}

// Использование
port := config.Get().Server.Port
dsn := config.Get().GetMySQLDSN()
```

