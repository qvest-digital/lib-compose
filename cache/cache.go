package cache

import (
	"github.com/hashicorp/golang-lru"
)

type Cache struct {
	lruBackend *lru.ARCCache
}

// NewCache creates a cache with max 100MB and max 10.000 Entries
func NewCache(entrySize int) *Cache {
	arc, err := lru.NewARC(entrySize)
	if err != nil {
		panic(err)
	}
	return &Cache{
		lruBackend: arc,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.lruBackend.Get(key)

}

func (c *Cache) Set(key string, sizeBytes int, cacheObject interface{}) {
	c.lruBackend.Add(key, cacheObject)
}
