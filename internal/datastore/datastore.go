package datastore

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"sync"
	"time"
)

type Datastore struct {
	mu   sync.RWMutex
	data map[string]*Entry

	expChunkSize int
}

// Entry is a struct that holds the value and the metadata related to it
type Entry struct {
	Value interface{}
	//The expiration date in unix millis
	Expiry int64
}

// KeyNotFoundError is an error struct which holds the missing key.
type KeyNotFoundError struct {
	key string
}

func NewDatastore() *Datastore {
	return &Datastore{data: make(map[string]*Entry), expChunkSize: 20}
}

func (d *Datastore) Set(key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[key] = newEntry(value, -1)
}

// SetWithExpiry sets the key/value pair with expiration.
// Expiry is the amount of millis after which the key will expire
func (d *Datastore) SetWithExpiry(key, value string, expiry int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[key] = newEntry(value, time.Now().UnixMilli()+expiry)
}

// SetWithExpiry sets the key/value pair with expiration.
// Expiry is the amount the timestamp in millis when the key becomes invalid
func (d *Datastore) SetWithExactExpiry(key, value string, expiry int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[key] = newEntry(value, expiry)
}

func (d *Datastore) StartExpiryCheck() {
	for {
		//TODO pause checks if no new keys are added
		d.ExpiryCheck()
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Datastore) ExpiryCheck() {
	keys := make([]string, 0)
	d.mu.RLock()
	keyCount := len(d.data)
	sampleSize := min(d.expChunkSize, keyCount)
	//copy the keys who have a set expiry
	for k := range maps.Keys(d.data) {
		if d.data[k].Expiry != -1 {
			if sampleSize > 0 {
				keys = append(keys, k)
				sampleSize -= 1
			} else {
				break
			}
		}
	}
	d.mu.RUnlock()

	for _, k := range keys {
		d.mu.Lock()
		delete(d.data, k)
		d.mu.Unlock()
	}
}

func (d *Datastore) Get(key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if value, ok := d.data[key]; ok {
		now := time.Now().UnixMilli()
		if value.Expiry == -1 || now < value.Expiry {
			var ret string
			switch value.Value.(type) {
			case int64:
				ret = fmt.Sprintf("%d", value.Value.(int64))
			case string:
				ret = value.Value.(string)
			}
			return ret, nil
		}
	}
	return "", errors.New("not found")
}

func (d *Datastore) Delete(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.data[key]; ok {
		delete(d.data, key)
		return nil
	}
	return errors.New("not found")
}

func (d *Datastore) Increment(key string) (int64, error) {
	return d.sumWith(key, 1)
}

func (d *Datastore) Decrement(key string) (int64, error) {
	return d.sumWith(key, -1)
}

func (d *Datastore) sumWith(key string, change int64) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var val int64
	var exp int64 = -1
	value, ok := d.data[key]
	if ok {
		switch value.Value.(type) {
		case int64:
			val = value.Value.(int64)
		case string:
			var err error
			val, err = strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return 0, err
			}
		}
		val += change
		exp = value.Expiry
		newEntry := Entry{Value: val, Expiry: exp}
		d.data[key] = &newEntry
		return val, nil
	}
	return 0, KeyNotFoundError{key: key}
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("%s not found in datastore", e.key)
}

func newEntry(value interface{}, expiry int64) *Entry {
	switch value.(type) {
	case string:
		v, err := strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			return &Entry{Value: v, Expiry: expiry}
		}

	}
	return &Entry{Value: value, Expiry: expiry}
}
