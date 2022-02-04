package utils

import (
	"math/rand"
	"strings"
	"time"
)

var Rand *rand.Rand

func init() {
	s := rand.NewSource(time.Now().UnixNano())
	Rand = rand.New(s)
}

func RandomStr(n int) string {

	builder := strings.Builder{}
	t := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < n; i++ {
		r := Rand.Intn(len(t))
		builder.WriteString(t[r : r+1])
	}
	return builder.String()
}
