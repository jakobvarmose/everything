package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type Type int

const (
	Type_Invalid = iota
	Type_Int
	Type_Map
	Type_Array
	Type_String
	Type_Nil
	Type_Bool
	Type_Bin
	Type_Float
)

type Object struct {
	d []byte
	n int
}

func New(buf []byte) (Object, error) {
	if len(buf) == 0 {
		return Object{}, errors.New("empty data")
	}
	item, rest := read(buf)
	if len(rest) != 0 {
		return Object{}, errors.New("invalid data")
	}
	return Object{item, 0}, nil
}

func MustNew(buf []byte) Object {
	x, err := New(buf)
	if err != nil {
		panic(err)
	}
	return x
}

func (x Object) headerSize() int {
	if len(x.d) < 1 {
		return 1
	}
	if x.d[0] < 0xc4 || x.d[0] >= 0xe0 {
		return 1
	}
	m := [0xe0 - 0xc4]byte{
		2, 3, 5,
		2, 3, 5,
		5, 9,
		2, 3, 5, 9,
		2, 3, 5, 9,
		1, 1, 1, 1, 1,
		2, 3, 5,
		3, 5,
		3, 5,
	}
	return int(m[x.d[0]-0xc4])
}

func (x Object) Type() Type {
	if len(x.d) == 0 {
		if x.n != 0 {
			return Type_Int
		}
		return Type_Invalid
	}
	if x.d[0] < 0xc0 {
		if x.d[0] < 0x80 {
			return Type_Int
		} else if x.d[0] < 0x90 {
			return Type_Map
		} else if x.d[0] < 0xa0 {
			return Type_Array
		} else {
			return Type_String
		}
	}
	if x.d[0] >= 0xe0 {
		return Type_Int
	}
	m := [0xe0 - 0xc0]byte{
		Type_Nil, Type_Invalid, Type_Bool, Type_Bool,
		Type_Bin, Type_Bin, Type_Bin,
		Type_Invalid, Type_Invalid, Type_Invalid,
		Type_Float, Type_Float,
		Type_Int, Type_Int, Type_Int, Type_Int,
		Type_Int, Type_Int, Type_Int, Type_Int,
		Type_Invalid, Type_Invalid, Type_Invalid, Type_Invalid, Type_Invalid,
		Type_String, Type_String, Type_String,
		Type_Array, Type_Array,
		Type_Map, Type_Map,
	}
	return Type(m[x.d[0]-0xc0])

	/*if x.d[0] <= 0x8f {
		return Type_Map
	} else if x.d[0] <= 0x9f {
		return Type_Array
	} else if x.d[0] <= 0xbf {
		return Type_String
	} else if x.d[0] <= 0xc0 {
		return Type_Nil
	} else if x.d[0] <= 0xc1 {
		return Type_Invalid
	} else if x.d[0] <= 0xc3 {
		return Type_Bool
	} else if x.d[0] <= 0xc6 {
		return Type_Bin
	} else if x.d[0] <= 0xc9 {
		return Type_Invalid
	} else if x.d[0] <= 0xcb {
		return Type_Float
	} else if x.d[0] <= 0xd3 {
		return Type_Int
	} else if x.d[0] <= 0xd8 {
		return Type_Invalid
	} else if x.d[0] <= 0xdb {
		return Type_String
	} else if x.d[0] <= 0xdd {
		return Type_Array
	} else if x.d[0] <= 0xdf {
		return Type_Map
	} else {
		return Type_Int
	}*/
}

func (x Object) int64() (int64, bool) {
	if len(x.d) == 1 && x.d[0] >= 0xe0 { // fixint
		return int64(int8(x.d[0])), true
	}
	if len(x.d) == 2 && x.d[0] == 0xd0 { // int 8
		return int64(int8(x.d[1])), true
	}
	if len(x.d) == 3 && x.d[0] == 0xd1 { // int 16
		return int64(int16(binary.BigEndian.Uint16(x.d[1:]))), true
	}
	if len(x.d) == 5 && x.d[0] == 0xd2 { // int 32
		return int64(int32(binary.BigEndian.Uint32(x.d[1:]))), true
	}
	if len(x.d) == 9 && x.d[0] == 0xd3 { // int 64
		return int64(binary.BigEndian.Uint64(x.d[1:])), true
	}
	return 0, false
}

func (x Object) uint64() (uint64, bool) {
	if len(x.d) == 1 && x.d[0] <= 0x7f { // fixint
		return uint64(x.d[0]), true
	}
	if len(x.d) == 2 && x.d[0] == 0xcc { // uint 8
		return uint64(x.d[1]), true
	}
	if len(x.d) == 3 && x.d[0] == 0xcd { // uint 16
		return uint64(binary.BigEndian.Uint16(x.d[1:])), true
	}
	if len(x.d) == 5 && x.d[0] == 0xce { // uint 32
		return uint64(binary.BigEndian.Uint32(x.d[1:])), true
	}
	if len(x.d) == 9 && x.d[0] == 0xcf { // uint 64
		return binary.BigEndian.Uint64(x.d[1:]), true
	}
	return 0, false
}

