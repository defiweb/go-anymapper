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
// It is configured to use the "map" struct tag, the "." separator. It also
// provides additional mapping rules for time.Time, big.Int, big.Float and
// big.Rat. It can be modified to change the default behavior, but if the
// mapper is used by other packages, it is recommended to create a copy of the
// default mapper and modify the copy.
var DefaultMapper = &Mapper{
	Tag:       `map`,
	Separator: `.`,
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

	// Separator is the symbol that is used to separate fields in the
	// struct tag.
	Separator string

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

func (m *Mapper) Map(src, dst any) error {
	return m.MapRefl(reflect.ValueOf(src), reflect.ValueOf(dst))
}

func (m *Mapper) MapRefl(src, dst reflect.Value) error {
	src = m.srcValue(src)
	dst = m.dstValue(dst)
	if !src.IsValid() {
		return InvalidSrcErr
	}
	if !dst.IsValid() {
		return InvalidDstErr
	}
	if dst.Type() == anyTy && dst.CanSet() {
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

// Copy creates a copy of the current Mapper with the same configuration.
func (m *Mapper) Copy() *Mapper {
	cpy := &Mapper{
		Tag:         m.Tag,
		Separator:   m.Separator,
		FieldMapper: m.FieldMapper,
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

// mapFunc tries to map the source value to the destination value using the
// MapFrom and MapTo interfaces, and the MapFrom and MapTo maps.
//
// It tries to use every defined mapping function until one of them succeeds.
// If no mapping function succeeds, it returns an error from the last mapping
// function that was tried.
func (m *Mapper) mapFunc(src, dst reflect.Value) (ok bool, err error) {
	if src.Type().Implements(mapToTy) {
		if err = src.Interface().(MapTo).MapTo(m, dst); err == nil {
			return true, nil
		}
	}
	if dst.Type().Implements(mapFromTy) {
		if err = dst.Interface().(MapFrom).MapFrom(m, src); err == nil {
			return true, nil
		}
	}
	if f, ok := m.MapTo[dst.Type()]; ok {
		if err = f(m, src, dst); err == nil {
			return true, nil
		}
	}
	if f, ok := m.MapFrom[src.Type()]; ok {
		if err = f(m, src, dst); err == nil {
			return true, nil
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
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dst.Index(i)); err != nil {
				return err
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
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dst.Index(i)); err != nil {
				return err
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
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dst.Index(i)); err != nil {
				return err
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
			if err := m.MapRefl(src.Index(i), dst.Index(i)); err != nil {
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
		if err := m.MapRefl(src, m.structToPtrsMap(dst, true)); err != nil {
			return err
		}
		return nil
	case reflect.Map:
		dstKeyType := dst.Type().Key()
		for _, srcKey := range src.MapKeys() {
			// Map key.
			dstKey := srcKey
			if srcKey.Type() != dstKeyType {
				dstKey = reflect.New(dstKeyType).Elem()
				if err := m.MapRefl(srcKey, dstKey); err != nil {
					return NewInvalidMappingError(srcKey.Type(), dstKeyType, "unable to map key")
				}
			}
			// It is important here to use dstValue because we need to check
			// if the value can be set directly or if we need to create a new
			// value, the dstValue function will always return a value that
			// can be set, otherwise it will return an invalid value.
			dstVal := m.dstValue(dst.MapIndex(dstKey))
			if dstVal.IsValid() {
				if err := m.MapRefl(src.MapIndex(srcKey), dstVal); err != nil {
					return err
				}
			} else {
				v := reflect.New(dst.Type().Elem()).Elem()
				if err := m.MapRefl(src.MapIndex(srcKey), v); err != nil {
					return err
				}
				dst.SetMapIndex(dstKey, v)
			}
		}
		return nil
	}
	return fmt.Errorf("mapper: cannot map map to %v", dst.Type())
}

func (m *Mapper) mapStruct(src, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Struct:
		if src.Type() == dst.Type() {
			dst.Set(src)
			return nil
		}
		return m.MapRefl(m.structToPtrsMap(src, false), m.structToPtrsMap(dst, true))
	case reflect.Map:
		if dst.Type().Key().Kind() != reflect.String {
			return NewInvalidMappingError(src.Type(), dst.Type(), "map key must be string")
		}
		return m.MapRefl(m.structToPtrsMap(src, false), dst)
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

// structToPtrsMap converts a struct to a map where the keys are the field
// names and the values are pointers to the fields. If a struct field has a
// tag, it is used as the key. If the tag has a nested field (e.g. "foo.bar"),
// the resulting map will have a nested map (e.g. "foo" => "bar" => &field).
func (m *Mapper) structToPtrsMap(v reflect.Value, initialize bool) reflect.Value {
	r := make(map[string]any)
	t := v.Type()
	for idx := 0; idx < v.NumField(); idx++ {
		vField := v.Field(idx)
		tField := t.Field(idx)
		if initialize {
			// The value needs to be initialized here to make sure that
			// the value is addressable, so that it will be possible to
			// store in a map a pointer to it. The dstValue method will
			// do initialization if needed.
			vField = m.dstValue(vField)
		}
		var fields []string
		if tag, ok := tField.Tag.Lookup(m.Tag); ok {
			if tag == "-" {
				continue
			}
			if len(m.Separator) > 0 {
				fields = strings.Split(tag, m.Separator)
			} else {
				fields = []string{tag}
			}
		}
		if len(fields) == 0 {
			if m.FieldMapper != nil {
				fields = []string{m.FieldMapper(tField.Name)}
			} else {
				fields = []string{tField.Name}
			}
		}
		e := r
		for i, f := range fields {
			if i == len(fields)-1 {
				e[f] = addr(vField).Interface()
				break
			}
			if e[f] == nil {
				e[f] = make(map[string]any)
			}
			e = e[f].(map[string]any)
		}
	}
	return reflect.ValueOf(r)
}

// srcValue unpacks values from pointers and interfaces until it reaches a non-pointer,
// non-interface value or value that implements the MapFrom interface or a type that
// has a custom mapper.
func (m *Mapper) srcValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.Type().Implements(mapFromTy) {
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
		if v.Kind() == reflect.Pointer && v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Kind() == reflect.Map && v.IsNil() && v.CanSet() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		if v.Kind() == reflect.Slice && v.IsNil() && v.CanSet() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		if v.Type().Implements(mapToTy) {
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
)
