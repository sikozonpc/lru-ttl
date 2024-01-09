package main

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

type EvictCallback func(key string, value interface{})

type Options struct {
	withLogs      bool
	evictCallback EvictCallback
}

type LRU struct {
	keyMap map[string]*list.Element
	list *list.List
	size int
	opts Options
	sync.RWMutex
}

type Item struct {
	key    string
	value  interface{}
	expiry time.Time
}

func NewLRUWithTTL(size int, opts Options) *LRU {
	if size <= 0 {
		return nil
	}

	lru := &LRU{
		size: size,
		keyMap: make(map[string]*list.Element),
		list: list.New(),
		opts: opts,
	}

	return lru
}

// Set sets a new key and value into the cache with a ttl option.
// It returns true if a new element has been created.
func (l *LRU) Set(key string, value interface{}, ttl time.Duration) bool {
	l.Lock()
	defer l.Unlock()

	expiry := time.Now().Add(ttl)

	time.AfterFunc(ttl, func() {
		if l.opts.withLogs {
			log.Printf("Expiring: %s", key)
		}

		l.Lock()
		defer l.Unlock()

		if elem, ok := l.keyMap[key]; ok {
			item := elem.Value.(*Item)
			if time.Now().After(item.expiry) {
				l.list.MoveToBack(elem)

				if l.opts.withLogs {
					log.Printf("Elem moved back to list: %s", key)
				}
			}
		}
	})

	if elem, ok := l.keyMap[key]; ok {
		item := elem.Value.(*Item)
		item.value = value
		item.expiry = expiry

		l.list.MoveToFront(elem)

		if l.opts.withLogs {
			log.Printf("Elem %s updated to the front", key)
			l.printList()
		}

		return false
	}

	item := &Item{
		key:    key,
		value:  value,
		expiry: expiry,
	}

	elem := l.list.PushFront(item)

	l.keyMap[key] = elem

	if l.list.Len() > l.size {
		l.removeLastElement()
	}

	if l.opts.withLogs {
		log.Printf("Elem %s added to the front", key)
		l.printList()
	}

	return true
}

// Get gets by key and returns the value and if the entry is expired.
// If expired it is moved to the back of the list else it gets
// moved  to front as the most recently used
func (l *LRU) Get(key string) (value interface{}, expired bool) {
	l.Lock()
	defer l.Unlock()

	if elem, ok := l.keyMap[key]; ok {
		item := elem.Value.(*Item)

		expired := time.Now().After(item.expiry)
		if expired {
			l.list.MoveToBack(elem)
		} else {
			l.list.MoveToFront(elem)
		}

		return item.value, expired
	}

	return nil, false
}

func (l *LRU) printList() {
	for elem := l.list.Front(); elem != nil; elem = elem.Next() {
		item := elem.Value.(*Item)
		fmt.Printf("key: %s, value: %v \n", item.key, item.value)
	}
}

func (l *LRU) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	// remove from list
	l.list.Remove(e)

	// remove from map
	item := e.Value.(*Item)
	delete(l.keyMap, item.key)

	// evict element
	if l.opts.evictCallback != nil {
		l.opts.evictCallback(item.key, item.value)
	}
}

func (l *LRU) removeLastElement() {
	l.removeElement(l.list.Back())
}
