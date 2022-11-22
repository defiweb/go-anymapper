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

var mapTimeSrc MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	srcVal := src.Interface().(time.Time)
	switch dst.Kind() {
	case reflect.Bool:
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	case reflect.String:
		dst.SetString(srcVal.Format(time.RFC3339))
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		if dst.OverflowInt(srcVal.Unix()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(srcVal.Unix())
		return nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		if dst.OverflowUint(uint64(srcVal.Unix())) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(uint64(srcVal.Unix()))
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return NewInvalidMappingError(src.Type(), dst.Type(), "int8, int16, uint8, uint16 are too small")
	case reflect.Float32, reflect.Float64:
		unix := srcVal.Unix()
		nano := srcVal.Nanosecond()
		dst.SetFloat(float64(unix) + float64(nano)/1e9)
		return nil
	case reflect.Struct:
		switch dst.Type() {
		case timeTy:
			dst.Set(src)
			return nil
		case bigIntTy:
			dst.Set(reflect.ValueOf(big.NewInt(srcVal.Unix())).Elem())
			return nil
		case bigFloatTy:
			unix := srcVal.Unix()
			nano := srcVal.Nanosecond()
			bf := new(big.Float).SetInt64(unix)
			bn := new(big.Float).SetInt64(int64(nano))
			bn = bn.Quo(bn, big.NewFloat(1e9))
			bf = bf.Add(bf, bn)
			dst.Set(reflect.ValueOf(bf).Elem())
			return nil
		}
	}
	// Try to use int64 as an intermediate type.
	if err := m.MapRefl(reflect.ValueOf(src.Interface().(time.Time).Unix()), dst); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

var mapTimeDst MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.Bool:
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	case reflect.String:
		tm, err := time.Parse(time.RFC3339, src.String())
		if err != nil {
			return NewInvalidMappingError(src.Type(), dst.Type(), err.Error())
		}
		dst.Set(reflect.ValueOf(tm))
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		dst.Set(reflect.ValueOf(time.Unix(src.Int(), 0).UTC()))
		return nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		dst.Set(reflect.ValueOf(time.Unix(int64(src.Uint()), 0).UTC()))
		return nil
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return NewInvalidMappingError(src.Type(), dst.Type(), "int8, int16, uint8, uint16 are too small")
	case reflect.Float32, reflect.Float64:
		unix, frac := math.Modf(src.Float())
		dst.Set(reflect.ValueOf(time.Unix(int64(unix), int64(frac*(1e9))).UTC()))
		return nil
	case reflect.Struct:
		switch src.Type() {
		case timeTy:
			dst.Set(src)
			return nil
		case bigIntTy:
			dst.Set(reflect.ValueOf(time.Unix(src.Addr().Interface().(*big.Int).Int64(), 0).UTC()))
			return nil
		case bigFloatTy:
			bf := src.Addr().Interface().(*big.Float)
			unix, _ := bf.Int(nil)
			frac := new(big.Float).Sub(bf, new(big.Float).SetInt(unix))
			nano, _ := frac.Mul(frac, big.NewFloat(1e9)).Int(nil)
			dst.Set(reflect.ValueOf(time.Unix(unix.Int64(), nano.Int64()).UTC()))
			return nil
		}
	}
	// Try to use int64 as an intermediate type.
	var timestamp int64
	if err := m.MapRefl(src, reflect.ValueOf(&timestamp)); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	dst.Set(reflect.ValueOf(time.Unix(timestamp, 0).UTC()))
	return nil
}

