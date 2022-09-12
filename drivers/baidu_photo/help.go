package baiduphoto

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

//Tid生成
func getTid() string {
	return fmt.Sprintf("3%d%.0f", time.Now().Unix(), math.Floor(9000000*rand.Float64()+1000000))
}

// 检查名称
func checkName(name string) bool {
	return len(name) <= 20 && regexp.MustCompile("[\u4e00-\u9fa5A-Za-z0-9_-]").MatchString(name)
}

func toTime(t int64) *time.Time {
	tm := time.Unix(t, 0)
	return &tm
}

func fsidsFormat(ids ...string) string {
	var buf []string
	for _, id := range ids {
		e := splitID(id)
		buf = append(buf, fmt.Sprintf(`{"fsid":%s,"uk":%s}`, e[0], e[3]))
	}
	return fmt.Sprintf("[%s]", strings.Join(buf, ","))
}

func fsidsFormatNotUk(ids ...string) string {
	var buf []string
	for _, id := range ids {
		buf = append(buf, fmt.Sprintf(`{"fsid":%s}`, splitID(id)[0]))
	}
	return fmt.Sprintf("[%s]", strings.Join(buf, ","))
}

/*
结构

{fsid} 文件

{album_id}|{tid} 相册

{fsid}|{album_id}|{tid}|{uk} 相册文件
*/
func splitID(id string) []string {
	return strings.SplitN(id, "|", 4)[:4]
}

/*
结构

{fsid} 文件

{album_id}|{tid} 相册

{fsid}|{album_id}|{tid}|{uk} 相册文件
*/
func joinID(ids ...interface{}) string {
	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprint(id))
	}
	return strings.Join(idsStr, "|")
}

func getFileName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}

// 相册
func IsAlbum(obj model.Obj) bool {
	return obj.IsDir() && obj.GetPath() == "album"
}

// 根目录
func IsRoot(obj model.Obj) bool {
	return obj.IsDir() && obj.GetPath() == "" && obj.GetID() == ""
}

// 以相册为根目录
func IsAlbumRoot(obj model.Obj) bool {
	return obj.IsDir() && obj.GetPath() == "" && obj.GetID() != ""
}

// 根文件
func IsFile(obj model.Obj) bool {
	return !obj.IsDir() && obj.GetPath() == "file"
}

// 相册文件
func IsAlbumFile(obj model.Obj) bool {
	return !obj.IsDir() && obj.GetPath() == "albumfile"
}

func MustString(str string, err error) string {
	return str
}
