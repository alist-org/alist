package vtencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func QSignatureKey(timeKey string, signPath string, key string) string {
	signKey := hmac.New(sha1.New, []byte(key))
	signKey.Write([]byte(timeKey))
	signKeyBytes := signKey.Sum(nil)
	signKeyHex := hex.EncodeToString(signKeyBytes)
	sha := sha1.New()
	sha.Write([]byte(signPath))
	shaBytes := sha.Sum(nil)
	shaHex := hex.EncodeToString(shaBytes)

	O := "sha1\n" + timeKey + "\n" + shaHex + "\n"
	dataSignKey := hmac.New(sha1.New, []byte(signKeyHex))
	dataSignKey.Write([]byte(O))
	dataSignKeyBytes := dataSignKey.Sum(nil)
	dataSignKeyHex := hex.EncodeToString(dataSignKeyBytes)
	return dataSignKeyHex
}

func QTwoSignatureKey(timeKey string, key string) string {
	signKey := hmac.New(sha1.New, []byte(key))
	signKey.Write([]byte(timeKey))
	signKeyBytes := signKey.Sum(nil)
	signKeyHex := hex.EncodeToString(signKeyBytes)
	return signKeyHex
}
