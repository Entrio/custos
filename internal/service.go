package internal

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	instance.dictionary["users"] = memoryCacheValue{
		Value:  map[string]interface{}{},
		Expiry: nil,
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

func (mc *MemoryCache) AddUser(id string, user interface{}) error {
	mc.rwMutex.Lock()
	mc.dictionary["users"].Value.(map[string]interface{})[id] = user
	mc.rwMutex.Unlock()
	fmt.Println(fmt.Sprintf("Added user %s to cache", id))
	return nil
}

func (mc *MemoryCache) GetUser(name string) *User {
	if val, ok := mc.dictionary["users"]; ok {

		if ku, ok := val.Value.(map[string]interface{})[name]; ok {
			return ku.(*User)
		}
	}

	fmt.Println(fmt.Sprintf("Identity %s not found in cache, looking up in database", name))

	user := new(User)
	result := dbInstance.Debug().Where(map[string]interface{}{"id": name}).First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if err := mc.AddUser(name, *user); err != nil {
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
			Email:     v.VerifiableAddresses[0].Email,
			Position:  v.Traits.Position,
			FirstName: v.Traits.Name.First,
			LastName:  v.Traits.Name.Last,
			Verified:  v.VerifiableAddresses[0].Verified,
			Enabled:   false,
		}
		filteredusers = append(filteredusers, u)
		//memorycache.AddItem(v.ID, v, &expiry)
		memorycache.AddUser(v.ID, v)
	}

	dbInstance.Debug().Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(filteredusers, 5)

	return nil
}
