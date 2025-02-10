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
			expected: protocol.Error{Data: "ERR syntax error"}},
		"Set with 1 too few args": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key1")}}},
			expected: protocol.Error{Data: "ERR syntax error"}},
		"Set with existing key": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key")}, protocol.BulkString{Data: protocol.Ptr("value")}}},
			expected: protocol.SimpleString{Data: "OK"}},
		"Set with non existent key": {
			in:       protocol.Array{Items: []protocol.Resp{protocol.BulkString{Data: protocol.Ptr("set")}, protocol.BulkString{Data: protocol.Ptr("key1")}, protocol.BulkString{Data: protocol.Ptr("value1")}}},
			expected: protocol.SimpleString{Data: "OK"}},
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
