# 🚀 Development Plan: CT-System Migration (PHP → Go)

**Дата создания:** 14 декабря 2024  
**Дата восстановления:** 2026-02-10  
**Проект:** CT-System Web Application (Go)  
**Миграция:** PHP → Go (Gin Framework)

**Статус восстановления:**
- ✅ Восстановлен рабочий каркас и основные модули (auth, users, groups, exchanges, exchange accounts)
- ✅ Модуль `exchange_accounts` доведен до рабочего состояния (CRUD, валидация форм, UI ошибок, DataTables server-side sort/filter)
- ✅ Добавлено защищенное хранение ключей бирж (envelope encryption: DEK + HSM KEK) и fallback re-key при недоступном legacy KEK
- 🟡 Бизнес-модули восстановлены частично: `positions` реализован и подключен в роутинге; `market analysis`, `coins`, `daemon` пока не портированы в backend (часть UI-шаблонов legacy)
- ✅ Первая задача (логирование): унификация с hsm-service выполнена

---

## 📋 Обзор проекта

**CT-System** — веб-приложение для управления криптовалютными операциями:

- Управление пользователями и группами
- Управление биржами и аккаунтами
- Торговые позиции (Trade Positions)
- Рыночный анализ (Market Analysis)
- Управление демоном (Daemon)
- Управление монетами (Coins)

### Ключевые требования

| Требование | Описание |
|------------|----------|
| База данных | MySQL |
| Прокси | Nginx (SSL termination, rate limiting) |
| Безопасность | XSS, CSRF, SQL Injection защита |
| Аутентификация | Полная (все страницы требуют авторизации) |
| Группы доступа | Пользователь, Администратор |
| Формы | AJAX |
| Таблицы | DataTables + AJAX (server-side) |

---

## 🏗️ Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         NGINX (Reverse Proxy)                   │
│            - SSL Termination, Rate Limiting, Headers            │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Go Web Application                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     Security Middleware                     ││
│  │  - XSS Protection, CSRF, Rate Limiting, Input Sanitization  ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Auth Middleware                          ││
│  │  - Session/Cookie-based auth                                ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     Controllers (API)                       ││
│  │  Users│Groups│Exchanges│Positions│Market│Coins              ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     Services Layer                          ││
│  │              Business Logic, Validation                     ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     Repository Layer                        ││
│  │                 MySQL Database Access                       ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
                    ┌────────────────────┐
                    │      MySQL DB      │
                    └────────────────────┘
```

---

## 🔴 КРИТИЧНО: Унификация логирования

### ✅ Текущее состояние (на 2026-02-19)

| Параметр | Статус Web UI |
|----------|---------------|
| **Библиотека** | ✅ `log/slog` (stdlib) |
| **Формат** | ✅ JSON / Text через `slog` handlers |
| **Stdout вывод** | ✅ Да (`error` через `output`, `access/audit` через `*_to_stdout`) |
| **Ротация** | ✅ `lumberjack` |
| **Access/Error разделение** | ✅ `access.log` + `error.log` |
| **Audit log** | ✅ `audit.log` + `AuditLogMiddleware` |
| **Access middleware** | ✅ `internal/middleware/access_log.go` |
| **Request ID (сквозной)** | ✅ `X-Request-ID` middleware + context propagation |
| **Fail-fast при инициализации** | ✅ Ошибка при недоступной директории логов |
| **Graceful shutdown** | ✅ `SIGINT/SIGTERM` + `server.Shutdown()` + `logger.Close()` |
| **Module тег** | ✅ Автоопределение + ручной override |

### 🎯 Требуемая архитектура логирования

**Три логирующих потока (web-oriented):**

```
┌─────────────────────────────────────────────────────────────┐
│                    Web UI Go Application                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  HTTP Request    │         │ Application      │         │
│  │  Middleware      │         │ (Errors, Events) │         │
│  └────────┬─────────┘         └────────┬─────────┘         │
│           │ (method, path,             │ (errors, panics,  │
│           │  status, duration, IP)     │  events, tracing) │
│           ▼                             ▼                   │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  access.json     │         │  error.json      │         │
│  │  RotationManager │         │  RotationManager │         │
│  └────────┬─────────┘         └────────┬─────────┘         │
│           │                            │                   │
│           ├─────────────────────┬──────┘                   │
│           │                     │                          │
│           ▼                     ▼                          │
│     stdout + file         stdout + file                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Логирование:**
- **access.log**: HTTP запросы (Middleware level)
  - Format: JSON
  - Fields: timestamp, method, path, status_code, response_time_ms, client_ip, user_id (if auth)
  - Ротация: 50MB, keep 10, 7 дней
  - Вывод: stdout + file

