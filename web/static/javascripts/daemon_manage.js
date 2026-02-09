$(document).ready(function() {
    $('#daemon_check_button').on('click', function(e) {
        e.preventDefault();
        $('.darkness').css('display','block');
        $('.layer').css('display','block');   
        $.ajax({
            url: "/daemon/ajax_check_status.php",
            type: "POST", 
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
                     new PNotify({
                            /*title: 'OK',*/
                            text: 'Updated',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                }
                $('#daemon_status').html(ret.status);
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
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
                $('#daemon_status').html('');
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
            }  
        }); 
    });
    $('#daemon_start_button').on('click', function(e) {
        e.preventDefault();
        $('.darkness').css('display','block');
        $('.layer').css('display','block'); 
        $.ajax({
            url: "/daemon/ajax_start.php",
            type: "POST", 
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
                     new PNotify({
                            /*title: 'OK',*/
                            text: 'Daemon started',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                }
                $('#daemon_status').html(ret.status);
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
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
                $('#daemon_status').html('');
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
            }  
        });
    });
    $('#daemon_stop_button').on('click', function(e) {
        e.preventDefault();
        $('.darkness').css('display','block');
        $('.layer').css('display','block'); 
        $.ajax({
            url: "/daemon/ajax_stop.php",
            type: "POST", 
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
                     new PNotify({
                            /*title: 'OK',*/
                            text: 'Daemon stoped',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                }
                $('#daemon_status').html(ret.status);
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
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
                $('#daemon_status').html('');
                $('.darkness').css('display','none');
                $('.layer').css('display','none');
            }  
        });
    });
    
    if($('#daemon_check_button')) {
        $( "#daemon_check_button" ).click();   
    }
});

function getDaemonStatus () {
    var run = $('#daemon_status').text();
    if(run === 'ACTIVE') {
        $.ajax({
            url: "/daemon/ajax_daemon_stat.php",
            type: "POST", 
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
                    var diagram = ret.status;
                    //console.log(ret.status);
                    
                    $("#output").html('<pre class="mermaid">'+diagram+'</pre>');
                    
                    $("pre.mermaid").each(function(i, e) {
                        const containerId = `mermaid-${i}`;
                        renderDiagram(containerId, e);
                    });
                    
                   /* var canvas = document.getElementById('processes');
                    var ctx = canvas.getContext('2d');
                    
                    ctx.clearRect(0,0,1000,600);
                    
                    /*ctx.fillStyle = 'magenta';
                    ctx.fillRect(100,50,150,150);
                    ctx.fillStyle = 'blue';
                    ctx.fillRect(120,50,150,150);*/ 
                    
                    /*ctx.beginPath();
                    ctx.strokeStyle = '#111';
                    ctx.lineWidth = "1";
                    ctx.rect(100,100,100,100);
                    ctx.stroke();
                    ctx.fillStyle = '#EEE';
                    ctx.fill();
                    
                    ctx.beginPath();
                    ctx.moveTo(150,100);
                    ctx.lineTo(200,200);
                    ctx.lineCap = 'round'; //butt square
                    ctx.stroke();*/
                    
                    /* new PNotify({
                            text: 'Daemon stoped',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });*/
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
    }
    else {
        $("#output").html('');
    }
}
function calc_r (obj){
    var c = 1;
    if(obj.pid){
        //console.log('c='+c);
        var ch = obj.child;
        if(Array.isArray(ch)){
            if(obj.child.length > 0) {           
               obj.child.forEach(
                    (element) => {
                       var k = calc_r(element);
                       k = k + 1; 
                       if(k>c){
                           c = k;
                       }
                    }
               );
               
            }
        }
        else if(typeof ch === 'object') {
            for (const [key, value] of Object.entries(ch)) {
                 var k = calc_r(value);
                 k = k +1;
                 if(k > c){
                     c = k;
                 }
            }
        }
    }
    //console.log('return '+r);
    return c;
}