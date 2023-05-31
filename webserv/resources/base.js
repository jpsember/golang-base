
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

// An onChange event has occurred within an input field;
// send result back to server
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
			// console.log("html for "+id+" : "+$(qid).html())
			$(qid).replaceWith(this.responseText);
    	}
  	};
  	xhttp.open('GET', url);
	xhttp.send();
}
