package anymapper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// MapTo interface is implemented by types that can map themselves to
// another type.
type MapTo interface {
	// MapTo maps the receiver value to the destination value.
	MapTo(m *Mapper, dst reflect.Value) error
}

// MapFrom interface is implemented by types that can set their value from
// another type.
type MapFrom interface {
	// MapFrom sets the receiver value from the source value.
	MapFrom(m *Mapper, src reflect.Value) error
}

// DefaultMapper is the default Mapper used by the Map and MapRefl functions.
// It also provides additional mapping rules for time.Time, big.Int, big.Float
// and big.Rat. It can be modified to change the default behavior, but if the
// mapper is used by other packages, it is recommended to create a copy of the
// default mapper and modify the copy.
var DefaultMapper = &Mapper{
	Tag:       `map`,
	ByteOrder: binary.BigEndian,
	MapFrom: map[reflect.Type]MapFunc{
		timeTy:     mapTimeSrc,
		bigIntTy:   mapBigIntSrc,
		bigFloatTy: mapBigFloatSrc,
		bigRatTy:   mapBigRatSrc,
	},
	MapTo: map[reflect.Type]MapFunc{
		timeTy:     mapTimeDst,
		bigIntTy:   mapBigIntDst,
		bigFloatTy: mapBigFloatDst,
		bigRatTy:   mapBigRatDst,
	},
}

// MapFunc is a function that maps a src value to a dst value. It returns an
// error if the mapping is not possible. The src and dst values are never
// pointers.
type MapFunc func(m *Mapper, src, dst reflect.Value) error

// Mapper hold the mapper configuration.
type Mapper struct {
	// StrictTypes enabled strict type checking.
	StrictTypes bool

	// Tag is the name of the struct tag that is used by the mapper to
	// determine the name of the field to map to.
	Tag string

	// FieldMapper is a function that maps a struct field name to another name,
	// it is used only when the tag is not present.
	FieldMapper func(string) string

	// ByteOrder is the byte order used to map data to and from byte slices.
	ByteOrder binary.ByteOrder

	// MapTo is a map of types that can map themselves to another type.
	MapTo map[reflect.Type]MapFunc

	// MapFrom is a map of types that can map themselves from another type.
	MapFrom map[reflect.Type]MapFunc
}

// Map maps the source value to the destination value.
//
// It is shorthand for DefaultMapper.Map(src, dst).
func Map(src, dst any) error {
	return DefaultMapper.Map(src, dst)
}

// MapRefl maps the source value to the destination value.
//
// It is shorthand for DefaultMapper.MapRefl(src, dst).
func MapRefl(src, dst reflect.Value) error {
	return DefaultMapper.MapRefl(src, dst)
}

// Map maps the source value to the destination value.
func (m *Mapper) Map(src, dst any) error {
	return m.MapRefl(reflect.ValueOf(src), reflect.ValueOf(dst))
}

// MapRefl maps the source value to the destination value.
func (m *Mapper) MapRefl(src, dst reflect.Value) error {
	return m.mapRefl(m.srcValue(src), m.dstValue(dst))
}

// Copy creates a copy of the current Mapper with the same configuration.
func (m *Mapper) Copy() *Mapper {
	cpy := &Mapper{
		Tag:         m.Tag,
		FieldMapper: m.FieldMapper,
		ByteOrder:   m.ByteOrder,
	}
	if m.MapFrom != nil {
		cpy.MapFrom = make(map[reflect.Type]MapFunc)
		for k, v := range m.MapFrom {
			cpy.MapFrom[k] = v
		}
	}
	if m.MapTo != nil {
		cpy.MapTo = make(map[reflect.Type]MapFunc)
		for k, v := range m.MapTo {
			cpy.MapTo[k] = v
		}
	}
	return cpy
}

