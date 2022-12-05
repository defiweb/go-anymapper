package anymapper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
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

// MapFuncProvider is a function that returns a MapFunc for given src and dst
// types. If mapping is not supported, it returns nil.
type MapFuncProvider func(m *Mapper, src, dst reflect.Type) MapFunc

// MapFunc is a function that maps a src value to a dst value. It returns an
// error if the mapping is not possible. The src and dst values are never
// pointers.
type MapFunc func(m *Mapper, src, dst reflect.Value) error

// DefaultMapper is the default Mapper used by the Map and MapRefl functions.
// It also provides additional mapping rules for time.Time, big.Int, big.Float
// and big.Rat. It can be modified to change the default behavior, but if the
// mapper is used by other packages, it is recommended to create a copy of the
// default mapper and modify the copy.
var DefaultMapper = &Mapper{
	Tag:       `map`,
	ByteOrder: binary.BigEndian,
	Mappers: map[reflect.Type]MapFuncProvider{
		timeTy:     timeTypeMapper,
		bigIntTy:   bigIntTypeMapper,
		bigFloatTy: bigFloatTypeMapper,
		bigRatTy:   bigRatTypeMapper,
	},
	cacheMap: make(map[typePair]*typeMapper, 0),
}

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

	// Mappers is a map of custom mapper providers. The key is the type that
	// the mapper can map to and from. The value is a function that returns
	// a MapFunc that maps the source type to the destination type. Provider
	// can return nil if the mapping is not possible.
	//
	// If both source and destination types have defined providers, then
	// the provider for source value is used first, and if it returns nil,
	// then the provider for destination value is used.
	Mappers map[reflect.Type]MapFuncProvider

	// ByteOrder is the byte order used to map numbers to and from byte slices.
	ByteOrder binary.ByteOrder

	// DisableCache disables the cache of the type mappers.
	DisableCache bool

	cacheMu  sync.Mutex
	cacheMap map[typePair]*typeMapper
}

// Map maps the source value to the destination value.
//
// It is shorthand for DefaultMapper.mapRefl(src, dst).
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
	srcVal := m.srcValue(reflect.ValueOf(src))
	dstVal := m.dstValue(reflect.ValueOf(dst))
	if !srcVal.IsValid() {
		return InvalidSrcErr
	}
	if !dstVal.IsValid() {
		return InvalidDstErr
	}
	return m.mapperFor(srcVal.Type(), dstVal.Type()).mapRefl(m, srcVal, dstVal)
}

// MapRefl maps the source value to the destination value.
func (m *Mapper) MapRefl(src, dst reflect.Value) error {
	srcVal := m.srcValue(src)
	dstVal := m.dstValue(dst)
	if !srcVal.IsValid() {
		return InvalidSrcErr
	}
	if !dstVal.IsValid() {
		return InvalidDstErr
	}
	return m.mapperFor(srcVal.Type(), dstVal.Type()).mapRefl(m, srcVal, dstVal)
}

// Copy creates a copy of the current Mapper with the same configuration.
func (m *Mapper) Copy() *Mapper {
	cpy := &Mapper{
		Tag:         m.Tag,
		FieldMapper: m.FieldMapper,
		ByteOrder:   m.ByteOrder,
		cacheMap:    make(map[typePair]*typeMapper, 0),
	}
	if m.Mappers != nil {
		cpy.Mappers = make(map[reflect.Type]MapFuncProvider)
		for k, v := range m.Mappers {
			cpy.Mappers[k] = v
		}
	}
	return cpy
}

