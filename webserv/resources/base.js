const request_key_widget = 'w'
const request_key_value = 'v'
const widget_id_info = '__info__'

function warning() {
    _alert("WARNING",_argsToArray(arguments))
}

function todo() {
    _alert("TODO",_argsToArray(arguments))
}

function _alert(type, args) {
    const x = ["***",type + ";", where(2), ":"].concat(args)
    pr(...x)
}

function _argsToArray(x) {
    const args = [];
    for (let i = 0; i < x.length; i++) {
        args[i] = x[i];
    }
    return args
}

function var_info(x) {
    return typeof(x) + ': "'+String(x)+'"'
}

function where(skip) {
    if (skip == undefined) {
        skip = 0
    }
    skip += 2
    const err = new Error();
    const lines = err.stack.split("\n")
    let x = lines[skip]
    if (x == null) {
        x = "<unknown location>"
    } else {
        x = x.replace(/^\s*at\s*/,"")
    }
    return x
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

function processServerResponse(text) {
    if (text.length == 0) {
        return
    }
    const obj = JSON.parse(text)
    if ('w' in obj) {
        const widgetMap = obj.w
        for (const [id, markup] of Object.entries(widgetMap)) {
            const elem = document.getElementById(id);
            if (elem == null) {
              warning("can't find element with id:",id);
              continue;
            }
            pr("replacing outerHTML for widget with id:",id)
            pr("and markup:",markup)
            elem.outerHTML = markup;
        }
    }
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
    // see https://tobiasahlin.com/blog/move-from-jquery-to-vanilla-javascript
    // to add back in some useful jquery functions

    // The Widget id has id "<id>"
    // The HTML element for the input field has id "<id>.aux"
    const auxId = id + '.aux'
    const x = document.getElementById(auxId);
    const textValue = x.value;
    const xhttp = new XMLHttpRequest();
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + '/ajax');
    url.searchParams.set(request_key_widget, id);         // The widget id
    url.searchParams.set(request_key_value, textValue);	 // The new value
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('GET', url);
    xhttp.send();
}

// An onClick event has occurred within a button
function jsButton(id) {
    const xhttp = new XMLHttpRequest();
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + '/ajax');
    url.searchParams.set(request_key_widget, id);         // The widget id
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('GET', url);
    xhttp.send();
}

// An click event has occurred within a checkbox
function jsCheckboxClicked(id) {
    var checkbox = document.getElementById(id+'.aux');
    const xhttp = new XMLHttpRequest();
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + '/ajax');
    url.searchParams.set(request_key_widget, id);         // The widget id
    url.searchParams.set(request_key_value, checkbox.checked.toString());	 // The new value
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('GET', url);
    xhttp.send();
}

function jsGetDisplayProperties() {
    const xhttp = new XMLHttpRequest();
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + '/ajax');

    const info = {
        "sw": window.screen.width,
        "sh": window.screen.height,
        "dp": window.devicePixelRatio
    }

    url.searchParams.set(request_key_widget, widget_id_info);
    url.searchParams.set(request_key_value, JSON.stringify(info));
    xhttp.open('GET', url);
    xhttp.send();
}

function getResolution() {
    pr("getResolution, screen width/height:",window.screen.width, window.screen.height)
    pr("window.screen.width:",window.screen.width)
    pr("device pixel ratio:",window.devicePixelRatio)
}

getResolution();
