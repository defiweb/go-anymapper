package anymapper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func builtInTypesMapper(_ *Mapper, src, dst reflect.Type) MapFunc {
	switch src.Kind() {
	case reflect.Bool:
		switch dst.Kind() {
		case reflect.Bool:
			return mapBoolToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapBoolToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapBoolToUint
		case reflect.Float32, reflect.Float64:
			return mapBoolToFloat
		case reflect.String:
			return mapBoolToString
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch dst.Kind() {
		case reflect.Bool:
			return mapIntToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapIntToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapIntToUint
		case reflect.Float32, reflect.Float64:
			return mapIntToFloat
		case reflect.String:
			return mapIntToString
		case reflect.Slice, reflect.Array:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapIntToByteSliceOrByteArray
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch dst.Kind() {
		case reflect.Bool:
			return mapUintToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapUintToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapUintToUint
		case reflect.Float32, reflect.Float64:
			return mapUintToFloat
		case reflect.String:
			return mapUintToString
		case reflect.Slice, reflect.Array:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapUintToByteSliceOrByteArray
			}
		}
	case reflect.Float32, reflect.Float64:
		switch dst.Kind() {
		case reflect.Bool:
			return mapFloatToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapFloatToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapFloatToUint
		case reflect.Float32, reflect.Float64:
			return mapFloatToFloat
		case reflect.String:
			return mapFloatToString
		case reflect.Slice, reflect.Array:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapFloatToByteSliceOrByteArray
			}
		}
	case reflect.String:
		switch dst.Kind() {
		case reflect.Bool:
			return mapStringToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapStringToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapStringToUint
		case reflect.Float32, reflect.Float64:
			return mapStringToFloat
		case reflect.String:
			return mapStringToString
		case reflect.Slice:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapStringToByteSlice
			}
		case reflect.Array:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapStringToByteArray
			}
		}
	case reflect.Slice:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			if src.Elem().Kind() == reflect.Uint8 {
				return mapByteSliceToNumber
			}
		case reflect.String:
			if src.Elem().Kind() == reflect.Uint8 {
				return mapByteSliceToString
			}
		case reflect.Slice:
			return mapSliceToSlice
		case reflect.Array:
			return mapSliceToArray
		}
	case reflect.Array:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			if src.Elem().Kind() == reflect.Uint8 {
				return mapByteArrayToNumber
			}
		case reflect.String:
			if src.Elem().Kind() == reflect.Uint8 {
				return mapByteArrayToString
			}
		case reflect.Slice:
			return mapArrayToSlice
		case reflect.Array:
			return mapArrayToArray
		}
	case reflect.Map:
		switch dst.Kind() {
		case reflect.Map:
			return mapMapToMap
		case reflect.Struct:
			return mapMapToStruct
		}
	case reflect.Struct:
		switch dst.Kind() {
		case reflect.Struct:
			switch {
			case src == dst:
				return mapStructsOfSameType
			default:
				return mapStructsOfDifferentTypes
			}
		case reflect.Map:
			if dst.Key().Kind() == reflect.String {
				return mapStructToMap
			}
		}
	default:
		return nil
	}
	return nil
}

func mapBoolToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBool(src.Bool())
	return nil
}

func mapBoolToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Bool() {
		dst.SetInt(1)
	} else {
		dst.SetInt(0)
	}
	return nil
}

func mapBoolToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Bool() {
		dst.SetUint(1)
	} else {
		dst.SetUint(0)
	}
	return nil
}

func mapBoolToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Bool() {
		dst.SetFloat(1)
	} else {
		dst.SetFloat(0)
	}
	return nil
}

func mapBoolToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Bool() {
		dst.SetString("true")
	} else {
		dst.SetString("false")
	}
	return nil
}

func mapIntToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBool(src.Int() != 0)
	return nil
}

func mapIntToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if dst.OverflowInt(src.Int()) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(src.Int())
	return nil
}

func mapIntToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Int() < 0 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "negative value")
	}
	if dst.OverflowUint(uint64(src.Int())) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(uint64(src.Int()))
	return nil
}

func mapIntToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetFloat(float64(src.Int()))
	return nil
}

func mapIntToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(strconv.FormatInt(src.Int(), 10))
	return nil
}

func mapIntToByteSliceOrByteArray(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	return numberToBytes(ctx, src, dst)
}

func mapUintToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBool(src.Uint() != 0)
	return nil
}

func mapUintToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Uint() > math.MaxInt64 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	if dst.OverflowInt(int64(src.Uint())) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(int64(src.Uint()))
	return nil
}

func mapUintToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if dst.OverflowUint(src.Uint()) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(src.Uint())
	return nil
}

func mapUintToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetFloat(float64(src.Uint()))
	return nil
}

func mapUintToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(strconv.FormatUint(src.Uint(), 10))
	return nil
}

func mapUintToByteSliceOrByteArray(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	return numberToBytes(ctx, src, dst)
}

func mapFloatToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBool(src.Float() != 0)
	return nil
}

func mapFloatToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Float() > math.MaxInt64 || src.Float() < math.MinInt64 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	if dst.OverflowInt(int64(src.Float())) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(int64(src.Float()))
	return nil
}

func mapFloatToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Float() < 0 || src.Float() > math.MaxUint64 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	if dst.OverflowUint(uint64(src.Float())) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(uint64(src.Float()))
	return nil
}

func mapFloatToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if dst.OverflowFloat(src.Float()) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetFloat(src.Float())
	return nil
}

func mapFloatToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(strconv.FormatFloat(src.Float(), 'f', -1, 64))
	return nil
}

func mapFloatToByteSliceOrByteArray(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	return numberToBytes(ctx, src, dst)
}

func mapStringToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	switch src.String() {
	case "true":
		dst.SetBool(true)
	case "false":
		dst.SetBool(false)
	default:
		return NewInvalidMappingError(src.Type(), dst.Type(), "invalid string value")
	}
	return nil
}

func mapStringToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, err := strconv.ParseInt(src.String(), 10, 64)
	if err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
	}
	if dst.OverflowInt(v) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(v)
	return nil
}

func mapStringToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, err := strconv.ParseUint(src.String(), 10, 64)
	if err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
	}
	if dst.OverflowUint(v) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(v)
	return nil
}

func mapStringToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, err := strconv.ParseFloat(src.String(), 64)
	if err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
	}
	if dst.OverflowFloat(v) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetFloat(v)
	return nil
}

func mapStringToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(src.String())
	return nil
}

func mapStringToByteArray(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	b := []byte(src.String())
	if len(b) != dst.Len() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "length mismatch")
	}
	for i := 0; i < len(b); i++ {
		dst.Index(i).SetUint(uint64(b[i]))
	}
	return nil
}

func mapStringToByteSlice(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBytes([]byte(src.String()))
	return nil
}

func mapByteSliceToNumber(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	return numberFromBytes(ctx, src.Bytes(), dst)
}

func mapByteSliceToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(string(src.Bytes()))
	return nil
}

func mapByteArrayToNumber(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	b := make([]byte, src.Len())
	for i := 0; i < src.Len(); i++ {
		b[i] = byte(src.Index(i).Uint())
	}
	return numberFromBytes(ctx, b, dst)
}

func mapByteArrayToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	b := make([]byte, src.Len())
	for i := 0; i < src.Len(); i++ {
		b[i] = byte(src.Index(i).Uint())
	}
	dst.SetString(string(b))
	return nil
}

