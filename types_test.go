package anymapper

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	tm1 := time.Unix(1666666666, 0).UTC()
	tm2 := time.Unix(1666666666, int64(time.Millisecond*500)).UTC()
	tm3 := time.Unix(math.MaxInt64, 0).UTC()

	tests := []struct {
		name string
		src  any
		dst  any
		exp  any
		err  bool
	}{
		// time.Time <-> time.Time
		{name: "time.Time-time.Time", src: tm1, dst: new(time.Time), exp: tm1},

		// time.Time <-> int
		{name: "time.Time-int64", src: tm1, dst: new(int64), exp: tm1.Unix()},
		{name: "int64-time.Time", src: tm1.Unix(), dst: new(time.Time), exp: tm1},
		{name: "time.Time-int32", src: tm3, dst: new(int32), err: true},
		{name: "time.Time-int8", src: tm3, dst: new(int8), err: true},
		{name: "time.Time-int16", src: tm3, dst: new(int16), err: true},
		{name: "int8-time.Time", src: int8(1), dst: new(time.Time), err: true},
		{name: "int16-time.Time", src: int16(1), dst: new(time.Time), err: true},

		// time.Time <-> uint
		{name: "time.Time-uint64", src: tm1, dst: new(uint64), exp: uint64(tm1.Unix())},
		{name: "uint64-time.Time", src: uint64(tm1.Unix()), dst: new(time.Time), exp: tm1},
		{name: "time.Time-uint32", src: tm3, dst: new(uint32), err: true},

		// time.Time <-> float64
		{name: "time.Time-float64", src: tm2, dst: new(float64), exp: float64(tm2.UnixNano()) / float64(time.Second)},
		{name: "float64-time.Time", src: float64(tm2.UnixNano()) / float64(time.Second), dst: new(time.Time), exp: tm2},

		// time.Time <-> string
		{name: "time.Time-string", src: tm1, dst: new(string), exp: tm1.Format(time.RFC3339)},
		{name: "string-time.Time", src: tm1.Format(time.RFC3339), dst: new(time.Time), exp: tm1},
		{name: "string-time.Time#invalid", src: "foo", dst: new(time.Time), err: true},

		// time.Time <-> slice
		{name: "time.Time-[]byte", src: tm1, dst: new([]byte), exp: []byte{0x0, 0x0, 0x0, 0x0, 0x63, 0x57, 0x50, 0xaa}},
		{name: "[]byte-time.Time", src: []byte{0x0, 0x0, 0x0, 0x0, 0x63, 0x57, 0x50, 0xaa}, dst: new(time.Time), exp: tm1},

		// time.Time <-> big.Int
		{name: "time.Time-big.Int", src: tm1, dst: new(big.Int), exp: big.NewInt(tm1.Unix())},
		{name: "big.Int-time.Time", src: big.NewInt(tm1.Unix()), dst: new(time.Time), exp: tm1},

		// time.Time <-> big.Float
		{name: "time.Time-big.Float", src: tm2, dst: new(big.Float), exp: big.NewFloat(float64(tm2.UnixNano()) / float64(time.Second))},
		{name: "big.Float-time.Time", src: big.NewFloat(float64(tm2.UnixNano()) / float64(time.Second)), dst: new(time.Time), exp: tm2},

		// time.Time <-> invalid
		{name: "time.Time-bool", src: tm1, dst: new(bool), err: true},
		{name: "bool-time.Time", src: true, dst: new(time.Time), err: true},
		{name: "time.Time-slice", src: tm1, dst: new([]int), err: true},
		{name: "slice-time.Time", src: []int{1, 2, 3}, dst: new(time.Time), err: true},
		{name: "time.Time-map", src: tm1, dst: new(map[string]int), err: true},
		{name: "map-time.Time", src: map[string]int{"a": 1, "b": 2}, dst: new(time.Time), err: true},
		{name: "time.Time-struct", src: tm1, dst: new(struct{}), err: true},
		{name: "struct-time.Time", src: struct{}{}, dst: new(time.Time), err: true},

		// big.Int <-> big.Int
		{name: "big.Int-big.Int", src: big.NewInt(1), dst: new(big.Int), exp: big.NewInt(1)},

		// big.Int <-> bool
		{name: "big.Int-bool#true", src: big.NewInt(1), dst: new(bool), exp: true},
		{name: "big.Int-bool#false", src: big.NewInt(0), dst: new(bool), exp: false},
		{name: "bool-big.Int#true", src: true, dst: new(big.Int), exp: big.NewInt(1)},
		{name: "bool-big.Int#false", src: false, dst: new(big.Int), exp: big.NewInt(0)},

		// big.Int <-> int
		{name: "big.Int-int64", src: big.NewInt(1), dst: new(int64), exp: int64(1)},
		{name: "int-big.Int64", src: int64(1), dst: new(big.Int), exp: big.NewInt(1)},
		{name: "big.Int-int8", src: big.NewInt(128), dst: new(int8), err: true},
		{name: "int8-big.Int", src: int8(1), dst: new(big.Int), exp: big.NewInt(1)},

		// big.Int <-> uint
		{name: "big.Int-uint64", src: big.NewInt(1), dst: new(uint64), exp: uint64(1)},
		{name: "int-big.uInt64", src: uint64(1), dst: new(big.Int), exp: big.NewInt(1)},
		{name: "big.Int-int8", src: big.NewInt(259), dst: new(uint8), err: true},
		{name: "int8-big.Int", src: uint8(1), dst: new(big.Int), exp: big.NewInt(1)},

		// big.Int <-> float64
		{name: "big.Int-float64", src: big.NewInt(2), dst: new(float64), exp: float64(2)},
		{name: "float64-big.Int", src: float64(2), dst: new(big.Int), exp: big.NewInt(2)},
		{name: "big.Int-float64#overflow", src: new(big.Int).Lsh(big.NewInt(1), 1024), dst: new(float64), err: true},

		// big.Int <-> string
		{name: "big.Int-string", src: big.NewInt(2), dst: new(string), exp: "2"},
		{name: "string-big.Int", src: "2", dst: new(big.Int), exp: big.NewInt(2)},
		{name: "string-big.Int#invalid", src: "foo", dst: new(big.Int), err: true},

		// big.Int <-> slice
		{name: "big.Int-[]byte#positive", src: big.NewInt(2), dst: new([]byte), exp: []byte{0x2}},
		{name: "big.Int-[]byte#negative", src: big.NewInt(-2), dst: new([]byte), exp: []byte{0x2}},
		{name: "[]byte-big.Int", src: []byte{0x2}, dst: new(big.Int), exp: big.NewInt(2)},

		// big.Int <-> big.Float
		{name: "big.Int-big.Float", src: big.NewInt(2), dst: new(big.Float), exp: big.NewFloat(2)},
		{name: "big.Float-big.Int", src: big.NewFloat(math.E), dst: new(big.Int), exp: big.NewInt(2)},

		// big.Int <-> big.Rat
		{name: "big.Int-big.Rat", src: big.NewInt(2), dst: new(big.Rat), exp: big.NewRat(2, 1)},
		{name: "big.Rat-big.Int", src: big.NewRat(2, 1), dst: new(big.Int), exp: big.NewInt(2)},

		// big.Int <-> invalid
		{name: "big.Int-map", src: big.NewInt(1), dst: new(map[string]int), err: true},
		{name: "map-big.Int", src: map[string]int{"a": 1, "b": 2}, dst: new(big.Int), err: true},
		{name: "big.Int-struct", src: big.NewInt(1), dst: new(struct{}), err: true},
		{name: "struct-big.Int", src: struct{}{}, dst: new(big.Int), err: true},

		// big.Float <-> big.Float
		{name: "big.Float-big.Float", src: big.NewFloat(math.E), dst: new(big.Float), exp: big.NewFloat(math.E)},

		// big.Float <-> bool
		{name: "big.Float-bool#true", src: big.NewFloat(1), dst: new(bool), exp: true},
		{name: "big.Float-bool#false", src: big.NewFloat(0), dst: new(bool), exp: false},
		{name: "bool-big.Float#true", src: true, dst: new(big.Float), exp: big.NewFloat(1)},
		{name: "bool-big.Float#false", src: false, dst: new(big.Float), exp: big.NewFloat(0)},

		// big.Float <-> int
		{name: "big.Float-int64", src: big.NewFloat(math.E), dst: new(int64), exp: int64(2)},
		{name: "int-big.Float64", src: int64(2), dst: new(big.Float), exp: big.NewFloat(2)},
		{name: "big.Float-int8", src: big.NewFloat(128), dst: new(int8), err: true},
		{name: "int8-big.Float", src: int8(1), dst: new(big.Float), exp: big.NewFloat(1)},

		// big.Float <-> uint
		{name: "big.Float-uint64", src: big.NewFloat(math.E), dst: new(uint64), exp: uint64(2)},
		{name: "int-big.uFloat64", src: uint64(2), dst: new(big.Float), exp: big.NewFloat(2)},
		{name: "big.Float-int8", src: big.NewFloat(259), dst: new(uint8), err: true},
		{name: "int8-big.Float", src: uint8(1), dst: new(big.Float), exp: big.NewFloat(1)},

		// big.Float <-> float64
		{name: "big.Float-float64", src: big.NewFloat(math.E), dst: new(float64), exp: math.E},
		{name: "float64-big.Float", src: math.E, dst: new(big.Float), exp: big.NewFloat(math.E)},
		{name: "big.Float-float64#overflow", src: new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 1024)), dst: new(float64), err: true},

		// big.Float <-> string
		{name: "big.Float-string", src: big.NewFloat(1.5), dst: new(string), exp: "1.5"},
		{name: "string-big.Float", src: "1.5", dst: new(big.Float), exp: big.NewFloat(1.5)},
		{name: "string-big.Float#invalid", src: "foo", dst: new(big.Float), err: true},

		// big.Float <-> big.Rat
		{name: "big.Float-big.Rat", src: big.NewFloat(0.5), dst: new(big.Rat), exp: big.NewRat(1, 2)},
		{name: "big.Rat-big.Float", src: big.NewRat(1, 2), dst: new(big.Float), exp: big.NewFloat(0.5)},

		// big.Float <-> invalid
		{name: "big.Float-map", src: big.NewFloat(1), dst: new(map[string]int), err: true}, {name: "big.Float-chan", src: big.NewFloat(1), dst: new(chan int), err: true},
		{name: "map-big.Float", src: map[string]int{"foo": 1}, dst: new(big.Float), err: true},
		{name: "big.Float-struct", src: big.NewFloat(1), dst: new(struct{}), err: true},
		{name: "struct-big.Float", src: struct{}{}, dst: new(big.Float), err: true},

		// big.Rat <-> big.Rat
		{name: "big.Rat-big.Rat", src: big.NewRat(1, 2), dst: new(big.Rat), exp: big.NewRat(1, 2)},

		// big.Rat <-> bool
		{name: "big.Rat-bool#true", src: big.NewRat(1, 1), dst: new(bool), exp: true},
		{name: "big.Rat-bool#false", src: big.NewRat(0, 1), dst: new(bool), exp: false},
		{name: "bool-big.Rat#true", src: true, dst: new(big.Rat), exp: big.NewRat(1, 1)},
		{name: "bool-big.Rat#false", src: false, dst: new(big.Rat), exp: big.NewRat(0, 1)},

		// big.Rat <-> int
		{name: "big.Rat-int64", src: big.NewRat(2, 1), dst: new(int64), exp: int64(2)},
		{name: "int-big.Rat64", src: int64(2), dst: new(big.Rat), exp: big.NewRat(2, 1)},

		// big.Rat <-> uint
		{name: "big.Rat-uint64", src: big.NewRat(2, 1), dst: new(uint64), exp: uint64(2)},
		{name: "int-big.uRat64", src: uint64(2), dst: new(big.Rat), exp: big.NewRat(2, 1)},

		// big.Rat <-> float64
		{name: "big.Rat-float64", src: big.NewRat(2, 1), dst: new(float64), exp: float64(2)},
		{name: "float64-big.Rat", src: float64(2), dst: new(big.Rat), exp: big.NewRat(2, 1)},

		// big.Rat <-> string
		{name: "big.Rat-string", src: big.NewRat(1, 2), dst: new(string), exp: "1/2"},
		{name: "string-big.Rat", src: "1/2", dst: new(big.Rat), exp: big.NewRat(1, 2)},

		// big.Rat <-> slice
		{name: "big.Rat-[]byte", src: big.NewRat(1, 2), dst: new([]byte), exp: []byte{1, 2}},
		{name: "[]byte-big.Rat", src: []byte{1, 2}, dst: new(big.Rat), exp: big.NewRat(1, 2)},
		{name: "big.Rat-[]int8#invalid1", src: big.NewRat(128, 1), dst: new([]int8), err: true},
		{name: "big.Rat-[]int8#invalid2", src: big.NewRat(1, 128), dst: new([]int8), err: true},

		// big.Rat <-> array
		{name: "big.Rat-[2]byte", src: big.NewRat(1, 2), dst: new([2]byte), exp: [2]byte{1, 2}},
		{name: "[2]byte-big.Rat", src: [2]byte{1, 2}, dst: new(big.Rat), exp: big.NewRat(1, 2)},
		{name: "big.Rat-[1]byte#array-too-short", src: big.NewRat(1, 2), dst: new([1]byte), err: true},
		{name: "[1]byte-big.Rat#array-too-short", src: [1]byte{1}, dst: new(big.Rat), err: true},

		// big.Rat <-> invalid
		{name: "big.Rat-map", src: big.NewRat(1, 2), dst: new(map[string]string), err: true},
		{name: "map-big.Rat", src: map[string]string{"foo": "bar"}, dst: new(big.Rat), err: true},
		{name: "big.Rat-struct", src: big.NewRat(1, 2), dst: new(struct{}), err: true},
		{name: "struct-big.Rat", src: struct{}{}, dst: new(big.Rat), err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Map(tt.src, tt.dst)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				switch tdst := tt.dst.(type) {
				case *big.Float:
					assert.Equal(t, tt.exp.(*big.Float).Text('f', -1), tdst.Text('f', -1))
				default:
					assert.Equal(t, exp(tt.exp), dst(tt.dst))
				}
			}
		})
	}
}
