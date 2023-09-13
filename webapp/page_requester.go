package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type PageRequesterStruct struct {
}

type PageRequester = *PageRequesterStruct

func (r PageRequester) Process(sUnused Session, user User, path string) bool {

	var resultPath string

	if !IsUserLoggedIn(user.Id()) {
		resultPath = "/"
	} else {
		resultPath = "/feed"
	}
	Pr(resultPath)

	return false
}

func (r PageRequester) RegisterPage(pg BasicPage) {

}
func NewPageRequester() PageRequester {
	PrIf(false)
	t := &PageRequesterStruct{}
	return t
}
