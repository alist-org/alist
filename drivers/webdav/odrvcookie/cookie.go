package odrvcookie

import (
	"github.com/Xhofe/alist/utils/cookie"
	"net/http"
	"sync"
	"time"
)

type SpCookie struct {
	Cookie string
	expire time.Time
}

func (sp SpCookie) IsExpire() bool {
	return time.Now().Before(sp.expire)
}

var cookiesMap = struct {
	sync.Mutex
	m map[string]*SpCookie
}{m: make(map[string]*SpCookie)}

func GetCookie(username, password, siteUrl string) (string, error) {
	cookiesMap.Lock()
	defer cookiesMap.Unlock()
	spCookie, ok := cookiesMap.m[username]
	if ok {
		if !spCookie.IsExpire() {
			return spCookie.Cookie, nil
		}
	}
	ca := New(username, password, siteUrl)
	tokenConf, err := ca.Cookies()
	if err != nil {
		return "", err
	}
	spCookie = &SpCookie{
		Cookie: cookie.ToString([]*http.Cookie{&tokenConf.RtFa, &tokenConf.FedAuth}),
		expire: time.Now().Add(time.Hour * 12),
	}
	cookiesMap.m[username] = spCookie
	return spCookie.Cookie, nil
}
