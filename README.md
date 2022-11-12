# go-anymapper

The `go-anymapper` package helps you map anything to anything! It can map data between simple Go types such as:
`string`, `int`, and data structures. It also allows you to define your own mapping functions for any type.

## Installation

```bash
go get github.com/ethereum/go-anymapper
```

## Usage

The simplest way to use the `go-anymapper` package is to use the `Map` function. It takes two arguments: the source and
the destination. The function will try to map the source to the destination. If the both arguments are of different
types, the function will try to convert types using to following rules:

- If the dest value is an empty interface, the src value is assigned to it.
- `bool` ⇔ `intX`, `uintX`, `floatX` ⇒ `true` ⇔ `1`, `false` ⇔ `0` (if source is number, then `≠0` ⇒ `true`).
- `intX`, `uintX`, `floatX` ⇔ `intX`, `uintX`, `floatX` ⇒ cast numbers to the destination type if not overflow.
- `string` ⇔ `intX`, `uintX` ⇒ converts string to or from number using `big.Int.SetString` and `big.Int.String`.
- `string` ⇔ `floatX` ⇒ converts string to or from number using `big.Float.SetString` and `big.Float.String`.
- `string` ⇔ `[]byte` ⇒ converts string to or from byte array using `[]byte(s)` and `string(b)`.
- `[]byte` ⇔ `intX`, `uintX` ⇒ converts byte slice to or from number using `big.Int.SetBytes` and `big.Int.Bytes`.
- `[X]byte` ⇔ `intX`, `uintX` ⇒ converts byte array to or from number using `big.Int.SetBytes` and `big.Int.Bytes`.
- `slice` ⇔ `slice` ⇒ recursively map each slice element.
- `slice` ⇔ `array` ⇒ recursively map each slice element if lengths are the same.
- `map` ⇔ `map` ⇒ recursively map every key and value pair.
- `struct` ⇔ `struct` ⇒ recursively map every struct field.
- `struct` ⇔ `map[string]X` ⇒ map struct fields to map elements using field names as keys and vice versa.

Please note, that mapping may fail if the destination type it not large enough to hold the source value. For example,
mapping `int64` to `int8` may fail because `int64` can hold values larger than `int8`. This is also true for mapping
numbers to byte slices or arrays.

The above types refers to type kind, not to the actual type, hence `type MyInt int` is considered to be `int` as well.

In addition to the above rules, the default configuration of the mapper supports the following conversions:

- `time.Time` ⇔ `string` ⇒ converts string to or from time using RFC3339 format.
- `time.Time` ⇔  `uint`, `uint32`, `uint64`, `int`, `int32`, `int64` ⇒ converts time to or from Unix timestamp.
- `time.Time` ⇔  `uint8`, `uint16`, `int8`, `int16` ⇒ not allowed.
- `time.Time` ⇔  `floatX` ⇒ convert to or from unix timestamp, preserving the fractional part of a second.
- `time.Time` ⇔  `big.Int` ⇒ convert to or from unix timestamp
- `time.Time` ⇔  `big.Float` ⇒ convert to or from unix timestamp, preserving the fractional part of a second.
- `time.Time` ⇔  `X` ⇒ if none of the above rules apply, `time.Time` will be treated as `int64`.
- `big.Int` ⇔ `intX`, `uintX`, `floatX` ⇒ converts big integer to or from number if not overflow.
- `big.Int` ⇔ `string` ⇒ converts to or from string using `big.Int.String` and `big.Int.SetString`.
- `big.Int` ⇔ `[]byte` ⇒ converts to or from byte array using `big.Int.Bytes` and `big.Int.SetBytes`.
- `big.Int` ⇔ `[X]byte` ⇒ converts to or from byte array using `big.Int.Bytes` and `big.Int.SetBytes` if not overflow.
- `big.Int` ⇔ `big.Float` ⇒ coverts number to the destination type (float numbers are rounded down).
- `big.Float` ⇔ `intX`, `uintX`, `floatX` ⇒ converts big float to or from number if not overflow.
- `big.Float` ⇔ `string` ⇒ converts to or from string using `big.Float.String` and `big.Float.SetString`.

### Mapping structures

Structures are treated by mapper as key-value maps. The mapper will try to map recursively every field of the source
structure to the corresponding field of the destination structure or map.

Field names can be overridden with a tag (whose name is defined in `Mapper.Tag`, default is `map`). The tag can contain
a list of field names separated by `Mapper.Separator` (default is `.`). In this case, the field will be treated as a
nested field. For example, the following struct:

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

### `MapInto` and `MapFrom` interfaces:

The `go-anymapper` package provides two interfaces that can be implemented by the source and destination types to
customize the mapping process.

If the source value implements `MapInto` interface, the `MapInto` method will be used to map the source value to the
destination value.

If the destination value implements `MapFrom` interface, the `MapFrom` method will be used to map the source value to
the destination value.

If both source and destination values implement the `MapInto` and `MapFrom` interfaces then `MapInto` will be used
first, if it returns an error then `MapFrom`.

### `Mapper.MapInro` and `Mapper.MapFrom` maps:

If it is not possible to implement the above interfaces, the custom mapping functions can be registered with the
`Mapper.MapInto` and `Mapper.MapFrom` maps. The keys of these maps are the types of the destination and source values
respectively. The values are the mapping functions.

### Default mapper instance

The package defines the default mapper instance `DefaultMapper` that can be used to map values. The functions `Map`
and `MapRefl`are wrappers around the `DefaultMapper.Map` and `DefaultMapper.MapRefl` methods. It is possible to
change configuration of the default mapper but it may affect other packages that use the default mapper. To avoid
this, it is recommended to create a copy of the default mapper using the `DefaultMapper.Copy` method.

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

	fmt.Println(b) // { X1: 1, Y1: 2, X2: 3, Y2: 4 }
}
```

### MapFrom and MapInto interfaces

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

func (v *Val) MapFrom(m *anymapper.Mapper, x any) error {
	return m.Map(x, &v.X)
}

func (v *Val) MapInto(m *anymapper.Mapper, x any) error {
	return m.Map(v.X, x)
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

### MapFrom and MapInto maps

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
	anymapper.DefaultMapper.MapFrom[typ] = func(m *anymapper.Mapper, src, dest reflect.Value) error {
		return m.MapRefl(src, src.FieldByName("X").Addr())
	}
	anymapper.DefaultMapper.MapInto[typ] = func(m *anymapper.Mapper, src, dest reflect.Value) error {
		return m.Map(src.FieldByName("X"), dest)
	}

	err := anymapper.Map(a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Println(b.X.String()) // "42"
}
```

## Documentation

[https:pkg.go.dev/github.com/defiweb/go-anymapper](https:pkg.go.dev/github.com/defiweb/go-anymapper)
