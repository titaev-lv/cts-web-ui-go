# Примеры использования auth.js

## Базовая обработка формы входа

Форма автоматически обрабатывается при загрузке страницы:

```html
<form id="authForm" method="post">
    <input name="username" type="text" required />
    <input name="pwd" type="password" required />
    <input id="RememberMe" name="rememberme" type="checkbox"/>
    <button type="submit">Sign in</button>
</form>
```

## Формат ответа сервера

### Успешный вход:
```json
{
    "error": false,
    "success": true
}
```

### Ошибка входа:
```json
{
    "error": "Bad Login or Password",
    "success": false
}
```

## Обработка событий

### Успешный вход
- Автоматический редирект на главную страницу (`/`)
- Кнопки отправки разблокируются

### Ошибка входа
- Показывается PNotify уведомление с текстом ошибки
- Кнопки отправки разблокируются
- Форма остаётся на странице

### HTTP ошибки
- **401 Unauthorized**: "Error 401 Unauthorized! Incorrect Login or Password"
- **404 Not Found**: "Error 404 Not found. Auth endpoint not found!"
- **500 Internal Server Error**: "Error 500 Internal Server Error. Please try again later."
- **Timeout**: "Request timeout. Please check your connection and try again."

## Кастомизация

### Изменение URL для отправки формы

По умолчанию используется `/auth/login`. Для изменения:

```javascript
$("#authForm").submit(function(e) {
    e.preventDefault();
    $.ajax({
        url: '/your/custom/endpoint', // Изменить здесь
        // ...
    });
});
```

### Изменение редиректа после успешного входа

По умолчанию редирект на `/`. Для изменения:

```javascript
} else if (response.success === true) {
    window.location.href = '/your/custom/redirect'; // Изменить здесь
}
```

### Добавление дополнительной обработки

```javascript
$("#authForm").submit(function(e) {
    e.preventDefault();
    
    // Ваша дополнительная логика перед отправкой
    // Например, валидация полей
    
    $.ajax({
        // ...
        success: function (response) {
            // Ваша дополнительная логика после успешного входа
            // Например, сохранение данных в localStorage
            
            if (response.success === true) {
                window.location.href = '/';
            }
        }
    });
});
```

