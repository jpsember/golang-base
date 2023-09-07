package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

type WebCacheStruct struct {
	CacheMap *ConcurrentMap[int, string]
	CacheDir Path
}

type WebCache = *WebCacheStruct

func (c WebCache) SetCacheDir(dir Path) {
	c.CacheDir = dir
	dir.MkDirsM()
}

func newWebCache() WebCache {
	t := &WebCacheStruct{
		CacheMap: NewConcurrentMap[int, string](),
	}
	return t
}

var SharedWebCache = newWebCache()
