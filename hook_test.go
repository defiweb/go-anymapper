package anymapper

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	m := New()
	m.Hooks = MappingInterfaceHooks

	t.Run("mapFrom", func(t *testing.T) {
		var dst customType
		require.NoError(t, m.Map("foo", &dst))
		assert.Equal(t, "foo", dst.foo)
	})
	t.Run("mapTo", func(t *testing.T) {
		var dst string
		require.NoError(t, m.Map(customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("mapFromPtr", func(t *testing.T) {
		var dst *customType
		require.NoError(t, m.Map("foo", &dst))
		assert.Equal(t, "foo", dst.foo)
	})
	t.Run("mapToPtr", func(t *testing.T) {
		var dst string
		require.NoError(t, m.Map(&customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("mapToAny", func(t *testing.T) {
		var dst any
		require.NoError(t, m.Map(&customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst)
	})
	t.Run("both", func(t *testing.T) {
		var dst customType
		require.NoError(t, m.Map(customType{foo: "foo"}, &dst))
		assert.Equal(t, "foo", dst.foo)
	})
}
