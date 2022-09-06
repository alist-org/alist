package _189

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	myrand "github.com/alist-org/alist/v3/pkg/utils/random"
	log "github.com/sirupsen/logrus"
)

func random() string {
	return fmt.Sprintf("0.%17v", myrand.Rand.Int63n(100000000000000000))
}

func RsaEncode(origData []byte, j_rsakey string, hex bool) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		log.Errorf("err: %s", err.Error())
	}
	res := base64.StdEncoding.EncodeToString(b)
	if hex {
		return b64tohex(res)
	}
	return res
}

var b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz"

func int2char(a int) string {
	return strings.Split(BI_RM, "")[a]
}

func b64tohex(a string) string {
	d := ""
	e := 0
	c := 0
	for i := 0; i < len(a); i++ {
		m := strings.Split(a, "")[i]
		if m != "=" {
			v := strings.Index(b64map, m)
			if 0 == e {
				e = 1
				d += int2char(v >> 2)
				c = 3 & v
			} else if 1 == e {
				e = 2
				d += int2char(c<<2 | v>>4)
				c = 15 & v
			} else if 2 == e {
				e = 3
				d += int2char(c)
				d += int2char(v >> 2)
				c = 3 & v
			} else {
				e = 0
				d += int2char(c<<2 | v>>4)
				d += int2char(15 & v)
			}
		}
	}
	if e == 1 {
		d += int2char(c << 2)
	}
	return d
}

func qs(form map[string]string) string {
	f := make(url.Values)
	for k, v := range form {
		f.Set(k, v)
	}
	return EncodeParam(f)
	//strList := make([]string, 0)
	//for k, v := range form {
	//	strList = append(strList, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	//}
	//return strings.Join(strList, "&")
}

func EncodeParam(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	for _, k := range keys {
		vs := v[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			//if k == "fileName" {
			//	buf.WriteString(encode(v))
			//} else {
			buf.WriteString(v)
			//}
		}
	}
	return buf.String()
}

func encode(str string) string {
	//str = strings.ReplaceAll(str, "%", "%25")
	//str = strings.ReplaceAll(str, "&", "%26")
	//str = strings.ReplaceAll(str, "+", "%2B")
	//return str
	return url.QueryEscape(str)
}

func AesEncrypt(data, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	if block == nil {
		return []byte{}
	}
	data = PKCS7Padding(data, block.BlockSize())
	decrypted := make([]byte, len(data))
	size := block.BlockSize()
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], data[bs:be])
	}
	return decrypted
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func hmacSha1(data string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func getMd5(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func decodeURIComponent(str string) string {
	r, _ := url.PathUnescape(str)
	//r = strings.ReplaceAll(r, " ", "+")
	return r
}

func Random(v string) string {
	reg := regexp.MustCompilePOSIX("[xy]")
	data := reg.ReplaceAllFunc([]byte(v), func(msg []byte) []byte {
		var i int64
		t := int64(16 * myrand.Rand.Float32())
		if msg[0] == 120 {
			i = t
		} else {
			i = 3&t | 8
		}
		return []byte(strconv.FormatInt(i, 16))
	})
	return string(data)
}
