package protocol

import (
	"bytes"
	"strconv"
	"strings"
)

var (
	messageSeparatorS = "\r\n"
	messageSeparator  = []byte(messageSeparatorS)
	separatorLen      = len(messageSeparator)
)

type Resp interface {
	//Encodes the command to a byte array representation as per RESP
	Encode() []byte
	//User readable representation of the command
	String() string
}

// SimpleString represents a simple string as defined in Redis' RESP spec.
// The string transmits a short non-binary strings.
type SimpleString struct {
	Data string
}

// Error represents a simple error as defined in Redis' RESP spec.
// The error is similar to SimpleString, but is used for error handling.
type Error struct {
	Data string
}

// Integer represents a signed, base-10, 64-bit integer as defined in Redis' RESP spec.
type Integer struct {
	Value int64
}

// BulkString represents a single binary string as defined in Redis' RESP spec.
// The string can be of any size, and is configured by the Redis config.
type BulkString struct {
	Data *string
}

// Array is an implememtation of the RESP arrays.
// Some Redis commands that return collections of elements use arrays as their replies.
type Array struct {
	Items []Resp
}

func (r SimpleString) Encode() []byte {
	prefix := '+'
	var buf bytes.Buffer
	buf.WriteRune(prefix)
	buf.WriteString(r.Data)
	buf.WriteString(messageSeparatorS)
	return buf.Bytes()
}

func (r SimpleString) String() string {
	return r.Data
}

func (r Error) Encode() []byte {
	prefix := '-'
	var buf bytes.Buffer
	buf.WriteRune(prefix)
	buf.WriteString(r.Data)
	buf.WriteString(messageSeparatorS)
	return buf.Bytes()
}

func (r Error) String() string {
	return r.Data
}

func (r Integer) Encode() []byte {
	prefix := ':'
	var buf bytes.Buffer
	buf.WriteRune(prefix)
	buf.WriteString(string(strconv.AppendInt(nil, r.Value, 10)))
	buf.WriteString(messageSeparatorS)
	return buf.Bytes()
}

func (r Integer) String() string {
	return strconv.FormatInt(r.Value, 10)
}

func (r BulkString) Encode() []byte {
	prefix := '$'
	var buf bytes.Buffer
	if r.Data != nil {
		buf.WriteRune(prefix)
		buf.WriteString(strconv.Itoa(len(*r.Data)))
		buf.WriteString(messageSeparatorS)
		buf.WriteString(*r.Data)
		buf.WriteString(messageSeparatorS)
	} else {
		buf.WriteRune(prefix)
		buf.WriteString(strconv.Itoa(-1))
		buf.WriteString(messageSeparatorS)
	}
	return buf.Bytes()
}

func (r BulkString) String() string {
	return Val(r.Data)
}

func (r Array) Encode() []byte {
	prefix := '*'
	var buf bytes.Buffer
	buf.WriteRune(prefix)
	if r.Items != nil {
		buf.WriteString(strconv.Itoa(len(r.Items)))
		buf.WriteString(messageSeparatorS)
		for _, item := range r.Items {
			enc := item.Encode()
			buf.Write(enc)
		}
	} else {
		buf.WriteString(strconv.Itoa(-1))
		buf.WriteString(messageSeparatorS)
	}
	return buf.Bytes()
}

func (r Array) String() string {
	var sb strings.Builder
	for _, item := range r.Items {
		sb.WriteString(item.String())
		sb.WriteString(" ")
	}
	return sb.String()
}

func Ptr[T any](t T) *T {
	return &t
}

func Val[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// Extracts a frame from the provided buffer and returns the amount of bytes read.
func ExtractFrameFromBuffer(buffer []byte) (Resp, int) {
	sep := bytes.Index(buffer, messageSeparator)
	if sep == -1 {
		return nil, 0
	}
	switch b := buffer[0]; b {
	case '+':
		return SimpleString{Data: string(buffer[1:sep])}, sep + separatorLen
	case '-':
		return Error{Data: string(buffer[1:sep])}, sep + separatorLen
	case ':':
		v, err := strconv.ParseInt(string(buffer[1:sep]), 10, 64)
		if err != nil {
			return nil, 0
		}
		return Integer{Value: v}, sep + separatorLen
	case '$':
		l, err := strconv.Atoi(string(buffer[1:sep]))
		if err != nil {
			return nil, 0
		}
		if l == -1 {
			return nil, 5
		}
		if len(buffer) < sep+separatorLen+l+separatorLen {
			return nil, 0
		}
		eom := sep + separatorLen + l
		return BulkString{Data: Ptr(string(buffer[sep+2 : eom]))}, eom + separatorLen
	case '*':
		l, err := strconv.Atoi(string(buffer[1:sep]))
		if err != nil {
			return nil, 0
		}
		if l == -1 {
			return nil, sep + separatorLen
		}
		if l == 0 {
			return Array{Items: make([]Resp, 0)}, sep + separatorLen
		}
		items := make([]Resp, 0, l)
		var nextItem Resp
		var length = 0
		for i := 0; i < l; i++ {
			nextItem, length = ExtractFrameFromBuffer(buffer[sep+2:])
			if nextItem != nil && length > 0 {
				items = append(items, nextItem)
				sep += length
			} else {
				return nil, 0
			}
		}
		return Array{Items: items}, sep + separatorLen
	}
	return nil, 0
}

func Encode[T Resp](resp T) []byte {
	return resp.Encode()
}
