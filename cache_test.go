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

func testCache(t *testing.T, c Cache[string, int]) {
	expected := rand.Int()
	key := "foobar"
	assertVal(t, c, key, expected, func() (int, error) {
		return expected, nil
	})

	assertVal(t, c, key, expected, func() (int, error) {
		panic("This should never be called")
	})

	start := make(chan interface{})

	var complete sync.WaitGroup

	for i := 0; i < 1000; i++ {
		complete.Add(1)
		go func() {
			<-start

			assertVal(t, c, key, expected, func() (int, error) {
				panic("This should never be called")
			})
			complete.Done()
		}()
	}
	close(start)
	complete.Wait()
}

func testCacheConcurrency(t *testing.T, c Cache[string, int]) {
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

func TestBlockingCache(t *testing.T) {
	c := NewBlockingCache[string, int]()
	testCache(t, c)
}

func TestBlockingCacheConcurrency(t *testing.T) {
	c := NewBlockingCache[string, int]()
	testCacheConcurrency(t, c)
}

func TestNonBlockingCache(t *testing.T) {
	c := NewNonBlockingCache[string, int]()
	testCache(t, c)
}

func TestNonBlockingCacheConcurrency(t *testing.T) {
	c := NewNonBlockingCache[string, int]()
	testCacheConcurrency(t, c)
}
