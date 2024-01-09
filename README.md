## LRU with TTL

This is a simple LRU cache with TTL support. It is implemented using a doubly linked list and a hash map. The hash map is used to store the key-value pairs and the linked list is used to maintain the order of the keys. The least recently used key is always at the tail of the linked list and the most recently used key is always at the head of the linked list. The TTL is implemented using a background thread that runs every second and removes the expired keys from the cache.


### Usage

Check the `lruttl_test.go` file for usage examples.
