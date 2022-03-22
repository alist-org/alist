package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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

func Get16MD5Encode(data string) string {
	return GetMD5Encode(data)[8:24]
}

func SignWithPassword(name, password string) string {
	return Get16MD5Encode(fmt.Sprintf("alist-%s-%s", password, name))
}

func SignWithToken(name, token string) string {
	return Get16MD5Encode(fmt.Sprintf("alist-%s-%s", token, name))
}
