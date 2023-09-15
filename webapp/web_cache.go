package webapp

import (
	. "github.com/jpsember/golang-base/base"
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

func ReadImageIntoCache(blobId int) string {
	Todo("Make this a method of BlobCache, with a default value if blob not found")
	s := SharedWebCache
	blob := s.GetBlobWithId(blobId)
	var url string
	if blob.Id() == 0 {
		url = "missing.jpg"
	} else {
		url = webserv.BlobURLPrefix + blob.Name()
	}
	return url
}
