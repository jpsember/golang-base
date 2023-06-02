
function ajax(id) {
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
     document.getElementById(id).innerHTML = this.responseText;
    }
  };
  xhttp.open("GET", "ajax", true);
  xhttp.send();
}

function processServerResponse(text) {
	const obj = JSON.parse(text)
	console.log("received server response:")
	console.log(obj)
	if ('w' in obj) {
		console.log("found a w")
		const widgetMap = obj.w
		for (const [id, markup] of Object.entries(widgetMap)) {
			console.log("attempting to replace "+id+" with markup: "+markup)
			const qid = '#' + id
			$(qid).replaceWith(markup);
		}
	}
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
	var qid = '#' + id
    var textValue = $(qid).val();
	var xhttp = new XMLHttpRequest();
	var addr = window.location.href.split('?')[0];
	var url = new URL(addr + '/ajax');
	url.searchParams.set('w', id);           // The widget id
	url.searchParams.set('v', textValue);	 // The new value
	xhttp.onreadystatechange = function() {
    	if (this.readyState == 4 && this.status == 200) {
			  processServerResponse(this.responseText)
    	}
  	};
  	xhttp.open('GET', url);
	xhttp.send();
}
