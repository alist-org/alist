package vtencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

func QSignatureKey(SecretKey, KeyTime string) string {
	var hashFunc = sha1.New
	h := hmac.New(hashFunc, []byte(SecretKey))
	h.Write([]byte(KeyTime))
	signKey := h.Sum(nil)
	return fmt.Sprintf("%x", signKey)
}
