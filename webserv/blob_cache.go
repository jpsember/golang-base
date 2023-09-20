package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"sync"
)

type AbstractBlob interface {
	Id() int
	Name() string
	Data() []byte
}

type BlobHelper interface {
	ReadBlob(id int) (AbstractBlob, error)
	ReadBlobWithName(name string) (AbstractBlob, error)
}

type blobName = string
type blobData = AbstractBlob

type BlobCacheStruct struct {
	BaseObject
	helper  BlobHelper
	NameMap map[blobName]blobData
	IdMap   map[int]blobData
	MaxSize int
	lock    sync.RWMutex
}

func NewBlobCache(helper BlobHelper) BlobCache {
	t := &BlobCacheStruct{
		NameMap: make(map[blobName]blobData),
		IdMap:   make(map[int]blobData),
		MaxSize: 1000,
		helper:  helper,
	}
	if false && Alert("Using small cache size") {
		t.MaxSize = 20
	}
	t.SetName("SharedBlobCache")
	//t.AlertVerbose()
	return t
}

type BlobCache = *BlobCacheStruct

func logError(err error) bool {
	if err != nil {
		Alert("<1#50Error encountered:", err)
		return true
	}
	return false
}

func (c BlobCache) GetBlobWithId(id int) blobData {
	c.lock.RLock()
	data := c.IdMap[id]
	c.lock.RUnlock()

	if data == nil {
		c.Log("GetBlobWithId", id, " not found; reading into cache")
		blob, err := c.helper.ReadBlob(id)
		logError(err)
		data = blob
		if blob.Id() != 0 {
			c.add(blob)
		}
	}
	return data
}

func (c BlobCache) GetBlobWithName(name blobName) blobData {
	c.lock.RLock()
	data := c.NameMap[name]
	c.lock.RUnlock()

	if data == nil {
		c.Log("GetBlobWithName", name, " not found; reading into cache")
		blob, err := c.helper.ReadBlobWithName(name)
		logError(err)
		data = blob
		if blob.Id() != 0 {
			c.add(blob)
		}
	}
	return data
}

func (c BlobCache) GetBlobURL(blobId int) string {
	blob := c.GetBlobWithId(blobId)
	var url string
	if blob.Id() != 0 {
		url = BlobURLPrefix + blob.Name()
	}
	return url
}

func (c BlobCache) add(blob blobData) {
	c.lock.Lock()

	c.NameMap[blob.Name()] = blob
	c.IdMap[blob.Id()] = blob
	c.trim()

	defer c.lock.Unlock()
}

// This should only be performed while we have the write lock.
func (c BlobCache) trim() {
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
