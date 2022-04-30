package xunlei

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/url"
)

const (
	API_URL        = "https://api-pan.xunlei.com/drive/v1"
	FILE_API_URL   = API_URL + "/files"
	XLUSER_API_URL = "https://xluser-ssl.xunlei.com/v1"
)

const (
	FOLDER    = "drive#folder"
	FILE      = "drive#file"
	RESUMABLE = "drive#resumable"
)

const (
	UPLOAD_TYPE_UNKNOWN = "UPLOAD_TYPE_UNKNOWN"
	//UPLOAD_TYPE_FORM      = "UPLOAD_TYPE_FORM"
	UPLOAD_TYPE_RESUMABLE = "UPLOAD_TYPE_RESUMABLE"
	UPLOAD_TYPE_URL       = "UPLOAD_TYPE_URL"
)

func getAction(method string, u string) string {
	c, _ := url.Parse(u)
	return method + ":" + c.Path
}

// 计算文件Gcid
func getGcid(r io.Reader, size int64) (string, error) {
	calcBlockSize := func(j int64) int64 {
		if j >= 0 && j <= 0x8000000 {
			return 0x40000
		}
		if j <= 0x8000000 || j > 0x10000000 {
			if j <= 0x10000000 || j > 0x20000000 {
				return 0x200000
			}
			return 0x100000
		}
		return 0x80000
	}

	hash1 := sha1.New()
	hash2 := sha1.New()
	readSize := calcBlockSize(size)
	for {
		hash2.Reset()
		if n, err := io.CopyN(hash2, r, readSize); err != nil && n == 0 {
			if err != io.EOF {
				return "", err
			}
			break
		}
		hash1.Write(hash2.Sum(nil))
	}
	return hex.EncodeToString(hash1.Sum(nil)), nil
}
