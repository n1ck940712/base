package types

import (
	"fmt"
	"math"
)

func (f Float) String() String {
	return String(fmt.Sprint(f))
}

func (f Float) Int() Int {
	return Int(f)
}

func (i Float) Ptr() *float64 {
	return (*float64)(&i)
}

func (f Float) Float64() float64 {
	return float64(f)
}

func (f Float) Float32() float32 {
	return float32(f)
}

func (f Float) Round(precision int) Float {
	mul := math.Pow10(precision)

	return Float(float64(int(math.Round(f.Float64()*mul))) / mul)
}

func (f Float) Fixed(precision int) Float {
	mul := math.Pow10(precision)

	return Float(math.Floor((f.Float64()*mul)+0.000000001) / mul)
}

func (f Float) FixedStr(precision int) String {
	return f.Fixed(precision).String()
}

func (f Float) Update(call func(float64) float64) Float {
	return Float(call(f.Float64()))
}
