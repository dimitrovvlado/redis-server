package datastore

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSetAndGetFromDatastore(t *testing.T) {
	tests := map[string]struct {
		key         string
		expected    string
		expectedErr string
	}{
		"Simple Get":  {key: "key", expected: "value", expectedErr: ""},
		"Invalid Get": {key: "invalid key", expected: "", expectedErr: "not found"},
	}

	ds := NewDatastore()
	ds.Set("key", "value")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ds.Get(test.key)
			if err != nil {
				if !strings.Contains(err.Error(), test.expectedErr) {
					t.Errorf("Unexpected error %v", err)
				}
			} else {
				if got != test.expected {
					t.Errorf("Expected: %s got %s", test.expected, got)
				}
			}
		})
	}
}

func TestSetWithExpiry(t *testing.T) {
	ds := NewDatastore()
	ds.SetWithExpiry("key", "value", 500) //expire in 500 millis
	got, err := ds.Get("key")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if got != "value" {
		t.Errorf("Expected 'value', got '%s'", got)
	}
	time.Sleep(500 * time.Millisecond)
	_, err = ds.Get("key")
	if err == nil {
		t.Errorf("Key has not expired, but it should have")
	}
}

func TestSetWithExactExpiry(t *testing.T) {
	ds := NewDatastore()
	ds.SetWithExactExpiry("key", "value", time.Now().Add(500*time.Millisecond).UnixMilli()) //expire in 500 millis
	got, err := ds.Get("key")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if got != "value" {
		t.Errorf("Expected 'value', got '%s'", got)
	}
	time.Sleep(500 * time.Millisecond)
	_, err = ds.Get("key")
	if err == nil {
		t.Errorf("Key has not expired, but it should have")
	}
}

func TestExpiryCheck(t *testing.T) {
	ds := NewDatastore()
	for i := range 100 { //100 permanent keys
		key := fmt.Sprintf("key%d", i)
		ds.Set(key, "value")
	}
	for i := range 100 { //100 perishable keys
		key := fmt.Sprintf("key%d", i+100)
		ds.SetWithExactExpiry(key, "value", time.Now().Add(1*time.Millisecond).UnixMilli())
	}
	if len(ds.data) != 200 {
		t.Errorf("Expected 200 items, got %d", len(ds.data))
	}
	ds.ExpiryCheck() //should remove 20 items by default
	if len(ds.data) != 180 {
		t.Errorf("Expected 180 items, got %d", len(ds.data))
	}
}

func TestStartExpiryCheck(t *testing.T) {
	ds := NewDatastore()
	for i := range 100 {
		key := fmt.Sprintf("key%d", i)
		ds.Set(key, "value")
	}
	for i := range 100 {
		key := fmt.Sprintf("key%d", i+100)
		ds.SetWithExactExpiry(key, "value", time.Now().Add(1*time.Millisecond).UnixMilli())
	}
	if len(ds.data) != 200 {
		t.Errorf("Expected 200 items, got %d", len(ds.data))
	}
	go ds.StartExpiryCheck()
	time.Sleep(2 * time.Second)
	if len(ds.data) != 100 {
		t.Errorf("Expected 100 items, got %d", len(ds.data))
	}
}

func TestDelete(t *testing.T) {
	ds := NewDatastore()
	ds.Set("key", "value")
	err := ds.Delete("key")
	if err != nil {
		t.Errorf("Expected item to be deleted")
	}
	err = ds.Delete("key")
	if err == nil {
		t.Errorf("Expected: no item to be deleted")
	}
}
