package sftp

import (
	"os"

	"github.com/alist-org/alist/v3/internal/model"
)

func fileToObj(f os.FileInfo) model.Obj {
	return &model.Object{
		Name:     f.Name(),
		Size:     f.Size(),
		Modified: f.ModTime(),
		IsFolder: f.IsDir(),
	}
}
