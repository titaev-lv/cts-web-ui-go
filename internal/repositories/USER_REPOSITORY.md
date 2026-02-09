# UserRepository - Документация

## Описание

`UserRepository` предоставляет методы для работы с пользователями в базе данных. Все методы используют prepared statements для защиты от SQL injection.

## Основные методы

### FindByID(id int) (*models.User, error)

Находит пользователя по ID.

```go
repo := repositories.NewUserRepository()
user, err := repo.FindByID(1)
```

### FindByLogin(login string) (*models.User, error)

Находит пользователя по логину.

```go
user, err := repo.FindByLogin("admin")
```

### FindAll() ([]*models.User, error)

Возвращает всех пользователей (активных и неактивных), отсортированных по логину.

```go
users, err := repo.FindAll()
```

### FindAllActive() ([]*models.User, error)

Возвращает только активных пользователей, отсортированных по логину.

```go
activeUsers, err := repo.FindAllActive()
```

### Count() (int, error)

Возвращает общее количество пользователей в базе данных.

```go
count, err := repo.Count()
```

## Методы валидации

### ExistsByLogin(login string) (bool, error)

Проверяет, существует ли пользователь с указанным логином. Используется при создании нового пользователя.

```go
exists, err := repo.ExistsByLogin("admin")
if exists {
    return errors.ValidationError("Login already exists")
}
```

### ExistsByLoginExcludingID(login string, excludeID int) (bool, error)

Проверяет уникальность логина при обновлении пользователя.

```go
exists, err := repo.ExistsByLoginExcludingID("admin", 1)
if exists {
    return errors.ValidationError("Login already exists")
}
```

### ExistsByEmail(email string) (bool, error)

Проверяет, существует ли пользователь с указанным email.

```go
exists, err := repo.ExistsByEmail("user@example.com")
if exists {
    return errors.ValidationError("Email already exists")
}
```

### ExistsByEmailExcludingID(email string, excludeID int) (bool, error)

Проверяет уникальность email при обновлении пользователя.

```go
exists, err := repo.ExistsByEmailExcludingID("user@example.com", 1)
```

## Методы создания и обновления

### Create(user *models.User, groupIDs []int, userCreatedID int) (int, error)

Создаёт нового пользователя вместе с группами.

**ВАЖНО:** 
- Пароль должен быть уже захеширован (bcrypt) перед вызовом
- Использует транзакцию для атомарности операции

```go
hashedPassword, _ := utils.PasswordHash("MyPassword123")
user := &models.User{
    Login:    "newuser",
    Password: hashedPassword,
    Email:    "user@example.com",
    Name:     "John",
    LastName: "Doe",
    Active:   true,
    Timezone: "UTC",
}
userID, err := repo.Create(user, []int{2}, currentUserID)
```

### Update(user *models.User, groupIDs []int, updatePassword bool, userModifyID int) error

Обновляет существующего пользователя.

**Параметры:**
- `user` - данные пользователя (должен содержать ID)
- `groupIDs` - список ID групп (nil = не изменять, []int{} = удалить все группы)
- `updatePassword` - обновлять ли пароль (если true, пароль должен быть захеширован)
- `userModifyID` - ID пользователя, изменяющего запись

```go
user, _ := repo.FindByID(1)
user.Name = "Updated Name"
err := repo.Update(user, []int{1, 2}, false, currentUserID)
```

### SetUserGroups(tx *sql.Tx, userID int, groupIDs []int) error

Устанавливает группы для пользователя (удаляет старые и добавляет новые).

**ВАЖНО:** Должен вызываться внутри транзакции.

```go
tx, _ := db.BeginTransaction()
defer db.RollbackTransaction(tx)
err := repo.SetUserGroups(tx, userID, []int{1, 2})
```

## Методы для DataTables

### FindAllWithPagination(req *UserDataTablesRequest) (*UserDataTablesResponse, error)

Находит пользователей с поддержкой пагинации, сортировки и фильтрации для DataTables.

**Особенности:**
- Поддержка глобального поиска
- Поиск по отдельным колонкам
- Специальная обработка для имени (поиск по NAME и LAST_NAME)
- Поиск по группам (через JOIN)
- Форматирование дат
- Группировка по пользователю (GROUP BY) для корректного отображения групп

**Структура ответа:**
```go
type UserDataTablesResponse struct {
    Data            []*UserDataTablesRow `json:"data"`
    RecordsTotal    int                  `json:"recordsTotal"`
    RecordsFiltered int                  `json:"recordsFiltered"`
}

type UserDataTablesRow struct {
    ID          int    `json:"id"`
    Login       string `json:"login"`
    Groups      string `json:"groups"`      // "Admin, User"
    Active      string `json:"active"`       // "Active" или "Blocked"
    Name        string `json:"name"`         // "Doe John"
    Email       string `json:"email"`
    CreateDate  string `json:"create_date"`  // "01-12-2025 10:30:00"
    ModifyDate  string `json:"modify_date"`
    Timestamp   string `json:"timestamp"`
}
```

