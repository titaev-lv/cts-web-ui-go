# UserController - Документация

## Описание

`UserController` предоставляет HTTP handlers для управления пользователями. Все методы требуют авторизации, а операции создания/редактирования - прав администратора.

## Методы

### List() - Страница управления пользователями

**Маршрут:** `GET /users/`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Что делает:**
- Отображает страницу управления пользователями
- Пока использует шаблон `index.html` (в фазе 3.7 будет создан `users/index.html`)

**Пример использования:**
```go
users.GET("/", userController.List)
```

### AjaxGetUsers() - Получение списка пользователей для DataTables

**Маршрут:** `POST /users/ajax_get_users`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры:** DataTables параметры (см. DATATABLES_API.md)

**Формат ответа:**
```json
{
  "draw": 1,
  "recordsTotal": 100,
  "recordsFiltered": 50,
  "aaData": [...]
}
```

### AjaxGetUserById() - Получение пользователя по ID

**Маршрут:** `GET /users/ajax_getid_user?id=1` или `POST /users/ajax_getid_user`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры:**
- `id` - ID пользователя (query или form)

**Формат ответа:**
```json
{
  "error": false,
  "success": true,
  "data": {
    "id": "1",
    "login": "admin",
    "groups": "1,2",
    "status": "enable",
    "last_name": "Doe",
    "name": "John",
    "email": "admin@example.com"
  }
}
```

**Пример использования (JavaScript):**
```javascript
$.post('/users/ajax_getid_user', {id: 1}, function(response) {
    if (response.success) {
        // Использовать response.data
    }
});
```

### AjaxCreateUser() - Создание пользователя

**Маршрут:** `POST /users/ajax_create_user`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры (POST form data):**
- `create_user_login` - логин (обязательно)
- `create_user_password` - пароль (обязательно)
- `create_user_password_confirm` - подтверждение пароля (обязательно)
- `create_user_email` - email (обязательно)
- `create_user_groups` - группы через запятую, например "1,2" (обязательно)
- `create_user_status` - статус: "enable" или "disable" (обязательно)
- `create_user_name` - имя (обязательно)
- `create_user_last_name` - фамилия (обязательно)

**Валидация:**
- ✅ Проверка заполненности всех обязательных полей
- ✅ Валидация пароля (минимум 10 символов, буквы, цифры)
- ✅ Проверка совпадения пароля и подтверждения
- ✅ Валидация email (формат @.+\.)
- ✅ Проверка уникальности логина
- ✅ Проверка уникальности email
- ✅ Валидация статуса ("enable" или "disable")
- ✅ Проверка существования всех групп

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
    create_user_login: "newuser",
    create_user_password: "MyPassword123",
    create_user_password_confirm: "MyPassword123",
    create_user_email: "user@example.com",
    create_user_groups: "2",
    create_user_status: "enable",
    create_user_name: "John",
    create_user_last_name: "Doe"
};

$.post('/users/ajax_create_user', formData, function(response) {
    if (response.success) {
        // Пользователь создан
    } else {
        // Ошибка: response.error
    }
});
```

### AjaxEditUser() - Редактирование пользователя

**Маршрут:** `POST /users/ajax_edit_user`

**Требования:**
- Авторизация обязательна
- Только администраторы

**Параметры (POST form data):**
- `edit_user_id` - ID пользователя (обязательно)
- `edit_user_login` - логин (обязательно)
- `edit_user_password` - пароль (опционально, если пустой - не обновляется)
- `edit_user_password_confirm` - подтверждение пароля (обязательно, если пароль указан)
- `edit_user_email` - email (обязательно)
- `edit_user_groups` - группы через запятую, например "1,2" (обязательно)
- `edit_user_status` - статус: "enable" или "disable" (обязательно)
- `edit_user_name` - имя (обязательно)
- `edit_user_last_name` - фамилия (обязательно)

**Валидация:**
- ✅ Проверка заполненности всех обязательных полей
- ✅ Валидация пароля (если указан)
- ✅ Проверка совпадения пароля и подтверждения (если пароль указан)
- ✅ Валидация email (формат @.+\.)
- ✅ Проверка уникальности логина (исключая текущего пользователя)
- ✅ Валидация статуса ("enable" или "disable")
- ✅ Проверка существования всех групп

**Формат ответа:** Аналогично `AjaxCreateUser`

**Пример использования (JavaScript):**
```javascript
var formData = {
    edit_user_id: "1",
    edit_user_login: "updateduser",
    edit_user_email: "updated@example.com",
    edit_user_groups: "1,2",
    edit_user_status: "enable",
    edit_user_name: "Updated",
    edit_user_last_name: "Name"
    // Пароль не указан - не будет обновлён
};

$.post('/users/ajax_edit_user', formData, function(response) {
    if (response.success) {
        // Пользователь обновлён
    } else {
        // Ошибка: response.error
    }
});
```

## Валидация

### Пароль

Требования к паролю:
- Минимум 10 символов
- Должен содержать строчные буквы (a-z)
- Должен содержать заглавные буквы (A-Z)
- Должен содержать цифры (0-9)
- Может содержать только буквы и цифры (без спецсимволов)

**Используется:** `utils.PasswordValidate()` и `utils.PasswordValidateWithConfirm()`

### Email

Простая проверка формата (как в PHP: `/@.+\./`).

**Используется:** `utils.ValidateEmail()`

### Статус

Допустимые значения: "enable" или "disable".

**Используется:** `utils.ValidateStatus()`

### Группы

- Группы передаются строкой через запятую (например, "1,2,3")
- Проверяется, что все группы существуют в БД
- Должна быть указана хотя бы одна группа

**Используется:** `utils.ParseGroupIDs()` и `utils.ValidateGroupIDs()`

### Уникальность логина

- При создании: проверка, что логин не существует
- При редактировании: проверка, что логин не существует (исключая текущего пользователя)

**Используется:** `repositories.UserRepository.ExistsByLogin()` и `ExistsByLoginExcludingID()`

## Безопасность

- ✅ Проверка авторизации (только авторизованные пользователи)
- ✅ Проверка прав доступа (только администраторы)
- ✅ Валидация всех входных данных
- ✅ Хеширование паролей (bcrypt)
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
- Создание пользователя: `logger.Info()` с деталями
- Редактирование пользователя: `logger.Info()` с деталями
- Ошибки: `logger.Error()` с деталями

## Следующие шаги

После реализации UserController можно переходить к:
- 3.5: Group Controller (List, Create, Edit, GetById)
- 3.7: UI Templates (страницы Users, Groups)
- 3.8: Модальные формы создания/редактирования

