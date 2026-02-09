# User Model - Модель пользователя

**Дата:** 14 декабря 2024  
**Задача:** 2.1 - User Model (структура с группами)

## Обзор

Модель `User` представляет пользователя системы с поддержкой групп пользователей. Структура соответствует таблице `USER` в базе данных MySQL.

## Структура User

### Поля модели

| Поле | Тип | JSON | DB | Описание |
|------|-----|------|----|----------|
| ID | int | `id` | `ID` | Уникальный идентификатор |
| Login | string | `login` | `LOGIN` | Логин (уникальный) |
| Password | string | `-` | `PASSWORD` | Хеш пароля (не отправляется) |
| Email | string | `email` | `EMAIL` | Email адрес |
| Active | bool | `active` | `ACTIVE` | Активен ли пользователь |
| Name | string | `name` | `NAME` | Имя |
| LastName | string | `last_name` | `LAST_NAME` | Фамилия |
| Token | string | `-` | `TOKEN` | Токен для "Remember Me" |
| Timezone | string | `timezone` | `TIMEZONE` | Временная зона |
| DateCreate | time.Time | `date_create` | `DATE_CREATE` | Дата создания |
| DateModify | *time.Time | `date_modify` | `DATE_MODIFY` | Дата изменения |
| TimestampX | *time.Time | `timestamp_x` | `TIMESTAMP_X` | Временная метка активности |
| UserCreated | *int | `user_created` | `USER_CREATED` | ID создателя |
| UserModify | *int | `user_modify` | `USER_MODIFY` | ID изменившего |
| Groups | []int | `groups` | `-` | Список ID групп |

### Особенности

- **Password** и **Token** не отправляются в JSON (тег `json:"-"`)
- **Groups** загружается отдельно через JOIN с таблицей `USERS_GROUP`
- Поля с `*` могут быть `nil` (опциональные)

## Методы User

### GetFullName() string

Возвращает полное имя пользователя (Фамилия + Имя).

```go
fullName := user.GetFullName()
// Результат: "Иванов Иван"
```

### IsActive() bool

Проверяет, активен ли пользователь.

```go
if !user.IsActive() {
    return errors.UnauthorizedError("User is blocked")
}
```

### HasGroup(groupID int) bool

Проверяет, принадлежит ли пользователь к указанной группе.

```go
if user.HasGroup(1) {
    // Пользователь - администратор
}
```

### IsAdmin() bool

Проверяет, является ли пользователь администратором.

Группа с ID=1 - это администраторы.

```go
if !user.IsAdmin() {
    return errors.ForbiddenError("Admin access required")
}
```

### HasAnyGroup(groupIDs []int) bool

Проверяет, принадлежит ли пользователь хотя бы к одной из указанных групп.

```go
if user.HasAnyGroup([]int{1, 2}) {
    // Пользователь - администратор или модератор
}
```

### HasAllGroups(groupIDs []int) bool

Проверяет, принадлежит ли пользователь ко всем указанным группам.

```go
if user.HasAllGroups([]int{1, 2}) {
    // Пользователь - администратор И модератор
}
```

### SetGroups(groupIDs []int)

Устанавливает список групп пользователя.

```go
user.SetGroups([]int{1, 2})
```

### AddGroup(groupID int) bool

Добавляет пользователя в группу (если ещё не состоит).

```go
added := user.AddGroup(2)
```

### RemoveGroup(groupID int) bool

Удаляет пользователя из группы.

```go
removed := user.RemoveGroup(2)
```

### ClearGroups()

Удаляет все группы пользователя.

```go
user.ClearGroups()
```

### GetGroupsCount() int

Возвращает количество групп пользователя.

```go
count := user.GetGroupsCount()
```

## Структура Group

### Поля модели

| Поле | Тип | JSON | DB | Описание |
|------|-----|------|----|----------|
| ID | int | `id` | `ID` | Уникальный идентификатор |
| Name | string | `name` | `NAME` | Название группы |
| Description | string | `description` | `DESCRIPTION` | Описание |
| Active | bool | `active` | `ACTIVE` | Активна ли группа |
| DateCreate | time.Time | `date_create` | `DATE_CREATE` | Дата создания |
| DateModify | *time.Time | `date_modify` | `DATE_MODIFY` | Дата изменения |
| UserCreated | *int | `user_created` | `USER_CREATED` | ID создателя |
| UserModify | *int | `user_modify` | `USER_MODIFY` | ID изменившего |

### Методы Group

### IsActive() bool

Проверяет, активна ли группа.

```go
if !group.IsActive() {
    return errors.NotFoundError("group", groupID)
}
```

### IsAdminGroup() bool

Проверяет, является ли группа группой администраторов (ID=1).

```go
if group.IsAdminGroup() {
    // Это группа администраторов
}
```

## Связь с группами

Пользователи связаны с группами через таблицу `USERS_GROUP`:

```
USER (1) <---> (N) USERS_GROUP (N) <---> (1) GROUP
```

**Таблица USERS_GROUP:**
- `UID` - ID пользователя
- `GID` - ID группы

## Специальные группы

- **ID=1**: Администраторы (Admin)
- **ID=2**: Пользователи (User) - обычные пользователи

## Примеры использования

### Создание пользователя

```go
user := models.User{
    Login:    "admin",
    Email:    "admin@example.com",
    Active:   true,
    Name:     "Admin",
    LastName: "User",
    Timezone: "UTC",
}
```

### Проверка прав доступа

```go
// Проверка администратора
if !user.IsAdmin() {
    return errors.ForbiddenError("Admin access required")
}

// Проверка конкретной группы
if !user.HasGroup(2) {
    return errors.ForbiddenError("User group required")
}

// Проверка нескольких групп
if !user.HasAnyGroup([]int{1, 2}) {
    return errors.ForbiddenError("Access denied")
}
```

### Работа с группами

```go
// Добавить группу
user.AddGroup(2)

// Удалить группу
user.RemoveGroup(2)

// Установить группы
user.SetGroups([]int{1, 2})

// Очистить группы
user.ClearGroups()
```

## Соответствие с PHP

| PHP | Go |
|-----|-----|
| `$this->user_uid` | `user.ID` |
| `$this->user_name` | `user.GetFullName()` |
| `$this->user_email` | `user.Email` |
| `$this->user_grp` | `user.Groups` |
| `in_array("1", $this->user_grp)` | `user.IsAdmin()` |
| `$this->auth` | (будет в сессии) |

## Следующие шаги

- Задача 2.2: Auth Service (Login, Logout, Remember Me)
- Задача 2.3: Password Hashing (bcrypt)
- Задача 2.4: Session Management

---

*Модель User готова к использованию!*

