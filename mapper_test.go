package anymapper

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidValues(t *testing.T) {
	t.Run("invalid-src", func(t *testing.T) {
		var dst string
		err := MapRefl(reflect.Value{}, reflect.ValueOf(&dst))
		assert.Error(t, err)
	})
	t.Run("invalid-dst", func(t *testing.T) {
		err := MapRefl(reflect.ValueOf("foo"), reflect.Value{})
		assert.Error(t, err)
	})
	t.Run("unaddressable-dst", func(t *testing.T) {
		var dst string
		err := MapRefl(reflect.ValueOf("foo"), reflect.ValueOf(dst))
		assert.Error(t, err)
	})
}

type customType struct {
	foo string
}

func (c *customType) MapFrom(m *Mapper, src reflect.Value) error {
	return m.MapRefl(src, reflect.ValueOf(&c.foo))
}

func (c customType) MapTo(m *Mapper, dst reflect.Value) error {
	return m.MapRefl(reflect.ValueOf(c.foo), dst)
}

func TestCustomType(t *testing.T) {
	t.Run("mapFrom", func(t *testing.T) {
		var dst customType
		require.NoError(t, Map("foo", &dst))
		assert.Equal(t, "foo", dst.foo)
	})
	t.Run("mapTo", func(t *testing.T) {
		var dst string
		require.NoError(t, Map(customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("mapFromPtr", func(t *testing.T) {
		var dst *customType
		require.NoError(t, Map("foo", &dst))
		assert.Equal(t, "foo", dst.foo)
	})
	t.Run("mapToPtr", func(t *testing.T) {
		var dst string
		require.NoError(t, Map(&customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("mapToAny", func(t *testing.T) {
		var dst any
		require.NoError(t, Map(&customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("both", func(t *testing.T) {
		var dst customType
		require.NoError(t, Map(customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst.foo)
	})
}

func TestCustomMapFunc(t *testing.T) {
	type customType struct {
		Foo string
	}
	typ := reflect.TypeOf(customType{})
	m := DefaultMapper.Copy()
	m.Mappers[typ] = func(m *Mapper, src, dst reflect.Type) MapFunc {
		if src == typ {
			return func(m *Mapper, src, dst reflect.Value) error {
				return m.MapRefl(src.FieldByName("Foo"), dst)
			}
		}
		if dst == typ {
			return func(m *Mapper, src, dst reflect.Value) error {
				return m.MapRefl(src, reflect.ValueOf(&dst.Addr().Interface().(*customType).Foo))
			}
		}
		return nil
	}
	t.Run("mapFrom", func(t *testing.T) {
		var dst customType
		require.NoError(t, m.Map("foo", &dst))
		assert.Equal(t, "foo", dst.Foo)
	})
	t.Run("mapTo", func(t *testing.T) {
		var dst string
		require.NoError(t, m.Map(customType{Foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("mapFromPtr", func(t *testing.T) {
		var dst *customType
		require.NoError(t, m.Map("foo", &dst))
		assert.Equal(t, "foo", dst.Foo)
	})
	t.Run("mapToPtr", func(t *testing.T) {
		var dst string
		require.NoError(t, m.Map(&customType{Foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("both", func(t *testing.T) {
		var dst customType
		require.NoError(t, m.Map(customType{Foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst.Foo)
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

func TestEmptyTag(t *testing.T) {
	m := DefaultMapper.Copy()
	m.Tag = ""
	type Src struct {
		Foo string `map:"bar"`
	}
	var dst map[string]any
	err := m.Map(Src{
		Foo: "foo",
	}, &dst)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"Foo": "foo"}, dst)
}

func TestCopy(t *testing.T) {
	cpy := DefaultMapper.Copy()
	assert.Equal(t, DefaultMapper.Tag, cpy.Tag)
	assert.Equal(t, DefaultMapper.ByteOrder, cpy.ByteOrder)
	assert.Equal(t, &DefaultMapper.FieldMapper, &cpy.FieldMapper)
	assert.Equal(t, len(DefaultMapper.Mappers), len(cpy.Mappers))
	for k, v := range DefaultMapper.Mappers {
		rv1 := reflect.ValueOf(v)
		rv2 := reflect.ValueOf(cpy.Mappers[k])
		assert.Equal(t, rv1.Pointer(), rv2.Pointer())
	}
}

func TestInvalidMappingErr_WithReason(t *testing.T) {
	err := InvalidMappingErr{From: reflect.TypeOf(1), To: reflect.TypeOf("a"), Reason: "reason"}
	assert.Equal(t, "mapper: cannot map int to string: reason", err.Error())
}

func TestInvalidMappingErr_WithoutReason(t *testing.T) {
	err := InvalidMappingErr{From: reflect.TypeOf(1), To: reflect.TypeOf("a")}
	assert.Equal(t, "mapper: cannot map int to string", err.Error())
}

func Benchmark(b *testing.B) {
	b.Run("struct->struct", func(b *testing.B) {
		type Src struct {
			A int
			B int
			C int
			D int
		}
		type Dst struct {
			A string
			B string
			C string
			D string
		}
		src := Src{
			A: 1,
			B: 2,
			C: 3,
			D: 4,
		}
		dst := Dst{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("struct->map", func(b *testing.B) {
		type Src struct {
			A int
			B int
			C int
			D int
		}
		src := Src{
			A: 1,
			B: 2,
			C: 3,
			D: 4,
		}
		dst := map[string]string{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("map->struct", func(b *testing.B) {
		src := map[string]int{
			"A": 1,
			"B": 2,
			"C": 3,
			"D": 4,
		}
		type Dst struct {
			A string
			B string
			C string
			D string
		}
		dst := Dst{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("map->map", func(b *testing.B) {
		src := map[string]int{
			"A": 1,
			"B": 2,
			"C": 3,
			"D": 4,
		}
		dst := map[string]string{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("[]int->[]int", func(b *testing.B) {
		src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		var dst []int
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("[]int->MyInt", func(b *testing.B) {
		type MyInt int
		src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		var dst []MyInt
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
	b.Run("[]int->any", func(b *testing.B) {
		src := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		var dst []any
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Map(src, &dst)
		}
	})
}

func ptr(v any) any {
	r := reflect.New(reflect.TypeOf(v)).Elem()
	r.Set(reflect.ValueOf(v))
	return r.Addr().Interface()
}

func exp(v any) any {
	r := reflect.ValueOf(v)
	for r.Kind() == reflect.Interface {
		r = r.Elem()
	}
	if r.Kind() == reflect.Ptr {
		return r.Interface()
	}
	return ptr(r.Interface())
}

func dst(v any) any {
	r := reflect.ValueOf(v)
	for r.Kind() == reflect.Interface {
		r = r.Elem()
	}
	return r.Interface()
}

func anySlice() any {
	return []any{}
}