- **error.log**: Ошибки приложения (slog level)
  - Format: JSON  
  - Fields: timestamp, level, message, module, error_type, stack_trace
  - Ротация: 100MB, keep 5, 30 дней
  - Вывод: stdout + file

- **audit.log**: Security/admin события (auth + mutating actions)
    - Format: JSON
    - Fields: event_type, action, resource_type, status, result, request_id, user_id, user_login, ip
    - Ротация: 100MB, keep 5, 30 дней
    - Вывод: stdout + file

### 📝 Примеры логирования (после реализации)

**access.log:**
```json
{"time":"2026-02-10T14:30:45Z","http_method":"POST","path":"/api/trades","status":201,"response_time_ms":245,"client_ip":"192.168.1.100","user_id":5}
{"time":"2026-02-10T14:30:46Z","http_method":"GET","path":"/api/positions","status":200,"response_time_ms":89,"client_ip":"192.168.1.100","user_id":5}
{"time":"2026-02-10T14:30:47Z","http_method":"DELETE","path":"/api/orders/123","status":204,"response_time_ms":156,"client_ip":"192.168.1.100","user_id":5}
```

**error.log:**
```json
{"time":"2026-02-10T14:30:48Z","level":"error","message":"Database connection failed","module":"repository","error":"EOF","stack":"..."}
{"time":"2026-02-10T14:30:49Z","level":"warn","message":"Rate limit exceeded","module":"middleware","user_id":5}
{"time":"2026-02-10T14:30:50Z","level":"error","message":"HSM encryption failed","module":"service","error":"timeout"}
```

### ✅ Что уже сделано

- Миграция на `slog` и отказ от legacy logger API.
- Реализованы два потока логирования: `error.log` и `access.log`.
- Подключен `AccessLogMiddleware` в `cmd/web/main.go`.
- Добавлен Web-UI audit поток: `audit.log` + `AuditLogMiddleware`.
- Добавлены fail-fast проверки директории логов на запись.
- Добавлен graceful shutdown HTTP сервера и закрытие лог-файлов.
- В access-логи добавлен `user_id` (если пользователь авторизован).
- Добавлен `request_id` (`X-Request-ID`) с пробросом в context, response header и access/error/recovery логи.

### 🔗 Policy: Downstream `X-Request-ID` propagation (обязательно)

Для **любых новых исходящих HTTP-запросов** из `web-ui-go` обязательно использовать helper:

- `middleware.NewRequestWithRequestID(ctx, method, url, body)`

Допустимый альтернативный вариант (если request создается вручную):

- `middleware.SetRequestIDHeaderFromContext(req)`

Цель:
- не терять корреляцию `request_id` на пути `client -> nginx -> web-ui-go -> downstream`;
- упростить трассировку инцидентов и аудит.

Минимальный шаблон:

```go
req, err := middleware.NewRequestWithRequestID(c.Request.Context(), http.MethodGet, targetURL, nil)
if err != nil {
    // handle error
}
resp, err := http.DefaultClient.Do(req)
```

Проверка: в `internal/middleware/request_id_test.go` есть тесты на перенос `X-Request-ID` в исходящий `http.Request`.

### 🔧 Что осталось (актуальный scope)

#### Фаза 1: Полировка структуры логов (0.5-1 день)
- Стандартизовать имена полей в access/error (`method/path/status/latency_ms/ip/user_id/module`).
- Проверить единообразие `module` между middleware/service/repository.

#### Фаза 2: Документация и эксплуатация (0.5 дня)
- ✅ Обновлены docs по конфигу логирования и примерам `docker logs`.
- ✅ Добавлен runbook: `services/web-ui-go/LOGGING_RUNBOOK.md`.

