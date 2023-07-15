package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func GetSHA1Encode(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func GetSHA256Encode(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func GetMD5Encode(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func GetMD5EncodeStr(data string) string {
	return GetMD5Encode([]byte(data))
}

var DEC = map[string]string{
	"-": "+",
	"_": "/",
	".": "=",
}

func SafeAtob(data string) (string, error) {
	for k, v := range DEC {
		data = strings.ReplaceAll(data, k, v)
	}
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}
