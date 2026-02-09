/**
 * Authentication JavaScript
 * Обработка формы входа через AJAX
 */

$(function () {
    /**
     * Обработка формы входа
     * Отправляет данные формы на сервер через AJAX
     * При успешном входе перенаправляет на главную страницу
     * При ошибке показывает уведомление
     */
    $("#authForm").submit(function(e) {
        // Предотвращаем стандартную отправку формы
        e.preventDefault();

        // Получаем форму
        var $form = $(this);
        
        // Блокируем кнопки отправки, чтобы предотвратить повторную отправку
        var $submitButtons = $form.find('button[type="submit"]');
        $submitButtons.prop('disabled', true);

        // Отправляем AJAX запрос
        $.ajax({
            url: '/auth/login',
            type: 'POST',
            data: $form.serialize(), // Сериализуем данные формы (username, pwd, rememberme)
            dataType: 'json', // Ожидаем JSON ответ
            contentType: 'application/x-www-form-urlencoded',
            success: function (response) {
                // Проверяем ответ сервера
                // Формат ответа как в PHP: {"error": false/string, "success": true/false}
                
                // Проверяем наличие ошибки
                // В PHP: ret.error !== false && ret.error !== ''
                // В Go: error может быть строкой или false
                if (response.error && response.error !== false && response.error !== '') {
                    // Есть ошибка - показываем уведомление
                    new PNotify({
                        title: 'ERROR',
                        text: response.error,
                        type: 'error',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                    
                    // Разблокируем кнопки
                    $submitButtons.prop('disabled', false);
                } else if (response.success === true) {
                    // Успешный вход - перенаправляем на главную страницу
                    window.location.href = '/';
                } else {
                    // Неожиданный формат ответа
                    new PNotify({
                        title: 'ERROR',
                        text: 'Unexpected server response',
                        type: 'error',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                    
                    // Разблокируем кнопки
                    $submitButtons.prop('disabled', false);
                }
            },
            error: function (xhr, status, error) {
                // Обработка ошибок HTTP
                var errorMessage = 'An error occurred during login';
                
                // Определяем сообщение об ошибке в зависимости от статуса
                if (xhr.status === 401) {
                    errorMessage = 'Error 401 Unauthorized!<br>Incorrect Login or Password';
                } else if (xhr.status === 404) {
                    errorMessage = 'Error 404 Not found.<br>Auth endpoint not found!';
                } else if (xhr.status === 500) {
                    errorMessage = 'Error 500 Internal Server Error.<br>Please try again later.';
                } else if (status === 'timeout') {
                    errorMessage = 'Request timeout.<br>Please check your connection and try again.';
                } else if (status === 'parsererror') {
                    errorMessage = 'Error parsing server response.<br>Please try again.';
                }
                
                // Пытаемся получить сообщение об ошибке из ответа сервера
                if (xhr.responseJSON && xhr.responseJSON.error) {
                    errorMessage = xhr.responseJSON.error;
                }
                
                // Показываем уведомление об ошибке
                new PNotify({
                    title: 'ERROR',
                    text: errorMessage,
                    type: 'error',
                    addclass: 'stack-bar-top',
                    width: "100%"
                });
                
                // Разблокируем кнопки
                $submitButtons.prop('disabled', false);
            },
            complete: function () {
                // Эта функция вызывается всегда, даже при ошибке
                // Можно использовать для очистки или финальных действий
            }
        });
        
        // Возвращаем false, чтобы предотвратить стандартную отправку формы
        return false;
    });
});

