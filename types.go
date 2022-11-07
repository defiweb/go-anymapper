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
)

var mapTimeSrc MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	srcVal := src.Interface().(time.Time)
	switch dest.Kind() {
	case reflect.String:
		dest.SetString(srcVal.Format(time.RFC3339))
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		if dest.OverflowInt(srcVal.Unix()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(srcVal.Unix())
		return nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		if dest.OverflowUint(uint64(srcVal.Unix())) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(uint64(srcVal.Unix()))
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return NewInvalidMappingError(src.Type(), dest.Type(), "int8, int16, uint8, uint16 are too small")
	case reflect.Float32, reflect.Float64:
		unix := srcVal.Unix()
		nano := srcVal.Nanosecond()
		dest.SetFloat(float64(unix) + float64(nano)/1e9)
		return nil
	case reflect.Struct:
		switch dest.Type() {
		case timeTy:
			dest.Set(src)
			return nil
		case bigIntTy:
			dest.Set(reflect.ValueOf(big.NewInt(srcVal.Unix())).Elem())
			return nil
		case bigFloatTy:
			unix := srcVal.Unix()
			nano := srcVal.Nanosecond()
			bf := new(big.Float).SetInt64(unix)
			bn := new(big.Float).SetInt64(int64(nano))
			bn = bn.Quo(bn, big.NewFloat(1e9))
			bf = bf.Add(bf, bn)
			dest.Set(reflect.ValueOf(bf).Elem())
			return nil
		}
	}
	// If any of the above cases matched, try to map the time.Time to an
	// int64 and then try to map that.
	if err := m.MapRefl(reflect.ValueOf(src.Interface().(time.Time).Unix()), dest); err != nil {
		return NewInvalidMappingError(src.Type(), dest.Type(), "")
	}
	return nil
}

var mapTimeDest MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.String:
		tm, err := time.Parse(time.RFC3339, src.String())
		if err != nil {
			return NewInvalidMappingError(src.Type(), dest.Type(), err.Error())
		}
		dest.Set(reflect.ValueOf(tm))
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		dest.Set(reflect.ValueOf(time.Unix(src.Int(), 0).UTC()))
		return nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		dest.Set(reflect.ValueOf(time.Unix(int64(src.Uint()), 0).UTC()))
		return nil
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return NewInvalidMappingError(src.Type(), dest.Type(), "int8, int16, uint8, uint16 are too small")
	case reflect.Float32, reflect.Float64:
		unix, frac := math.Modf(src.Float())
		dest.Set(reflect.ValueOf(time.Unix(int64(unix), int64(frac*(1e9))).UTC()))
		return nil
	case reflect.Struct:
		switch src.Type() {
		case timeTy:
			dest.Set(src)
			return nil
		case bigIntTy:
			dest.Set(reflect.ValueOf(time.Unix(src.Addr().Interface().(*big.Int).Int64(), 0).UTC()))
			return nil
		case bigFloatTy:
			bf := src.Addr().Interface().(*big.Float)
			unix, _ := bf.Int(nil)
			frac := new(big.Float).Sub(bf, new(big.Float).SetInt(unix))
			nano, _ := frac.Mul(frac, big.NewFloat(1e9)).Int(nil)
			dest.Set(reflect.ValueOf(time.Unix(unix.Int64(), nano.Int64()).UTC()))
			return nil
		}
	}
	// If any of the above cases matched, try to map source to an int64 and
	// try to map that to the time.Time.
	var timestamp int64
	if err := m.MapRefl(src, reflect.ValueOf(&timestamp)); err != nil {
		return NewInvalidMappingError(src.Type(), dest.Type(), "")
	}
	dest.Set(reflect.ValueOf(time.Unix(timestamp, 0).UTC()))
	return nil
}

