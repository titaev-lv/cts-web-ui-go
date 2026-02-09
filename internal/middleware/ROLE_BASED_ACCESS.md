# Role-based Access Control (Admin vs User)

Система контроля доступа на основе ролей (групп) пользователей.

## Обзор

Система использует группы пользователей для управления правами доступа:
- **Группа ID=1**: Администраторы (Admin)
- **Группа ID=2**: Пользователи (User) - обычные пользователи
- Другие группы: могут быть добавлены по необходимости

## Компоненты

### 1. AdminMiddleware

Middleware для защиты маршрутов, требующих прав администратора.

```go
admin := r.Group("/admin")
admin.Use(middleware.AdminMiddleware())
admin.GET("/users", adminController.ListUsers)
```

### 2. Helper функции

Функции для проверки прав в контроллерах:

- **`RequireAdmin(c)`** - проверка прав администратора
- **`RequireGroup(c, groupID)`** - проверка принадлежности к группе
- **`RequireAnyGroup(c, groupIDs)`** - проверка принадлежности хотя бы к одной группе

### 3. Методы модели User

- **`IsAdmin()`** - проверка, является ли пользователь администратором
- **`HasGroup(groupID)`** - проверка принадлежности к группе
- **`HasAnyGroup(groupIDs)`** - проверка принадлежности хотя бы к одной группе

## Использование

### Middleware для защиты маршрутов

```go
// В main.go
r := gin.Default()
r.Use(middleware.AuthMiddleware())  // Сначала проверка авторизации

// Администраторские маршруты
admin := r.Group("/admin")
admin.Use(middleware.AdminMiddleware())  // Проверка прав администратора
admin.GET("/users", adminController.ListUsers)
admin.POST("/users", adminController.CreateUser)
admin.DELETE("/users/:id", adminController.DeleteUser)
```

### Проверка прав в контроллерах

```go
func (c *gin.Context) {
    // Вариант 1: Использование helper функции
    admin, isAdmin := middleware.RequireAdmin(c)
    if !isAdmin {
        return // ошибка уже обработана
    }
    
    // Вариант 2: Ручная проверка
    user, exists := middleware.GetUserFromContext(c)
    if !exists || !user.IsAdmin() {
        c.JSON(403, gin.H{"error": "Admin access required"})
        return
    }
}
```

### Проверка принадлежности к группе

```go
func (c *gin.Context) {
    // Проверка принадлежности к группе модераторов (ID=2)
    user, hasGroup := middleware.RequireGroup(c, 2)
    if !hasGroup {
        return // ошибка уже обработана
    }
    
    // Пользователь принадлежит к группе
}
```

### Проверка принадлежности к любой из групп

```go
func (c *gin.Context) {
    // Проверка, что пользователь - администратор или модератор
    user, hasAnyGroup := middleware.RequireAnyGroup(c, []int{1, 2})
    if !hasAnyGroup {
        return // ошибка уже обработана
    }
    
    // Пользователь принадлежит хотя бы к одной группе
}
```

## Проверка активности группы

При проверке прав администратора также проверяется активность группы администраторов в БД:

```go
// В AdminMiddleware и RequireAdmin
if !isAdminGroupActive(user) {
    // Группа администраторов неактивна
    // Доступ запрещён
}
```

Это дополнительная проверка безопасности, которая гарантирует, что даже если пользователь принадлежит к группе администраторов, группа должна быть активна.

## Поведение при отказе в доступе

### HTML запросы

- **Редирект** на главную страницу `/` (HTTP 302)
- Запрос прерывается (`c.Abort()`)

### API запросы

- **JSON ошибка** (HTTP 403):
  ```json
  {"error": "Admin access required"}
  ```
- Запрос прерывается (`c.Abort()`)

## Логирование

Middleware логирует:

- **Попытки неавторизованного доступа** к администраторским маршрутам (warn уровень)
- **Попытки доступа при неактивной группе** (warn уровень)
- **Ошибки при проверке группы** (error уровень)

## Сравнение с PHP

### PHP код:
```php
// Проверка прав администратора
if($User->isAdmin()) {
    // Администратор
}

// В коде
if($User->isAdmin()) {
    // Показать администраторские функции
}
```

### Go код:
```go
// В middleware
admin.Use(middleware.AdminMiddleware())

// В контроллере
admin, isAdmin := middleware.RequireAdmin(c)
if !isAdmin {
    return
}

// Или
user, _ := middleware.GetUserFromContext(c)
if !user.IsAdmin() {
    c.JSON(403, gin.H{"error": "Admin access required"})
    return
}
```

## Примеры использования

### Защита администраторских маршрутов

```go
// В main.go
admin := r.Group("/admin")
admin.Use(middleware.AdminMiddleware())

admin.GET("/users", adminController.ListUsers)
admin.POST("/users", adminController.CreateUser)
admin.PUT("/users/:id", adminController.UpdateUser)
admin.DELETE("/users/:id", adminController.DeleteUser)
```

### Условный доступ в контроллере

```go
func (c *gin.Context) {
    user, _ := middleware.GetUserFromContext(c)
    
    if user.IsAdmin() {
        // Администратор - показываем все данные
        users := getAllUsers()
        c.JSON(200, gin.H{"users": users})
    } else {
        // Обычный пользователь - показываем только свои данные
        c.JSON(200, gin.H{"user": user})
    }
}
```

### Условный рендеринг в шаблонах

```go
// В контроллере
func (c *gin.Context) {
    user, _ := middleware.GetUserFromContext(c)
    isAdmin := user != nil && user.IsAdmin()
    
    c.HTML(200, "index.html", gin.H{
        "User":    user,
        "IsAdmin": isAdmin,
    })
}
```

```html
<!-- В шаблоне -->
{{if .IsAdmin}}
    <a href="/admin/users">Manage Users</a>
    <a href="/admin/groups">Manage Groups</a>
{{end}}
```

## Безопасность

### Многоуровневая проверка

1. **Проверка авторизации** - AuthMiddleware проверяет, что пользователь авторизован
2. **Проверка группы** - AdminMiddleware проверяет принадлежность к группе администраторов
3. **Проверка активности группы** - проверяется, что группа администраторов активна в БД

### Логирование

Все попытки несанкционированного доступа логируются для аудита безопасности.

## Зависимости

- `ctweb/internal/models` - модели User и Group
- `ctweb/internal/repositories` - GroupRepository для проверки активности группы
- `ctweb/internal/middleware` - AuthMiddleware (должен быть применён первым)

## См. также

- `internal/middleware/auth.go` - AuthMiddleware для проверки авторизации
- `internal/models/user.go` - методы IsAdmin(), HasGroup()
- `internal/middleware/admin_examples.go` - примеры использования

