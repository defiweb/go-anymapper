# go-anymapper

The `go-anymapper` package is a fast and convenient tool for mapping data between different types, including basic Go
types like strings and integers, as well as more complex data structures. It allows you to create custom mapping rules
to fit the unique requirements of your application. This means you can use `go-anymapper` to easily convert data in the
most useful way for your specific needs.

## Installation

```bash
go get -u github.com/defiweb/go-anymapper
```

## Usage

The simplest way to use the `go-anymapper` package is to use the `Map` function. It takes two arguments: the source and
the destination. The function will try to map the source to the destination using the following rules:

- If the dst value is an empty interface, the src value is assigned to it.
- `bool` ⇔ `intX`, `uintX`, `floatX` ⇒ `true` ⇔ `1`, `false` ⇔ `0` (if source is number, then `≠0` ⇒ `true`).
- `intX`, `uintX`, `floatX` ⇔ `intX`, `uintX`, `floatX` ⇒ cast numbers to the destination type.
- `intX`, `uintX`, `floatX` ⇔ `[]byte` ⇒ converts using `binary.Read` and `binary.Write`.
- `intX`, `uintX`, `floatX` ⇔ `[X]byte` ⇒ converts using `binary.Read` and `binary.Write`.
- `string` ⇔ `intX`, `uintX` ⇒ converts using `big.Int.SetString` and `big.Int.String`.
- `string` ⇔ `floatX` ⇒ converts string to or from number using `big.Float.SetString` and `big.Float.String`.
- `string` ⇔ `[]byte` ⇒ converts using `[]byte(s)` and `string(b)`.
- `slice` ⇔ `slice` ⇒ recursively map each slice element.
- `slice` ⇔ `array` ⇒ recursively map each slice element if lengths are the same.
- `array` ⇔ `array` ⇒ recursively map each array element if lengths are the same.
- `map` ⇔ `map` ⇒ recursively map every key and value pair.
- `struct` ⇔ `struct` ⇒ recursively map every struct field.
- `struct` ⇔ `map[string]X` ⇒ map struct fields to map elements using field names as keys and vice versa.

The above types refer to the type kind, not the actual type, hence `type MyInt int` is also considered as `int`.

In addition to the above rules, the default configuration of the mapper supports the following conversions:

- `time.Time` ⇔ `string` ⇒ converts string to or from time using RFC3339 format.
- `time.Time` ⇔  `uint`, `uint32`, `uint64`, `int`, `int32`, `int64` ⇒ convert using Unix timestamp.
- `time.Time` ⇔  `uint8`, `uint16`, `int8`, `int16` ⇒ not allowed.
- `time.Time` ⇔  `floatX` ⇒ convert to or from unix timestamp, preserving the fractional part of a second.
- `time.Time` ⇔  `big.Int` ⇒ convert using Unix timestamp.
- `time.Time` ⇔  `big.Float` ⇒ convert using Unix timestamp, preserving the fractional part of a second.
- `time.Time` ⇔  _other_ ⇒ try to convert using `int64` as intermediate value.
- `big.Int` ⇔ `intX`, `uintX`, `floatX` ⇒ convert using `big.Int.Int64` and `big.Int.SetUint64`.
- `big.Int` ⇔ `string` ⇒ converts using `big.Int.String` and `big.Int.SetString`.
- `big.Int` ⇔ `[]byte` ⇒ converts using `big.Int.Bytes` and `big.Int.SetBytes`.
- `big.Int` ⇔ `big.Float` ⇒ coverts using `big.Float.Int` and `big.Float.SetInt`.
- `big.Float` ⇔ `intX`, `uintX` ⇒ convert using `big.Float.Int64` and `big.Float.SetUint64`.
- `big.Float` ⇔ `floatX` ⇒ convert using `big.Float.Float64` and `big.Float.SetFloat64`.
- `big.Float` ⇔ `string` ⇒ converts to or from string using `big.Float.String` and `big.Float.SetString`.
- `big.Rat` ⇔ `string` ⇒ converts to or from string using `big.Rat.String` and `big.Rat.SetString`.
- `big.Rat` ⇔ `big.Float` ⇒ converts using `big.Float.SetRat` and `big.Float.Rat`.
- `big.Rat` ⇔ `slice`, `[2]array` ⇒ convert first element to/from numerator and second to/form denominator.
- `big.Rat` ⇔ _other_ ⇒ try to convert using `big.Float` as intermediate value.

