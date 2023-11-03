const request_key_widget = 'w'
const request_key_value = 'v'
const request_key_info = 'i'

const _db = true && warning("db is true")

let id_with_focus = null;

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

function where(skip, count) {
    if (skip == undefined) {
        skip = 0
    }
    if (count == undefined) {
        count = 1
    }
    skip += 2
    const err = new Error();
    const lines = err.stack.split("\n")
    let y = ""
    for (i = 0; i < count; i++) {
        let x = lines[skip + i]
        if (x == null) {
            if (i == 0) {
                x = "<unknown location>"
            } else {
                x = ""
            }
            break
        } else {
            x = x.replace(/^\s*at\s*/, "")
        }
        if (i == 0) {
            y = x
        } else {
            y = y + "\n" + x
        }
    }
    return y
}

const respKeyWidgetsToRefresh = 'w'
const respKeyURLExpr = 'u'


function processServerResponse(text) {
    if (text.length == 0) {
        return
    }
    const obj = JSON.parse(text)
    //pr("procesServerResponse:",obj)
    if (respKeyWidgetsToRefresh in obj) {
        const widgetMap = obj.w
        for (const [id, markup] of Object.entries(widgetMap)) {
            const elem = document.getElementById(id);
            if (elem == null) {
              warning("can't find element with id:",id);
              continue;
            }

            // We would like to preserve the focus on this element, if it had it

            elem.outerHTML = markup;
        }

        // Now that we've rendered the requested elements, restore the focus (if that element still exists)
        // TODO: restore the select range as well
        const f = id_with_focus;
        if (f != null && f.url == window.location.href) {
            db("attempting to restore focus to:",f)
            const elem = document.getElementById(f.id)
            if (elem != null) {
                db("found element")
                elem.focus()
            }
        }
    }

    if (respKeyURLExpr in obj) {
        const url = window.origin + obj.u
        //pr("calling history.pushState with:",url,"because key was in the server response:",response_key_url_expr)
        history.pushState(null, null, url);
        // I think it is ok to call pushState() here because this never happens as a result of
        // the user hitting the back button.
    }
}

function makeAjaxCall(...args) {
    const addr = location.origin
    const url = new URL(addr + '/ajax');
    for (let i = 0; i < args.length; i+=2) {
        url.searchParams.set(args[i],args[i+1])
    }
    const xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status == 200) {
            db("processServerResponse")
            processServerResponse(this.responseText)
        }
    };
    db("makeAjaxCall",args)
    xhttp.open('GET', url);
    xhttp.send();
}

// An onchange event has occurred within an input field;
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

// An onfocus event has occurred within an input field
function jsFocus(id, active) {
    db("jsFocus",id,"active:",active,"from:",where(0,4))
    id_with_focus = null;
    if (active) {
        const x = document.getElementById(id);
        if (x != null) {
            id_with_focus = {id: id, url: window.location.href};
            db("updated id_with_focus to:", id_with_focus)
        }
    }
}

// An onchange event has occurred within a file upload.
// It assumes that the <form> element has id '<id>.form'
//
function jsUpload(id) {
    db("jsUpload",id)
    const addr = window.origin;
    const url = new URL(addr + '/upload/' + id);
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
}


// An onclick event has occurred within a button
function jsButton(id) {
    db("jsButton",id)
    makeAjaxCall(request_key_widget, id)
}

// An click event has occurred within a checkbox
function jsCheckboxClicked(id) {
    db("jsCheckboxClicked", id)
    auxId = id + '.aux'

    const checkbox = document.getElementById(auxId)
    if (checkbox == null) {
        warning("Cannot find checkbox element with id:",auxId)
        return
    }
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

function jsPopStateEventHandler(e) {
    pr("jsPopStateEventHandler, e:",e)
    var rel_path = window.location.pathname;
    pr("redirecting user to page:",rel_path)
    window.location.href = window.origin + rel_path
}
window.addEventListener('popstate', jsPopStateEventHandler);

pr("...base.js has loaded")

