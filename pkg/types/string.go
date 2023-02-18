package types

import (
	"encoding/json"
	"strconv"
	"strings"
)

func (s String) Bytes() Bytes {
	return Bytes(s)
}

func (s String) Ptr() *string {
	return (*string)(&s)
}

func (s String) Int() Int {
	return Int(s.Float())
}

func (s String) Float() Float {
	return Float(s.parseFloat(64))
}

func (s String) Bool() bool {
	if s == "t" || s == "T" || s == "true" || s == "True" || s == "TRUE" {
		return true
	}
	return false
}

func (s String) JSON() any {
	var jsonObj any

	json.Unmarshal(s.Bytes(), &jsonObj)
	return jsonObj
}

func (s String) Mask(char string, leading int, tailing int) String {
	sLen := len(s)
	slead := String(strings.Repeat(char, int(leading)))

	if sLen > tailing {
		return slead + s[sLen-tailing:]
	} else if sLen > leading {
		return slead + s[leading:]
	}
	return slead + s
}

//Splits string with separator
func (s String) Split(sep string) []String {
	sSR := Array[String]{}
	ss := strings.Split(string(s), sep)

	for _, s := range ss {
		sSR = append(sSR, String(s))
	}
	return sSR
}

//Trims leading and tailing spaces
func (s String) TrimSpace() String {
	return String(strings.TrimSpace(string(s)))
}

//Updates raw string
func (s String) Update(call func(string) string) String {
	return String(call(string(s)))
}

func (s String) parseFloat(bitSize int) float64 {
	f, _ := strconv.ParseFloat(string(s), bitSize)

	return f
}
