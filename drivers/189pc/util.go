package _189

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	rand2 "math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Xhofe/alist/model"
)

const (
	APP_ID      = "8025431004"
	CLIENT_TYPE = "10020"
	VERSION     = "6.2"

	WEB_URL    = "https://cloud.189.cn"
	AUTH_URL   = "https://open.e.189.cn"
	API_URL    = "https://api.cloud.189.cn"
	UPLOAD_URL = "https://upload.cloud.189.cn"

	RETURN_URL = "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html"

	PC  = "TELEPC"
	MAC = "TELEMAC"

	CHANNEL_ID = "web_cloud.189.cn"
)

func clientSuffix() map[string]string {
	return map[string]string{
		"clientType": PC,
		"version":    VERSION,
		"channelId":  CHANNEL_ID,
		"rand":       fmt.Sprintf("%d_%d", rand2.Int63n(1e5), rand2.Int63n(1e10)),
	}
}

// 带params的SignatureOfHmac HMAC签名
func signatureOfHmac(sessionSecret, sessionKey, operate, fullUrl, dateOfGmt, param string) string {
	u, _ := url.Parse(fullUrl)
	mac := hmac.New(sha1.New, []byte(sessionSecret))
	data := fmt.Sprintf("SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s", sessionKey, operate, u.Path, dateOfGmt)
	if param != "" {
		data += fmt.Sprintf("&params=%s", param)
	}
	mac.Write([]byte(data))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

// 获取http规范的时间
func getHttpDateStr() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

// RAS 加密用户名密码
func rsaEncrypt(publicKey, origData string) string {
	block, _ := pem.Decode([]byte(publicKey))
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	data, _ := rsa.EncryptPKCS1v15(rand.Reader, pubInterface.(*rsa.PublicKey), []byte(origData))
	return base64ToHex(base64.StdEncoding.EncodeToString(data))
}

// aes 加密params
func AesECBEncrypt(data, key string) string {
	block, _ := aes.NewCipher([]byte(key))
	paddingData := PKCS7Padding([]byte(data), block.BlockSize())
	decrypted := make([]byte, len(paddingData))
	size := block.BlockSize()
	for src, dst := paddingData, decrypted; len(src) > 0; src, dst = src[size:], dst[size:] {
		block.Encrypt(dst[:size], src[:size])
	}
	return strings.ToUpper(hex.EncodeToString(decrypted))
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 时间戳
func timestamp() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}

func base64ToHex(a string) string {
	v, _ := base64.StdEncoding.DecodeString(a)
	return strings.ToUpper(hex.EncodeToString(v))
}

func isFamily(account *model.Account) bool {
	return account.InternalType == "Family"
}

func toFamilyOrderBy(o string) string {
	switch o {
	case "filename":
		return "1"
	case "filesize":
		return "2"
	case "lastOpTime":
		return "3"
	default:
		return "1"
	}
}

func MapToUrlValues(m map[string]interface{}) url.Values {
	url := make(url.Values, len(m))
	for k, v := range m {
		url.Add(k, fmt.Sprint(v))
	}
	return url
}

func decodeURIComponent(str string) string {
	r, _ := url.QueryUnescape(str)
	//r, _ := url.PathUnescape(str)
	//r = strings.ReplaceAll(r, " ", "+")
	return r
}

func MustToBytes(b []byte, err error) []byte {
	return b
}

func BoolToNumber(b bool) int {
	if b {
		return 1
	}
	return 0
}
