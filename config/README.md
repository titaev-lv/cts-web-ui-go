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

См. `config.proxy.example.yaml` и `config.direct.example.yaml` для полной структуры. Основные секции:

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

