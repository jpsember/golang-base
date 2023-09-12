const request_key_widget = 'w'
const request_key_value = 'v'
const request_key_info = 'i'

const _db = false && warning("db is true")

function pr() {
    const args = arguments
    let s = ""
    for (let i = 0; i < args.length; i++) {
        let a = args[i]
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

function db() {
    if (_db) {
        pr(...arguments)
    }
}

function warning() {
    return _alert("WARNING", ...arguments)
}

function todo() {
    return _alert("TODO",   ...arguments)
}

function _alert(type, ...args) {
    pr("***",type+";",where(2),":",...args)
    return true
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

function processServerResponse(text) {
    if (text.length == 0) {
        return
    }
    pr("processServerResponse, text:",text)
    const obj = JSON.parse(text)
    if ('w' in obj) {
        const widgetMap = obj.w
        for (const [id, markup] of Object.entries(widgetMap)) {
            const elem = document.getElementById(id);
            if (elem == null) {
              warning("can't find element with id:",id);
              continue;
            }
            elem.outerHTML = markup;
        }
    }
}

function makeAjaxCall(...args) {
    const addr = window.location.href.split('?')[0];
    const url = new URL(addr + 'ajax');
    for (let i = 0; i < args.length; i+=2) {
        url.searchParams.set(args[i],args[i+1])
    }
    const xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('GET', url);
    xhttp.send();
}

// An onChange event has occurred within an input field;
// send widget id and value back to server; process response
function jsVal(id) {
    db("jsVal",id)
    // see https://tobiasahlin.com/blog/move-from-jquery-to-vanilla-javascript
    // to add back in some useful jquery functions
    // The Widget id has id "<id>"
    // The HTML element for the input field has id "<id>.aux"
    const auxId = id + '.aux'
    const x = document.getElementById(auxId);
    makeAjaxCall(request_key_widget,id,request_key_value,x.value)
}

// An onchange event has occurred within a file upload
function jsUpload(id) {
    db("jsUpload",id)


//     function uploadFiles() {
 //const url = 'https://httpbin.org/post';


     const addr = window.location.href.split('?')[0];
    const url = new URL(addr + 'upload/' + id);
    // url.searchParams.
    // for (let i = 0; i < args.length; i+=2) {
    //     url.searchParams.set(args[i],args[i+1])
    // }
    const xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            processServerResponse(this.responseText)
        }
    };
    xhttp.open('POST', url);
    formElem = document.getElementById(id+'.form')
     const data = new FormData(formElem);
    xhttp.send(data);
//   const method = 'post';
//
//   const xhr = new XMLHttpRequest();
//

//
//   xhr.open(method, url);
//   xhr.send(data);
// }
//

  //  makeAjaxCall(request_key_widget, id)
}


// An onclick event has occurred within a button
function jsButton(id) {
    db("jsButton",id)
    makeAjaxCall(request_key_widget, id)
}
// An click event has occurred within a checkbox
function jsCheckboxClicked(id) {
    db("jsCheckboxClicked", id)
    const checkbox = document.getElementById(id+'.aux');
    makeAjaxCall(request_key_widget,id,request_key_value,checkbox.checked.toString())
}

function jsGetDisplayProperties() {
    const info = {
        "sw": window.screen.width,
        "sh": window.screen.height,
        "dp": window.devicePixelRatio
    }
    makeAjaxCall(request_key_info, JSON.stringify(info))
}

pr("...base.js has loaded")
