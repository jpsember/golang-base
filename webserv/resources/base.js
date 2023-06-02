function pr() {
    let s = ""
    for (let i = 0; i < arguments.length; i++) {
        let a = arguments[i]
        const t = typeof a
        if (t == "object") {
            a = "\n" + JSON.stringify(a, null, 2) + "\n"
        }
        if (i != 0) {
            s += " "
        }
        s += String(a)
    }
    console.log(s)
}

function ajax(id) {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            document.getElementById(id).innerHTML = this.responseText;
        }
    };
    xhttp.open("GET", "ajax", true);
    xhttp.send();
}

function processServerResponse(text) {
    pr("processServerResponse:", text)
    if (text.length == 0) {
        return
    }
    const obj = JSON.parse(text)
    pr("parsed to JSON object:", obj)
    if ('w' in obj) {
        const widgetMap = obj.w
        for (const [id, markup] of Object.entries(widgetMap)) {
            const qid = '#' + id
            $(qid).replaceWith(markup);
        }
    }
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
    const qid = '#' + id
    const textValue = $(qid).val();
    const xhttp = new XMLHttpRequest();
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + '/ajax');
    url.searchParams.set('w', id);         // The widget id
    url.searchParams.set('v', textValue);	 // The new value
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('GET', url);
    xhttp.send();
}
