package protocol

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var (
	messageSeparator = []byte("\r\n")
	separatorLen     = len(messageSeparator)
)

type Resp interface {
	Encode() []byte
	String() string
}

type SimpleString struct {
	Data string
}

type Error struct {
	Data string
}

type Integer struct {
	Value int64
}

type BulkString struct {
	Data *string
}

type Array struct {
	Items []Resp
}

func (r SimpleString) Encode() []byte {
	prefix := "+"
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s%s%s", prefix, r.Data, messageSeparator)
	return buf.Bytes()
}

func (r SimpleString) String() string {
	return r.Data
}

func (r Error) Encode() []byte {
	prefix := "-"
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s%s%s", prefix, r.Data, messageSeparator)
	return buf.Bytes()
}

func (r Error) String() string {
	return r.Data
}

func (r Integer) Encode() []byte {
	prefix := ":"
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s%d%s", prefix, r.Value, messageSeparator)
	return buf.Bytes()
}

func (r Integer) String() string {
	return strconv.FormatInt(r.Value, 10)
}

func (r BulkString) Encode() []byte {
	prefix := "$"
	var buf bytes.Buffer
	if r.Data != nil {
		fmt.Fprintf(&buf, "%s%d%s%s%s", prefix, len(*r.Data), messageSeparator, *r.Data, messageSeparator)
	} else {
		fmt.Fprintf(&buf, "%s-1%s", prefix, messageSeparator)
	}
	return buf.Bytes()
}

func (r BulkString) String() string {
	return Val(r.Data)
}

func (r Array) Encode() []byte {
	prefix := "*"
	var buf bytes.Buffer
	if r.Items != nil {
		fmt.Fprintf(&buf, "%s%d%s", prefix, len(r.Items), messageSeparator)
		for _, item := range r.Items {
			enc := item.Encode()
			buf.Write(enc)
		}
	} else {
		fmt.Fprintf(&buf, "%s-1%s", prefix, messageSeparator)
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
