package dns

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type itemCache struct {
	msg        *dns.Msg
	expiration time.Time
}

type MemoryCache struct {
	items map[string]itemCache
	mu    sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]itemCache),
	}
}

func (c *MemoryCache) Get(key string) (*dns.Msg, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.msg.Copy(), true
}

func (c *MemoryCache) Set(key string, msg *dns.Msg) {
	if len(msg.Answer) == 0 {
		return
	}
	ttl := time.Duration(msg.Answer[0].Header().Ttl) * time.Second
	toSave := msg.Copy()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = itemCache{
		msg:        toSave,
		expiration: time.Now().Add(ttl),
	}
}

func MakeKey(q dns.Question) string {
	return q.Name + "|" + string(rune(q.Qtype))
}