// mapperFor returns the typeMapper that can map values of the given types.
// If mapping is not possible, the returned typeMapper has a nil MapFunc.
func (m *Mapper) mapperFor(src, dst reflect.Type) (tm *typeMapper) {
	// Try to find a mapper in the cache.
	if !m.DisableCache {
		m.cacheMu.Lock()
		if v, ok := m.cacheMap[typePair{src: src, dst: dst}]; ok {
			m.cacheMu.Unlock()
			return v
		}
		defer func() {
			m.cacheMap[typePair{src: src, dst: dst}] = tm
			m.cacheMu.Unlock()
		}()
	}

	tm = &typeMapper{
		SrcType: src,
		DstType: dst,
	}
	var isSrcSimple, isDstSimple, sameTypes bool
	if src == dst {
		isSrcSimple = isSimpleType(src)
		isDstSimple = isSrcSimple
		sameTypes = true
	} else {
		isSrcSimple = isSimpleType(src)
		isDstSimple = isSimpleType(dst)
	}

	// If both types are simple, e.g. int, string, etc. map the value directly
	// using reflect.Set.
	if (sameTypes && isSrcSimple) || dst == anyTy {
		tm.MapFunc = mapDirect
		return
	}

	// If src or dst implements MapTo or MapFrom, use the interface methods to
	// map the value.
	if !isSrcSimple && implMapTo(src) {
		tm.MapFunc = mapToInterface
		return
	}
	if !isDstSimple && !sameTypes && implMapFrom(dst) {
		tm.MapFunc = mapFromInterface
		return
	}

	// Try to find a mapper using mapper providers. It looks for providers
	// for src and dst types. First it tries to use providers for src. If
	// it returns a mapper, it uses it. If it returns nil, it tries to use
	// providers for dst. If both return nil, then mapping is not possible.
	var srcMapper, dstMapper MapFuncProvider
	var hasSrcMapper, hasDstMapper bool
	if !isSrcSimple {
		srcMapper, hasSrcMapper = m.Mappers[src]
	}
	if hasSrcMapper {
		tm.MapFunc = srcMapper(m, src, dst)
		if tm.MapFunc != nil {
			return
		}
	}
	if !sameTypes && !isDstSimple {
		dstMapper, hasDstMapper = m.Mappers[dst]
	}
	if hasDstMapper {
		tm.MapFunc = dstMapper(m, src, dst)
		if tm.MapFunc != nil {
			return
		}
	}
	if hasSrcMapper || hasDstMapper {
		return
	}

	// If none of the types implements MapTo or MapFrom, and there are no
	// mapper providers for them, try to use mapper for built-in types.
	tm.MapFunc = builtInTypesMapper(m, src, dst)
	return
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
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
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
	if v.Kind() != reflect.Interface && v.Kind() != reflect.Pointer && v.CanSet() {
		return v
	}
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
			for v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			return v
		}
		if m.Mappers[v.Type()] != nil {
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

// isSimpleType indicates whether a type is simple type.
//
// A type is considered simple if it is a built-in type, or it is a slice,
// array or map that is composed of build-in types. In other words, it is a
// type that cannot implement the MapFrom or MapTo interfaces nor contain
// fields or elements that can implement these interfaces, e.g. a custom type
// like `type MyInt int` is not considered simple because it can implement
// interfaces.
//
// Structs are never considered simple because they are rarely used without a
// custom type, and verifying if a struct is simple is too expensive.
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
		return strings.HasPrefix(p.String(), "[") && isSimpleType(p.Elem())
	case reflect.Array:
		return strings.HasPrefix(p.String(), "[") && isSimpleType(p.Elem())
	case reflect.Map:
		return strings.HasPrefix(p.String(), "map[") && isSimpleType(p.Elem()) && isSimpleType(p.Key())
	}
	return false
}

func implMapTo(t reflect.Type) bool {
	_, ok := reflect.Zero(t).Interface().(MapTo)
	return ok
}

func implMapFrom(t reflect.Type) bool {
	_, ok := reflect.Zero(t).Interface().(MapFrom)
	return ok
}

func mapDirect(_ *Mapper, src, dst reflect.Value) error {
	dst.Set(src)
	return nil
}

func mapFromInterface(m *Mapper, src, dst reflect.Value) error {
	return dst.Interface().(MapFrom).MapFrom(m, src)
}

func mapToInterface(m *Mapper, src, dst reflect.Value) error {
	return src.Interface().(MapTo).MapTo(m, dst)
}

// typeMapper contains a hint about how to map a value. It may ba passed to
// mapRefl method to help it decide how to map a value.
type typeMapper struct {
	SrcType reflect.Type
	DstType reflect.Type
	MapFunc MapFunc
}

func (tm *typeMapper) match(src, dst reflect.Type) bool {
	return tm.SrcType == src && tm.DstType == dst
}

func (tm *typeMapper) mapRefl(m *Mapper, src, dst reflect.Value) error {
	if tm.MapFunc == nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return tm.MapFunc(m, src, dst)
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

type typePair struct {
	src reflect.Type
	dst reflect.Type
}

var (
	anyTy     = reflect.TypeOf((*any)(nil)).Elem()
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
