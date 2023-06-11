function warning() {
    // TODO: this isn't doing quite what I want
    pr(...["*** WARNING:",console.trace()].concat(arguments))
}

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
            elem = document.getElementById(id);
            if (elem == null) {
              warning("can't find element with id:",id);
              return;
            }
            elem.innerHTML = this.responseText;
            warning("changed innerHTML of elem with id:",id)
            pr("changed innerHTML...")
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
            elem = document.getElementById(id);
            if (elem == null) {
              warning("can't find element with id:",id);
              continue;
            }
            elem.outerHTML = markup;
        }
    }
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
    // see https://tobiasahlin.com/blog/move-from-jquery-to-vanilla-javascript
    // to add back in some useful jquery functions
    x = document.getElementById(id);
    const textValue = x.value;
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

