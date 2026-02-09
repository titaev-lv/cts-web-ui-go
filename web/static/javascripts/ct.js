$(function () {
    /*
    * 1. Auth request
    */
    // Sending AJAX request for auth
    $("#authForm").submit(function(e) {
        $.ajax ({
            url: '/auth/login',
            type: 'POST',
            data: $(this).serialize(),
            contentType: false,
            processData: false,
            dataType: 'html',
            beforeSend: function(xhr) {
                xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
            },
            success: function (response) {
                var ret = JSON.parse(response);
                //console.log(ret);
                if(ret.error !== false && ret.error !== '') {
                    new PNotify({
                            title: 'ERROR',
                            text: ret.error,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                }
                else if(ret.success === true) {
                    $(location).attr('href', '/');
                }
            },
            statusCode: {
                404: function(response) {
                    new PNotify({
                            title: 'ERROR 404',
                            text: 'Error 404 Not found. <br> Auth file not found!',
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                },
                401: function(response) {
                     new PNotify({
                            title: 'ERROR 401',
                            text: 'Error 401 Unauthorized! <br>Incorrect Login or Password',
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                }
            }   
        });
        
        return false;
    });
    
    /*Modal*/
    $('.modal-with-form').magnificPopup({
        type: 'inline',
        preloader: false,
        focus: '#name',
        modal: true,
        closeOnContentClick: false,
        closeOnBgClick:false,
        // When elemened is focused, some mobile browsers in some cases zoom in
        // It looks not nice, so we disable it:
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
    $('.modal-basic').magnificPopup({
                type: 'inline',
                preloader: false,
                modal: true
    });
    /*
    Modal Dismiss
    */ 
    $(document).on('click', '.modal-dismiss', function (e) {
        e.preventDefault();
        $.magnificPopup.close();
        $('#create-user-form').trigger("reset");
        $('#create-group-form').trigger("reset");
        //$('#edit-form').trigger("reset");
    });
    
});

$(document).ready(function() {
    
});

function validateEmptyFormFieldsExample(form) {
	var count = form.length;
	var empty_field = false; 
	var email_err = false;
	var inn_err = false;
	//radio-button имеют спецефическое выделение, т.к. выделяются не поля, а блок в виде таблицы.
	//Класс not_empty
	//В массиве перечисляем поля, которые должны быть обязательными. Это поля из всех форм ресурса.
	//radio-button должны быть заключены в таблицу с id=id+'_table'
        var radio_not_empty = new Array();
	var radio_not_empty_status_true = new Array();
	var count_radio = radio_not_empty.length;
	
	var checkbox_block = new Array(); //сюда добавляются id родительских блоков, в которых ch обязательны
	var checkbox_block_checked = new Array(); //сюда добавляются id родительских блоков, 
	//------------------
	for (var i=0; i<count;i++) {
	    //Обработка checkbox
	    if(form.elements[i].type == 'checkbox') {
		//Поиск среди checkbox элементов, имеющих класс js-require-group-once, это класс, указывающий, что должен быть выделен
		//хотя бы один элемент из группы
 		if(form.elements[i].className.indexOf('js-require-group-once') > -1) {
		    var block = document.getElementsByName(form.elements[i].name)[0].parentNode.parentNode.parentNode.parentNode.id;
		    if (checkbox_block.indexOf(block) == -1) {
			checkbox_block.push(block);
		    }
		    if(form.elements[i].checked === true) {
			if (checkbox_block_checked.indexOf(block) == -1) {
			    checkbox_block_checked.push(block);
			}
		    }
		}
	    }
	    //Обработка исключений RADIO
	    if(form.elements[i].type == 'radio') {
		if(document.getElementById(form.elements[i].name + '_table')) {
		    document.getElementById(form.elements[i].name + '_table').className = '';
		}
		for(var j=0;j<count_radio;j++){
		    if(form.elements[i].name == radio_not_empty[j]) {
		       if(form.elements[i].checked == true) {
			   radio_not_empty_status_true.push(form.elements[i].name); 
		       }
		    }
		}
	    }
	    //-----------------
	    //Обработка полей типа text, select
	    if(form.elements[i].classList.contains('js-form-field-require')) {
		el = document.getElementsByName(form.elements[i].name)[0];
		if(el.type == 'select-one') {
		    el = document.getElementsByName(form.elements[i].name)[0].parentNode;
		}
		var arr;
		arr = el.className.split(" ");

		if(form.elements[i].value.length === 0 || !form.elements[i].value.trim()) {
		    if (arr.indexOf('_error') == -1) {
			el.className += " " + '_error';
		    }
		    empty_field = true;
		}
		else {
		    if (arr.indexOf('_error') > 0 ) {
			el.className = el.className.replace(/_error/g, "" );
		    }
		}
		//Исключния, если родительский блок выключен display:none
		if(el.parentNode.style.display == 'none') {
		    empty_field = false;
		    if (arr.indexOf('_error') > 0 ) {
			el.className = el.className.replace(/_error/g, "" );
		    }
		}
	    }
	    //Обработка email
	    if(form.elements[i].classList.contains('js-email-require')) {
		if(form.elements[i].value.length > 0) {
		    el = document.getElementsByName(form.elements[i].name)[0];
		    var r = /^[_a-zA-Z0-9-]+(\.[_a-zA-Z0-9-]+)*@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*(\.[a-zA-Z]{2,4})$/;
		    var valid = (el.value.match(r) == null) ? false : true;
		    var arr;
		    arr = el.className.split(" ");
		    if(valid === false) {
			if (arr.indexOf('_error') == -1) {
			    el.className += " " + '_error';
			}
			email_err = true;
		    }
		    else {
			if (arr.indexOf('_error') > 0 ) {
			    el.className = el.className.replace(/_error/g, "" );
			}
		    }
		}
	    }
	    //Обработака ИНН
	    if(form.elements[i].classList.contains('js-inn')) {
		if(form.elements[i].value.length > 0) {
		    el = document.getElementsByName(form.elements[i].name)[0];
		    var inn_msg = '';
		    if (el.value.match(/\D/)) {
			inn_err = true;
			inn_msg = "MESSAGE_INN_ERR1";
		    }
		    else {
			var inn = el.value.match(/(\d)/g);

			if (inn.length == 10) {
				if(inn[9] != String(((2 * inn[0] + 4 * inn[1] + 10 * inn[2]+3 * inn[3] + 5 * inn[4] + 9 * inn[5]+4 * inn[6] + 6 * inn[7] + 8 * inn[8]) % 11) % 10) ) {
				    inn_err = true;
				    inn_msg = "MESSAGE_INN_ERR2";
				}
			}
			else if (inn.length == 12) {
				if(inn[10] != String(((7 * inn[0] + 2 * inn[1] + 4 * inn[2] +10 * inn[3] + 3 * inn[4] + 5 * inn[5] +9 * inn[6] + 4 * inn[7] + 6 * inn[8] +8 * inn[9]) % 11) % 10) || inn[11] != String(((
					3 * inn[0] + 7 * inn[1] + 2 * inn[2] +
					4 * inn[3] + 10 * inn[4] + 3 * inn[5] +
					5 * inn[6] + 9 * inn[7] + 4 * inn[8] +
					6 * inn[9] + 8 * inn[10]
					) % 11) % 10)) {
				    inn_err = true;
				    inn_msg = "MESSAGE_INN_ERR2";
			    }
			}
			else {
			    inn_err = true;
			    inn_msg = "MESSAGE_INN_ERR1";
			}
			if(inn_err == true) {
			    if (arr.indexOf('_error') == -1) {
				el.className += " " + '_error';
			    }
			}
			else {
			    if (arr.indexOf('_error') > 0 ) {
				el.className = el.className.replace(/_error/g, "" );
			    }
			}
		    }
		} 
	    }
	}   
	//Обработка radio
	var count_radio_not_empty_status_true = radio_not_empty_status_true.length;
	    for (var i=0; i<count_radio;i++) {
		var flag = false;
		for(var j=0;j<count_radio_not_empty_status_true;j++) {
		    if(radio_not_empty[i] == radio_not_empty_status_true[j]) {
			flag = true;
		    }
		}
		if(flag == false) {
		    document.getElementById(radio_not_empty[i] + '_table').className = 'not_empty';
		    empty_field = true;
		}
	    }
	//Обработка checkbox
	var count_checkbox_block = checkbox_block.length;
	var count_checkbox_block_checked = checkbox_block_checked.length;
	for (var i=0; i<count_checkbox_block;i++) {
	    var flag_ch = false;
	    document.getElementById(checkbox_block[i]).className = '';
	    for (var j=0; j<count_checkbox_block_checked;j++) {
		if(checkbox_block[i] == checkbox_block_checked[j]) {
		    flag_ch = true;
		}
	    }
	    if(flag_ch == false) {
		document.getElementById(checkbox_block[i]).className = 'not_empty';
		empty_field = true;
	    }
	}
	
	if(empty_field == true) {
	    document.getElementById("jerror").style.display = "block";
	    document.getElementById("jerror").innerText = "MESSAGE_EMPTY_FIELDS";
	    //console.log(document.getElementById("jerror"));
	    window.scrollTo(0, 250);
	}
	else if(email_err == true) {
	    document.getElementById("jerror").style.display = "block";
	    document.getElementById("jerror").innerText = "MESSAGE_ERROR_EMAIL_FIELDS";
	    window.scrollTo(0, 250);
	}
	else if(inn_err == true) {
	    document.getElementById("jerror").style.display = "block";
	    document.getElementById("jerror").innerText = inn_msg;
	    window.scrollTo(0, 250);
	}
	else {
	    document.getElementById("jerror").style.display = "none";
	    document.getElementById("jerror").innerText = "";
	    form.submit();
	}
	return true;
}

function validateEmptyFormFields(form_id) {
    var isNotValid = false;
    $("#"+form_id).find('input, textarea, select').each(function(e,elements) {
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
    return !isNotValid;
}

window.addEventListener("DOMContentLoaded", function() {
  [].forEach.call( document.querySelectorAll('.money'), function(input) {
    var keyCode;
    function mask(event) {
        //event.keyCode && (keyCode = event.keyCode);
        var val = this.value;
        var ns = val.replace(/[^0-9\.-]/g, '');
        this.value = ns;
    }

    input.addEventListener("input", mask, false);
    input.addEventListener("focus", mask, false);
    input.addEventListener("blur", mask, false);
    input.addEventListener("keydown", mask, false);

  });

});