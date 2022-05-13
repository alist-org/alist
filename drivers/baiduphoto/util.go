package baiduphoto

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/Xhofe/alist/model"
)

const (
	API_URL         = "https://photo.baidu.com/youai"
	ALBUM_API_URL   = API_URL + "/album/v1"
	FILE_API_URL_V1 = API_URL + "/file/v1"
	FILE_API_URL_V2 = API_URL + "/file/v2"
)

var (
	ErrNotSupportName = errors.New("only chinese and english, numbers and underscores are supported, and the length is no more than 20")
)

//Tid生成
func getTid() string {
	return fmt.Sprintf("3%d%.0f", time.Now().Unix(), math.Floor(9000000*rand.Float64()+1000000))
}

// 检查名称
func checkName(name string) bool {
	return len(name) <= 20 && regexp.MustCompile("[\u4e00-\u9fa5A-Za-z0-9_]").MatchString(name)
}

func getTime(t int64) *time.Time {
	tm := time.Unix(t, 0)
	return &tm
}

func fsidsFormat(ids ...string) string {
	var buf []string
	for _, id := range ids {
		e := strings.Split(id, "|")
		buf = append(buf, fmt.Sprintf("{\"fsid\":%s,\"uk\":%s}", e[0], e[1]))
	}
	return fmt.Sprintf("[%s]", strings.Join(buf, ","))
}

func fsidsFormatNotUk(ids ...string) string {
	var buf []string
	for _, id := range ids {
		buf = append(buf, fmt.Sprintf("{\"fsid\":%s}", strings.Split(id, "|")[0]))
	}
	return fmt.Sprintf("[%s]", strings.Join(buf, ","))
}

func splitID(id string) []string {
	return strings.SplitN(id, "|", 3)[:3]
}

func joinID(ids ...interface{}) string {
	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprint(id))
	}
	return strings.Join(idsStr, "|")
}

func IsAlbum(file *model.File) bool {
	return file.Id != "" && file.IsDir()
}

func IsAlbumFile(file *model.File) bool {
	return file.Id != "" && !file.IsDir()
}

func IsRoot(file *model.File) bool {
	return file.Id == "" && file.IsDir()
}

func MustString(str string, err error) string {
	return str
}
