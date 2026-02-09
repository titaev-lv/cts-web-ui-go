$(document).ready(function() {
    $('#update_coins_button').on('click', function(e) {
        $('#result_update_coins').html('<img src="/assets/images/loading2.gif" style="padding-right:10px">');
        $.ajax({
            url: "/coins/ajax_update_coins.php",
            type: "POST", 
            dataType: "html",
            data: {'action':"update"},
            success: function(response) { //Данные отправлены успешно
                var ret = JSON.parse(response);
                $('#result_update_coins').html('');
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
                            text: 'Coins updated',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                    $('#result_update_coins').html(ret.data);
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
    });
     $('#gen_icon_list_coins_button').on('click', function(e) {
        alert(2);
    });
});