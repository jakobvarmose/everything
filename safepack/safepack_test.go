package main

import (
	"bytes"
	"testing"
)

type test struct {
	input    []byte
	expected interface{}
}

var tests = []test{
	// nil
	{[]byte{0xc0}, nil},

	// bool
	{[]byte{0xc2}, false},
	{[]byte{0xc3}, true},

	// int
	{[]byte{0x00}, 0},
	{[]byte{0x7f}, 127},
	{[]byte{0xff}, -1},
	{[]byte{0xe0}, -32},
	{[]byte{0xcc, 0x12}, 0x12},
	{[]byte{0xcd, 0x12, 0x34}, 0x1234},
	{[]byte{0xce, 0x12, 0x34, 0x56, 0x78}, 0x12345678},
	{[]byte{0xcf, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, 0x123456789abcdef0},
	{[]byte{0xd0, 0x12}, 0x12},
	{[]byte{0xd1, 0x12, 0x34}, 0x1234},
	{[]byte{0xd2, 0x12, 0x34, 0x56, 0x78}, 0x12345678},
	{[]byte{0xd3, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, 0x123456789abcdef0},
	//TODO also test negative values and large unsigned values

	// float
	{[]byte{0xca, 0x3f, 0xc0, 0x00, 0x00}, 1.5},
	{[]byte{0xcb, 0x3f, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 1.5},

	// str
	{[]byte{0xa0}, ""},
	{[]byte{0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello"},
	{[]byte{0xd9, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello"},
	{[]byte{0xda, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello"},
	{[]byte{0xdb, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, "hello"},

	// bin
	{[]byte{0xc4, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, []byte("hello")},
	{[]byte{0xc5, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, []byte("hello")},
	{[]byte{0xc6, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, []byte("hello")},
}

func verify(t *testing.T, test test, expected Type) {
	input := MustNew(test.input)
	result := input.Type()
	if result != expected {
		t.Errorf("%v Type: %v, expected %v", test.input, result, expected)
	}
}

func TestTable(t *testing.T) {
	//TODO implement this
	for _, test := range tests {
		input, err := New(test.input)
		if err != nil {
			t.Errorf("%v Value: %s", test.input, err)
		}
		switch expected := test.expected.(type) {
		case nil:
			verify(t, test, Type_Nil)
		case bool:
			verify(t, test, Type_Bool)
			result := input.Bool()
			if result != expected {
				t.Errorf("%v Value: %v, expected %v", test.input, result, expected)
			}
		case int:
			verify(t, test, Type_Int)
			result := input.Int()
			if result != expected {
				t.Errorf("%v Value: %v, expected %v", test.input, result, expected)
			}
		case float64:
			verify(t, test, Type_Float)
			result := input.Float64()
			if result != expected {
				t.Errorf("%v Value: %v, expected %v", test.input, result, expected)
			}
		case string:
			verify(t, test, Type_String)
			result := input.String()
			if result != expected {
				t.Errorf("%v Value: %v, expected %v", test.input, result, expected)
			}
		case []byte:
			verify(t, test, Type_Bin)
			result := input.Bytes()
			if !bytes.Equal(result, expected) {
				t.Errorf("%v Value: %v, expected %v", test.input, result, expected)
			}
		default:
			panic("NOT TESTED")
		}
		if _, ok := test.expected.(bool); !ok {
			result := input.Bool()
			if result != false {
				t.Errorf("%v Value: %v, expected false", test.input, result)
			}
		}
		if _, ok := test.expected.(int); !ok {
			result := input.Int()
			if result != 0 {
				t.Errorf("%v Value: %v, expected 0", test.input, result)
			}
		}
		if _, ok := test.expected.(float64); !ok {
			result := input.Float64()
			if result != 0.0 {
				t.Errorf("%v Value: %v, expected 0.0", test.input, result)
			}
		}
		_, ok1 := test.expected.(string)
		_, ok2 := test.expected.([]byte)
		if !ok1 && !ok2 {
			result := input.Bytes()
			if result != nil {
				t.Errorf("%v Value: %v, expected \"\"", test.input, result)
			}
		}
	}
}

func TestTypes(t *testing.T) {
	x, err := New(nil)
	if err == nil {
		t.Error("Nil error")
	}
	if x.Type() != Type_Invalid {
		t.Error("Invalid type")
	}
	if MustNew([]byte{0xc0}).Type() != Type_Nil {
		t.Error("Invalid type")
	}
}

func TestString(t *testing.T) {
	hello := MustNew([]byte("\xa5hello"))
	if hello.Type() != Type_String {
		t.Error("invalid type")
	}
	if hello.String() != "hello" {
		t.Error("invalid type")
	}
	if string(hello.Bytes()) != "hello" {
		t.Error("invalid type")
	}
}

func TestFloat(t *testing.T) {

}

func TestInt(t *testing.T) {
	if MustNew([]byte{0x00}).Int() != 0 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0x7f}).Int() != 127 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0xff}).Int() != -1 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0xe0}).Int() != -32 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0xcc, 0x12}).Int() != 0x12 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0xcd, 0x12, 0x34}).Int() != 0x1234 {
		t.Error("invalid value")
	}
	if MustNew([]byte{0xce, 0x12, 0x34, 0x56, 0x78}).Int() != 0x12345678 {
		t.Error("invalid value")
	}
}
