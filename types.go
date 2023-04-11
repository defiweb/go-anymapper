package anymapper

import (
	"math"
	"math/big"
	"reflect"
	"time"
)

var (
	timeTy     = reflect.TypeOf((*time.Time)(nil)).Elem()
	bigIntTy   = reflect.TypeOf((*big.Int)(nil)).Elem()
	bigFloatTy = reflect.TypeOf((*big.Float)(nil)).Elem()
	bigRatTy   = reflect.TypeOf((*big.Rat)(nil)).Elem()
)

func timeTypeMapper(_ *Mapper, src, dst reflect.Type) MapFunc {
	if src == dst {
		return mapDirect
	}
	switch {
	case src == timeTy:
		switch dst.Kind() {
		case reflect.String:
			return mapTimeToString
		case reflect.Int, reflect.Int32, reflect.Int64:
			return mapTimeToInt
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			return mapTimeToUint
		case reflect.Float32, reflect.Float64:
			return mapTimeToFloat
		case reflect.Struct:
			switch dst {
			case bigIntTy:
				return mapTimeToBigInt
			case bigFloatTy:
				return mapTimeToBigFloat
			}
		case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
			return nil
		}
		return mapFromTimeViaInt64
	case dst == timeTy:
		switch src.Kind() {
		case reflect.String:
			return mapStringToTime
		case reflect.Int, reflect.Int32, reflect.Int64:
			return mapIntToTime
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			return mapUintToTime
		case reflect.Float32, reflect.Float64:
			return mapFloatToTime
		case reflect.Struct:
			switch src {
			case bigIntTy:
				return mapBigIntToTime
			case bigFloatTy:
				return mapBigFloatToTime
			}
		case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
			return nil
		}
		return mapToTimeViaInt64
	}
	return nil
}

func bigIntTypeMapper(_ *Mapper, src, dst reflect.Type) MapFunc {
	if src == dst {
		return mapDirect
	}
	switch {
	case src == bigIntTy:
		switch dst.Kind() {
		case reflect.Bool:
			return mapBigIntToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapBigIntToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapBigIntToUint
		case reflect.Float32, reflect.Float64:
			return mapBigIntToFloat
		case reflect.String:
			return mapBigIntToString
		case reflect.Slice:
			if dst.Elem().Kind() == reflect.Uint8 {
				return mapBigIntToBytes
			}
		case reflect.Struct:
			if bigFloatTy == dst {
				return mapBigIntToBigFloat
			}
		}
	case dst == bigIntTy:
		switch src.Kind() {
		case reflect.Bool:
			return mapBoolToBigInt
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapIntToBigInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapUintToBigInt
		case reflect.Float32, reflect.Float64:
			return mapFloatToBigInt
		case reflect.String:
			return mapStringToBigInt
		case reflect.Slice:
			if src.Elem().Kind() == reflect.Uint8 {
				return mapBytesToBigInt
			}
		case reflect.Struct:
			if bigFloatTy == src {
				return mapBigFloatToBigInt
			}
		}
	}
	return nil
}

func bigRatTypeMapper(_ *Mapper, src, dst reflect.Type) MapFunc {
	if src == bigRatTy && dst == bigRatTy {
		return mapDirect
	}
	switch {
	case src == bigRatTy:
		switch dst.Kind() {
		case reflect.String:
			return mapBigRatToString
		case reflect.Slice, reflect.Array:
			return mapBigRatToSliceOrArray
		}
		return mapFromBigRatViaBigFloat
	case dst == bigRatTy:
		switch src.Kind() {
		case reflect.String:
			return mapStringToBigRat
		case reflect.Slice, reflect.Array:
			return mapSliceOrArrayToBigRat
		}
		return mapToBigRatViaBigFloat
	}
	return nil
}

func bigFloatTypeMapper(_ *Mapper, src, dst reflect.Type) MapFunc {
	if src == dst {
		return mapDirect
	}
	switch {
	case src == bigFloatTy:
		switch dst.Kind() {
		case reflect.Bool:
			return mapBigFloatToBool
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapBigFloatToInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapBigFloatToUint
		case reflect.Float32, reflect.Float64:
			return mapBigFloatToFloat
		case reflect.String:
			return mapBigFloatToString
		case reflect.Struct:
			if bigIntTy == dst {
				return mapBigFloatToBigInt
			}
		}
	case dst == bigFloatTy:
		switch src.Kind() {
		case reflect.Bool:
			return mapBoolToBigFloat
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return mapIntToBigFloat
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return mapUintToBigFloat
		case reflect.Float32, reflect.Float64:
			return mapFloatToBigFloat
		case reflect.String:
			return mapStringToBigFloat
		case reflect.Struct:
			if bigIntTy == src {
				return mapBigIntToBigFloat
			}
		}
	}
	return nil
}

func mapTimeToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(src.Interface().(time.Time).Format(time.RFC3339))
	return nil
}

func mapTimeToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	unix := src.Interface().(time.Time).Unix()
	if dst.OverflowInt(unix) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(unix)
	return nil
}

func mapTimeToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	unix := src.Interface().(time.Time).Unix()
	if dst.OverflowUint(uint64(unix)) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(uint64(unix))
	return nil
}

func mapTimeToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm := src.Interface().(time.Time)
	unix := tm.Unix()
	nano := tm.Nanosecond()
	dst.SetFloat(float64(unix) + float64(nano)/1e9)
	return nil
}

func mapTimeToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	unix := src.Interface().(time.Time).Unix()
	dst.Set(reflect.ValueOf(big.NewInt(unix)).Elem())
	return nil
}

func mapTimeToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm := src.Interface().(time.Time)
	unix := tm.Unix()
	nano := tm.Nanosecond()
	bf := new(big.Float).SetInt64(unix)
	bn := new(big.Float).SetInt64(int64(nano))
	bn = bn.Quo(bn, big.NewFloat(1e9))
	bf = bf.Add(bf, bn)
	dst.Set(reflect.ValueOf(bf).Elem())
	return nil
}

func mapStringToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm, err := time.Parse(time.RFC3339, src.String())
	if err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
	}
	dst.Set(reflect.ValueOf(tm))
	return nil
}

func mapIntToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm := time.Unix(src.Int(), 0).UTC()
	dst.Set(reflect.ValueOf(tm))
	return nil
}

func mapUintToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm := time.Unix(int64(src.Uint()), 0).UTC()
	dst.Set(reflect.ValueOf(tm))
	return nil
}

func mapFloatToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	f := src.Float()
	unix := int64(f)
	nano := int64((f - float64(unix)) * 1e9)
	tm := time.Unix(unix, nano).UTC()
	dst.Set(reflect.ValueOf(tm))
	return nil
}

func mapBigIntToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	tm := time.Unix(src.Addr().Interface().(*big.Int).Int64(), 0).UTC()
	dst.Set(reflect.ValueOf(tm))
	return nil
}

func mapBigFloatToTime(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	bf := src.Addr().Interface().(*big.Float)
	unix, _ := bf.Int(nil)
	frac := new(big.Float).Sub(bf, new(big.Float).SetInt(unix))
	nano, _ := frac.Mul(frac, big.NewFloat(1e9)).Int(nil)
	dst.Set(reflect.ValueOf(time.Unix(unix.Int64(), nano.Int64()).UTC()))
	return nil
}

func mapFromTimeViaInt64(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	aux := src.Interface().(time.Time).Unix()
	if err := m.MapRefl(reflect.ValueOf(aux), dst); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

func mapToTimeViaInt64(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	var aux int64
	if err := m.MapRefl(src, reflect.ValueOf(&aux)); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	dst.Set(reflect.ValueOf(time.Unix(aux, 0).UTC()))
	return nil
}

func mapBigIntToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetBool(src.Addr().Interface().(*big.Int).Cmp(big.NewInt(0)) != 0)
	return nil
}

func mapBigIntToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Int)
	n := v.Int64()
	if !v.IsInt64() || dst.OverflowInt(n) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(n)
	return nil
}

func mapBigIntToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Int)
	n := v.Uint64()
	if !v.IsUint64() || dst.OverflowUint(n) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(n)
	return nil
}

func mapBigIntToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Int)
	n, a := new(big.Float).SetInt(v).Float64()
	if dst.OverflowFloat(n) || (math.IsInf(n, 0) && (a == big.Below || a == big.Above)) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetFloat(n)
	return nil
}

func mapBigIntToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(src.Addr().Interface().(*big.Int).String())
	return nil
}

func mapBigIntToBytes(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Int)
	if v.Sign() < 0 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "cannot convert negative big.Int to bytes")
	}
	dst.SetBytes(v.Bytes())
	return nil
}

func mapBigIntToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(new(big.Float).SetInt(src.Addr().Interface().(*big.Int))).Elem())
	return nil
}

func mapBoolToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Bool() {
		dst.Set(reflect.ValueOf(big.NewInt(1)).Elem())
	} else {
		dst.Set(reflect.ValueOf(big.NewInt(0)).Elem())
	}
	return nil
}

func mapIntToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(big.NewInt(src.Int())).Elem())
	return nil
}

func mapUintToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(big.NewInt(0).SetUint64(src.Uint())).Elem())
	return nil
}

func mapFloatToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, _ := new(big.Float).SetFloat64(src.Float()).Int(nil)
	dst.Set(reflect.ValueOf(new(big.Int).Set(v)).Elem())
	return nil
}

func mapStringToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, ok := new(big.Int).SetString(src.String(), 0)
	if !ok {
		return NewInvalidMappingError(src.Type(), dst.Type(), "invalid string")
	}
	dst.Set(reflect.ValueOf(v).Elem())
	return nil
}

func mapBytesToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(new(big.Int).SetBytes(src.Bytes())).Elem())
	return nil
}

func mapBigFloatToBigInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, _ := src.Addr().Interface().(*big.Float).Int(nil)
	dst.Set(reflect.ValueOf(new(big.Int).Set(v)).Elem())
	return nil
}

func mapBigFloatToBool(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Float)
	if v.Sign() == 0 {
		dst.SetBool(false)
	} else {
		dst.SetBool(true)
	}
	return nil
}

func mapBigFloatToInt(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, _ := src.Addr().Interface().(*big.Float).Int(nil)
	n := v.Int64()
	if !v.IsInt64() || dst.OverflowInt(n) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetInt(n)
	return nil
}

func mapBigFloatToUint(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, _ := src.Addr().Interface().(*big.Float).Int(nil)
	n := v.Uint64()
	if !v.IsUint64() || dst.OverflowUint(n) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetUint(n)
	return nil
}

func mapBigFloatToFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v := src.Addr().Interface().(*big.Float)
	n, a := v.Float64()
	if dst.OverflowFloat(n) || (math.IsInf(n, 0) && (a == big.Below || a == big.Above)) {
		return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
	}
	dst.SetFloat(n)
	return nil
}

func mapBigFloatToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(src.Addr().Interface().(*big.Float).String())
	return nil
}

func mapBoolToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	switch src.Bool() {
	case true:
		dst.Set(reflect.ValueOf(big.NewFloat(1)).Elem())
	case false:
		dst.Set(reflect.ValueOf(big.NewFloat(0)).Elem())
	}
	return nil
}

func mapIntToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(new(big.Float).SetInt64(src.Int())).Elem())
	return nil
}

func mapUintToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(new(big.Float).SetUint64(src.Uint())).Elem())
	return nil
}

func mapFloatToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.Set(reflect.ValueOf(new(big.Float).SetFloat64(src.Float())).Elem())
	return nil
}

func mapStringToBigFloat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, ok := new(big.Float).SetString(src.String())
	if !ok {
		return NewInvalidMappingError(src.Type(), dst.Type(), "string is not a valid float number")
	}
	dst.Set(reflect.ValueOf(v).Elem())
	return nil
}

func mapBigRatToString(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	dst.SetString(src.Addr().Interface().(*big.Rat).String())
	return nil
}

func mapBigRatToSliceOrArray(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if dst.Kind() == reflect.Slice {
		dst.Set(reflect.MakeSlice(dst.Type(), 2, 2))
	}
	if dst.Kind() == reflect.Array && dst.Len() != 2 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "array must have length 2")
	}
	v := src.Addr().Interface().(*big.Rat)
	if err := m.MapRefl(reflect.ValueOf(v.Num()), dst.Index(0)); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	if err := m.MapRefl(reflect.ValueOf(v.Denom()), dst.Index(1)); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

func mapStringToBigRat(_ *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	v, ok := new(big.Rat).SetString(src.String())
	if !ok {
		return NewInvalidMappingError(src.Type(), dst.Type(), "string is not a valid rational number")
	}
	dst.Set(reflect.ValueOf(v).Elem())
	return nil
}

func mapSliceOrArrayToBigRat(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	if src.Len() != 2 {
		return NewInvalidMappingError(src.Type(), dst.Type(), "array must have length 2")
	}
	var num, den big.Int
	if err := m.MapRefl(src.Index(0), reflect.ValueOf(&num).Elem()); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	if err := m.MapRefl(src.Index(1), reflect.ValueOf(&den).Elem()); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	dst.Set(reflect.ValueOf(new(big.Rat).SetFrac(&num, &den)).Elem())
	return nil
}

func mapFromBigRatViaBigFloat(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	aux := new(big.Float).SetRat(src.Addr().Interface().(*big.Rat))
	if err := m.MapRefl(reflect.ValueOf(aux), dst); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

func mapToBigRatViaBigFloat(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if ctx.StrictTypes {
		return NewStrictMappingError(src.Type(), dst.Type())
	}
	aux := reflect.New(bigFloatTy).Elem()
	if err := m.MapRefl(src, aux); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	rat, _ := aux.Addr().Interface().(*big.Float).Rat(nil)
	dst.Set(reflect.ValueOf(rat).Elem())
	return nil
}
