package main

import (
	"testing"
	"time"
)

func TestSetOutRange(t *testing.T) {
	opts := Options{
		evictCallback: nil,
	}

	c := NewLRUWithTTL(3, opts)

	c.Set("a", 1, time.Second*3)
	testOrder(t, c, []string{"a"})

	c.Set("b", 2, time.Second*5)
	testOrder(t, c, []string{"b", "a"})

	c.Set("c", 3, time.Second*2)
	c.Set("a", 5, time.Second*5)
	testOrder(t, c, []string{"a", "c", "b"})

	// should not break
	c.Set("d", 7, time.Second*1)
	c.Set("e", 10, time.Second*5)

	testOrder(t, c, []string{"e", "d", "a"})
}

func TestGet(t *testing.T) {
	opts := Options{
		evictCallback: nil,
	}
	c := NewLRUWithTTL(3, opts)

	c.Set("a", 1, time.Second*3)
	c.Set("b", 2, time.Second*5)
	c.Set("c", 3, time.Second*2)
	c.Set("a", 5, time.Second*5)

	// test updated value
	val, expired := c.Get("a")
	if val != 5 {
		t.Errorf("Wrong value: %v expected: %v", 1, val)
	}
	if expired {
		t.Errorf("should not be expired yet")
	}
}

func TestEvictCallback(t *testing.T) {
	var evictedKey string

	opts := Options{
		evictCallback: func(key string, value interface{}) {
			evictedKey = key
		},
	}
	c := NewLRUWithTTL(3, opts)

	c.Set("a", 12, time.Second*3)
	c.Set("b", 2, time.Second*3)
	c.Set("c", 42, time.Second*2)

	// should evict after adding a new elem at max cache capacity
	c.Set("d", 3, time.Second*2)

	if evictedKey != "a" {
		t.Errorf("evicted key should be 'a' and not %s", evictedKey)
	}
}

func TestExpiry(t *testing.T) {
	opts := Options{
		evictCallback: nil,
	}
	c := NewLRUWithTTL(3, opts)

	c.Set("a", 1, time.Second*3)
	c.Set("b", 2, time.Second*5)
	c.Set("c", 3, time.Microsecond*500)
	c.Set("a", 5, time.Second*5)

	// test expired key
	val, expired := c.Get("c")
	if val != 3 {
		t.Errorf("Wrong value: %v expected: %v", 3, val)
	}
	if expired {
		t.Errorf("should not be expired yet")
	}

	time.Sleep(time.Microsecond*600)

	val, expired = c.Get("c")
	if val != 3 {
		t.Errorf("Wrong value: %v expected: %v", 3, val)
	}
	if !expired {
		t.Errorf("should be expired yet")
	}
}

func testOrder(t *testing.T, lru *LRU, want []string) {
	// check cache order
	i := 0
	for elem := lru.list.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*Item)

		if want[i] != item.key {
			t.Errorf("Invalid order of key %s", item.key)
		}

		i++
	}
}
