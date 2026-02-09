$(document).ready(function() {
    //read get parameter position ID
    const params = new URLSearchParams(window.location.search);
    var position_id = parseInt(params.get("position"));
    
    getPosition(position_id);
    
    var dttrans = document.getElementById('dt-trans');
    if(dttrans) {
        var nm = Array(
                "",
                "u_id", 
                "u_type",
                "u_price",         
                "u_volume",
                "u_fee_base",
                "u_fee",
                "u_funding",
                "u_trans_date",
        );
        /*$('#dt-trans thead tr th').each(function (i) {
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
        
        table =  $('#dt-trans').DataTable( {
            "processing": true,
            "serverSide": true,
            "searching": true,
            "ordering": false,
            "scrollX": false,
            "pageLength": 50,
            "lengthMenu": [15,30, 50, 100, 500, 1000, 5000 ],
            "pagingExtraNumberForNext": true,
           /* "bAutoWidth": false,*/
            "bScrollCollapse": true,
            //"rowId": 'id',
            "columns": [
                { "data":null, render:function(){return "<input type='checkbox' class='t-row' value=''/>";}}, //0
                { "data": "ID" },    //1             
                { "data": "TYPE"},            //2
                { "data": "PRICE"},   //3
                { "data": "VOLUME"},     //4
                { "data": "FEE_BASE"},          //5
                { "data": "FEE"},  //6
                { "data": "FUNDING"},   //7
                { "data": "TRANS_DATE"}  //8
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
                "url": "/positions_calc/position/ajax_get_trans.php",
                "data": function ( d ) {
                    //d.filterMyInWork = $('#button-filter-my-in-work').val();
                    //d.filterMy = $('#button-filter-my').val();
                    d.position_id = position_id;
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
        var search = document.getElementById('dt-trans_filter');
        if (search) {
              document.getElementById('dt-trans_filter').style.display = 'none';
        } 
        
        //Highlight row on mouse hover
        $('#dt-trans tbody').on('click', 'tr', function () {
        if ($(this).hasClass('selectedd')) {
            $(this).removeClass('selectedd');
        } else {
            table.$('tr.selectedd').removeClass('selectedd');
            $(this).addClass('selectedd');
        }
        }); 
        $('#dt-trans tbody').on('mouseover', 'tr', function () {
        if ($(this).hasClass('hoveredd')) {
        } else {
            $(this).addClass('hoveredd');
        }
        });
        $('#dt-trans tbody').on('mouseout', 'tr', function () {
            if ($(this).hasClass('hoveredd')) {
                $(this).removeClass('hoveredd');
            } 
        });
        
        //Double click 
        $('#dt-trans tbody').on( 'click', 'tr', function (e) {
            var data = table.row(this).data();   
            var id_record = data.ID;

            //location.href = '/positions_calc/position/?position='+parseInt(id_record);
        });
        
        //Checkbox - select all
        $('#checkallTrans').on('click', function() {
            var cells = table.column(0).nodes(), // Cells from 1st column
            state = this.checked;
            for (var i = 0; i < cells.length; i += 1) {
                cells[i].querySelector("input[type='checkbox']").checked = state;
            }
            calcAVGPriceChecked();
        });
    }

});

let exchaange;

function getPosition(position_id) {
    if(position_id) {
        var formData = new FormData();
        formData.append('position_id', position_id);   
        $.ajax({
            url: "/positions_calc/position/ajax_get_position.php",
            type: "POST", 
            data:formData,
            processData: false,
            contentType: false,
            /*beforeSend: function(xhr) {
                xhr.setRequestHeader("Content-type", "multipart/form-data");
            },*/
            success: function(response) { //Данные отправлены успешно
                var ret = JSON.parse(response);
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
            
                    $('#p_position_id').text(ret.POSITION_ID);
                    $('#p_contract_name').text(ret.CONTRACT_NAME);
                    $('#import_trans_csv_contract_name').val(ret.CONTRACT_NAME);
                    $('#p_exchange_name').text(ret.EXCHANGE_NAME);
                    $('#p_market').text(ret.MARKET_TYPE);
                    $('#p_status').text(ret.STATUS);
                    $('#p_date_open').text(ret.OPENED);
                    $('#import_trans_csv_start_date').val(ret.OPENED);
                    $('#p_date_close').text(ret.CLOSED);
                    $('#p_amount').text(ret.AMOUNT);
                    if(ret.AVG_PRICE.length === 0) $('#p_avg_price').text("—");
                    else $('#p_avg_price').text(ret.AVG_PRICE);
                    $('#p_fee_base_curr').text(ret.FEE_BASE_CURR);
                    $('#p_fee_quote_curr').text(ret.FEE_QUOTE_CURR);
                    $('#p_funding').text(ret.FUNDING);
                    $('#p_total_realized_pnl').text(ret.TOTAL_REALIZED_PNL);
                    $('#p_trans_count').text(ret.TRANS_COUNT);
                    
                    if(ret.STATUS == 'OPEN') {
                        document.getElementById('close-pos-btn').style.setProperty('display','inline');
                    }
                    else {
                        document.getElementById('close-pos-btn').style.setProperty('display','none');
                    }
                     
                    if(ret.AMOUNT > 0) {
                        document.getElementById('p_amount').style.setProperty('color', 'green', 'important');
                    }
                    else if (ret.AMOUNT < 0) {
                        document.getElementById('p_amount').style.setProperty('color', 'red', 'important');
                    }
                    else {
                        document.getElementById('p_amount').style.setProperty('color', 'black', 'important');
                    }
                    if(ret.FEE_BASE_CURR > 0) {
                        document.getElementById('p_fee_base_curr').style.setProperty('color', 'red', 'important');
                    }
                    else {
                        document.getElementById('p_fee_base_curr').style.setProperty('color', 'black', 'important');
                    }
                    if(ret.FEE_QUOTE_CURR > 0) {
                        document.getElementById('p_fee_quote_curr').style.setProperty('color', 'red', 'important');
                        
                    }
                    else {
                        document.getElementById('p_fee_quote_curr').style.setProperty('color', 'black', 'important');
                    }
                    if(ret.FUNDING > 0) {
                        document.getElementById('p_funding').style.setProperty('color', 'green', 'important');
                        document.getElementById('p_funding').textContent = '+'+document.getElementById('p_funding').innerText;
                    }
                    else if (ret.FUNDING < 0) {
                        document.getElementById('p_funding').style.setProperty('color', 'red', 'important');
                        document.getElementById('p_funding').textContent = ''+document.getElementById('p_funding').innerText;
                    }
                    else {
                        document.getElementById('p_funding').style.setProperty('color', 'black', 'important');
                    }
                    if(ret.TOTAL_REALIZED_PNL > 0) {
                        document.getElementById('p_total_realized_pnl').style.setProperty('color', 'green', 'important');
                        document.getElementById('p_total_realized_pnl').textContent = '+'+document.getElementById('p_total_realized_pnl').innerText;
                    }
                    else if (ret.TOTAL_REALIZED_PNL < 0) {
                        document.getElementById('p_total_realized_pnl').style.setProperty('color', 'red', 'important');
                        document.getElementById('p_total_realized_pnl').textContent = ''+document.getElementById('p_total_realized_pnl').innerText;
                    }
                    else {
                        document.getElementById('p_total_realized_pnl').style.setProperty('color', 'black', 'important');
                    }
                    
                    // Check if exchange changed and reconnect if needed
                    const currentExchangeName = ret.EXCHANGE_NAME;
                    const currentMarket = ret.MARKET_TYPE;
                    const currentSymbol = ret.CONTRACT_NAME;
                    
                    if (exchange && (
                        exchange.constructor.name.toLowerCase() !== currentExchangeName.toLowerCase() ||
                        exchange.currentSymbol !== currentSymbol
                    )) {
                        console.log('Exchange or symbol changed, reconnecting...', {
                            old: exchange.constructor.name + '/' + (exchange.currentSymbol || 'unknown'),
                            new: currentExchangeName + '/' + currentSymbol
                        });
                        
                        // Close current connection
                        exchange.closeWS();
                        
                        // Create new exchange instance
                        exchange = ExchangeFactory.create(currentExchangeName, currentMarket);
                        
                        // Store current symbol for future comparisons
                        exchange.currentSymbol = currentSymbol;
                        
                        // Reconnect with new symbol if position is open
                        if (ret.STATUS === 'OPEN' && currentSymbol) {
                            exchange.fetchInitialPrice(currentSymbol);
                            exchange.connectWS(currentSymbol);
                        }
                    } else if (exchange && !exchange.currentSymbol) {
                        // Store symbol if not set yet
                        exchange.currentSymbol = currentSymbol;
                    }
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
}
//Обработка каждого выделения или снятия выделения строки (checkbox) для подсчета Сумма удержания с ТСП RUB
$('#dt-trans tbody').on('change', 'input[type="checkbox"]', function () {
    calcAVGPriceChecked();   
});
function calcAVGPriceChecked() {
    const data = table.columns().rows().nodes();
    const market = document.getElementById('p_market').innerText.trim();
    const dt = [];
    var pos = 0;
    var avg = 0;
    var rpnl = 0;
    var se = false;
    
    data.toArray().forEach(function(item, i, arr) {            
        if(item.cells !== undefined) {
            const trans_id = parseInt(item.cells[1].firstChild.nodeValue.trim());
            const trans_type = item.cells[2].firstChild.nodeValue.trim();
            const trans_price = parseFloat(item.cells[3].firstChild.nodeValue.trim());
            const trans_volume = parseFloat(item.cells[4].firstChild.nodeValue.trim());
            const trans_fee_base = parseFloat(item.cells[5].firstChild.nodeValue.trim());
            const trans_fee = parseFloat(item.cells[6].firstChild.nodeValue.trim());
            const trans_funding = parseFloat(item.cells[7].firstChild.nodeValue.trim());
            const inp = item.firstChild.firstChild;
            //Read selected checkbox
            var chk = $(inp).prop('checked');
            if(chk === true) {
                dt.push({ trans_id, trans_type, trans_price, trans_volume, trans_fee_base, trans_fee, trans_funding });
                se = true;
            }
        }
    });
    //sort
    dt.sort((a,b) => a.trans_id - b.trans_id);
    
    var i=0;
    for(const obj of dt) {
        let avg_prev = avg;
        switch (market) {
            case 'SPOT':
                if(obj.trans_volume > 0) {
                    if(i===0){
                        avg = (obj.trans_price * obj.trans_volume) / (obj.trans_volume - obj.trans_fee_base);
                    }
                    else {
                        if((pos + obj.trans_volume - obj.trans_fee_base) !== 0) {
                            avg = (pos * avg + (obj.trans_volume - obj.trans_fee_base)*((obj.trans_price * obj.trans_volume)/(obj.trans_volume - obj.trans_fee_base)) - rpnl) / (pos + obj.trans_volume - obj.trans_fee_base);
                        }
                        else {
                            avg = 0;
                        }
                    };
                    if((pos + obj.trans_volume - obj.trans_fee_base) === 0) {
                        rpnl = (obj.trans_price - avg_prev) * Math.min(Math.abs(obj.trans_volume), Math.abs(pos)) * Math.sign(pos);
                    }
                    else {
                        rpnl = 0;
                    }
                    pos = pos + (obj.trans_volume - obj.trans_fee_base);
                }
                else if(obj.trans_volume < 0){
                    if((pos+obj.trans_volume) !== 0) {
                        avg = (pos*avg + obj.trans_volume*obj.trans_price + obj.trans_fee) / (pos+obj.trans_volume);
                    }
                    else {
                        avg = 0;
                    }
                    
                    if(pos + obj.trans_volume === 0) {
                        rpnl = (obj.trans_price - avg_prev) * Math.min(Math.abs(obj.trans_volume), Math.abs(pos)) * Math.sign(pos) - obj.trans_fee;
                    }
                    else {
                        rpnl = 0;
                    }
                    pos = pos + obj.trans_volume;
                }
                break;
            case 'FUTURES':
            default:
                if(i===0){
                    if(obj.trans_type == 'TRADE') {
                        avg = ((obj.trans_price*obj.trans_volume + obj.trans_fee) / obj.trans_volume);                      
                        rpnl = 0;
                    }
                    else {
                        //funding
                        if(pos !== 0) {
                            avg = (pos * avg - obj.trans_funding)/pos;
                        }
                        else {
                             avg = 0;
                             rpnl = rpnl + obj.trans_funding;
                        }
                    }
                    
                }
                else {
                    if(obj.trans_type == 'TRADE') {
                        //THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + o.VOLUME*o.PRICE + o.FEE - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME,0)
                        if((pos+obj.trans_volume) !== 0) {
                            avg = (pos * avg + obj.trans_volume * obj.trans_price + obj.trans_fee - rpnl) / (pos + obj.trans_volume);
                            rpnl = 0;
                        }
                        else {
                            avg = 0;
                            rpnl = (obj.trans_price - avg_prev) * Math.min(Math.abs(obj.trans_volume), Math.abs(pos)) * Math.sign(pos)-obj.trans_fee;
                        }
                    }
                    else {
                        //funding
                        //THEN (prev.POS*prev.AVG_PRICE - o.FUNDING_AMOUNT) / NULLIF(prev.POS,0)
                        if(pos !== 0) {
                            avg = (pos * avg - obj.trans_funding)/pos;
                        }
                        else {
                            avg = 0;
                            rpnl = rpnl + obj.trans_funding;
                        }
                    }
                    
                    //WHEN (prev.POS + o.VOLUME) = 0
                    /*if(pos + obj.trans_volume == 0) {
                        //THEN (o.PRICE - prev.AVG_PRICE) * LEAST(ABS(o.VOLUME),ABS(prev.POS)) * SIGN(prev.POS) - o.FEE
                        rpnl = (obj.trans_price - avg_prev) * Math.min(Math.abs(obj.trans_volume), Math.abs(pos)) * Math.sign(pos);
                    }
                    else {
                        rpnl = 0;
                    }*/
                }
                pos = pos + obj.trans_volume;
        }
        i++;
    }
    if(se) {
        let insrt = 'Selected transactions: \
                        Position = <b><span id="pq_pos">'+pos+'</span></b>&nbsp;&nbsp;&nbsp; \
                        AVG Price = <b><span id="pq_avg">'+avg+'</span></b>&nbsp;&nbsp;&nbsp; \
                        PnL = <b><span id="pq_pnl"></span></b>';
        if(rpnl !== 0) {
            insrt = 'Selected transactions: Realized PnL = <b>'+rpnl+'</b>';
        }
        $('#pq').html(insrt);
    }
    else {
        $('#pq').text('');
    }
    
    const lastPriceInput = toNumberSafe(document.getElementById("p_last_price").textContent.trim());
    const lastPrice = toNumberSafe(lastPriceInput);
    exchange.calcAndRender(lastPrice);
   // console.log('Position='+pos.toFixed(8)+' Pavg='+avg.toFixed(8)+' pnl='+rpnl);
    /*
    if(count > 0) {
        $('#resultsum').css('display', 'block');
    }
    else {
         $('#resultsum').css('display', 'none');
    }*/
    /*
    sum_total = sum_total.toFixed(2);
    sum_total = sum_total.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
    sum_total = sum_total.replace(/\./, ",");
    
    $('#resultsumcount').text(count);
    $('#resultsumsum').text(sum_total + ' RUB');
    */
}

//Close Position Button
$('#close-pos-btn').on('click', function(e) {
    const params = new URLSearchParams(window.location.search);
    var position_id = parseInt(params.get("position"));
    
    $.ajax({
        url: "/positions_calc/ajax_close_position.php", 
        type: 'POST',
        dataType: 'html',
        data: 'position_id='+position_id,
        processData: false,
         /*beforeSend: function(xhr) {
            xhr.setRequestHeader("Content-type", "multipart/form-data");
        },*/
        success: function(response) { //Данные отправлены успешно
            var ret = JSON.parse(response);
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
                    text: 'Position closed',
                    type: 'success',
                    addclass: 'stack-bar-top',
                    width: "100%"
                });
                const params = new URLSearchParams(window.location.search);
                var position_id = parseInt(params.get("position"));
                getPosition(position_id);
                /*var pr = document.getElementById('p_last_price').innerText;
                if(pr > 0) {
                    setTimeout(1000,exchange.calcAndRender(pr));
                }*/
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
});

//Edit Position
$('.modal-with-form[href="#modalForm-edit-position"]').on('click', function(e) {
    e.preventDefault();
    
    // Get position data to populate the form
    const contractName = $('#p_contract_name').text();
    const exchangeName = $('#p_exchange_name').text();
    const dateStart = $('#p_date_open').text();
    
    // Populate the form fields
    $('#edit_position_name_contract').val(contractName);
    $("#edit_position_exchange option:contains("+exchangeName+")").attr('selected', true);
    $('#edit_position_date_start').val(dateStart);
    
    // Open the modal
    $.magnificPopup.open({
        items: [{
            src: '#modalForm-edit-position',
            type: 'inline',
            modal: true
        }],
        closeOnContentClick: false,
        closeOnBgClick: false,
        callbacks: {
            beforeOpen: function() {
                if($(window).width() < 700) {
                    this.st.focus = false;
                } else {
                    this.st.focus = '#edit_position_name_contract';
                }
            }
        }
    });
});

// Edit Position Submit
$('#edit_position_button').on('click', function(e) {
    e.preventDefault();
    
    const params = new URLSearchParams(window.location.search);
    const position_id = parseInt(params.get("position"));
    
    // Validate required fields
    const contractName = $('#edit_position_name_contract').val().trim();
    const exchangeId = $('#edit_position_exchange').val();
    const dateStart = $('#edit_position_date_start').val().trim();
    
    var isValid = validateEmptyFormFields('edit-position-form');
    
    if(isValid) {
        // Prepare data for AJAX
        const formData = {
            position_id: position_id,
            name_contract: contractName,
            exchange_id: exchangeId,
            date_start: dateStart
        };

        // Send AJAX request
        $.ajax({
            url: 'ajax_edit_position.php',
            type: 'POST',
            data: formData,
            dataType: 'json',
            success: function(response) {
                var ret = response;
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
                    $.magnificPopup.close();
                    new PNotify({
                        /*title: 'OK',*/
                        text: 'Position updated successfully',
                        type: 'success',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                    getPosition(position_id);
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
                $.magnificPopup.close();
            }  
        });
    }
});

//Delete Position
$('#del-pos-btn').on('click', function(e) {
    $.magnificPopup.open({
        items: [
            {
                src: '#modalDeletePos',
                type: 'inline',
                modal: true
            }],
        closeOnContentClick: false,
        closeOnBgClick:false,
        callbacks: {
            beforeOpen: function() {
                if($(window).width() < 700) {
                    this.st.focus = false;
                } else {
                    this.st.focus = '#delete_pos_confirm';
                }
            }
        }
    });
});
//Delete Position Confirm
$('#delete_pos_confirm').on('click', function(e) {
   const params = new URLSearchParams(window.location.search);
   let position_id = parseInt(params.get("position"));
   $.ajax({
        url: "/positions_calc/ajax_delete_position.php", 
        type: 'POST',
        dataType: 'html',
        data: 'position_id='+position_id,
        processData: false,
         /*beforeSend: function(xhr) {
            xhr.setRequestHeader("Content-type", "multipart/form-data");
        },*/
        success: function(response) { //Данные отправлены успешно
            var ret = JSON.parse(response);
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
                    text: 'Position deleted',
                    type: 'success',
                    addclass: 'stack-bar-top',
                    width: "100%"
                });
                setTimeout(function(){ location.href = '/positions_calc/'; }, 400);
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
   
});

//Edit Transaction

//Delete Transaction
$('#del-trans-btn').on('click', function(e) {
    // Check if any transactions are selected
    var selectedTransactions = [];
    var tableNodes = table.columns().rows().nodes();
    
    tableNodes.toArray().forEach(function(item) {
        if(item.cells !== undefined) {
            const checkbox = item.firstChild.firstChild;
            var isChecked = $(checkbox).prop('checked');
            if(isChecked === true) {
                const trans_id = parseInt(item.cells[1].firstChild.nodeValue.trim());
                selectedTransactions.push(trans_id);
            }
        }
    });
    
    if(selectedTransactions.length === 0) {
        new PNotify({
            title: 'Warning',
            text: 'Please select at least one transaction to delete',
            type: 'warning',
            addclass: 'stack-bar-top',
            width: "100%"
        });
        return;
    }
    
    $.magnificPopup.open({
        items: [
            {
                src: '#modalDeleteTrans',
                type: 'inline',
                modal: true
            }],
        closeOnContentClick: false,
        closeOnBgClick:false,
        callbacks: {
            beforeOpen: function() {
                if($(window).width() < 700) {
                    this.st.focus = false;
                } else {
                    this.st.focus = '#name';
                }
            }
        }
    });
});
//
//Delete Transaction Confirm
$('#delete_trans_confirm').on('click', function(e) {
    e.preventDefault();
    
    // Collect selected transaction IDs
    var selectedTransactionIds = [];
    var tableNodes = table.columns().rows().nodes();
    
    tableNodes.toArray().forEach(function(item) {
        if(item.cells !== undefined) {
            const checkbox = item.firstChild.firstChild;
            var isChecked = $(checkbox).prop('checked');
            if(isChecked === true) {
                const trans_id = parseInt(item.cells[1].firstChild.nodeValue.trim());
                selectedTransactionIds.push(trans_id);
            }
        }
    });
    
    if(selectedTransactionIds.length === 0) {
        new PNotify({
            title: 'Error',
            text: 'No transactions selected',
            type: 'error',
            addclass: 'stack-bar-top',
            width: "100%"
        });
        $.magnificPopup.close();
        return;
    }
    
    // Send DELETE request to server
    var formData = new FormData();
    formData.append('transaction_ids', JSON.stringify(selectedTransactionIds));
    
    const params = new URLSearchParams(window.location.search);
    var position_id = parseInt(params.get("position"));
    formData.append('position_id', position_id);
    
    $.ajax({
        url: "/positions_calc/position/ajax_delete_trans.php",
        type: "POST", 
        data: formData,
        processData: false,
        contentType: false,
        success: function(response) {
            var ret = JSON.parse(response);
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
                    text: 'Transactions deleted successfully',
                    type: 'success',
                    addclass: 'stack-bar-top',
                    width: "100%"
                });
                
                // Close modal and refresh data
                $.magnificPopup.close();
                table.draw(); // Refresh DataTable
                
                // Refresh position data
                getPosition(position_id);
                
                // Update PnL calculations if exchange is available
                if(typeof exchange !== 'undefined' && exchange) {
                    var lastPrice = document.getElementById('p_last_price').innerText;
                    if(lastPrice > 0) {
                        exchange.calcAndRender(lastPrice);
                    }
                }
                calcAVGPriceChecked();
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
            $.magnificPopup.close();
        }  
    });
});

//Create Transaction
$('#add_trans_button').on('click', function(e) {
    e.preventDefault();
    updateTransactionTypeOptions();
    var isNotValid = false;
    $("#add-trans-form").find('input, textarea, select').each(function(e,elements) {
        if(elements.required === true && elements.disabled === false) {
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
        var data = $('#add-trans-form').serializeArray();
        $.each(data,function(key,input){
           formData.append(input.name,input.value);   
        });
        $.ajax({
            url: "/positions_calc/position/ajax_create_trans.php",
            type: "POST", 
            data:formData,
            processData: false,
            contentType: false,
            /*beforeSend: function(xhr) {
                xhr.setRequestHeader("Content-type", "multipart/form-data");
            },*/
            success: function(response) { //Данные отправлены успешно
                var ret = JSON.parse(response);
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
                        text: 'Transaction added',
                        type: 'success',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                    $.magnificPopup.close();
                    //$('#add-position-form').trigger("reset");
                    table.draw();
                    //setTimeout(function(){ location.reload(); }, 2000);
                    //сброс полей формы
                    $("form#add-trans-form").trigger('reset');
                    $('#add_trans_type').val('');
                    updateTransactionTypeOptions();
                    updateTransactionAddFields();
                    const params = new URLSearchParams(window.location.search);
                    var position_id = parseInt(params.get("position"));
                    getPosition(position_id);
                    var pr = document.getElementById('p_last_price').innerText;
                    if(pr > 0) {
                        setTimeout(1000,exchange.calcAndRender(pr));
                    }
                    calcAVGPriceChecked();
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
    
// Функция для обновления опций транзакций
function updateTransactionTypeOptions() {
    var market = document.getElementById('p_market').innerText;
    switch(market) {
        case 'FUTURES':
            if ($('#add_trans_type').find('option[value="funding"]').length === 0) {
                $('#add_trans_type').append('<option value="funding">FUNDING</option>');
            }
            if ($('#add_trans_type').find('option[value="trade"]').length === 0) {
                $('#add_trans_type').append('<option value="trade">TRADE</option>');
            }
            break;
        case 'SPOT':
            if ($('#add_trans_type').find('option[value="trade"]').length === 0) {
                $('#add_trans_type').append('<option value="trade">TRADE</option>');
            }
            if ($('#add_trans_type').find('option[value="funding"]').length > 0) {
                 $('#add_trans_type').find('option[value="funding"]').remove();
            }
            break;
        default:
            $('#add_trans_type option[]').remove();
            $('#add_trans_type').append('<option value=""></option>');
    }
}

function updateTransactionAddFields() {
    //read selected type
    var type = $("#add_trans_type option:selected").val();
    if(type == 'funding') {
        $('#div_add_trans_funding').css("display","block");
        $('#add_trans_funding').prop('disabled', false);
    }
    else {
        $('#div_add_trans_funding').css("display","none");
        $('#add_trans_funding').prop('disabled', true);
    }
    if(type == 'trade') {
        var market = document.getElementById('p_market').innerText;
        $('#div_add_trans_action').css("display","block");
        $('#add_trans_action').prop('disabled', false);
        $('#div_add_trans_price').css("display","block");
        $('#add_trans_price').prop('disabled', false);
        $('#div_add_trans_volume').css("display","block");
        $('#add_trans_volume').prop('disabled', false);
        if(market == 'FUTURES') {
            $('#div_add_trans_fee_quote').css("display","block");
            $('#add_trans_fee_quote').prop('disabled', false);
            $('#div_add_trans_fee_base').css("display","none");
            $('#add_trans_fee_base').prop('disabled', true);
        }
        else if(market == 'SPOT') {
            selectSpotFees();
        }
    }
    else {
        $('#div_add_trans_action').css("display","none");
        $('#add_trans_action').prop('disabled', true);
        $('#div_add_trans_price').css("display","none");
        $('#add_trans_price').prop('disabled', true);
        $('#div_add_trans_volume').css("display","none");
        $('#add_trans_volume').prop('disabled', true);
        $('#div_add_trans_fee_quote').css("display","none");
        $('#add_trans_fee_quote').prop('disabled', true);
        $('#div_add_trans_fee_base').css("display","none");
        $('#add_trans_fee_base').prop('disabled', true);
    }
}

$('#add_trans_type').on('focus mousedown', function(e) {
    // Обновляем опции перед открытием списка
    updateTransactionTypeOptions();
});

$('#add_trans_type').on('change', function(e) {
    updateTransactionAddFields();
});


function selectSpotFees() {
    var market = document.getElementById('p_market').innerText;
    if(market == 'SPOT') {
        var action = $("#add_trans_action option:selected").val();
        if(action == 'buy') {
            $('#div_add_trans_fee_base').css("display","block");
            $('#add_trans_fee_base').prop('disabled', false);
            $('#div_add_trans_fee_quote').css("display","none");
            $('#add_trans_fee_quote').prop('disabled', true);
        }
        else {
            $('#div_add_trans_fee_base').css("display","none");
            $('#add_trans_fee_base').prop('disabled', true);
            $('#div_add_trans_fee_quote').css("display","block");
            $('#add_trans_fee_quote').prop('disabled', false);
        }
    }
}

//Create Transaction
$('#import_trans_csv_button').on('click', function(e) {
    e.preventDefault();
    var isNotValid = false;
    $("#import-trans-csv-form").find('input, textarea, select').each(function(e,elements) {
        if(elements.required === true && elements.disabled === false) {
            if(elements.type === 'file') {
                if(elements.value === null || elements.value === '') {
                    let ell = elements.parentElement;
                    while (ell) {
                        if (ell.matches('div')) {
                            $(ell).css("background-color", "#f48e70");
                            break; 
                        }
                        ell = ell.previousElementSibling;
                    }
                }
                else {
                    let ell = elements.parentElement;
                    while (ell) {
                        if (ell.matches('div')) {
                            $(ell).css("background-color", "#ffffff");
                            break; 
                        }
                        ell = ell.previousElementSibling;
                    }
                }
            } 
            else {
                if(elements.value === null || elements.value === '') {
                    $(elements).addClass("err");
                    isNotValid = true;
                }
                else {
                    $(elements).removeClass("err");
                }
            }
        }
    });
    if(isNotValid === false) {
        var formData = new FormData();
        var data = $('#import-trans-csv-form').serializeArray();
        $.each(data,function(key,input){
           formData.append(input.name,input.value);   
        });
        var f = $('#import_trans_csv_file')[0].files[0];
        formData.append('file', f);
        $.ajax({
            url: "/positions_calc/position/ajax_upload_trans_csv.php",
            type: "POST", 
            data:formData,
            processData: false,
            contentType: false,
            /*beforeSend: function(xhr) {
                xhr.setRequestHeader("Content-type", "multipart/form-data");
            },*/
            success: function(response) { //Данные отправлены успешно
                var ret = JSON.parse(response);
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
                        text: 'Loaded '+ret.data+" transactions",
                        type: 'success',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                    $.magnificPopup.close();
                    //$('#add-position-form').trigger("reset");
                    table.draw();
                    //setTimeout(function(){ location.reload(); }, 2000);
                    //сброс полей формы
                    $("form#import-trans-csv-form").trigger('reset');
                    const params = new URLSearchParams(window.location.search);
                    var position_id = parseInt(params.get("position"));
                    getPosition(position_id);
                    var pr = document.getElementById('p_last_price').innerText;
                    if(pr > 0) {
                        setTimeout(1000,exchange.calcAndRender(pr));
                    }
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

// ===== Helper для статуса =====
function setWSStatus(state) {
  const el = document.getElementById("ws_status");
  switch (state) {
    case "connected":
      el.textContent = "🟢 Connected";
      el.style.background = "#065f46"; // зелёный
      break;
    case "reconnecting":
      el.textContent = "🟡 Reconnecting...";
      el.style.background = "#92400e"; // оранжевый
      break;
    case "disconnected":
    default:
      el.textContent = "🔴 Disconnected";
      el.style.background = "#7f1d1d"; // красный
      break;
  }
}

// ================= Helper: ws status UI =================
function setWSStatus(state, label = null) {
  // states: connecting, connected, disconnected, reconnecting, error
  const el = document.getElementById("ws_status");
  if (!el) return;
  el.className = ""; // reset classes
  switch (state) {
    case "connecting":
      el.classList.add("connecting");
      el.textContent = "🟡 Connecting...";
      break;
    case "connected":
      el.classList.add("connected");
      el.textContent = "🟢 Connected";
      break;
    case "reconnecting":
      el.classList.add("reconnecting");
      el.textContent = "🟠 Reconnecting...";
      break;
    case "error":
      el.classList.add("error");
      el.textContent = "⚠️ Error";
      break;
    case "disconnected":
    default:
      el.classList.add("disconnected");
      el.textContent = "🔴 Disconnected";
      break;
  }
  if (label) el.textContent += " " + label;
}

// init default
setWSStatus("disconnected");

// ================= Heartbeat Manager =================
class HeartbeatManager {
  constructor(ws, options = {}) {
    this.ws = ws;
    this.pingInterval = options.pingInterval || null;
    this.pingMessage = options.pingMessage || null;
    this.autoPong = options.autoPong || false;
    this.timeout = options.timeout || 30000;
    this.debug = options.debug || false;
    this.lastMessageTime = Date.now();
    this.checker = null;
    this.pinger = null;
  }
  log(...args) { if (this.debug) console.log("[Heartbeat]", ...args); }
  start(onTimeout) {
    this.log("Запуск Heartbeat...");
    if (this.pingInterval && this.pingMessage) {
      this.pinger = setInterval(() => {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          const msg = typeof this.pingMessage === "function" ? this.pingMessage() : this.pingMessage;
          try { this.ws.send(msg); this.log("Отправлен ping:", msg); } catch(e){ this.log("Ping send failed", e); }
        }
      }, this.pingInterval);
    }
    this.checker = setInterval(() => {
      if (Date.now() - this.lastMessageTime > this.timeout) {
        this.log("Таймаут Heartbeat: нет входящих сообщений в " + this.timeout + "ms");
        this.stop();
        setWSStatus("reconnecting");
        try { this.ws.close(); } catch(e){}
        if (onTimeout) onTimeout();
      }
    }, Math.min(5000, this.timeout / 2));
  }
  stop() {
    this.log("Остановка Heartbeat");
    if (this.checker) { clearInterval(this.checker); this.checker = null; }
    if (this.pinger) { clearInterval(this.pinger); this.pinger = null; }
  }
  handleMessage(raw, onMessage) {
    this.lastMessageTime = Date.now();
    let data;
    try { data = JSON.parse(raw); } catch { return; }
    if (this.debug) this.log("Получено сообщение:", data);
    if (this.autoPong) {
      if (data.ping) {
        const pong = JSON.stringify({ pong: data.ping });
        try { this.ws.send(pong); this.log("Отправлен pong:", pong); } catch(e){}
      } else if (data.op === "ping") {
        const pong = JSON.stringify({ op: "pong", ts: Date.now() });
        try { this.ws.send(pong); this.log("Отправлен pong:", pong); } catch(e){}
      }
    }
    onMessage(data);
  }
}

// ================= utility =================
function toNumberSafe(v) {
  if (v == null) return NaN;
  if (typeof v === "number") return v;
  if (typeof v === "string") {
    if (v.trim() === "-" || v.trim() === "") return NaN;
    return Number(v.replace(/\s+/g, "").replace(",", "."));
  }
  if (typeof v === "object") {
    if ("price" in v) return toNumberSafe(v.price);
    if ("last" in v) return toNumberSafe(v.last);
    if ("c" in v) return toNumberSafe(v.c);
  }
  return Number(v);
}

// ================= Base Exchange =================
class Exchange {
  constructor(market) {
    this.market = (market || "SPOT").toUpperCase();
    this.ws = null;
    this.hb = null;
    this.fees = { spot: 0.001, futures: 0.0005 }; // default, override in subclasses
    this.reconnectDelay = 5000;
  }

  getTakerFee() {
    return this.market === "SPOT" ? this.fees.spot : this.fees.futures;
  }

  // resilient calc and render
  calcAndRender(lastPriceInput) {
    const lastPrice = toNumberSafe(lastPriceInput);
    if (!Number.isFinite(lastPrice)) {
      console.warn("[PNL] Некорректная цена:", lastPriceInput);
      return;
    }
    $("#p_last_price").text(lastPrice.toFixed(8));

    const avgText = $("#p_avg_price").text().trim();
    const avgPrice = toNumberSafe(avgText);
    const amount = toNumberSafe($("#p_amount").text().trim());

    // if avg not provided or '-', show '-'
    if (avgText === "" || avgText === "-" || !Number.isFinite(avgPrice) || !Number.isFinite(amount)) {
      $("#p_unrealized_pnl").text("—");
      $("#p_cost").text("—");
      document.getElementById("p_unrealized_pnl")?.style.removeProperty("color");
      return;
    }

    const fee = this.getTakerFee();
    const pnlValue = lastPrice * amount - avgPrice * amount - lastPrice * fee * amount;
    $("#p_unrealized_pnl").text(pnlValue.toFixed(8));
    const pnlEl = document.getElementById("p_unrealized_pnl");
    if (pnlValue > 0) pnlEl.style.setProperty("color", "green", "important");
    else if (pnlValue < 0) pnlEl.style.setProperty("color", "red", "important");
    else pnlEl.style.setProperty("color", "black", "important");
    
    const cost =  amount * lastPrice;
    $("#p_cost").text(Math.abs(cost.toFixed(8)));
    
    //calc selected
    if(document.getElementById("pq_pos") && document.getElementById("pq_avg")) {
        const pq_pos = toNumberSafe(document.getElementById("pq_pos").textContent.trim());
        const pq_avg = toNumberSafe(document.getElementById("pq_avg").textContent.trim());
        console.log(pq_pos);
        const pq_calcpnl = lastPrice * pq_pos - pq_avg * pq_pos - lastPrice * fee * pq_pos;
        $("#pq_pnl").text(pq_calcpnl.toFixed(8));
    } 
  }

  async fetchInitialPrice(symbol) { /* override per exchange */ }
  connectWS(symbol) { throw "Not implemented"; }

  // safe close & set UI
  closeWS() {
    if (this.hb) {
      try { this.hb.stop(); } catch(e){}
      this.hb = null;
    }
    if (this.ws) {
      try { this.ws.close(); } catch(e){}
      this.ws = null;
    }
    setWSStatus("disconnected");
  }

  // guard to avoid double connect
  isOpenOrConnecting() {
    return this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING);
  }
}

// ================= BINANCE =================
class Binance extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.001, futures: 0.0004 };
  }
  mapSymbol(symbol) { return symbol.replace("/", "").toLowerCase(); }
  async fetchInitialPrice(symbol) {
    const baseUrl = this.market === "SPOT"
      ? "https://api.binance.com/api/v3/ticker/price"
      : "https://fapi.binance.com/fapi/v1/ticker/price";
    const url = `${baseUrl}?symbol=${symbol.replace("/", "").toUpperCase()}`;
    try {
      const res = await fetch(url);
      const data = await res.json();
      if (data?.price) this.calcAndRender(parseFloat(data.price));
    } catch (e) { console.error("[Binance] REST error:", e); }
  }
  connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[Binance] WS already open/connecting"); return; }
    const mapped = this.mapSymbol(symbol);
    const url = this.market === "SPOT"
      ? `wss://stream.binance.com:9443/ws/${mapped}@ticker`
      : `wss://fstream.binance.com/ws/${mapped}@ticker`;

    setWSStatus("connecting");
    try { this.ws = new WebSocket(url); } catch(e) { console.error(e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      console.log("[Binance] ws open");
      setWSStatus("connected");
    };
    this.ws.onmessage = (msg) => {
      let d; try { d = JSON.parse(msg.data); } catch { return; }
      if (d?.c !== undefined) this.calcAndRender(parseFloat(d.c));
    };
    this.ws.onerror = (e) => {
      console.error("[Binance] ws error", e);
      setWSStatus("error");
    };
    this.ws.onclose = (ev) => {
      console.log("[Binance] ws close", ev);
      setWSStatus("disconnected");
      // do not auto-reconnect here — control logic manages open/close based on p_status.
    };
  }
}

// ================= BYBIT =================
class Bybit extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.001, futures: 0.00055 };
  }
  async fetchInitialPrice(symbol) {
    const category = this.market === "SPOT" ? "spot" : "linear";
    const url = `https://api.bybit.com/v5/market/tickers?category=${category}&symbol=${symbol}`;
    try {
      const res = await fetch(url);
      const data = await res.json();
      const price = data?.result?.list?.[0]?.lastPrice;
      if (price) this.calcAndRender(parseFloat(price));
    } catch (e) { console.error("[Bybit] REST error:", e); }
  }
  connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[Bybit] WS already open/connecting"); return; }
    const category = this.market === "SPOT" ? "spot" : "linear";
    const url = `wss://stream.bybit.com/v5/public/${category}`;

    setWSStatus("connecting");
    try { this.ws = new WebSocket(url); } catch(e) { console.error(e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      console.log("[Bybit] ws open, subscribing");
      setWSStatus("connected");
      this.ws.send(JSON.stringify({ op: "subscribe", args: [`tickers.${symbol}`] }));
    };
    this.ws.onmessage = (msg) => {
      if (!this.hb) this.hb = new HeartbeatManager(this.ws, { timeout: 60000, debug: false });
      this.hb.handleMessage(msg.data, (d) => {
        if (d.topic && d.data?.lastPrice) this.calcAndRender(parseFloat(d.data.lastPrice));
      });
    };
    this.ws.onerror = (e) => { console.error("[Bybit] ws error", e); setWSStatus("error"); };
    this.ws.onclose = (ev) => { console.log("[Bybit] ws close", ev); setWSStatus("disconnected"); };
    // start heartbeat after open
    this.hb = new HeartbeatManager(this.ws, { timeout: 60000, debug: false });
    this.hb.start(() => { /* onTimeout */ });
  }
}

// ================= KUCOIN =================
class Kucoin extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.001, futures: 0.0006 };
  }
  mapSymbol(symbol) {
    if (this.market === "SPOT") return symbol.replace("/", "-");
    return symbol.replace("/", "") + "M";
  }
  async fetchInitialPrice(symbol) {
    try {
      // Use server proxy to avoid CORS issues
      const formData = new FormData();
      formData.append('symbol', symbol);
      formData.append('market', this.market);
      
      const res = await fetch('/positions_calc/position/ajax_kucoin_price.php', {
        method: 'POST',
        body: formData
      });
      
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      
      const result = await res.json();
      if (result.error) {
        throw new Error(result.error);
      }
      
      if (result.price) {
        this.calcAndRender(parseFloat(result.price));
      }
    } catch (e) { 
      console.error("[Kucoin] REST error:", e); 
    }
  }
  async connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[Kucoin] WS already open/connecting"); return; }
    const mapped = this.mapSymbol(symbol);

    setWSStatus("connecting");
    try {
      // Use server proxy to avoid CORS issues
      const formData = new FormData();
      formData.append('market', this.market);
      
      const res = await fetch('/positions_calc/position/ajax_kucoin_token.php', {
        method: 'POST',
        body: formData
      });
      
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      
      const result = await res.json();
      if (result.error) {
        throw new Error(result.error);
      }
      
      const token = result.data.token;
      const instance = result.data.instanceServers[0];
      const url = `${instance.endpoint}?token=${token}&connectId=${Date.now()}`;
      this.ws = new WebSocket(url);
    } catch (e) { console.error("[Kucoin] bullet error", e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      console.log("[Kucoin] ws open, subscribing");
      setWSStatus("connected");
      this.ws.send(JSON.stringify({
        id: Date.now(),
        type: "subscribe",
        topic: `/market/ticker:${mapped}`,
        privateChannel: false,
        response: true
      }));
    };
    this.ws.onmessage = (msg) => {
      if (!this.hb) this.hb = new HeartbeatManager(this.ws, {
        pingInterval: 20000,
        pingMessage: () => JSON.stringify({ id: Date.now(), type: "ping" }),
        timeout: 60000,
        debug: false
      });
      this.hb.handleMessage(msg.data, (d) => {
        if (d.topic && d.data?.price) this.calcAndRender(parseFloat(d.data.price));
      });
    };
    this.ws.onerror = (e) => { console.error("[Kucoin] ws error", e); setWSStatus("error"); };
    this.ws.onclose = (ev) => { console.log("[Kucoin] ws close", ev); setWSStatus("disconnected"); };
    // start hb
    if (!this.hb) {
      this.hb = new HeartbeatManager(this.ws, {
        pingInterval: 20000,
        pingMessage: () => JSON.stringify({ id: Date.now(), type: "ping" }),
        timeout: 60000,
        debug: false
      });
    }
    this.hb.start(() => {/* onTimeout */});
  }
}

// ================= HTX (Huobi) =================
class Htx extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.002, futures: 0.0005 };
  }
  mapSymbol(symbol) { return symbol.replace("/", "").toLowerCase(); }
  async fetchInitialPrice(symbol) {
    const mapped = this.mapSymbol(symbol);
    const baseUrl = this.market === "SPOT"
      ? `https://api.huobi.pro/market/trade?symbol=${mapped}`
      : `https://api.hbdm.com/linear-swap-ex/market/trade?contract_code=${symbol}`;
    try {
      const res = await fetch(baseUrl);
      const data = await res.json();
      const price = data?.tick?.data?.[0]?.price;
      if (price) this.calcAndRender(parseFloat(price));
    } catch (e) { console.error("[HTX] REST error:", e); }
  }
  connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[HTX] WS already open/connecting"); return; }
    const mapped = this.mapSymbol(symbol);
    const url = this.market === "SPOT"
      ? "wss://api.huobi.pro/ws"
      : "wss://api.hbdm.com/linear-swap-ws";
    setWSStatus("connecting");
    try { this.ws = new WebSocket(url); } catch(e) { console.error(e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      console.log("[HTX] ws open, subscribing");
      setWSStatus("connected");
      const sub = this.market === "SPOT"
        ? { sub: `market.${mapped}.trade.detail`, id: Date.now() }
        : { sub: `market.${mapped}.trade.detail`, id: Date.now() };
      this.ws.send(JSON.stringify(sub));
    };
    this.ws.onmessage = (msg) => {
      if (!this.hb) this.hb = new HeartbeatManager(this.ws, { autoPong: true, timeout: 60000, debug: false });
      this.hb.handleMessage(msg.data, (d) => {
        if (d.tick?.data?.[0]?.price) this.calcAndRender(parseFloat(d.tick.data[0].price));
      });
    };
    this.ws.onerror = (e) => { console.error("[HTX] ws error", e); setWSStatus("error"); };
    this.ws.onclose = (ev) => { console.log("[HTX] ws close", ev); setWSStatus("disconnected"); };
    if (!this.hb) this.hb = new HeartbeatManager(this.ws, { autoPong: true, timeout: 60000, debug: false });
    this.hb.start(() => {/* onTimeout */});
  }
}

// ================= COINEX =================
class Coinex extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.001, futures: 0.0005 };
  }
  async fetchInitialPrice(symbol) {
    const url = this.market === "SPOT"
      ? `https://api.coinex.com/v1/market/ticker?market=${symbol.replace("/", "")}`
      : `https://api.coinex.com/perpetual/v1/market/ticker?market=${symbol.replace("/", "")}`;
    try {
      const res = await fetch(url);
      const data = await res.json();
      const price = data?.data?.ticker?.last;
      if (price) this.calcAndRender(parseFloat(price));
    } catch (e) { console.error("[Coinex] REST error:", e); }
  }
  connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[Coinex] WS already open/connecting"); return; }
    const url = this.market === "SPOT"
      ? "wss://socket.coinex.com/v1/spot"
      : "wss://perpetual.coinex.com/ws";
    setWSStatus("connecting");
    try { this.ws = new WebSocket(url); } catch(e) { console.error(e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      setWSStatus("connected");
      this.ws.send(JSON.stringify({
        method: "subscribe",
        params: [`market.${symbol.replace("/", "")}.ticker`],
        id: Date.now()
      }));
    };
    this.ws.onmessage = (msg) => {
      if (!this.hb) this.hb = new HeartbeatManager(this.ws, { pingInterval: 20000, pingMessage: JSON.stringify({ method: "ping" }), timeout: 60000, debug: false });
      this.hb.handleMessage(msg.data, (d) => {
        if (d.method === "ticker.update" && d.params?.[0]?.last) this.calcAndRender(parseFloat(d.params[0].last));
      });
    };
    this.ws.onerror = (e) => { console.error("[Coinex] ws error", e); setWSStatus("error"); };
    this.ws.onclose = (ev) => { console.log("[Coinex] ws close", ev); setWSStatus("disconnected"); };
    if (!this.hb) this.hb = new HeartbeatManager(this.ws, { pingInterval: 20000, pingMessage: JSON.stringify({ method: "ping" }), timeout: 60000, debug: false });
    this.hb.start(() => {/* onTimeout */});
  }
}

// ================= POLONIEX =================
class Poloniex extends Exchange {
  constructor(market) {
    super(market);
    this.fees = { spot: 0.0015, futures: 0.0005 };
  }
  mapSymbol(symbol) { return this.market === "SPOT" ? symbol.replace("/", "_") : symbol.replace("/", ""); }
  async fetchInitialPrice(symbol) {
    const mapped = this.mapSymbol(symbol);
    const url = this.market === "SPOT"
      ? `https://api.poloniex.com/markets/${mapped}/ticker24h`
      : `https://futures-api.poloniex.com/v1/ticker?symbol=${mapped}`;
    try {
      const res = await fetch(url);
      const data = await res.json();
      const price = data?.price || data?.last;
      if (price) this.calcAndRender(parseFloat(price));
    } catch (e) { console.error("[Poloniex] REST error:", e); }
  }
  connectWS(symbol) {
    if (this.isOpenOrConnecting()) { console.log("[Poloniex] WS already open/connecting"); return; }
    const mapped = this.mapSymbol(symbol);
    const url = this.market === "SPOT"
      ? "wss://ws.poloniex.com/ws/public"
      : "wss://futures-apiws.poloniex.com/ws/v1";
    setWSStatus("connecting");
    try { this.ws = new WebSocket(url); } catch(e) { console.error(e); setWSStatus("error"); return; }

    this.ws.onopen = () => {
      setWSStatus("connected");
      const sub = this.market === "SPOT"
        ? { event: "subscribe", channel: "ticker", symbols: [mapped] }
        : { op: "subscribe", args: [`ticker.${mapped}`] };
      this.ws.send(JSON.stringify(sub));
    };
    this.ws.onmessage = (msg) => {
      if (!this.hb) this.hb = new HeartbeatManager(this.ws, { autoPong: true, timeout: 60000, debug: false });
      this.hb.handleMessage(msg.data, (d) => {
        if (d.data?.[0]?.price) this.calcAndRender(parseFloat(d.data[0].price));
      });
    };
    this.ws.onerror = (e) => { console.error("[Poloniex] ws error", e); setWSStatus("error"); };
    this.ws.onclose = (ev) => { console.log("[Poloniex] ws close", ev); setWSStatus("disconnected"); };
    if (!this.hb) this.hb = new HeartbeatManager(this.ws, { autoPong: true, timeout: 60000, debug: false });
    this.hb.start(() => {/* onTimeout */});
  }
}

// ================= Factory =================
class ExchangeFactory {
  static create(exchange, market) {
    switch (exchange.toLowerCase()) {
      case "binance": return new Binance(market);
      case "bybit": return new Bybit(market);
      case "kucoin": return new Kucoin(market);
      case "htx": return new Htx(market);
      case "coinex": return new Coinex(market);
      case "poloniex": return new Poloniex(market);
      default: throw new Error("Unsupported exchange " + exchange);
    }
  }
}

// ================= Control logic =================
let exchange = null;
let statusObserver = null;

function initPNL() {
  const symbol = $("#p_contract_name").text().trim();
  const exchangeName = $("#p_exchange_name").text().trim();
  const market = $("#p_market").text().trim();
  
  // Check if we're on a position detail page
  if (!symbol || !exchangeName || !market) {
    //console.log("[PNL] Not on position detail page, skipping initialization");
    return;
  }
  
  const statusEl = document.getElementById("p_status");
  if (!statusEl) {
    console.warn("Element #p_status not found — creating default CLOSED");
    $("body").append('<span id="p_status" style="display:none">CLOSED</span>');
  }

  exchange = ExchangeFactory.create(exchangeName, market);

  // always do one REST initial fetch
  exchange.fetchInitialPrice(symbol);

  const handleStatus = () => {
    const st = $("#p_status").text().trim();
    if (st === "OPEN") {
      console.log("[PNL] status OPEN -> ensure WS up");
      // if already open/connecting - keep it
      if (!exchange.isOpenOrConnecting()) {
        exchange.closeWS();
        exchange.connectWS(symbol);
      } else {
        console.log("[PNL] WS already open/connecting, skipping connect");
      }
    } else {
      console.log("[PNL] status not OPEN -> ensure WS closed and do REST");
      exchange.closeWS();
      exchange.fetchInitialPrice(symbol);
    }
  };

  handleStatus();
  statusObserver = new MutationObserver(handleStatus);
  statusObserver.observe(document.getElementById("p_status"), { childList: true, subtree: true, characterData: true });
}

// start after small delay (DOM readiness)
setTimeout(initPNL, 500);