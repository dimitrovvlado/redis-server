package datastore

import (
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
