# Configuration

Конфигурационные файлы приложения.

## Файлы

- **config.proxy.yaml** - Рабочий конфиг для `proxy` режима (не коммитится в git)
- **config.direct.yaml** - Рабочий конфиг для `direct` режима (не коммитится в git)
- **config.proxy.example.yaml** - Пример конфига для `proxy` режима (коммитится в git)
- **config.direct.example.yaml** - Пример конфига для `direct` режима (коммитится в git)

## Быстрый старт

1. Скопируйте примеры конфигурации:
   ```bash
   cp config/config.proxy.example.yaml config/config.proxy.yaml
   cp config/config.direct.example.yaml config/config.direct.yaml
   ```

2. Отредактируйте файлы под нужный профиль:
   - `config/config.proxy.yaml` для `proxy`
   - `config/config.direct.yaml` для `direct`

   Укажите свои значения:
   - Параметры подключения к БД
   - Секретные ключи (Session, CSRF)
   - Режим запуска (`proxy` или `direct`)

3. Запустите приложение - конфигурация загрузится автоматически.

## Переменные окружения

Конфигурация также может быть переопределена через переменные окружения с префиксом `CT_`:

```bash
# Примеры
export CT_SERVER_PORT=8080
export CT_DATABASES_SYSTEM_MYSQL_HOST=mysql.example.com
export CT_SECURITY_SESSION_SECRET=my-session-secret
```

## Минимальные примеры и дефолты

`config.proxy.example.yaml` и `config.direct.example.yaml` теперь содержат минимально необходимый набор параметров для запуска.

Часть полей намеренно опущена и берётся из дефолтов в коде (см. `internal/config/config.go` и `internal/logger/logger.go`):
- `server.timeouts.*`
- `server.limits.max_header_bytes`
- `server.http2.*`
- `rate_limit.*.burst`
- расширенные настройки ротации/пути `logging.*` (кроме `error_path`)

Для stdout-маршрутизации логов:
- `logging.access_to_stdout` — дублирование `access.log` в stdout
- `logging.audit_to_stdout` — дублирование `audit.log` в stdout
- `logging.out_request_to_stdout` — зарезервировано (поток `out_request` пока не выделен)
- `logging.output` остаётся legacy fallback для базовой маршрутизации stdout/file

Полезно знать:
- `server.tls.cert_path` и `server.tls.key_path` обязательны только при `server.tls.enabled=true`.
- Для proxy-режима оставляйте `proxy.trust_forward_headers=true` и `proxy.trusted_cidrs` с сетью вашего nginx.

Основные секции:

- **server** - Настройки HTTP сервера
   - `server.tls.enabled` включает HTTPS
   - `server.tls.cert_path` и `server.tls.key_path` задают сертификат и ключ
   - `server.tls.ca_path` сейчас резервный параметр (под будущую проверку цепочки/mTLS), на старте HTTP сервера пока не используется
   - `server.timeouts.read`, `write`, `idle`, `read_header`, `shutdown_grace` управляют таймаутами сервера
   - `server.limits.max_header_bytes` ограничивает общий размер HTTP заголовков
   - `server.http2.*` (опционально) позволяет тюнить HTTP/2: `max_concurrent_streams`, `initial_window_size`, `max_frame_size`, `max_header_list_size`, `idle_timeout_seconds`, `max_upload_buffer_per_conn`, `max_upload_buffer_per_stream`
- **databases** - Настройки подключений к БД по функциональным зонам
   - `databases.system.engine` — движок основной БД (сейчас `mysql`)
   - `databases.system.mysql.host|port|user|password|database|charset|parse_time` — базовые параметры подключения
   - `databases.system.mysql.pool.*` — параметры пула (`max_open_conns`, `max_idle_conns`, `conn_max_lifetime`, `conn_max_idle_time`)
   - `databases.system.mysql.tls.*` — TLS для исходящего подключения к MySQL (`enabled`, `ca_path`, `cert_path`, `key_path`)
   - `databases.system.mysql.retry.*` — retry-политика (`max_attempts`, `initial_delay`, `max_delay`, `multiplier`)
   - `databases.audit` и `databases.quotes` — зарезервированы под отдельные хранилища
