package netease_music

import (
	"regexp"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Cookie    string `json:"cookie" type:"text" required:"true" help:""`
	SongLimit uint64 `json:"song_limit" default:"200" type:"number" help:"only get 200 songs by default"`
}

func (ad *Addition) getCookie(name string) string {
	re := regexp.MustCompile(name + "=([^(;|$)]+)")
	matches := re.FindStringSubmatch(ad.Cookie)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

var config = driver.Config{
	Name: "NeteaseMusic",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &NeteaseMusic{}
	})
}
