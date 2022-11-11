package anymapper

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// MapInto interface is implemented by types that can map themselves to
// another type.
type MapInto interface {
	// MapInto maps the receiver value to the destination value.
	MapInto(m *Mapper, dest reflect.Value) error
}

// MapFrom interface is implemented by types that can set their value from
// another type.
type MapFrom interface {
	// MapFrom sets the receiver value from the source value.
	MapFrom(m *Mapper, src reflect.Value) error
}

// DefaultMapper is the default Mapper used by the Map and MapRefl functions.
// It is configured to use the "map" struct tag, the "." separator. It also
// provides additional mapping rules for time.Time, big.Int and big.Float.
// It can be modified to change the default behavior, but if the mapper is
// used by other packages, it is recommended to create a copy of the default
// mapper and modify the copy.
var DefaultMapper = &Mapper{
	Tag:       `map`,
	Separator: `.`,
	MapFrom: map[reflect.Type]MapFunc{
		timeTy:     mapTimeSrc,
		bigIntTy:   mapBigIntSrc,
		bigFloatTy: mapBigFloatSrc,
	},
	MapInto: map[reflect.Type]MapFunc{
		timeTy:     mapTimeDest,
		bigIntTy:   mapBigIntDest,
		bigFloatTy: mapBigFloatDest,
	},
}

// MapFunc is a function that maps a src value to a dest value. It returns an
// error if the mapping is not possible. The src and dest values are never
// pointers.
type MapFunc func(m *Mapper, src, dest reflect.Value) error

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

	// MapInto is a map of types that can map themselves to another type.
	MapInto map[reflect.Type]MapFunc

	// MapFrom is a map of types that can map themselves from another type.
	MapFrom map[reflect.Type]MapFunc
}

// Map maps the source value to the destination value.
//
// It is shorthand for DefaultMapper.Map(src, dest).
func Map(src, dest any) error {
	return DefaultMapper.Map(src, dest)
}

// MapRefl maps the source value to the destination value.
//
// It is shorthand for DefaultMapper.MapRefl(src, dest).
func MapRefl(src, dest reflect.Value) error {
	return DefaultMapper.MapRefl(src, dest)
}

func (m *Mapper) Map(src, dest any) error {
	return m.MapRefl(reflect.ValueOf(src), reflect.ValueOf(dest))
}

