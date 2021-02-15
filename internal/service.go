package internal

import (
	"fmt"
	"sync"
	"time"
)

type (
	MemoryCache struct {
		dictionary map[string]memoryCacheValue
		terminate  chan bool
		rwMutex    sync.RWMutex
	}
	memoryCacheValue struct {
		Value  interface{}
		Expiry *time.Time
	}
)

func NewMemoryCache() *MemoryCache {
	instance := &MemoryCache{
		dictionary: map[string]memoryCacheValue{},
		terminate:  make(chan bool),
	}
	go instance.startCleanup()
	return instance
}

func (mc *MemoryCache) startCleanup() {
	for {
		mc.rwMutex.Lock()
		// perform cleanup
		for k, v := range mc.dictionary {
			// We have no expiration, skip it
			if v.Expiry == nil {
				continue
			}

			if v.Expiry.Before(time.Now()) {
				fmt.Println(fmt.Sprintf("Stale entry for %s, deleting", k))
				// we have expired
				delete(mc.dictionary, k)
			}
		}
		// Sleep for a time
		mc.rwMutex.Unlock()
		fmt.Println(fmt.Sprintf("Items in cache: %d", len(mc.dictionary)))
		time.Sleep(time.Second * 30)
		select {
		case <-mc.terminate:
			fmt.Println("Terminating NewMemoryCache cleanup routine")
			return
		default:
		}
	}
}

func (mc *MemoryCache) AddItem(name string, value interface{}, expiry *time.Time) error {
	mc.rwMutex.Lock()
	mc.dictionary[name] = memoryCacheValue{
		Value:  value,
		Expiry: expiry,
	}
	mc.rwMutex.Unlock()
	fmt.Println(fmt.Sprintf("Added %s to cache", name))
	return nil
}

func (mc *MemoryCache) GetUser(name string) *KratosUser {
	mc.rwMutex.RLock()
	defer mc.rwMutex.RUnlock()

	if val, ok := mc.dictionary[name]; ok {
		if ku, ok := val.Value.(KratosUser); ok {
			return &ku
		}
	}

	return nil
}
