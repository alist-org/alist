package netease_music

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"strings"

	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
)

var (
	linuxapiKey = []byte("rFgB&h#%2?^eDg:Q")
	eapiKey     = []byte("e82ckenh8dichen8")
	iv          = []byte("0102030405060708")
	presetKey   = []byte("0CoJUm6Qyw8W8jud")
	publicKey   = []byte("-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDgtQn2JZ34ZC28NWYpAUd98iZ37BUrX/aKzmFbt7clFSs6sXqHauqKWqdtLkF2KexO40H1YTX8z2lSgBBOAxLsvaklV8k4cBFK9snQXE9/DDaFt6Rr7iVZMldczhC0JNgTz+SHXT6CBHuX3e9SdB1Ua44oncaTWz7OBGLbCiK45wIDAQAB\n-----END PUBLIC KEY-----")
	stdChars    = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func aesKeyPending(key []byte) []byte {
	k := len(key)
	count := 0
	switch true {
	case k <= 16:
		count = 16 - k
	case k <= 24:
		count = 24 - k
	case k <= 32:
		count = 32 - k
	default:
		return key[:32]
	}
	if count == 0 {
		return key
	}

	return append(key, bytes.Repeat([]byte{0}, count)...)
}

func pkcs7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func aesCBCEncrypt(src, key, iv []byte) []byte {
	block, _ := aes.NewCipher(aesKeyPending(key))
	src = pkcs7Padding(src, block.BlockSize())
	dst := make([]byte, len(src))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(dst, src)

	return dst
}

func aesECBEncrypt(src, key []byte) []byte {
	block, _ := aes.NewCipher(aesKeyPending(key))

	src = pkcs7Padding(src, block.BlockSize())
	dst := make([]byte, len(src))

	ecbCryptBlocks(block, dst, src)

	return dst
}

func ecbCryptBlocks(block cipher.Block, dst, src []byte) {
	bs := block.BlockSize()

	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
}

func rsaEncrypt(buffer, key []byte) []byte {
	buffers := make([]byte, 128-16, 128)
	buffers = append(buffers, buffer...)
	block, _ := pem.Decode(key)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	c := new(big.Int).SetBytes([]byte(buffers))
	return c.Exp(c, big.NewInt(int64(pub.E)), pub.N).Bytes()
}

func getSecretKey() ([]byte, []byte) {
	key := make([]byte, 16)
	reversed := make([]byte, 16)
	for i := 0; i < 16; i++ {
		result := stdChars[random.RangeInt64(0, 62)]
		key[i] = result
		reversed[15-i] = result
	}
	return key, reversed
}

func weapi(data map[string]string) map[string]string {
	text, _ := utils.Json.Marshal(data)
	secretKey, reversedKey := getSecretKey()
	params := []byte(base64.StdEncoding.EncodeToString(aesCBCEncrypt(text, presetKey, iv)))
	return map[string]string{
		"params":    base64.StdEncoding.EncodeToString(aesCBCEncrypt(params, reversedKey, iv)),
		"encSecKey": hex.EncodeToString(rsaEncrypt(secretKey, publicKey)),
	}
}

func eapi(url string, data map[string]interface{}) map[string]string {
	text, _ := utils.Json.Marshal(data)
	msg := "nobody" + url + "use" + string(text) + "md5forencrypt"
	h := md5.New()
	h.Write([]byte(msg))
	digest := hex.EncodeToString(h.Sum(nil))
	params := []byte(url + "-36cd479b6b5-" + string(text) + "-36cd479b6b5-" + digest)
	return map[string]string{
		"params": hex.EncodeToString(aesECBEncrypt(params, eapiKey)),
	}
}

func linuxapi(data map[string]interface{}) map[string]string {
	text, _ := utils.Json.Marshal(data)
	return map[string]string{
		"eparams": strings.ToUpper(hex.EncodeToString(aesECBEncrypt(text, linuxapiKey))),
	}
}