func (x Object) float64() (float64, bool) {
	if len(x.d) == 5 && x.d[0] == 0xca { // float 32
		var f32 float32
		binary.Read(bytes.NewReader(x.d[1:]), binary.BigEndian, &f32)
		return float64(f32), true
	}
	if len(x.d) == 9 && x.d[0] == 0xcb { // float 64
		var f64 float64
		binary.Read(bytes.NewReader(x.d[1:]), binary.BigEndian, &f64)
		return f64, true
	}
	return 0, false
}

func (x Object) Int() int {
	if val, ok := x.uint64(); ok {
		return int(val)
	}
	if val, ok := x.int64(); ok {
		return int(val)
	}
	if x.n != 0 {
		return x.n - 1
	}
	return 0
}

func (x Object) Int64() int64 {
	if val, ok := x.uint64(); ok {
		return int64(val)
	}
	if val, ok := x.int64(); ok {
		return val
	}
	if x.n != 0 {
		return int64(x.n - 1)
	}
	return 0
}

func (x Object) Float64() float64 {
	if val, ok := x.float64(); ok {
		return val
	}
	return 0
}

func (x Object) String() string {
	return string(x.Bytes())
}

func (x Object) Bytes() []byte {
	if len(x.d) < 1 {
		return nil
	}
	if x.d[0] >= 0xa0 && x.d[0] <= 0xbf { // fixstr
		if int(x.d[0]-0xa0) != len(x.d[1:]) {
			return nil
		}
		return x.d[1:]
	}
	if x.d[0] == 0xc4 || x.d[0] == 0xd9 {
		if len(x.d) < 2 {
			return nil
		}
		if int(x.d[1]) != len(x.d[2:]) {
			return nil
		}
		return x.d[2:]
	}
	if x.d[0] == 0xc5 || x.d[0] == 0xda {
		if len(x.d) < 3 {
			return nil
		}
		if int(binary.BigEndian.Uint16(x.d[1:3])) != len(x.d[3:]) {
			return nil
		}
		return x.d[3:]
	}
	if x.d[0] == 0xc6 || x.d[0] == 0xdb {
		if len(x.d) < 5 {
			return nil
		}
		if int(binary.BigEndian.Uint32(x.d[1:5])) != len(x.d[5:]) {
			return nil
		}
		return x.d[5:]
	}
	return nil
}

func (x Object) Bool() bool {
	return len(x.d) == 1 && x.d[0] == 0xc3
}

func sizeMap(b []byte, s int, n int) int {
	for i := 0; i < n; i++ {
		s += size(b[s:])
		s += size(b[s:])
	}
	return s
}

func sizeArray(b []byte, s int, n int) int {
	for i := 0; i < n; i++ {
		s += size(b[s:])
	}
	return s
}

func size(b []byte) int {
	if b[0] <= 0x7f { // positive fixint
		return 1
	} else if b[0] <= 0x8f { // fixmap
		return sizeMap(b, 1, int(b[0]-0x80))
	} else if b[0] <= 0x9f { // fixarray
		return sizeArray(b, 1, int(b[0]-0x90))
	} else if b[0] <= 0xbf { // fixstr
		s := int(b[0] - 0xa0)
		return 1 + s
	} else if b[0] <= 0xc0 { // nil
		return 1
	} else if b[0] <= 0xc1 { // invalid
		return 0
	} else if b[0] <= 0xc3 { // bool
		return 1
	} else if b[0] == 0xc4 { // bin 8
		return 2 + int(b[1])
	} else if b[0] == 0xc5 { // bin 16
		s := binary.BigEndian.Uint16(b[1:3])
		return 3 + int(s) // FIXME is this safe?
	} else if b[0] == 0xc6 { // bin 32
		s := binary.BigEndian.Uint32(b[1:5])
		return 5 + int(s) // FIXME is this safe?
	} else if b[0] <= 0xc9 {
		return 0
	} else if b[0] == 0xca {
		return 5
	} else if b[0] == 0xcb {
		return 9
	} else if b[0] == 0xcc {
		return 2
	} else if b[0] == 0xcd {
		return 3
	} else if b[0] == 0xce {
		return 5
	} else if b[0] == 0xcf {
		return 9
	} else if b[0] == 0xd0 {
		return 2
	} else if b[0] == 0xd1 {
		return 3
	} else if b[0] == 0xd2 {
		return 5
	} else if b[0] == 0xd3 {
		return 9
	} else if b[0] <= 0xd8 {
		return 0
	} else if b[0] == 0xd9 {
		return 2 + int(b[1])
	} else if b[0] == 0xda {
		s := binary.BigEndian.Uint16(b[1:3])
		return 3 + int(s) // FIXME is this safe?
	} else if b[0] == 0xdb {
		s := binary.BigEndian.Uint32(b[1:5])
		return 5 + int(s) // FIXME is this safe?
	} else if b[0] == 0xdc {
		n := binary.BigEndian.Uint16(b[1:3])
		return sizeArray(b, 3, int(n))
	} else if b[0] == 0xdd {
		n := binary.BigEndian.Uint32(b[1:5])
		return sizeArray(b, 5, int(n))
	} else if b[0] == 0xde {
		n := binary.BigEndian.Uint16(b[1:3])
		return sizeMap(b, 3, int(n))
	} else if b[0] == 0xdf {
		n := binary.BigEndian.Uint32(b[1:5])
		return sizeMap(b, 5, int(n))
	} else {
		return 1
	}
}