- **security** - Секретные ключи и настройки безопасности
- **security.csrf** - Явное управление CSRF middleware
   - `security.csrf.enabled` — включает/выключает CSRF middleware
   - `security.csrf.cookie_name`, `security.csrf.header_name` — имена cookie/заголовка токена
   - `security.csrf.trusted_origins` — список доверенных origin для cross-origin POST

Примеры `security.csrf.trusted_origins`:

```yaml
# 1) Same-origin (UI и backend на одном домене) — обычно пусто
security:
   csrf:
      trusted_origins: []
```

```yaml
# 2) Proxy/поддомены (frontend на app.example.com, backend/API на api.example.com)
security:
   csrf:
      trusted_origins:
         - "https://app.example.com"
```

```yaml
# 3) Dev SPA + local backend
security:
   csrf:
      trusted_origins:
         - "http://localhost:3000"
         - "http://127.0.0.1:3000"
```

Важно:
- Указывайте только origin (схема + хост + порт), без путей (`/api`, `/login`).
- Для production указывайте только HTTPS origin'ы реальных фронтенд-доменов.
- Не добавляйте `*` и лишние origin'ы — это ослабляет защиту.
- `security.session.cookie_secure` автоматически считается включенным при `server.tls.enabled=true`
- **rate_limit** - Лимиты запросов приложения
   - `rate_limit.login.requests_per_second`, `rate_limit.login.burst`
   - `rate_limit.api.requests_per_second`, `rate_limit.api.burst`
   - `security.rate_limit_login` и `security.rate_limit_api` считаются устаревшими (fallback только если `rate_limit.*` не задан)
   - при одновременной установке legacy и новых полей с разными значениями конфиг считается невалидным
- **logging** - Настройки логирования

## Proxy mode (`proxy.*`)

Для запуска за reverse proxy (nginx) используйте секцию:

```yaml
proxy:
   enabled: true
   trust_forward_headers: true
   trusted_hops: 1
   trusted_cidrs:
      - "172.16.0.0/12"
   static_via_nginx: true
```

Рекомендации:
- В proxy mode backend обычно запускается без TLS (`server.tls.enabled: false`), TLS и HTTP/2 терминируются на edge (`nginx`).
- При `static_via_nginx: true` Gin не регистрирует `r.Static("/assets", ...)`.
- `X-Forwarded-*` учитываются только при `proxy.enabled=true`, `trust_forward_headers=true` и доверенном источнике (`trusted_cidrs`).

## Troubleshooting forwarded headers

- **Неверная схема (`http` вместо `https`) в логах/cookie logic**
   - Проверьте, что nginx отправляет `X-Forwarded-Proto $scheme`.
   - Проверьте, что IP proxy попадает в `proxy.trusted_cidrs`.
   - Проверьте `proxy.trusted_hops` (обычно `1` для single nginx).

- **`trusted_proxy=false` в access/audit логах web-ui-go**
   - Запрос пришел не от доверенного proxy источника.
   - Или `proxy.trust_forward_headers=false`.

- **Потеря `X-Request-ID`**
   - Проверьте `proxy_set_header X-Request-ID $request_id;` на nginx.
   - Убедитесь, что в web-ui-go включен middleware `RequestIDMiddleware` (в `cmd/web/main.go`).

## Короткий итог реализации (2026-02)

- `proxy.*` добавлен в schema/loader/validation и используется в runtime.
- Поддерживаются оба режима одного бинарника: direct и behind nginx.
- Для proxy mode добавлены проверки trust policy, effective scheme/ip/host и корреляции `X-Request-ID`.
- Smoke-покрытие доступно через `make smoke-web-ui` (включая optional secure-cookie check).

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

