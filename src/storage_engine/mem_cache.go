package storage_engine

import (
	"sync"
	"time"

	"github.com/allegro/bigcache"
)

// use bigcache as cache engine

var (
	memCacheInstace  *MemCache
	initMemcacheOnce sync.Once
)

type MemCache struct {
	*bigcache.BigCache
}

func GetMemcacheInstance() *MemCache {
	var err error
	initMemcacheOnce.Do(func() {
		memCacheInstace, err = newMemCache()
		if err != nil {
			panic(err)
		}
	})
	return memCacheInstace
}

func newMemCache() (*MemCache, error) {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		return nil, err
	}
	return &MemCache{cache}, nil
}

func (c *MemCache) Set(key string, value []byte) error {
	return c.BigCache.Set(key, value)
}

func (c *MemCache) Get(key string) ([]byte, error) {
	return c.BigCache.Get(key)
}

func (c *MemCache) Delete(key string) error {
	return c.BigCache.Delete(key)
}

func (c *MemCache) Close() error {
	return c.BigCache.Close()
}

func (c *MemCache) Len() int {
	return c.BigCache.Len()
}

func (c *MemCache) Reset() error {
	return c.BigCache.Reset()
}

func (c *MemCache) Stats() bigcache.Stats {
	return c.BigCache.Stats()
}
