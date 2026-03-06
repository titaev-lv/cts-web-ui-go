/**
 * 2FA challenge JavaScript
 * Обработка второго шага входа через AJAX
 */

$(function () {
    $('#auth2FAForm').submit(function (e) {
        e.preventDefault();

        var $form = $(this);
        var $code = $('#twofa_code');
        var code = ($code.val() || '').trim();

        if (code === '') {
            $code.addClass('err').css('background-color', '#f48e70');
            new PNotify({
                title: 'ERROR',
                text: 'Enter one-time code or recovery code',
                type: 'error',
                addclass: 'stack-bar-top',
                width: '100%'
            });
            return false;
        }

        var $submitButtons = $form.find('button[type="submit"]');
        $submitButtons.prop('disabled', true);

        $.ajax({
            url: '/auth/login',
            type: 'POST',
            data: $form.serialize(),
            dataType: 'json',
            contentType: 'application/x-www-form-urlencoded',
            success: function (response) {
                if (response.error && response.error !== false && response.error !== '') {
                    new PNotify({
                        title: 'ERROR',
                        text: response.error,
                        type: 'error',
                        addclass: 'stack-bar-top',
                        width: '100%'
                    });
                    $submitButtons.prop('disabled', false);
                    return;
                }

                if (response.success === true) {
                    window.location.href = '/';
                    return;
                }

                new PNotify({
                    title: 'ERROR',
                    text: 'Unexpected server response',
                    type: 'error',
                    addclass: 'stack-bar-top',
                    width: '100%'
                });
                $submitButtons.prop('disabled', false);
            },
            error: function () {
                new PNotify({
                    title: 'ERROR',
                    text: 'An error occurred during verification',
                    type: 'error',
                    addclass: 'stack-bar-top',
                    width: '100%'
                });
                $submitButtons.prop('disabled', false);
            }
        });

        return false;
    });

    $('#twofa_code').on('input', function () {
        if ($(this).val().length > 0) {
            $(this).removeClass('err').css('background-color', '');
        }
    });
});
