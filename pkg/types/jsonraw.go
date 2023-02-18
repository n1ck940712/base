package types

import (
	"strings"
)

func (jsonR *JSONRaw) RawString() string {
	return string(*jsonR)
}

func (jr *JSONRaw) String() string {
	return string(jr.cleanString())
}

func (jr *JSONRaw) Int() int {
	return int(jr.cleanString().Int())
}

func (jr *JSONRaw) Int8() int8 {
	return int8(jr.Int())
}

func (jr *JSONRaw) Int16() int16 {
	return int16(jr.Int())
}

func (jr *JSONRaw) Int32() int32 {
	return int32(jr.Int())
}

func (jr *JSONRaw) Int64() int64 {
	return int64(jr.Int())
}

func (jr *JSONRaw) ToFloat32() float32 {
	return float32(jr.Int())
}

func (jr *JSONRaw) ToFloat64() float64 {
	return float64(jr.Int())
}

func (jr *JSONRaw) cleanString() String {
	if jr.RawString() == "null" {
		return ""
	}
	return String(strings.ReplaceAll(jr.RawString(), `"`, ""))
}