func read(b []byte) (y []byte, z []byte) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err, hex.EncodeToString(b))
			z = b
		}
	}()
	m := size(b)
	return b[:m], b[m:]
}

func (x Object) equal(val interface{}) bool {
	switch val := val.(type) {
	case string:
		if x.Type() == Type_String {
			return bytes.Equal(x.Bytes(), []byte(val))
		}
	case []byte:
		if x.Type() == Type_Bin {
			return bytes.Equal(x.Bytes(), val)
		}
	default:
		panic("NOT implemented")
	}
	return false
}

func (x Object) Len() int {
	if len(x.d) < 1 {
		return -1
	}
	if x.d[0] >= 0x80 && x.d[0] <= 0x8f { // fixmap
		return int(x.d[0] - 0x80)
	} else if x.d[0] >= 0x90 && x.d[0] <= 0x9f { // fixarray
		return int(x.d[0] - 0x90)
	} else if x.d[0] == 0xdc || x.d[0] == 0xde { // array 16 / map 16
		return int(binary.BigEndian.Uint16(x.d[1:3]))
	} else if x.d[0] == 0xdd || x.d[0] == 0xdf { // array 32 / map 32
		size := int(binary.BigEndian.Uint32(x.d[1:3]))
		if size < 0 { // handle overflow on 32 bit platforms
			return -1
		}
		return size
	}
	return -1
}

func (x Object) ArrayEach(callback func(int, Object)) {
	if x.d[0] >= 0x90 && x.d[0] <= 0x9f { // fixarray
		arrayEach(x.d[1:], callback)
	} else if x.d[0] == 0xdc { // array 16
		arrayEach(x.d[3:], callback)
	} else if x.d[0] == 0xdd { // array 32
		arrayEach(x.d[5:], callback)
	}
}

func arrayEach(buf []byte, callback func(int, Object)) {
	i := 0
	for len(buf) > 0 {
		item, rest := read(buf)
		buf = rest
		if item == nil {
			return
		}
		callback(i, Object{item, 0})
		i++
	}
}

func (x Object) ArrayGet(index int) Object {
	var res Object
	x.ArrayEach(func(i int, obj Object) {
		if i == index {
			res = obj
		}
	})
	return res
}

func (x Object) MapEach(callback func(Object, Object)) {
	if x.d[0] >= 0x80 && x.d[0] <= 0x8f { // fixmap
		mapEach(x.d[1:], callback)
	} else if x.d[0] == 0xde { // map 16
		mapEach(x.d[3:], callback)
	} else if x.d[0] == 0xdf { // map 32
		mapEach(x.d[5:], callback)
	}
}

func mapEach(buf []byte, callback func(Object, Object)) {
	for len(buf) > 0 {
		key, rest := read(buf)
		if key == nil {
			return
		}
		value, rest := read(rest)
		if value == nil {
			return
		}
		buf = rest
		callback(Object{key, 0}, Object{value, 0})
	}
}

func (x Object) MapGet(key interface{}) Object {
	var r Object
	x.MapEach(func(k Object, v Object) {
		if k.equal(key) {
			r = v
		}
	})
	return r
}

func main() {
	x, err := New([]byte{0x83, 0xa1, 0x58, 0x08, 0xa0, 0x0a, 0xa1, 0x37, 0x0c})
	if err != nil {
		panic(err)
	}
	fmt.Println(x.Type())
	fmt.Println(x.Len())
	fmt.Printf("X: %v\n", x.MapGet("X"))
	x.MapEach(func(i Object, item Object) {
		fmt.Printf("%v = %v\n", i.String(), item.Int())
	})
}
