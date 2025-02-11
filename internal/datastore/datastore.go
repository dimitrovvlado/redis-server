package datastore

import (
	"errors"
	"maps"
	"sync"
	"time"
)

type Datastore struct {
	mu   sync.RWMutex
	data map[string]*Entry
}

// Entry is a struct that holds the value and the metadata related to it
type Entry struct {
	Value string
	//The expiration date in unix millis
	Expiry int64
}

func NewDatastore() *Datastore {
	return &Datastore{data: make(map[string]*Entry)}
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

func (d *Datastore) CheckExpiredKeys() {
	keys := make([]string, 0)
	d.mu.RLock()
	//copy the keys who have a set expiry
	for k := range maps.Keys(d.data) {
		if d.data[k].Expiry != -1 {
			keys = append(keys, k)
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
			return value.Value, nil
		}
	}
	return "", errors.New("not found")
}

func newEntry(value string, expiry int64) *Entry {
	return &Entry{Value: value, Expiry: expiry}
}
