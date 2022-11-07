package anymapper

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapBasicTypes(t *testing.T) {
	// bool
	t.Run("bool-bool#true", func(t *testing.T) {
		var dst bool
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("bool-bool#false", func(t *testing.T) {
		var dst bool
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, false, dst)
	})
	t.Run("bool-int#true", func(t *testing.T) {
		var dst int
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("bool-int#false", func(t *testing.T) {
		var dst int
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, dst)
	})
	t.Run("bool-uint#true", func(t *testing.T) {
		var dst uint
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("bool-uint#false", func(t *testing.T) {
		var dst uint
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), dst)
	})
	t.Run("bool-float#true", func(t *testing.T) {
		var dst float64
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), dst)
	})
	t.Run("bool-float#false", func(t *testing.T) {
		var dst float64
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(0), dst)
	})
	t.Run("bool-string#false", func(t *testing.T) {
		var dst string
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "false", dst)
	})
	t.Run("bool-string#true", func(t *testing.T) {
		var dst string
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "true", dst)
	})
	t.Run("bool-slice", func(t *testing.T) {
		var dst []bool
		err := Map(true, &dst)
		assert.Error(t, err)
	})
	t.Run("bool-array", func(t *testing.T) {
		var dst [1]bool
		err := Map(true, &dst)
		assert.Error(t, err)
	})
	t.Run("bool-map", func(t *testing.T) {
		var dst map[string]bool
		err := Map(true, &dst)
		assert.Error(t, err)
	})
	t.Run("bool-struct", func(t *testing.T) {
		var dst struct{ Bool bool }
		err := Map(true, &dst)
		assert.Error(t, err)
	})

	// int
	t.Run("int-bool", func(t *testing.T) {
		var dst bool
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("int-int", func(t *testing.T) {
		var dst int
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("int-uint", func(t *testing.T) {
		var dst uint
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("int-float", func(t *testing.T) {
		var dst float64
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), dst)
	})
	t.Run("int-string", func(t *testing.T) {
		var dst string
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "1", dst)
	})
	t.Run("int-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(math.MaxInt32, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x7f, 0xff, 0xff, 0xff}, dst)
	})
	t.Run("int-[4]byte", func(t *testing.T) {
		var dst [4]byte
		err := Map(int64(math.MaxInt32), &dst)
		assert.NoError(t, err)
		assert.Equal(t, [4]byte{0x7f, 0xff, 0xff, 0xff}, dst)
	})
	t.Run("int-[4]byte#overflow", func(t *testing.T) {
		var dst [4]byte
		err := Map(int64(math.MaxUint32)+1, &dst)
		assert.Error(t, err)
	})
	t.Run("int-struct", func(t *testing.T) {
		var dst struct{ Int int }
		err := Map(1, &dst)
		assert.Error(t, err)
	})

	// uint
	t.Run("uint-bool", func(t *testing.T) {
		var dst bool
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("uint-int", func(t *testing.T) {
		var dst int
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("uint-uint", func(t *testing.T) {
		var dst uint
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("uint-float", func(t *testing.T) {
		var dst float64
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), dst)
	})
	t.Run("uint-string", func(t *testing.T) {
		var dst string
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, "1", dst)
	})
	t.Run("uint-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(uint64(math.MaxUint32), &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, dst)
	})
	t.Run("uint-[4]byte", func(t *testing.T) {
		var dst [4]byte
		err := Map(uint64(math.MaxUint32), &dst)
		assert.NoError(t, err)
		assert.Equal(t, [4]byte{0xff, 0xff, 0xff, 0xff}, dst)
	})
	t.Run("uint-[4]byte#overflow", func(t *testing.T) {
		var dst [4]byte
		err := Map(uint64(math.MaxUint32)+1, &dst)
		assert.Error(t, err)
	})
	t.Run("uint-struct", func(t *testing.T) {
		var dst struct{ Uint uint }
		err := Map(uint(1), &dst)
		assert.Error(t, err)
	})

	// float
	t.Run("float-bool#false", func(t *testing.T) {
		var dst bool
		err := Map(float64(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("float-bool#true", func(t *testing.T) {
		var dst bool
		err := Map(float64(0), &dst)
		assert.NoError(t, err)
		assert.Equal(t, false, dst)
	})
	t.Run("float-int", func(t *testing.T) {
		var dst int
		err := Map(float64(1.9), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("float-int#overflow", func(t *testing.T) {
		var dst int
		err := Map(float64(math.MaxInt64)+1025, &dst)
		assert.Error(t, err)
	})
	t.Run("float-int#underflow", func(t *testing.T) {
		var dst int
		err := Map(float64(math.MinInt64)-1025, &dst)
		assert.Error(t, err)
	})
	t.Run("float-uint", func(t *testing.T) {
		var dst uint
		err := Map(float64(1.9), &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("float-uint#overflow", func(t *testing.T) {
		var dst uint
		err := Map(float64(math.MaxUint64)+2049, &dst)
		assert.Error(t, err)
	})
	t.Run("float-uint#underflow", func(t *testing.T) {
		var dst uint
		err := Map(float64(-1), &dst)
		assert.Error(t, err)
	})
	t.Run("float-float", func(t *testing.T) {
		var dst float64
		err := Map(float64(1.9), &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1.9), dst)
	})
	t.Run("float-string#1.9", func(t *testing.T) {
		var dst string
		err := Map(float64(1.9), &dst)
		assert.NoError(t, err)
		assert.Equal(t, "1.9", dst)
	})
	t.Run("float-string#MaxFloat64", func(t *testing.T) {
		var dst string
		err := Map(math.MaxFloat64, &dst)
		assert.NoError(t, err)
		assert.Equal(t, strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64), dst)
	})
	t.Run("float-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(math.MaxFloat64, &dst)
		assert.Error(t, err)
	})
	t.Run("float-[4]byte", func(t *testing.T) {
		var dst [4]byte
		err := Map(math.MaxFloat64, &dst)
		assert.Error(t, err)
	})
	t.Run("float-struct", func(t *testing.T) {
		var dst struct{ Float float64 }
		err := Map(float64(1.9), &dst)
		assert.Error(t, err)
	})

	// string
	t.Run("string-bool#false", func(t *testing.T) {
		var dst bool
		err := Map("false", &dst)
		assert.NoError(t, err)
		assert.Equal(t, false, dst)
	})
	t.Run("string-bool#true", func(t *testing.T) {
		var dst bool
		err := Map("true", &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("string-bool#invalid", func(t *testing.T) {
		var dst bool
		err := Map("foo", &dst)
		assert.Error(t, err)
	})
	t.Run("string-int#decimal", func(t *testing.T) {
		var dst int
		err := Map("255", &dst)
		assert.NoError(t, err)
		assert.Equal(t, 255, dst)
	})
	t.Run("string-int#hex", func(t *testing.T) {
		var dst int
		err := Map("0xFF", &dst)
		assert.NoError(t, err)
		assert.Equal(t, 255, dst)
	})
	t.Run("string-int#nagative-hex", func(t *testing.T) {
		var dst int
		err := Map("-0xFF", &dst)
		assert.NoError(t, err)
		assert.Equal(t, -255, dst)
	})
	t.Run("string-int#invalid", func(t *testing.T) {
		var dst int
		err := Map("foo", &dst)
		assert.Error(t, err)
	})
	t.Run("string-uint#decimal", func(t *testing.T) {
		var dst uint
		err := Map("255", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(255), dst)
	})
	t.Run("string-uint#hex", func(t *testing.T) {
		var dst uint
		err := Map("0xFF", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(255), dst)
	})
	t.Run("string-uint#negative-hex", func(t *testing.T) {
		var dst uint
		err := Map("-0xFF", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(255), dst)
	})
	t.Run("string-uint#invalid", func(t *testing.T) {
		var dst uint
		err := Map("foo", &dst)
		assert.Error(t, err)
	})
	t.Run("string-float#1.9", func(t *testing.T) {
		var dst float64
		err := Map("1.9", &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1.9), dst)
	})
	t.Run("string-float#MaxFloat64", func(t *testing.T) {
		var dst float64
		err := Map(strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64), &dst)
		assert.NoError(t, err)
		assert.Equal(t, math.MaxFloat64, dst)
	})
	t.Run("string-float#invalid", func(t *testing.T) {
		var dst float64
		err := Map("foo", &dst)
		assert.Error(t, err)
	})
	t.Run("string-string", func(t *testing.T) {
		var dst string
		err := Map("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, "foo", dst)
	})
	t.Run("string-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte("foo"), dst)
	})
	t.Run("string-[3]byte", func(t *testing.T) {
		var dst [3]byte
		err := Map("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, [3]byte{'f', 'o', 'o'}, dst)
	})
	t.Run("string-[2]byte#overflow", func(t *testing.T) {
		var dst [2]byte
		err := Map("foo", &dst)
		assert.Error(t, err)
	})
	t.Run("string-struct", func(t *testing.T) {
		var dst struct{ String string }
		err := Map("foo", &dst)
		assert.Error(t, err)
	})

	// []byte
	t.Run("[]byte-bool", func(t *testing.T) {
		var dst bool
		err := Map([]byte("false"), &dst)
		assert.Error(t, err)
	})
	t.Run("[]byte-int", func(t *testing.T) {
		var dst int
		err := Map([]byte{0x7f, 0xff, 0xff, 0xff}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, math.MaxInt32, dst)
	})
	t.Run("[]byte-uint", func(t *testing.T) {
		var dst uint
		err := Map([]byte{0xff, 0xff, 0xff, 0xff}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(math.MaxUint32), dst)
	})
	t.Run("[]byte-float", func(t *testing.T) {
		var dst float64
		err := Map([]byte("1.9"), &dst)
		assert.Error(t, err)
	})
	t.Run("[]byte-string", func(t *testing.T) {
		var dst string
		err := Map([]byte("foo"), &dst)
		assert.NoError(t, err)
		assert.Equal(t, "foo", dst)
	})
	t.Run("[]byte-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map([]byte("foo"), &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte("foo"), dst)
	})
	t.Run("[]int-[]float", func(t *testing.T) {
		var dst []float64
		err := Map([]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1, 2, 3}, dst)
	})
	t.Run("[]string-[]float#invalid", func(t *testing.T) {
		var dst []float64
		err := Map([]string{"1", "2", "foo"}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]byte-[3]byte", func(t *testing.T) {
		var dst [3]byte
		err := Map([]byte("foo"), &dst)
		assert.NoError(t, err)
		assert.Equal(t, [3]byte{'f', 'o', 'o'}, dst)
	})
	t.Run("[]string-[3]float#invalid", func(t *testing.T) {
		var dst [3]float64
		err := Map([]string{"1", "2", "foo"}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]byte-[3]byte#overflow", func(t *testing.T) {
		var dst [3]byte
		err := Map([]byte("foobar"), &dst)
		assert.Error(t, err)
	})
	t.Run("[]byte-struct", func(t *testing.T) {
		var dst struct{ String string }
		err := Map([]byte("foo"), &dst)
		assert.Error(t, err)
	})

	// slices
	t.Run("[]int-bool", func(t *testing.T) {
		var dst bool
		err := Map([]int{1}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]int-int", func(t *testing.T) {
		var dst int
		err := Map([]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]int-uint", func(t *testing.T) {
		var dst uint
		err := Map([]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]int-float", func(t *testing.T) {
		var dst float64
		err := Map([]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]int-string", func(t *testing.T) {
		var dst string
		err := Map([]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[]int-[]int", func(t *testing.T) {
		var dst []int
		err := Map([]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, dst)
	})
	t.Run("[]int-[]string", func(t *testing.T) {
		var dst []string
		err := Map([]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, dst)
	})
	t.Run("[]interface-[]int", func(t *testing.T) {
		var dst []int
		err := Map([]any{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, dst)
	})
	t.Run("[]int-[]interface", func(t *testing.T) {
		var dst []any
		err := Map([]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []any{1, 2, 3}, dst)
	})
	t.Run("[]int-struct", func(t *testing.T) {
		var dst struct{ String string }
		err := Map([]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})

	// arrays
	t.Run("[3]int-bool", func(t *testing.T) {
		var dst bool
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-int", func(t *testing.T) {
		var dst int
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-uint", func(t *testing.T) {
		var dst uint
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-float", func(t *testing.T) {
		var dst float64
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-string", func(t *testing.T) {
		var dst string
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-[]int", func(t *testing.T) {
		var dst []int
		err := Map([3]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, dst)
	})
	t.Run("[3]string-[]float#invalid", func(t *testing.T) {
		var dst []float64
		err := Map([3]string{"1", "2", "foo"}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-[3]int", func(t *testing.T) {
		var dst [3]int
		err := Map([3]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, [3]int{1, 2, 3}, dst)
	})
	t.Run("[3]int-[3]string", func(t *testing.T) {
		var dst [3]string
		err := Map([3]int{1, 2, 3}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, [3]string{"1", "2", "3"}, dst)
	})
	t.Run("[3]string-[3]float#invalid", func(t *testing.T) {
		var dst [3]float64
		err := Map([3]string{"1", "2", "foo"}, &dst)
		assert.Error(t, err)
	})
	t.Run("[2]int-[3]string#underflow", func(t *testing.T) {
		var dst [3]string
		err := Map([2]int{1, 2}, &dst)
		assert.Error(t, err)
	})
	t.Run("[4]int-[3]string#overflow", func(t *testing.T) {
		var dst [3]string
		err := Map([4]int{1, 2, 3, 4}, &dst)
		assert.Error(t, err)
	})
	t.Run("[3]int-struct", func(t *testing.T) {
		var dst struct{ String string }
		err := Map([3]int{1, 2, 3}, &dst)
		assert.Error(t, err)
	})

	// maps
	t.Run("map-bool", func(t *testing.T) {
		var dst bool
		err := Map(map[string]any{"foo": "bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("map-int", func(t *testing.T) {
		var dst int
		err := Map(map[string]any{"foo": "bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("map-uint", func(t *testing.T) {
		var dst uint
		err := Map(map[string]any{"foo": "bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("map-float", func(t *testing.T) {
		var dst float64
		err := Map(map[string]any{"foo": "bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("map-string", func(t *testing.T) {
		var dst string
		err := Map(map[string]any{"foo": "bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("map-map", func(t *testing.T) {
		var dst map[string]int
		err := Map(map[string]int{"foo": 1}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]int{"foo": 1}, dst)
	})
	t.Run("map-struct", func(t *testing.T) {
		type Dst struct {
			Foo string
			Bar string `map:"a.bar"`
			Baz string `map:"a.baz"`
			Qux string `map:"-"`
		}
		var dst Dst
		err := Map(map[string]any{
			"Foo": 1,
			"a": map[string]any{
				"bar": 2,
				"baz": true,
			},
			"qux": "qux",
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{Foo: "1", Bar: "2", Baz: "true", Qux: ""}, dst)
	})
	t.Run("map-struct#invalid", func(t *testing.T) {
		type Dst struct{ Foo float64 }
		var dst Dst
		err := Map(map[string]any{"Foo": "bar"}, &dst)
		assert.Error(t, err)
	})

	// structs
	t.Run("struct-bool", func(t *testing.T) {
		var dst bool
		err := Map(struct{ Foo string }{"bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("struct-int", func(t *testing.T) {
		var dst int
		err := Map(struct{ Foo string }{"bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("struct-uint", func(t *testing.T) {
		var dst uint
		err := Map(struct{ Foo string }{"bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("struct-float", func(t *testing.T) {
		var dst float64
		err := Map(struct{ Foo string }{"bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("struct-string", func(t *testing.T) {
		var dst string
		err := Map(struct{ Foo string }{"bar"}, &dst)
		assert.Error(t, err)
	})
	t.Run("struct-map", func(t *testing.T) {
		type Src struct {
			Foo int
		}
		var dst map[string]any
		err := Map(Src{
			Foo: 1,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"Foo": 1,
		}, dst)
	})
	t.Run("struct-map#tag", func(t *testing.T) {
		type Src struct {
			Foo int `map:"foo"`
		}
		var dst map[string]any
		err := Map(Src{
			Foo: 1,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"foo": 1,
		}, dst)
	})
	t.Run("struct-map#tag-nested-fields", func(t *testing.T) {
		type Src struct {
			Foo int `map:"a.b.foo"`
			Bar int `map:"a.b.bar"`
			Baz int `map:"a.baz"`
		}
		var dst map[string]any
		err := Map(Src{
			Foo: 1,
			Bar: 2,
			Baz: 3,
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"a": map[string]any{
				"b": map[string]any{
					"foo": 1,
					"bar": 2,
				},
				"baz": 3,
			},
		}, dst)
	})
	t.Run("struct-struct", func(t *testing.T) {
		type Src struct {
			Foo int
			Bar string
			Baz []int
		}
		type Dst struct {
			Foo string
			Bar int
			Baz []string
		}
		var dst Dst
		err := Map(Src{
			Foo: 1,
			Bar: "2",
			Baz: []int{3, 4, 5},
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			Foo: "1",
			Bar: 2,
			Baz: []string{"3", "4", "5"},
		}, dst)
	})
	t.Run("struct-struct#tags", func(t *testing.T) {
		type Src struct {
			Foo int    `map:"X"`
			Bar string `map:"Bar"`
			Baz []int  `map:"C"`
		}
		type Dst struct {
			A string `map:"X"`
			B int    `map:"Bar"`
			C []string
		}
		var dst Dst
		err := Map(Src{
			Foo: 1,
			Bar: "2",
			Baz: []int{3, 4, 5},
		}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, Dst{
			A: "1",
			B: 2,
			C: []string{"3", "4", "5"},
		}, dst)
	})
}

func TestNilValues(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		var dst *string
		err := Map("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, "foo", *dst)
	})
	t.Run("map", func(t *testing.T) {
		var dst map[string]any
		err := Map(map[string]any{"Foo": "foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"Foo": "foo"}, dst)
	})
	t.Run("pointer-map", func(t *testing.T) {
		var dst *map[string]any
		err := Map(map[string]any{"Foo": "foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"Foo": "foo"}, *dst)
	})
	t.Run("slice", func(t *testing.T) {
		var dst []string
		err := Map([]string{"foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo"}, dst)
	})
	t.Run("pointer-slice", func(t *testing.T) {
		var dst *[]string
		err := Map([]string{"foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo"}, *dst)
	})
	t.Run("struct", func(t *testing.T) {
		type Dst struct{ Foo string }
		var dst *Dst
		err := Map(map[string]any{"Foo": "foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, &Dst{Foo: "foo"}, dst)
	})
	t.Run("struct-field-pointer", func(t *testing.T) {
		type Dst struct{ Foo *string }
		var dst *Dst
		err := Map(map[string]any{"Foo": "foo"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, &Dst{Foo: strPtr("foo")}, dst)
	})
	t.Run("struct-field-map", func(t *testing.T) {
		type Dst struct{ Foo map[string]any }
		var dst *Dst
		err := Map(map[string]any{"Foo": map[string]any{"Bar": "bar"}}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, &Dst{Foo: map[string]any{"Bar": "bar"}}, dst)
	})
	t.Run("struct-field-slice", func(t *testing.T) {
		type Dst struct{ Foo []string }
		var dst *Dst
		err := Map(map[string]any{"Foo": []string{"foo"}}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, &Dst{Foo: []string{"foo"}}, dst)
	})
}

func TestStrictTypes(t *testing.T) {
	m := DefaultMapper.Copy()
	m.StrictTypes = true
	t.Run("bool", func(t *testing.T) {
		require.NoError(t, m.Map(true, new(bool)))
		require.Error(t, m.Map(true, new(int)))
		require.Error(t, m.Map(true, new(uint)))
		require.Error(t, m.Map(true, new(float64)))
		require.Error(t, m.Map(true, new(string)))
	})
	t.Run("int", func(t *testing.T) {
		require.Error(t, m.Map(1, new(bool)))
		require.NoError(t, m.Map(1, new(int)))
		require.Error(t, m.Map(1, new(uint)))
		require.Error(t, m.Map(1, new(float64)))
		require.Error(t, m.Map(1, new(string)))
		require.Error(t, m.Map(1, make([]byte, 8)))
		require.Error(t, m.Map(1, new([8]byte)))
	})
	t.Run("uint", func(t *testing.T) {
		require.Error(t, m.Map(uint(1), new(bool)))
		require.Error(t, m.Map(uint(1), new(int)))
		require.NoError(t, m.Map(uint(1), new(uint)))
		require.Error(t, m.Map(uint(1), new(float64)))
		require.Error(t, m.Map(uint(1), new(string)))
		require.Error(t, m.Map(uint(1), make([]byte, 8)))
		require.Error(t, m.Map(uint(1), new([8]byte)))
	})
	t.Run("float", func(t *testing.T) {
		require.Error(t, m.Map(1.0, new(bool)))
		require.Error(t, m.Map(1.0, new(int)))
		require.Error(t, m.Map(1.0, new(uint)))
		require.NoError(t, m.Map(1.0, new(float64)))
		require.Error(t, m.Map(1.0, new(string)))
	})
	t.Run("string", func(t *testing.T) {
		require.Error(t, m.Map("true", new(bool)))
		require.Error(t, m.Map("1", new(int)))
		require.Error(t, m.Map("1", new(uint)))
		require.Error(t, m.Map("1", new(float64)))
		require.NoError(t, m.Map("foo", new(string)))
		require.Error(t, m.Map("foo", new([]byte)))
		require.Error(t, m.Map("foo", new([3]byte)))
	})
	t.Run("slice", func(t *testing.T) {
		require.Error(t, m.Map([]byte{0x01}, new(int)))
		require.Error(t, m.Map([]byte{0x01}, new(uint)))
		require.NoError(t, m.Map([]string{"foo"}, new([]string)))
		require.Error(t, m.Map([]int{1}, new([]string)))
		require.Error(t, m.Map([]string{"foo"}, new([1]string)))
	})
	t.Run("array", func(t *testing.T) {
		require.Error(t, m.Map([1]byte{0x01}, new(int)))
		require.Error(t, m.Map([1]byte{0x01}, new(uint)))
		require.NoError(t, m.Map([1]string{"foo"}, new([1]string)))
		require.Error(t, m.Map([1]int{1}, new([]string)))
		require.Error(t, m.Map([1]int{1}, new([1]string)))
	})
	t.Run("map", func(t *testing.T) {
		require.NoError(t, m.Map(map[string]string{"foo": "bar"}, new(map[string]string)))
		require.Error(t, m.Map(map[string]int{"foo": 1}, new(map[string]string)))
		require.NoError(t, m.Map(map[string]string{"Foo": "bar"}, new(struct{ Foo string })))
		require.Error(t, m.Map(map[string]int{"Foo": 1}, new(struct{ Foo string })))
	})
	t.Run("struct", func(t *testing.T) {
		require.NoError(t, m.Map(struct{ Foo string }{"bar"}, new(struct{ Foo string })))
		require.Error(t, m.Map(struct{ Foo int }{1}, new(struct{ Foo string })))
		require.NoError(t, m.Map(struct{ Foo string }{"bar"}, new(map[string]string)))
		require.Error(t, m.Map(struct{ Foo int }{1}, new(map[string]string)))
	})
}

func TestUnaddressableValues(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		var dst *string
		err := Map("bar", dst)
		assert.Error(t, err)
	})
	t.Run("bool", func(t *testing.T) {
		var dst bool
		err := Map(true, dst)
		assert.Error(t, err)
	})
	t.Run("int", func(t *testing.T) {
		var dst int
		err := Map(1, dst)
		assert.Error(t, err)
	})
	t.Run("uint", func(t *testing.T) {
		var dst uint
		err := Map(uint(1), dst)
		assert.Error(t, err)
	})
	t.Run("float", func(t *testing.T) {
		var dst float64
		err := Map(1.0, dst)
		assert.Error(t, err)
	})
	t.Run("string", func(t *testing.T) {
		var dst string
		err := Map("foo", dst)
		assert.Error(t, err)
	})
	t.Run("slice", func(t *testing.T) {
		var dst []string
		err := Map([]string{"foo"}, dst)
		assert.Error(t, err)
	})
	t.Run("map", func(t *testing.T) {
		var dst map[string]any
		err := Map(map[string]any{"foo": "bar"}, dst)
		assert.Error(t, err)
	})
	t.Run("struct", func(t *testing.T) {
		type Dst struct{ Foo string }
		var dst Dst
		err := Map(map[string]any{"Foo": "foo"}, dst)
		assert.Error(t, err)
	})
}

func TestFieldMapper(t *testing.T) {
	m := DefaultMapper.Copy()
	m.FieldMapper = func(name string) string {
		return strings.ToLower(name)
	}
	type Src struct {
		FOO string
		BAR string `map:"BAR"` // field mapper is ignored for tagged fields
	}
	var dst map[string]any
	err := m.Map(Src{
		FOO: "foo",
		BAR: "bar",
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"foo": "foo",
		"BAR": "bar",
	}, dst)
}

func TestCustomTagAndSeparator(t *testing.T) {
	m := DefaultMapper.Copy()
	m.Tag = "tag"
	m.Separator = ":"
	type Src struct {
		Foo string `tag:"foo:foo"`
	}
	var dst map[string]any
	err := m.Map(Src{
		Foo: "foo",
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"foo": map[string]any{"foo": "foo"},
	}, dst)
}

func strPtr(s string) *string {
	return &s
}
