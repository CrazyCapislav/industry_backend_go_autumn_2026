package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	С := &LRUCache[K, V]{
		capacity: capacity,
		ll:       list.List{},
		items:    make(map[K]*list.Element),
	}
	return С
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if _, ok = c.items[key]; ok {
		c.ll.MoveToFront(c.items[key])
		c.items[key].Value = entry[K, V]{key, c.items[key].Value.(entry[K, V]).value}
		return c.items[key].Value.(entry[K, V]).value, true
	}
	return value, false
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if _, ok := c.items[key]; ok {
		c.items[key].Value = entry[K, V]{key, value}
		return
	}
	if c.ll.Len() >= c.capacity {
		delete(c.items, c.ll.Back().Value.(entry[K, V]).key)
		c.ll.Remove(c.ll.Back())

	}
	c.ll.PushFront(entry[K, V]{key, value})
	c.items[key] = c.ll.Front()

}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