func (m *Mapper) mapRefl(src, dst reflect.Value) error {
	if !src.IsValid() {
		return InvalidSrcErr
	}
	if !dst.IsValid() {
		return InvalidDstErr
	}
	if canSetDirectly(src.Type(), dst.Type()) && dst.CanSet() {
		dst.Set(src)
		return nil
	}
	if ok, err := m.mapFunc(src, dst); ok {
		return err
	}
	switch src.Kind() {
	case reflect.Bool:
		return m.mapBool(src, dst)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m.mapInt(src, dst)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return m.mapUint(src, dst)
	case reflect.Float32, reflect.Float64:
		return m.mapFloat(src, dst)
	case reflect.String:
		return m.mapString(src, dst)
	case reflect.Slice:
		return m.mapSlice(src, dst)
	case reflect.Array:
		return m.mapArray(src, dst)
	case reflect.Map:
		return m.mapMap(src, dst)
	case reflect.Struct:
		return m.mapStruct(src, dst)
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

// mapFunc tries to map the source value to the destination value using the
// MapFrom and MapTo interfaces, and the MapFrom and MapTo maps.
//
// It tries to use every defined mapping function until one of them succeeds.
// If no mapping function succeeds, it returns an error from the last mapping
// function that was tried.
func (m *Mapper) mapFunc(src, dst reflect.Value) (ok bool, err error) {
	isSrcSimpleType := isSimpleType(src.Type())
	isDstSimpleType := isSimpleType(dst.Type())
	if !isSrcSimpleType {
		mapTo, ok := src.Interface().(MapTo)
		if ok {
			if err = mapTo.MapTo(m, dst); err == nil {
				return true, nil
			}
		}
	}
	if !isDstSimpleType {
		mapFrom, ok := dst.Interface().(MapFrom)
		if ok {
			if err = mapFrom.MapFrom(m, src); err == nil {
				return true, nil
			}
		}
	}
	if !isDstSimpleType {
		if f, ok := m.MapTo[dst.Type()]; ok {
			if err = f(m, src, dst); err == nil {
				return true, nil
			}
		}
	}
	if !isSrcSimpleType {
		if f, ok := m.MapFrom[src.Type()]; ok {
			if err = f(m, src, dst); err == nil {
				return true, nil
			}
		}
	}
	if err != nil {
		// If the error is not nil, it means that there was a mapping rule
		// defined, but it failed.
		return true, err
	}
	return false, err
}

func (m *Mapper) mapBool(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src.Bool() {
			dst.SetInt(1)
		} else {
			dst.SetInt(0)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src.Bool() {
			dst.SetUint(1)
		} else {
			dst.SetUint(0)
		}
		return nil
	case reflect.Float32, reflect.Float64:
		if src.Bool() {
			dst.SetFloat(1)
		} else {
			dst.SetFloat(0)
		}
		return nil
	case reflect.String:
		if src.Bool() {
			dst.SetString("true")
		} else {
			dst.SetString("false")
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapInt(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src.Int() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dst.OverflowInt(src.Int()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(src.Int())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := src.Int()
		if n < 0 || dst.OverflowUint(uint64(n)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(uint64(n))
		return nil
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(float64(src.Int()))
		return nil
	case reflect.String:
		dst.SetString(strconv.FormatInt(src.Int(), 10))
		return nil
	case reflect.Slice, reflect.Array:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			return m.toBytes(src, dst)
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapUint(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src.Uint() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := src.Uint()
		if n > uint64(math.MaxInt64) || dst.OverflowInt(int64(n)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(int64(n))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if dst.OverflowUint(src.Uint()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(src.Uint())
		return nil
	case reflect.Float32, reflect.Float64:
		dst.SetFloat(float64(src.Uint()))
		return nil
	case reflect.String:
		dst.SetString(strconv.FormatUint(src.Uint(), 10))
		return nil
	case reflect.Slice, reflect.Array:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			return m.toBytes(src, dst)
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapFloat(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src.Float() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := src.Float()
		if n < math.MinInt64 || n > math.MaxInt64 || dst.OverflowInt(int64(n)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(int64(n))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := src.Float()
		if n < 0 || n > math.MaxUint64 || dst.OverflowUint(uint64(n)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(uint64(n))
		return nil
	case reflect.Float32, reflect.Float64:
		n := src.Float()
		if dst.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetFloat(n)
		return nil
	case reflect.String:
		dst.SetString(strconv.FormatFloat(src.Float(), 'f', -1, 64))
		return nil
	case reflect.Slice, reflect.Array:
		return m.toBytes(src, dst)
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapString(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Bool:
		switch src.String() {
		case "true":
			dst.SetBool(true)
			return nil
		case "false":
			dst.SetBool(false)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		if !n.IsInt64() || dst.OverflowInt(n.Int64()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(n.Int64())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		if !n.IsUint64() || dst.OverflowUint(n.Uint64()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(n.Uint64())
		return nil
	case reflect.Float32, reflect.Float64:
		bn, ok := new(big.Float).SetString(src.String())
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		n, a := bn.Float64()
		if dst.OverflowFloat(n) || (math.IsInf(n, 0) && (a == big.Below || a == big.Above)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetFloat(n)
		return nil
	case reflect.String:
		dst.SetString(src.String())
		return nil
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			dst.SetBytes([]byte(src.String()))
			return nil
		}
	case reflect.Array:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			b := []byte(src.String())
			if len(b) != dst.Len() {
				return NewInvalidMappingError(src.Type(), dst.Type(), "length mismatch")
			}
			for i := 0; i < len(b); i++ {
				dst.Index(i).SetUint(uint64(b[i]))
			}
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapSlice(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() && dst.Kind() != reflect.Map {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			return m.fromBytes(src.Bytes(), dst)
		}
	case reflect.String:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dst.SetString(string(src.Bytes()))
			return nil
		}
	case reflect.Slice:
		// Instead of creating a new slice, we reuse the existing one and only
		// adjust the length. That way the mapper will map values to the existing
		// elements, instead of creating new ones. It may be especially useful
		// when dst is a slice of interfaces.
		if src.Len() > dst.Len() {
			dst.Set(reflect.AppendSlice(
				dst,
				reflect.MakeSlice(dst.Type(), src.Len()-dst.Len(), src.Len()-dst.Len())),
			)
		}
		if canSetDirectly(src.Type(), dst.Type()) {
			dst.Set(src)
		} else {
			for i := 0; i < src.Len(); i++ {
				if err := m.mapRefl(m.srcValue(src.Index(i)), m.dstValue(dst.Index(i))); err != nil {
					return err
				}
			}
		}
		// If the source slice is shorter than the destination, we need to
		// zero out the remaining elements.
		for i := src.Len(); i < dst.Len(); i++ {
			dst.Index(i).Set(reflect.Zero(dst.Type().Elem()))
		}
		return nil
	case reflect.Array:
		if src.Len() != dst.Len() {
			return NewInvalidMappingError(src.Type(), dst.Type(), "length mismatch")
		}
		if canSetDirectly(src.Type().Elem(), dst.Type().Elem()) {
			reflect.Copy(dst, src)
		} else {
			for i := 0; i < src.Len(); i++ {
				if err := m.mapRefl(m.srcValue(src.Index(i)), m.dstValue(dst.Index(i))); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapArray(src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, src.Len())
			for i := 0; i < src.Len(); i++ {
				b[i] = byte(src.Index(i).Uint())
			}
			return m.fromBytes(b, dst)
		}
	case reflect.String:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, src.Len())
			for i := 0; i < src.Len(); i++ {
				b[i] = byte(src.Index(i).Uint())
			}
			dst.SetString(string(b))
			return nil
		}
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
		if canSetDirectly(src.Type().Elem(), dst.Type().Elem()) {
			reflect.Copy(dst, src)
		} else {
			for i := 0; i < src.Len(); i++ {
				if err := m.mapRefl(m.srcValue(src.Index(i)), m.dstValue(dst.Index(i))); err != nil {
					return err
				}
			}
		}
		return nil
	case reflect.Array:
		if src.Type() == dst.Type() {
			dst.Set(src)
			return nil
		}
		if src.Len() != dst.Len() {
			return NewInvalidMappingError(src.Type(), dst.Type(), "length mismatch")
		}
		for i := 0; i < src.Len(); i++ {
			if err := m.mapRefl(m.srcValue(src.Index(i)), m.dstValue(dst.Index(i))); err != nil {
				return err
			}
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

func (m *Mapper) mapMap(src, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Struct:
		dstNum := dst.Type().NumField()
		for i := 0; i < dstNum; i++ {
			dstFieldVal := dst.Field(i)
			dstFieldTyp := dst.Type().Field(i)
			tag, skip := m.parseTag(dstFieldTyp)
			if skip {
				continue
			}
			key := reflect.ValueOf(tag)
			val := m.srcValue(src.MapIndex(key))
			if canSetDirectly(val.Type(), dstFieldTyp.Type) {
				dstFieldVal.Set(val)
			} else {
				aux := reflect.New(dstFieldTyp.Type).Elem()
				if err := m.mapRefl(val, aux); err != nil {
					return err
				}
				dstFieldVal.Set(aux)
			}
		}
		return nil
	case reflect.Map:
		srcKeyTyp := src.Type().Key()
		dstKeyTyp := dst.Type().Key()
		srcElemTyp := src.Type().Elem()
		dstElemTyp := dst.Type().Elem()
		sameKeys := srcKeyTyp == dstKeyTyp
		canSetDir := canSetDirectly(srcElemTyp, dstElemTyp)
		for _, srcKey := range src.MapKeys() {
			dstKey := srcKey
			if !sameKeys {
				dstKey = reflect.New(dstKeyTyp).Elem()
				if err := m.mapRefl(m.srcValue(srcKey), m.dstValue(dstKey)); err != nil {
					return NewInvalidMappingError(srcKey.Type(), dstKeyTyp, "unable to map key")
				}
			}
			if canSetDir {
				dst.SetMapIndex(dstKey, src.MapIndex(srcKey))
			} else {
				// It is important here to use dstValue because we need to check
				// if the value can be set directly or if we need to create a new
				// value, the dstValue function will always return a value that
				// can be set, otherwise it will return an invalid value.
				dstVal := m.dstValue(dst.MapIndex(dstKey))
				if dstVal.IsValid() {
					if err := m.mapRefl(m.srcValue(src.MapIndex(srcKey)), dstVal); err != nil {
						return err
					}
				} else {
					v := reflect.New(dstElemTyp).Elem()
					if err := m.mapRefl(m.srcValue(src.MapIndex(srcKey)), m.dstValue(v)); err != nil {
						return err
					}
					dst.SetMapIndex(dstKey, v)
				}
			}
		}
		return nil
	}
	return fmt.Errorf("mapper: cannot map map to %v", dst.Type())
}

func (m *Mapper) mapStruct(src, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Struct:
		srcTyp := src.Type()
		dstTyp := dst.Type()
		if srcTyp == dstTyp {
			n := src.NumField()
			for i := 0; i < n; i++ {
				srcVal := src.Field(i)
				dstVal := dst.Field(i)
				if _, skip := m.parseTag(srcTyp.Field(i)); skip {
					continue
				}
				if err := m.mapRefl(m.srcValue(srcVal), m.dstValue(dstVal)); err != nil {
					return err
				}
			}
			return nil
		}
		srcNum := srcTyp.NumField()
		dstNum := dstTyp.NumField()
		srcVals := map[string]reflect.Value{}
		for i := 0; i < srcNum; i++ {
			val := src.Field(i)
			typ := srcTyp.Field(i)
			tag, skip := m.parseTag(typ)
			if skip {
				continue
			}
			srcVals[tag] = m.srcValue(val)
		}
		for i := 0; i < dstNum; i++ {
			val := dst.Field(i)
			typ := dstTyp.Field(i)
			tag, skip := m.parseTag(typ)

			if skip {
				continue
			}
			if srcVal, ok := srcVals[tag]; ok {
				if canSetDirectly(srcVal.Type(), typ.Type) {
					val.Set(srcVal)
				} else {
					if err := m.mapRefl(srcVal, m.dstValue(val)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	case reflect.Map:
		if dst.Type().Key().Kind() != reflect.String {
			return NewInvalidMappingError(src.Type(), dst.Type(), "map key must be string")
		}
		dstTyp := dst.Type().Elem()
		srcNum := src.Type().NumField()
		for i := 0; i < srcNum; i++ {
			srcFieldVal := src.Field(i)
			srcFieldTyp := src.Type().Field(i)
			tag, skip := m.parseTag(srcFieldTyp)
			if skip {
				continue
			}
			key := reflect.ValueOf(tag)
			if canSetDirectly(srcFieldTyp.Type, dstTyp) {
				dst.SetMapIndex(key, srcFieldVal)
			} else {
				val := m.dstValue(dst.MapIndex(key))
				if val.IsValid() {
					if err := m.mapRefl(m.srcValue(srcFieldVal), val); err != nil {
						return err
					}
				} else {
					aux := reflect.New(dstTyp).Elem()
					if err := m.mapRefl(m.srcValue(srcFieldVal), aux); err != nil {
						return err
					}
					dst.SetMapIndex(key, aux)
				}
			}
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

// srcValue unpacks values from pointers and interfaces until it reaches a non-pointer,
// non-interface value or value that implements the MapFrom interface or a type that
// has a custom mapper.
func (m *Mapper) srcValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if isSimpleType(v.Type()) {
			return v
		}
		if _, ok := v.Interface().(MapFrom); ok {
			return v
		}
		v = v.Elem()
	}
	return v
}

// dstValue unpacks values from pointers and interfaces until it reaches a
// settable non-pointer or non-interface value, value that implements the
// MapTo interface, has a custom mapper, or a value that is a map, slice or
// array. It returns an invalid value if it cannot find a value that meets
// these conditions. If the value is a pointer, map or slice, it will be
// initialized if needed.
func (m *Mapper) dstValue(v reflect.Value) reflect.Value {
	settable := reflect.Value{}
	for {
		if !v.IsValid() {
			break
		}
		m.initValue(v)
		if v.CanSet() && isSimpleType(v.Type()) {
			return v
		}
		if _, ok := v.Interface().(MapTo); ok {
			return v
		}
		if m.MapTo[v.Type()] != nil {
			return v
		}
		if v.Kind() == reflect.Map && !v.IsNil() {
			return v
		}
		if v.CanSet() {
			settable = v
		}
		if v.Kind() != reflect.Interface && v.Kind() != reflect.Pointer {
			break
		}
		v = v.Elem()
	}
	return settable
}

// initValue initializes a value if it is a pointer, map or slice.
func (m *Mapper) initValue(v reflect.Value) {
	if v.Kind() < reflect.Map || v.Kind() > reflect.Slice || !v.IsNil() || !v.CanSet() {
		return
	}
	switch {
	case v.Kind() == reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
	case v.Kind() == reflect.Map:
		v.Set(reflect.MakeMap(v.Type()))
	case v.Kind() == reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}
}

// parseTag parses the tag of the given field and returns the tag name and
// whether the field should be skipped.
func (m *Mapper) parseTag(f reflect.StructField) (fields string, skip bool) {
	tag, ok := f.Tag.Lookup(m.Tag)
	if !ok {
		if m.FieldMapper != nil {
			return m.FieldMapper(f.Name), false
		} else {
			return f.Name, false
		}
	}
	if tag == "-" {
		return "", true
	}
	return tag, false
}

// toBytes converts a value to a byte slice using binary.Write.
func (m *Mapper) toBytes(src, dst reflect.Value) error {
	// binary.Write does not work with Int and Uint types, so we need to
	// convert them to int64 and uint64. To make mapped values compatible
	// between 32 and 64-bit architectures, we always use int64 and uint64.
	switch src.Kind() {
	case reflect.Int:
		src = reflect.ValueOf(src.Int())
	case reflect.Uint:
		src = reflect.ValueOf(src.Uint())
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, m.ByteOrder, src.Interface()); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
	}
	switch dst.Kind() {
	case reflect.Slice:
		if dst.Type().Elem().Kind() != reflect.Uint8 {
			return NewInvalidMappingError(src.Type(), dst.Type(), "")
		}
		dst.SetBytes(buf.Bytes())
	case reflect.Array:
		if dst.Type().Elem().Kind() != reflect.Uint8 {
			return NewInvalidMappingError(src.Type(), dst.Type(), "")
		}
		if dst.Len() != buf.Len() {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid array length")
		}
		reflect.Copy(dst, reflect.ValueOf(buf.Bytes()))
	default:
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

// fromBytes converts a byte slice to a value using binary.Read.
func (m *Mapper) fromBytes(src []byte, dst reflect.Value) error {
	if len(src) != int(dst.Type().Size()) {
		return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "invalid byte slice length")
	}
	switch dst.Kind() {
	case reflect.Int:
		var v int64
		if err := binary.Read(bytes.NewReader(src), m.ByteOrder, &v); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
		if dst.OverflowInt(v) {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "overflow")
		}
		dst.SetInt(v)
	case reflect.Uint:
		var v uint64
		if err := binary.Read(bytes.NewReader(src), m.ByteOrder, &v); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
		if dst.OverflowUint(v) {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "overflow")
		}
		dst.SetUint(v)
	default:
		if err := binary.Read(bytes.NewBuffer(src), m.ByteOrder, dst.Addr().Interface()); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
	}
	return nil
}

// addr returns the address of the value if it is addressable and it not a
// pointer already. Otherwise, it returns the value itself.
func addr(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer || !v.CanAddr() {
		return v
	}
	return v.Addr()
}

// canSetDirectly reports whether a src value can be set directly to a dst.
func canSetDirectly(src, dst reflect.Type) bool {
	return (src == dst || dst == anyTy) && isSimpleType(src)
}

// isSimpleType returns true if the type is a simple Go type that do not
// implement any interfaces.
func isSimpleType(p reflect.Type) bool {
	switch p.Kind() {
	case reflect.Bool:
		return p == boolTy
	case reflect.Int:
		return p == intTy
	case reflect.Int8:
		return p == int8Ty
	case reflect.Int16:
		return p == int16Ty
	case reflect.Int32:
		return p == int32Ty
	case reflect.Int64:
		return p == int64Ty
	case reflect.Uint:
		return p == uintTy
	case reflect.Uint8:
		return p == uint8Ty
	case reflect.Uint16:
		return p == uint16Ty
	case reflect.Uint32:
		return p == uint32Ty
	case reflect.Uint64:
		return p == uint64Ty
	case reflect.Float32:
		return p == float32Ty
	case reflect.Float64:
		return p == float64Ty
	case reflect.String:
		return p == stringTy
	case reflect.Slice:
		return strings.HasPrefix(p.String(), "[")
	case reflect.Array:
		return strings.HasPrefix(p.String(), "[")
	case reflect.Map:
		return strings.HasPrefix(p.String(), "map[")
	case reflect.Struct:
		return strings.HasPrefix(p.String(), "struct {")
	}
	return false
}

// InvalidSrcErr is returned when reflect.IsValid returns false for the source
// value.
var InvalidSrcErr = errors.New("mapper: invalid source value")

// InvalidDstErr is returned when reflect.IsValid returns false for the
// destination value. It may happen when the destination value was not
// passed as a pointer.
var InvalidDstErr = errors.New("mapper: invalid destination value")

type InvalidMappingErr struct {
	From, To reflect.Type
	Reason   string
}

func NewInvalidMappingError(from, to reflect.Type, reason string) *InvalidMappingErr {
	return &InvalidMappingErr{From: from, To: to, Reason: reason}
}

func (e *InvalidMappingErr) Error() string {
	if len(e.Reason) == 0 {
		return fmt.Sprintf("mapper: cannot map %v to %v", e.From, e.To)
	}
	return fmt.Sprintf("mapper: cannot map %v to %v: %s", e.From, e.To, e.Reason)
}

var (
	mapFromTy = reflect.TypeOf((*MapFrom)(nil)).Elem()
	mapToTy   = reflect.TypeOf((*MapTo)(nil)).Elem()
	anyTy     = reflect.TypeOf((*any)(nil)).Elem()

	// Common types
	boolTy    = reflect.TypeOf((*bool)(nil)).Elem()
	intTy     = reflect.TypeOf((*int)(nil)).Elem()
	int8Ty    = reflect.TypeOf((*int8)(nil)).Elem()
	int16Ty   = reflect.TypeOf((*int16)(nil)).Elem()
	int32Ty   = reflect.TypeOf((*int32)(nil)).Elem()
	int64Ty   = reflect.TypeOf((*int64)(nil)).Elem()
	uintTy    = reflect.TypeOf((*uint)(nil)).Elem()
	uint8Ty   = reflect.TypeOf((*uint8)(nil)).Elem()
	uint16Ty  = reflect.TypeOf((*uint16)(nil)).Elem()
	uint32Ty  = reflect.TypeOf((*uint32)(nil)).Elem()
	uint64Ty  = reflect.TypeOf((*uint64)(nil)).Elem()
	float32Ty = reflect.TypeOf((*float32)(nil)).Elem()
	float64Ty = reflect.TypeOf((*float64)(nil)).Elem()
	stringTy  = reflect.TypeOf((*string)(nil)).Elem()
)
