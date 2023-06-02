
function pr() {
	s = ""
	for (var i = 0; i < arguments.length; i++) {
		a = arguments[i]
		//console.log("arg:" + a)
		const t = typeof a
		if (t == "object") {
			a = "\n" + JSON.stringify(a, null, 2) + "\n"
			//console.log("changed arg to:" + a)
		}
		s = s + " " + String(a)
	}
    console.log(s.trim())
}

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
	pr("processServerResponse:" , text)
	const obj = JSON.parse(text)
	pr("received server response:", obj)
	if ('w' in obj) {
		pr("found a 'w'")
		const widgetMap = obj.w
		for (const [id, markup] of Object.entries(widgetMap)) {
			pr("attempting to replace",id,"with markup:",markup)
			const qid = '#' + id
			$(qid).replaceWith(markup);
		}
	}
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
	const qid = '#' + id
	pr("jsVal called with qid:",qid)
    const textValue = $(qid).val();
	const xhttp = new XMLHttpRequest();
	const addr = window.location.href.split('?')[0];
	const url = new URL(addr + '/ajax');
	url.searchParams.set('w', id);           // The widget id
	url.searchParams.set('v', textValue);	 // The new value
	xhttp.onreadystatechange = function() {
		pr("onreadstatchange, readState:",this.readyState,"status:",this.status)
		pr("responseText:",this.responseText)
		pr("this:",this.toString())
    	if (this.readyState == 4 && this.status == 200) {
			  processServerResponse(this.responseText)
    	}
  	};
  	xhttp.open('GET', url);
	xhttp.send();
}
