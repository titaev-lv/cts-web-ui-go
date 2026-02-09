# GroupController - Документация

## Описание

`GroupController` предоставляет HTTP handlers для управления группами пользователей. Все методы требуют авторизации, а операции создания/редактирования - прав администратора.

## Методы

### List() - Страница управления группами

**Маршрут:** `GET /groups/`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Что делает:**
- Отображает страницу управления группами
- Пока использует шаблон `index.html` (в фазе 3.7 будет создан `groups/index.html`)

**Пример использования:**
```go
groups.GET("/", groupController.List)
```

### AjaxGetGroups() - Получение списка групп для DataTables

**Маршрут:** `POST /groups/ajax_get_groups`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры:** DataTables параметры (см. DATATABLES_API.md)

**Формат ответа:**
```json
{
  "draw": 1,
  "recordsTotal": 10,
  "recordsFiltered": 5,
  "aaData": [...]
}
```

### AjaxGetGroupById() - Получение группы по ID

**Маршрут:** `GET /groups/ajax_getid_group?id=1` или `POST /groups/ajax_getid_group`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры:**
- `id` - ID группы (query или form)

**Формат ответа:**
```json
{
  "error": false,
  "success": true,
  "data": {
    "id": "1",
    "name": "Administrators",
    "description": "System administrators",
    "status": "enable"
  }
}
```

**Пример использования (JavaScript):**
```javascript
$.post('/groups/ajax_getid_group', {id: 1}, function(response) {
    if (response.success) {
        // Использовать response.data
    }
});
```

### AjaxCreateGroup() - Создание группы

**Маршрут:** `POST /groups/ajax_create_group`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры (POST form data):**
- `create_group_name` - название группы (обязательно)
- `create_group_description` - описание группы (опционально)
- `create_group_status` - статус: "enable" или "disable" (обязательно)

**Валидация:**
- ✅ Проверка заполненности обязательных полей (name, status)
- ✅ Валидация статуса ("enable" или "disable")
- ✅ Проверка уникальности имени группы

**Формат ответа (успех):**
```json
{
  "error": false,
  "success": true
}
```

**Формат ответа (ошибка):**
```json
{
  "error": "Error message",
  "success": false
}
```

**Пример использования (JavaScript):**
```javascript
var formData = {
    create_group_name: "Moderators",
    create_group_description: "Moderator group",
    create_group_status: "enable"
};

$.post('/groups/ajax_create_group', formData, function(response) {
    if (response.success) {
        // Группа создана
    } else {
        // Ошибка: response.error
    }
});
```

### AjaxEditGroup() - Редактирование группы

**Маршрут:** `POST /groups/ajax_edit_group`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры (POST form data):**
- `edit_group_id` - ID группы (обязательно)
- `edit_group_name` - название группы (обязательно)
- `edit_group_description` - описание группы (опционально)
- `edit_group_status` - статус: "enable" или "disable" (обязательно)

**Валидация:**
- ✅ Проверка заполненности обязательных полей (id, name, status)
- ✅ Валидация статуса ("enable" или "disable")
- ✅ Проверка уникальности имени группы (исключая текущую группу)

**Формат ответа:** Аналогично `AjaxCreateGroup`

**Пример использования (JavaScript):**
```javascript
var formData = {
    edit_group_id: "1",
    edit_group_name: "Updated Group",
    edit_group_description: "Updated description",
    edit_group_status: "enable"
};

$.post('/groups/ajax_edit_group', formData, function(response) {
    if (response.success) {
        // Группа обновлена
    } else {
        // Ошибка: response.error
    }
});
```

## Валидация

### Статус

Допустимые значения: "enable" или "disable".

**Используется:** `utils.ValidateStatus()`

### Уникальность имени группы

- При создании: проверка, что имя не существует
- При редактировании: проверка, что имя не существует (исключая текущую группу)

**Используется:** 
- `repositories.GroupRepository.ExistsByName()`
- `repositories.GroupRepository.ExistsByNameExcludingID()`

## Безопасность

- ✅ Проверка авторизации (только авторизованные пользователи)
- ✅ Проверка прав доступа (только администраторы)
- ✅ Валидация всех входных данных
- ✅ Prepared statements в репозиториях (защита от SQL injection)
- ✅ Логирование всех операций создания/редактирования

## Обработка ошибок

Все методы возвращают ответ в формате:
```json
{
  "error": false/string,
  "success": true/false
}
```

При ошибке:
- `error` содержит описание ошибки
- `success` = false

При успехе:
- `error` = false
- `success` = true

## Логирование

Все операции логируются:
- Создание группы: `logger.Info()` с деталями
- Редактирование группы: `logger.Info()` с деталями
- Ошибки: `logger.Error()` с деталями

## Специальные группы

В системе есть предопределённые группы:
- **ID=1**: Администраторы (Admin) - системная группа
- **ID=2**: Пользователи (User) - обычные пользователи

**ВАЖНО:** Эти группы не должны быть удалены или переименованы, так как они используются системой для авторизации.

## Следующие шаги

После реализации GroupController можно переходить к:
- 3.7: UI Templates (страницы Users, Groups)
- 3.8: Модальные формы создания/редактирования

