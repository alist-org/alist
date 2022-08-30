package random

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var Rand *rand.Rand

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func String(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[Rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Token() string {
	return "alist-" + uuid.NewString() + String(64)
}

func RangeInt64(left, right int64) int64 {
	return rand.Int63n(left+right) - left
}

func init() {
	s := rand.NewSource(time.Now().UnixNano())
	Rand = rand.New(s)
}