Mapping will fail if the target type is not large enough to hold the source value. For example, mapping `int64`
to `int8` may fail because `int64` can store values larger than `int8`.

When mapping numbers from a byte slice or array, the length of the slice/array *must* be the same as the size of the
variable in bytes. The size of `int`, `uint` is always considered as 64 bits.

The mapper will not overwrite the values in the destination if they do not have corresponding values in the source. For
slices, if the destination slice is longer than the source slice, the extra elements will remain unchanged.

When using the mapper to convert values to interface types, it will attempt to use existing elements in the destination
if possible. For example, mapping `[]int{1, 2}` to `[]any{"", 0}` will result in `[]any{"1", 2}`, allowing to easily
assign values to a specific implementation of an interface.

### Mapping structures

Structures are treated by mapper as key-value maps. The mapper will try to map recursively every field of the source
structure to the corresponding field of the destination structure or map.

Field names can be overridden with a tag (whose name is defined in `Mapper.Tag`, default is `map`).

As a special case, if the field tag is "-", the field is always omitted.

If the tag is not set, struct field names will be mapped using the `Mapper.FieldNameMapper` function.

Tags can be defined for both source and target structures. In this case, the names used in the tags must be the same for
both structures.

If destination structure has fields that are not present in the source structure, the mapper will set zero values for
those fields.

### Strict types

If `Mapper.StrictTypes` is set to true, strict type checking will be enforced for the mapping process. This means that the
source and destination types must be exactly the same for the mapping to be successful. However, mapping between
different data structures, such as `struct` ⇔ `struct`, `struct` ⇔ `map` and `map` ⇔ `map` is always allowed. If the
destination type is an empty interface, the source value will be assigned to it regardless of the strict type check
setting.

Additionally, the strict type check applies to custom types as well. For example, a custom type `type MyInt int` will
not be treated as `int` anymore.

### Custom mapping functions

If it is not possible to implement the above interfaces, custom mapping functions can be registered with the
`Mapper.Mapper` map. The keys of this map are the types of the destination or source values, and the values are
functions that return a `MapFunc` function that can map the source value to the destination value.

If the function returns a `nil` value, it means that the mapping is not possible. If both the source and destination
types are registered, the source type will be used first. If it returns a nil value, the destination type will be used.
If neither of them returns a `nil` value, the mapping will fail.

### `MapTo` and `MapFrom` interfaces:

**This feature is disabled by default. To enable it, set `Mapper.Hooks` to `Mapper.MappingInterfaceHooks`.**

The `go-anymapper` package provides two interfaces that can be implemented by the source and destination types to
customize the mapping process.

If the source value implements `MapTo` interface, the `MapTo` method will be used to map the source value to the
destination value.

If the destination value implements `MapFrom` interface, the `MapFrom` method will be used to map the source value to
the destination value.

If both source and destination values implement the `MapTo` and `MapFrom` interfaces then only `MapTo` will be used.

### Default mapper instance

The package defines the default mapper instance `Default` that is used by `Map` and `MapRefl` functions. It is
possible to change configuration of the default mapper, but it may affect other packages that use the default mapper. To
avoid this, it is recommended to create a new instance of the mapper using the `New` method.

## Examples

### Mapping between simple types

```go
package main

import (
	"fmt"

	"github.com/defiweb/go-anymapper"
)

func main() {
	var a int = 42
	var b string

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b) // "42"
}
```

### Mapping between structure and map

