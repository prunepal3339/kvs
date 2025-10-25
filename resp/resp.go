package resp

import (
	"bufio"
	"errors"
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

type Tag byte

const (
	TAG_NIL Tag = iota
	TAG_STR
	TAG_BULK
	TAG_INT
	TAG_ARR
	TAG_ERR
)

func (t Tag) String() string {
	switch t {
	case TAG_NIL:
		return "nil"
	case TAG_STR:
		return "string"
	case TAG_BULK:
		return "bulk"
	case TAG_INT:
		return "int"
	case TAG_ARR:
		return "array"
	case TAG_ERR:
		return "error"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(t))
	}
}

type Value struct {
	tag Tag
	val any
}

func (v Value) Tag() Tag {
	return v.tag
}
func (v Value) Val() any {
	return v.val
}

func (v Value) String() string {
	switch v.tag {
	case TAG_NIL:
		return "nil"
	case TAG_STR:
		return fmt.Sprintf("(string %s)", v.val.(string))
	case TAG_ARR:
		arrValue := v.val.([]Value)
		var s string
		for i := 0; i < len(arrValue); i++ {
			s += arrValue[i].String()
		}
		return fmt.Sprintf("(array %s)", s)
	case TAG_ERR:
		return fmt.Sprintf("(error %s)", v.val.(string))
	case TAG_INT:
		return fmt.Sprintf("(int %d)", v.val.(int))
	case TAG_BULK:
		return fmt.Sprintf("(bulk %s)", v.val.(string))
	default:
		return "(unknown value)"
	}
}

func NewValue(tag Tag, val any) Value {
	return Value{
		tag,
		val,
	}
}
func (v Value) Marshal() []byte {
	switch v.tag {
	case TAG_ARR:
		return v.marshalArray()
	case TAG_BULK:
		return v.marshalBulk()
	case TAG_NIL:
		return v.marshalNil()
	case TAG_STR:
		return v.marshalString()
	case TAG_ERR:
		return v.marshalError()
	case TAG_INT:
		return v.marshalInt()
	default:
		return []byte{}
	}
}
func (v Value) marshalInt() []byte {
	var bytes []byte
	bytes = append(bytes, INTEGER)
	intval := v.val.(int)
	bytes = append(bytes, strconv.Itoa(intval)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}
func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	strval := v.val.(string)
	bytes = append(bytes, strval...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalArray() []byte {
	var bytes []byte
	bytes = append(bytes, ARRAY)
	arrval := v.val.([]Value)
	size := len(arrval)
	bytes = append(bytes, strconv.Itoa(size)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < size; i++ {
		bytes = append(bytes, arrval[i].Marshal()...)
	}
	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	strval := v.val.(string)
	length := len(strval)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')

	bytes = append(bytes, strval...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNil() []byte {
	var bytes []byte
	bytes = append(bytes, '$', '-', '1', '\r', '\n')
	return bytes
}
func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	strval := v.val.(string)
	bytes = append(bytes, strval...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}
func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
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
	case ERROR:
		return r.readError()
	case STRING:
		return r.readString()
	case INTEGER:
		return r.readInt()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}
func (r *Resp) readInt() (Value, error) {
	v := Value{}
	v.tag = TAG_INT
	ival, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	v.val = ival

	return v, nil
}
func (r *Resp) readError() (Value, error) {
	v := Value{}
	v.tag = TAG_ERR

	errval, err := r.readString()

	if err != nil {
		return v, err
	}
	v.val = errval

	return v, nil
}
func (r *Resp) readString() (Value, error) {
	v := Value{}
	v.tag = TAG_STR
	strval, err := r.reader.ReadString('\r')
	if err != nil {
		return v, err
	}
	lf, err := r.reader.ReadByte()
	if err != nil {
		return v, err
	}
	if lf != '\n' {
		return v, errors.New("invalid line terminator")
	}

	v.val = strval

	return v, nil
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
