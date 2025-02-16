package commands

import (
	"reflect"
	"testing"

	"github.com/dimitrovvlado/redis-server/internal/datastore"
	"github.com/dimitrovvlado/redis-server/internal/protocol"
)

func TestHandleCommand(t *testing.T) {
	tests := map[string]struct {
		in       protocol.Resp
		expected protocol.Resp
	}{
		"PING": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("PING")}}},
			expected: protocol.SimpleString{Data: "PONG"}},
		"ping": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("ping")}}},
			expected: protocol.SimpleString{Data: "PONG"}},
		"ping with a param": {
			in: protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("ping")},
				protocol.BulkString{Data: protocol.Ptr("param")}}},
			expected: protocol.BulkString{Data: protocol.Ptr("param")}},
		"ping with multiple params": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("ping")},
				protocol.BulkString{Data: protocol.Ptr("p1")},
				protocol.BulkString{Data: protocol.Ptr("p2")}}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'ping' command"}},
		"echo hello world": {
			in: protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("echo")},
				protocol.BulkString{Data: protocol.Ptr("Hello World")}}},
			expected: protocol.BulkString{Data: protocol.Ptr("Hello World")}},
		"unknown command": {
			in: protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("foo")},
				protocol.BulkString{Data: protocol.Ptr("bar")},
				protocol.BulkString{Data: protocol.Ptr("baz")}}},
			expected: protocol.Error{Data: "ERR unknown command 'foo', with args beginning with: 'bar' 'baz'"}},
		"Set with 2 too few args": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"}},
		"Set with 1 too few args": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key1")}}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"}},
		"Set with existing key": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key")}, protocol.BulkString{Data: protocol.Ptr("value")}}},
			expected: protocol.SimpleString{Data: "OK"}},
		"Set with non existent key": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key1")}, protocol.BulkString{Data: protocol.Ptr("value1")}}},
			expected: protocol.SimpleString{Data: "OK"}},
		"Set EX Error": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("EX")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"},
		},
		"Set PX Error": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("PX")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"},
		},
		"Set Invalid Expiry Type": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("MADEUPX")},
				protocol.BulkString{Data: protocol.Ptr("1")},
			}},
			expected: protocol.Error{Data: "ERR syntax error"},
		},
		"Set Invalid Expiry Value": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("PX")},
				protocol.BulkString{Data: protocol.Ptr("ten")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"},
		},
		"Set EXAT Error": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("EXAT")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"},
		},
		"Set PXAT Error": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("PXAT")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'set' command"},
		},
		"Set Invalid EXAT Expiry Value": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("set")},
				protocol.BulkString{Data: protocol.Ptr("k")},
				protocol.BulkString{Data: protocol.Ptr("v")},
				protocol.BulkString{Data: protocol.Ptr("EXAT")},
				protocol.BulkString{Data: protocol.Ptr("-1")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"},
		},
	}
	ds := datastore.NewDatastore()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := HandleCommand(test.in, ds)
			if err != nil {
				t.Errorf("HandleCommand() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("expected: %v, got: %v", test.expected, got)
			}
		})
	}
}

