package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func assertVal[K comparable, V comparable](t *testing.T, c Cache[K, V], key K, expected V, f func() (V, error)) {
	result, err := c.Get(key, f)

	if result != expected {
		t.Errorf("Expected value %+v, but got %+v instead", expected, result)
	}

	if err != nil {
		t.Errorf("Unexpected error! %s", err)
	}
}

func TestBlockingCache(t *testing.T) {
	expected := rand.Int()
	key := "foobar"
	c := NewBlockingCache[string, int]()
	assertVal[string, int](t, c, key, expected, func() (int, error) {
		return expected, nil
	})

	assertVal[string, int](t, c, key, expected, func() (int, error) {
		panic("This should never be called")
	})

	start := make(chan interface{})

	var complete sync.WaitGroup

	for i := 0; i < 1000; i++ {
		complete.Add(1)
		go func() {
			<-start

			assertVal[string, int](t, c, key, expected, func() (int, error) {
				panic("This should never be called")
			})
			complete.Done()
		}()
	}
	close(start)
	complete.Wait()
}

func TestBlockingCacheConcurrency(t *testing.T) {
	c := NewBlockingCache[string, int]()
	start := make(chan interface{})

	var complete sync.WaitGroup

	for i := 0; i < 1000; i++ {
		complete.Add(1)
		key := fmt.Sprintf("key-%d", i)
		go func() {
			<-start

			c.Get(key, func() (int, error) {
				return rand.Int(), nil
			})
			complete.Done()
		}()
	}
	close(start)
	complete.Wait()
}
