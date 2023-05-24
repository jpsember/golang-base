package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"log"
	"net/http"
)

var _ = Pr

// https://github.com/denji/golang-tls

// How to get a certificate: server.crt
// https://www.vultr.com/docs/secure-a-golang-web-server-with-a-selfsigned-or-lets-encrypt-ssl-certificate/

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}

func Demo() {
	http.HandleFunc("/hello", HelloServer)
	err := http.ListenAndServeTLS(":443", "webserv/keys/server.crt", "webserv/keys/server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
