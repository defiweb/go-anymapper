# go-anymapper

[![Run Tests](https://github.com/defiweb/go-anymapper/actions/workflows/test.yml/badge.svg)](https://github.com/defiweb/go-anymapper/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/defiweb/go-anymapper.svg)](https://pkg.go.dev/github.com/defiweb/go-anymapper)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`go-anymapper` maps data between any Go types at runtime. 

```go
type Order struct {
	ID     int    `map:"id"`
	Amount string `map:"amount"`
}

var order Order
err := anymapper.Map(map[string]any{"id": "42", "amount": 1.5}, &order)
// order = Order{ID: 42, Amount: "1.5"}
```

**Contents**

<!-- TOC -->
* [go-anymapper](#go-anymapper)
  * [Installation](#installation)
  * [Usage](#usage)
  * [Mapping rules](#mapping-rules)
  * [Additional types](#additional-types)
  * [Destination values](#destination-values)
  * [Structures](#structures)
  * [Context](#context)
  * [Strict types](#strict-types)
  * [Examples](#examples)
    * [Mapping between simple types](#mapping-between-simple-types)
    * [Mapping between structure and map](#mapping-between-structure-and-map)
    * [Mapping to interfaces](#mapping-to-interfaces)
    * [MapFrom and MapTo interfaces](#mapfrom-and-mapto-interfaces)
    * [Custom mapping function](#custom-mapping-function)
  * [Benchmark](#benchmark)
  * [Documentation](#documentation)
  * [License](#license)
<!-- TOC -->

## Installation

```bash
go get -u github.com/defiweb/go-anymapper
```

The package requires Go 1.18 or later.

## Usage

The package provides four functions. The destination must be a pointer, or another value that the mapper can set.

| Function                             | Description                                                       |
|--------------------------------------|-------------------------------------------------------------------|
| `Map(src, dst any) error`            | Maps the source value to the destination value.                   |
| `MapContext(ctx, src, dst) error`    | Same as `Map`, but uses the given [context](#context).            |
| `MapRefl(src, dst reflect.Value)`    | Same as `Map`, but takes `reflect.Value` arguments.               |
| `MapReflContext(ctx, src, dst)`      | Same as `MapRefl`, but uses the given [context](#context).        |

These functions use the [default mapper instance](#default-mapper-instance). The same functions are available as methods
of the `Mapper` type.

## Mapping rules

The mapper uses the rules below. The types in the table refer to the type kind, not to the actual type, hence
`type MyInt int` is also considered as `int`.

| Types                                                 | Conversion                                                                                         |
|-------------------------------------------------------|----------------------------------------------------------------------------------------------------|
| _any type_ ⇒ `any` (empty interface)                  | Assigns the source value to the interface. See [destination values](#destination-values).          |
| `bool` ⇔ `intX`, `uintX`, `floatX`                    | `false` ⇔ `0` and `true` ⇔ `1`. Every number other than `0` becomes `true`.                        |
| `bool` ⇔ `string`                                     | `false` ⇔ `"false"` and `true` ⇔ `"true"`. Every other string is an error.                         |
| `intX`, `uintX`, `floatX` ⇔ `intX`, `uintX`, `floatX` | Casts the number to the destination type.                                                          |
| `intX`, `uintX`, `floatX` ⇔ `[]byte`, `[X]byte`       | Converts with `binary.Read` and `binary.Write`.                                                    |
| `string` ⇔ `intX`, `uintX`                            | Converts with `strconv.ParseInt`/`strconv.ParseUint` and `strconv.FormatInt`/`strconv.FormatUint`. |
| `string` ⇔ `floatX`                                   | Converts with `strconv.ParseFloat` and `strconv.FormatFloat`.                                      |
| `string` ⇔ `[]byte`, `[X]byte`                        | Converts with `[]byte(s)` and `string(b)`.                                                         |
| `slice` ⇔ `slice`                                     | Maps every element.                                                                                |
| `slice` ⇔ `array`                                     | Maps every element, if the lengths are the same.                                                   |
| `array` ⇔ `array`                                     | Maps every element, if the lengths are the same.                                                   |
| `map` ⇔ `map`                                         | Maps every key and every value.                                                                    |
| `struct` ⇔ `struct`                                   | Maps every exported field. See [structures](#structures).                                          |
| `struct` ⇔ `map[string]X`                             | Uses the field names as map keys.                                                                  |

The mapping fails if the destination type is not large enough to hold the source value. For example, mapping `int64` to
`int8` may fail, because `int64` can store values larger than `int8`.

Numbers and byte slices use the byte order from `Context.ByteOrder`, which is big-endian by default. When the mapper
converts a byte slice or a byte array to a number, the length of the slice or array *must* be the same as the size of
the number in bytes. The size of `int` and `uint` is always considered as 64 bits, so the same data can be mapped on
32-bit and 64-bit architectures.

## Additional types

The default configuration adds the rules below for `time.Time` and for the `math/big` types.

| Types                                                             | Conversion                                                                                                       |
|-------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `time.Time` ⇔ `string`                                            | Converts with the RFC3339 format.                                                                                |
| `time.Time` ⇔ `int`, `int32`, `int64`, `uint`, `uint32`, `uint64` | Converts with the Unix timestamp.                                                                                |
| `time.Time` ⇔ `floatX`                                            | Converts with the Unix timestamp, and keeps the fractional part of a second.                                     |
| `time.Time` ⇔ `big.Int`                                           | Converts with the Unix timestamp.                                                                                |
| `time.Time` ⇔ `big.Float`                                         | Converts with the Unix timestamp, and keeps the fractional part of a second.                                     |
| `time.Time` ⇔ `int8`, `int16`, `uint8`, `uint16`                  | Not allowed.                                                                                                     |
| `time.Time` ⇔ _other_                                             | Uses `int64` as an intermediate value.                                                                           |
| `big.Int` ⇔ `bool`, `intX`, `uintX`, `floatX`                     | Converts with the `big.Int` methods, e.g. `big.Int.Int64` and `big.NewInt`.                                      |
| `big.Int` ⇔ `string`                                              | Converts with `big.Int.String` and `big.Int.SetString`. The base of the string comes from its prefix, e.g. `0x`. |
| `big.Int` ⇔ `[]byte`                                              | Converts with `big.Int.Bytes` and `big.Int.SetBytes`. A negative value is an error.                              |
| `big.Int` ⇔ `big.Float`                                           | Converts with `big.Float.Int` and `big.Float.SetInt`.                                                            |
| `big.Float` ⇔ `bool`, `intX`, `uintX`, `floatX`                   | Converts with the `big.Float` methods, e.g. `big.Float.Float64` and `big.Float.SetFloat64`.                      |
| `big.Float` ⇔ `string`                                            | Converts with `big.Float.String` and `big.Float.SetString`.                                                      |
| `big.Rat` ⇔ `string`                                              | Converts with `big.Rat.String` and `big.Rat.SetString`.                                                          |
| `big.Rat` ⇔ `slice`, `[2]array`                                   | Maps the first element to the numerator and the second to the denominator.                                       |
| `big.Rat` ⇔ `big.Float`                                           | Converts with `big.Float.SetRat` and `big.Float.Rat`.                                                            |
| `big.Rat` ⇔ _other_                                               | Uses `big.Float` as an intermediate value.                                                                       |

## Destination values

The mapper always tries to reuse the destination value. It does not change the fields, keys, or elements of the
destination that have no equivalent in the source.

Slices always get the length of the source slice. The mapper reuses the elements that exist in both slices, and sets 
the new elements to their zero value. An array cannot be resized, hence the source and the destination must have the 
same length.

Interfaces keep the type of the value that they hold. If the destination is an interface that is not empty, the
mapper maps the source to the type of the value in that interface. This is an easy way to select the type of the result:

```go
dst := []any{"", 0, new(big.Int)}
err := anymapper.Map([]string{"1", "2", "3"}, &dst)
// dst = []any{"1", 2, big.NewInt(3)}
```

The source can be a slice of interfaces too. Elements that the mapper adds to the destination slice have no type, so it
assigns them directly from the source.

If the destination is a nil pointer, a nil map, or a nil slice, the mapper creates the value before it maps the data.

## Structures

The mapper treats a structure as a key-value map. It maps every exported field of the source structure to the field of
the destination structure, or to the element of the destination map, that has the same name.

A tag changes the name of a field. The name of the tag is defined in `Context.Tag`, and is `map` by default:

```go
type Data struct {
	Foo int `map:"bar"`
	Bar int `map:"foo"`
}
```

- If the tag is `-`, the mapper always omits the field.
- If the field has no tag, the mapper uses the `Context.FieldMapper` function to change the field name. If that function
  is `nil`, it uses the field name.
- Tags can be set in the source structure, in the destination structure, or in both. If both structures have tags, the
  names in the tags must be the same.
- Fields of the destination structure that have no equivalent in the source structure stay unchanged.

## Context

A `Context` holds the settings of one mapping operation. Use it to change the behavior of the mapper without a change of
the global state.

| Field          | Default            | Description                                                          |
|----------------|--------------------|----------------------------------------------------------------------|
| `Tag`          | `map`              | Name of the structure tag.                                           |
| `ByteOrder`    | `binary.BigEndian` | Byte order for the conversions between numbers and bytes.            |
| `StrictTypes`  | `false`            | Enables [strict types](#strict-types).                               |
| `DisableCache` | `false`            | Disables the cache of the mapping functions.                         |
| `FieldMapper`  | `nil`              | Function that changes the name of a structure field that has no tag. |
| `Custom`       | `nil`              | Any value that custom mapping functions can use.                     |

## Strict types

If `Context.StrictTypes` is `true`, the source and the destination must have exactly the same type. The check also
applies to custom types, hence `type MyInt int` is no longer treated as `int`.

There are two exceptions:

- Mapping between structs and maps is always allowed.
- If the destination is an empty interface, the mapper always assigns the source value to it.

## Examples

### Mapping between simple types

```go
package main

import (
	"fmt"

	"github.com/defiweb/go-anymapper"
)

func main() {
	a := 42
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

### Mapping to interfaces

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/defiweb/go-anymapper"
)

func main() {
	// The values in the destination slice define the target types.
	b := []any{new(big.Int), 0, 0.0}

	err := anymapper.Map([]string{"1", "2", "3"}, &b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T %T %T\n", b[0], b[1], b[2]) // *big.Int int float64
}
```

### MapFrom and MapTo interfaces

```go
package main

import (
	"fmt"
	"math/big"
	"reflect"

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
	a := 42
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
	"math/big"
	"reflect"

	"github.com/defiweb/go-anymapper"
)

type Val struct {
	X *big.Int
}

func main() {
	a := 42
	var b Val

	typ := reflect.TypeOf(Val{})
	anymapper.Default.Mappers[typ] = func(m *anymapper.Mapper, src, dst reflect.Type) anymapper.MapFunc {
		if src == typ {
			return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
				return m.MapRefl(src.FieldByName("X"), dst)
			}
		}
		if dst == typ {
			return func(m *anymapper.Mapper, _ *anymapper.Context, src, dst reflect.Value) error {
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

## Benchmark

Measured on Apple M1 Max, Go 1.18, `-benchtime=500ms -count=3`.

| Case                                     | go-anymapper                  | mapstructure                    | Result           |
|------------------------------------------|-------------------------------|---------------------------------|------------------|
| `map` ⇒ `struct`                         | 1079 ns, 408 B, 16 allocs     | 2686 ns, 3580 B, 64 allocs      | **2.5x faster**  |
| `struct` ⇒ `map`                         | 969 ns, 584 B, 14 allocs      | 546 ns, 672 B, 16 allocs        | 1.8x slower      |
| `[]string` ⇒ `[]int`                     | 371 ns, 160 B, 5 allocs       | 646 ns, 377 B, 22 allocs        | **1.7x faster**  |
| `string` ⇒ `int`                         | 60 ns, 8 B, 1 alloc           | 60 ns, 96 B, 3 allocs           | even             |

<details>
<summary>Benchmark source</summary>

```go
package main

import (
	"testing"

	"github.com/defiweb/go-anymapper"
	"github.com/mitchellh/mapstructure"
)

type Object struct {
	A string
	B int
	C []string
	D []any
	E map[string]string
}

// weakDecode enables the type conversions that mapstructure does not do by
// default. go-anymapper does these conversions without configuration.
func weakDecode(input, result any) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           result,
	})
	if err != nil {
		return err
	}
	return dec.Decode(input)
}

func Benchmark(b *testing.B) {
	object := map[string]any{
		"A": "a",
		"B": 1,
		"C": []string{"a", "b", "c"},
		"D": []any{1, "2", 3.0},
		"E": map[string]string{"a": "a", "b": "b", "c": "c"},
	}
	list := []string{"1", "2", "3", "4", "5", "6", "7", "8"}

	b.Run("map-struct/anymapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result Object
			if err := anymapper.Map(object, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("map-struct/mapstructure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result Object
			if err := mapstructure.Decode(object, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("struct-map/anymapper", func(b *testing.B) {
		input := Object{
			A: "a",
			B: 1,
			C: []string{"a", "b", "c"},
			D: []any{1, "2", 3.0},
			E: map[string]string{"a": "a", "b": "b", "c": "c"},
		}
		for i := 0; i < b.N; i++ {
			var result map[string]any
			if err := anymapper.Map(input, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("struct-map/mapstructure", func(b *testing.B) {
		input := Object{
			A: "a",
			B: 1,
			C: []string{"a", "b", "c"},
			D: []any{1, "2", 3.0},
			E: map[string]string{"a": "a", "b": "b", "c": "c"},
		}
		for i := 0; i < b.N; i++ {
			var result map[string]any
			if err := mapstructure.Decode(input, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("slice/anymapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result []int
			if err := anymapper.Map(list, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("slice/mapstructure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result []int
			if err := weakDecode(list, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("scalar/anymapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result int
			if err := anymapper.Map("42", &result); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("scalar/mapstructure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result int
			if err := weakDecode("42", &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}
```

</details>

## Documentation

[https://pkg.go.dev/github.com/defiweb/go-anymapper](https://pkg.go.dev/github.com/defiweb/go-anymapper)

## License

[MIT](LICENSE)
