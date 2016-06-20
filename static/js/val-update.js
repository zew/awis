
	// using the enter key 
	// to trigger an update in any input element
	$( document ).ready(function() {

		$( "input" ).keydown(function(evt) {
			if (evt.which == 13) {
				evt.preventDefault();
				console.log("enter prevented")
				jsUpdateVal(evt.target);
			}
		});

		$( "input" ).blur(function(evt) {
			console.log("blur => update")
			jsUpdateVal(evt.target);
		});

	});



	// Instant update upon leaving input field
	function jsUpdateVal(src) {

		// check for dirty
		if (src.defaultValue == src.value) {
			console.log("unchanged");
			return;
		}
		src.defaultValue = src.value;
		

		// var param_id    = $(src).attr("name");
		var param_id    = $(src).attr("data-param-id");
		var val_id      = $(src).attr("data-val-id");
		var mode        = $(src).attr("data-mode");
		var val_val     = $(src).val();

		var msg = "sending param_id " + param_id + "; val_id " + val_id +  "; val_val "  + val_val + "; mode "  + mode ;
		console.log(msg);


		var url  = PREFIX + '/val-update';
		var data = {
			param_id:  param_id,
			val_id:    val_id,
			val_val:   val_val,
			mode:      mode,
		};
		var fctSucc = function(data) { 
			if(data.Status == 'success'){
				var MsgAttrChanged = "";
				if (data["InsertedId"]) {
					if($(src).attr("data-val-id") == "0"){
						$(src).attr("data-val-id",data.InsertedId);
						MsgAttrChanged =  " - data-val-id from 0 to " + data.InsertedId;
					}	
				}
				showMsg(src,data.MsgSh,data.MsgLg + MsgAttrChanged);
			}else if(data.Status == 'error'){
				showMsg(src,data.MsgSh,data.MsgLg);
			}else{
				showMsg(src,"error","empty json.status");
				showMsg(src,"error", data);
			}		
		}


		var fctDone = function(data) { 
			// console.log( "jsUpdateVal "+msg+" - success 2: " + data ); 
		}
		var fctFail = function() { 
			showMsg(src,"error","ajax fail");
		}
		// var fctError = function(XMLHttpRequest, textStatus, errorThrown) {
		// 	if (XMLHttpRequest.readyState == 4) {
		// 		showMsg(src,"jsUpdateVal - error - HTTP" + XMLHttpRequest.status + " - " +  XMLHttpRequest.statusText);
		// 	} else if (XMLHttpRequest.readyState == 0) {
		// 		showMsg(src,"jsUpdateVal - error - network - connection refused - access denied - CORS etc");
		// 	} else {
		// 		showMsg(src,"jsUpdateVal - error - completely unforeseen");
		// 	}
		// }


		$.ajax({
			type:     "POST",
			url:      url,
			data:     data,
			success:  fctSucc,
			timeout:  4000,
			// error:    fctError,
			// dataType: dataType
		}).done(fctDone).fail(fctFail);


	}

	// Appending an absolutely positioned bubble to the updated input field
	// Message
	function showMsg(src,msgSh,msgLg){
		if (msgLg === undefined)  {
			msgLg = "msgLg undefined";
		}
		console.log(msgLg);
		msgLg = msgLg.replace(/(?:\r\n|\r|\n)/g, '<br />');
		var attr = $(src).next().attr("data-update-msg");
		if (typeof attr !== typeof undefined && attr !== false) {
			$(src).next().html(msgSh);
			$(src).next().show();
			console.log('reused');
		} else {
			var bubbleHtml = "<span data-update-msg='1' onclick=\"this.style.display ='none';\" style='right:4px; top:-2px;s' class='ajax-msg-bubble' >"+msgSh+"</span>";
			$(bubbleHtml).insertAfter( $(src) );
			console.log('appended');
		}
		setTimeout(function(arg1) {
			$(arg1).css({display:'none'});
			// arg1.style.display ='none';
		}, 2220, $(src).next());		

	}