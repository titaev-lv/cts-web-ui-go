# Auth Middleware - Проверка сессий и редирект

Middleware для проверки аутентификации пользователей через сессии и cookies "Remember Me".

## Обзор

AuthMiddleware проверяет:
1. Является ли маршрут публичным (не требует авторизации)
2. Наличие активной сессии пользователя
3. Восстановление из cookies "Remember Me" (если сессии нет)
4. Активность пользователя и наличие групп
5. Редирект на `/login` для неавторизованных пользователей

## Использование

### Регистрация в main.go

```go
import "ctweb/internal/middleware"

r := gin.New()
r.Use(middleware.AuthMiddleware())
```

### Получение пользователя в контроллерах

```go
import "ctweb/internal/middleware"

func (c *gin.Context) {
    user, exists := middleware.GetUserFromContext(c)
    if !exists {
        // Пользователь не найден (не должно произойти для защищённых маршрутов)
        c.JSON(500, gin.H{"error": "User not found"})
        return
    }

    // Используем данные пользователя
    c.JSON(200, gin.H{
        "user_id": user.ID,
        "name":    user.GetFullName(),
    })
}
```

## Публичные маршруты

Следующие маршруты **не требуют** авторизации:

- `/` - главная страница
- `/login` - страница входа
- `/auth/login` - обработка формы входа
- `/assets/*` - статические файлы

Все остальные маршруты требуют авторизации.

## Поток проверки

```
1. Запрос приходит в AuthMiddleware
   │
   ├─ Маршрут публичный?
   │  └─ Да → Продолжить обработку
   │
   └─ Нет → Проверка авторизации
      │
      ├─ Получить пользователя из сессии
      │  └─ Успешно? → Продолжить
      │
      ├─ Сессии нет → Восстановить из cookies "Remember Me"
      │  └─ Успешно? → Продолжить
      │
      ├─ Пользователь не активен?
      │  └─ Да → Очистить сессию → Редирект на /login
      │
      ├─ Нет активных групп?
      │  └─ Да → Очистить сессию → Редирект на /login
      │
      └─ Пользователь авторизован → Сохранить в контекст → Продолжить
```

## Поведение для разных типов запросов

### HTML запросы

Если пользователь не авторизован:
- **Редирект** на `/login` (HTTP 302)
- Запрос прерывается (`c.Abort()`)

### API запросы

Если пользователь не авторизован:
- **JSON ответ** с ошибкой (HTTP 401):
  ```json
  {"error": "Unauthorized"}
  ```
- Запрос прерывается (`c.Abort()`)

API запросы определяются по:
- Заголовок `Accept` содержит `application/json`
- Или путь начинается с `/api/`

## Проверки безопасности

### 1. Активность пользователя

Если пользователь заблокирован (`Active = false`):
- Сессия очищается
- Редирект на `/login` или JSON ошибка 403

### 2. Активные группы

Если у пользователя нет активных групп:
- Сессия очищается
- Редирект на `/login` или JSON ошибка 403

### 3. Восстановление из cookies

При восстановлении из cookies "Remember Me":
- Проверяется токен в БД
- Проверяется активность пользователя
- Проверяется наличие активных групп
- Создаётся новая сессия

## Контекст Gin

После успешной авторизации пользователь сохраняется в контекст Gin:

```go
c.Set(middleware.ContextKeyUser, user)
```

Получение пользователя:

```go
user, exists := middleware.GetUserFromContext(c)
```

## Примеры использования

### Публичный маршрут

```go
// В main.go
r.GET("/", userController.Home)

// В контроллере
func (c *gin.Context) {
    // Пользователь может быть не авторизован
    user, exists := middleware.GetUserFromContext(c)
    if exists {
        // Пользователь авторизован
        c.HTML(200, "index.html", gin.H{"User": user})
    } else {
        // Пользователь не авторизован
        c.HTML(200, "index.html", gin.H{"User": nil})
    }
}
```

### Защищённый маршрут

```go
// В main.go
r.GET("/users", userController.ListUsers)

// В контроллере
func (c *gin.Context) {
    // Пользователь всегда авторизован (middleware проверяет)
    user, _ := middleware.GetUserFromContext(c)
    
    c.JSON(200, gin.H{
        "current_user": user.GetFullName(),
        "users":        getUsersList(),
    })
}
```

### Проверка прав администратора

```go
func (c *gin.Context) {
    user, _ := middleware.GetUserFromContext(c)
    
    if !user.IsAdmin() {
        c.JSON(403, gin.H{"error": "Forbidden: Admin access required"})
        return
    }
    
    // Пользователь - администратор
    c.JSON(200, gin.H{"message": "Admin access granted"})
}
```

## Логирование

Middleware логирует:

- **Ошибки** при получении сессии или восстановлении из cookies
- **Попытки неавторизованного доступа** (debug уровень)
- **Восстановление из cookies** (debug уровень)
- **Попытки доступа заблокированных пользователей** (warn уровень)
- **Попытки доступа пользователей без групп** (warn уровень)

## Сравнение с PHP

### PHP код:
```php
// В prolog.php
if(defined('AUTH')) {
    if (!$User->checkAuth()) {
        header('HTTP/1.1 401 Unauthorized');
        include_once $_SERVER['DOCUMENT_ROOT'].'/sign-in.html';
        exit();
    }
}
```

### Go код:
```go
// В middleware
r.Use(middleware.AuthMiddleware())

// Middleware автоматически:
// - Проверяет сессию
// - Восстанавливает из cookies
// - Редиректит на /login
```

## Зависимости

- `ctweb/internal/session` - управление сессиями
- `ctweb/internal/models` - модели данных
- `ctweb/internal/logger` - логирование

## См. также

- `internal/session/SESSION.md` - документация по сессиям
- `internal/middleware/auth_examples.go` - примеры использования

