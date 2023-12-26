package baiduphoto

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/pkg/utils"
)

// Tid生成
func getTid() string {
	return fmt.Sprintf("3%d%.0f", time.Now().Unix(), math.Floor(9000000*rand.Float64()+1000000))
}

func toTime(t int64) *time.Time {
	tm := time.Unix(t, 0)
	return &tm
}

func fsidsFormatNotUk(ids ...int64) string {
	buf := utils.MustSliceConvert(ids, func(id int64) string {
		return fmt.Sprintf(`{"fsid":%d}`, id)
	})
	return fmt.Sprintf("[%s]", strings.Join(buf, ","))
}

func getFileName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}

func MustString(str string, err error) string {
	return str
}

/*
*	处理文件变化
*	最大程度利用重复数据
**/
func copyFile(file *AlbumFile, cf *CopyFile) *File {
	return &File{
		Fsid:     cf.Fsid,
		Path:     cf.Path,
		Ctime:    cf.Ctime,
		Mtime:    cf.Ctime,
		Size:     file.Size,
		Thumburl: file.Thumburl,
	}
}

func moveFileToAlbumFile(file *File, album *Album, uk int64) *AlbumFile {
	return &AlbumFile{
		File:    *file,
		AlbumID: album.AlbumID,
		Tid:     album.Tid,
		Uk:      uk,
	}
}

func renameAlbum(album *Album, newName string) *Album {
	return &Album{
		AlbumID:      album.AlbumID,
		Tid:          album.Tid,
		JoinTime:     album.JoinTime,
		CreationTime: album.CreationTime,
		Title:        newName,
		Mtime:        time.Now().Unix(),
	}
}

func BoolToIntStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