var mapBigIntSrc MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	srcVal := src.Addr().Interface().(*big.Int)
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(srcVal.Cmp(big.NewInt(0)) != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := srcVal.Int64()
		if !srcVal.IsInt64() || dst.OverflowInt(n) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := srcVal.Uint64()
		if !srcVal.IsUint64() || dst.OverflowUint(n) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, a := new(big.Float).SetInt(srcVal).Float64()
		if dst.OverflowFloat(n) || (math.IsInf(n, 0) && (a == big.Below || a == big.Above)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetFloat(n)
		return nil
	case reflect.String:
		dst.SetString(srcVal.String())
		return nil
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			dst.SetBytes(srcVal.Bytes())
			return nil
		}
	case reflect.Struct:
		switch dst.Type() {
		case bigIntTy:
			dst.Set(src)
			return nil
		case bigFloatTy:
			dst.Set(reflect.ValueOf(new(big.Float).SetInt(srcVal)).Elem())
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

var mapBigIntDst MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			dst.Set(reflect.ValueOf(big.NewInt(1)).Elem())
		} else {
			dst.Set(reflect.ValueOf(big.NewInt(0)).Elem())
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.Set(reflect.ValueOf(new(big.Int).SetInt64(src.Int())).Elem())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.Set(reflect.ValueOf(new(big.Int).SetUint64(src.Uint())).Elem())
		return nil
	case reflect.Float32, reflect.Float64:
		bi, _ := new(big.Float).SetFloat64(src.Float()).Int(nil)
		dst.Set(reflect.ValueOf(new(big.Int).Set(bi)).Elem())
		return nil
	case reflect.String:
		bn, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		dst.Set(reflect.ValueOf(bn).Elem())
		return nil
	case reflect.Slice:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dst.Set(reflect.ValueOf(new(big.Int).SetBytes(src.Bytes())).Elem())
			return nil
		}
	case reflect.Struct:
		switch src.Type() {
		case bigIntTy:
			dst.Set(src)
			return nil
		case bigFloatTy:
			bf := src.Addr().Interface().(*big.Float)
			bi, _ := bf.Int(nil)
			dst.Set(reflect.ValueOf(bi).Elem())
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

var mapBigFloatSrc MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	srcVal := src.Addr().Interface().(*big.Float)
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(srcVal.Sign() != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bn, _ := srcVal.Int(nil)
		if !bn.IsInt64() || dst.OverflowInt(bn.Int64()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetInt(bn.Int64())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bn, _ := srcVal.Int(nil)
		if !bn.IsUint64() || dst.OverflowUint(bn.Uint64()) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetUint(bn.Uint64())
		return nil
	case reflect.Float32, reflect.Float64:
		n, a := srcVal.Float64()
		if dst.OverflowFloat(n) || (math.IsInf(n, 0) && (a == big.Below || a == big.Above)) {
			return NewInvalidMappingError(src.Type(), dst.Type(), "overflow")
		}
		dst.SetFloat(n)
		return nil
	case reflect.String:
		dst.SetString(srcVal.String())
		return nil
	case reflect.Struct:
		switch dst.Type() {
		case bigIntTy:
			bi, _ := srcVal.Int(nil)
			dst.Set(reflect.ValueOf(bi).Elem())
			return nil
		case bigFloatTy:
			dst.Set(src)
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

var mapBigFloatDst MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			dst.Set(reflect.ValueOf(big.NewFloat(1)).Elem())
		} else {
			dst.Set(reflect.ValueOf(big.NewFloat(0)).Elem())
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.Set(reflect.ValueOf(new(big.Float).SetInt64(src.Int())).Elem())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.Set(reflect.ValueOf(new(big.Float).SetUint64(src.Uint())).Elem())
		return nil
	case reflect.Float32, reflect.Float64:
		dst.Set(reflect.ValueOf(new(big.Float).SetFloat64(src.Float())).Elem())
		return nil
	case reflect.String:
		bn, ok := new(big.Float).SetString(src.String())
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		dst.Set(reflect.ValueOf(bn).Elem())
		return nil
	case reflect.Struct:
		switch src.Type() {
		case bigIntTy:
			bn := src.Addr().Interface().(*big.Int)
			dst.Set(reflect.ValueOf(new(big.Float).SetInt(bn)).Elem())
			return nil
		case bigFloatTy:
			dst.Set(src)
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dst.Type(), "")
}

var mapBigRatSrc MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(src.Addr().Interface().(*big.Rat).String())
		return nil
	case reflect.Slice, reflect.Array:
		if dst.Kind() == reflect.Slice {
			dst.Set(reflect.MakeSlice(dst.Type(), 2, 2))
		}
		if dst.Kind() == reflect.Array && dst.Len() != 2 {
			return NewInvalidMappingError(src.Type(), dst.Type(), "array must have length 2")
		}
		bn := src.Addr().Interface().(*big.Rat)
		if err := m.MapRefl(reflect.ValueOf(bn.Num()), dst.Index(0)); err != nil {
			return NewInvalidMappingError(src.Type(), dst.Type(), "")
		}
		if err := m.MapRefl(reflect.ValueOf(bn.Denom()), dst.Index(1)); err != nil {
			return NewInvalidMappingError(src.Type(), dst.Type(), "")
		}
		return nil
	case reflect.Struct:
		switch dst.Type() {
		case bigRatTy:
			dst.Set(src)
			return nil
		case bigFloatTy:
			dst.Set(reflect.ValueOf(new(big.Float).SetRat(src.Addr().Interface().(*big.Rat))).Elem())
			return nil
		}
	}
	if dst.Kind() == reflect.String {
		dst.SetString(src.Addr().Interface().(*big.Rat).String())
		return nil
	}
	// Try to use big.Float as an intermediate.
	bf := new(big.Float).SetRat(src.Addr().Interface().(*big.Rat))
	if err := m.MapRefl(reflect.ValueOf(bf), dst); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}

var mapBigRatDst MapFunc = func(m *Mapper, src, dst reflect.Value) error {
	if m.StrictTypes && src.Type() != dst.Type() {
		return NewInvalidMappingError(src.Type(), dst.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.String:
		bn, ok := new(big.Rat).SetString(src.String())
		if !ok {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid number")
		}
		dst.Set(reflect.ValueOf(bn).Elem())
		return nil
	case reflect.Slice, reflect.Array:
		if src.Len() != 2 {
			return NewInvalidMappingError(src.Type(), dst.Type(), "invalid slice length")
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
	case reflect.Struct:
		switch src.Type() {
		case bigRatTy:
			dst.Set(src)
			return nil
		case bigFloatTy:
			br, _ := src.Addr().Interface().(*big.Float).Rat(nil)
			dst.Set(reflect.ValueOf(br).Elem())
			return nil
		}
	}
	// Try to use big.Float as an intermediate.
	bf := new(big.Float)
	if err := m.MapRefl(src, reflect.ValueOf(bf)); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	if err := m.MapRefl(reflect.ValueOf(bf), dst); err != nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return nil
}
