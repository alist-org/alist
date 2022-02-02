package baidu

import (
	"net/url"
	"strings"
	"time"
)

func getTime(t int64) *time.Time {
	tm := time.Unix(t, 0)
	return &tm
}

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.ReplaceAll(r, "+", "%20")
	return r
}
