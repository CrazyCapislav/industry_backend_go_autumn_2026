package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	С := &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]V),
	}
	return С
}
func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	v, ok = c.items[k]
	if c.items == nil {
		return v, false
	} else {
		return v, ok
	}
}
func (c *Cache[K, V]) Set(k K, v V) bool {
	if _, ok := c.items[k]; ok {
		c.items[k] = v
		return true
	}

	if len(c.items) >= c.capacity {
		return false
	}
	c.items[k] = v
	return true
}
