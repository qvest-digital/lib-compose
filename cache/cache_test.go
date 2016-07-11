package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Cache_PanicsOnCreateWithWrongSize(t *testing.T) {
	a := assert.New(t)
	a.Panics(func() {
		NewCache("my-cache", -1, 0, time.Millisecond)
	})
}

func Test_Cache_TTL(t *testing.T) {
	a := assert.New(t)

	// given a cache 1ms ttl
	c := NewCache("my-cache", 5, 100, time.Millisecond)

	// when i store an entry
	c.Set("foo", "", 0, "bar")

	// then I can retrieve it
	v, found := c.Get("foo")
	a.True(found)
	a.Equal("bar", v.(string))

	// but when I wait for the TTL
	time.Sleep(time.Millisecond)

	// then it is not found any more
	_, found = c.Get("foo")
	a.False(found)
}

func Test_Cache_LRU_MaxEntries(t *testing.T) {
	a := assert.New(t)

	// given a cache of size 3
	// with 3 entries
	c := NewCache("my-cache", 3, 100, time.Hour)
	c.Set("a", "", 0, "a")
	c.Set("b", "", 0, "b")
	c.Set("c", "", 0, "c")
	a.Equal(3, c.Len())

	// when I and access the oldest
	v, found := c.Get("a")
	a.True(found)
	a.Equal("a", v.(string))

	// and add one more
	c.Set("newcommer", "", 0, "newcommer")

	// then the recently used are in
	_, found = c.Get("a")
	a.True(found)
	_, found = c.Get("c")
	a.True(found)
	_, found = c.Get("newcommer")
	a.True(found)

	// but one is out
	_, found = c.Get("b")
	a.False(found)
}

func Test_Cache_MaxBytes(t *testing.T) {
	a := assert.New(t)

	// given a cache with max 1 mega byte, filled with 8 bytes
	c := NewCache("my-cache", 100, 1, time.Hour)
	c.Set("a", "", 42*1024, "a")
	c.Set("a", "", 400*1024, "a") // the same item only shoud count once, with the lastest bytes
	c.Set("b", "", 400*1024, "b")
	a.Equal(800*1024, c.SizeByte())
	a.Equal(2, c.Len())

	// when I add and 2 more bytes
	c.Set("c", "", 200*1024, "c")

	// then they fit
	a.Equal(1000*1024, c.SizeByte())
	a.Equal(3, c.Len())

	// but when i add even more
	c.Set("d", "", 200*1024, "c")

	// then the last accessed entry was taken out
	a.Equal(800*1024, c.SizeByte())
	a.Equal(3, c.Len())
	_, found := c.Get("b")
	a.True(found)
	_, found = c.Get("c")
	a.True(found)
	_, found = c.Get("d")
	a.True(found)
}

func Test_Cache_Stats(t *testing.T) {
	a := assert.New(t)

	c := NewCache("my-cache", 3, 100, time.Hour)
	c.Set("a", "", 42, "a")
	c.Get("a")
	c.Get("a")
	c.Get("b")

	c.calculateStats(time.Hour)
	a.Equal(1, c.stats["cache_entries"])
	a.Equal(42, c.stats["cache_size_bytes"])
	a.Equal(2, c.stats["cache_hits"])
	a.Equal(1, c.stats["cache_misses"])
	a.Equal(66, c.stats["cache_hit_ratio"])
}

func Test_Cache_PurgeOldEntries(t *testing.T) {
	a := assert.New(t)

	c := NewCache("my-cache", 100, 100, time.Millisecond)
	c.Set("a", "", 1, "a")
	c.Set("b", "", 1, "a")
	c.Set("c", "", 1, "a")
	time.Sleep(time.Millisecond)
	c.Set("d", "", 42, "a")
	c.Set("e", "", 42, "a")

	c.PurgeOldEntries()

	a.Equal(2, c.Len())
	a.Equal(84, c.SizeByte())
}
