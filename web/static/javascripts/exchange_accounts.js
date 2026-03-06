// Client-side logic for Exchange Accounts page (DataTables + modals).
(function() {
    let table;

    function extractAjaxError(xhr, fallback) {
        if (xhr && xhr.responseJSON && xhr.responseJSON.error) {
            return xhr.responseJSON.error;
        }
        if (xhr && xhr.responseText) {
            try {
                var parsed = JSON.parse(xhr.responseText);
                if (parsed && parsed.error) {
                    return parsed.error;
                }
            } catch (e) {
                // Ignore parse error and fallback to raw text.
            }
            if (xhr.responseText.trim() !== '') {
                return xhr.responseText;
            }
        }
        return fallback || 'Request failed';
    }

    function showError(text) {
        new PNotify({ title: 'Error', text: text, type: 'error', addclass: 'stack-bar-top', width: '100%' });
    }

    function showSuccess(text) {
        new PNotify({ title: 'Success', text: text, type: 'success', addclass: 'stack-bar-top', width: '100%' });
    }

    function isSelectedExchangeActive($exchangeSelect) {
        var active = $exchangeSelect.find('option:selected').data('active');
        return active === 1 || active === '1' || active === true;
    }

    function renderStatusOptions($statusSelect, exchangeActive, preferredStatus) {
        var options = exchangeActive
            ? [
                { value: 'Active', text: 'Active' },
                { value: 'Blocked', text: 'Blocked' }
            ]
            : [
                { value: 'Blocked', text: 'Blocked' }
            ];

        var nextStatus = preferredStatus || 'Blocked';
        if (!exchangeActive && nextStatus === 'Active') {
            nextStatus = 'Blocked';
        }

        $statusSelect.empty();
        options.forEach(function(opt) {
            var $o = $('<option>').val(opt.value).text(opt.text);
            if (opt.value === nextStatus) {
                $o.prop('selected', true);
            }
            $statusSelect.append($o);
        });
    }

    function syncStatusByExchange(exchangeSelectSelector, statusSelectSelector, preferredStatus) {
        var $exchange = $(exchangeSelectSelector);
        var $status = $(statusSelectSelector);
        var active = isSelectedExchangeActive($exchange);
        renderStatusOptions($status, active, preferredStatus || $status.val());
    }

    function initTable() {
        // Добавляем input поля поиска в заголовки таблицы
        var columnNames = Array(
            "",
            "ex_acc_id",
            "ex_acc_exchange",
            "ex_acc_account_name",
            "ex_acc_note",
            "ex_acc_priority",
            "ex_acc_status",
        );

        $('#dt-exchange-accounts thead tr th').each(function (i) {
            var title = $(this).text();
            if(i > 0) {
                $(this).html(title + ' <input type="text" name="' + columnNames[i] + '@' + i + '" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchExchangeAccount(event)" />');
            }
        });

        table = $('#dt-exchange-accounts').DataTable({
            processing: true,
            serverSide: true,
            pageLength: 15,
            lengthMenu: [15, 30, 50, 100],
            pagingExtraNumberForNext: true,
            bScrollCollapse: true,
            ajax: {
                url: '/exchange_accounts/ajax_get_accounts',
                type: 'POST'
            },
            columns: [
                { data: null, render: function(){ return "<input type='checkbox' class='t-row chbx-ch' value=''/>"; }},
                { data: 'id' },
                { data: 'exchange_name' },
                { data: 'account_name' },
                { data: 'note' },
                { data: 'priority' },
                { data: 'status' }
            ],
            columnDefs: [
                {
                    targets: "_all",
                    className: 'dt-body-left',
                    searchable: true
                },
                {
                    searchable: false, 
                    orderable: false, 
                    visible: true,
                    className: 'no-sort',
                    targets: [0]
                }
            ],
            order: [1, 'asc'],
            language: {
                processing: "Processing...",
                lengthMenu: "_MENU_ accounts per page",
                zeroRecords: "Data not found",
                info: "Filtered from _START_ to _END_ of _TOTAL_",
                infoEmpty: "Data not found",
                infoFiltered: "(Total accounts _MAX_)"
            }
        });

        // Скрываем глобальное поле поиска (используем поиск по столбцам)
        var search = document.getElementById('dt-exchange-accounts_filter');
        if (search) {
            document.getElementById('dt-exchange-accounts_filter').style.display = 'none';
        }

        $('#dt-exchange-accounts tbody').on('dblclick', 'tr', function() {
            const data = table.row(this).data();
            if (!data) return;
            loadAccountForEdit(data.id);
        });
    }

    function loadAccountForEdit(id) {
        $.post('/exchange_accounts/ajax_getid_accounts', { id: id }, function(resp) {
            if (resp.error) {
                showError(resp.error);
                return;
            }
            const form = $('#form-edit-exaccount')[0];
            form.reset();
            $('[name=edit_exchange_account_id]').val(resp.id);
            $('[name=edit_exchange_account_exid]').val(resp.exchange_id);
            $('[name=edit_exchange_account_account_name]').val(resp.account_name);
            $('[name=edit_exchange_account_priority]').val(resp.priority);
            syncStatusByExchange('#edit_exchange_account_exid', '#edit_exchange_account_status', resp.status);
            // Do not show secret values in edit form; allow replace-only input.
            $('[name=edit_exchange_account_api_key]').val('');
            $('[name=edit_exchange_account_secret_key]').val('');
            $('[name=edit_exchange_account_add_key]').val('');

            $('#edit_api_key_state').text(resp.has_api_key ? 'Stored in DB' : 'Not set in DB');
            $('#edit_secret_key_state').text(resp.has_secret_key ? 'Stored in DB' : 'Not set in DB');
            $('#edit_add_key_state').text(resp.has_add_key ? 'Stored in DB' : 'Not set in DB');
            $('[name=edit_exchange_account_note]').val(resp.note || '');
            $.magnificPopup.open({
                type: 'inline',
                items: {
                    src: '#modalAccountEdit'
                },
                preloader: false,
                modal: true,
                closeOnContentClick: false,
                closeOnBgClick: false
            });
        }, 'json');
    }

    function bindCreate() {
        $('#btn-save-exaccount').on('click', function() {
            const form = $('#form-create-exaccount');
            if (!validateEmptyFormFields('form-create-exaccount')) {
                showError('Please fill in all required fields');
                return;
            }

            $.ajax({
                url: '/exchange_accounts/ajax_create_account',
                type: 'POST',
                data: form.serialize(),
                dataType: 'json',
                success: function(resp) {
                    if (resp && resp.error) {
                        showError(resp.error);
                        return;
                    }
                    showSuccess('Account created');
                    $.magnificPopup.close();
                    form[0].reset();
                    table.ajax.reload(null, false);
                },
                error: function(xhr) {
                    showError(extractAjaxError(xhr, 'Failed to create account'));
                }
            });
        });
    }

    function bindEdit() {
        $('#btn-update-exaccount').on('click', function() {
            const form = $('#form-edit-exaccount');
            if (!validateEmptyFormFields('form-edit-exaccount')) {
                showError('Please fill in all required fields');
                return;
            }

            $.ajax({
                url: '/exchange_accounts/ajax_edit_account',
                type: 'POST',
                data: form.serialize(),
                dataType: 'json',
                success: function(resp) {
                    if (resp && resp.error) {
                        showError(resp.error);
                        return;
                    }
                    showSuccess('Account updated');
                    $.magnificPopup.close();
                    table.ajax.reload(null, false);
                },
                error: function(xhr) {
                    showError(extractAjaxError(xhr, 'Failed to update account'));
                }
            });
        });
    }

    $(function() {
        initTable();
        bindCreate();
        bindEdit();

        // Keep status options consistent with selected exchange activity.
        syncStatusByExchange('#create_exchange_account_exid', '#create_exchange_account_status', 'Active');
        syncStatusByExchange('#edit_exchange_account_exid', '#edit_exchange_account_status', 'Blocked');

        $('#create_exchange_account_exid').on('change', function() {
            syncStatusByExchange('#create_exchange_account_exid', '#create_exchange_account_status');
        });
        $('#edit_exchange_account_exid').on('change', function() {
            syncStatusByExchange('#edit_exchange_account_exid', '#edit_exchange_account_status');
        });

        // Remove red highlight when user fixes required fields.
        $('#form-create-exaccount, #form-edit-exaccount').on('input change', 'input[required], textarea[required], select[required]', function() {
            if ($(this).val() !== null && String($(this).val()).trim() !== '') {
                $(this).removeClass('err');
            }
        });
    });
})();

// Функция поиска по столбцам для аккаунтов
function keysearchExchangeAccount(event) {
    if(event.keyCode === 13) {
        var table = $('#dt-exchange-accounts').DataTable();
        var input = event.target;
        var col_index = input.name.match(/\d+/)[0];
        var col_name = input.name.replace(/@\d+/, '');
        
        table.columns(col_index).search(input.value).draw();
    }
}
