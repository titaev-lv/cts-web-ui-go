# Session Management (Cookie + DB token)

Система управления сессиями для аутентификации пользователей, полностью совместимая с PHP версией приложения.

## Обзор

Session Manager использует:
- **Cookies** для хранения сессий (через `gorilla/sessions`)
- **Cookies "Remember Me"** для долгосрочной аутентификации (Login + CTToken)
- **Токены в БД** для проверки "Remember Me" (поле `TOKEN` в таблице `USER`)

## Архитектура

### Компоненты

1. **SessionManager** - основной класс для управления сессиями
2. **CookieStore** - хранилище сессий на основе cookies
3. **UserRepository** - для работы с токенами в БД

### Поток аутентификации

```
1. Пользователь входит → AuthService.Login()
   ├─ Проверка логина/пароля
   ├─ Генерация токена (если remember = true)
   └─ Сохранение токена в БД

2. Контроллер Login
   ├─ SessionManager.SetUser() - сохранение в сессию
   └─ SessionManager.SetRememberMeCookies() - установка cookies (если remember)

3. Последующие запросы
   ├─ Middleware проверяет сессию
   ├─ Если сессии нет → проверяет cookies "Remember Me"
   └─ Если cookies валидны → восстанавливает сессию
```

## Использование

### Инициализация

```go
import "ctweb/internal/session"

// В main.go при старте приложения
session.Init()
```

### Сохранение пользователя в сессию

```go
sm := session.GetSessionManager()
err := sm.SetUser(r, w, user)
if err != nil {
    // обработка ошибки
}
```

### Получение пользователя из сессии

```go
sm := session.GetSessionManager()
user, isAuth, err := sm.GetUser(r)
if err != nil {
    // обработка ошибки
}
if isAuth {
    // пользователь авторизован
}
```

### Восстановление из cookies "Remember Me"

```go
sm := session.GetSessionManager()
user, restored, err := sm.RestoreUserFromCookies(r, w)
if err != nil {
    // обработка ошибки
}
if restored {
    // пользователь восстановлен из cookies
}
```

### Очистка сессии (выход)

```go
sm := session.GetSessionManager()
err := sm.ClearUser(r, w)
if err != nil {
    // обработка ошибки
}
```

## Структура сессии

Данные, хранящиеся в сессии (как в PHP):

```go
session.Values["ct_auth"] = true              // Флаг авторизации
session.Values["ct_user_uid"] = 1             // ID пользователя
session.Values["ct_user_name"] = "John Doe"   // Имя пользователя
session.Values["ct_user_email"] = "user@..."  // Email
session.Values["ct_user_grp"] = [1, 2]        // ID групп
session.Values["ct_user_timezone"] = "UTC"    // Часовой пояс
```

## Cookies "Remember Me"

### Имена cookies (как в PHP)

- `Login` - логин пользователя
- `CTToken` - токен из БД (поле `TOKEN` в таблице `USER`)

### Установка cookies

```go
sm := session.GetSessionManager()
sm.SetRememberMeCookies(w, "admin", "abc123...")
```

### Параметры cookies

- **Path**: `/` (доступен для всего сайта)
- **Expires**: `RememberMeDays` дней (по умолчанию 7 дней)
- **HttpOnly**: `true` (защита от XSS)
- **Secure**: из конфигурации (HTTPS в продакшн)
- **SameSite**: из конфигурации (Lax/Strict/None)

## Генерация токена

```go
sm := session.GetSessionManager()
token, err := sm.GenerateRememberMeToken()
if err != nil {
    // обработка ошибки
}
// Сохраняем токен в БД
userRepo.UpdateToken(userID, token)
```

**Безопасность:**
- Использует `crypto/rand` для генерации случайных байт
- Длина токена: 32 байта (256 бит) в hex формате
- Безопаснее, чем в PHP (где используется `mt_rand(2, 100)` байт)

## Конфигурация

Настройки в `config/config.yaml`:

```yaml
server:
    tls:
        enabled: true
        cert_path: "pki/server/web-ui.crt"
        key_path: "pki/server/web-ui.key"
        ca_path: "pki/ca/ca.crt"

security:
  session_secret: "your-secret-key-here"  # Секретный ключ для сессий
  session_cookie_name: "ct_session"        # Имя cookie для сессии
    session_cookie_secure: false            # автоматически true при server.tls.enabled=true
  session_cookie_http_only: true          # Защита от XSS
  session_cookie_same_site: "Lax"         # Политика SameSite
  session_max_age: 86400                  # Время жизни сессии (24 часа)
  remember_me_days: 7                     # Количество дней для "Remember Me"
```

## Сравнение с PHP

### PHP код:
```php
// Сохранение в сессию
$_SESSION['ct_auth'] = true;
$_SESSION['ct_user']['uid'] = $user_id;
$_SESSION['ct_user']['name'] = $user_name;

// Cookies "Remember Me"
setcookie("Login", $login, $expires, '/');
setcookie("CTToken", $token, $expires, '/');

// Восстановление из cookies
$login = $_COOKIE['Login'];
$token = $_COOKIE['CTToken'];
$user = $DB->select("SELECT * FROM USER WHERE LOGIN = ? AND TOKEN = ?", [$login, $token]);
```

### Go код:
```go
// Сохранение в сессию
sm.SetUser(r, w, user)

// Cookies "Remember Me"
sm.SetRememberMeCookies(w, login, token)

// Восстановление из cookies
user, restored, err := sm.RestoreUserFromCookies(r, w)
```

## Безопасность

### Защита от XSS
- Cookies с флагом `HttpOnly: true` - недоступны через JavaScript
- Сессии подписываются секретным ключом

### Защита от CSRF
- Используется политика `SameSite` для cookies
- Сессии проверяются на каждом запросе

### Защита токенов
- Токены хранятся только в БД и cookies (не в сессии)
- Токены генерируются криптографически стойким генератором
- Токены проверяются при каждом восстановлении сессии

## Интеграция с AuthService

Session Manager интегрирован с `AuthService`:

1. **Login** - создаёт сессию и устанавливает cookies
2. **Logout** - очищает сессию и удаляет cookies
3. **AuthenticateByToken** - использует `RestoreUserFromCookies`

## Примеры

См. `examples.go` для подробных примеров использования.

## Зависимости

- `github.com/gorilla/sessions` - управление сессиями
- `ctweb/internal/config` - конфигурация
- `ctweb/internal/models` - модели данных
- `ctweb/internal/repositories` - работа с БД

