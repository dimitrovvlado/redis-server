package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestProtocolParser(t *testing.T) {
	tests := map[string]struct {
		buffer       []byte
		expected     Resp
		expectedSize int
	}{
		// Test cases for Simple strings
		"Partial message":                         {buffer: []byte("+Par"), expected: nil, expectedSize: 0},
		"Full simple string":                      {buffer: []byte("+OK\r\n"), expected: SimpleString{"OK"}, expectedSize: 5},
		"Full, followed by partial simple string": {buffer: []byte("+OK\r\n+Next"), expected: SimpleString{"OK"}, expectedSize: 5},

		// Test cases for Errors
		"Partial error":                   {buffer: []byte("-Err"), expected: nil, expectedSize: 0},
		"Full error":                      {buffer: []byte("-Error Message\r\n"), expected: Error{"Error Message"}, expectedSize: 16},
		"Full, followed by partial error": {buffer: []byte("-Error Message\r\n+Other"), expected: Error{"Error Message"}, expectedSize: 16},

		// Test cases for Integers
		"Partial integer":                   {buffer: []byte(":10"), expected: nil, expectedSize: 0},
		"Full Integer":                      {buffer: []byte(":100\r\n"), expected: Integer{100}, expectedSize: 6},
		"Full, followed by partial integer": {buffer: []byte(":100\r\n+OK"), expected: Integer{100}, expectedSize: 6},

		// Test cases for Bulk Strings
		"Partial bulk string":        {buffer: []byte("$5\r\nHel"), expected: nil, expectedSize: 0},
		"Full bulk string":           {buffer: []byte("$5\r\nHello\r\n"), expected: BulkString{Ptr("Hello")}, expectedSize: 11},
		"Longer bulk string":         {buffer: []byte("$12\r\nHello, World\r\n"), expected: BulkString{Ptr("Hello, World")}, expectedSize: 19},
		"Bulk string with separator": {buffer: []byte("$12\r\nHello\r\nWorld\r\n"), expected: BulkString{Ptr("Hello\r\nWorld")}, expectedSize: 19},
		"Empty bulk string":          {buffer: []byte("$0\r\n\r\n"), expected: BulkString{Ptr("")}, expectedSize: 6},
		"Null":                       {buffer: []byte("$-1\r\n"), expected: BulkString{nil}, expectedSize: 5},

		// Test cases for Arrays
		"Partial array":                {buffer: []byte("*0"), expected: nil, expectedSize: 0},
		"Empty array":                  {buffer: []byte("*0\r\n"), expected: nil, expectedSize: 4},
		"Null (array version)":         {buffer: []byte("*-1\r\n"), expected: Array{nil}, expectedSize: 5},
		"Partial array with Data":      {buffer: []byte("*2\r\n$5\r\nhello\r\n$5\r\n"), expected: nil, expectedSize: 0},
		"Array of two bulk strings":    {buffer: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), expected: Array{[]Resp{BulkString{Ptr("hello")}, BulkString{Ptr("world")}}}, expectedSize: 26},
		"Array of two BS and parial":   {buffer: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n+OK"), expected: Array{[]Resp{BulkString{Ptr("hello")}, BulkString{Ptr("world")}}}, expectedSize: 26},
		"Partial array of ints":        {buffer: []byte("*3\r\n:1\r\n:"), expected: nil, expectedSize: 0},
		"Array of ints":                {buffer: []byte("*3\r\n:1\r\n:2\r\n:3\r\n"), expected: Array{[]Resp{Integer{1}, Integer{2}, Integer{3}}}, expectedSize: 16},
		"Array of ints and by partial": {buffer: []byte("*3\r\n:1\r\n:2\r\n:3\r\n+OK"), expected: Array{[]Resp{Integer{1}, Integer{2}, Integer{3}}}, expectedSize: 16},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			frame, frameSize := ExtractFrameFromBuffer(test.buffer)
			if frameSize != test.expectedSize {
				t.Errorf("Incorrect framesize: %d", frameSize)
			}
			if frame != nil && test.expected != nil && !reflect.DeepEqual(frame, test.expected) {
				t.Errorf("Failed to parse frame")
			}
		})
	}
}

func TestProtocolEncoding(t *testing.T) {
	tests := map[string]struct {
		in       Resp
		expected []byte
	}{
		// Tests
		"Simple String":                  {SimpleString{"OK"}, []byte("+OK\r\n")},
		"Error":                          {Error{"Error"}, []byte("-Error\r\n")},
		"Integer":                        {Integer{100}, []byte(":100\r\n")},
		"Null Bulk String":               {BulkString{nil}, []byte("$-1\r\n")},
		"Empty Bulk String":              {BulkString{Ptr("")}, []byte("$0\r\n\r\n")},
		"Simple Bulk String":             {BulkString{Ptr("Simple string")}, []byte("$13\r\nSimple string\r\n")},
		"Bulk String - embedded newline": {BulkString{Ptr("Hello\r\nWorld")}, []byte("$12\r\nHello\r\nWorld\r\n")},
		"Null array":                     {Array{nil}, []byte("*-1\r\n")},
		"Empty array":                    {Array{[]Resp{}}, []byte("*0\r\n")},
		"Array - Bulk Strings":           {Array{[]Resp{BulkString{Ptr("Hello\r\nWorld")}, BulkString{Ptr("Coding\r\nChallenges")}}}, []byte("*2\r\n$12\r\nHello\r\nWorld\r\n$18\r\nCoding\r\nChallenges\r\n")},
		"Array - Bulk String and Int":    {Array{[]Resp{BulkString{Ptr("Hello\r\nWorld")}, Integer{42}}}, []byte("*2\r\n$12\r\nHello\r\nWorld\r\n:42\r\n")},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Encode(test.in)
			if !bytes.Equal(got, test.expected) {
				t.Errorf("Incorrect Encoding got: '%s' wanted: '%s'", got, test.expected)
			}

		})
	}
}