#### Фаза 3: Валидация в docker-compose (0.5 дня)
✅ Выполнено: проверены debug/release профили, `request_id` в header+logs, наличие `access.log`/`error.log`, наличие rotated артефактов.

```bash
# Проверить поток JSON-логов
docker logs ct-system-web-ui-1 | head -20

# Проверить наличие файлов
ls -lh /var/log/web-ui/

# Проверить ротацию
ls -lh /var/log/web-ui/access.log* /var/log/web-ui/error.log*
```

### 📋 Чек-лист реализации

```
Web UI Logging Unification (Phase 4.1)
[x] Перевести Web UI на `log/slog`
[x] Реализовать разделение `access.log` / `error.log`
[x] Подключить `AccessLogMiddleware`
[x] Настроить `lumberjack` для обоих потоков
[x] Добавить fail-fast проверку директории логов
[x] Добавить graceful shutdown + `logger.Close()`
[x] Добавить `request_id` middleware и сквозной проброс (`X-Request-ID`)
[x] Завершить финальную валидацию в docker окружении
[x] Обновить эксплуатационную документацию (runbook)
```

**Time estimate (остаток):** 0-0.5 дня  
**Difficulty:** Low-Medium  
**Impact:** HIGH (наблюдаемость и диагностика в проде)

---

## 📁 Структура проекта Go (актуально на 2026-03-25)

```
web-ui-go/
├── cmd/
│   └── web/
│       └── main.go
├── internal/
│   ├── config/                     # config.go, http2.go
│   ├── controllers/                # user/group/exchange/exchange_account/position
│   ├── db/
│   ├── dto/
│   ├── errors/
│   ├── hsm/                        # интеграция с hsm-service
│   ├── logger/
│   ├── middleware/                 # auth, admin, security, csrf, request_id, access/audit/recovery
│   ├── models/                     # user/group/exchange/exchange_account/position
│   ├── repositories/               # user/group/exchange/exchange_account/position
│   ├── services/                   # auth, user, exchange, position, csv/exchange import
│   ├── session/
│   └── utils/                      # password, validation, timezone, totp, qr
├── config/
│   ├── config.direct.example.yaml
│   ├── config.direct.yaml
│   ├── config.proxy.example.yaml
│   └── config.proxy.yaml
├── web/
│   ├── static/
│   └── templates/                  # users/groups/exchanges/exchange_accounts/positions + legacy market/coins/daemon
├── migrations/
├── php/                            # legacy PHP artifacts for migration parity
├── LOGGING_RUNBOOK.md
├── go.mod
└── Dockerfile
```

---

## 📦 Зависимости Go

```go
require (
    github.com/gin-gonic/gin              // Web framework
    github.com/go-sql-driver/mysql        // MySQL driver
    github.com/gorilla/sessions           // Session management
    github.com/spf13/viper                // Configuration
    golang.org/x/crypto                   // bcrypt password hashing
    github.com/microcosm-cc/bluemonday    // HTML sanitization (XSS)
    github.com/go-playground/validator/v10 // Input validation
    github.com/gorilla/csrf               // CSRF protection
    github.com/gorilla/sessions           // Session management
)
```

---

## 📊 Фазы разработки

### Priority 0: Унификация логирования (первая задача) ✅

**Статус:** Базовая реализация завершена, идёт финальная полировка.

**Структура логирования (3 потока):**

1. **access.log** - логирование HTTP запросов (request/response logging)
   - Middleware для отслеживания всех входящих запросов
   - Информация: метод (GET/POST), путь, статус код, время обработки, IP клиента
   - Формат: JSON
   - Ротация: ежедневная или по размеру (max 100MB)
   - Пример: `{"timestamp":"2026-02-10T10:30:45Z","method":"POST","path":"/auth/login","status":200,"duration_ms":123,"client_ip":"192.168.1.100"}`

