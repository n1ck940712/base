package types

import (
	"fmt"
	"math/rand"
)

func (i Int) Bytes() Bytes {
	return i.String().Bytes()
}

func (i Int) String() String {
	return String(fmt.Sprint(i))
}

func (i Int) Ptr() *int {
	return (*int)(&i)
}

func (i Int) Int8() int8 {
	return int8(i)
}

func (i Int) Int16() int16 {
	return int16(i)
}

func (i Int) Int32() int32 {
	return int32(i)
}

func (i Int) Int64() int64 {
	return int64(i)
}

func (i Int) Float() Float {
	return Float(i)
}

func (i Int) Bool() bool {
	return i > 0
}

func (i Int) GetASCII() string {
	aToZ := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return fmt.Sprintf("%c", aToZ[i])
}

func (i Int) LeadingZeroes(width int) string {
	return fmt.Sprintf("%0*d", width, i)
}

func (i Int) RandSeq() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, i)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
