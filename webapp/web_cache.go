package webapp

import (
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"github.com/jpsember/golang-base/webserv"
)

type WebCacheStruct struct {
}

type WebCache = *WebCacheStruct

func NewWebCache() WebCache {
	t := &WebCacheStruct{}
	return t
}

func (w WebCache) ReadBlob(id int) (webserv.AbstractBlob, error) {
	return webapp_data.ReadBlob(id)
}

func (w WebCache) ReadBlobWithName(name string) (webserv.AbstractBlob, error) {
	return webapp_data.ReadBlobWithName(name)
}

var SharedWebCache = webserv.NewBlobCache(NewWebCache())
