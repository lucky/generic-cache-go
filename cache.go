package cache

import "sync"

type Cache[K comparable, V any] interface {
	Get(key K, f func() (V, error)) (V, error)
}

type entry[V any] struct {
	v   V
	err error
}

type BlockingCache[K comparable, V any] struct {
	vals  map[K]entry[V]
	mutex *sync.Mutex // Does this need to be a pointer?
}

func NewBlockingCache[K comparable, V any]() *BlockingCache[K, V] {
	return &BlockingCache[K, V]{
		vals:  make(map[K]entry[V]),
		mutex: &sync.Mutex{},
	}
}

func (c *BlockingCache[K, V]) Get(key K, f func() (V, error)) (V, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ntry, ok := c.vals[key]; ok {
		return ntry.v, ntry.err
	}

	v, err := f()
	ntry := entry[V]{v: v, err: err}
	c.vals[key] = ntry
	return ntry.v, ntry.err
}