**Пример использования:**
```go
req := &repositories.UserDataTablesRequest{
    Start:  0,
    Length: 10,
    Search: "admin",
    Order: []struct{
        Column int
        Dir    string
    }{
        {Column: 0, Dir: "asc"},
    },
    Columns: []struct{
        Data       string
        Searchable bool
        Search     struct{ Value string }
    }{
        {Data: "name", Searchable: true, Search: struct{ Value string }{"John"}},
    },
}
response, err := repo.FindAllWithPagination(req)
```

## Методы для работы с группами

### FindGroupsByUserID(userID int) ([]int, error)

Находит все активные группы пользователя.

```go
groups, err := repo.FindGroupsByUserID(1)
// Результат: []int{1, 2}
```

## Методы для Remember Me

### FindByLoginAndToken(login, token string) (*models.User, error)

Находит пользователя по логину и токену (для восстановления сессии из cookie).

```go
user, err := repo.FindByLoginAndToken("admin", "token123")
```

### UpdateToken(userID int, token string) error

Обновляет токен пользователя (для Remember Me).

```go
err := repo.UpdateToken(userID, "new_token")
```

### UpdateTimestamp(userID int) error

Обновляет временную метку последней активности пользователя.

```go
err := repo.UpdateTimestamp(userID)
```

## Примеры использования

### Создание пользователя с валидацией

```go
func CreateUser(login, password, email, name, lastName string, groupIDs []int, userID int) error {
    repo := repositories.NewUserRepository()
    
    // Проверка уникальности логина
    exists, err := repo.ExistsByLogin(login)
    if err != nil {
        return err
    }
    if exists {
        return errors.ValidationError("Login already exists")
    }
    
    // Проверка уникальности email
    exists, err = repo.ExistsByEmail(email)
    if err != nil {
        return err
    }
    if exists {
        return errors.ValidationError("Email already exists")
    }
    
    // Хеширование пароля
    hashedPassword, err := utils.PasswordHash(password)
    if err != nil {
        return err
    }
    
    // Создание пользователя
    user := &models.User{
        Login:    login,
        Password: hashedPassword,
        Email:    email,
        Name:     name,
        LastName: lastName,
        Active:   true,
        Timezone: "UTC",
    }
    
    _, err = repo.Create(user, groupIDs, userID)
    return err
}
```

### Обновление пользователя с валидацией

```go
func UpdateUser(userID int, login, email, name, lastName string, groupIDs []int, currentUserID int) error {
    repo := repositories.NewUserRepository()
    
    // Получаем существующего пользователя
    user, err := repo.FindByID(userID)
    if err != nil {
        return err
    }
    
    // Проверка уникальности логина (если изменился)
    if user.Login != login {
        exists, err := repo.ExistsByLoginExcludingID(login, userID)
        if err != nil {
            return err
        }
        if exists {
            return errors.ValidationError("Login already exists")
        }
    }
    
    // Проверка уникальности email (если изменился)
    if user.Email != email {
        exists, err := repo.ExistsByEmailExcludingID(email, userID)
        if err != nil {
            return err
        }
        if exists {
            return errors.ValidationError("Email already exists")
        }
    }
    
    // Обновление данных
    user.Login = login
    user.Email = email
    user.Name = name
    user.LastName = lastName
    
    return repo.Update(user, groupIDs, false, currentUserID)
}
```

## Структура таблицы USER

```sql
CREATE TABLE USER (
    ID INT PRIMARY KEY AUTO_INCREMENT,
    LOGIN VARCHAR(255) NOT NULL UNIQUE,
    PASSWORD VARCHAR(255) NOT NULL,
    EMAIL VARCHAR(255) NOT NULL,
    ACTIVE TINYINT(1) NOT NULL DEFAULT 1,
    NAME VARCHAR(255) NOT NULL,
    LAST_NAME VARCHAR(255) NOT NULL,
    TOKEN VARCHAR(255) NULL,
    TIMEZONE VARCHAR(50) NOT NULL DEFAULT 'UTC',
    DATE_CREATE DATETIME NOT NULL,
    DATE_MODIFY DATETIME NULL,
    TIMESTAMP_X DATETIME NULL,
    USER_CREATED INT NULL,
    USER_MODIFY INT NULL
);
```

## Связь с группами

Пользователи связаны с группами через таблицу `USERS_GROUP`:
- `UID` - ID пользователя
- `GID` - ID группы

## Безопасность

- ✅ Все методы используют prepared statements
- ✅ Защита от SQL injection
- ✅ Транзакции для атомарности операций
- ✅ Валидация уникальности логина и email
- ✅ Пароли хешируются перед сохранением (bcrypt)
- ✅ Обработка nullable полей

## Следующие шаги

После реализации UserRepository можно переходить к:
- 3.3: DataTables API (server-side processing)
- 3.4: User Controller (List, Create, Edit, GetById)
- 3.6: Валидация (пароль, email, уникальность login)

