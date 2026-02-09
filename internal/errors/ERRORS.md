# Централизованная обработка ошибок

**Дата:** 14 декабря 2024

## Обзор

Централизованная система обработки ошибок обеспечивает:
- ✅ Единый формат ответов
- ✅ Правильные HTTP статус коды
- ✅ Автоматическое логирование
- ✅ Совместимость с PHP форматом ответов
- ✅ Безопасность (внутренние ошибки не показываются пользователю)

## Типы ошибок

### ValidationError (400 Bad Request)

**Когда использовать:** Невалидный пользовательский ввод

```go
if username == "" {
    errors.HandleError(c, errors.ValidationError("Username is required", nil))
    return
}
```

**Ответ:**
```json
{
  "error": "Username is required",
  "success": false
}
```

### NotFoundError (404 Not Found)

**Когда использовать:** Ресурс не найден

```go
user, err := getUserByID(userID)
if err == sql.ErrNoRows {
    errors.HandleError(c, errors.NotFoundError("user", userID))
    return
}
```

**Ответ:**
```json
{
  "error": "User with ID 123 not found",
  "success": false,
  "details": {
    "resource": "user",
    "id": 123
  }
}
```

### UnauthorizedError (401 Unauthorized)

**Когда использовать:** Пользователь не авторизован или неверные креды

```go
if !checkPassword(password, user.Password) {
    errors.HandleError(c, errors.UnauthorizedError("Invalid credentials"))
    return
}
```

**Ответ:**
```json
{
  "error": "Invalid credentials",
  "success": false
}
```

### ForbiddenError (403 Forbidden)

**Когда использовать:** Пользователь авторизован, но нет прав доступа

```go
if !user.IsAdmin() {
    errors.HandleError(c, errors.ForbiddenError("Admin access required"))
    return
}
```

**Ответ:**
```json
{
  "error": "Admin access required",
  "success": false
}
```

### ConflictError (409 Conflict)

**Когда использовать:** Конфликт с текущим состоянием (дубликат)

```go
if userExists(login) {
    errors.HandleError(c, errors.ConflictError("User already exists"))
    return
}
```

**Ответ:**
```json
{
  "error": "User already exists",
  "success": false
}
```

### InternalError (500 Internal Server Error)

**Когда использовать:** Неожиданная внутренняя ошибка

```go
result, err := db.Exec("INSERT INTO users ...")
if err != nil {
    errors.HandleError(c, errors.InternalError("Failed to create user", err))
    return
}
```

**Ответ:**
```json
{
  "error": "Failed to create user",
  "success": false
}
```

**ВАЖНО:** Детали внутренней ошибки логируются, но не показываются пользователю!

## Обработка ошибок БД

### DatabaseError

Автоматически обрабатывает ошибки базы данных:

```go
user, err := getUserByID(userID)
if err != nil {
    // Автоматически преобразует:
    // - sql.ErrNoRows -> NotFoundError
    // - Другие ошибки -> InternalError
    errors.HandleDatabaseError(c, "user", userID, err)
    return
}
```

## Успешные ответы

### HandleSuccess

Отправляет успешный ответ в формате, совместимом с PHP:

```go
errors.HandleSuccess(c, gin.H{
    "user_id": user.ID,
    "username": user.Username,
})
```

**Ответ:**
```json
{
  "error": false,
  "success": true,
  "data": {
    "user_id": 123,
    "username": "user123"
  }
}
```

## Множественные ошибки валидации

### HandleValidationErrors

Обрабатывает несколько ошибок валидации одновременно:

```go
validationErrors := map[string]string{
    "username": "Username is required",
    "email":    "Invalid email format",
    "password": "Password too short",
}

if len(validationErrors) > 0 {
    errors.HandleValidationErrors(c, validationErrors)
    return
}
```

**Ответ:**
```json
{
  "error": "Validation failed",
  "success": false,
  "details": {
    "username": "Username is required",
    "email": "Invalid email format",
    "password": "Password too short"
  }
}
```

## Логирование

Ошибки автоматически логируются:

- **400-499 (Client Errors):** Логируются на уровне `Warn`
- **500+ (Server Errors):** Логируются на уровне `Error` с полными деталями

**Пример лога:**
```
2024-12-14T18:30:45+03:00 WRN Client error status_code=404 message="User not found" path=/api/users/123 method=GET client_ip=192.168.1.1
```

## Формат ответов (совместимость с PHP)

### PHP формат:
```php
$data = array(
    "error"   => $error,
    "success" => $success,
);
```

### Go формат (ошибка):
```json
{
  "error": "Error message",
  "success": false
}
```

### Go формат (успех):
```json
{
  "error": false,
  "success": true,
  "data": {...}
}
```

## Полный пример

```go
func CreateUser(c *gin.Context) {
    // 1. Валидация
    username := c.PostForm("username")
    if username == "" {
        errors.HandleError(c, errors.ValidationError("Username is required", nil))
        return
    }

    // 2. Проверка на дубликат
    if userExists(username) {
        errors.HandleError(c, errors.ConflictError("User already exists"))
        return
    }

    // 3. Создание пользователя
    user, err := createUserInDB(username)
    if err != nil {
        errors.HandleDatabaseError(c, "user", nil, err)
        return
    }

    // 4. Успешный ответ
    errors.HandleSuccess(c, gin.H{
        "user_id": user.ID,
        "username": user.Username,
    })
}
```

## Лучшие практики

### ✅ ДЕЛАТЬ:

1. **Всегда использовать HandleError**
   ```go
   if err != nil {
       errors.HandleError(c, err)
       return
   }
   ```

2. **Использовать правильный тип ошибки**
   ```go
   // ПЛОХО
   errors.HandleError(c, errors.InternalError("User not found", nil))
   
   // ХОРОШО
   errors.HandleError(c, errors.NotFoundError("user", userID))
   ```

3. **Логировать внутренние ошибки**
   ```go
   // Внутренние ошибки автоматически логируются
   errors.HandleError(c, errors.InternalError("Failed to process", err))
   ```

### ❌ НЕ ДЕЛАТЬ:

1. **Не показывать детали внутренних ошибок**
   ```go
   // ПЛОХО
   errors.HandleError(c, errors.InternalError(err.Error(), err))
   
   // ХОРОШО
   errors.HandleError(c, errors.InternalError("Failed to process", err))
   ```

2. **Не использовать неправильные статус коды**
   ```go
   // ПЛОХО
   c.JSON(500, gin.H{"error": "User not found"})
   
   // ХОРОШО
   errors.HandleError(c, errors.NotFoundError("user", userID))
   ```

## Сравнение с PHP

| PHP | Go |
|-----|-----|
| `$error = "message"` | `errors.ValidationError("message", nil)` |
| `header('HTTP/1.1 401')` | `errors.UnauthorizedError("message")` |
| `echo json_encode(["error" => $error])` | `errors.HandleError(c, err)` |

---

*Единообразная обработка ошибок упрощает поддержку и отладку!*

