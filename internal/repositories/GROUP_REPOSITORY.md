# GroupRepository - Документация

## Описание

`GroupRepository` предоставляет методы для работы с группами пользователей в базе данных. Все методы используют prepared statements для защиты от SQL injection.

## Основные методы

### FindByID(id int) (*models.Group, error)

Находит группу по ID.

```go
repo := repositories.NewGroupRepository()
group, err := repo.FindByID(1)
if err != nil {
    // Обработка ошибки
}
```

### FindAll() ([]*models.Group, error)

Возвращает все группы (активные и неактивные), отсортированные по имени.

```go
groups, err := repo.FindAll()
```

### FindAllActive() ([]*models.Group, error)

Возвращает только активные группы, отсортированные по имени.

```go
activeGroups, err := repo.FindAllActive()
```

### FindByName(name string) (*models.Group, error)

Находит группу по имени. Используется для проверки уникальности.

```go
group, err := repo.FindByName("Administrators")
if err != nil {
    // Группа не найдена
}
```

### Count() (int, error)

Возвращает общее количество групп в базе данных.

```go
count, err := repo.Count()
```

### Create(group *models.Group, userCreatedID int) (int, error)

Создаёт новую группу. Возвращает ID созданной группы.

**ВАЖНО:** Использует транзакцию для атомарности операции.

```go
group := &models.Group{
    Name:        "Moderators",
    Description: "Moderator group",
    Active:      true,
}
groupID, err := repo.Create(group, currentUserID)
if err != nil {
    // Обработка ошибки
}
```

### Update(group *models.Group, userModifyID int) error

Обновляет существующую группу. Группа должна содержать ID.

**ВАЖНО:** Использует транзакцию для атомарности операции.

```go
group, _ := repo.FindByID(1)
group.Name = "Updated Name"
group.Active = false
err := repo.Update(group, currentUserID)
```

### FindAllWithPagination(req *DataTablesRequest) (*DataTablesResponse, error)

Находит группы с поддержкой пагинации, сортировки и фильтрации для DataTables.

**Параметры запроса:**
- `Start` - начальная позиция (offset)
- `Length` - количество записей (limit)
- `Search` - глобальный поисковый запрос
- `Order` - параметры сортировки
- `Columns` - параметры фильтрации по колонкам

**Возвращает:**
- `Data` - список групп
- `RecordsTotal` - общее количество записей
- `RecordsFiltered` - количество записей после фильтрации

```go
req := &repositories.DataTablesRequest{
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
        {Data: "name", Searchable: true, Search: struct{ Value string }{"admin"}},
    },
}
response, err := repo.FindAllWithPagination(req)
```

## Методы валидации

### ExistsByName(name string) (bool, error)

Проверяет, существует ли группа с указанным именем. Используется при создании новой группы.

```go
exists, err := repo.ExistsByName("Administrators")
if exists {
    return errors.ValidationError("Group name already exists")
}
```

### ExistsByNameExcludingID(name string, excludeID int) (bool, error)

Проверяет, существует ли группа с указанным именем, исключая группу с указанным ID. Используется при обновлении группы.

```go
exists, err := repo.ExistsByNameExcludingID("Administrators", 1)
if exists {
    return errors.ValidationError("Group name already exists")
}
```

## Структура таблицы GROUP

```sql
CREATE TABLE `GROUP` (
    ID INT PRIMARY KEY AUTO_INCREMENT,
    NAME VARCHAR(255) NOT NULL UNIQUE,
    DESCRIPTION TEXT,
    ACTIVE TINYINT(1) NOT NULL DEFAULT 1,
    DATE_CREATE DATETIME NOT NULL,
    DATE_MODIFY DATETIME NULL,
    USER_CREATED INT NULL,
    USER_MODIFY INT NULL
);
```

## Специальные группы

- **ID=1**: Администраторы (Admin) - системная группа
- **ID=2**: Пользователи (User) - обычные пользователи

## Примеры использования

### Создание группы с валидацией

```go
func CreateGroup(name, description string, active bool, userID int) error {
    repo := repositories.NewGroupRepository()
    
    // Проверка уникальности имени
    exists, err := repo.ExistsByName(name)
    if err != nil {
        return err
    }
    if exists {
        return errors.ValidationError("Group name already exists")
    }
    
    // Создание группы
    group := &models.Group{
        Name:        name,
        Description: description,
        Active:      active,
    }
    
    _, err = repo.Create(group, userID)
    return err
}
```

### Обновление группы с валидацией

```go
func UpdateGroup(groupID int, name, description string, active bool, userID int) error {
    repo := repositories.NewGroupRepository()
    
    // Получаем существующую группу
    group, err := repo.FindByID(groupID)
    if err != nil {
        return err
    }
    
    // Проверка уникальности имени (исключая текущую группу)
    if group.Name != name {
        exists, err := repo.ExistsByNameExcludingID(name, groupID)
        if err != nil {
            return err
        }
        if exists {
            return errors.ValidationError("Group name already exists")
        }
    }
    
    // Обновление данных
    group.Name = name
    group.Description = description
    group.Active = active
    
    return repo.Update(group, userID)
}
```

### Получение списка групп для DataTables

```go
func GetGroupsForDataTables(c *gin.Context) {
    repo := repositories.NewGroupRepository()
    
    // Парсим параметры от DataTables
    req := &repositories.DataTablesRequest{
        Start:  c.GetInt("start"),
        Length: c.GetInt("length"),
        Search: c.GetString("search[value]"),
        // ... остальные параметры
    }
    
    response, err := repo.FindAllWithPagination(req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, response)
}
```

## Безопасность

- ✅ Все методы используют prepared statements
- ✅ Защита от SQL injection
- ✅ Транзакции для атомарности операций
- ✅ Валидация уникальности имени
- ✅ Обработка nullable полей

## Обработка ошибок

Все методы возвращают ошибки, которые нужно обрабатывать:

```go
group, err := repo.FindByID(1)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // Группа не найдена
        return errors.NotFoundError("group", 1)
    }
    // Другая ошибка БД
    return errors.InternalError("Failed to find group", err)
}
```

## Следующие шаги

После реализации GroupRepository можно переходить к:
- 3.2: User Model и Repository (если ещё не реализовано)
- 3.3: DataTables API (server-side processing)
- 3.4: Group Controller (List, Create, Edit, GetById)

