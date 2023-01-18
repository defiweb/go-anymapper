package anymapper

import "reflect"

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

// MappingInterfaceHooks is a set of hooks that checks if the source or
// destination type implements the MapTo or MapFrom interface. If so, it
// will use one of those interfaces to map the value. If both interfaces
// are implemented, MapTo will be used.
var MappingInterfaceHooks = Hooks{
	MapFuncHook: func(m *Mapper, src, dst reflect.Type) MapFunc {
		if isSimpleType(src) && isSimpleType(dst) {
			return nil
		}
		if implMapTo(src) {
			return mapToInterface
		}
		if implMapFrom(dst) {
			return mapFromInterface
		}
		return nil
	},
	SourceValueHook: func(v reflect.Value) reflect.Value {
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			if _, ok := v.Interface().(MapTo); ok {
				return v
			}
			v = v.Elem()
		}
		return reflect.Value{}
	},
	DestinationValueHook: func(v reflect.Value) reflect.Value {
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			if v.Kind() == reflect.Ptr && v.IsNil() {
				if !v.CanSet() {
					return reflect.Value{}
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			if _, ok := v.Interface().(MapFrom); ok {
				return v
			}
			v = v.Elem()
		}
		return reflect.Value{}
	},
}

// mapFromInterface is the MapFunc that is used to map a value using the
// MapFrom interface.
func mapFromInterface(m *Mapper, src, dst reflect.Value) error {
	return dst.Interface().(MapFrom).MapFrom(m, src)
}

// mapToInterface is the MapFunc that is used to map a value using the
// MapTo interface.
func mapToInterface(m *Mapper, src, dst reflect.Value) error {
	return src.Interface().(MapTo).MapTo(m, dst)
}

// implMapTo returns true if the type implements the MapTo interface.
func implMapTo(t reflect.Type) bool {
	_, ok := reflect.Zero(t).Interface().(MapTo)
	return ok
}

// implMapFrom returns true if the type implements the MapFrom interface.
func implMapFrom(t reflect.Type) bool {
	_, ok := reflect.Zero(t).Interface().(MapFrom)
	return ok
}
