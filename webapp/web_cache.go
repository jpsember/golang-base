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
	Map             *ConcurrentMap[blobName, blobData]
	MaxSize         int
	maintenanceLock sync.RWMutex
}

type WebCache = *WebCacheStruct

func logError(err error) bool {
	if err != nil {
		Alert("<1#50Error encountered:", err)
		return true
	}
	return false
}

func (c WebCache) GetData(name blobName) blobData {
	data := c.Map.Get(name)
	if data == nil {
		blob, err := ReadBlobWithName(name)
		data = blob
		logError(err)
		if blob.Id() != 0 {
			c.maintenanceLock.Lock()
			c.Map.Put(data.Name(), data)
			c.trim()
			c.maintenanceLock.Unlock()
		}
	}
	c.Log("GetData", name, "=>", data.Id())
	return data
}

// This should only be performed while the maintenanceLock is locked.
func (c WebCache) trim() {
	currentSize := c.Map.Size()
	if currentSize < c.MaxSize {
		return
	}
	c.Log("Trimming, size:", currentSize, "exceeds max:", c.MaxSize)
	oldKeys, oldValues := c.Map.GetAll()
	newWrappedMap := make(map[blobName]blobData)
	for i, k := range oldKeys {
		if (i & 1) == 0 {
			newWrappedMap[k] = oldValues[i]
		}
	}
	c.Map = NewConcurrentMapWith[blobName, blobData](newWrappedMap)
}

func newWebCache() WebCache {
	t := &WebCacheStruct{
		Map:     NewConcurrentMap[blobName, blobData](),
		MaxSize: 1000,
	}
	return t
}

var SharedWebCache = newWebCache()