2. **error.log** - логирование приложения и ошибок
   - Система логирования приложения (slog)
   - Информация: уровень (debug/info/warn/error), сообщение, стек, контекст
   - Формат: JSON
   - Включает: события (login, user_create, etc), ошибки БД, паники, внутренние ошибки
   - Ротация: lumberjack (max 100MB, keep 5 backups, 30 days)
   - Пример: `{"timestamp":"2026-02-10T10:30:45Z","level":"error","message":"Failed to load user groups","user_id":123,"error":"connection timeout"}`

3. **audit.log** - аудит security/admin/mutating событий
    - Middleware: `AuditLogMiddleware`
    - Формат: JSON
    - Поля: event_type, action, resource_type, result, request_id, user_id, ip
    - Ротация: lumberjack

4. **Stdout** (реал-тайм мониторинг в Docker logs)
   - Все логи дублируются в stdout для docker logs
    - Все три потока (access + error + audit) видны в docker logs <service>

**Конфигурация (config.yaml):**
```yaml
logging:
    level: "info"            # info, debug, warn, error
    format: "json"           # json или text
    output: "both"           # legacy fallback для stdout/file маршрутизации
    error_path: "/var/log/web-ui/error.log"
    access_path: "/var/log/web-ui/access.log"
    access_to_stdout: true
    out_request_to_stdout: true  # зарезервировано (поток out_request пока не выделен)
    audit_to_stdout: true
    max_size_mb: 100          # MB
    max_backups: 5            # старые файлы
    max_age_days: 30          # дни
    compress: true            # gzip архивирование
```

**Остаточные задачи:**
- [x] Проверить output в docker для debug/release режимов
- [x] Актуализировать служебную документацию по эксплуатации
- [ ] Подтвердить ротацию под нагрузкой (длительный soak-тест)
- [ ] Добавить алерты на деградацию log I/O (error/access/audit)

### Фаза 1: Базовая инфраструктура (1-2 недели)

| # | Задача | Статус |
|---|--------|--------|
| 1.1 | Создать структуру директорий проекта | ✅ |
| 1.2 | Настроить конфигурацию (YAML config loader) | ✅ |
| 1.3 | Подключение к БД (connection pool, транзакции) | ✅ |
| 1.4 | Система логирования (slog + access/error split) | ✅ |
| 1.5 | Security Middleware (XSS, CSRF, headers) | ⚠️ |
| 1.6 | Централизованная обработка ошибок | ✅ |
| 1.7 | Recovery middleware (panic handling) | ✅ |

**Результат:** Готовая инфраструктура для разработки

---

### Фаза 2: Аутентификация и авторизация (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 2.1 | User Model (структура с группами) | ✅ |
| 2.2 | Auth Service (Login, Logout, Remember Me) | ✅ |
| 2.3 | Password Hashing (bcrypt, как в PHP) | ✅ |
| 2.4 | Session Management (Cookie + DB token) | ✅ |
| 2.5 | Auth Middleware (проверка сессии, редирект) | ✅ |
| 2.6 | Role-based Access (Admin vs User) | ✅ |
| 2.7 | Login Page (шаблон + AJAX) | ✅ |
| 2.8 | 404 Page (в едином стиле) | ✅ |
| 2.9 | Logout функционал | ✅ |

**Логика аутентификации:**
1. Session-based authentication
2. Cookie "Remember Me" с токеном в БД
3. Проверка активности пользователя
4. Проверка активности групп пользователя

**Результат:** Полностью работающая система авторизации

---

### Фаза 3: Управление пользователями и группами (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 3.1 | Group Model и Repository | ✅ |
| 3.2 | User Model и Repository | ✅ |
| 3.3 | DataTables API (server-side processing) | ✅ |
| 3.4 | User Controller (List, Create, Edit, GetById) | ✅ |
| 3.5 | Group Controller (List, Create, Edit, GetById) | ✅ |
| 3.6 | Валидация (пароль, email, уникальность login) | ⚠️ |
| 3.7 | UI Templates (страницы Users, Groups) | ✅ |
| 3.8 | Модальные формы создания/редактирования | ⚠️ |

