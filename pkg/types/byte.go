package types

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

func (b Bytes) SHA256() string {
	hash := sha256.Sum256(b)

	return fmt.Sprintf("%x", hash[:])
}

func (b Bytes) MD5() string {
	hash := md5.Sum(b)

	return fmt.Sprintf("%x", hash[:])
}

func (b Bytes) Ptr() *[]byte {
	return (*[]byte)(&b)
}

func (b Bytes) String() String {
	return String(b)
}

func (b Bytes) Int() Int {
	return b.String().Int()
}

func (b Bytes) Float() Float {
	return b.String().Float()
}

func (b Bytes) Bool() bool {
	return b.String().Bool()
}

func (b Bytes) Update(call func([]byte) []byte) Bytes {
	return Bytes(call(b))
}
