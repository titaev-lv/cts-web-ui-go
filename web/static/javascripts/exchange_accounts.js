// Client-side logic for Exchange Accounts page (DataTables + modals).
(function() {
    let table;

    function initTable() {
        // Добавляем input поля поиска в заголовки таблицы
        var columnNames = Array(
            "",
            "ex_acc_id",
            "ex_acc_exchange_id",
            "ex_acc_account_name",
            "ex_acc_priority",
            "ex_acc_status",
            "ex_acc_api_key",
            "ex_acc_note"
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
                { data: 'exchange_id' },
                { data: 'account_name' },
                { data: 'priority' },
                { data: 'status' },
                { data: 'api_key' },
                { data: 'note' }
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
                new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                return;
            }
            const form = $('#form-edit-exaccount')[0];
            form.reset();
            $('[name=edit_exchange_account_id]').val(resp.id);
            $('[name=edit_exchange_account_exid]').val(resp.exchange_id);
            $('[name=edit_exchange_account_account_name]').val(resp.account_name);
            $('[name=edit_exchange_account_priority]').val(resp.priority);
            $('[name=edit_exchange_account_status]').val(resp.status);
            $('[name=edit_exchange_account_api_key]').val(resp.api_key);
            $('[name=edit_exchange_account_secret_key]').val(resp.secret_key);
            $('[name=edit_exchange_account_add_key]').val(resp.add_key);
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
            $.post('/exchange_accounts/ajax_create_account', form.serialize(), function(resp) {
                if (resp.error) {
                    new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                    return;
                }
                new PNotify({ title: 'Success', text: 'Account created', type: 'success', addclass: 'stack-bar-top', width: '100%' });
                $.magnificPopup.close();
                form[0].reset();
                table.ajax.reload(null, false);
            }, 'json');
        });
    }

    function bindEdit() {
        $('#btn-update-exaccount').on('click', function() {
            const form = $('#form-edit-exaccount');
            $.post('/exchange_accounts/ajax_edit_account', form.serialize(), function(resp) {
                if (resp.error) {
                    new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                    return;
                }
                new PNotify({ title: 'Success', text: 'Account updated', type: 'success', addclass: 'stack-bar-top', width: '100%' });
                $.magnificPopup.close();
                table.ajax.reload(null, false);
            }, 'json');
        });
    }

    $(function() {
        initTable();
        bindCreate();
        bindEdit();
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
