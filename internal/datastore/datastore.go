package datastore

import (
	"errors"
	"sync"
)

type Datastore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewDatastore() *Datastore {
	return &Datastore{data: make(map[string]string)}
}

func (d *Datastore) Set(key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[key] = value
}

func (d *Datastore) Get(key string) (string, error) {
	l := d.mu.RLocker()
	l.Lock()
	defer l.Unlock()

	value, ok := d.data[key]
	if ok {
		return value, nil
	}
	return "", errors.New("not found")
}
