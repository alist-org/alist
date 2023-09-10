package _123Link

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

func SignURL(originURL, privateKey string, uid uint64, validDuration time.Duration) (newURL string, err error) {
	if privateKey == "" {
		return originURL, nil
	}
	var (
		ts     = time.Now().Add(validDuration).Unix() // 有效时间戳
		rInt   = rand.Int()                           // 随机正整数
		objURL *url.URL
	)
	objURL, err = url.Parse(originURL)
	if err != nil {
		return "", err
	}
	authKey := fmt.Sprintf("%d-%d-%d-%x", ts, rInt, uid, md5.Sum([]byte(fmt.Sprintf("%s-%d-%d-%d-%s",
		objURL.Path, ts, rInt, uid, privateKey))))
	v := objURL.Query()
	v.Add("auth_key", authKey)
	objURL.RawQuery = v.Encode()
	return objURL.String(), nil
}
