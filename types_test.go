package anymapper

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	tm := time.Unix(1666666666, 0).UTC()
	t.Run("time-string", func(t *testing.T) {
		var dst string
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, "2022-10-25T02:57:46Z", dst)
	})
	t.Run("time-int8", func(t *testing.T) {
		var dst int8
		err := Map(tm, &dst)
		assert.Error(t, err)
	})
	t.Run("time-int16", func(t *testing.T) {
		var dst int16
		err := Map(tm, &dst)
		assert.Error(t, err)
	})
	t.Run("time-int32", func(t *testing.T) {
		var dst int32
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, int32(1666666666), dst)
	})
	t.Run("time-int64", func(t *testing.T) {
		var dst int64
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, int64(1666666666), dst)
	})
	t.Run("time-uint8", func(t *testing.T) {
		var dst uint8
		err := Map(tm, &dst)
		assert.Error(t, err)
	})
	t.Run("time-uint16", func(t *testing.T) {
		var dst uint16
		err := Map(tm, &dst)
		assert.Error(t, err)
	})
	t.Run("time-uint32", func(t *testing.T) {
		var dst uint32
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint32(1666666666), dst)
	})
	t.Run("time-uint64", func(t *testing.T) {
		var dst uint64
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1666666666), dst)
	})
	t.Run("time-float64", func(t *testing.T) {
		var dst float64
		err := Map(tm.Add(time.Millisecond*500), &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1666666666.5), dst)
	})
	t.Run("time-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x63, 0x57, 0x50, 0xaa}, dst)
	})
	t.Run("string-time", func(t *testing.T) {
		var dst time.Time
		err := Map("2022-10-25T02:57:46Z", &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
	t.Run("int8-time", func(t *testing.T) {
		var dst time.Time
		err := Map(int8(1), &dst)
		assert.Error(t, err)
	})
	t.Run("int16-time", func(t *testing.T) {
		var dst time.Time
		err := Map(int16(1), &dst)
		assert.Error(t, err)
	})
	t.Run("int32-time", func(t *testing.T) {
		var dst time.Time
		err := Map(int32(1666666666), &dst)
		assert.NoError(t, err)
		assert.Equal(t, time.Unix(1666666666, 0).UTC(), dst)
	})
	t.Run("int64-time", func(t *testing.T) {
		var dst time.Time
		err := Map(int64(1666666666), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
	t.Run("uint8-time", func(t *testing.T) {
		var dst time.Time
		err := Map(uint8(1), &dst)
		assert.Error(t, err)
	})
	t.Run("uint16-time", func(t *testing.T) {
		var dst time.Time
		err := Map(uint16(1), &dst)
		assert.Error(t, err)
	})
	t.Run("uint32-time", func(t *testing.T) {
		var dst time.Time
		err := Map(uint32(1666666666), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
	t.Run("uint64-time", func(t *testing.T) {
		var dst time.Time
		err := Map(uint64(1666666666), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
	t.Run("float64-time", func(t *testing.T) {
		var dst time.Time
		err := Map(float64(1666666666.5), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm.Add(time.Millisecond*500), dst)
	})
	t.Run("[]byte-time", func(t *testing.T) {
		var dst time.Time
		err := Map([]byte{0x63, 0x57, 0x50, 0xaa}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
}

func TestBigInt(t *testing.T) {
	t.Run("bigInt-bool#false", func(t *testing.T) {
		var dst bool
		err := Map(big.NewInt(0), &dst)
		assert.NoError(t, err)
		assert.Equal(t, false, dst)
	})
	t.Run("bigInt-bool#true", func(t *testing.T) {
		var dst bool
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("bigInt-int", func(t *testing.T) {
		var dst int
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("bigInt-uint", func(t *testing.T) {
		var dst uint
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("bigInt-float64", func(t *testing.T) {
		var dst float64
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), dst)
	})
	t.Run("bigInt-string", func(t *testing.T) {
		var dst string
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, "1", dst)
	})
	t.Run("bigInt-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01}, dst)
	})
	t.Run("bigInt-[4]byte", func(t *testing.T) {
		var dst [4]byte
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, [4]byte{0x01, 0x00, 0x00, 0x00}, dst)
	})
	t.Run("bigInt-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("bool-bigInt#true", func(t *testing.T) {
		var dst big.Int
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("bool-bigInt#false", func(t *testing.T) {
		var dst big.Int
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(0), &dst)
	})
	t.Run("int-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("uint-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("float64-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map(float64(1.5), &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("string-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map("1", &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("[]byte-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map([]byte{0x01}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
	t.Run("[4]byte-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map([4]byte{0x00, 0x00, 0x00, 0x01}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
}

func TestBigFloat(t *testing.T) {
	t.Run("bigFloat-bool#false", func(t *testing.T) {
		var dst bool
		err := Map(big.NewFloat(0), &dst)
		assert.NoError(t, err)
		assert.Equal(t, false, dst)
	})
	t.Run("bigFloat-bool#true", func(t *testing.T) {
		var dst bool
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, true, dst)
	})
	t.Run("bigFloat-int", func(t *testing.T) {
		var dst int
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1, dst)
	})
	t.Run("bigFloat-uint", func(t *testing.T) {
		var dst uint
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dst)
	})
	t.Run("bigFloat-float64", func(t *testing.T) {
		var dst float64
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), dst)
	})
	t.Run("bigFloat-string", func(t *testing.T) {
		var dst string
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, "1", dst)
	})
	t.Run("bigFloat-[]byte", func(t *testing.T) {
		var dst []byte
		err := Map(big.NewFloat(1), &dst)
		assert.Error(t, err)
	})
	t.Run("bigFloat-[4]byte", func(t *testing.T) {
		var dst [4]byte
		err := Map(big.NewFloat(1), &dst)
		assert.Error(t, err)
	})
	t.Run("bigFloat-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("bool-bigFloat#true", func(t *testing.T) {
		var dst big.Float
		err := Map(true, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("bool-bigFloat#false", func(t *testing.T) {
		var dst big.Float
		err := Map(false, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(0).Cmp(&dst))
	})
	t.Run("int-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map(1, &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("uint-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map(uint(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("float64-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map(float64(1.5), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1.5).Cmp(&dst))
	})
	t.Run("string-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map("1", &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("[]byte-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map([]byte{0x01}, &dst)
		assert.Error(t, err)
	})
	t.Run("[4]byte-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map([4]byte{0x00, 0x00, 0x00, 0x01}, &dst)
		assert.Error(t, err)
	})
}

func TestCombinations(t *testing.T) {
	tm := time.Unix(1666666666, 0).UTC()
	t.Run("time-bigInt", func(t *testing.T) {
		var dst *big.Int
		err := Map(tm, &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1666666666), dst)
	})
	t.Run("time-bigFloat", func(t *testing.T) {
		var dst *big.Float
		err := Map(tm.Add(time.Millisecond*500), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1666666666.5).Cmp(dst))
	})
	t.Run("bigInt-time", func(t *testing.T) {
		var dst time.Time
		err := Map(big.NewInt(1666666666), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm, dst)
	})
	t.Run("bigInt-bigFloat", func(t *testing.T) {
		var dst big.Float
		err := Map(big.NewInt(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, 0, big.NewFloat(1).Cmp(&dst))
	})
	t.Run("bigFloat-time", func(t *testing.T) {
		var dst time.Time
		err := Map(big.NewFloat(1666666666.5), &dst)
		assert.NoError(t, err)
		assert.Equal(t, tm.Add(time.Millisecond*500).UTC(), dst)
	})
	t.Run("bigFloat-bigInt", func(t *testing.T) {
		var dst big.Int
		err := Map(big.NewFloat(1), &dst)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1), &dst)
	})
}
