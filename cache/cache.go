package cache

import (
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/tarent/lib-compose/logging"
	"sync"
	"sync/atomic"
	"time"
)

// Cache is a LRU cache with the following features
// - limits on max entries
// - memory size limit
// - ttl for entries
type Cache struct {
	name             string
	lock             sync.RWMutex
	lruBackend       *simplelru.LRU
	maxAge           time.Duration
	maxSizeBytes     int32
	currentSizeBytes int32
}

type CacheEntry struct {
	key         string
	label       string
	size        int
	fetchTime   time.Time
	cacheObject interface{}
	hits        int
}

// NewCache creates a new cache
func NewCache(name string, maxEntries int, maxSizeMB int, maxAge time.Duration) *Cache {
	c := &Cache{
		name:         name,
		maxAge:       maxAge,
		maxSizeBytes: int32(maxSizeMB) * 1024 * 1024,
	}

	var err error
	c.lruBackend, err = simplelru.NewLRU(maxEntries, simplelru.EvictCallback(c.onEvicted))
	if err != nil {
		panic(err)
	}
	return c
}

// LogEvery Start a Goroutine, which logs statisitcs periodically.
func (c *Cache) LogEvery(d time.Duration) {
	go c.logEvery(d)
}

func (c *Cache) logEvery(d time.Duration) {
	for {
		select {
		case <-time.After(d):
			logging.Logger.WithFields(logrus.Fields{
				"type":             "metric",
				"metric_name":      "cachestatus",
				"cache_entries":    c.Len(),
				"cache_size_bytes": c.SizeByte(),
			}).Infof("cache status #%v, %vbytes", c.Len(), c.SizeByte())
		}
	}
}

func (c *Cache) onEvicted(key, value interface{}) {
	entry := value.(*CacheEntry)
	atomic.AddInt32(&c.currentSizeBytes, -1*int32(entry.size))
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, found := c.lruBackend.Get(key)
	if found {
		entry := e.(*CacheEntry)
		if time.Since(entry.fetchTime) < c.maxAge {
			entry.hits++
			return entry.cacheObject, true
		}
	}
	return nil, false
}

func (c *Cache) Set(key string, label string, sizeBytes int, cacheObject interface{}) {
	entry := &CacheEntry{
		key:         key,
		label:       label,
		size:        sizeBytes,
		fetchTime:   time.Now(),
		cacheObject: cacheObject,
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	atomic.AddInt32(&c.currentSizeBytes, int32(sizeBytes))
	c.lruBackend.Add(key, entry)

	for atomic.LoadInt32(&c.currentSizeBytes) > int32(c.maxSizeBytes) {
		c.lruBackend.RemoveOldest()
	}
}

func (c *Cache) Invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lruBackend.Purge()
	c.currentSizeBytes = 0
	return
}

// SizeByte returns the total memory consumption of the cache
func (c *Cache) SizeByte() int {
	return int(atomic.LoadInt32(&c.currentSizeBytes))
}

// Len returns the total number of entries in the cache
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lruBackend.Len()
}