var mapBigIntSrc MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	srcVal := src.Addr().Interface().(*big.Int)
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(srcVal.Cmp(big.NewInt(0)) != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := srcVal.Int64()
		if dest.OverflowInt(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if srcVal.Sign() < 0 {
			return NewInvalidMappingError(src.Type(), dest.Type(), "negative value")
		}
		n := srcVal.Uint64()
		if dest.OverflowUint(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, _ := new(big.Float).SetInt(srcVal).Float64()
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(srcVal.String())
		return nil
	case reflect.Slice:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			dest.SetBytes(srcVal.Bytes())
			return nil
		}
	case reflect.Array:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			b := srcVal.Bytes()
			if len(b) > dest.Len() {
				return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
			}
			for i := 0; i < len(b); i++ {
				dest.Index(i).SetUint(uint64(b[i]))
			}
			return nil
		}
	case reflect.Struct:
		switch dest.Type() {
		case bigIntTy:
			dest.Set(src)
			return nil
		case bigFloatTy:
			dest.Set(reflect.ValueOf(new(big.Float).SetInt(srcVal)).Elem())
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

var mapBigIntDest MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			dest.Set(reflect.ValueOf(big.NewInt(1)).Elem())
		} else {
			dest.Set(reflect.ValueOf(big.NewInt(0)).Elem())
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dest.Set(reflect.ValueOf(new(big.Int).SetInt64(src.Int())).Elem())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dest.Set(reflect.ValueOf(new(big.Int).SetUint64(src.Uint())).Elem())
		return nil
	case reflect.Float32, reflect.Float64:
		bi, _ := new(big.Float).SetFloat64(src.Float()).Int(nil)
		dest.Set(reflect.ValueOf(new(big.Int).Set(bi)).Elem())
		return nil
	case reflect.String:
		bn, ok := new(big.Int).SetString(src.String(), 0)
		if !ok {
			return NewInvalidMappingError(src.Type(), dest.Type(), "invalid number")
		}
		dest.Set(reflect.ValueOf(bn).Elem())
		return nil
	case reflect.Slice:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			dest.Set(reflect.ValueOf(new(big.Int).SetBytes(src.Bytes())).Elem())
			return nil
		}
	case reflect.Array:
		if src.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, src.Len())
			for i := 0; i < src.Len(); i++ {
				b[i] = byte(src.Index(i).Uint())
			}
			dest.Set(reflect.ValueOf(new(big.Int).SetBytes(b)).Elem())
			return nil
		}
	case reflect.Struct:
		switch src.Type() {
		case bigIntTy:
			dest.Set(src)
			return nil
		case bigFloatTy:
			bf := src.Addr().Interface().(*big.Float)
			bi, _ := bf.Int(nil)
			dest.Set(reflect.ValueOf(bi).Elem())
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

var mapBigFloatSrc MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	srcVal := src.Addr().Interface().(*big.Float)
	switch dest.Kind() {
	case reflect.Bool:
		dest.SetBool(srcVal.Cmp(big.NewFloat(0)) != 0)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bn, _ := srcVal.Int(nil)
		if !bn.IsInt64() || dest.OverflowInt(bn.Int64()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetInt(bn.Int64())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bn, _ := srcVal.Int(nil)
		if !bn.IsUint64() || dest.OverflowUint(bn.Uint64()) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetUint(bn.Uint64())
		return nil
	case reflect.Float32, reflect.Float64:
		n, _ := srcVal.Float64()
		if dest.OverflowFloat(n) {
			return NewInvalidMappingError(src.Type(), dest.Type(), "overflow")
		}
		dest.SetFloat(n)
		return nil
	case reflect.String:
		dest.SetString(srcVal.String())
		return nil
	case reflect.Struct:
		switch dest.Type() {
		case bigIntTy:
			bi, _ := srcVal.Int(nil)
			dest.Set(reflect.ValueOf(bi).Elem())
			return nil
		case bigFloatTy:
			dest.Set(src)
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}

var mapBigFloatDest MapFunc = func(m *Mapper, src, dest reflect.Value) error {
	if m.StrictTypes && src.Type() != dest.Type() {
		return NewInvalidMappingError(src.Type(), dest.Type(), "strict mode")
	}
	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			dest.Set(reflect.ValueOf(big.NewFloat(1)).Elem())
		} else {
			dest.Set(reflect.ValueOf(big.NewFloat(0)).Elem())
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dest.Set(reflect.ValueOf(new(big.Float).SetInt64(src.Int())).Elem())
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dest.Set(reflect.ValueOf(new(big.Float).SetUint64(src.Uint())).Elem())
		return nil
	case reflect.Float32, reflect.Float64:
		dest.Set(reflect.ValueOf(new(big.Float).SetFloat64(src.Float())).Elem())
		return nil
	case reflect.String:
		bn, ok := new(big.Float).SetString(src.String())
		if !ok {
			return NewInvalidMappingError(src.Type(), dest.Type(), "invalid number")
		}
		dest.Set(reflect.ValueOf(bn).Elem())
		return nil
	case reflect.Struct:
		switch src.Type() {
		case bigIntTy:
			bn := src.Addr().Interface().(*big.Int)
			dest.Set(reflect.ValueOf(new(big.Float).SetInt(bn)).Elem())
			return nil
		case bigFloatTy:
			dest.Set(src)
			return nil
		}
	}
	return NewInvalidMappingError(src.Type(), dest.Type(), "")
}