func mapSliceToSlice(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	mapper := m.mapperFor(ctx, src.Type().Elem(), dst.Type().Elem())
	if src.Type() == dst.Type() && dst.CanSet() {
		dst.Set(src)
		return nil
	}
	if src.Len() > dst.Len() {
		if dst.Cap() >= src.Len() {
			dst.SetLen(src.Len())
		} else {
			dst.Set(reflect.AppendSlice(
				dst,
				reflect.MakeSlice(dst.Type(), src.Len()-dst.Len(), src.Len()-dst.Len())),
			)
		}
	}
	for i := 0; i < src.Len(); i++ {
		srcVal := m.srcValue(src.Index(i))
		dstVal := m.dstValue(dst.Index(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
			return err
		}
	}
	return nil
}

func mapSliceToArray(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Len() != dst.Len() {
		return NewInvalidMappingError(
			src.Type(),
			dst.Type(),
			fmt.Sprintf("length mismatch: %d != %d", src.Len(), dst.Len()),
		)
	}
	srcTyp := src.Type().Elem()
	dstTyp := dst.Type().Elem()
	mapper := m.mapperFor(ctx, srcTyp, dstTyp)
	if srcTyp == dstTyp && dst.CanSet() {
		reflect.Copy(dst, src)
		return nil
	}
	for i := 0; i < src.Len(); i++ {
		srcVal := m.srcValue(src.Index(i))
		dstVal := m.dstValue(dst.Index(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, m.srcValue(src.Index(i)), m.dstValue(dst.Index(i))); err != nil {
			return err
		}
	}
	for i := src.Len(); i < dst.Len(); i++ {
		dst.Index(i).Set(reflect.Zero(dst.Type().Elem()))
	}
	return nil
}

func mapArrayToSlice(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	srcTyp := src.Type().Elem()
	dstTyp := dst.Type().Elem()
	mapper := m.mapperFor(ctx, srcTyp, dstTyp)
	if srcTyp == dstTyp && dst.CanSet() {
		dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
		reflect.Copy(dst, src)
	} else {
		if src.Len() > dst.Len() {
			if dst.Cap() >= src.Len() {
				dst.SetLen(src.Len())
			} else {
				dst.Set(reflect.AppendSlice(
					dst,
					reflect.MakeSlice(dst.Type(), src.Len()-dst.Len(), src.Len()-dst.Len())),
				)
			}
		}
		for i := 0; i < src.Len(); i++ {
			srcVal := m.srcValue(src.Index(i))
			dstVal := m.dstValue(dst.Index(i))
			srcValTyp := srcVal.Type()
			dstValTyp := dstVal.Type()
			if !mapper.match(srcValTyp, dstValTyp) {
				mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
			}
			if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
				return err
			}
		}
	}
	return nil
}

func mapArrayToArray(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes && src.Type() != dst.Type() {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Len() != dst.Len() {
		return NewInvalidMappingError(
			src.Type(),
			dst.Type(),
			fmt.Sprintf("length mismatch: %d != %d", src.Len(), dst.Len()),
		)
	}
	srcTyp := src.Type().Elem()
	dstTyp := dst.Type().Elem()
	mapper := m.mapperFor(ctx, srcTyp, dstTyp)
	if srcTyp == dstTyp && dst.CanSet() {
		reflect.Copy(dst, src)
		return nil
	}
	for i := 0; i < src.Len(); i++ {
		srcVal := m.srcValue(src.Index(i))
		dstVal := m.dstValue(dst.Index(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
			return err
		}
	}
	return nil
}

func mapMapToStruct(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	mapper := &typeMapper{}
	dstNum := dst.Type().NumField()
	for i := 0; i < dstNum; i++ {
		dstFld := dst.Type().Field(i)
		if !dstFld.IsExported() {
			continue
		}
		tag, skip := m.parseTag(ctx, dstFld)
		if skip {
			// If the tag is "-", skip it.
			continue
		}
		srcKey := reflect.ValueOf(tag)
		srcVal := m.srcValue(src.MapIndex(srcKey))
		if !srcVal.IsValid() {
			// If the source map doesn't have a value for the key, skip it.
			continue
		}
		dstVal := m.dstValue(dst.Field(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
			return err
		}
	}
	return nil
}

func mapMapToMap(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	var (
		srcKeyTyp  = src.Type().Key()
		dstKeyTyp  = dst.Type().Key()
		srcElemTyp = src.Type().Elem()
		dstElemTyp = dst.Type().Elem()
		keyMapper  = m.mapperFor(ctx, srcKeyTyp, dstKeyTyp)
		elemMapper = m.mapperFor(ctx, srcElemTyp, dstElemTyp)
		sameKeys   = srcKeyTyp == dstKeyTyp
	)
	for _, srcKey := range src.MapKeys() {
		dstKey := srcKey
		if !sameKeys {
			dstKey = reflect.New(dstKeyTyp).Elem()
			if err := keyMapper.mapRefl(m, ctx, m.srcValue(srcKey), m.dstValue(dstKey)); err != nil {
				return NewInvalidMappingError(srcKey.Type(), dstKeyTyp, "unable to map key")
			}
		}
		srcVal := m.srcValue(src.MapIndex(srcKey))
		dstVal := m.dstValue(dst.MapIndex(dstKey))
		if dstVal.IsValid() {
			// If the destination map already has a value for the key.
			srcValTyp := srcVal.Type()
			dstValTyp := dstVal.Type()
			if !elemMapper.match(srcValTyp, dstValTyp) {
				elemMapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
			}
			if err := elemMapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
				return err
			}
		} else {
			// If the destination map doesn't have a value for the key.
			newVal := reflect.New(dstElemTyp).Elem()
			dstVal := m.dstValue(newVal)
			srcValTyp := srcVal.Type()
			dstValTyp := dstVal.Type()
			if !dstVal.IsValid() {
				continue
			}
			if !elemMapper.match(srcValTyp, dstValTyp) {
				elemMapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
			}
			if err := elemMapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
				return err
			}
			dst.SetMapIndex(dstKey, newVal)
		}
	}
	return nil
}

func mapStructsOfSameType(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	var (
		mapper = &typeMapper{}
		srcTyp = src.Type()
		srcNum = src.NumField()
	)
	for i := 0; i < srcNum; i++ {
		srcFld := srcTyp.Field(i)
		if !srcFld.IsExported() {
			continue
		}
		if _, skip := m.parseTag(ctx, srcFld); skip {
			// If the tag is "-", skip it.
			continue
		}
		srcVal := m.srcValue(src.Field(i))
		dstVal := m.dstValue(dst.Field(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
			return err
		}
	}
	return nil
}

func mapStructsOfDifferentTypes(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	var (
		mapper = &typeMapper{}
		srcTyp = src.Type()
		dstTyp = dst.Type()
		srcNum = srcTyp.NumField()
		dstNum = dstTyp.NumField()
		valMap = map[string]reflect.Value{}
	)
	// Map the source struct to a map of values.
	for i := 0; i < srcNum; i++ {
		srcVal := src.Field(i)
		srcFld := srcTyp.Field(i)
		if !srcFld.IsExported() {
			continue
		}
		tag, skip := m.parseTag(ctx, srcFld)
		if skip {
			continue
		}
		valMap[tag] = srcVal
	}
	// Map the values to the destination struct.
	for i := 0; i < dstNum; i++ {
		dstFld := dst.Type().Field(i)
		if !dstFld.IsExported() {
			continue
		}
		tag, skip := m.parseTag(ctx, dstFld)
		if skip {
			// If the tag is "-", skip it.
			continue
		}
		var srcVal reflect.Value
		if val, ok := valMap[tag]; ok {
			srcVal = m.srcValue(val)
		} else {
			// If the source struct doesn't have a value for the key, skip it.
			continue
		}
		dstVal := m.dstValue(dst.Field(i))
		srcValTyp := srcVal.Type()
		dstValTyp := dstVal.Type()
		if !mapper.match(srcValTyp, dstValTyp) {
			mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
		}
		if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
			return err
		}
	}
	return nil
}

func mapStructToMap(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	var (
		mapper     = &typeMapper{}
		srcNum     = src.Type().NumField()
		dstElemTyp = dst.Type().Elem()
	)
	for i := 0; i < srcNum; i++ {
		srcFld := src.Type().Field(i)
		if !srcFld.IsExported() {
			continue
		}
		tag, skip := m.parseTag(ctx, srcFld)
		if skip {
			// If the tag is "-", skip it.
			continue
		}
		dstKey := reflect.ValueOf(tag)
		srcVal := m.srcValue(src.Field(i))
		dstVal := m.dstValue(dst.MapIndex(dstKey))
		if dstVal.IsValid() {
			// If the destination map already has a value for the key.
			srcValTyp := srcVal.Type()
			dstValTyp := dstVal.Type()
			if !mapper.match(srcValTyp, dstValTyp) {
				mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
			}
			if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
				return err
			}
		} else {
			// If the destination map doesn't have a value for the key.
			newVal := reflect.New(dstElemTyp).Elem()
			dstVal := m.dstValue(newVal)
			srcValTyp := srcVal.Type()
			dstValTyp := dstVal.Type()
			if !dstVal.IsValid() {
				continue
			}
			if !mapper.match(srcValTyp, dstValTyp) {
				mapper = m.mapperFor(ctx, srcValTyp, dstValTyp)
			}
			if err := mapper.mapRefl(m, ctx, srcVal, dstVal); err != nil {
				return err
			}
			dst.SetMapIndex(dstKey, newVal)
		}
	}
	return nil
}

// numberToBytes converts an int or uint to a byte slice using binary.Write.
func numberToBytes(ctx *Context, src, dst reflect.Value) error {
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
	if err := binary.Write(&buf, ctx.ByteOrder, src.Interface()); err != nil {
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

// numberFromBytes converts a byte slice to an int ot uint using binary.Read.
func numberFromBytes(ctx *Context, src []byte, dst reflect.Value) error {
	if len(src) != int(dst.Type().Size()) {
		return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "invalid byte slice length")
	}
	switch dst.Kind() {
	case reflect.Int:
		var v int64
		if err := binary.Read(bytes.NewReader(src), ctx.ByteOrder, &v); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
		if dst.OverflowInt(v) {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "overflow")
		}
		dst.SetInt(v)
	case reflect.Uint:
		var v uint64
		if err := binary.Read(bytes.NewReader(src), ctx.ByteOrder, &v); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
		if dst.OverflowUint(v) {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), "overflow")
		}
		dst.SetUint(v)
	default:
		if err := binary.Read(bytes.NewBuffer(src), ctx.ByteOrder, dst.Addr().Interface()); err != nil {
			return NewInvalidMappingError(reflect.TypeOf(src), dst.Type(), err.Error())
		}
	}
	return nil
}