**API Endpoints:**
```
GET  /users/                   → Страница пользователей
POST /users/ajax_get_users     → DataTables данные
POST /users/ajax_create_user   → Создать пользователя
POST /users/ajax_edit_user     → Редактировать пользователя
GET  /users/ajax_getid_user    → Получить по ID

GET  /groups/                  → Страница групп
POST /groups/ajax_get_groups   → DataTables данные
POST /groups/ajax_create_group → Создать группу
POST /groups/ajax_edit_group   → Редактировать группу
GET  /groups/ajax_getid_group  → Получить по ID
```

**Результат:** CRUD для пользователей и групп

---

### Фаза 4: Управление биржами (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 4.1 | Exchange Model и Repository | ✅ |
| 4.2 | Exchange Account Model и Repository | ✅ |
| 4.3 | Exchange Service (валидация, бизнес-логика) | ✅ |
| 4.4 | Exchange Controller (List, Create, Edit) | ✅ |
| 4.5 | Exchange Account Controller | ✅ |
| 4.6 | UI Templates (Exchanges, Exchange Accounts) | ✅ |

#### Актуализация фазы 4 (2026-03-07)

**Сделано по `exchange_accounts`:**
- ✅ Исправлены SQL/schema рассинхронизации (`*_ENC`, nullable scan, ambiguous columns)
- ✅ В таблице используется `exchange_name` (JOIN с `EXCHANGE`), а не `exchange_id`
- ✅ DataTables переведен на реальную server-side фильтрацию и сортировку, включая join-поля
- ✅ Реализована валидация обязательных полей форм через `validateEmptyFormFields(...)` с подсветкой `.err`
- ✅ Ошибки backend (`400`) корректно отображаются в UI (PNotify)
- ✅ Реализовано шифрование API/Secret/Add key через DEK + HSM KEK
- ✅ Добавлен fallback сценарий: re-key/reset ключей при недешифруемом legacy DEK (retired KEK)
- ✅ Исправлен ложный кейс `RowsAffected=0` при update без изменений

**Что еще имеет смысл сделать в этой фазе (минимальный хвост):**
- [ ] Добавить интеграционные тесты create/edit для encrypted credentials и re-key fallback
- [ ] Добавить audit-событие для forced re-key (legacy DEK unrecoverable)
- [ ] Добавить runbook шаг для операторов: как перевыпускать ключи при retired KEK

**API Endpoints:**
```
GET  /exchange_manage/                  → Страница управления биржами
POST /exchange_manage/ajax_get_exchanges
POST /exchange_manage/ajax_create_exchange
POST /exchange_manage/ajax_edit_exchange
GET  /exchange_manage/ajax_getid_exchange

GET  /exchange_accounts/                → Страница аккаунтов
POST /exchange_accounts/ajax_get_accounts
POST /exchange_accounts/ajax_create_account
POST /exchange_accounts/ajax_edit_account
GET  /exchange_accounts/ajax_getid_accounts
```

**Результат:** Управление биржами и аккаунтами

---

### Фаза 5: Торговые позиции (1-2 недели)

| # | Задача | Статус |
|---|--------|--------|
| 5.1 | Position Model (позиции, транзакции) | ✅ |
| 5.2 | Position Repository | ✅ |
| 5.3 | Transaction Model и Repository | ✅ |
| 5.4 | Position Service (создание, закрытие, расчёты) | ✅ |
| 5.5 | Transaction Service (CRUD) | ✅ |
| 5.6 | CSV Upload (загрузка транзакций) | ✅ |
| 5.7 | KuCoin API Integration (цены, токены) | ✅ |
| 5.8 | Position Controller | ✅ |
| 5.9 | UI Templates (список позиций, детали) | ✅ |

**API Endpoints:**
```
GET  /positions_calc/                        → Список позиций
POST /positions_calc/ajax_get_positions
POST /positions_calc/ajax_create_position
POST /positions_calc/ajax_close_position
POST /positions_calc/ajax_delete_position

GET  /positions_calc/position/               → Детали позиции
POST /positions_calc/position/ajax_get_position
POST /positions_calc/position/ajax_edit_position
POST /positions_calc/position/ajax_get_trans
POST /positions_calc/position/ajax_create_trans
POST /positions_calc/position/ajax_delete_trans
POST /positions_calc/position/ajax_upload_trans_csv
POST /positions_calc/position/ajax_kucoin_price
POST /positions_calc/position/ajax_kucoin_token
```

