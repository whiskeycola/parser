package jas

import "sync"

// type cache map[position]atomSize
// caching atomic sizes
type cache struct {
	mp  map[int]int
	mux sync.Mutex
}

func newCache() *cache {
	return &cache{
		mp: make(map[int]int),
	}
}
func (c *cache) Add(start, size int) {
	c.mux.Lock()
	c.mp[start] = size
	c.mux.Unlock()
}
func (c *cache) Get(start int) (size int, ok bool) {
	c.mux.Lock()
	size, ok = c.mp[start]
	c.mux.Unlock()
	return
}

func (c *cache) takeAtom(vector []byte, start int) *atom {
	size, ok := c.Get(start)
	if ok {
		return &atom{
			vector: vector[start : start+size],
			cache:  copyCache(c, start, size),
		}
	}
	return nil
}
func copyCache(c *cache, start, size int) *cache {
	n := newCache()
	c.mux.Lock()
	for s, sz := range c.mp {
		if s >= start && s < start+size {
			n.mp[s-start] = sz
		}
	}
	c.mux.Unlock()
	return n
}
