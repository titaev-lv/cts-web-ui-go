/**
 * User and Group Management JavaScript
 * 
 * Этот файл содержит JavaScript код для управления пользователями и группами:
 * - Инициализация DataTables для пользователей и групп
 * - Обработка модальных форм создания/редактирования
 * - Валидация форм
 * - AJAX запросы к серверу
 */

$(document).ready(function() {
    /*
     * 1. Инициализация DataTables для пользователей
     */  
    var dtuser = document.getElementById('dt-user');
    var tableUsers = null;
    
    if(dtuser) {
        // Вставляем поля поиска в заголовки таблицы
        var nm = Array(
            "",
            "u_id", 
            "u_login",
            "u_groups",
            "u_active",         
            "u_name",
            "u_email",
            "u_date_create",
            "u_date_modify",
            "u_timestamp",
        );
        
        $('#dt-user thead tr th').each(function (i) {
            var title = $(this).text();
            if(i > 0) {
                $(this).html(title+' <input type="text" name="'+nm[i]+'@'+i+'" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchUser(event)" />');
            }
        });
        	
        $.fn.dataTable.ext.errMode = 'throw';
        
        tableUsers = $('#dt-user').DataTable({
            "processing": true,
            "serverSide": true,
            "searching": true,
            "pageLength": 15,
            "lengthMenu": [15, 30, 50, 100],
            "pagingExtraNumberForNext": true,
            "bScrollCollapse": true,
            "columns": [
                { "data": null, render: function(){ return "<input type='checkbox' class='t-row chbx-ch' value=''/>"; }}, // 0
                { "data": "id" },              // 1             
                { "data": "login"},            // 2
                { "data": "groups"},           // 3
                { "data": "active" },           // 4
                { "data": "name" },            // 5
                { "data": "email" },           // 6
                { "data": "create_date" },      // 7
                { "data": "modify_date" },      // 8
                { "data": "timestamp" },        // 9
            ],
            "language": {
                "processing": "Processing...",
                "lengthMenu": "_MENU_ user per page",
                "zeroRecords": "Data not found",
                "info": "Filtered from _START_ to _END_ of _TOTAL_",
                "infoEmpty": "Data not found",
                "infoFiltered": "(Total users _MAX_)"
            },
            "ajax": {
                "method": "POST",
                "url": "/users/ajax_get_users",
                "data": function (d) {
                    // Можно добавить дополнительные фильтры
                },
                "error": function (xhr, error, thrown) {
                    console.error("DataTables error:", error);
                },
                "statusCode": {
                    401: function (xhr, error, thrown) {
                        location.reload();
                    }
                }
            },
            "createdRow": function (row, data, dataIndex, cells) {
                // Можно добавить кастомную обработку строк
            },
            "columnDefs": [
                {
                    "targets": "_all",
                    "className": 'dt-body-left',
                    "searchable": true
                },
                {
                    "searchable": false, 
                    "orderable": false, 
                    "visible": true,
                    "targets": [0]
                },
            ],
            "select": {
                "style": 'os',
                "selector": 'td:first-child'
            },
            "order": [1, 'asc'],
            "drawCallback": function(settings) {
                // Можно добавить кастомную обработку после отрисовки
            },     
        });
        
        // Скрываем глобальное поле поиска (используем поиск по колонкам)
        var search = document.getElementById('dt-user_filter');
        if (search) {
            document.getElementById('dt-user_filter').style.display = 'none';
        } 
        
        // Подсветка строки при наведении мыши
        $('#dt-user tbody').on('click', 'tr', function () {
            if ($(this).hasClass('selectedd')) {
                $(this).removeClass('selectedd');
            } else {
                tableUsers.$('tr.selectedd').removeClass('selectedd');
                $(this).addClass('selectedd');
            }
        }); 
        
        $('#dt-user tbody').on('mouseover', 'tr', function () {
            if (!$(this).hasClass('hoveredd')) {
                $(this).addClass('hoveredd');
            }
        });
        
        $('#dt-user tbody').on('mouseout', 'tr', function () {
            if ($(this).hasClass('hoveredd')) {
                $(this).removeClass('hoveredd');
            } 
        });
        
        // Обработка checkbox "выделить/не выделить все записи"
        $('#checkallUser').on('click', function() {
            var cells = tableUsers.column(0).nodes();
            var state = this.checked;
            for (var i = 0; i < cells.length; i += 1) {
                cells[i].querySelector("input[type='checkbox']").checked = state;
            }
        });
        
        // Кнопка создания пользователя
        $('#create_user_button').on('click', function(e) {
            e.preventDefault();
            var isNotValid = false;
            var passError = false; 
            var lenghtPassError = false;
            
            // Валидация обязательных полей
            $("#create-user-form").find('input, textarea, select').each(function(e, elements) {
                if(elements.required === true) {
                    if(elements.value === null || elements.value === '') {
                        $(elements).addClass("err");
                        isNotValid = true;
                    } else {
                        $(elements).removeClass("err");
                    }
                }
            });
            
            // Проверка совпадения паролей
            let p1 = $("#create_user_password").val();
            let p2 = $("#create_user_password_confirm").val();
            if(p1 !== p2) {
                passError = true;
                $('#create_user_password_confirm').addClass("err");
            }
            
            // Проверка длины пароля
            if(p1.length < 10) {
                lenghtPassError = true;
            }

            if(isNotValid === false && passError === false && lenghtPassError === false) {
                var formData = new FormData();
                var data = $('#create-user-form').serializeArray();
                var grp = Array();
                
                // Собираем группы в массив
                $.each(data, function(key, input) {
                    if(input.name === 'create_user_groups') {
                        grp.push(input.value);
                    }
                });
                formData.append('create_user_groups', grp.join(','));
                
                // Добавляем остальные поля
                $.each(data, function(key, input) {
                    if(input.name !== 'create_user_groups') {
                        formData.append(input.name, input.value);
                    }   
                });
                
                $.ajax({
                    url: "/users/ajax_create_user",
                    type: "POST",
                    data: formData,
                    processData: false,
                    contentType: false,
                    success: function(response) {
                        if(response.error !== false && response.error !== '') {
                            new PNotify({
                                title: 'Error',
                                text: response.error,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                        } else if(response.success === true) {
                            new PNotify({
                                text: 'User created',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                            $.magnificPopup.close();
                            tableUsers.draw();
                            $("form#create-user-form").trigger('reset');
                        }
                    },
                    error: function (data, textStatus) {
                        if(data.status == 401) {
                            setTimeout(function(){ location.reload(); }, 1000);
                        }
                        new PNotify({
                            title: 'Error',
                            text: "Error " + data.status + " " + data.statusText,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                        });
                    }  
                });
            } else {
                var msg = '';
                if(passError === true) {
                    msg = 'Confirm Password is failed';
                } else if(lenghtPassError === true) {
                    msg = 'Password must be longer than or equal 10 characters';
                } else {
                    msg = 'Required fields is empty';
                }
                
                new PNotify({
                    title: 'Error',
                    text: msg,
                    addclass: 'stack-bar-top',
                    type: 'error',
                    width: "100%"
                });
            }
        });
        
        // Двойной клик для редактирования пользователя
        $('#dt-user tbody').on('dblclick', 'tr', function (e) {
            var data = tableUsers.row(this).data();   
            var id_record = data.DT_RowId;
            id_record = id_record.replace(/row_/, "");
            
            $.magnificPopup.open({
                type: 'inline',
                items: {
                    src: '#modalForm-edit-user'
                },
                preloader: false,
                focus: '#name',
                modal: true,
                closeOnContentClick: false,
                closeOnBgClick: false,
                callbacks: {
                    beforeOpen: function() {
                        if($(window).width() < 700) {
                            this.st.focus = false;
                        } else {
                            this.st.focus = '#name';
                        }
                        
                        // Сброс полей формы
                        $("form#edit-user-form").trigger('reset');
                        $("#edit-user-form").find('input, textarea, select').each(function(e, elements) {
                            if(elements.required === true) {
                                $(elements).removeClass("err");
                            }
                        });

                        // Сброс всех checkbox
                        var cells = tableUsers.column(0).nodes();
                        var state = false;
                        for (var i = 0; i < cells.length; i += 1) {
                            cells[i].querySelector("input[type='checkbox']").checked = state;
                        }
                        
                        // Выделение checkbox выбранной строки
                        var data = tableUsers.columns().rows().nodes();
                        data.toArray().forEach(function(item, i, arr) {
                            var ids = item.id.replace(/row_/, "");
                            if(ids === id_record) {
                                var inp = item.firstChild.firstChild;
                                $(inp).prop('checked', "checked");
                            }
                        });
          
                        // Загрузка данных пользователя
                        $.ajax({
                            url: "/users/ajax_getid_user",
                            type: "POST",
                            dataType: "json",
                            data: {'id': id_record},
                            success: function(response) {
                                if(response.error !== false && response.error !== '') {
                                    new PNotify({
                                        title: 'Error',
                                        text: response.error,
                                        type: 'error',
                                        addclass: 'stack-bar-top',
                                        width: "100%"
                                    });
                                }
                                if(response.data) {
                                    for (var key in response.data) {
                                        var ins = '#edit_user_' + key;
                                        var dat = response.data[key];
                                        
                                        if ($(ins).length) {
                                            if($(ins).prop('nodeName').toLowerCase() === 'input') {
                                                $(ins).val(dat);
                                            }
                                            if($(ins).prop('nodeName').toLowerCase() === 'select') {
                                                if(key === 'groups') {
                                                    // Разбиваем строку групп по запятой и удаляем пробелы
                                                    let grps = dat.split(',').map(function(g) { 
                                                        return String(g.trim()); 
                                                    }).filter(function(g) { 
                                                        return g.length > 0; 
                                                    });
                                                    
                                                    // Сбрасываем все выбранные опции
                                                    $(ins + ' option').prop('selected', false);
                                                    
                                                    // Устанавливаем selected для каждой группы
                                                    // Используем перебор всех опций для надежного сравнения
                                                    $(ins + ' option').each(function() {
                                                        var optionValue = String($(this).val());
                                                        // Проверяем, есть ли это значение в списке групп
                                                        for(var i = 0; i < grps.length; i++) {
                                                            if(String(grps[i]) === optionValue) {
                                                                $(this).prop('selected', true);
                                                                break;
                                                            }
                                                        }
                                                    });
                                                } else {
                                                    $(ins + ' option').prop('selected', false);
                                                    $(ins + ' option[value="' + dat + '"]').prop('selected', true);
                                                }
                                            } 
                                            if($(ins).prop('nodeName').toLowerCase() === 'textarea') {
                                                $(ins).val(dat);
                                            } 
                                        }
                                    }
                                }
                            },
                            error: function (data, textStatus) {
                                if(data.status == 401) {
                                    setTimeout(function(){ location.reload(); }, 1000);
                                }
                                new PNotify({
                                    title: 'Error',
                                    text: "Error " + data.status + " " + data.statusText,
                                    type: 'error',
                                    addclass: 'stack-bar-top',
                                    width: "100%"
                                });
                                $.magnificPopup.close();
                            }  
                        });
                    }
                }       
            });
        }); 
        
        // Кнопка редактирования пользователя
        $('#edit_user_button').on('click', function(e) {
            e.preventDefault();
            var isNotValid = false;
            var passError = false; 
            var lenghtPassError = false;
            
            // Валидация обязательных полей
            $("#edit-user-form").find('input, textarea, select').each(function(e, elements) {
                if(elements.required === true) {
                    if(elements.value === null || elements.value === '') {
                        $(elements).addClass("err");
                        isNotValid = true;
                    } else {
                        $(elements).removeClass("err");
                    }
                } else {
                    $(elements).removeClass("err");
                }
            });
            
            // Проверка совпадения паролей (если указаны)
            let p1 = $("#edit_user_password").val();
            let p2 = $("#edit_user_password_confirm").val();
            if(p1.length > 0 || p2.length > 0) {
                if(p1 !== p2) {
                    passError = true;
                    $('#edit_user_password_confirm').addClass("err");
                }
                // Проверка длины пароля
                if(p1.length > 0 && p1.length < 10) {
                    lenghtPassError = true;
                }
            }

            if(isNotValid === false && passError === false && lenghtPassError === false) {
                var formData = new FormData();
                var data = $('#edit-user-form').serializeArray();
                var grp = Array();
                
                // Собираем группы в массив
                $.each(data, function(key, input) {
                    if(input.name === 'edit_user_groups') {
                        grp.push(input.value);
                    }
                });
                formData.append('edit_user_groups', grp.join(','));
                
                // Добавляем остальные поля
                $.each(data, function(key, input) {
                    if(input.name !== 'edit_user_groups') {
                        formData.append(input.name, input.value);
                    }   
                });
                
                $.ajax({
                    url: "/users/ajax_edit_user",
                    type: "POST",
                    data: formData,
                    processData: false,
                    contentType: false,
                    success: function(response) {
                        if(response.error !== false && response.error !== '') {
                            new PNotify({
                                title: 'Error',
                                text: response.error,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                        } else if(response.success === true) {
                            new PNotify({
                                text: 'User updated',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                            $.magnificPopup.close();
                            tableUsers.draw();
                            $("form#edit-user-form").trigger('reset');
                        }
                    },
                    error: function (data, textStatus) {
                        if(data.status == 401) {
                            setTimeout(function(){ location.reload(); }, 1000);
                        }
                        new PNotify({
                            title: 'Error',
                            text: "Error " + data.status + " " + data.statusText,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                        });
                    }  
                });
            } else {
                var msg = '';
                if(passError === true) {
                    msg = 'Confirm Password is failed';
                } else if(lenghtPassError === true) {
                    msg = 'Password must be longer than or equal 10 characters';
                } else {
                    msg = 'Required fields is empty';
                }
                
                new PNotify({
                    title: 'Error',
                    text: msg,
                    addclass: 'stack-bar-top',
                    type: 'error',
                    width: "100%"
                });
            }
        });
    }
    
    /*
     * 2. Инициализация DataTables для групп
     */  
    var dtgroups = document.getElementById('dt-groups');
    var tableGroups = null;
    // Делаем переменную глобальной для использования в keysearchGroup
    window.tableGroups = null;
    
    if(dtgroups) {
        // Вставляем поля поиска в заголовки таблицы
        var nm = Array(
            "",
            "g_id", 
            "g_name",
            "g_status",         
            "g_description",
        );
        
        $('#dt-groups thead tr th').each(function (i) {
            var title = $(this).text();
            if(i > 0) {
                $(this).html(title+' <input type="text" name="'+nm[i]+'@'+i+'" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchGroup(event)" />');
            }
        });
        	
        $.fn.dataTable.ext.errMode = 'throw';
        
        tableGroups = $('#dt-groups').DataTable({
            "processing": true,
            "serverSide": true,
            "searching": true,
            "pageLength": 15,
            "lengthMenu": [15, 30, 50, 100],
            "pagingExtraNumberForNext": true,
            "bScrollCollapse": true,
            "columns": [
                { "data": null, render: function(){ return "<input type='checkbox' class='t-row chbx-ch' value=''/>"; }}, // 0
                { "data": "id" },              // 1             
                { "data": "name"},             // 2
                { "data": "status"},           // 3
                { "data": "description" },     // 4
            ],
            "language": {
                "processing": "Processing...",
                "lengthMenu": "_MENU_ groups per page",
                "zeroRecords": "Data not found",
                "info": "Filtered from _START_ to _END_ of _TOTAL_",
                "infoEmpty": "Data not found",
                "infoFiltered": "(Total groups _MAX_)"
            },
            "ajax": {
                "method": "POST",
                "url": "/groups/ajax_get_groups",
                "data": function (d) {
                    // Можно добавить дополнительные фильтры
                },
                "error": function (xhr, error, thrown) {
                    console.error("DataTables error:", error);
                },
                "statusCode": {
                    401: function (xhr, error, thrown) {
                        location.reload();
                    }
                }
            },
            "createdRow": function (row, data, dataIndex, cells) {
                // Можно добавить кастомную обработку строк
            },
            "columnDefs": [
                {
                    "targets": "_all",
                    "className": 'dt-body-left',
                    "searchable": true,
                    "orderable": true
                },
                {
                    "searchable": false, 
                    "orderable": false, 
                    "visible": true,
                    "targets": [0]
                },
            ],
            "select": {
                "style": 'os',
                "selector": 'td:first-child'
            },
            "order": [1, 'asc'],
            "drawCallback": function(settings) {
                // Можно добавить кастомную обработку после отрисовки
            },     
        });
        
        // Сохраняем ссылку на таблицу в глобальной переменной для использования в keysearchGroup
        window.tableGroups = tableGroups;
        
        // Скрываем глобальное поле поиска
        var search = document.getElementById('dt-groups_filter');
        if (search) {
            document.getElementById('dt-groups_filter').style.display = 'none';
        } 
        
        // Обработка checkbox "выделить/не выделить все записи"
        $('#checkallGroups').on('click', function() {
            var cells = tableGroups.column(0).nodes();
            var state = this.checked;
            for (var i = 0; i < cells.length; i += 1) {
                cells[i].querySelector("input[type='checkbox']").checked = state;
            }
        });
        
        // Кнопка создания группы
        $('#create_group_button').on('click', function(e) {
            e.preventDefault();
            var isNotValid = false;
         
            // Валидация обязательных полей
            $("#create-group-form").find('input, textarea, select').each(function(e, elements) {
                if(elements.required === true) {
                    if(elements.value === null || elements.value === '') {
                        $(elements).addClass("err");
                        isNotValid = true;
                    } else {
                        $(elements).removeClass("err");
                    }
                }
            });
 
            if(isNotValid === false) {
                var formData = new FormData();
                var data = $('#create-group-form').serializeArray();

                $.each(data, function(key, input) {
                    formData.append(input.name, input.value);
                });
                
                $.ajax({
                    url: "/groups/ajax_create_group",
                    type: "POST",
                    data: formData,
                    processData: false,
                    contentType: false,
                    success: function(response) {
                        if(response.error !== false && response.error !== '') {
                            new PNotify({
                                title: 'Error',
                                text: response.error,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                        } else if(response.success === true) {
                            new PNotify({
                                text: 'Group created',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                            $.magnificPopup.close();
                            tableGroups.draw();
                            $("form#create-group-form").trigger('reset');
                        }
                    },
                    error: function (data, textStatus) {
                        if(data.status == 401) {
                            setTimeout(function(){ location.reload(); }, 1000);
                        }
                        new PNotify({
                            title: 'Error',
                            text: "Error " + data.status + " " + data.statusText,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                        });
                    }  
                });
            } else {
                var msg = 'Required fields is empty';
                new PNotify({
                    title: 'Error',
                    text: msg,
                    addclass: 'stack-bar-top',
                    type: 'error',
                    width: "100%"
                });
            }
        });
        
        // Двойной клик для редактирования группы
        $('#dt-groups tbody').on('dblclick', 'tr', function (e) {
            var data = tableGroups.row(this).data();   
            var id_record = data.DT_RowId;
            id_record = id_record.replace(/row_/, "");
            
            $.magnificPopup.open({
                type: 'inline',
                items: {
                    src: '#modalForm-edit-group'
                },
                preloader: false,
                focus: '#name',
                modal: true,
                closeOnContentClick: false,
                closeOnBgClick: false,
                callbacks: {
                    beforeOpen: function() {
                        if($(window).width() < 700) {
                            this.st.focus = false;
                        } else {
                            this.st.focus = '#name';
                        }
                        
                        // Сброс полей формы
                        $("form#edit-group-form").trigger('reset');
                        $("#edit-group-form").find('input, textarea, select').each(function(e, elements) {
                            if(elements.required === true) {
                                $(elements).removeClass("err");
                            }
                        });

                        // Сброс всех checkbox
                        var cells = tableGroups.column(0).nodes();
                        var state = false;
                        for (var i = 0; i < cells.length; i += 1) {
                            cells[i].querySelector("input[type='checkbox']").checked = state;
                        }
                        
                        // Выделение checkbox выбранной строки
                        var data = tableGroups.columns().rows().nodes();
                        data.toArray().forEach(function(item, i, arr) {
                            var ids = item.id.replace(/row_/, "");
                            if(ids === id_record) {
                                var inp = item.firstChild.firstChild;
                                $(inp).prop('checked', "checked");
                            }
                        });
          
                        // Загрузка данных группы
                        $.ajax({
                            url: "/groups/ajax_getid_group",
                            type: "POST",
                            dataType: "json",
                            data: {'id': id_record},
                            success: function(response) {
                                if(response.error !== false && response.error !== '') {
                                    new PNotify({
                                        title: 'Error',
                                        text: response.error,
                                        type: 'error',
                                        addclass: 'stack-bar-top',
                                        width: "100%"
                                    });
                                }
                                if(response.data) {
                                    for (var key in response.data) {
                                        var ins = '#edit_group_' + key;
                                        var dat = response.data[key];
                                        
                                        if ($(ins).length) {
                                            if($(ins).prop('nodeName').toLowerCase() === 'input') {
                                                $(ins).val(dat);
                                            }
                                            if($(ins).prop('nodeName').toLowerCase() === 'select') {
                                                $(ins + ' option').prop('selected', false);
                                                $(ins + ' option[value="' + dat + '"]').prop('selected', true);
                                            } 
                                            if($(ins).prop('nodeName').toLowerCase() === 'textarea') {
                                                $(ins).val(dat);
                                            } 
                                        }
                                    }
                                }
                            },
                            error: function (data, textStatus) {
                                if(data.status == 401) {
                                    setTimeout(function(){ location.reload(); }, 1000);
                                }
                                new PNotify({
                                    title: 'Error',
                                    text: "Error " + data.status + " " + data.statusText,
                                    type: 'error',
                                    addclass: 'stack-bar-top',
                                    width: "100%"
                                });
                                $.magnificPopup.close();
                            }  
                        });
                    }
                }       
            });
        });
       
        // Кнопка редактирования группы
        $('#edit_group_button').on('click', function(e) {
            e.preventDefault();
            var isNotValid = false;
            
            // Валидация обязательных полей
            $("#edit-group-form").find('input, textarea, select').each(function(e, elements) {
                if(elements.required === true) {
                    if(elements.value === null || elements.value === '') {
                        $(elements).addClass("err");
                        isNotValid = true;
                    } else {
                        $(elements).removeClass("err");
                    }
                } else {
                    $(elements).removeClass("err");
                }
            });

            if(isNotValid === false) {
                var formData = new FormData();
                var data = $('#edit-group-form').serializeArray();
                
                $.each(data, function(key, input) {
                    formData.append(input.name, input.value);  
                });
                
                $.ajax({
                    url: "/groups/ajax_edit_group",
                    type: "POST",
                    data: formData,
                    processData: false,
                    contentType: false,
                    success: function(response) {
                        if(response.error !== false && response.error !== '') {
                            new PNotify({
                                title: 'Error',
                                text: response.error,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                        } else if(response.success === true) {
                            new PNotify({
                                text: 'Group updated',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                            });
                            $.magnificPopup.close();
                            tableGroups.draw();
                            $("form#edit-group-form").trigger('reset');
                        }
                    },
                    error: function (data, textStatus) {
                        if(data.status == 401) {
                            setTimeout(function(){ location.reload(); }, 1000);
                        }
                        new PNotify({
                            title: 'Error',
                            text: "Error " + data.status + " " + data.statusText,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                        });
                    }  
                });
            } else {
                var msg = 'Required fields is empty';
                new PNotify({
                    title: 'Error',
                    text: msg,
                    addclass: 'stack-bar-top',
                    type: 'error',
                    width: "100%"
                });
            }
        });
    }
});

/**
 * Функция поиска по колонкам для пользователей
 * Вызывается при нажатии Enter в поле поиска
 */
/**
 * Функция поиска по колонкам для пользователей
 * Вызывается при нажатии Enter в поле поиска
 * 
 * Логика работы:
 * 1. Проходим по всем заголовкам колонок таблицы
 * 2. Для каждого заголовка находим поле ввода (input)
 * 3. Извлекаем индекс колонки из атрибута name (формат: "name@index")
 * 4. Устанавливаем значение поиска для этой колонки через DataTables API
 * 5. Вызываем draw() для отправки запроса на сервер
 */
function keysearchUser(event) { 
    if(event.keyCode === 13) {
        event.preventDefault();
        var table = $('#dt-user').DataTable();
        if (!table) {
            console.error("DataTables for users is not initialized.");
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
                
                // Формат name: "u_login@2" -> индекс колонки = 2
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

/**
 * Функция поиска по колонкам для групп
 * Вызывается при нажатии Enter в поле поиска
 * 
 * Логика работы:
 * 1. Проходим по всем заголовкам колонок таблицы
 * 2. Для каждого заголовка находим поле ввода (input)
 * 3. Извлекаем индекс колонки из атрибута name (формат: "name@index")
 * 4. Устанавливаем значение поиска для этой колонки через DataTables API
 * 5. Вызываем draw() для отправки запроса на сервер
 */
function keysearchGroup(event) { 
    if(event.keyCode === 13) {
        event.preventDefault();
        // Используем глобальную переменную tableGroups
        var table = window.tableGroups || $('#dt-groups').DataTable();
        if (!table) {
            console.error("DataTables for groups is not initialized.");
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
                
                // Формат name: "g_name@2" -> индекс колонки = 2
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
