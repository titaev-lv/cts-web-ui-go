# 🚀 Development Plan: CT-System Migration (PHP → Go)

**Дата создания:** 14 декабря 2024  
**Проект:** CT-System Web Application  
**Миграция:** PHP → Go (Gin Framework)

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
│  │  - Session/Cookie-based auth, JWT tokens                    ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     Controllers (API)                       ││
│  │  Users│Groups│Exchanges│Positions│Market│Daemon│Coins       ││
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

## 📁 Структура проекта Go

```
www-go/
├── cmd/
│   └── web/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loader
│   ├── middleware/
│   │   ├── auth.go                 # Authentication middleware
│   │   ├── security.go             # XSS, CSRF, headers
│   │   ├── logging.go              # Request logging
│   │   └── recovery.go             # Panic recovery
│   ├── models/
│   │   ├── user.go
│   │   ├── group.go
│   │   ├── exchange.go
│   │   ├── exchange_account.go
│   │   ├── position.go
│   │   └── coin.go
│   ├── repositories/
│   │   ├── user_repository.go
│   │   ├── group_repository.go
│   │   ├── exchange_repository.go
│   │   ├── position_repository.go
│   │   └── coin_repository.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── group_service.go
│   │   ├── exchange_service.go
│   │   ├── position_service.go
│   │   ├── daemon_service.go
│   │   └── coin_service.go
│   ├── controllers/
│   │   ├── auth_controller.go
│   │   ├── user_controller.go
│   │   ├── group_controller.go
│   │   ├── exchange_controller.go
│   │   ├── position_controller.go
│   │   ├── daemon_controller.go
│   │   ├── market_controller.go
│   │   └── coin_controller.go
│   ├── dto/
│   │   ├── request/                # Request DTOs with validation
│   │   └── response/               # Response DTOs
│   ├── db/
│   │   └── mysql.go
│   ├── utils/
│   │   ├── password.go             # bcrypt hashing
│   │   ├── validator.go            # Input validation
│   │   └── sanitizer.go            # HTML/XSS sanitization
│   └── logger/
│       └── logger.go
├── web/
│   ├── static/                     # Static assets (from PHP assets)
│   │   ├── images/
│   │   ├── javascripts/
│   │   ├── stylesheets/
│   │   └── vendor/
│   └── templates/                  # HTML templates
│       ├── layouts/
│       │   ├── base.html
│       │   ├── header.html
│       │   └── footer.html
│       ├── auth/
│       │   └── login.html
│       ├── errors/
│       │   ├── 404.html
│       │   └── 500.html
│       ├── users/
│       ├── groups/
│       ├── exchanges/
│       ├── positions/
│       ├── market/
│       ├── daemon/
│       └── coins/
├── config/
│   └── config.yaml                 # Application config
├── go.mod
├── go.sum
└── Dockerfile
```

---

## 📦 Зависимости Go

