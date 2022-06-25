package random

import (
	"math/rand"
	"time"
)

var Rand *rand.Rand

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func RandomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[Rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	s := rand.NewSource(time.Now().UnixNano())
	Rand = rand.New(s)
}
