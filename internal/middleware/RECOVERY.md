# Recovery Middleware - Обработка паник

**Дата:** 14 декабря 2024

## Обзор

Recovery Middleware перехватывает паники (panic) и предотвращает краш приложения, возвращая пользователю корректный ответ вместо падения сервера.

## Что такое panic?

**Panic** в Go - это критическая ошибка, которая останавливает выполнение программы.

**Примеры ситуаций, когда может возникнуть panic:**
- Обращение к nil указателю
- Выход за границы массива/слайса
- Вызов `panic()` вручную
- Ошибки в сторонних библиотеках

**Без Recovery middleware:**
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x...]
```

**С Recovery middleware:**
```json
{
  "error": "An unexpected error occurred",
  "success": false
}
```

## Как работает

1. **Перехват паники** - `recover()` перехватывает panic
2. **Логирование** - Полная информация о панике записывается в лог
3. **Ответ пользователю** - Возвращается 500 Internal Server Error
4. **Продолжение работы** - Приложение продолжает работать

## Использование

### Базовое использование

```go
r := gin.New() // Без встроенного Recovery
r.Use(middleware.RecoveryMiddleware())
```

### С опцией показа stack trace (только для debug)

```go
cfg := config.Get()
showStack := cfg.Server.Mode == "debug"
r.Use(middleware.RecoveryMiddlewareWithStack(showStack))
```

## Что логируется

При возникновении паники логируется:

- **panic** - значение, переданное в panic()
- **stack** - полный stack trace (где произошла паника)
- **path** - путь запроса
- **method** - HTTP метод
- **client_ip** - IP адрес клиента
- **user_agent** - User-Agent браузера

**Пример лога:**
```
2024-12-14T18:30:45+03:00 ERR Panic recovered panic="runtime error: invalid memory address" path=/api/users method=POST client_ip=192.168.1.1
```

## Безопасность

### ✅ В продакшн

- Детали паники **НЕ показываются** пользователю
- Stack trace **только в логах**
- Пользователь видит общее сообщение: "An unexpected error occurred"

### ⚠️ В режиме разработки (debug)

- Можно показать stack trace в ответе (для отладки)
- **ВАЖНО:** Никогда не используйте это в продакшн!

## Примеры

### Пример 1: Обращение к nil указателю

```go
func GetUser(c *gin.Context) {
    var user *User = nil
    // Паника: обращение к nil указателю
    c.JSON(200, gin.H{"name": user.Name})
}
```

**Без Recovery:** Приложение крашится  
**С Recovery:** Возвращается 500 ошибка, приложение продолжает работать

### Пример 2: Выход за границы массива

```go
func GetItem(c *gin.Context) {
    items := []string{"a", "b"}
    // Паника: индекс 10 не существует
    item := items[10]
    c.JSON(200, gin.H{"item": item})
}
```

**Без Recovery:** Приложение крашится  
**С Recovery:** Возвращается 500 ошибка, приложение продолжает работать

### Пример 3: Ручной вызов panic

```go
func ProcessData(c *gin.Context) {
    if someCondition {
        panic("Critical error occurred")
    }
}
```

**Без Recovery:** Приложение крашится  
**С Recovery:** Возвращается 500 ошибка, приложение продолжает работать

## Порядок middleware

Recovery middleware должен быть **первым** в цепочке:

```go
r := gin.New()

// 1. Recovery (первым!)
r.Use(middleware.RecoveryMiddleware())

// 2. Security Headers
r.Use(middleware.SecurityHeadersMiddleware())

// 3. Auth
r.Use(middleware.AuthMiddleware())

// 4. Остальные middleware...
```

**Почему первым?**
- Recovery должен перехватывать паники из всех последующих middleware
- Если Recovery будет после других middleware, паники из них не будут перехвачены

## Интеграция с системой обработки ошибок

Recovery middleware использует нашу систему обработки ошибок:

```go
appErr := errors.InternalError("An unexpected error occurred", panicErr)
errors.HandleError(c, appErr)
```

Это обеспечивает:
- Единый формат ответов
- Автоматическое логирование
- Правильные HTTP статус коды

## Сравнение с PHP

| PHP | Go |
|-----|-----|
| `try { ... } catch (Exception $e) { ... }` | `defer { recover() }` |
| `exit()` | `panic()` |
| Обработка через try-catch | Обработка через Recovery middleware |

## Лучшие практики

### ✅ ДЕЛАТЬ:

1. **Всегда использовать Recovery middleware**
   ```go
   r.Use(middleware.RecoveryMiddleware())
   ```

2. **Размещать Recovery первым**
   ```go
   r.Use(middleware.RecoveryMiddleware()) // Первым!
   ```

3. **Логировать все паники**
   - Recovery автоматически логирует все паники

### ❌ НЕ ДЕЛАТЬ:

1. **Не показывать stack trace в продакшн**
   ```go
   // ПЛОХО (в продакшн)
   showStack := true
   
   // ХОРОШО
   showStack := cfg.Server.Mode == "debug"
   ```

2. **Не игнорировать паники**
   - Всегда используйте Recovery middleware
   - Не полагайтесь на то, что паник не будет

3. **Не вызывать panic() без необходимости**
   ```go
   // ПЛОХО
   if err != nil {
       panic(err)
   }
   
   // ХОРОШО
   if err != nil {
       errors.HandleError(c, errors.InternalError("Failed", err))
       return
   }
   ```

## Отладка паник

### 1. Проверьте логи

Все паники логируются с полным stack trace:
```
ERR Panic recovered panic="..." stack="goroutine 1 [running]:\n..."
```

### 2. Используйте debug режим

В режиме разработки можно увидеть stack trace в ответе:
```json
{
  "error": "An unexpected error occurred",
  "success": false,
  "details": {
    "panic": "...",
    "stack": "..."
  }
}
```

### 3. Проверьте stack trace

Stack trace показывает:
- Где произошла паника (файл и строка)
- Цепочку вызовов функций
- Параметры функций

## Дополнительные ресурсы

- [Go Panic and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Gin Recovery Middleware](https://gin-gonic.com/docs/middleware/recovery/)

---

*Recovery middleware защищает приложение от неожиданных паник!*

