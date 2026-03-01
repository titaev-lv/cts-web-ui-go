$(document).ready(function() {

    function trimTrailingZeros(value) {
        if (typeof value !== 'string' || value.indexOf('e') !== -1 || value.indexOf('E') !== -1) {
            return value;
        }
        return value
            .replace(/(\.\d*?[1-9])0+$/,'$1')
            .replace(/\.0+$/,'')
            .replace(/\.$/,'');
    }

    function formatAdaptivePrice(value) {
        if (value === null || value === undefined || value === '') {
            return '—';
        }

        var numeric = Number(String(value).replace(/\s+/g, '').replace(',', '.'));
        if (!isFinite(numeric)) {
            return String(value);
        }

        var abs = Math.abs(numeric);
        var formatted;

        if (abs === 0) {
            return '0';
        }

        if (abs >= 1) {
            formatted = numeric.toFixed(4);
        } else if (abs >= 0.000001) {
            formatted = numeric.toPrecision(6);
        } else {
            formatted = numeric.toExponential(3);
        }

        return trimTrailingZeros(formatted);
    }

    function parseAjaxResponse(response) {
        if (response && typeof response === 'object') {
            return response;
        }
        if (typeof response === 'string') {
            return JSON.parse(response);
        }
        return {};
    }

    /*
    * 1. Get Position table 
    */  
    var dtpos = document.getElementById('dt-positions');
    if(dtpos) {
        //Insert into header table input field for search
        var nm = Array(
                "",
                "u_id", 
                "u_login",
                "u_active",         
                "u_email",
                "u_name",
                "u_date_create",
                "u_date_modify",
                "u_timestamp",
                );
        /*$('#dt-positions thead tr th').each(function (i) {
            var title = $(this).text();
            if(i>0) {
                $(this).html(title+' <input type="text" name="'+nm[i]+'@'+i+'" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchUser(event)" />');
            }
        });*/
        	
        $.fn.dataTable.ext.errMode = 'throw';
        
        /*$('#dt-positions thead tr')
        .clone(true)
        .addClass('filters')
        .appendTo('#dt-user thead');*/
        
        table =  $('#dt-positions').DataTable( {
            "processing": true,
            "serverSide": true,
            "searching": true,
            "ordering": false,
            "scrollX": false,
            "pageLength": 50,
            "lengthMenu": [15,30, 50, 100 ],
            "pagingExtraNumberForNext": true,
           /* "bAutoWidth": false,*/
            "bScrollCollapse": true,
            //"rowId": 'id',
            "columns": [
                //{ "data":null, render:function(){return "<input type='checkbox' class='t-row chbx-ch' value=''/>";}}, //0
                { "data": "POSITION_ID" },    //1             
                { "data": "CONTRACT_NAME"},            //2
                { "data": "EXCHANGE_NAME"},   //3
                { "data": "MARKET_TYPE"},     //4
                { "data": "STATUS"},          //5
                { "data": "FINAL_POSITION"},  //6
                {
                    "data": "FINAL_AVG_PRICE",   //7
                    "render": function(data, type) {
                        if (type !== 'display') {
                            return data;
                        }
                        return formatAdaptivePrice(data);
                    }
                },
                { "data": "FEE_BASE_TOTAL"},  //8
                { "data": "FEE_TOTAL"},       //9
                { "data": "FUNDING_TOTAL"},   //10
                { 
                    "data": "TOTAL_REALIZED_PNL",    //11
                    "render": function(data, type, row) {
                        if (type === 'display') {
                            var value = parseFloat(data);
                            if (isNaN(value)) {
                                return data;
                            }
                            var formattedValue = value.toFixed(2);
                            if (value > 0) {
                                return '<span style="color: green; font-weight: bold; text-align: right; display: block;">+' + formattedValue + '</span>';
                            } else if (value < 0) {
                                return '<span style="color: red; font-weight: bold; text-align: right; display: block;">' + formattedValue + '</span>';
                            } else {
                                return '<span style="text-align: right; display: block;">' + value + '</span>';
                            }
                        }
                        return data;
                    }
                }
            ],
            //"pagingType": "first_last_numbers",
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
                "url": "/positions_calc/ajax_get_positions.php",
                "data": function ( d ) {
                    //d.filterMyInWork = $('#button-filter-my-in-work').val();
                    //d.filterMy = $('#button-filter-my').val();
                },
                "error": function (xhr, error, thrown) {
                },
                /*"success": function (xhr, error, thrown) {
                   console.log('');
                },*/
                "statusCode": {
                    401: function (xhr, error, thrown) {
                        location.reload();
                    }
                }
            },
            "createdRow":function  ( row, data, dataIndex, cells ) {
            },
            "columnDefs": [
                    {
                        "targets": "_all",
                        "className": 'dt-body-left',
                        "searchable": true
                    },
                    {
                        "searchable": false, "orderable": false, "visible": true,
                        "targets":  [0]
                    },
            ],
            "select": {
                "style":    'os',
                "selector": 'td:first-child'
            },
            "order": [[ 1, 'desc' ]],
            "drawCallback": function( settings ) {
            },     
        } );
        //Hide field search
        var search = document.getElementById('dt-positions_filter');
        if (search) {
              document.getElementById('dt-positions_filter').style.display = 'none';
        } 
        
        //Highlight row on mouse hover
        $('#dt-positions tbody').on('click', 'tr', function () {
        if ($(this).hasClass('selectedd')) {
            $(this).removeClass('selectedd');
        } else {
            table.$('tr.selectedd').removeClass('selectedd');
            $(this).addClass('selectedd');
        }
        }); 
        $('#dt-positions tbody').on('mouseover', 'tr', function () {
        if ($(this).hasClass('hoveredd')) {
        } else {
            $(this).addClass('hoveredd');
        }
        });
        $('#dt-positions tbody').on('mouseout', 'tr', function () {
            if ($(this).hasClass('hoveredd')) {
                $(this).removeClass('hoveredd');
            } 
        });
        
        //Double click 
        $('#dt-positions tbody').on( 'click', 'tr', function (e) {
            var data = table.row(this).data();   
            var id_record = data.POSITION_ID;
            location.href = '/positions_calc/position/?position='+parseInt(id_record);
        });
        
        //Checkbox - select all
        /*$('#checkallPositions').on('click', function() {
            var cells = table.column(0).nodes(), // Cells from 1st column
              state = this.checked;
            for (var i = 0; i < cells.length; i += 1) {
                cells[i].querySelector("input[type='checkbox']").checked = state;
            }
        });*/
    }
        
    //Create Position
    $('#add_position_button').on('click', function(e) {
        e.preventDefault();
        var isNotValid = false;
        $("#add-position-form").find('input, textarea, select').each(function(e,elements) {
            if(elements.required === true) {
                if(elements.value === null || elements.value === '') {
                    $(elements).addClass("err");
                    isNotValid = true;
                }
                else {
                    $(elements).removeClass("err");
                }
            }
        });
        if(isNotValid === false) {
            var formData = new FormData();
            var data = $('#add-position-form').serializeArray();
            $.each(data,function(key,input){
               formData.append(input.name,input.value);   
            });
            $.ajax({
                url: "/positions_calc/ajax_create_position.php",
                type: "POST", 
                data:formData,
                processData: false,
                contentType: false,
                /*beforeSend: function(xhr) {
                    xhr.setRequestHeader("Content-type", "multipart/form-data");
                },*/
                success: function(response) { //Данные отправлены успешно
                    var ret = parseAjaxResponse(response);
                    //console.log(ret);
                    if(ret.error !== false && ret.error !== '') {
                        new PNotify({
                                title: 'Error',
                                text: ret.error,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                        });
                    }
                    else if(ret.success === true) {
                        new PNotify({
                            /*title: 'OK',*/
                            text: 'Position added',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                        });
                        $.magnificPopup.close();
                        //$('#add-position-form').trigger("reset");
                        table.draw();
                        //setTimeout(function(){ location.reload(); }, 2000);
                        //сброс полей формы
                        $("form#add-position-form").trigger('reset');
                    }
                },
                error: function (data, textStatus) {
                    if(data.status == 401) {
                        setTimeout(function(){ location.reload(); }, 800);
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
        }
        else {
            msg =  'Required fiels is empty';
            new PNotify({
                title: 'Error',
                text: msg,
                addclass: 'stack-bar-top',
                type: 'error',
                width: "100%"
            });
        }
    });
});