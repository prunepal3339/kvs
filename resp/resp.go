package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

//	type Value struct {
//		typ   string
//		str   string
//		num   int
//		bulk  string
//		array []value
//	}
type Tag string

const (
	TAG_NIL  Tag = "nil"
	TAG_STR  Tag = "str"
	TAG_BULK Tag = "bulk"
	TAG_INT  Tag = "int"
	TAG_ARR  Tag = "array"
	TAG_ERR  Tag = "error"
)

type Value struct {
	tag Tag
	val any
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

// Read the buffer to return the Value.
func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}
func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.tag = TAG_ARR

	//read length of the array.
	length, _, err := r.readInteger()
	if err != nil {
		_ = length
		return v, err
	}
	v.val = make([]Value, length)
	arr := v.val.([]Value)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		arr[i] = val
	}
	v.val = arr
	return v, nil
}
func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.tag = TAG_BULK

	//read the length of the string.
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	bulk := make([]byte, length)
	r.reader.Read(bulk)
	v.val = string(bulk)
	r.readLine()
	return v, nil
}