**Результат:** Модуль `positions` реализован и подключен; дальнейшая задача — расширение интеграционных тестов и polish

---

### Фаза 6: Market Analysis (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 6.1 | Market Controller | ☐ |
| 6.2 | K-Lines между биржами | ☐ |
| 6.3 | Direct Arbitration между биржами | ☐ |
| 6.4 | Exchange API Clients (KuCoin, Binance, Bybit, etc.) | ☐ |
| 6.5 | Analytics Service (расчёты арбитража) | ☐ |
| 6.6 | UI Templates (графики amCharts) | ☐ |

**API Endpoints:**
```
GET  /market_analysis/                       → K-Lines страница
GET  /market_analysis/direct_exs.php         → Direct Arbitration
POST /market_analysis/ajax_exchanges_step_1
POST /market_analysis/ajax_exchanges_step_2
POST /market_analysis/ajax_exchanges_step_3
POST /market_analysis/ajax_exchanges_step_4
POST /market_analysis/ajax_direct_exs
```

**Результат:** Анализ рынка и арбитражные возможности

---

### Фаза 7: Coins Management (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 7.1 | Coin Model и Repository | ☐ |
| 7.2 | Coin Controller (обновление списка) | ☐ |
| 7.3 | CoinMarketCap API Integration | ☐ |
| 7.4 | UI Templates (Coins) | ☐ |

**API Endpoints:**
```
GET  /coins/                     → Страница монет
POST /coins/ajax_update_coins
```

**Результат:** Управление монетами

---

### Фаза 8: Безопасность и оптимизация (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 8.1 | XSS Protection (bluemonday sanitization) | ⚠️ |
| 8.2 | CSRF Protection (gorilla/csrf) | ⚠️ |
| 8.3 | SQL Injection (prepared statements) | ✅ |
| 8.4 | Rate Limiting (защита от brute-force) | ⚠️ |
| 8.5 | Secure Headers (X-Frame-Options, CSP) | ✅ |
| 8.6 | Input Validation (все входные данные) | ⚠️ |
| 8.7 | Nginx Configuration | ☐ |
| 8.8 | Dockerfile | ✅ |
| 8.9 | Тестирование безопасности | ☐ |
| 8.10 | Документация API | ☐ |

**Результат:** Безопасное production-ready приложение

---

## 🔐 Безопасность

### Security Headers

```go
// internal/middleware/security.go
func SecurityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", 
            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        c.Next()
    }
}
```

### Nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name ct-system.local;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;

    # Security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Rate limiting zones
    limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=api:10m rate=100r/s;

    location / {
        proxy_pass http://localhost:8443;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    location /auth/login {
        limit_req zone=login burst=3 nodelay;
        proxy_pass http://localhost:8443;
    }

    location /assets {
        alias /app/web/static;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
    }
}
```

---

## 📝 Модели данных

### User Model

```go
type User struct {
    ID          int        `json:"id" db:"ID"`
    Login       string     `json:"login" db:"LOGIN"`
    Password    string     `json:"-" db:"PASSWORD"`
    Email       string     `json:"email" db:"EMAIL"`
    Name        string     `json:"name" db:"NAME"`
    LastName    string     `json:"last_name" db:"LAST_NAME"`
    Active      bool       `json:"active" db:"ACTIVE"`
    Token       *string    `json:"-" db:"TOKEN"`
    Timezone    string     `json:"timezone" db:"TIMEZONE"`
    Groups      []Group    `json:"groups"`
    UserCreated *int       `json:"user_created" db:"USER_CREATED"`
    UserModify  *int       `json:"user_modify" db:"USER_MODIFY"`
    DateCreate  time.Time  `json:"date_create" db:"DATE_CREATE"`
    DateModify  *time.Time `json:"date_modify" db:"DATE_MODIFY"`
    TimestampX  time.Time  `json:"timestamp_x" db:"TIMESTAMP_X"`
}
```

### Group Model

```go
type Group struct {
    ID          int        `json:"id" db:"ID"`
    Name        string     `json:"name" db:"NAME"`
    Active      bool       `json:"active" db:"ACTIVE"`
    Description *string    `json:"description" db:"DESCRIPTION"`
    UserCreated *int       `json:"user_created" db:"USER_CREATED"`
    UserModify  *int       `json:"user_modify" db:"USER_MODIFY"`
    DateCreate  time.Time  `json:"date_create" db:"DATE_CREATE"`
    DateModify  *time.Time `json:"date_modify" db:"DATE_MODIFY"`
}
```

### Exchange Model

```go
type Exchange struct {
    ID             int        `json:"id" db:"ID"`
    Name           string     `json:"name" db:"NAME"`
    URL            string     `json:"url" db:"URL"`
    BaseURL        string     `json:"base_url" db:"BASE_URL"`
    ClassToFactory string     `json:"class" db:"CLASS_TO_FACTORY"`
    Active         bool       `json:"active" db:"ACTIVE"`
    Deleted        bool       `json:"deleted" db:"DELETED"`
    Description    *string    `json:"description" db:"DESCRIPTION"`
    UserCreated    *int       `json:"user_created" db:"USER_CREATED"`
    UserModify     *int       `json:"user_modify" db:"USER_MODIFY"`
    DateCreate     time.Time  `json:"date_create" db:"DATE_CREATE"`
    DateModify     *time.Time `json:"date_modify" db:"DATE_MODIFY"`
}
```

### Position Model

```go
type Position struct {
    ID         int       `json:"id" db:"ID"`
    Name       string    `json:"name" db:"NAME"`
    ExchangeID int       `json:"exchange_id" db:"EXID"`
    MarketType string    `json:"market_type" db:"MARKET_TYPE"`
    UserID     int       `json:"user_id" db:"USER_ID"`
    Created    time.Time `json:"created" db:"CREATED"`
    Closed     bool      `json:"closed" db:"CLOSED"`
}
```

---

## 📅 Timeline

| Фаза | Описание | Длительность | Накопительно |
|------|----------|--------------|--------------|
| 1 | Базовая инфраструктура | 1-2 недели | 2 нед |
| 2 | Аутентификация | 1 неделя | 3 нед |
| 3 | Users & Groups | 1 неделя | 4 нед |
| 4 | Exchanges | 1 неделя | 5 нед |
| 5 | Positions | 1-2 недели | 7 нед |
| 6 | Market Analysis | 1 неделя | 8 нед |
| 7 | Coins | 1 неделя | 9 нед |
| 8 | Security & Polish | 1 неделя | 10 нед |

**Общий срок: ~10 недель (2.5 месяца)**

---

## 🗄️ Структура базы данных (существующая)

### Основные таблицы

| Таблица | Описание |
|---------|----------|
| `USER` | Пользователи системы |
| `GROUP` | Группы пользователей |
| `USERS_GROUP` | Связь пользователей и групп (M:N) |
| `EXCHANGE` | Криптовалютные биржи |
| `EXCHANGE_ACCOUNT` | Аккаунты на биржах |
| `POS_POSITIONS` | Торговые позиции |
| `POS_TRANSACTIONS` | Транзакции по позициям |
| `COINS` | Список криптовалют |

---

## ✅ Чек-лист готовности к production

- [ ] Все endpoints защищены аутентификацией
- [ ] CSRF токены на всех формах
- [ ] XSS санитизация входных данных
- [ ] Prepared statements для всех SQL запросов
- [ ] Rate limiting на критичных endpoints
- [ ] Secure headers настроены
- [ ] HTTPS обязателен
- [ ] Пароли хешируются bcrypt
- [ ] Логирование всех действий
- [ ] Graceful shutdown
- [ ] Health check endpoint
- [ ] Dockerfile готов
- [ ] Nginx конфигурация готова
- [ ] Документация API

---

## 📞 Контакты и ресурсы

- **PHP код (для анализа):** `/php/`
- **Go код:** `/cmd/`, `/internal/`, `/web/`
- **Статические файлы:** `/web/static/`
- **Шаблоны:** `/web/templates/`
- **Конфигурация:** `/config/config.yaml`

---

*Документ создан автоматически. Последнее обновление: 7 марта 2026*

