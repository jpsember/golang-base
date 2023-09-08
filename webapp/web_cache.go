package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"sync"
)

type blobName = string
type blobData = Blob

type WebCacheStruct struct {
	BaseObject
	NameMap map[blobName]blobData
	IdMap   map[int]blobData
	MaxSize int
	lock    sync.RWMutex
}

type WebCache = *WebCacheStruct

func logError(err error) bool {
	if err != nil {
		Alert("<1#50Error encountered:", err)
		return true
	}
	return false
}

func (c WebCache) GetBlobWithId(id int) blobData {
	c.lock.RLock()
	data := c.IdMap[id]
	c.lock.RUnlock()

	if data == nil {
		c.Log("GetBlobWithId", id, " not found; reading into cache")
		blob, err := ReadBlob(id)
		logError(err)
		data = blob
		if blob.Id() != 0 {
			c.add(blob)
		}
	}
	return data
}

func (c WebCache) GetBlobWithName(name blobName) blobData {
	c.lock.RLock()
	data := c.NameMap[name]
	c.lock.RUnlock()

	if data == nil {
		c.Log("GetBlobWithName", name, " not found; reading into cache")
		blob, err := ReadBlobWithName(name)
		logError(err)
		data = blob
		if blob.Id() != 0 {
			c.add(blob)
		}
	}
	return data
}

func (c WebCache) add(blob blobData) {
	c.lock.Lock()

	c.NameMap[blob.Name()] = blob
	c.IdMap[blob.Id()] = blob
	c.trim()

	defer c.lock.Unlock()
}

// This should only be performed while we have the write lock.
func (c WebCache) trim() {
	currentSize := len(c.IdMap)
	if currentSize <= c.MaxSize {
		return
	}
	c.Log("Trimming, size:", currentSize, "exceeds max:", c.MaxSize)

	_, oldBlobs := GetMapKeysAndValues(c.IdMap)
	newIdMap := make(map[int]blobData)
	newNameMap := make(map[blobName]blobData)
	for i, b := range oldBlobs {
		if (i & 1) == 0 {
			newIdMap[b.Id()] = b
			newNameMap[b.Name()] = b
		}
	}
	c.Log("...new size:", len(newIdMap))
	c.IdMap = newIdMap
	c.NameMap = newNameMap
}

func newWebCache() WebCache {
	t := &WebCacheStruct{
		NameMap: make(map[blobName]blobData),
		IdMap:   make(map[int]blobData),
		MaxSize: 1000,
	}
	if Alert("Using small cache size") {
		t.MaxSize = 20
	}
	t.SetName("SharedWebCache")
	t.AlertVerbose()
	return t
}

var SharedWebCache = newWebCache()
