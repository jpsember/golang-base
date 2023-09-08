package webapp

import (
	_ "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type AbstractPage interface {
	Generate()
}

type BasicPageStruct struct {
	session    Session
	parentPage Widget
}

type BasicPage = *BasicPageStruct

func NewBasicPage(session Session, parentPage Widget) BasicPage {
	t := &BasicPageStruct{
		session:    session,
		parentPage: parentPage,
	}
	return t
}
