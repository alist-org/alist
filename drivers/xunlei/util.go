package xunlei

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"

	"github.com/Xhofe/alist/utils"
)

const (
	// 小米浏览器
	CLIENT_ID      = "X7MtiU0Gb5YqWv-6"
	CLIENT_SECRET  = "84MYEih3Eeu2HF4RrGce3Q"
	CLIENT_VERSION = "5.1.0.51045"

	ALG_VERSION  = "1"
	PACKAGE_NAME = "com.xunlei.xcloud.lib"
)

var Algorithms = []string{
	"",
	"BXza40wm+P4zw8rEFpHA",
	"UfZLfKfYRmKTA0",
	"OMBGVt/9Wcaln1XaBz",
	"Jn217F4rk5FPPWyhoeV",
	"w5OwkGo0pGpb0Xe/XZ5T3",
	"5guM3DNiY4F78x49zQ97q75",
	"QXwn4D2j884wJgrYXjGClM/IVrJX",
	"NXBRosYvbHIm6w8vEB",
	"2kZ8Ie1yW2ib4O2iAkNpJobP",
	"11CoVJJQEc",
	"xf3QWysVwnVsNv5DCxU+cgNT1rK",
	"9eEfKkrqkfw",
	"T78dnANexYRbiZy",
}

const (
	API_URL        = "https://api-pan.xunlei.com/drive/v1"
	FILE_API_URL   = API_URL + "/files"
	XLUSER_API_URL = "https://xluser-ssl.xunlei.com/v1"
)

const (
	FOLDER = "drive#folder"
	FILE   = "drive#file"

	RESUMABLE = "drive#resumable"
)

const (
	UPLOAD_TYPE_UNKNOWN = "UPLOAD_TYPE_UNKNOWN"
	//UPLOAD_TYPE_FORM      = "UPLOAD_TYPE_FORM"
	UPLOAD_TYPE_RESUMABLE = "UPLOAD_TYPE_RESUMABLE"
	UPLOAD_TYPE_URL       = "UPLOAD_TYPE_URL"
)

func captchaSign(driverID string, time int64) string {
	str := fmt.Sprint(CLIENT_ID, CLIENT_VERSION, PACKAGE_NAME, driverID, time)
	for _, algorithm := range Algorithms {
		str = utils.GetMD5Encode(fmt.Sprint(str, algorithm))
	}
	return fmt.Sprint(ALG_VERSION, ".", str)
}

func getAction(method string, u string) string {
	c, _ := url.Parse(u)
	return fmt.Sprint(method, ":", c.Path)
}

func getGcid(r io.Reader, size int64) (string, error) {
	calcBlockSize := func(j int64) int64 {
		if j >= 0 && j <= 134217728 {
			return 262144
		}
		if j <= 134217728 || j > 268435456 {
			if j <= 268435456 || j > 536870912 {
				return 2097152
			}
			return 1048576
		}
		return 524288
	}
	/*
		calcBlockSize := func(j int64) int64 {
			psize := int64(0x40000)
			for j/psize > 0x200 {
				psize <<= 1
			}
			return psize
		}
	*/

	hash1 := sha1.New()
	hash2 := sha1.New()
	for {
		hash2.Reset()
		if n, err := io.CopyN(hash2, r, calcBlockSize(size)); err != nil && n == 0 {
			if err != io.EOF {
				return "", err
			}
			break
		}
		hash1.Write(hash2.Sum(nil))
	}
	return hex.EncodeToString(hash1.Sum(nil)), nil
}
