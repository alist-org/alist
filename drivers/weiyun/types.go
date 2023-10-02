package weiyun

import (
	"github.com/alist-org/alist/v3/pkg/utils"
	"time"

	weiyunsdkgo "github.com/foxxorcat/weiyun-sdk-go"
)

type File struct {
	PFolder *Folder
	weiyunsdkgo.File
}

func (f *File) GetID() string      { return f.FileID }
func (f *File) GetSize() int64     { return f.FileSize }
func (f *File) GetName() string    { return f.FileName }
func (f *File) ModTime() time.Time { return time.Time(f.FileMtime) }
func (f *File) IsDir() bool        { return false }
func (f *File) GetPath() string    { return "" }

func (f *File) GetPKey() string {
	return f.PFolder.DirKey
}
func (f *File) CreateTime() time.Time {
	return time.Time(f.FileCtime)
}

func (f *File) GetHash() utils.HashInfo {
	return utils.NewHashInfo(utils.SHA1, f.FileSha)
}

type Folder struct {
	PFolder *Folder
	weiyunsdkgo.Folder
}

func (f *Folder) CreateTime() time.Time {
	return time.Time(f.DirCtime)
}

func (f *Folder) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (f *Folder) GetID() string      { return f.DirKey }
func (f *Folder) GetSize() int64     { return 0 }
func (f *Folder) GetName() string    { return f.DirName }
func (f *Folder) ModTime() time.Time { return time.Time(f.DirMtime) }
func (f *Folder) IsDir() bool        { return true }
func (f *Folder) GetPath() string    { return "" }

func (f *Folder) GetPKey() string {
	return f.PFolder.DirKey
}
