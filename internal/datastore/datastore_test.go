package datastore

import (
	"strings"
	"testing"
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