```go
require (
    github.com/gin-gonic/gin              // Web framework
    github.com/go-sql-driver/mysql        // MySQL driver
    github.com/golang-jwt/jwt/v5          // JWT tokens
    github.com/rs/zerolog                 // Structured logging
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

### Фаза 1: Базовая инфраструктура (1-2 недели)

| # | Задача | Статус |
|---|--------|--------|
| 1.1 | Создать структуру директорий проекта | ☐ |
| 1.2 | Настроить конфигурацию (YAML config loader) | ☐ |
| 1.3 | Подключение к БД (connection pool, транзакции) | ☐ |
| 1.4 | Система логирования (zerolog) | ☐ |
| 1.5 | Security Middleware (XSS, CSRF, headers) | ☐ |
| 1.6 | Централизованная обработка ошибок | ☐ |
| 1.7 | Recovery middleware (panic handling) | ☐ |

**Результат:** Готовая инфраструктура для разработки

---

### Фаза 2: Аутентификация и авторизация (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 2.1 | User Model (структура с группами) | ☐ |
| 2.2 | Auth Service (Login, Logout, Remember Me) | ☐ |
| 2.3 | Password Hashing (bcrypt, как в PHP) | ☐ |
| 2.4 | Session Management (Cookie + DB token) | ☐ |
| 2.5 | Auth Middleware (проверка сессии, редирект) | ☐ |
| 2.6 | Role-based Access (Admin vs User) | ☐ |
| 2.7 | Login Page (шаблон + AJAX) | ☐ |
| 2.8 | 404 Page (в едином стиле) | ☐ |
| 2.9 | Logout функционал | ☐ |

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
| 3.1 | Group Model и Repository | ☐ |
| 3.2 | User Model и Repository | ☐ |
| 3.3 | DataTables API (server-side processing) | ☐ |
| 3.4 | User Controller (List, Create, Edit, GetById) | ☐ |
| 3.5 | Group Controller (List, Create, Edit, GetById) | ☐ |
| 3.6 | Валидация (пароль, email, уникальность login) | ☐ |
| 3.7 | UI Templates (страницы Users, Groups) | ☐ |
| 3.8 | Модальные формы создания/редактирования | ☐ |

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
| 4.1 | Exchange Model и Repository | ☐ |
| 4.2 | Exchange Account Model и Repository | ☐ |
| 4.3 | Exchange Service (валидация, бизнес-логика) | ☐ |
| 4.4 | Exchange Controller (List, Create, Edit) | ☐ |
| 4.5 | Exchange Account Controller | ☐ |
| 4.6 | UI Templates (Exchanges, Exchange Accounts) | ☐ |

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
| 5.1 | Position Model (позиции, транзакции) | ☐ |
| 5.2 | Position Repository | ☐ |
| 5.3 | Transaction Model и Repository | ☐ |
| 5.4 | Position Service (создание, закрытие, расчёты) | ☐ |
| 5.5 | Transaction Service (CRUD) | ☐ |
| 5.6 | CSV Upload (загрузка транзакций) | ☐ |
| 5.7 | KuCoin API Integration (цены, токены) | ☐ |
| 5.8 | Position Controller | ☐ |
| 5.9 | UI Templates (список позиций, детали) | ☐ |

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
GET  /positions_calc/position/ajax_kucoin_price
GET  /positions_calc/position/ajax_kucoin_token
```

**Результат:** Полное управление торговыми позициями

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

### Фаза 7: Daemon & Coins Management (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 7.1 | Daemon Model и Service | ☐ |
| 7.2 | Daemon Controller (Start, Stop, Status) | ☐ |
| 7.3 | Daemon Launcher (запуск внешнего процесса) | ☐ |
| 7.4 | Coin Model и Repository | ☐ |
| 7.5 | Coin Controller (обновление списка) | ☐ |
| 7.6 | CoinMarketCap API Integration | ☐ |
| 7.7 | UI Templates (Daemon, Coins) | ☐ |

**API Endpoints:**
```
GET  /daemon/                    → Страница управления демоном
POST /daemon/ajax_start
POST /daemon/ajax_stop
GET  /daemon/ajax_check_status
GET  /daemon/ajax_daemon_stat

GET  /coins/                     → Страница монет
POST /coins/ajax_update_coins
```

**Результат:** Управление демоном и монетами

---

### Фаза 8: Безопасность и оптимизация (1 неделя)

| # | Задача | Статус |
|---|--------|--------|
| 8.1 | XSS Protection (bluemonday sanitization) | ☐ |
| 8.2 | CSRF Protection (gorilla/csrf) | ☐ |
| 8.3 | SQL Injection (prepared statements) | ☐ |
| 8.4 | Rate Limiting (защита от brute-force) | ☐ |
| 8.5 | Secure Headers (X-Frame-Options, CSP) | ☐ |
| 8.6 | Input Validation (все входные данные) | ☐ |
| 8.7 | Nginx Configuration | ☐ |
| 8.8 | Dockerfile | ☐ |
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
| 7 | Daemon & Coins | 1 неделя | 9 нед |
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

*Документ создан автоматически. Последнее обновление: 14 декабря 2024*

