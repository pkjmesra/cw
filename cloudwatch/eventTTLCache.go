package cloudwatch

import (
	"log"
	"sync"
	"time"
)

const purgeFreq = 10 * time.Second

type eventCache struct {
	seen         map[string]int64
	mostRecentTS int64
	sync.RWMutex
}

// type eventCacheOld struct {
// 	seen     map[string]bool
// 	creation map[string]time.Time
// 	sync.RWMutex
// }

func createCache(ttl time.Duration, log *log.Logger) *eventCache {
	cache := &eventCache{seen: make(map[string]int64)} // creation: make(map[string]time.Time)

	log.Printf("cache: ttl:%s check-time:%s\n", ttl.String(), purgeFreq.String())

	cachePurge := func(c *eventCache, ttl time.Duration, freq time.Duration) {
		cacheTicker := time.NewTicker(purgeFreq)
		for range cacheTicker.C {
			c.Lock()

			var ids []string
			now := time.Now()
			for id, ts := range c.seen {
				if ts != c.mostRecentTS { //keep most recent TS
					t := time.Unix(ts/1000, 0)
					purgeCandidate := now.Sub(t).Seconds() >= ttl.Seconds()
					if purgeCandidate {
						ids = append(ids, id)
					}
				}
			}
			log.Println("entries to purge:", len(ids))

			for _, id := range ids {
				delete(c.seen, id)
			}
			c.Unlock()
		}
	}

	go cachePurge(cache, ttl, purgeFreq)

	return cache
}

func (c *eventCache) Has(eventID string) bool {
	c.RLock()
	defer c.RUnlock()
	return c.seen[eventID] != 0
}

func (c *eventCache) Add(eventID string, ts int64) {
	c.Lock()
	defer c.Unlock()
	c.seen[eventID] = ts
	c.mostRecentTS = ts
}

func (c *eventCache) Size() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.seen)
}
