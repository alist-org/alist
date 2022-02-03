package utils

import (
	"math/rand"
	"strings"
)

func RandomStr(n int) string {
	builder := strings.Builder{}
	t := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < n; i++ {
		r := rand.Intn(len(t))
		builder.WriteString(t[r : r+1])
	}
	return builder.String()
}
