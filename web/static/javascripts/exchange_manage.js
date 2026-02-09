// Client-side logic for Exchanges page (DataTables + modals).
(function() {
    let table;

    function initTable() {
        // Добавляем input поля поиска в заголовки таблицы
        var columnNames = Array(
            "",
            "ex_id",
            "ex_name",
            "ex_status",
            "ex_url",
            "ex_base_url",
            "ex_websocket_url",
            "ex_class"
        );

        $('#dt-exchange-manage thead tr th').each(function (i) {
            var title = $(this).text();
            if(i > 0) {
                $(this).html(title + ' <input type="text" name="' + columnNames[i] + '@' + i + '" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchExchange(event)" />');
            }
        });

        table = $('#dt-exchange-manage').DataTable({
            processing: true,
            serverSide: true,
            pageLength: 15,
            lengthMenu: [15, 30, 50, 100],
            pagingExtraNumberForNext: true,
            bScrollCollapse: true,
            ajax: {
                url: '/exchange_manage/ajax_get_exchanges',
                type: 'POST'
            },
            columns: [
                { data: null, render: function(){ return "<input type='checkbox' class='t-row chbx-ch' value=''/>"; }},
                { data: 'id' },
                { data: 'name' },
                { data: 'status' },
                { data: 'url' },
                { data: 'base_url' },
                { data: 'websocket_url' },
                { data: 'class' }
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
                lengthMenu: "_MENU_ exchanges per page",
                zeroRecords: "Data not found",
                info: "Filtered from _START_ to _END_ of _TOTAL_",
                infoEmpty: "Data not found",
                infoFiltered: "(Total exchanges _MAX_)"
            }
        });

        // Скрываем глобальное поле поиска (используем поиск по столбцам)
        var search = document.getElementById('dt-exchange-manage_filter');
        if (search) {
            document.getElementById('dt-exchange-manage_filter').style.display = 'none';
        }

        $('#dt-exchange-manage tbody').on('dblclick', 'tr', function() {
            const data = table.row(this).data();
            if (!data) return;
            loadExchangeForEdit(data.id);
        });
    }

    function loadExchangeForEdit(id) {
        $.post('/exchange_manage/ajax_getid_exchange', { id: id }, function(resp) {
            if (resp.error) {
                new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                return;
            }
            const form = $('#form-edit-exchange')[0];
            form.reset();
            $('[name=edit_exchange_id]').val(resp.id);
            $('[name=edit_exchange_name]').val(resp.name);
            $('[name=edit_exchange_url]').val(resp.url);
            $('[name=edit_exchange_base_url]').val(resp.base_url);
            $('[name=edit_exchange_websocket_url]').val(resp.websocket_url || '');
            $('[name=edit_exchange_class]').val(resp.class);
            $('[name=edit_exchange_status]').val(resp.status);
            $('[name=edit_exchange_description]').val(resp.description || '');
            $.magnificPopup.open({
                type: 'inline',
                items: {
                    src: '#modalExchangeEdit'
                },
                preloader: false,
                modal: true,
                closeOnContentClick: false,
                closeOnBgClick: false
            });
        }, 'json');
    }

    function bindCreate() {
        $('#btn-save-exchange').on('click', function() {
            const form = $('#form-create-exchange');
            $.post('/exchange_manage/ajax_create_exchange', form.serialize(), function(resp) {
                if (resp.error) {
                    new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                    return;
                }
                new PNotify({ title: 'Success', text: 'Exchange created', type: 'success', addclass: 'stack-bar-top', width: '100%' });
                $.magnificPopup.close();
                form[0].reset();
                table.ajax.reload(null, false);
            }, 'json');
        });
    }

    function bindEdit() {
        $('#btn-update-exchange').on('click', function() {
            const form = $('#form-edit-exchange');
            $.post('/exchange_manage/ajax_edit_exchange', form.serialize(), function(resp) {
                if (resp.error) {
                    new PNotify({ title: 'Error', text: resp.error, type: 'error', addclass: 'stack-bar-top', width: '100%' });
                    return;
                }
                new PNotify({ title: 'Success', text: 'Exchange updated', type: 'success', addclass: 'stack-bar-top', width: '100%' });
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

// Функция поиска по столбцам для бирж
function keysearchExchange(event) {
    if(event.keyCode === 13) {
        event.preventDefault();
        var table = $('#dt-exchange-manage').DataTable();
        if (!table) {
            console.error("DataTables for exchanges is not initialized.");
            return;
        }
        
        // Получаем все заголовки колонок
        var h = table.columns().header();
        h.toArray().forEach(function(item, i, arr) {
            // Ищем поле ввода в заголовке
            if(item.children[0] && item.children[0].tagName === 'INPUT') {
                var input = item.children[0];
                var v = input.value;
                var nameParts = input.name.split('@');
                
                // Формат name: "ex_name@1" -> индекс колонки = 1
                if(nameParts.length >= 2) {
                    var colIndex = parseInt(nameParts[1]);
                    if (!isNaN(colIndex)) {
                        // Устанавливаем поиск для колонки
                        // DataTables автоматически отправит columns[colIndex][search][value] на сервер
                        table.column(colIndex).search(v);
                    }
                }
            }
        });
        
        // Вызываем draw() для отправки запроса на сервер с новыми параметрами поиска
        table.draw();
    }
}