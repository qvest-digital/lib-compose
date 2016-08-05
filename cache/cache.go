package cache

import (
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/tarent/lib-compose/logging"
	"sync"
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
	maxSizeBytes     int
	currentSizeBytes int
	hits             int
	misses           int
	stats            map[string]interface{}
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
		maxSizeBytes: maxSizeMB * 1024 * 1024,
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
			c.PurgeOldEntries()
			c.calculateStats(d)
		}
	}
}

func (c *Cache) calculateStats(reportingDuration time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ratio := 100
	if c.hits+c.misses != 0 {
		ratio = 100 * c.hits / (c.hits + c.misses)
	}

	c.stats = map[string]interface{}{
		"type":                     "metric",
		"metric_name":              "cachestatus",
		"cache_entries":            c.lruBackend.Len(),
		"cache_size_bytes":         c.currentSizeBytes,
		"cache_reporting_duration": reportingDuration,
		"cache_hits":               c.hits,
		"cache_misses":             c.misses,
		"cache_hit_ratio":          ratio,
	}

	c.hits = 0
	c.misses = 0
	logging.Logger.
		WithFields(logrus.Fields(c.stats)).
		Infof("cache status #%v, %vbytes, %v%% hits", c.lruBackend.Len(), c.currentSizeBytes, ratio)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, found := c.lruBackend.Get(key)
	if found {
		entry := e.(*CacheEntry)
		if time.Since(entry.fetchTime) < c.maxAge {
			entry.hits++
			c.hits++
			return entry.cacheObject, true
		}
	}
	c.misses++
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

	// first remove, to have correct size counting
	c.lruBackend.Remove(key)

	c.currentSizeBytes += sizeBytes
	c.lruBackend.Add(key, entry)

	for c.currentSizeBytes > c.maxSizeBytes {
		c.lruBackend.RemoveOldest()
	}
}

// called by the cache api, if items are removed,
// because of an overfilled cache.
// Attention: This method does not locking, because it
//     will be triggered as a subcall of Set()
func (c *Cache) onEvicted(key, value interface{}) {
	entry := value.(*CacheEntry)
	c.currentSizeBytes -= entry.size
}

// PurgeOldEntries removes all entries which are out of their ttl
func (c *Cache) PurgeOldEntries() {
	c.lock.RLock()
	keys := c.lruBackend.Keys()
	c.lock.RUnlock()
	purged := 0
	for _, key := range keys {
		c.lock.RLock()
		e, found := c.lruBackend.Peek(key)
		c.lock.RUnlock()

		if found {
			entry := e.(*CacheEntry)
			if time.Since(entry.fetchTime) > c.maxAge {
				c.lock.Lock()
				c.lruBackend.Remove(key)
				c.lock.Unlock()
				purged++
			}
		}
	}
	logging.Logger.
		WithFields(logrus.Fields(c.stats)).
		Infof("purged %v out of %v cache entries", purged, len(keys))
}

// Purge Entries with a specific hash
func (c *Cache) PurgeEntries(keys []string) {
	purged := 0
	purgedKeys := []string{}
	for _, key := range keys {
		c.lock.RLock()
		_, found := c.lruBackend.Peek(key)
		c.lock.RUnlock()

		if found {
			c.lock.Lock()
			c.lruBackend.Remove(key)
			c.lock.Unlock()
			purged++
			purgedKeys = append(purgedKeys, key)
		}
	}
	logging.Logger.
		WithFields(logrus.Fields(c.stats)).
		Infof("Following cache entries become purged: %v", c.PurgedKeysAsString(keys))
}

func (c *Cache) PurgedKeysAsString(keys []string) string {
	count := 0
	keyString := ""
	for _, key := range keys {
		if count != 0 {
			keyString += ", " + key
		} else {
			keyString += key
		}
		count ++
	}
	return keyString
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
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.currentSizeBytes
}

// Len returns the total number of entries in the cache
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lruBackend.Len()
}
