package cache

import "sync"

type Cache[K comparable, V any] interface {
	Get(key K, f func() (V, error)) (V, error)
}

type result[V any] struct {
	v   V
	err error
}

type entry[V any] struct {
	res   result[V]
	ready chan struct{}
}

type BlockingCache[K comparable, V any] struct {
	vals  map[K]*result[V]
	mutex sync.Mutex
}

func NewBlockingCache[K comparable, V any]() *BlockingCache[K, V] {
	return &BlockingCache[K, V]{
		vals: make(map[K]*result[V]),
	}
}

func (c *BlockingCache[K, V]) Get(key K, f func() (V, error)) (V, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ntry, ok := c.vals[key]; ok {
		return ntry.v, ntry.err
	}

	v, err := f()
	ntry := &result[V]{v: v, err: err}
	c.vals[key] = ntry
	return ntry.v, ntry.err
}

type NonBlockingCache[K comparable, V any] struct {
	vals  map[K]*entry[V]
	mutex sync.Mutex
}

func NewNonBlockingCache[K comparable, V any]() *NonBlockingCache[K, V] {
	return &NonBlockingCache[K, V]{
		vals: make(map[K]*entry[V]),
	}
}

// Okay, so this blocks on the mutex, but only briefly so it can place an entry
// with a ready channel in the map. This is how the book does it and I'm really
// just copying that naming and methodology
func (c *NonBlockingCache[K, V]) Get(key K, f func() (V, error)) (V, error) {
	c.mutex.Lock()
	e, ok := c.vals[key]
	if !ok {
		e = &entry[V]{
			ready: make(chan struct{}),
		}
		c.vals[key] = e
		c.mutex.Unlock()
		e.res.v, e.res.err = f()
		close(e.ready)
	} else {
		c.mutex.Unlock()
		<-e.ready
	}
	return e.res.v, e.res.err
}
