package tool

import (
	"io"
	"os"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type AddUriArgs struct {
	Uri     string
	UID     string
	TempDir string
	Signal  chan int
}

type Status struct {
	Progress  float64
	NewTID    string
	Completed bool
	Status    string
	Err       error
}

type Tool interface {
	// Items return the setting items the tool need
	Items() []model.SettingItem
	Init() (string, error)
	IsReady() bool
	// AddURI add an uri to download, return the task id
	AddURI(args *AddUriArgs) (string, error)
	// Remove the download if task been canceled
	Remove(tid string) error
	// Status return the status of the download task, if an error occurred, return the error in Status.Err
	Status(tid string) (*Status, error)
	// GetFile return an io.ReadCloser as the download file, if nil, means walk the temp dir to get the files
	GetFile(tid string) *File
}

type File struct {
	io.ReadCloser
	Name     string
	Size     int64
	Path     string
	Modified time.Time
}

func (f *File) GetReadCloser() (io.ReadCloser, error) {
	if f.ReadCloser != nil {
		return f.ReadCloser, nil
	}
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
