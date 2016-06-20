
	function activeTabToClipboard(element) {


		// CTRL+ALT+V in Excel
		// var text = $(activeTabContent).text(); // .html()
		// copyToClipboard(text);


		var arrLabels = [];
		var	selectorLabel = '#tab-body label';
		$( selectorLabel ).each(function() {
			// $( this ).css({"border-width":"4px",});
			console.log('label: -' + $( this ).text() + '- ');
			arrLabels.push($( this ).text());
		});


		var arrValues = [];
		var	selectorInput = '#tab-body input';
		$( selectorInput ).each(function() {
			var inputType = $( this ).attr('type');
			if ( inputType == 'button' ) {
			} else if ( typeof  inputType == 'undefined' ) {
			} else {
				// $( this ).css({"border-width":"4px",});
				console.log('inpt: -' + inputType + '-  -' + $( this ).val() + '- ');
				arrValues.push($( this ).val());
			}
		});

		var columnWise = "";
		var arrayLength = selectorLabel.length;
		for (var i = 0; i < arrayLength; i++) {

			if ( typeof  arrValues[i] === 'undefined' ) {
				arrValues[i] = "";
				if ( typeof  arrLabels[i] === 'undefined' ) {
					break;
				}			
			}			
			columnWise += arrLabels[i] + '\t' + arrValues[i] + '\n';
		}
		copyToClipboard(columnWise);



	}

	function copyToClipboard(text) {
		// display:none; => impossible
		var $temp = $("<textarea style=''></textarea>");
		$('body').append($temp);
		$temp.val(text).select();

		try {
			var successful = document.execCommand('copy');
			var msg = successful ? 'successful' : 'unsuccessful';
			console.log('copy result: ' + msg + ';  --' + text + '-- ');
		} catch (err) {
			console.log('unable to copy  --' + text + '-- ');
		}

		$temp.remove();
	}


