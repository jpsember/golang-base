package webapp

import (
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"github.com/jpsember/golang-base/webserv"
)

type webCacheBlobHelper struct {
}

func (w webCacheBlobHelper) ReadBlob(id int) (webserv.AbstractBlob, error) {
	return webapp_data.ReadBlob(id)
}

func (w webCacheBlobHelper) ReadBlobWithName(name string) (webserv.AbstractBlob, error) {
	return webapp_data.ReadBlobWithName(name)
}

func ConstructSharedWebCache() webserv.BlobCache {
	return webserv.NewBlobCache(webCacheBlobHelper{})
}

var SharedWebCache webserv.BlobCache
