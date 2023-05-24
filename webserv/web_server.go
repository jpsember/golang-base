package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
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

	var keyDir = NewPathM("webserv/keys")
	var certPath = keyDir.JoinM("server.crt")
	var keyPath = keyDir.JoinM("server.key")
	
	http.HandleFunc("/hello", HelloServer)
	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
