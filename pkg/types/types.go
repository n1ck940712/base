package types

import "encoding/json"

// type Any interface{}
type Bytes []byte
type String string
type Int int
type Float float64
type JSONRaw json.RawMessage
type Odds float64
type Bool bool

type Array[T any] []T