```go
package main

import (
	"fmt"

	"github.com/defiweb/go-anymapper"
)

type Data struct {
	Foo int `map:"bar"`
	Bar int `map:"foo"`
}

func main() {
	a := Data{Foo: 42, Bar: 1337}
	b := make(map[string]uint64)

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b) // map[bar:42 foo:1337]
}
```

### MapFrom and MapTo interfaces

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/defiweb/go-anymapper"
)

type Val struct {
	X *big.Int
}

func (v *Val) MapFrom(m *anymapper.Mapper, x reflect.Value) error {
	return m.Map(x.Interface(), &v.X)
}

func (v *Val) MapTo(m *anymapper.Mapper, x reflect.Value) error {
	if v.X == nil {
		return m.Map(0, x.Addr().Interface())
	}
	return m.Map(v.X, x.Addr().Interface())
}

func main() {
	var a int = 42
	var b Val
	
	// Enable MapTo and MapFrom interfaces:
	anymapper.Default.Hooks = anymapper.MappingInterfaceHooks

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.X.String()) // "42"
}
```

### Custom mapping function

```go
package main

import (
	"fmt"
	"reflect"
	"math/big"

	"github.com/defiweb/go-anymapper"
)

type Val struct {
	X *big.Int
}

func main() {
	var a int = 42
	var b Val

	typ := reflect.TypeOf(Val{})
	anymapper.DefaultMapper.Mappers[typ] = func(m *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
		if src == typ {
			return func(m *anymapper.Mapper, src, dst reflect.Value) error {
				return m.MapRefl(src.FieldByName("X"), dst)
			}
		}
		if dst == typ {
			return func(m *anymapper.Mapper, src, dst reflect.Value) error {
				return m.MapRefl(src, reflect.ValueOf(&dst.Addr().Interface().(*Val).X))
			}
		}
		return nil
	}

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.X.String()) // "42"
}
```

### Benchmark

Following benchmarks compare the performance of the `go-anymapper` package with the `mapstructure` package.

```go
package main

import (
	"testing"

	"github.com/defiweb/go-anymapper"
	"github.com/mitchellh/mapstructure"
)

func Benchmark(b *testing.B) {
	type Object struct {
		A string
		B int
		C []string
		D []any
		E map[string]string
	}
	b.Run("anymapper/map-struct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			input := map[string]interface{}{
				"A": "a",
				"B": 1,
				"C": []string{"a", "b", "c"},
				"D": []any{1, "2", 3.0},
				"E": map[string]string{"a": "a", "b": "b", "c": "c"},
			}
			var result Object
			err := anymapper.Map(input, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("anymapper/struct-map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			input := Object{
				A: "a",
				B: 1,
				C: []string{"a", "b", "c"},
				D: []any{1, "2", 3.0},
				E: map[string]string{"a": "a", "b": "b", "c": "c"},
			}
			var result map[string]any
			err := anymapper.Map(input, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("mapstructure/map-struct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			input := map[string]interface{}{
				"A": "a",
				"B": 1,
				"C": []string{"a", "b", "c"},
				"D": []any{1, "2", 3.0},
				"E": map[string]string{"a": "a", "b": "b", "c": "c"},
			}
			var result Object
			err := mapstructure.Decode(input, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("mapstructure/struct-map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			input := Object{
				A: "a",
				B: 1,
				C: []string{"a", "b", "c"},
				D: []any{1, "2", 3.0},
				E: map[string]string{"a": "a", "b": "b", "c": "c"},
			}
			var result map[string]any
			err := mapstructure.Decode(input, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
```

Results:

```
BenchmarK/anymapper/map-struct         	  972992	      1174 ns/op
Benchmark/anymapper/struct-map         	  903348	      1311 ns/op
BenchmarK/mapstructure/map-struct      	  339668	      3501 ns/op
Benchmark/mapstructure/struct-map      	 1354458	      889.5 ns/op
```

## Documentation

[https://pkg.go.dev/github.com/defiweb/go-anymapper](https://pkg.go.dev/github.com/defiweb/go-anymapper)