func (m *Mapper) MapRefl(src, dest reflect.Value) error {
	src = m.srcValue(src)
	dest = m.destValue(dest)
	if !src.IsValid() {
		return InvalidSrcErr
	}
	if !dest.IsValid() {
		return InvalidDestErr
	}
	if dest.Type() == anyTy && dest.CanSet() {
		dest.Set(src)
		return nil
	}
	if ok, err := m.mapFunc(src, dest); ok {
		return err
	}
	switch src.Kind() {
	case reflect.Bool:
		return m.mapBool(src, dest)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m.mapInt(src, dest)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return m.mapUint(src, dest)
	case reflect.Float32, reflect.Float64:
		return m.mapFloat(src, dest)
	case reflect.String:
		return m.mapString(src, dest)
	case reflect.Slice:
		return m.mapSlice(src, dest)
	case reflect.Array:
		return m.mapArray(src, dest)
	case reflect.Map:
		return m.mapMap(src, dest)
	case reflect.Struct:
		return m.mapStruct(src, dest)
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
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
	if m.MapInto != nil {
		cpy.MapInto = make(map[reflect.Type]MapFunc)
		for k, v := range m.MapInto {
			cpy.MapInto[k] = v
		}
	}
	return cpy
}

// mapFunc tries to map the source value to the destination value using the
// MapFrom and MapInto interfaces, and the MapFrom and MapInto maps.
//
// It tries to use every defined mapping function until one of them succeeds.
// If no mapping function succeeds, it returns an error from the last mapping
// function that was tried.
func (m *Mapper) mapFunc(src, dest reflect.Value) (ok bool, err error) {
	if src.Type().Implements(mapIntoTy) {
		if err = src.Interface().(MapInto).MapInto(m, dest); err == nil {
			return true, nil
		}
	}
	if dest.Type().Implements(mapFromTy) {
		if err = dest.Interface().(MapFrom).MapFrom(m, src); err == nil {
			return true, nil
		}
	}
	if f, ok := m.MapInto[dest.Type()]; ok {
		if err = f(m, src, dest); err == nil {
			return true, nil
		}
	}
	if f, ok := m.MapFrom[src.Type()]; ok {
		if err = f(m, src, dest); err == nil {
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

func (m *Mapper) mapBool(src, dest reflect.Value) error {
	if m.StrictTypes && dest.Kind() != reflect.Bool {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(src.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src.Bool() {
			dest.SetInt(1)
		} else {
			dest.SetInt(0)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src.Bool() {
			dest.SetUint(1)
		} else {
			dest.SetUint(0)
		}
		return nil
	case reflect.Float32, reflect.Float64:
		if src.Bool() {
			dest.SetFloat(1)
		} else {
			dest.SetFloat(0)
		}
		return nil
	case reflect.String:
		if src.Bool() {
			dest.SetString("true")
		} else {
			dest.SetString("false")
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapInt(src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(src.Int() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dest.OverflowInt(src.Int()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(src.Int())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := uint64(src.Int())
		if dest.OverflowUint(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n := float64(src.Int())
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(strconv.FormatInt(src.Int(), 10))
		return nil
	case reflect.Slice:
		// If the destination is a slice of bytes, store the integer as a
		// big-endian byte slice.
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetBytes(new(big.Int).SetInt64(src.Int()).Bytes())
			return nil
		}
	case reflect.Array:
		// If the destination is an array of bytes, store the integer as a
		// big-endian byte array, but only if the array length is large enough.
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			b := new(big.Int).SetInt64(src.Int()).Bytes()
			if len(b) > dest.Len() {
				return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
			}
			for i := 0; i < len(b); i++ {
				dest.Index(i).SetUint(uint64(b[i]))
			}
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapUint(src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(src.Uint() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int64(src.Uint())
		if dest.OverflowInt(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if dest.OverflowUint(src.Uint()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(src.Uint())
		return nil
	case reflect.Float32, reflect.Float64:
		n := float64(src.Uint())
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(strconv.FormatUint(src.Uint(), 10))
		return nil
	case reflect.Slice:
		// If the destination is a slice of bytes, store the integer as a
		// big-endian byte slice.
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetBytes(new(big.Int).SetUint64(src.Uint()).Bytes())
			return nil
		}
		// If the destination is an array of bytes, store the integer as a
		// big-endian byte array, but only if the array length is large enough.
	case reflect.Array:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			b := new(big.Int).SetUint64(src.Uint()).Bytes()
			if len(b) > dest.Len() {
				return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
			}
			for i := 0; i < len(b); i++ {
				dest.Index(i).SetUint(uint64(b[i]))
			}
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapFloat(src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(src.Float() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := src.Float()
		if n < math.MinInt64 || n > math.MaxInt64 || dest.OverflowInt(int64(n)) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(int64(n))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := src.Float()
		if n < 0 || n > math.MaxUint64 || dest.OverflowUint(uint64(n)) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(uint64(n))
		return nil
	case reflect.Float32, reflect.Float64:
		n := src.Float()
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(strconv.FormatFloat(src.Float(), 'f', -1, 64))
		return nil
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapString(src, dest reflect.Value) error {
	if m.StrictTypes && dest.Kind() != reflect.String {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Bool:
		switch src.String() {
		case "true":
			dest.SetBool(true)
			return nil
		case "false":
			dest.SetBool(false)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dest.Type(), "invalid number")
		}
		if dest.OverflowInt(n.Int64()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(n.Int64())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dest.Type(), "invalid number")
		}
		if dest.OverflowUint(n.Uint64()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(n.Uint64())
		return nil
	case reflect.Float32, reflect.Float64:
		bn, ok := new(big.Float).SetString(src.String())
		if !ok {
			return NewInvalidMappingError(src.Type(), dest.Type(), "invalid number")
		}
		n, _ := bn.Float64()
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(src.String())
		return nil
	case reflect.Slice:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetBytes([]byte(src.String()))
			return nil
		}
	case reflect.Array:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			b := []byte(src.String())
			if len(b) != dest.Len() {
				return NewInvalidMappingError(src.Type(), dest.Type(), "length mismatch")
			}
			for i := 0; i < len(b); i++ {
				dest.Index(i).SetUint(uint64(b[i]))
			}
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapSlice(src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() && dest.Kind() != reflect.Map {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetInt(new(big.Int).SetBytes(src.Bytes()).Int64())
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetUint(new(big.Int).SetBytes(src.Bytes()).Uint64())
			return nil
		}
	case reflect.String:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetString(string(src.Bytes()))
			return nil
		}
	case reflect.Slice:
		if src.Type() == dest.Type() {
			dest.Set(reflect.MakeSlice(dest.Type(), src.Len(), src.Cap()))
			reflect.Copy(dest, src)
			return nil
		}
		dest.Set(reflect.MakeSlice(dest.Type(), src.Len(), src.Len()))
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dest.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Array:
		if src.Len() != dest.Len() {
			return NewInvalidMappingError(src.Type(), dest.Type(), "length mismatch")
		}
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dest.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapArray(src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch dest.Kind() {
	case reflect.Slice:
		dest.Set(reflect.MakeSlice(dest.Type(), src.Len(), src.Len()))
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dest.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Array:
		if src.Type() == dest.Type() {
			dest.Set(src)
			return nil
		}
		if src.Len() != dest.Len() {
			return NewInvalidMappingError(src.Type(), dest.Type(), "length mismatch")
		}
		for i := 0; i < src.Len(); i++ {
			if err := m.MapRefl(src.Index(i), dest.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

func (m *Mapper) mapMap(src, dest reflect.Value) error {
	switch dest.Kind() {
	case reflect.Struct:
		if err := m.MapRefl(src, m.structToPtrsMap(dest, true)); err != nil {
			return err
		}
		return nil
	case reflect.Map:
		for _, key := range src.MapKeys() {
			// It is important here to use destValue because we need to check
			// if the value can be set directly or if we need to create a new
			// value, the destValue function will always return a value that
			// can be set, otherwise it will return an invalid value.
			destVal := m.destValue(dest.MapIndex(key))
			if destVal.IsValid() {
				if err := m.MapRefl(src.MapIndex(key), destVal); err != nil {
					return err
				}
			} else {
				v := reflect.New(dest.Type().Elem()).Elem()
				if err := m.MapRefl(src.MapIndex(key), v); err != nil {
					return err
				}
				dest.SetMapIndex(key, v)
			}
		}
		return nil
	}
	return fmt.Errorf("mapper: cannot map map to %v", dest.Type())
}

func (m *Mapper) mapStruct(src, dest reflect.Value) error {
	switch dest.Kind() {
	case reflect.Struct:
		if src.Type() == dest.Type() {
			dest.Set(src)
			return nil
		}
		return m.MapRefl(m.structToPtrsMap(src, false), m.structToPtrsMap(dest, true))
	case reflect.Map:
		if dest.Type().Key().Kind() != reflect.String {
			return NewInvalidMappingError(src.Type(), dest.Type(), "map key must be string")
		}
		return m.MapRefl(m.structToPtrsMap(src, false), dest)
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

// structToPtrsMap returns a map where the keys are the field names and the
// values are pointers to the fields. If struct field has a tag, the tag is
// used as the key. If the tag has a nested field (e.g. "foo.bar"), the
// resulting map will have a nested map (e.g. "foo" => "bar" => &field).
func (m *Mapper) structToPtrsMap(v reflect.Value, initialize bool) reflect.Value {
	r := make(map[string]any)
	t := v.Type()
	for idx := 0; idx < v.NumField(); idx++ {
		vField := v.Field(idx)
		tField := t.Field(idx)
		if initialize {
			// The value needs to be initialized here to make sure that
			// the value is addressable, so that it will be possible to
			// store in a map a pointer to it. The destValue method will
			// do initialization if needed.
			vField = m.destValue(vField)
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
		if m.MapFrom[v.Type()] != nil {
			return v
		}
		v = v.Elem()
	}
	return v
}

// destValue unpacks values from pointers and interfaces until it reaches a
// settable non-pointer or non-interface value, value that implements the
// MapInto interface, has a custom mapper, or a value that is a map, slice or
// array. It returns an invalid value if it cannot find a value that meets
// these conditions. If the value is a pointer, map or slice, it will be
// initialized if needed.
func (m *Mapper) destValue(v reflect.Value) reflect.Value {
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
		if v.Type().Implements(mapIntoTy) {
			return v
		}
		if m.MapInto[v.Type()] != nil {
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

// InvalidDestErr is returned when reflect.IsValid returns false for the
// destination value. It may happen when the destination value was not
// passed as a pointer.
var InvalidDestErr = errors.New("mapper: invalid destination value")

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
	mapIntoTy = reflect.TypeOf((*MapInto)(nil)).Elem()
	anyTy     = reflect.TypeOf((*any)(nil)).Elem()
)
