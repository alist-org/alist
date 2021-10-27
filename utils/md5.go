package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// GetMD5Encode
func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Get16MD5Encode
func Get16MD5Encode(data string) string {
	return GetMD5Encode(data)[8:24]
}
