package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

func GetSHA1Encode(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
