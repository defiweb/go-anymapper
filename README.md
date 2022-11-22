# go-anymapper

The `go-anymapper` package helps you map anything to anything! It can map data between simple Go types such as:
`string`, `int`, and data structures. It also allows you to define your own mapping functions for any type.

## Installation

```bash
go get github.com/defiweb/go-anymapper
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

Mapping will fail if the target type is not large enough to hold the source value. For example, mapping `int64`
to `int8` may fail because `int64` can store values larger than `int8`.

If the destination slice is longer than the source, the extra elements will be zeroed. If the destination slice is
shorter than the source, the mapper will append new elements to the destination slice.

When mapping numbers from a byte slice or array, the length of the slice/array *must* be the same as the size of the
variable in bytes. The size of `int`, `uint` is always considered as 64 bits.

When mapping to slices, arrays or maps, the mapper will try to reuse the existing elements of the destination
if possible. For example, mapping `[]int{1, 2}` to `[]any{"", 0}` will result in `[]any{"1", 2}`.

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

### Mapping structures

Structures are treated by mapper as key-value maps. The mapper will try to map recursively every field of the source
structure to the corresponding field of the destination structure or map.

Field names can be overridden with a tag (whose name is defined in `Mapper.Tag`, default is `map`). The tag can contain
a list of field names separated by `Mapper.Separator` (default is `.`). For example, the following struct:

```
type Foo struct {
    Bar string `map:"a.b"`
}
```

will be treated as the following map:

```
map[string]any{"a": map[string]any{"b": "bar"}}
```

As a special case, if the field tag is "-", the field is always omitted.

If the tag is not set, struct field names will be mapped using the `Mapper.FieldNameMapper` function.

Tags can be defined for both source and target structures. In this case, the names used in the tags must be the same for
both structures.

### Strict types

It is possible to enforce strict type checking by setting `Mapper.StrictTypes` to `true`. If enabled, the source and
destination types must be exactly the same for the mapping to be possible. Although, mapping between different data
structures, like `struct` ⇔ `struct`, `struct` ⇔ `map` and `map` ⇔ `map` is always possible. If the destination type is
an empty interface, the source value will be assigned to it regardless of the value of `Mapper.StrictTypes`.

The strict type check also applies to custom types, for example, `type MyInt int` will not be treated as `int` anymore.

### `MapTo` and `MapFrom` interfaces:

The `go-anymapper` package provides two interfaces that can be implemented by the source and destination types to
customize the mapping process.

If the source value implements `MapTo` interface, the `MapTo` method will be used to map the source value to the
destination value.

If the destination value implements `MapFrom` interface, the `MapFrom` method will be used to map the source value to
the destination value.

If both source and destination values implement the `MapTo` and `MapFrom` interfaces then `MapTo` will be used
first, then `MapFrom` if the first one fails.

### `Mapper.MapTo` and `Mapper.MapFrom` maps:

If it is not possible to implement the above interfaces, the custom mapping functions can be registered with the
`Mapper.MapTo` and `Mapper.MapFrom` maps. The keys of these maps are the types of the destination and source values
respectively. The values are the mapping functions.

### Default mapper instance

The package defines the default mapper instance `DefaultMapper` that is used by `Map` and `MapRefl` functions. It is
possible to change configuration of the default mapper, but it may affect other packages that use the default mapper. To
avoid this, it is recommended to create a copy of the default mapper using the `DefaultMapper.Copy` method.

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

### Nested data structures

```go
package main

import (
	"fmt"

	"github.com/defiweb/go-anymapper"
)

type Point struct {
	X, Y int
}

type Points struct {
	P1 Point
	P2 Point
}

type PointsFlat struct {
	X1 int `map:"P1.X"`
	Y1 int `map:"P1.Y"`
	X2 int `map:"P2.X"`
	Y2 int `map:"P2.Y"`
}

func main() {
	a := Points{
		P1: Point{X: 1, Y: 2},
		P2: Point{X: 3, Y: 4},
	}
	b := PointsFlat{}

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", b) // {X1:1 Y1:2 X2:3 Y2:4}
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

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.X.String()) // "42"
}
```

### MapFrom and MapTo maps

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
	anymapper.DefaultMapper.MapFrom[typ] = func(m *anymapper.Mapper, src, dst reflect.Value) error {
		x := src.FieldByName("X")
		if x.IsNil() {
			return m.MapRefl(reflect.ValueOf(0), dst)
		}
		return m.MapRefl(src.FieldByName("X"), dst)
	}
	anymapper.DefaultMapper.MapTo[typ] = func(m *anymapper.Mapper, src, dst reflect.Value) error {
		return m.MapRefl(src, dst.FieldByName("X").Addr())
	}

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.X.String()) // "42"
}
```

## Documentation

[https://pkg.go.dev/github.com/defiweb/go-anymapper](https://pkg.go.dev/github.com/defiweb/go-anymapper)
