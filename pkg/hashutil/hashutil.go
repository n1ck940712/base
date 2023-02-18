package hashutil

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type SliceIndex struct {
	PrevIndex int
	PostIndex int
}

type Hash interface {
	Raw() string
	Extract(index int, size int) Hash
	Int64() int64
	Error() error
}

type hash struct {
	raw string
	err error
}

func NewHash(raw string) Hash {
	return &hash{raw: raw}
}

func (h *hash) Raw() string {
	return h.raw
}

func (h *hash) Extract(index int, size int) Hash {
	if maxSize := index + size; maxSize <= len(h.raw) {
		return NewHash(h.raw[index:maxSize])
	}
	eHash := hash{raw: ""}

	eHash.err = errors.New("(index + size) is out of range of hash length")
	return &eHash
}

func (h *hash) Int64() int64 {
	value, err := strconv.ParseInt(h.raw, 16, 64)

	if err != nil {
		h.err = errors.New("hash parseInt error: " + err.Error())
		return 0
	}
	return value
}

func (h *hash) Error() error {
	return h.err
}

func GenerateHash(seeds ...string) string {
	seed := ""

	if len(seeds) > 0 {
		seed = seeds[0]
	}
	str := time.Now().String() + "h@sh-go-mg--r4pid" + seed
	hash256 := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%x", hash256)
}
