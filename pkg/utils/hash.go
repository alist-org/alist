package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"strings"
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
