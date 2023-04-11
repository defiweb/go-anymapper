package anymapper

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltInTypes(t *testing.T) {
	type (
		myBool   bool
		myInt    int
		myUint   uint
		myFloat  float64
		myString string
		mySlice  []string
		myArray  [1]string
		myMap    map[string]string
	)

	tests := []struct {
		name string
		src  any
		dst  any
		exp  any
		err  bool
	}{
		// bool <-> bool
		{name: `bool(true)->bool`, src: true, dst: new(bool), exp: true},
		{name: `bool(false)->bool`, src: false, dst: new(bool), exp: false},
		{name: `bool(true)->myBool`, src: true, dst: new(myBool), exp: myBool(true)},
		{name: `myBool(true)->bool`, src: myBool(true), dst: new(bool), exp: true},

		// bool <-> int
		{name: `bool(true)->int`, src: true, dst: new(int), exp: 1},
		{name: `bool(false)->int`, src: false, dst: new(int), exp: 0},
		{name: `int(1)->bool`, src: 1, dst: new(bool), exp: true},
		{name: `int(0)->bool`, src: 0, dst: new(bool), exp: false},

		// bool <-> uint
		{name: `bool(true)->uint`, src: true, dst: new(uint), exp: uint(1)},
		{name: `bool(false)->uint`, src: false, dst: new(uint), exp: uint(0)},
		{name: `uint(1)->bool`, src: uint(1), dst: new(bool), exp: true},
		{name: `uint(0)->bool`, src: uint(0), dst: new(bool), exp: false},

		// bool <-> float
		{name: `bool(true)->float64`, src: true, dst: new(float64), exp: float64(1)},
		{name: `bool(false)->float64`, src: false, dst: new(float64), exp: float64(0)},
		{name: `float64(1)->bool`, src: float32(1), dst: new(bool), exp: true},
		{name: `float64(0)->bool`, src: float32(0), dst: new(bool), exp: false},

		// bool <-> string
		{name: `bool(true)->string`, src: true, dst: new(string), exp: "true"},
		{name: `bool(false)->string`, src: false, dst: new(string), exp: "false"},
		{name: `string("true")->bool`, src: "true", dst: new(bool), exp: true},
		{name: `string("false")->bool`, src: "false", dst: new(bool), exp: false},
		{name: `string("foo")->bool`, src: "foo", dst: new(bool), err: true}, // error

		// bool <-> invalid
		{name: `bool->[]byte`, src: true, dst: new([]byte), err: true},             // error
		{name: `bool->[1]bool`, src: true, dst: new([1]bool), err: true},           // error
		{name: `bool->map[int]bool`, src: true, dst: new(map[int]bool), err: true}, // error
		{name: `bool->struct`, src: true, dst: new(struct{}), err: true},           // error

		// int <-> int
		{name: `int(1)->int`, src: 1, dst: new(int), exp: 1},
		{name: `int(259)->int8`, src: 259, dst: new(int8), err: true}, // error
		{name: `int(1)->myInt`, src: 1, dst: new(myInt), exp: myInt(1)},
		{name: `myInt(1)->int`, src: myInt(1), dst: new(int), exp: 1},

		// int <-> uint
		{name: `int(1)->uint`, src: 1, dst: new(uint), exp: uint(1)},
		{name: `uint(1)->int`, src: uint(1), dst: new(int), exp: 1},
		{name: `int(-1)->uint`, src: -1, dst: new(uint), err: true},                                      // error
		{name: `int(259)->uint8`, src: 259, dst: new(uint8), err: true},                                  // error
		{name: `uint(259)->int8`, src: uint(259), dst: new(int8), err: true},                             // error
		{name: `uint64(math.MaxUint64)->int64`, src: uint64(math.MaxUint64), dst: new(int64), err: true}, // error

		// int <-> float
		{name: `int(1)->float64`, src: 1, dst: new(float64), exp: float64(1)},
		{name: `float64(1)->int`, src: float64(1), dst: new(int), exp: 1},
		{name: `float64(math.MathFloat64)->int`, src: float64(math.MaxFloat64), dst: new(int), err: true}, // error
		{name: `float64(257)->int8`, src: float64(257), dst: new(int8), err: true},                        // error

		// int <-> string
		{name: `int(1)->string`, src: 1, dst: new(string), exp: "1"},
		{name: `string("1")->int`, src: "1", dst: new(int), exp: 1},
		{name: `string("1.0")->int`, src: "1.0", dst: new(int), err: true},                                     // error
		{name: `string("foo")->int`, src: "foo", dst: new(int), err: true},                                     // error
		{name: `string("257")->int8`, src: "257", dst: new(int8), err: true},                                   // error
		{name: `string("9223372036854775808")->int64`, src: "9223372036854775808", dst: new(int64), err: true}, // error

		// int <-> slice
		{name: `int->[]byte#positive`, src: math.MaxInt32, dst: new([]byte), exp: []byte{0x0, 0x0, 0x0, 0x0, 0x7f, 0xff, 0xff, 0xff}},
		{name: `int->[]byte#negative`, src: math.MinInt32, dst: new([]byte), exp: []byte{0xff, 0xff, 0xff, 0xff, 0x80, 0x0, 0x0, 0x0}},
		{name: `[]byte->int#positive`, src: []byte{0x0, 0x0, 0x0, 0x0, 0x7f, 0xff, 0xff, 0xff}, dst: new(int), exp: math.MaxInt32},
		{name: `[]byte->int#negative`, src: []byte{0xff, 0xff, 0xff, 0xff, 0x80, 0x0, 0x0, 0x0}, dst: new(int), exp: math.MinInt32},
		{name: `int8->[]byte`, src: int8(math.MaxInt8), dst: new([]byte), exp: []byte{0x7f}},
		{name: `int16->[]byte`, src: int16(math.MaxInt16), dst: new([]byte), exp: []byte{0x7f, 0xff}},
		{name: `int32->[]byte`, src: int32(math.MaxInt32), dst: new([]byte), exp: []byte{0x7f, 0xff, 0xff, 0xff}},
		{name: `int64->[]byte`, src: int64(math.MaxInt64), dst: new([]byte), exp: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `int->[]byte`, src: int(math.MaxInt64), dst: new([]byte), exp: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[]byte->int8`, src: []byte{0x7f}, dst: new(int8), exp: int8(math.MaxInt8)},
		{name: `[]byte->int16`, src: []byte{0x7f, 0xff}, dst: new(int16), exp: int16(math.MaxInt16)},
		{name: `[]byte->int32`, src: []byte{0x7f, 0xff, 0xff, 0xff}, dst: new(int32), exp: int32(math.MaxInt32)},
		{name: `[]byte->int64`, src: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(int64), exp: int64(math.MaxInt64)},
		{name: `[]byte->int`, src: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(int), exp: int(math.MaxInt64)},
		{name: `[]byte->int32#slice-too-short`, src: []byte{0x7f}, dst: new(int32), err: true},                        // error
		{name: `[]byte->int32#slice-too-long`, src: []byte{0x7f, 0x7f, 0x7f, 0x7f, 0x7f}, dst: new(int32), err: true}, // error

		// int <-> array
		{name: `int8->[1]byte`, src: int8(math.MaxInt8), dst: new([1]byte), exp: [1]byte{0x7f}},
		{name: `int16->[2]byte`, src: int16(math.MaxInt16), dst: new([2]byte), exp: [2]byte{0x7f, 0xff}},
		{name: `int32->[4]byte`, src: int32(math.MaxInt32), dst: new([4]byte), exp: [4]byte{0x7f, 0xff, 0xff, 0xff}},
		{name: `int64->[8]byte`, src: int64(math.MaxInt64), dst: new([8]byte), exp: [8]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `int->[8]byte`, src: int(math.MaxInt64), dst: new([8]byte), exp: [8]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[1]byte->int8`, src: [1]byte{0x7f}, dst: new(int8), exp: int8(math.MaxInt8)},
		{name: `[2]byte->int16`, src: [2]byte{0x7f, 0xff}, dst: new(int16), exp: int16(math.MaxInt16)},
		{name: `[4]byte->int32`, src: [4]byte{0x7f, 0xff, 0xff, 0xff}, dst: new(int32), exp: int32(math.MaxInt32)},
		{name: `[8]byte->int64`, src: [8]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(int64), exp: int64(math.MaxInt64)},
		{name: `[8]byte->int`, src: [8]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(int64), exp: int64(math.MaxInt64)},
		{name: `[1]byte->int16#array-too-short`, src: [1]byte{0x7f}, dst: new(int16), err: true},            // error
		{name: `[3]byte->int16#array-too-long`, src: [3]byte{0x7f, 0x7f, 0x7f}, dst: new(int16), err: true}, // error
		{name: `int16->[1]byte#array-too-short`, src: int16(math.MaxInt16), dst: new([1]byte), err: true},   // error
		{name: `int16->[3]byte#array-too-long`, src: int16(math.MaxInt16), dst: new([3]byte), err: true},    // error

		// int <-> invalid
		{name: `int->map[int]int`, src: 1, dst: new(map[int]bool), err: true},
		{name: `int->struct`, src: 1, dst: new(struct{}), err: true},

		// uint <-> uint
		{name: `uint(1)->uint`, src: uint(1), dst: new(uint), exp: uint(1)},
		{name: `uint(259)->uint8`, src: uint(259), dst: new(uint8), err: true}, // error
		{name: `uint(1)->myUint`, src: uint(1), dst: new(myUint), exp: myUint(1)},
		{name: `myUint(1)->uint`, src: myUint(1), dst: new(uint), exp: uint(1)},

		// uint <-> float
		{name: `uint(1)->float64`, src: uint(1), dst: new(float64), exp: float64(1)},
		{name: `float64(1)->uint`, src: float64(1), dst: new(uint), exp: uint(1)},
		{name: `float64(math.MaxFloat64)->uint`, src: float64(math.MaxFloat64), dst: new(uint), err: true}, // error
		{name: `float64(257)->uint8`, src: float64(257), dst: new(uint8), err: true},                       // error

		// uint <-> string
		{name: `uint(1)->string`, src: uint(1), dst: new(string), exp: "1"},
		{name: `string("1")->uint`, src: "1", dst: new(uint), exp: uint(1)},
		{name: `string("1.0")->uint`, src: "1.0", dst: new(uint), err: true},                                       // error
		{name: `string("foo")->uint`, src: "foo", dst: new(uint), err: true},                                       // error
		{name: `string("257")->uint8`, src: "257", dst: new(uint8), err: true},                                     // error
		{name: `string("18446744073709551616")->uint64`, src: "18446744073709551616", dst: new(uint64), err: true}, // error

		// uint <-> slice
		{name: `uint8->[]byte`, src: uint8(math.MaxUint8), dst: new([]byte), exp: []byte{0xff}},
		{name: `uint16->[]byte`, src: uint16(math.MaxUint16), dst: new([]byte), exp: []byte{0xff, 0xff}},
		{name: `uint32->[]byte`, src: uint32(math.MaxUint32), dst: new([]byte), exp: []byte{0xff, 0xff, 0xff, 0xff}},
		{name: `uint64->[]byte`, src: uint64(math.MaxUint64), dst: new([]byte), exp: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `uint->[]byte`, src: uint(math.MaxUint64), dst: new([]byte), exp: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[]byte->uint8`, src: []byte{0xff}, dst: new(uint8), exp: uint8(math.MaxUint8)},
		{name: `[]byte->uint16`, src: []byte{0xff, 0xff}, dst: new(uint16), exp: uint16(math.MaxUint16)},
		{name: `[]byte->uint32`, src: []byte{0xff, 0xff, 0xff, 0xff}, dst: new(uint32), exp: uint32(math.MaxUint32)},
		{name: `[]byte->uint64`, src: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(uint64), exp: uint64(math.MaxUint64)},
		{name: `[]byte->uint`, src: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(uint), exp: uint(math.MaxUint64)},
		{name: `[]byte->uint32#slice-too-short`, src: []byte{0xff, 0xff, 0xff}, dst: new(uint32), err: true},            // error
		{name: `[]byte->uint32#slice-too-long`, src: []byte{0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(uint32), err: true}, // error

		// uuint <-> array
		{name: `uint8->[1]byte`, src: uint8(math.MaxUint8), dst: new([1]byte), exp: [1]byte{0xff}},
		{name: `uint16->[2]byte`, src: uint16(math.MaxUint16), dst: new([2]byte), exp: [2]byte{0xff, 0xff}},
		{name: `uint32->[4]byte`, src: uint32(math.MaxUint32), dst: new([4]byte), exp: [4]byte{0xff, 0xff, 0xff, 0xff}},
		{name: `uint64->[8]byte`, src: uint64(math.MaxUint64), dst: new([8]byte), exp: [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `uint->[8]byte`, src: uint(math.MaxUint64), dst: new([8]byte), exp: [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[1]byte->uint8`, src: [1]byte{0xff}, dst: new(uint8), exp: uint8(math.MaxUint8)},
		{name: `[2]byte->uint16`, src: [2]byte{0xff, 0xff}, dst: new(uint16), exp: uint16(math.MaxUint16)},
		{name: `[4]byte->uint32`, src: [4]byte{0xff, 0xff, 0xff, 0xff}, dst: new(uint32), exp: uint32(math.MaxUint32)},
		{name: `[8]byte->uint64`, src: [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(uint64), exp: uint64(math.MaxUint64)},
		{name: `[8]byte->uint`, src: [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(uint), exp: uint(math.MaxUint64)},
		{name: `[1]byte->uint16#array-too-short`, src: [1]byte{0xff}, dst: new(uint16), err: true},            // error
		{name: `[3]byte->uint16#array-too-long`, src: [3]byte{0xff, 0xff, 0xff}, dst: new(uint16), err: true}, // error
		{name: `uint16->[1]byte#array-too-short`, src: uint16(math.MaxUint16), dst: new([1]byte), err: true},  // error
		{name: `uint16->[3]byte#array-too-long`, src: uint16(math.MaxUint16), dst: new([3]byte), err: true},   // error

		// uint <-> invalid
		{name: `uint->map[int]uint`, src: uint(1), dst: new(map[uint]bool), err: true},
		{name: `uint->struct`, src: uint(1), dst: new(struct{}), err: true},

		// float <-> float
		{name: `float64(1)->float64`, src: 1.0, dst: new(float64), exp: float64(1)},
		{name: `float64(math.MaxFloat64)->float32`, src: float64(math.MaxFloat64), dst: new(float32), err: true}, // error
		{name: `float32(1)->myFloat`, src: float32(1), dst: new(myFloat), exp: myFloat(1)},
		{name: `myFloat(1)->float32`, src: myFloat(1), dst: new(float32), exp: float32(1)},

		// float <-> string
		{name: `float64(1)->string`, src: float64(1), dst: new(string), exp: "1"},
		{name: `string("1")->float64`, src: "1", dst: new(float64), exp: float64(1)},
		{name: `string("1.0")->float64`, src: "1.0", dst: new(float64), exp: float64(1)},
		{name: `string("foo")->float64`, src: "foo", dst: new(float64), err: true},
		{name: `string("1e39")->float32`, src: "1e39", dst: new(float32), err: true},   // error
		{name: `string("1e309")->float64`, src: "1e309", dst: new(float64), err: true}, // error

		// float <-> slice
		{name: `float32(math.MaxFloat32)->[]byte`, src: float32(math.MaxFloat32), dst: new([]byte), exp: []byte{0x7f, 0x7f, 0xff, 0xff}},
		{name: `float64(math.MaxFloat64)->[]byte`, src: float64(math.MaxFloat64), dst: new([]byte), exp: []byte{0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[]byte(...)->float32`, src: []byte{0x7f, 0x7f, 0xff, 0xff}, dst: new(float32), exp: float32(math.MaxFloat32)},
		{name: `[]byte{...}->float64`, src: []byte{0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(float64), exp: float64(math.MaxFloat64)},
		{name: `[]byte{...}->float32#slice-too-short`, src: []byte{0xff}, dst: new(float32), err: true},                        // error
		{name: `[]byte{...}->float32#slice-too-long`, src: []byte{0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(float32), err: true}, // error

		// float <-> array
		{name: `float32(math.MaxFloat32)->[4]byte`, src: float32(math.MaxFloat32), dst: new([4]byte), exp: [4]byte{0x7f, 0x7f, 0xff, 0xff}},
		{name: `float64(math.MaxFloat64)->[8]byte`, src: float64(math.MaxFloat64), dst: new([8]byte), exp: [8]byte{0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: `[4]byte{...}->float32`, src: [4]byte{0x7f, 0x7f, 0xff, 0xff}, dst: new(float32), exp: float32(math.MaxFloat32)},
		{name: `[8]byte{...}->float64`, src: [8]byte{0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(float64), exp: float64(math.MaxFloat64)},
		{name: `[1]byte{...}->float32#array-too-short`, src: [1]byte{0xff}, dst: new(float32), err: true},                        // error
		{name: `[5]byte{...}->float32#array-too-long`, src: [9]byte{0xff, 0xff, 0xff, 0xff, 0xff}, dst: new(float32), err: true}, // error
		{name: `float32->[1]byte#array-too-short`, src: float32(math.MaxFloat32), dst: new([1]byte), err: true},                  // error
		{name: `float32->[5]byte#array-too-long`, src: float32(math.MaxFloat32), dst: new([9]byte), err: true},                   // error

		// float <-> invalid
		{name: `float64->map[int]float64`, src: float64(1), dst: new(map[uint]bool), err: true}, // error
		{name: `float64->struct`, src: float64(1), dst: new(struct{}), err: true},               // error

		// string <-> string
		{name: `string("foo")->string`, src: "foo", dst: new(string), exp: "foo"},
		{name: `string("foo")->myString`, src: "foo", dst: new(myString), exp: myString("foo")},
		{name: `myString("foo")->string`, src: myString("foo"), dst: new(string), exp: "foo"},

		// string <-> slice
		{name: `string("foo")->[]byte`, src: "foo", dst: new([]byte), exp: []byte("foo")},
		{name: `[]byte("foo")->string`, src: []byte("foo"), dst: new(string), exp: "foo"},

		// string <-> array
		{name: `string("foo")->[3]byte`, src: "foo", dst: new([3]byte), exp: [3]byte{'f', 'o', 'o'}},
		{name: `[3]byte("foo")->string`, src: [3]byte{'f', 'o', 'o'}, dst: new(string), exp: "foo"},
		{name: `string("foo")->[2]byte#array-too-short`, src: "foo", dst: new([2]byte), err: true}, // error
		{name: `string("foo")->[4]byte#array-too-long`, src: "foo", dst: new([4]byte), err: true},  // error

		// string <-> invalid
		{name: `string->map[int]string`, src: "foo", dst: new(map[uint]bool), err: true}, // error
		{name: `string->struct`, src: "foo", dst: new(struct{}), err: true},              // error

		// slice <-> slice
		{name: `[]byte("foo")->[]byte`, src: []byte("foo"), dst: new([]byte), exp: []byte("foo")},
		{name: `[]int{1,2,3}->any{0,"0",0.0}`, src: []int{1, 2, 3}, dst: ptr([]any{0, "0", 0.0}), exp: []any{1, "2", 3.0}},
		{name: `[]int{1,2,3}->make([]uint8,0,3)`, src: []int{1, 2, 3}, dst: ptr(make([]uint8, 0, 3)), exp: []uint8{1, 2, 3}},
		{name: `[]int->[]string`, src: []int{1, 2, 3}, dst: new([]string), exp: []string{"1", "2", "3"}},
		{name: `[]string->[]int`, src: []string{"1", "2", "3"}, dst: new([]int), exp: []int{1, 2, 3}},
		{name: `[]string->[]int#invalid`, src: []string{"foo"}, dst: new([]int), err: true}, // error
		{name: `[]int{1}->[]int{0,1}`, src: []int{1}, dst: ptr([]int{0, 1}), exp: []int{1}},
		{name: `[]int{1}->[]any{}`, src: []int{1}, dst: ptr(anySlice()), exp: []any{1}},
		{name: `[]string{"foo"}->mySlice`, src: []string{"foo"}, dst: new(mySlice), exp: mySlice{"foo"}},
		{name: `mySlice{"foo"}->[]string`, src: mySlice{"foo"}, dst: new([]string), exp: []string{"foo"}},

		// slice <-> array
		{name: `[]byte("foo")->[3]byte`, src: []byte("foo"), dst: new([3]byte), exp: [3]byte{'f', 'o', 'o'}},
		{name: `[3]int{1,2,3}->make([]uint8,0,3)`, src: [3]int{1, 2, 3}, dst: ptr(make([]uint8, 0, 3)), exp: []uint8{1, 2, 3}},
		{name: `[3]byte("foo")->[]byte`, src: [3]byte{'f', 'o', 'o'}, dst: new([]byte), exp: []byte("foo")},
		{name: `[]string->[1]int`, src: []string{"1"}, dst: new([1]int), exp: [1]int{1}},
		{name: `[]string->[1]int#invalid`, src: []string{"foo"}, dst: new([1]int), err: true},              // error
		{name: `[1]string->[]int#invalid`, src: [1]string{"foo"}, dst: new([]int), err: true},              // error
		{name: `[]byte("foo")->[2]byte#array-too-short`, src: []byte("foo"), dst: new([2]byte), err: true}, // error
		{name: `[]byte("foo")->[4]byte#array-too-long`, src: []byte("foo"), dst: new([4]byte), err: true},  // error

		// slice <-> invalid
		{name: `[]byte->map[int][]byte`, src: []byte("foo"), dst: new(map[uint]bool), err: true}, // error
		{name: `[]byte->struct`, src: []byte("foo"), dst: new(struct{}), err: true},              // error

		// array <-> array
		{name: `[1]byte{1}->[1]byte`, src: [1]byte{1}, dst: new([1]byte), exp: [1]byte{1}},
		{name: `[1]string{1}->[1]int`, src: [1]string{"1"}, dst: new([1]int), exp: [1]int{1}},
		{name: `[1]string{1}->[1]int#invalid`, src: [1]string{"foo"}, dst: new([1]int), err: true},   // error
		{name: `[1]byte{1}->[2]byte#array-too-long`, src: [1]byte{1}, dst: new([2]byte), err: true},  // error
		{name: `[2]byte{1}->[1]byte#array-too-short`, src: [2]byte{1}, dst: new([1]byte), err: true}, // error
		{name: `[1]string{"foo"}->myArray`, src: [1]string{"foo"}, dst: new(myArray), exp: myArray{"foo"}},
		{name: `myArray{"foo"}->[1]string`, src: myArray{"foo"}, dst: new([1]string), exp: [1]string{"foo"}},

		// array <-> invalid
		{name: `[1]byte->map[int][1]byte`, src: [1]byte{1}, dst: new(map[uint]bool), err: true}, // error
		{name: `[1]byte->struct`, src: [1]byte{1}, dst: new(struct{}), err: true},               // error

		// map <-> map
		{name: `map[int]string{1:"foo"}->map[int]string`, src: map[int]string{1: "foo"}, dst: new(map[int]string), exp: map[int]string{1: "foo"}},
		{name: `map[int]string{1:"1"}->map[string]int`, src: map[int]string{1: "1"}, dst: new(map[string]int), exp: map[string]int{"1": 1}},
		{name: `map[int]string{1:"foo"}->map[string]int#invalid`, src: map[int]string{1: "foo"}, dst: new(map[string]int), err: true}, // error
		{name: `map[string]int{"foo":1}->map[int]string`, src: map[string]int{"foo": 1}, dst: new(map[int]string), err: true},         // error
		{name: `map[string]int{"foo":1}->map[int]string#invalid`, src: map[string]int{"foo": 1}, dst: new(map[int]string), err: true}, // error
		{name: `map[string]string{"foo":"bar"}->myMap`, src: map[string]string{"foo": "bar"}, dst: new(myMap), exp: myMap{"foo": "bar"}},
		{name: `myMap{"foo":"bar"}->map[string]string`, src: myMap{"foo": "bar"}, dst: new(map[string]string), exp: map[string]string{"foo": "bar"}},

		// map <-> struct
		{name: `map[string]string{"Foo":"bar"}->struct{Foo string}`, src: map[string]string{"Foo": "bar"}, dst: new(struct{ Foo string }), exp: struct{ Foo string }{"bar"}},
		{name: `struct{Foo string}{Foo:"bar"}->map[string]string`, src: struct{ Foo string }{"bar"}, dst: new(map[string]string), exp: map[string]string{"Foo": "bar"}},

		// struct <-> struct
		{name: `struct{A int}{1}->struct{A int}`, src: struct{ A int }{1}, dst: new(struct{ A int }), exp: struct{ A int }{1}},
		{name: `struct{Foo string}{Foo:"bar"}->struct{Foo string}`, src: struct{ Foo string }{"bar"}, dst: new(struct{ Foo string }), exp: struct{ Foo string }{"bar"}},
		{name: `struct{Foo string}{Foo:"bar"}->struct{Foo int}`, src: struct{ Foo string }{"bar"}, dst: new(struct{ Foo int }), err: true},         // error
		{name: `struct{Foo string}{Foo:"bar"}->struct{Foo int}#invalid`, src: struct{ Foo string }{"bar"}, dst: new(struct{ Foo int }), err: true}, // error

		// nil values
		{name: `bool(true)->(*bool)(nil)`, src: true, dst: new(*bool), exp: ptr(ptr(true))},
		{name: `int(1)->(*int)(nil)`, src: 1, dst: new(*int), exp: ptr(ptr(1))},
		{name: `uint(1)->(*uint)(nil)`, src: uint(1), dst: new(*uint), exp: ptr(ptr(uint(1)))},
		{name: `float64(1)->(*float64)(nil)`, src: float64(1), dst: new(*float64), exp: ptr(ptr(float64(1)))},
		{name: `string("foo")->(*string)(nil)`, src: "foo", dst: new(*string), exp: ptr(ptr("foo"))},
		{name: `[]byte("foo")->(*[]byte)(nil)`, src: []byte("foo"), dst: new(*[]byte), exp: ptr(ptr([]byte("foo")))},
		{name: `[3]byte("foo")->(*[3]byte)(nil)`, src: [3]byte{'f', 'o', 'o'}, dst: new(*[3]byte), exp: ptr(ptr([3]byte{'f', 'o', 'o'}))},
		{name: `map[int]string{1:"foo"}->(*map[int]string)(nil)`, src: map[int]string{1: "foo"}, dst: new(*map[int]string), exp: ptr(ptr(map[int]string{1: "foo"}))},
		{name: `struct{Foo string}{Foo:"bar"}->(*struct{Foo string})(nil)`, src: struct{ Foo string }{"bar"}, dst: new(*struct{ Foo string }), exp: ptr(ptr(struct{ Foo string }{"bar"}))},

		{name: `(*bool)(nil)->bool`, src: new(*bool), dst: new(bool), err: true},                                         // error
		{name: `(*int)(nil)->int`, src: new(*int), dst: new(int), err: true},                                             // error
		{name: `(*uint)(nil)->uint`, src: new(*uint), dst: new(uint), err: true},                                         // error
		{name: `(*float64)(nil)->float64`, src: new(*float64), dst: new(float64), err: true},                             // error
		{name: `(*string)(nil)->string`, src: new(*string), dst: new(string), err: true},                                 // error
		{name: `(*[]byte)(nil)->[]byte`, src: new(*[]byte), dst: new([]byte), err: true},                                 // error
		{name: `(*[1]byte)(nil)->[1]byte`, src: new(*[1]byte), dst: new([1]byte), err: true},                             // error
		{name: `(*map[int]string)(nil)->map[int]string`, src: new(*map[int]string), dst: new(map[int]string), err: true}, // error
		{name: `(*struct{})(nil)->struct{}`, src: new(*struct{}), dst: new(struct{}), err: true},                         // error
		{name: `nil->nil`, src: nil, dst: nil, err: true},                                                                // error
		{name: `nil->[]byte`, src: nil, dst: new([]byte), err: true},                                                     // error
		{name: `[]byte->nil`, src: []byte("foo"), dst: nil, err: true},                                                   // error

		// unaddressable values
		{name: `bool->bool#unaddressable`, src: true, dst: true, err: true},                                              // error
		{name: `int->int#unaddressable`, src: 1, dst: 1, err: true},                                                      // error
		{name: `uint->uint#unaddressable`, src: uint(1), dst: uint(1), err: true},                                        // error
		{name: `float64->float64#unaddressable`, src: float64(1), dst: float64(1), err: true},                            // error
		{name: `string->string#unaddressable`, src: "foo", dst: "foo", err: true},                                        // error
		{name: `[]byte->[]byte#unaddressable`, src: []byte("foo"), dst: []byte{}, err: true},                             // error
		{name: `[3]byte->[3]byte#unaddressable`, src: [3]byte{'f', 'o', 'o'}, dst: [3]byte{}, err: true},                 // error
		{name: `struct->struct#unaddressable`, src: struct{ Foo string }{"bar"}, dst: struct{ Foo string }{}, err: true}, // error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Map(tt.src, tt.dst)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, exp(tt.exp), dst(tt.dst))
			}
		})
	}
}

func TestStrictTypes(t *testing.T) {
	type (
		myBool   bool
		myInt    int
		myUint   uint
		myFloat  float64
		myString string
		mySlice  []string
		myArray  [1]string
	)

	tests := []struct {
		name string
		src  any
		dst  any
		exp  any
		err  bool
	}{
		{name: `bool->bool`, src: true, dst: new(bool), exp: true},
		{name: `bool->int`, src: true, dst: new(int), err: true},            // error
		{name: `bool->uint`, src: true, dst: new(uint), err: true},          // error
		{name: `bool->float64`, src: true, dst: new(float64), err: true},    // error
		{name: `bool->string`, src: true, dst: new(string), err: true},      // error
		{name: `bool->[]byte`, src: true, dst: new([]byte), err: true},      // error
		{name: `bool->[1]byte`, src: true, dst: new([1]byte), err: true},    // error
		{name: `bool->map`, src: true, dst: new(map[int]string), err: true}, // error
		{name: `bool->struct`, src: true, dst: new(struct{}), err: true},    // error
		{name: `bool-myBool`, src: true, dst: new(myBool), err: true},       // error
		{name: `int->bool`, src: 1, dst: new(bool), err: true},              // error
		{name: `int->int`, src: 1, dst: new(int), exp: 1},
		{name: `int->int8`, src: 1, dst: new(int8), err: true},          // error
		{name: `int->uint`, src: 1, dst: new(uint), err: true},          // error
		{name: `int->float64`, src: 1, dst: new(float64), err: true},    // error
		{name: `int->string`, src: 1, dst: new(string), err: true},      // error
		{name: `int->[]byte`, src: 1, dst: new([]byte), err: true},      // error
		{name: `int->[1]byte`, src: 1, dst: new([1]byte), err: true},    // error
		{name: `int->map`, src: 1, dst: new(map[int]string), err: true}, // error
		{name: `int->struct`, src: 1, dst: new(struct{}), err: true},    // error
		{name: `int-myInt`, src: 1, dst: new(myInt), err: true},         // error
		{name: `uint->bool`, src: uint(1), dst: new(bool), err: true},   // error
		{name: `uint->int`, src: uint(1), dst: new(int), err: true},     // error
		{name: `uint->uint`, src: uint(1), dst: new(uint), exp: uint(1)},
		{name: `uint->uint8`, src: uint(1), dst: new(uint8), err: true},        // error
		{name: `uint->float64`, src: uint(1), dst: new(float64), err: true},    // error
		{name: `uint->string`, src: uint(1), dst: new(string), err: true},      // error
		{name: `uint->[]byte`, src: uint(1), dst: new([]byte), err: true},      // error
		{name: `uint->[1]byte`, src: uint(1), dst: new([1]byte), err: true},    // error
		{name: `uint->map`, src: uint(1), dst: new(map[int]string), err: true}, // error
		{name: `uint->struct`, src: uint(1), dst: new(struct{}), err: true},    // error
		{name: `uint-myUint`, src: uint(1), dst: new(myUint), err: true},       // error
		{name: `float64->bool`, src: float64(1), dst: new(bool), err: true},    // error
		{name: `float64->int`, src: float64(1), dst: new(int), err: true},      // error
		{name: `float64->uint`, src: float64(1), dst: new(uint), err: true},    // error
		{name: `float64->float64`, src: float64(1), dst: new(float64), exp: float64(1)},
		{name: `float64->float32`, src: float64(1), dst: new(float32), err: true},    // error
		{name: `float64->string`, src: float64(1), dst: new(string), err: true},      // error
		{name: `float64->[]byte`, src: float64(1), dst: new([]byte), err: true},      // error
		{name: `float64->[1]byte`, src: float64(1), dst: new([1]byte), err: true},    // error
		{name: `float64->map`, src: float64(1), dst: new(map[int]string), err: true}, // error
		{name: `float64->struct`, src: float64(1), dst: new(struct{}), err: true},    // error
		{name: `float64-myFloat`, src: float64(1), dst: new(myFloat), err: true},     // error
		{name: `string->bool`, src: "1", dst: new(bool), err: true},                  // error
		{name: `string->int`, src: "1", dst: new(int), err: true},                    // error
		{name: `string->uint`, src: "1", dst: new(uint), err: true},                  // error
		{name: `string->float64`, src: "1", dst: new(float64), err: true},            // error
		{name: `string->string`, src: "1", dst: new(string), exp: "1"},
		{name: `string->[]byte`, src: "1", dst: new([]byte), err: true},           // error
		{name: `string->[1]byte`, src: "1", dst: new([1]byte), err: true},         // error
		{name: `string->map`, src: "1", dst: new(map[int]string), err: true},      // error
		{name: `string->struct`, src: "1", dst: new(struct{}), err: true},         // error
		{name: `string-myString`, src: "1", dst: new(myString), err: true},        // error
		{name: `[]byte->bool`, src: []byte("1"), dst: new(bool), err: true},       // error
		{name: `[]byte->int`, src: []byte("1"), dst: new(int), err: true},         // error
		{name: `[]byte->uint`, src: []byte("1"), dst: new(uint), err: true},       // error
		{name: `[]byte->float64`, src: []byte("1"), dst: new(float64), err: true}, // error
		{name: `[]byte->string`, src: []byte("1"), dst: new(string), err: true},   // error
		{name: `[]byte->[]byte`, src: []byte("1"), dst: new([]byte), exp: []byte("1")},
		{name: `[]byte->[]int`, src: []byte("1"), dst: new([]int), err: true},        // error
		{name: `[]byte->[1]byte`, src: []byte("1"), dst: new([1]byte), err: true},    // error
		{name: `[]byte->map`, src: []byte("1"), dst: new(map[int]string), err: true}, // error
		{name: `[]byte->struct`, src: []byte("1"), dst: new(struct{}), err: true},    // error
		{name: `[]byte-myBytes`, src: []byte("1"), dst: new(mySlice), err: true},     // error
		{name: `[1]byte->bool`, src: [1]byte{'1'}, dst: new(bool), err: true},        // error
		{name: `[1]byte->int`, src: [1]byte{'1'}, dst: new(int), err: true},          // error
		{name: `[1]byte->uint`, src: [1]byte{'1'}, dst: new(uint), err: true},        // error
		{name: `[1]byte->float64`, src: [1]byte{'1'}, dst: new(float64), err: true},  // error
		{name: `[1]byte->string`, src: [1]byte{'1'}, dst: new(string), err: true},    // error
		{name: `[1]byte->[]byte`, src: [1]byte{'1'}, dst: new([]byte), err: true},    // error
		{name: `[1]byte->[1]byte`, src: [1]byte{'1'}, dst: new([1]byte), exp: [1]byte{'1'}},
		{name: `[1]byte->[1]int`, src: [1]byte{'1'}, dst: new([1]int), err: true},            // error
		{name: `[1]byte->map`, src: [1]byte{'1'}, dst: new(map[int]string), err: true},       // error
		{name: `[1]byte->struct`, src: [1]byte{'1'}, dst: new(struct{}), err: true},          // error
		{name: `[1]byte-myBytes`, src: [1]byte{'1'}, dst: new(myArray), err: true},           // error
		{name: `map->map`, src: map[int]string{1: "1"}, dst: new(map[string]int), err: true}, // error
		{name: `map->map#same`, src: map[int]int{1: 1}, dst: new(map[int]int), exp: map[int]int{1: 1}},
		{name: `map->struct`, src: map[string]int{"A": 1}, dst: new(struct{ A int }), exp: struct{ A int }{1}},
		{name: `struct->map`, src: struct{ A int }{1}, dst: new(map[string]int), exp: map[string]int{"A": 1}},
		{name: `struct->struct`, src: struct{ A int }{1}, dst: new(struct{ A int }), exp: struct{ A int }{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapContext(Default.Context.WithStrictTypes(true), tt.src, tt.dst)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, exp(tt.exp), dst(tt.dst))
			}
		})
	}
}

func TestTags(t *testing.T) {
	t.Run("struct-map", func(t *testing.T) {
		type Src struct {
			Foo int    `map:"foo"`
			Bar string `map:"bar"`
			Baz int    `map:"-"`
			qaz int
		}
		var dst map[string]any
		err := Map(Src{
			Foo: 1,
			Bar: "2",
			Baz: 3,
			qaz: 4,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"foo": 1,
			"bar": "2",
		}, dst)
	})
	t.Run("map-struct", func(t *testing.T) {
		type Dst struct {
			Foo int    `map:"foo"`
			Bar string `map:"bar"`
			Baz int    `map:"-"`
			qaz int
		}
		var dst Dst
		err := Map(map[string]any{
			"foo": 1,
			"bar": 2,
			"baz": 3,
			"qaz": 4,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			Foo: 1,
			Bar: "2",
		}, dst)
	})
	t.Run("struct-struct", func(t *testing.T) {
		type Src struct {
			Foo int    `map:"foo"`
			Bar string `map:"bar"`
			Baz int    `map:"-"`
			qaz int
		}
		type Dst struct {
			A int `map:"foo"`
			B int `map:"bar"`
			C int `map:"baz"`
		}
		var dst Dst
		err := Map(Src{
			Foo: 1,
			Bar: "2",
			Baz: 3,
			qaz: 4,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			A: 1,
			B: 2,
		}, dst)
	})
	t.Run("struct-struct#tag-src", func(t *testing.T) {
		type Str struct {
			Foo int    `map:"A"`
			Bar string `map:"B"`
			Baz []int  `map:"C"`
		}
		type Dst struct {
			A int
			B string
			C []int
		}
		var dst Dst
		err := Map(Str{
			Foo: 1,
			Bar: "2",
			Baz: []int{3, 4, 5},
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			A: 1,
			B: "2",
			C: []int{3, 4, 5},
		}, dst)
	})
	t.Run("struct-struct#tag-dst", func(t *testing.T) {
		type Str struct {
			Foo int
			Bar string
			Baz []int
		}
		type Dst struct {
			A int    `map:"Foo"`
			B string `map:"Bar"`
			C []int  `map:"Baz"`
		}
		var dst Dst
		err := Map(Str{
			Foo: 1,
			Bar: "2",
			Baz: []int{3, 4, 5},
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			A: 1,
			B: "2",
			C: []int{3, 4, 5},
		}, dst)
	})
	t.Run("struct-struct#same", func(t *testing.T) {
		type Str struct {
			Foo int `map:"foo"`
			Bar int `map:"-"`
			baz int
		}
		var dst Str
		err := Map(Str{
			Foo: 1,
			Bar: 2,
			baz: 3,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Str{
			Foo: 1,
		}, dst)
	})
}

func TestMapToStruct(t *testing.T) {
	type Str struct {
		Foo int
		Bar *big.Int
		Baz any
	}
	dst := Str{
		Baz: new(big.Int),
	}
	err := Map(map[string]any{
		"Foo": 1,
		"Bar": 2,
		"Baz": 3,
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, Str{
		Foo: 1,
		Bar: big.NewInt(2),
		Baz: big.NewInt(3),
	}, dst)
}

func TestMapToMap(t *testing.T) {
	dst := map[string]any{
		"foo": nil,
		"bar": new(big.Int),
	}
	err := Map(map[string]any{
		"foo": 1,
		"bar": 2,
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"foo": 1,
		"bar": big.NewInt(2),
	}, dst)
}

func TestStructToMap(t *testing.T) {
	type Str struct {
		Foo int
		Bar *big.Int
		Baz any
	}
	dst := map[string]any{
		"Foo": nil,
		"Bar": new(big.Int),
		"Baz": new(big.Int),
	}
	err := Map(Str{
		Foo: 1,
		Bar: big.NewInt(2),
		Baz: big.NewInt(3),
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"Foo": 1,
		"Bar": big.NewInt(2),
		"Baz": big.NewInt(3),
	}, dst)
}