func TestExistsCommand(t *testing.T) {
	tests := map[string]struct {
		in       protocol.Resp
		expected protocol.Resp
	}{
		"Exists with too few args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("Exists")}}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'exists' command"}},
		"Exists with valid key": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("Exists")},
				protocol.BulkString{Data: protocol.Ptr("key")}}},
			expected: protocol.Integer{Value: 1}},
		"Exists with non existent key": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("Exists")},
				protocol.BulkString{Data: protocol.Ptr("invalid key")}}},
			expected: protocol.Integer{Value: 0}},
		"Exists with multiple keys": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("Exists")},
				protocol.BulkString{Data: protocol.Ptr("key")},
				protocol.BulkString{Data: protocol.Ptr("keyexists")},
				protocol.BulkString{Data: protocol.Ptr("invalid key")},
			}},
			expected: protocol.Integer{Value: 2}},
	}

	// Test Datastore
	ds := datastore.NewDatastore()
	ds.Set("key", "val")
	ds.Set("keyexists", "valexists")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := HandleCommand(test.in, ds)
			if err != nil {
				t.Errorf("HandleCommand() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("expected: %v, got: %v", test.expected, got)
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	tests := map[string]struct {
		in       protocol.Resp
		expected protocol.Resp
	}{
		"Too few args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("DEL")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'del' command"}},
		"Key not present": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("DEL")},
				protocol.BulkString{Data: protocol.Ptr("invalid key")},
			}},
			expected: protocol.Integer{Value: 0}},
		"One key there, one not": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("DEL")},
				protocol.BulkString{Data: protocol.Ptr("key")},
				protocol.BulkString{Data: protocol.Ptr("invalid key")},
			}},
			expected: protocol.Integer{Value: 1}},
		"Multiple keys there, multiple not": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("DEL")},
				protocol.BulkString{Data: protocol.Ptr("k1")},
				protocol.BulkString{Data: protocol.Ptr("invalid key1")},
				protocol.BulkString{Data: protocol.Ptr("k2")},
				protocol.BulkString{Data: protocol.Ptr("invalid key3")},
				protocol.BulkString{Data: protocol.Ptr("k3")},
				protocol.BulkString{Data: protocol.Ptr("k4")},
			}},
			expected: protocol.Integer{Value: 4}},
	}
	// Test Datastore
	ds := datastore.NewDatastore()
	ds.Set("key", "val")
	ds.Set("k1", "v1")
	ds.Set("k2", "v2")
	ds.Set("k3", "v3")
	ds.Set("k4", "v4")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := HandleCommand(test.in, ds)
			if err != nil {
				t.Errorf("HandleCommand() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("expected: %v, got: %v", test.expected, got)
			}
		})
	}
}

func TestIncrCommand(t *testing.T) {
	tests := map[string]struct {
		in       protocol.Resp
		expected protocol.Resp
	}{
		"Not enough args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'incr' command"}},
		"Too many args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
				protocol.BulkString{Data: protocol.Ptr("key1")},
				protocol.BulkString{Data: protocol.Ptr("key2")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'incr' command"}},
		"None existent key": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
				protocol.BulkString{Data: protocol.Ptr("key")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"}},
		"Key with string val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
				protocol.BulkString{Data: protocol.Ptr("keystring")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"}},
		"Key with int val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
				protocol.BulkString{Data: protocol.Ptr("keyint")},
			}},
			expected: protocol.Integer{Value: 2}},
		"Key with negative int val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("incr")},
				protocol.BulkString{Data: protocol.Ptr("keyintneg")},
			}},
			expected: protocol.Integer{Value: -2}},
	}
	ds := datastore.NewDatastore()
	ds.Set("keystring", "one")
	ds.Set("keyint", "1")
	ds.Set("keyintneg", "-3")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := HandleCommand(test.in, ds)
			if err != nil {
				t.Errorf("handleIncrCommand() error = %v", err)
			}
			if got != test.expected {
				t.Errorf("expected: %v, got: %v", test.expected, got)
			}
		})
	}
}

func TestDecrCommand(t *testing.T) {
	tests := map[string]struct {
		in       protocol.Resp
		expected protocol.Resp
	}{
		"Not enough args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'decr' command"}},
		"Too many args": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
				protocol.BulkString{Data: protocol.Ptr("key1")},
				protocol.BulkString{Data: protocol.Ptr("key2")},
			}},
			expected: protocol.Error{Data: "ERR wrong number of arguments for 'decr' command"}},
		"None existent key": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
				protocol.BulkString{Data: protocol.Ptr("key")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"}},
		"Key with string val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
				protocol.BulkString{Data: protocol.Ptr("keystring")},
			}},
			expected: protocol.Error{Data: "ERR value is not an integer or out of range"}},
		"Key with int val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
				protocol.BulkString{Data: protocol.Ptr("keyint")},
			}},
			expected: protocol.Integer{Value: 1}},
		"Key with negative int val": {
			in: protocol.Array{Items: []protocol.Resp{
				protocol.BulkString{Data: protocol.Ptr("decr")},
				protocol.BulkString{Data: protocol.Ptr("keyintneg")},
			}},
			expected: protocol.Integer{Value: -2}},
	}
	ds := datastore.NewDatastore()
	ds.Set("keystring", "one")
	ds.Set("keyint", "2")
	ds.Set("keyintneg", "-1")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := HandleCommand(test.in, ds)
			if err != nil {
				t.Errorf("handleIncrCommand() error = %v", err)
			}
			if got != test.expected {
				t.Errorf("expected: %v, got: %v", test.expected, got)
			}
		})
	}
}
