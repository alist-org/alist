package google

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"path"
	"strconv"
	"time"
)

type File struct {
	Id            string     `json:"id"`
	Name          string     `json:"name"`
	MimeType      string     `json:"mimeType"`
	ModifiedTime  *time.Time `json:"modifiedTime"`
	Size          string     `json:"size"`
	ThumbnailLink string     `json:"thumbnailLink"`
}

func (f File) GetSize() uint64 {
	if f.GetType() == conf.FOLDER {
		return 0
	}
	size, _ := strconv.ParseUint(f.Size, 10, 64)
	return size
}

func (f File) GetName() string {
	return f.Name
}

func (f File) GetType() int {
	mimeType := f.MimeType
	if mimeType == "application/vnd.google-apps.folder" || mimeType == "application/vnd.google-apps.shortcut" {
		return conf.FOLDER
	}
	return utils.GetFileType(path.Ext(f.Name))
}
