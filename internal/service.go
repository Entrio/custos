package internal

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"gorm.io/gorm"
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

func (mc *MemoryCache) GetUser(name string) *User {
	mc.rwMutex.RLock()
	defer mc.rwMutex.RUnlock()

	if val, ok := mc.dictionary[name]; ok {
		if ku, ok := val.Value.(User); ok {
			return &ku
		}
	}

	fmt.Println(fmt.Sprintf("Identity %s not found in cache, looking up in database", name))

	user := new(User)
	result := dbInstance.Debug().Where(map[string]interface{}{"id": name}).First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if err := mc.AddItem(name, *user, nil); err != nil {
		return nil
	}

	return user
}

func ProcessUsers(users *[]KratosUser) error {

	if users == nil {
		// this is a nullptr, do nothing
		return nil
	}

	// Add each user to cache and make sure that they exist in the database
	//expiry := time.Now().Add(time.Second * 10)
	filteredusers := make([]User, 0)
	for _, v := range *users {
		u := User{
			Base: Base{
				ID: uuid.FromStringOrNil(v.ID),
			},
			Email:    v.VerifiableAddresses[0].Email,
			Position: v.Traits.Position,
			FirsName: v.Traits.Name.First,
			LastName: v.Traits.Name.Last,
			Verified: v.VerifiableAddresses[0].Verified,
			Enabled:  false,
		}
		filteredusers = append(filteredusers, u)
		//memorycache.AddItem(v.ID, v, &expiry)
		//memorycache.AddItem(v.ID, v, nil)
	}

	//TODO: Process teh database
	fmt.Println(filteredusers)

	return nil
}
