package fs

import (
	"github.com/alist-org/alist/v3/internal/model"
	"sync"
)

var FsPool = sync.Pool{
	New: func() any {
		return &FileSystem{}
	},
}

func getEmptyFs() *FileSystem {
	return FsPool.Get().(*FileSystem)
}

func New(user *model.User) Fs {
	fs := getEmptyFs()
	fs.User = user
	return fs
}

func Recycle(fs Fs) {
	FsPool.Put(fs)
}
